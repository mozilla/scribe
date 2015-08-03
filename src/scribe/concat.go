// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package scribe

func criteriaConcat(in []evaluationCriteria, concat string) []evaluationCriteria {
	ret := make([]evaluationCriteria, 0)
	if len(in) == 0 {
		return ret
	}
	debugPrint("criteriaConcat(): applying concat with \"%v\"\n", concat)
	n := evaluationCriteria{}
	for _, x := range in {
		if len(n.testValue) == 0 {
			n.identifier = "concat:" + x.identifier
			n.testValue = x.testValue
		} else {
			n.identifier = n.identifier + "," + x.identifier
			n.testValue = n.testValue + concat + x.testValue
		}
	}
	debugPrint("criteriaConcat(): concat result is \"%v\"\n", n.testValue)
	ret = append(ret, n)
	return ret
}
