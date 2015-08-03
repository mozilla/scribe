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

type object struct {
	Object      string      `json:"object"`
	FileContent filecontent `json:"filecontent"`
	FileName    filename    `json:"filename"`
	Package     pkg         `json:"package"`
	Raw         raw         `json:"raw"`

	prepared bool  // True if object has been prepared.
	err      error // The last error condition encountered during preparation.
}

type genericSource interface {
	prepare() error
	getCriteria() []evaluationCriteria
	expandVariables([]variable)
	validate() error
}

func (o *object) getSourceInterface() genericSource {
	if o.Package.Name != "" {
		return &o.Package
	} else if o.FileContent.Path != "" {
		return &o.FileContent
	} else if o.FileName.Path != "" {
		return &o.FileName
	} else if len(o.Raw.Identifiers) > 0 {
		return &o.Raw
	}
	return nil
}

func (o *object) prepare(d *Document) error {
	if o.prepared {
		return nil
	}
	o.prepared = true

	p := o.getSourceInterface()
	if p == nil {
		o.err = fmt.Errorf("object has no valid interface")
		return o.err
	}
	p.expandVariables(d.Variables)
	err := p.prepare()
	if err != nil {
		o.err = err
		return err
	}
	return nil
}
