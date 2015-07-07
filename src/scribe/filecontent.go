// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

type FileContent struct {
	Path       string `json:"path"`
	File       string `json:"file"`
	Expression string `json:"expression"`

	Matches []ContentMatch
}

type ContentMatch struct {
	Path    string
	Matches []string
}

func (f *FileContent) expandVariables(v []Variable) {
	f.Path = variableExpansion(v, f.Path)
	f.File = variableExpansion(v, f.File)
}

func (f *FileContent) getCriteria() (ret []EvaluationCriteria) {
	for _, x := range f.Matches {
		for _, y := range x.Matches {
			n := EvaluationCriteria{}
			n.Identifier = x.Path
			n.TestValue = y
			ret = append(ret, n)
		}
	}
	return ret
}

func (f *FileContent) prepare() error {
	debugPrint("prepare(): analyzing file system, path %v, file \"%v\"\n", f.Path, f.File)

	sfl := NewSimpleFileLocator()
	sfl.root = f.Path
	err := sfl.Locate(f.File, true)
	if err != nil {
		return err
	}

	for _, x := range sfl.matches {
		m, err := FileContentCheck(x, f.Expression)
		// XXX These soft errors during preparation are ignored right
		// now, but they should probably be tracked somewhere.
		if err != nil {
			continue
		}
		if m == nil || len(m) == 0 {
			continue
		}

		for _, i := range m {
			if len(i) < 2 {
				continue
			}
			ncm := ContentMatch{}
			ncm.Path = x
			ncm.Matches = make([]string, 0)
			for j := 1; j < len(i); j++ {
				debugPrint("prepare(): matched %v, \"%v\"\n", ncm.Path, i[j])
				ncm.Matches = append(ncm.Matches, i[j])
			}
			f.Matches = append(f.Matches, ncm)
		}
	}

	return nil
}

type SimpleFileLocator struct {
	executed bool
	root     string
	curDepth int
	maxDepth int
	matches  []string
}

func NewSimpleFileLocator() (ret SimpleFileLocator) {
	// XXX This needs to be fixed to work with Windows.
	ret.root = "/"
	ret.maxDepth = 10
	ret.matches = make([]string, 0)
	return ret
}

func (s *SimpleFileLocator) Locate(target string, useRegexp bool) error {
	if s.executed {
		return fmt.Errorf("locator has already been executed")
	}
	s.executed = true
	return s.locateInner(target, useRegexp, "")
}

func (s *SimpleFileLocator) locateInner(target string, useRegexp bool, path string) error {
	var (
		spath string
		re    *regexp.Regexp
		err   error
	)

	// If processing this directory would result in us exceeding the
	// specified search depth, just ignore it.
	if (s.curDepth + 1) > s.maxDepth {
		return nil
	}

	if useRegexp {
		re, err = regexp.Compile(target)
		if err != nil {
			return err
		}
	}

	s.curDepth++
	defer func() {
		s.curDepth--
	}()

	if path == "" {
		spath = s.root
	} else {
		spath = path
	}
	dirents, err := ioutil.ReadDir(spath)
	if err != nil {
		// If we encounter an error while reading a directory, just
		// ignore it and keep going until we are finished.
		return nil
	}
	for _, x := range dirents {
		fname := filepath.Join(spath, x.Name())
		if x.IsDir() {
			err = s.locateInner(target, useRegexp, fname)
			if err != nil {
				return err
			}
		} else if x.Mode().IsRegular() {
			if !useRegexp {
				if x.Name() == target {
					s.matches = append(s.matches, fname)
				}
			} else {
				if re.MatchString(x.Name()) {
					s.matches = append(s.matches, fname)
				}
			}
		}
	}
	return nil
}

func FileContentCheck(path string, regex string) ([][]string, error) {
	re, err := regexp.Compile(regex)
	if err != nil {
		return nil, err
	}
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		fd.Close()
	}()

	rdr := bufio.NewReader(fd)
	ret := make([][]string, 0)
	for {
		ln, err := rdr.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		mtch := re.FindStringSubmatch(ln)
		if len(mtch) > 0 {
			newmatch := make([]string, 0)
			newmatch = append(newmatch, mtch[0])
			for i := 1; i < len(mtch); i++ {
				newmatch = append(newmatch, mtch[i])
			}
			ret = append(ret, newmatch)
		}
	}

	if len(ret) == 0 {
		return nil, nil
	}
	return ret, nil
}
