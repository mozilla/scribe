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

// A scribe document. Contains all tests and other information used to execute
// the document.
type Document struct {
	Variables []variable `json:"variables"`
	Objects   []object   `json:"objects"`
	Tests     []test     `json:"tests"`
}

// Validate a scribe document for consistency. This identifies any errors in
// the document that are not JSON syntax related, including missing fields or
// references to tests that do not exist. Returns an error if validation fails.
func (d *Document) Validate() error {
	for i := range d.Objects {
		err := d.Objects[i].validate(d)
		if err != nil {
			return err
		}
	}
	for i := range d.Tests {
		err := d.Tests[i].validate(d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Return the test IDs of all tests present in a document.
func (d *Document) GetTestIdentifiers() []string {
	ret := make([]string, 0)
	for _, x := range d.Tests {
		ret = append(ret, x.TestID)
	}
	return ret
}

func (d *Document) prepareObjects() error {
	// Note that prepare() will return an error if something goes wrong
	// but we don't propagate this back. Errors within object preparation
	// are kept localized to the object, and are not considered fatal to
	// execution of the entire document.
	for i := range d.Objects {
		d.Objects[i].prepare(d)
	}
	return nil
}

func (d *Document) objectPrepared(obj string) (bool, error) {
	var objptr *object
	for i := range d.Objects {
		if d.Objects[i].Object == obj {
			objptr = &d.Objects[i]
		}
	}
	if objptr == nil {
		return false, fmt.Errorf("unknown object \"%v\"", obj)
	}
	return objptr.prepared, nil
}

func (d *Document) runTests() error {
	// As documented prepareObjects(), we don't propagate errors here but
	// instead keep them localized to the test.
	for i := range d.Tests {
		d.Tests[i].runTest(d)
	}
	return nil
}

// Return a pointer to a test instance of the test whose identifier matches
func (d *Document) getTest(testid string) (*test, error) {
	for i := range d.Tests {
		if d.Tests[i].TestID == testid {
			return &d.Tests[i], nil
		}
	}
	return nil, fmt.Errorf("unknown test \"%v\"", testid)
}

// Given an object name, return a generic source interface for the object.
func (d *Document) getObjectInterface(obj string) (genericSource, error) {
	for i := range d.Objects {
		if d.Objects[i].Object == obj {
			return d.Objects[i].getSourceInterface(), nil
		}
	}
	return nil, fmt.Errorf("invalid object \"%v\"", obj)
}
