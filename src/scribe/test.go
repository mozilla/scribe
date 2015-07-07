// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

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
	results   []evaluationResult
}

type evaluationResult struct {
	criteria evaluationCriteria
	result   bool
}

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
		return nil
	}
	return p.prepare()
}

func (t *Test) runTest(d *Document) error {
	if t.evaluated {
		return nil
	}

	debugPrint("runTest(): running \"%v\"\n", t.Name)
	t.evaluated = true
	// First, see if this test has any dependencies. If so, run those
	// before we execute this one.
	for _, x := range t.If {
		t, err := d.getTest(x)
		if err != nil {
			return err
		}
		t.runTest(d)
	}

	ev := t.getEvaluationInterface()
	for _, x := range t.getSourceInterface().getCriteria() {
		ev.evaluate(x)
	}

	return nil
}
