// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package scribe

import (
	"fmt"
	"strings"
)

// Describes the results of a test. The type can be marshaled into a JSON
// string as required.
type TestResult struct {
	Name       string   `json:"name"`       // The name of the test.
	Identifier string   `json:"identifier"` // The identifier for the test.
	Aliases    []string `json:"aliases"`    // Aliases for the test.

	IsError bool   `json:"iserror"` // True of error is encountered during evaluation.
	Error   string `json:"error"`   // Error associated with test.

	MasterResult   bool `json:"masterresult"`   // Master result of test.
	HasTrueResults bool `json:"hastrueresults"` // True if > 0 evaluations resulted in true.

	Results []TestSubResult `json:"results"` // The sub-results for the test.
}

// For a given test, a number of sources can be identified that match the
// criteria. For example, multiple files can be identifier with a given
// filename. Each test tracks individual results for these cases.
type TestSubResult struct {
	Result     bool   `json:"result"`     // The result of evaluation for an identifier source.
	Identifier string `json:"identifier"` // The identifier for the source.
}

// Return test results for a given test. Returns an error if for
// some reason the results can not be returned.
func GetResults(d *Document, name string) (TestResult, error) {
	t, err := d.getTest(name)
	if err != nil {
		return TestResult{}, err
	}
	ret := TestResult{}
	ret.Identifier = t.Identifier
	if t.err != nil {
		ret.Error = fmt.Sprintf("%v", t.err)
		ret.IsError = true
		return ret, nil
	}
	ret.MasterResult = t.masterResult
	ret.HasTrueResults = t.hasTrueResults
	for _, x := range t.results {
		nr := TestSubResult{}
		nr.Result = x.result
		nr.Identifier = x.criteria.identifier
		ret.Results = append(ret.Results, nr)
	}
	return ret, nil
}

// A helper function to convert Testresult r into a slice of greppable single
// line results. Note that each line returned is not terminated with a line
// feed.
func (r *TestResult) SingleLineResults() []string {
	lns := make([]string, 0)

	rs := "[error]"
	if !r.IsError {
		if r.MasterResult {
			rs = "[true]"
		} else {
			rs = "[false]"
		}
	}
	buf := fmt.Sprintf("master %v name:\"%v\" hastrue:%v error:\"%v\"", rs, r.Identifier, r.HasTrueResults, r.Error)
	lns = append(lns, buf)

	for _, x := range r.Results {
		if x.Result {
			rs = "[true]"
		} else {
			rs = "[false]"
		}
		buf := fmt.Sprintf("sub %v name:\"%v\" identifier:\"%v\"", rs, r.Identifier, x.Identifier)
		lns = append(lns, buf)
	}

	return lns
}

// A helper function to convert TestResult into a human readable result
// suitable for display.
func (r *TestResult) String() string {
	lns := make([]string, 0)
	lns = append(lns, fmt.Sprintf("result for \"%v\"", r.Identifier))
	if r.MasterResult {
		lns = append(lns, "\tmaster result: true")
	} else {
		buf := "\tmaster result: false"
		if r.HasTrueResults {
			buf = buf + ", has true results, failure caused by dependency"
		}
		lns = append(lns, buf)
	}
	if r.IsError {
		buf := fmt.Sprintf("\t[error] error: %v", r.Error)
		lns = append(lns, buf)
	}
	for _, x := range r.Results {
		buf := fmt.Sprintf("\t[%v] identifier: \"%v\"", x.Result, x.Identifier)
		lns = append(lns, buf)
	}
	return strings.Join(lns, "\n")
}
