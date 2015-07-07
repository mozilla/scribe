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
	FileContent FileContent `json:"filecontent"`
	EVR         EVRTest     `json:"evr"`
	Regexp      Regexp      `json:"regexp"`
	If          []string    `json:"if"`

	Expected bool `json:"expectedresult"`

	evaluated bool

	Err            error
	MasterResult   bool
	HasTrueResults bool
	Results        []EvaluationResult
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
}

type genericEvaluator interface {
	evaluate(EvaluationCriteria) (EvaluationResult, error)
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
	return &NOOP{}
}

func (t *Test) getSourceInterface() genericSource {
	if t.Package.Name != "" {
		return &t.Package
	} else if t.FileContent.Path != "" {
		return &t.FileContent
	}
	return nil
}

func (t *Test) prepare(v []Variable) error {
	p := t.getSourceInterface()
	if p == nil {
		t.Err = fmt.Errorf("source has no valid interface")
		return t.Err
	}
	p.expandVariables(v)
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
			t.Err = fmt.Errorf("a test dependency failed")
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
