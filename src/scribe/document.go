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
	Tests     []test     `json:"tests"`
}

// Validate a scribe document for consistency. This identifies any errors in
// the document that are not JSON syntax related, including missing fields or
// references to tests that do not exist. Returns an error if validation fails.
func (d *Document) Validate() error {
	for i := range d.Tests {
		err := d.Tests[i].validate(d)
		if err != nil {
			return err
		}
	}
	return nil
}

// Return the identifiers of all tests present in a document.
func (d *Document) GetTestIdentifiers() []string {
	ret := make([]string, 0)
	for _, x := range d.Tests {
		ret = append(ret, x.Identifier)
	}
	return ret
}

func (d *Document) runTests() error {
	// Note that prepare() and runTest() will return an error if something
	// goes wrong, but we don't propagate this back. Errors within a test
	// are kept localized in that test, and aren't considered to be a fatal
	// condition.
	for i := range d.Tests {
		d.Tests[i].prepare(d)
	}
	for i := range d.Tests {
		d.Tests[i].runTest(d)
	}
	return nil
}

// Return a pointer to a test instance of the test whose identifier matches
func (d *Document) getTest(identifier string) (*test, error) {
	for i := range d.Tests {
		if d.Tests[i].Identifier == identifier {
			return &d.Tests[i], nil
		}
	}
	return nil, fmt.Errorf("unknown test \"%v\"", identifier)
}
