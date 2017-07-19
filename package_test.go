// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package scribe_test

import (
	"github.com/mozilla/scribe"
	"strings"
	"testing"
)

// Used in TestPackagePolicy
var packagePolicyDoc = `
{
	"objects": [
	{
		"object": "openssl-package",
		"package": {
			"name": "openssl"
		}
	},

	{
		"object": "libbind-package",
		"package": {
			"name": "libbind"
		}
	},

	{
		"object": "grub-common-package",
		"package": {
			"name": "grub-common"
		}
	},

	{
		"object": "kernel-package-newest",
		"package": {
			"name": "kernel",
			"onlynewest": true
		}
	}
	],

	"tests": [
	{
		"test": "package0",
		"expectedresult": true,
		"object": "openssl-package"
	},

	{
		"test": "package1",
		"expectedresult": true,
		"object": "libbind-package",
		"evr": {
			"operation": "=",
			"value": "1:9.9.5.dfsg-4.3"
		}
	},

	{
		"test": "package2",
		"expectedresult": false,
		"object": "grub-common-package",
		"evr": {
			"operation": "<",
			"value": "2.02-beta1"
		}
	},

	{
		"test": "package3",
		"expectedresult": false,
		"object": "grub-common-package",
		"evr": {
			"operation": "<",
			"value": "2.02-beta2"
		}
	},

	{
		"test": "package4",
		"expectedresult": false,
		"object": "grub-common-package",
		"evr": {
			"operation": "<",
			"value": "2.01-beta2"
		}
	},

	{
		"test": "package5",
		"expectedresult": false,
		"object": "grub-common-package",
		"evr": {
			"operation": "<",
			"value": "2.02-beta3"
		},
		"if": [ "package2" ]
	},

	{
		"test": "package6",
		"expecterror": true,
		"object": "openssl-package",
		"evr": {
			"operation": "badop",
			"value": "1.0.1e"
		}
	},

	{
		"test": "package7",
		"expectedresult": true,
		"object": "openssl-package",
		"evr": {
			"operation": ">",
			"value": "1.0.1d"
		}
	},

	{
		"test": "package8",
		"expectedresult": false,
		"object": "kernel-package-newest",
		"evr": {
			"operation": "<",
			"value": "2.6.32-573.8.1.el6.x86_64"
		}
	}
	]
}
`

func TestPackagePolicy(t *testing.T) {
	rdr := strings.NewReader(packagePolicyDoc)
	scribe.Bootstrap()
	scribe.TestHooks(true)
	doc, err := scribe.LoadDocument(rdr)
	if err != nil {
		t.Fatalf("scribe.LoadDocument: %v", err)
	}
	err = scribe.AnalyzeDocument(doc)
	if err != nil {
		t.Fatalf("scribe.AnalyzeDocument: %v", err)
	}
	// Get results for each test and make sure the result matches what
	// expectedresult is set to
	testids := doc.GetTestIdentifiers()
	for _, x := range testids {
		stest, err := doc.GetTest(x)
		if err != nil {
			t.Fatalf("Document.GetTest: %v", err)
		}
		sres, err := scribe.GetResults(&doc, x)
		if err != nil {
			t.Fatalf("scribe.GetResults: %v", err)
		}
		if stest.ExpectError {
			if !sres.IsError {
				t.Fatalf("test %v should have been an error", x)
			}
		} else {
			if sres.IsError {
				t.Fatalf("test %v resulted in an error", x)
			}
			if sres.MasterResult != stest.ExpectedResult {
				t.Fatalf("result incorrect for test %v", x)
			}
		}
	}
}

func TestPackageQuery(t *testing.T) {
	scribe.Bootstrap()
	scribe.TestHooks(true)
	pinfo := scribe.QueryPackages()
	for _, x := range pinfo {
		t.Logf("%v %v %v", x.Name, x.Version, x.Type)
	}
	if len(pinfo) != 7 {
		t.FailNow()
	}
}
