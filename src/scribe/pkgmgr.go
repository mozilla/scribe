// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

import (
	"os/exec"
	"strings"
)

var pkgmgrInitialized bool
var pkgmgrCache []pkgmgrInfo

type pkgmgrResult struct {
	results []pkgmgrInfo
}

type pkgmgrInfo struct {
	name    string
	version string
	pkgtype string
}

func getPackage(name string) (ret pkgmgrResult) {
	ret.results = make([]pkgmgrInfo, 0)
	if !pkgmgrInitialized {
		pkgmgrInit()
	}
	debugPrint("getPackage(): looking for \"%v\"\n", name)
	for _, x := range pkgmgrCache {
		if x.name != name {
			continue
		}
		debugPrint("getPackage(): found %v, %v, %v\n", x.name, x.version, x.pkgtype)
		ret.results = append(ret.results, x)
	}
	debugPrint("getPackage(): returning %v entries\n", len(ret.results))
	return
}

func pkgmgrInit() {
	debugPrint("pkgmgrInit(): initializing package manager...\n")
	pkgmgrCache = make([]pkgmgrInfo, 0)
	if sRuntime.testHooks {
		pkgmgrCache = append(pkgmgrCache, testGetPackages()...)
	} else {
		pkgmgrCache = append(pkgmgrCache, rpmGetPackages()...)
		pkgmgrCache = append(pkgmgrCache, dpkgGetPackages()...)
	}
	pkgmgrInitialized = true
	debugPrint("pkgmgrInit(): initialized with %v packages\n", len(pkgmgrCache))
}

func rpmGetPackages() []pkgmgrInfo {
	ret := make([]pkgmgrInfo, 0)

	c := exec.Command("rpm", "-qa", "--queryformat", "%{NAME} %{EVR}\\n")
	buf, err := c.Output()
	if err != nil {
		return ret
	}

	slist := strings.Split(string(buf), "\n")
	for _, x := range slist {
		s := strings.Fields(x)

		if len(s) < 2 {
			continue
		}
		newpkg := pkgmgrInfo{}
		newpkg.name = s[0]
		newpkg.version = s[1]
		newpkg.pkgtype = "rpm"
		ret = append(ret, newpkg)
	}
	return ret
}

func dpkgGetPackages() []pkgmgrInfo {
	ret := make([]pkgmgrInfo, 0)

	c := exec.Command("dpkg", "-l")
	buf, err := c.Output()
	if err != nil {
		return nil
	}

	slist := strings.Split(string(buf), "\n")
	for _, x := range slist {
		s := strings.Fields(x)

		if len(s) < 3 {
			continue
		}
		// Only process packages that have been fully installed.
		if s[0] != "ii" {
			continue
		}
		newpkg := pkgmgrInfo{}
		newpkg.name = s[1]
		newpkg.version = s[2]
		newpkg.pkgtype = "dpkg"
		ret = append(ret, newpkg)
	}
	return ret
}

// Functions and data related to package tests

var testPkgTable = []struct {
	name string
	ver  string
}{
	{"openssl", "1.0.1e"},
	{"bash", "4.3-11"},
	{"upstart", "1.13.2"},
	{"grub-common", "2.02-beta2"},
	{"libbind", "1:9.9.5.dfsg-4.3"},
}

func testGetPackages() []pkgmgrInfo {
	ret := make([]pkgmgrInfo, 0)
	for _, x := range testPkgTable {
		newpkg := pkgmgrInfo{}
		newpkg.name = x.name
		newpkg.version = x.ver
		newpkg.pkgtype = "test"
		ret = append(ret, newpkg)
	}
	return ret
}
