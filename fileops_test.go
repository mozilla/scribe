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

// Used in TestHasLinePolicy
var hasLinePolicyDoc = `
{
        "variables": [
        { "key": "root", "value": "./test/hasline" }
        ],

        "objects": [
        {
                "object": "file-hasline",
                "hasline": {
                        "path": "${root}",
                        "file": ".*\\.txt",
                        "expression": ".*test.*"
                }
        }
        ],

        "tests": [
        {
                "test": "files-without-line",
                "expectedresult": true,
                "object": "file-hasline",
                "exactmatch": {
                        "value": "true"
                }
        }
        ]
}
`

func TestHasLinePolicy(t *testing.T) {
	rdr := strings.NewReader(hasLinePolicyDoc)
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
