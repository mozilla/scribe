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

type test struct {
	Name        string      `json:"name"`
	Identifier  string      `json:"identifier"`
	Aliases     []string    `json:"aliases"`
	Package     pkg         `json:"package"`
	Modifier    modifier    `json:"modifier"`
	FileContent filecontent `json:"filecontent"`
	FileName    filename    `json:"filename"`
	EVR         evrtest     `json:"evr"`
	Regexp      regex       `json:"regexp"`
	If          []string    `json:"if"`

	Expected bool `json:"expectedresult"`

	prepared  bool // True if test has been prepared.
	evaluated bool // True if test has been evaluated at least once.

	err error // The last error condition encountered during preparation or execution.

	// The final result for this test, a rolled up version of the results
	// of this test for any identified candidates. If at least one
	// candidate for the test evaluated to true, the master result will be
	// true.
	masterResult   bool               // The final result for the test.
	hasTrueResults bool               // True if at least one result evaluated to true.
	results        []evaluationResult // A slice of results for the test.
}

// The result of evaluation of a test. There can be more then one
// EvaluationResult present in the results of a test, if the source
// information returned more than one matching object.
type evaluationResult struct {
	criteria evaluationCriteria // Criteria used during evaluation.
	result   bool               // The result of the evaluation.
}

// Generic criteria for an evaluation. A source object should always support
// conversion from the specific type to a set of evaluation criteria.
//
// An identifier is used to track the source of an evaluation. For example,
// this may be a filename or a package name. In those examples, the testValue
// may be matched content from the file, or a package version string.
type evaluationCriteria struct {
	identifier string // The identifier used to track the source.
	testValue  string // the actual test data passed to the evaluator.
}

type genericSource interface {
	prepare() error
	getCriteria() []evaluationCriteria
	expandVariables([]variable)
	validate() error
	isModifier() bool
}

type genericEvaluator interface {
	evaluate(evaluationCriteria) (evaluationResult, error)
}

func (t *test) validate(d *Document) error {
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

func (t *test) getEvaluationInterface() genericEvaluator {
	if t.EVR.Value != "" {
		return &t.EVR
	} else if t.Regexp.Value != "" {
		return &t.Regexp
	}
	// If no evaluation criteria exists, use a no op evaluator
	// which will always return true for the test if any source objects
	// are identified.
	return &noop{}
}

func (t *test) getSourceInterface() genericSource {
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

func (t *test) prepare(d *Document) error {
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
				t.err = err
				return t.err
			}
			err = dt.prepare(d)
			if err != nil {
				t.err = err
				return t.err
			}
			err = t.Modifier.Sources[i].selectCriteria(dt)
			if err != nil {
				t.err = err
				return t.err
			}
			t.Modifier.addMergeTarget(&t.Modifier.Sources[i])
		}
	}

	p := t.getSourceInterface()
	if p == nil {
		t.err = fmt.Errorf("source has no valid interface")
		return t.err
	}
	p.expandVariables(d.Variables)
	err := p.prepare()
	if err != nil {
		t.err = err
		return err
	}
	return nil
}

func (t *test) runTest(d *Document) error {
	if t.evaluated {
		return nil
	}

	// If this test has failed at some point, return the error.
	if t.err != nil {
		return t.err
	}

	debugPrint("runTest(): running \"%v\"\n", t.Name)
	t.evaluated = true
	// First, see if this test has any dependencies. If so, run those
	// before we execute this one.
	for _, x := range t.If {
		dt, err := d.getTest(x)
		if err != nil {
			t.err = err
			return t.err
		}
		err = dt.runTest(d)
		if err != nil {
			t.err = fmt.Errorf("a test dependency failed (\"%v\")", x)
			return t.err
		}
	}

	ev := t.getEvaluationInterface()
	if ev == nil {
		t.err = fmt.Errorf("test has no valid evaluation interface")
		return t.err
	}
	si := t.getSourceInterface()
	if si == nil {
		t.err = fmt.Errorf("test has no valid source interface")
		return t.err
	}
	for _, x := range si.getCriteria() {
		res, err := ev.evaluate(x)
		if err != nil {
			t.err = err
			return t.err
		}
		t.results = append(t.results, res)
	}

	// Set the master result for the test. If any of the dependent tests
	// are false from a master result perspective, this one is also false.
	// If at least one result for this test is true, the master result for
	// the test is true.
	t.hasTrueResults = false
	for _, x := range t.results {
		if x.result {
			t.hasTrueResults = true
		}
	}
	t.masterResult = false
	if t.hasTrueResults {
		t.masterResult = true
	}
	for _, x := range t.If {
		dt, err := d.getTest(x)
		if err != nil {
			t.err = err
			t.masterResult = false
			return t.err
		}
		if !dt.masterResult {
			t.masterResult = false
			break
		}
	}

	// See if there is a test expected result handler installed, if so
	// validate it and call the handler if required.
	if sRuntime.excall != nil {
		if t.masterResult != t.Expected {
			tr, err := GetResults(d, t.Name)
			if err != nil {
				panic("GetResults() in expected handler")
			}
			sRuntime.excall(tr)
		}
	}

	return nil
}
