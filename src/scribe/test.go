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

	evaluated bool
	err       error
	results   []evaluationResult
}

// The result of evaluation of a test. There can be more then one
// evaluationResult present in the results of a test, if the source
// information returned more than one matching object.
type evaluationResult struct {
	criteria evaluationCriteria
	result   bool
}

// Generic criteria for an evaluation. A source object should always support
// conversion from the specific type to a set of evaluation criteria.
type evaluationCriteria struct {
	Identifier string
	TestValue  string
}

type genericSource interface {
	prepare() error
	getCriteria() []evaluationCriteria
}

type genericEvaluator interface {
	evaluate(evaluationCriteria) evaluationResult
}

func (t *Test) getEvaluationInterface() genericEvaluator {
	if t.EVR.Value != "" {
		return &t.EVR
	} else if t.Regexp.Value != "" {
		return &t.Regexp
	}
	return nil
}

func (t *Test) getSourceInterface() genericSource {
	if t.Package.Name != "" {
		return &t.Package
	} else if t.FileContent.Path != "" {
		return &t.FileContent
	}
	return nil
}

func (t *Test) prepare() error {
	p := t.getSourceInterface()
	if p == nil {
		t.err = fmt.Errorf("source has no valid interface")
		return t.err
	}
	err := p.prepare()
	if err != nil {
		t.err = err
		return err
	}
	return nil
}

func (t *Test) runTest(d *Document) error {
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
		t, err := d.getTest(x)
		if err != nil {
			t.err = err
			return t.err
		}
		err = t.runTest(d)
		if err != nil {
			t.err = fmt.Errorf("a test dependency failed")
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
		ev.evaluate(x)
	}

	return nil
}
