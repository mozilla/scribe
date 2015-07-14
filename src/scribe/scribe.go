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

func Bootstrap() (err error) {
	return err
}

func ExpectedCallback(f func(Test)) {
	sRuntime.excall = f
}

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

func SetDebug(f bool, w io.Writer) {
	sRuntime.debugging = f
	sRuntime.debugWriter = w
	debugPrint("debugging enabled\n")
}
