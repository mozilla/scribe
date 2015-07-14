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
)

type FileName struct {
	Path string `json:"path"`
	File string `json:"file"`

	matches []nameMatch
}

type nameMatch struct {
	path  string
	match string
}

func (f *FileName) validate() error {
	if len(f.Path) == 0 {
		return fmt.Errorf("filename path must be set")
	}
	if len(f.File) == 0 {
		return fmt.Errorf("filename file must be set")
	}
	return nil
}

func (f *FileName) expandVariables(v []Variable) {
	f.Path = variableExpansion(v, f.Path)
}

func (f *FileName) getCriteria() (ret []EvaluationCriteria) {
	for _, x := range f.matches {
		n := EvaluationCriteria{}
		n.Identifier = x.path
		n.TestValue = x.match
		ret = append(ret, n)
	}
	return ret
}

func (f *FileName) prepare() error {
	debugPrint("prepare(): analyzing file system, path %v, file \"%v\"\n", f.Path, f.File)

	sfl := newSimpleFileLocator()
	sfl.root = f.Path
	err := sfl.locate(f.File, true)
	if err != nil {
		return err
	}

	re, err := regexp.Compile(f.File)
	if err != nil {
		return err
	}

	for _, x := range sfl.matches {
		mtch := re.FindStringSubmatch(x)
		if len(mtch) < 2 {
			continue
		}
		nnm := nameMatch{}
		nnm.path = x
		nnm.match = mtch[1]
		f.matches = append(f.matches, nnm)
	}

	return nil
}
