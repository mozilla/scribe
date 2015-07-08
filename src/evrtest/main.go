// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package main

import (
	"bufio"
	"fmt"
	"os"
	"scribe"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, "specify input test data file as argument\n")
		os.Exit(1)
	}
	fmt.Println("starting evr comparison tests")

	fd, err := os.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	scanner := bufio.NewScanner(fd)
	for scanner.Scan() {
		buf := strings.TrimSpace(scanner.Text())
		if len(buf) == 0 {
			continue
		}
		fmt.Printf("%v\n", buf)

		var opmode int
		s0 := strings.Fields(buf)
		switch s0[0] {
		case "=":
			opmode = scribe.EVROP_EQUALS
		case "<":
			opmode = scribe.EVROP_LESS_THAN
		default:
			fmt.Fprintf(os.Stderr, "unknown operation %v\n", s0[0])
			os.Exit(1)
		}
		result := scribe.TestEvrCompare(opmode, s0[1], s0[2])
		if !result {
			fmt.Println("FAIL")
			os.Exit(2)
		}
		fmt.Println("PASS")
	}
	fd.Close()

	fmt.Println("end evr comparison tests")
}
