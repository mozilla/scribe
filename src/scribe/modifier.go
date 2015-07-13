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

type Modifier struct {
	Sources []ModifierSource `json:"sources"`
	Concat  ConcatModifier   `json:"concat"`
}

func (m *Modifier) addMergeTarget(ms *ModifierSource) {
	if m.Concat.Operator != "" {
		m.Concat.setMergeTarget(ms)
	} else {
		return
	}
	debugPrint("addMergeTarget(): merge target at %p\n", ms)
}

type ModifierSource struct {
	Name   string `json:"name"`
	Select string `json:"select"`

	criteria modifierData
}

func (m *ModifierSource) selectCriteria(t *Test) error {
	debugPrint("selectCriteria(): modifier selecting criteria from \"%v\"\n", t.Name)
	// XXX Just support "all" for now, this could change to select specific
	// elements of the source criteria slice.
	if m.Select != "all" {
		return fmt.Errorf("invalid selection criteria in modifier source")
	}
	s := t.getSourceInterface()
	if s == nil {
		return fmt.Errorf("source has no valid interface")
	}
	m.criteria.testName = t.Name
	m.criteria.criteria = s.getCriteria()
	debugPrint("selectCriteria(): copied %v criteria elements\n", len(m.criteria.criteria))
	return nil
}

type modifierData struct {
	testName string
	criteria []EvaluationCriteria
}

type mergingModifier struct {
	targets  []*ModifierSource
	criteria []EvaluationCriteria
}

func (m *mergingModifier) setMergeTarget(ms *ModifierSource) {
	m.targets = append(m.targets, ms)
}

func (m *mergingModifier) mergeTargets() {
	m.criteria = make([]EvaluationCriteria, 0)
	for _, x := range m.targets {
		for _, y := range x.criteria.criteria {
			m.criteria = append(m.criteria, y)
		}
	}
}

type ConcatModifier struct {
	Operator string `json:"operator"`
	mergingModifier
}

func (c *ConcatModifier) prepare() error {
	c.mergeTargets()
	return nil
}

func (c *ConcatModifier) expandVariables(v []Variable) {
}

func (c *ConcatModifier) getCriteria() []EvaluationCriteria {
	ret := make([]EvaluationCriteria, 0)
	nc := EvaluationCriteria{}
	ncid := ""
	buf := ""
	for _, x := range c.criteria {
		if len(buf) == 0 {
			ncid = "concat:" + x.Identifier
			buf = x.TestValue
		} else {
			ncid = ncid + "," + x.Identifier
			buf = buf + c.Operator + x.TestValue
		}
	}
	nc.Identifier = ncid
	nc.TestValue = buf
	ret = append(ret, nc)
	return ret
}
