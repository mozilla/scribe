// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"fmt"
	"io"
)

type runtime struct {
	debugging   bool
	debugWriter io.Writer
	excall      func(Test)
	testHooks   bool
}

const Version = "0.5"

var sRuntime runtime

func (r *runtime) initialize() {
}

func init() {
	sRuntime.initialize()
}

// Bootstrap the scribe library. This function is currently not used but code
// should call this function before any other functions in the library. An
// error is returned if one occurs.
func Bootstrap() (err error) {
	return err
}

// Set an expected result callback. f should be a function that takes a Test
// type as an argument. When this is set, if the result of a test does not
// match the value set in "expectedresult" for the test, the function is
// immediately called with the applicable test as an argument.
func ExpectedCallback(f func(Test)) {
	sRuntime.excall = f
}

// Enable or disable test hooks. If test hooks are enabled, certain functions
// such as requesting package data from the host system are bypassed in favor
// of test tables.
func TestHooks(f bool) {
	sRuntime.testHooks = f
}

func debugPrint(s string, args ...interface{}) {
	if !sRuntime.debugging {
		return
	}
	buf := fmt.Sprintf(s, args...)
	fmt.Fprintf(sRuntime.debugWriter, "[scribe] %v", buf)
}

// Enable or disable debugging. If debugging is enabled, output is written
// to the io.Writer specified by w.
func SetDebug(f bool, w io.Writer) {
	sRuntime.debugging = f
	sRuntime.debugWriter = w
	debugPrint("debugging enabled\n")
}
