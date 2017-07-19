// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"fmt"
	"github.com/mozilla/scribe"
)

type centosRelease struct {
	identifier   int
	versionMatch string
	expression   string
	filematch    string
}

const centos_expression = ".*CentOS.*(release \\d+)\\..*"

var centosReleases = []centosRelease{
	{PLATFORM_CENTOS_7, "release 7", centos_expression, "^centos-release$"},
	{PLATFORM_CENTOS_6, "release 6", centos_expression, "^centos-release$"},
}

// Adds a release test to scribe document doc. The release test is a dependency
// for each other vuln check, and validates if a given package is vulnerable that the
// platform is also what is expected (e.g., package X is vulnerable and operating system
// is also X.
func centosReleaseTest(platform supportedPlatform, doc *scribe.Document) (tid string, err error) {
	var (
		test    scribe.Test
		obj     scribe.Object
		release centosRelease
	)

	// Set the name and referenced object for the release test
	test.TestID = fmt.Sprintf("test-release-%v", platform.name)
	test.Object = "test-release"

	// Set our match value on the test to the release string
	found := false
	for _, x := range centosReleases {
		if x.identifier == platform.platformId {
			found = true
			release = x
			break
		}
	}
	if !found {
		err = fmt.Errorf("unable to locate release version match for %v", platform.name)
		return
	}
	test.EMatch.Value = release.versionMatch

	// Add our object, which will be the file we will match against to determine
	// if the platform is in scope
	obj.Object = test.Object
	obj.FileContent.Path = "/etc"
	obj.FileContent.File = release.filematch
	obj.FileContent.Expression = release.expression

	doc.Tests = append(doc.Tests, test)
	doc.Objects = append(doc.Objects, obj)
	tid = test.TestID
	return
}
