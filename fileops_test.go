// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package scribe_test

import (
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
	genericTestExec(t, hasLinePolicyDoc)
}
