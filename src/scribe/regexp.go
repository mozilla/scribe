// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"regexp"
)

type Regexp struct {
	Value string `json:"value"`
}

func (r *Regexp) evaluate(c evaluationCriteria) (ret evaluationResult) {
	debugPrint("evaluate(): regexp %v \"%v\", \"%v\"\n", c.Identifier, c.TestValue, r.Value)
	re, err := regexp.Compile(r.Value)
	if err != nil {
		return
	}
	if re.MatchString(c.TestValue) {
		ret.criteria = c
		ret.result = true
	}
	return
}
