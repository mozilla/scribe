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

// Describes a scribe document for interpretation. This structure represents
// the root of a description for analysis.
type Document struct {
	Variables []Variable `json:"variables"`
	Tests     []Test     `json:"tests"`
}

func (d *Document) runTests() error {
	for i := range d.Tests {
		d.Tests[i].prepare(d.Variables)
	}
	for i := range d.Tests {
		d.Tests[i].runTest(d)
	}
	return nil
}

// Return a pointer to a test instance. Will locate the test whos name matches
// name, or has an alias that matches name.
func (d *Document) getTest(name string) (*Test, error) {
	for i := range d.Tests {
		if d.Tests[i].Name == name {
			return &d.Tests[i], nil
		}
		for _, x := range d.Tests[i].Aliases {
			if x == name {
				return &d.Tests[i], nil
			}
		}
	}

	return nil, fmt.Errorf("unknown test \"%v\"", name)
}
