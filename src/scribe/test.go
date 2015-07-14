// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"fmt"
)

type Test struct {
	Name        string      `json:"name"`
	Identifier  string      `json:"identifier"`
	Aliases     []string    `json:"aliases"`
	Package     Package     `json:"package"`
	Modifier    Modifier    `json:"modifier"`
	FileContent FileContent `json:"filecontent"`
	FileName    FileName    `json:"filename"`
	EVR         EVRTest     `json:"evr"`
	Regexp      Regexp      `json:"regexp"`
	If          []string    `json:"if"`

	Expected bool `json:"expectedresult"`

	// true if test has been prepared
	prepared bool

	// true if test has been evaluated at least once
	evaluated bool

	// The last error condition encountered during preparation
	// or execution.
	Err error

	// The final result for this test, a rolled up version of the results
	// of this test for any identified candidates.
	MasterResult bool
	// true if any candidates for this test returned a true result.
	HasTrueResults bool

	// Stores a slice of results for this test.
	Results []EvaluationResult
}

// The result of evaluation of a test. There can be more then one
// EvaluationResult present in the results of a test, if the source
// information returned more than one matching object.
type EvaluationResult struct {
	Criteria EvaluationCriteria
	Result   bool
}

// Generic criteria for an evaluation. A source object should always support
// conversion from the specific type to a set of evaluation criteria.
type EvaluationCriteria struct {
	Identifier string
	TestValue  string
}

type genericSource interface {
	prepare() error
	getCriteria() []EvaluationCriteria
	expandVariables([]Variable)
	validate() error
	isModifier() bool
}

type genericEvaluator interface {
	evaluate(EvaluationCriteria) (EvaluationResult, error)
}

func (t *Test) validate(d *Document) error {
	if len(t.Name) == 0 {
		return fmt.Errorf("a test in document has no name")
	}
	if len(t.Identifier) == 0 {
		return fmt.Errorf("%v: no identifier", t.Name)
	}
	si := t.getSourceInterface()
	if si == nil {
		return fmt.Errorf("%v: no valid source interface", t.Name)
	}
	err := si.validate()
	if err != nil {
		return fmt.Errorf("%v: %v", t.Name, err)
	}
	// If this is a modifier, ensure the modifier sources are valid.
	if si.isModifier() {
		for _, x := range t.Modifier.Sources {
			_, err := d.getTest(x.Name)
			if err != nil {
				return fmt.Errorf("%v: %v", t.Name, err)
			}
			if x.Select != "all" {
				return fmt.Errorf("%v: modifier source must include selector", t.Name)
			}
		}
	}
	if t.getEvaluationInterface() == nil {
		return fmt.Errorf("%v: no valid evaluation interface", t.Name)
	}
	for _, x := range t.Aliases {
		if len(x) == 0 {
			return fmt.Errorf("%v: bad alias within test", t.Name)
		}
	}
	for _, x := range t.If {
		ptr, err := d.getTest(x)
		if err != nil {
			return fmt.Errorf("%v: %v", t.Name, err)
		}
		if ptr == t {
			return fmt.Errorf("%v: test cannot reference itself", t.Name)
		}
	}
	return nil
}

func (t *Test) GetResults() ([]EvaluationResult, error) {
	if t.Err != nil {
		return nil, t.Err
	}
	return t.Results, nil
}

func (t *Test) getEvaluationInterface() genericEvaluator {
	if t.EVR.Value != "" {
		return &t.EVR
	} else if t.Regexp.Value != "" {
		return &t.Regexp
	}
	// If no evaluation criteria exists, use a no op evaluator
	// which will always return true for the test if any source objects
	// are identified.
	return &NOOP{}
}

func (t *Test) getSourceInterface() genericSource {
	if t.Package.Name != "" {
		return &t.Package
	} else if t.FileContent.Path != "" {
		return &t.FileContent
	} else if t.FileName.Path != "" {
		return &t.FileName
	} else if t.Modifier.Concat.Operator != "" {
		return &t.Modifier.Concat
	}
	return nil
}

func (t *Test) prepare(d *Document) error {
	if t.prepared {
		return nil
	}
	t.prepared = true

	// If this test is a modifier, prepare all the source tests first.
	if len(t.Modifier.Sources) != 0 {
		debugPrint("prepare(): readying modifier \"%v\"\n", t.Name)
		for i := range t.Modifier.Sources {
			nm := t.Modifier.Sources[i].Name
			debugPrint("prepare(): preparing modifier source \"%v\"\n", nm)
			dt, err := d.getTest(nm)
			if err != nil {
				t.Err = err
				return t.Err
			}
			err = dt.prepare(d)
			if err != nil {
				t.Err = err
				return t.Err
			}
			err = t.Modifier.Sources[i].selectCriteria(dt)
			if err != nil {
				t.Err = err
				return t.Err
			}
			t.Modifier.addMergeTarget(&t.Modifier.Sources[i])
		}
	}

	p := t.getSourceInterface()
	if p == nil {
		t.Err = fmt.Errorf("source has no valid interface")
		return t.Err
	}
	p.expandVariables(d.Variables)
	err := p.prepare()
	if err != nil {
		t.Err = err
		return err
	}
	return nil
}

func (t *Test) runTest(d *Document) error {
	if t.evaluated {
		return nil
	}

	// If this test has failed at some point, return the error.
	if t.Err != nil {
		return t.Err
	}

	debugPrint("runTest(): running \"%v\"\n", t.Name)
	t.evaluated = true
	// First, see if this test has any dependencies. If so, run those
	// before we execute this one.
	for _, x := range t.If {
		dt, err := d.getTest(x)
		if err != nil {
			t.Err = err
			return t.Err
		}
		err = dt.runTest(d)
		if err != nil {
			t.Err = fmt.Errorf("a test dependency failed (\"%v\")", x)
			return t.Err
		}
	}

	ev := t.getEvaluationInterface()
	if ev == nil {
		t.Err = fmt.Errorf("test has no valid evaluation interface")
		return t.Err
	}
	si := t.getSourceInterface()
	if si == nil {
		t.Err = fmt.Errorf("test has no valid source interface")
		return t.Err
	}
	for _, x := range si.getCriteria() {
		res, err := ev.evaluate(x)
		if err != nil {
			t.Err = err
			return t.Err
		}
		t.Results = append(t.Results, res)
	}

	// Set the master result for the test. If any of the dependent tests
	// are false from a master result perspective, this one is also false.
	// If at least one result for this test is true, the master result for
	// the test is true.
	t.HasTrueResults = false
	for _, x := range t.Results {
		if x.Result {
			t.HasTrueResults = true
		}
	}
	t.MasterResult = false
	if t.HasTrueResults {
		t.MasterResult = true
	}
	for _, x := range t.If {
		dt, err := d.getTest(x)
		if err != nil {
			t.Err = err
			t.MasterResult = false
			return t.Err
		}
		if !dt.MasterResult {
			t.MasterResult = false
			break
		}
	}

	// See if there is a test expected result handler installed, if so
	// validate it and call the handler if required.
	if sRuntime.excall != nil {
		if t.MasterResult != t.Expected {
			sRuntime.excall(*t)
		}
	}

	return nil
}
