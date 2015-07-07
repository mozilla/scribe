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

type EVRTest struct {
	Operation string `json:"operation"`
	Value     string `json:"value"`
}

func (e *EVRTest) evaluate(c evaluationCriteria) (ret evaluationResult, err error) {
	debugPrint("evaluate(): evr %v \"%v\", %v \"%v\"\n", c.Identifier, c.TestValue, e.Operation, e.Value)
	evrop := evrLookupOperation(e.Operation)
	if evrop == EVROP_UNKNOWN {
		return ret, fmt.Errorf("invalid evr operation \"%v\"", e.Operation)
	}
	if evrCompare(evrop, c.TestValue, e.Value) {
		debugPrint("evaluate(): evr comparison operation was true\n")
		ret.result = true
		ret.criteria = c
	}
	return
}
