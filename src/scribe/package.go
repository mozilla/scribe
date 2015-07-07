// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com
package scribe

type Package struct {
	Name    string `json:"name"`
	pkgInfo []PackageInfo
}

type PackageInfo struct {
	Name    string
	Version string
}

func (p *Package) getCriteria() (ret []EvaluationCriteria) {
	for _, x := range p.pkgInfo {
		n := EvaluationCriteria{}
		n.Identifier = x.Name
		n.TestValue = x.Version
		ret = append(ret, n)
	}
	return ret
}

func (p *Package) prepare() error {
	debugPrint("prepare(): preparing information for package \"%v\"\n", p.Name)
	p.pkgInfo = make([]PackageInfo, 0)
	ret := getPackage(p.Name)
	for _, x := range ret.results {
		n := PackageInfo{}
		n.Name = x.name
		n.Version = x.version
		p.pkgInfo = append(p.pkgInfo, n)
	}
	return nil
}

func (p *Package) expandVariables(v []Variable) {
	p.Name = variableExpansion(v, p.Name)
}
