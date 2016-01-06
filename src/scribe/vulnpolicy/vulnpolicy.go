// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package vulnpolicy

import (
	"encoding/json"
	"fmt"
	"scribe"
)

type Policy struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

type Vulnerability struct {
	OS       string   `json:"os,omitempty"`
	Release  string   `json:"release,omitempty"`
	Package  string   `json:"package,omitempty"`
	Version  string   `json:"version,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	Description string   `json:"description"`
	CVE         []string `json:"cve"`
}

var testcntr int

func getReleaseTest(doc *scribe.Document, vuln Vulnerability) (string, error) {
	if vuln.OS == "ubuntu" {
		return ubuntuGetReleaseTest(doc, vuln)
	}
	return "", fmt.Errorf("unable to create release definition")
}

func addTest(doc *scribe.Document, vuln Vulnerability) error {
	// Get the release definition for the test, if it's missing from
	// the document it will be added
	reltestid, err := getReleaseTest(doc, vuln)
	if err != nil {
		return err
	}

	// See if we already have an object definition for the package, if
	// not add it
	objid := ""
	for _, x := range doc.Objects {
		if x.Package.Name == vuln.Package {
			objid = x.Object
			break
		}
	}
	if objid == "" {
		objid = fmt.Sprintf("obj-package-%v", vuln.Package)
		obj := scribe.Object{}
		obj.Object = objid
		obj.Package.Name = vuln.Package
		doc.Objects = append(doc.Objects, obj)
	}

	testidstr := fmt.Sprintf("test-%v-%v-%v-%v", vuln.OS, vuln.Release,
		vuln.Package, testcntr)
	test := scribe.Test{}
	test.TestID = testidstr
	test.Object = objid
	test.EVR.Value = vuln.Version
	test.EVR.Operation = "<"
	test.If = append(test.If, reltestid)
	doc.Tests = append(doc.Tests, test)
	testcntr++

	return nil
}

func DocumentFromPolicy(buf []byte) (ret scribe.Document, err error) {
	policy := Policy{}
	err = json.Unmarshal(buf, &policy)
	if err != nil {
		return
	}

	// Create a test for each vulnerability that is listed in the
	// policy. Create depedency release tests in the document as
	// well as we go.
	testcntr = 0
	for _, x := range policy.Vulnerabilities {
		err = addTest(&ret, x)
		if err != nil {
			return
		}
	}
	return
}
