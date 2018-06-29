// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Zack Mullaly zack@mozilla.com

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	clairListVulnsEndptFmt string = "http://127.0.0.1:6060/v1/namespaces/%s/vulnerabilities?limit=99999"
	clairGetVulnsEndptFmt  string = "http://127.0.0.1:6060/v1/namespaces/%s/vulnerabilities/%s?fixedIn"
)

type ShortVulnResponse struct {
	Vulns    []ShortVuln `json:"Vulnerabilities"`
	Error    *string     `json:"Error"`
	NextPage string      `json:"NextPage"`
}

type LongVulnResponse struct {
	Vuln LongVuln `json:"Vulnerability"`
}

type ShortVuln struct {
	Name     string `json:"Name"`
	Link     string `json:"Link"`
	Severity string `json:"Severity"`
}

type LongVuln struct {
	Name        string `json:"Name"`
	Link        string `json:"Link"`
	Severity    string `json:"Severity"`
	Description string `json:"Description"`
	FixedIn     []Fix  `json:"FixedIn"`
}

type Fix struct {
	Name    string `json:"Name"`
	Version string `json:"Version"`
}

func listVulnsForNamespace(namespace string) ([]ShortVuln, error) {
	respData := ShortVulnResponse{}
	vulns := []ShortVuln{}
	page := ""

	for {
		url := fmt.Sprintf(clairListVulnsEndptFmt, namespace)
		if page != "" {
			url = fmt.Sprintf("%s&page=%s", url, page)
		}
		response, err := http.Get(url)
		if err != nil {
			return []ShortVuln{}, err
		}

		decoder := json.NewDecoder(response.Body)
		defer response.Body.Close()
		decodeErr := decoder.Decode(&respData)
		if decodeErr != nil {
			return []ShortVuln{}, decodeErr
		}

		if respData.Error != nil {
			return []ShortVuln{}, errors.New(*respData.Error)
		}

		if respData.NextPage == "" {
			break
		}
		page = respData.NextPage
		vulns = append(vulns, respData.Vulns...)
	}

	return respData.Vulns, nil
}

func describeVuln(namespace string, vuln ShortVuln) (LongVuln, error) {
	respData := LongVulnResponse{}

	vulnName := strings.Split(vuln.Name, " ")[0]
	url := fmt.Sprintf(clairGetVulnsEndptFmt, namespace, vulnName)
	response, err := http.Get(url)
	if err != nil {
		return LongVuln{}, err
	}

	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()
	decodeErr := decoder.Decode(&respData)
	if decodeErr != nil {
		return LongVuln{}, decodeErr
	}

	return respData.Vuln, nil
}

func vulnsInNamespace(namespace string) ([]VulnInfo, error) {
	vulns := []VulnInfo{}

	allVulns, err := listVulnsForNamespace(namespace)
	if err != nil {
		return []VulnInfo{}, err
	}

	for _, vuln := range allVulns {
		vulnDesc, err := describeVuln(namespace, vuln)
		if err != nil {
			return []VulnInfo{}, err
		}

		if len(vulnDesc.FixedIn) == 0 {
			continue
		}

		for _, fix := range vulnDesc.FixedIn {
			vulns = append(vulns, VulnInfo{
				Package:        fix.Name,
				Vulnerability:  vulnDesc.Name,
				Severity:       vulnDesc.Severity,
				Description:    vulnDesc.Description,
				FixedInVersion: fix.Version,
				InfoLink:       vulnDesc.Link,
			})
		}
	}

	return vulns, nil
}
