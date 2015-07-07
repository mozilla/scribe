// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"encoding/json"
	"io"
	"io/ioutil"
)

func LoadDocument(r io.Reader) (Document, error) {
	var ret Document

	debugPrint("loading new document\n")
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return ret, err
	}
	err = json.Unmarshal(b, &ret)
	if err != nil {
		return ret, err
	}

	debugPrint("new document has %v test(s)\n", len(ret.Tests))

	return ret, nil
}

func AnalyzeDocument(d Document) error {
	debugPrint("analyzing document...\n")
	err := d.runTests()
	if err != nil {
		return err
	}
	return nil
}
