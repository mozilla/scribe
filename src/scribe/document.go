// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

// Describes a scribe document for interpretation. This structure represents
// the root of a description for analysis.
type Document struct {
	Tests []Test `json:"tests"`
}

func (d *Document) runTests() error {
	for i := range d.Tests {
		d.Tests[i].prepare()
	}
	return nil
}
