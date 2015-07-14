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

type TestResult struct {
	// The name of the test.
	Name string `json:"name"`

	// True if error encountered during evaluation.
	IsError bool `json:"iserror"`
	// Error associated with test.
	Error string `json:"error"`

	// Master result of test.
	MasterResult bool `json:"masterresult"`
	// True if some source evaluations resulted in the test being true.
	HasTrueResults bool `json:"hastrueresults"`

	Results []TestSubResult `json:"results"`
}

type TestSubResult struct {
	// The result of evaluation for an identifier source.
	Result bool `json:"result"`
	// The identifier for the source.
	Identifier string `json:"identifier"`
}

// Return test results for a given test. Returns an error if an
// error occured during test preparation or execution.
func GetResults(d Document, name string) (TestResult, error) {
	t, err := d.getTest(name)
	if err != nil {
		return TestResult{}, err
	}
	if t.err != nil {
		return TestResult{}, t.err
	}
	ret := TestResult{}
	ret.Name = t.Name
	ret.Error = fmt.Sprintf("%v", t.err)
	ret.MasterResult = t.masterResult
	ret.HasTrueResults = t.hasTrueResults
	return TestResult{}, nil
}
