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

// Used in testConcatPolicy
var concatPolicyDoc = `
{
	"variables": [
	{ "key": "root", "value": "./test/concat" }
	],

	"objects": [
	{
		"object": "testfile0-content",
		"filecontent": {
			"path": "${root}",
			"file": "testfile0",
			"expression": "var = \\((\\S+), (\\S+)\\)",
			"concat": "."
		}
	}
	],

	"tests": [
	{
		"test": "testfile0-noop",
		"expectedresult": true,
		"object": "testfile0-content"
	}
	]
}
`

func TestConcatPolicy(t *testing.T) {
	genericTestExec(t, concatPolicyDoc)
}
