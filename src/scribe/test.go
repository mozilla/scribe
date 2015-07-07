// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

type Test struct {
	Name        string      `json:"name"`
	Identifier  string      `json:"identifier"`
	Aliases     []string    `json:"aliases"`
	Package     Package     `json:"package"`
	FileContent FileContent `json:"filecontent"`
	EVR         EVR         `json:"evr"`
	Regexp      Regexp      `json:"regexp"`
}

type genericSource interface {
	prepare() error
}

func (t *Test) getSourceInterface() genericSource {
	if t.Package.Name != "" {
		return &t.Package
	} else if t.FileContent.Path != "" {
		return &t.FileContent
	}
	return nil
}

func (t *Test) prepare() error {
	p := t.getSourceInterface()
	if p == nil {
		return nil
	}
	return p.prepare()
}
