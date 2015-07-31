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

type modifier struct {
	Sources []modifierSource `json:"sources"`
	Concat  concatModifier   `json:"concat"`
}

func (m *modifier) addMergeTarget(ms *modifierSource) {
	if m.Concat.Operator != "" {
		m.Concat.setMergeTarget(ms)
	} else {
		return
	}
	debugPrint("addMergeTarget(): merge target at %p\n", ms)
}

type modifierSource struct {
	Identifier string `json:"identifier"`
	Select     string `json:"select"`

	criteria modifierData
}

func (m *modifierSource) selectCriteria(t *test) error {
	debugPrint("selectCriteria(): modifier selecting criteria from \"%v\"\n", t.Identifier)
	// XXX Just support "all" for now, this could change to select specific
	// elements of the source criteria slice.
	if m.Select != "all" {
		return fmt.Errorf("invalid selection criteria in modifier source")
	}
	s := t.getSourceInterface()
	if s == nil {
		return fmt.Errorf("source has no valid interface")
	}
	m.criteria.testName = t.Identifier
	m.criteria.criteria = s.getCriteria()
	debugPrint("selectCriteria(): copied %v criteria elements\n", len(m.criteria.criteria))
	return nil
}

type modifierData struct {
	testName string
	criteria []evaluationCriteria
}

type mergingModifier struct {
	targets  []*modifierSource
	criteria []evaluationCriteria
}

func (m *mergingModifier) setMergeTarget(ms *modifierSource) {
	m.targets = append(m.targets, ms)
}

func (m *mergingModifier) mergeTargets() {
	m.criteria = make([]evaluationCriteria, 0)
	for _, x := range m.targets {
		for _, y := range x.criteria.criteria {
			m.criteria = append(m.criteria, y)
		}
	}
}

type concatModifier struct {
	Operator string `json:"operator"`
	mergingModifier
}

func (c *concatModifier) prepare() error {
	c.mergeTargets()
	return nil
}

func (c *concatModifier) isModifier() bool {
	return true
}

func (c *concatModifier) validate() error {
	if len(c.Operator) == 0 {
		return fmt.Errorf("must specify concat operator")
	}
	return nil
}

func (c *concatModifier) expandVariables(v []variable) {
}

func (c *concatModifier) getCriteria() []evaluationCriteria {
	ret := make([]evaluationCriteria, 0)
	if len(c.criteria) == 0 {
		return ret
	}
	nc := evaluationCriteria{}
	ncid := ""
	buf := ""
	for _, x := range c.criteria {
		if len(buf) == 0 {
			ncid = "concat:" + x.identifier
			buf = x.testValue
		} else {
			ncid = ncid + "," + x.identifier
			buf = buf + c.Operator + x.testValue
		}
	}
	nc.identifier = ncid
	nc.testValue = buf
	ret = append(ret, nc)
	return ret
}
