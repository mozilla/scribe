// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package scribe

import (
	"fmt"
	"regexp"
)

// Normally when a Pkg object is prepared, the pkgInfo list will be filled with
// information related to any package installed which exactly matches Name. If
// the optional CollectMatch parameter is set, this regular expression will be
// used to match packages, but the package name will still be Name.
//
// For example, if Name is set to linux-image-generic, and CollectMatch is set
// to ^linux-image-.*-generic$, it will result in the object being populated
// with a set of version strings from all installed packages that match the
// regexp.
//
// This is intended to handle a case where we want to do a kernel package
// version comparison, but the kernel package name actually includes the a
// version string which makes a direct package name -> name comparison harder.
type Pkg struct {
	Name         string `json:"name,omitempty"`
	CollectMatch string `json:"collectmatch,omitempty"`
	pkgInfo      []packageInfo
}

type packageInfo struct {
	Name    string
	Version string
}

func (p *Pkg) isChain() bool {
	return false
}

func (p *Pkg) validate(d *Document) error {
	if len(p.Name) == 0 {
		return fmt.Errorf("package must specify name")
	}
	if len(p.CollectMatch) > 0 {
		_, err := regexp.Compile(p.CollectMatch)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Pkg) fireChains(d *Document) ([]evaluationCriteria, error) {
	return nil, nil
}

func (p *Pkg) mergeCriteria(c []evaluationCriteria) {
}

func (p *Pkg) getCriteria() (ret []evaluationCriteria) {
	for _, x := range p.pkgInfo {
		n := evaluationCriteria{}
		n.identifier = x.Name
		n.testValue = x.Version
		ret = append(ret, n)
	}
	return ret
}

func (p *Pkg) prepare() error {
	debugPrint("prepare(): preparing information for package \"%v\"\n", p.Name)
	p.pkgInfo = make([]packageInfo, 0)
	ret := getPackage(p.Name, p.CollectMatch)
	for _, x := range ret.results {
		n := packageInfo{}
		n.Name = x.name
		n.Version = x.version
		p.pkgInfo = append(p.pkgInfo, n)
	}
	return nil
}

func (p *Pkg) expandVariables(v []Variable) {
	p.Name = variableExpansion(v, p.Name)
}
