// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"crypto/md5"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/mozilla/scribe"
	"os"
)

// Our list of platforms we will support policy generation for, this maps the
// platform constants to clair namespace identifiers
type supportedPlatform struct {
	name           string
	clairNamespace string
	releaseTest    func(supportedPlatform, *scribe.Document) (string, error)
	pkgNewest      func(string) bool
}

var supportedPlatforms = []supportedPlatform{
	{"centos6", "centos:6", centosReleaseTest, centosOnlyNewest},
	{"centos7", "centos:7", centosReleaseTest, centosOnlyNewest},
}

// Given a clair namespace, return the supportedPlatform entry for it if it is
// supported, otherwise return an error
func getPlatform(clairNamespace string) (ret supportedPlatform, err error) {
	for _, x := range supportedPlatforms {
		if clairNamespace == x.clairNamespace {
			ret = x
			return
		}
	}
	err = fmt.Errorf("platform %v not supported", clairNamespace)
	return
}

type VulnInfo struct {
	Package        string `json:"package"`
	Vulnerability  string `json:"vulnerability"`
	Severity       string `json:"severity"`
	Description    string `json:"description"`
	FixedInVersion string `json:"fixedInVersion"`
	InfoLink       string `json:"link"`
}

// Generate a test identifier; this needs to be unique in the document. Here we
// just use a few elements from the vulnerability and platform and return an MD5
// digest.
func generateTestID(v VulnInfo, p supportedPlatform) (string, error) {
	h := md5.New()
	h.Write([]byte(v.Vulnerability))
	h.Write([]byte(p.name))
	h.Write([]byte(v.Package))
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}

func generatePolicy(p string) error {
	var (
		platform supportedPlatform
		doc      scribe.Document
	)
	// First make sure this is a supported platform, and this will also get us the namespace ID
	supported := false
	for _, pform := range supportedPlatforms {
		if pform.name == p {
			platform = pform
			supported = true
			break
		}
	}
	if !supported {
		return fmt.Errorf("platform %v not supported for policy generation", p)
	}

	// Add the release test which will be used as a dependency on all checks
	// in the final test document
	reltestid, err := platform.releaseTest(platform, &doc)
	if err != nil {
		return err
	}

	// Get all vulnerabilities for the platform from the Clair API
	vulns, err := vulnsInNamespace(platform.clairNamespace)
	if err != nil {
		return err
	}

	// Add a test for each vulnerability
	for _, x := range vulns {
		var (
			newtest scribe.Test
			newobj  scribe.Object
			objname string
		)

		// See if we already have an object in the document that references
		// the package we want to lookup, if so we don't need to add a second
		// one
		found := false
		objname = fmt.Sprintf("obj-package-%v", x.Package)
		for _, y := range doc.Objects {
			if y.Package.Name == x.Package {
				found = true
				break
			}
		}
		if !found {
			newobj.Object = objname
			newobj.Package.Name = x.Package
			newobj.Package.OnlyNewest = platform.pkgNewest(x.Package)
			doc.Objects = append(doc.Objects, newobj)
		}

		newtest.TestName = x.Vulnerability
		newtest.Object = objname
		newtest.EVR.Value = x.FixedInVersion
		newtest.EVR.Operation = "<"
		newtest.If = append(newtest.If, reltestid)
		newtest.TestID, err = generateTestID(x, platform)
		if err != nil {
			return err
		}
		// Add some tags to the test we can use when we parse results
		pkgtag := scribe.TestTag{Key: "package", Value: x.Package}
		newtest.Tags = append(newtest.Tags, pkgtag)
		sevtag := scribe.TestTag{Key: "severity", Value: x.Severity}
		newtest.Tags = append(newtest.Tags, sevtag)
		linktag := scribe.TestTag{Key: "link", Value: x.InfoLink}
		newtest.Tags = append(newtest.Tags, linktag)
		doc.Tests = append(doc.Tests, newtest)
	}

	// Finally, display the policy on stdout
	outbuf, err := json.MarshalIndent(doc, "", "    ")
	if err != nil {
		return err
	}
	fmt.Printf("%v\n", string(outbuf))

	return nil
}

func main() {
	var (
		genPlatform  string
		showVersions bool
		err          error
	)
	flag.BoolVar(&showVersions, "V", false, "show distributions we can generate policies for and exit")
	flag.Parse()
	if len(flag.Args()) >= 1 {
		genPlatform = flag.Args()[0]
	}

	if showVersions {
		for _, platform := range supportedPlatforms {
			fmt.Println(platform.name)
		}
		return
	}

	if genPlatform == "" {
		fmt.Fprintf(os.Stderr, "error: platform to generate policy for not specified\n")
		os.Exit(1)
	}
	err = generatePolicy(genPlatform)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating policy: %v\n", err)
	}
}
