// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

const (
	_ = iota
	EVROP_LESS_THAN
	EVROP_EQUALS
	EVROP_UNKNOWN
)

type EVR struct {
	epoch   string
	version string
	release string
}

func evrLookupOperation(s string) int {
	switch s {
	case "<":
		return EVROP_LESS_THAN
	case "=":
		return EVROP_EQUALS
	}
	return EVROP_UNKNOWN
}

func evrOperationStr(val int) string {
	switch val {
	case EVROP_LESS_THAN:
		return "<"
	case EVROP_EQUALS:
		return "="
	default:
		return "?"
	}
}

func evrIsDigit(c rune) bool {
	return unicode.IsDigit(c)
}

func evrExtract(s string) (EVR, error) {
	var ret EVR
	var idx int

	for _, c := range s {
		if !evrIsDigit(c) {
			break
		}
		idx++
	}

	if idx >= len(s) {
		return ret, fmt.Errorf("evrExtract: all digits")
	}

	if s[idx] == ':' {
		ret.epoch = s[:idx]
		idx++
	} else {
		ret.epoch = "0"
		idx = 0
	}

	if idx >= len(s) {
		return ret, fmt.Errorf("evrExtract: only epoch")
	}
	remain := s[idx:]

	rp0 := strings.LastIndex(remain, "-")
	if rp0 != -1 {
		ret.version = remain[:rp0]
		rp0++
		if rp0 >= len(remain) {
			return ret, fmt.Errorf("evrExtract: ends in dash")
		}
		ret.release = remain[rp0:]
	} else {
		ret.version = remain
		ret.release = ""
	}

	debugPrint("evrExtract(): epoch=%v, version=%v, revision=%v\n", ret.epoch, ret.version, ret.release)
	return ret, nil
}

func evrRpmTokenizer(s string) []string {
	re := regexp.MustCompile("[A-Za-z0-9]+")
	buf := re.FindAllString(s, -1)
	ret := make([]string, 0)
	var isnum bool
	var cmp string
	for _, x := range buf {
		cmp = ""
		for _, c := range x {
			if len(cmp) == 0 {
				if evrIsDigit(c) {
					isnum = true
				} else {
					isnum = false
				}
				cmp += string(c)
			} else {
				if isnum {
					if !evrIsDigit(c) {
						ret = append(ret, cmp)
						cmp = string(c)
						isnum = false
					} else {
						cmp += string(c)
					}
				} else {
					if evrIsDigit(c) {
						ret = append(ret, cmp)
						cmp = string(c)
						isnum = true
					} else {
						cmp += string(c)
					}
				}
			}
		}
		ret = append(ret, cmp)
	}
	return ret
}

func evrTrimZeros(s string) string {
	if len(s) == 1 {
		return s
	}
	_, err := strconv.Atoi(s)
	if err != nil {
		return s
	}
	return strings.TrimLeft(s, "0")
}

func evrRpmVerCmp(actual string, check string) int {
	if actual == check {
		return 0
	}

	acttokens := evrRpmTokenizer(actual)
	chktokens := evrRpmTokenizer(check)

	for i := range chktokens {
		if i >= len(acttokens) {
			// There are more tokens in the check value, the
			// check wins.
			return 1
		}

		// If the values are pure numbers, trim any leading 0's.
		acttest := evrTrimZeros(acttokens[i])
		chktest := evrTrimZeros(chktokens[i])

		// Numeric component will always win out over alpha.
		if evrIsDigit(rune(acttest[0])) && !evrIsDigit(rune(chktest[0])) {
			return -1
		}
		if evrIsDigit(rune(chktest[0])) && !evrIsDigit(rune(acttest[0])) {
			return 1
		}

		// Do a lexical string comparison here, this should work
		// even with pure integer values.
		if chktest > acttest {
			return 1
		} else if chktest < acttest {
			return -1
		}
	}

	// If we get this far, see if the actual value still has more tokens
	// for comparison, if so actual wins.
	if len(acttokens) > len(chktokens) {
		return -1
	}

	return 0
}

func evrRpmCompare(actual EVR, check EVR) (int, error) {
	aepoch, err := strconv.Atoi(actual.epoch)
	if err != nil {
		return 0, fmt.Errorf("evrRpmCompare: bad actual epoch")
	}
	cepoch, err := strconv.Atoi(check.epoch)
	if err != nil {
		return 0, fmt.Errorf("evrRpmCompare: bad check epoch")
	}
	if cepoch > aepoch {
		return 1, nil
	} else if cepoch < aepoch {
		return -1, nil
	}

	ret := evrRpmVerCmp(actual.version, check.version)
	if ret != 0 {
		return ret, nil
	}

	ret = evrRpmVerCmp(actual.release, check.release)
	if ret != 0 {
		return ret, nil
	}

	return 0, nil
}

func evrCompare(op int, actual string, check string) (bool, error) {
	debugPrint("evrCompare(): %v %v %v\n", actual, evrOperationStr(op), check)

	evract, err := evrExtract(actual)
	if err != nil {
		return false, err
	}
	evrchk, err := evrExtract(check)
	if err != nil {
		return false, err
	}

	ret, err := evrRpmCompare(evract, evrchk)
	if err != nil {
		return false, err
	}
	switch op {
	case EVROP_EQUALS:
		if ret != 0 {
			return false, nil
		}
		return true, nil
	case EVROP_LESS_THAN:
		if ret == 1 {
			return true, nil
		}
		return false, nil
	}
	return false, fmt.Errorf("evrCompare: unknown operator")
}

// Exported version of evrCompare() used for testing in evrtest.
func TestEvrCompare(op int, actual string, check string) (bool, error) {
	return evrCompare(op, actual, check)
}
