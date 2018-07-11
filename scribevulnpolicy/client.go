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

// errorMessage contains error data that the Clair API may return.
type errorMessage struct {
	Message string `json:"Message"`
}

// shortVulnResponse contains the response data we expect from a request for a
// list of vulnerabilities for a namespace.
type shortVulnResponse struct {
	Vulns    []shortVuln   `json:"Vulnerabilities"`
	Error    *errorMessage `json:"Error"`
	NextPage string        `json:"NextPage"`
}

// longVulnResponse contains response data we expect from a request for more
// detailed information about a specific vulnerability.
type longVulnResponse struct {
	Vuln longVuln `json:"Vulnerability"`
}

// shortVuln contains only the information we need about vulnerabilities
// retrieved when we request a list of vulns for a namespace.
type shortVuln struct {
	Name string `json:"Name"`
}

// longVuln contains the information we need about specific vulnerabilities.
type longVuln struct {
	Name        string `json:"Name"`
	Link        string `json:"Link"`
	Severity    string `json:"Severity"`
	Description string `json:"Description"`
	FixedIn     []fix  `json:"FixedIn"`
}

// fix contains information about what version of a piece of software fixed a
// particular vulnerability.
type fix struct {
	Name    string `json:"Name"`
	Version string `json:"Version"`
}

func listVulnsForNamespace(namespace string) ([]shortVuln, error) {
	respData := shortVulnResponse{}
	vulns := []shortVuln{}

	// After we get a first response from Clair, we'll have the first
	// "NextPage" value, which we can use to page through results if
	// needed.  Once we *don't* get a "NextPage", we know we're done.
	page := ""

	for {
		url := fmt.Sprintf(clairListVulnsEndptFmt, namespace)
		if page != "" {
			url = fmt.Sprintf("%s&page=%s", url, page)
		}
		response, err := http.Get(url)
		if err != nil {
			return []shortVuln{}, err
		}

		decoder := json.NewDecoder(response.Body)
		decodeErr := decoder.Decode(&respData)
		response.Body.Close()
		if decodeErr != nil {
			return []shortVuln{}, decodeErr
		}

		if respData.Error != nil {
			return []shortVuln{}, errors.New(respData.Error.Message)
		}

		if respData.NextPage == "" {
			break
		}
		page = respData.NextPage
		vulns = append(vulns, respData.Vulns...)
	}

	return respData.Vulns, nil
}

func describeVuln(namespace string, vuln shortVuln) (longVuln, error) {
	respData := longVulnResponse{}

	vulnName := strings.Split(vuln.Name, " ")[0]
	url := fmt.Sprintf(clairGetVulnsEndptFmt, namespace, vulnName)
	response, err := http.Get(url)
	if err != nil {
		return longVuln{}, err
	}

	decoder := json.NewDecoder(response.Body)
	defer response.Body.Close()
	decodeErr := decoder.Decode(&respData)
	if decodeErr != nil {
		return longVuln{}, decodeErr
	}

	return respData.Vuln, nil
}

// VulnsInNamespace collects a list of vulnerabilities affecting a particular
// namespace.
func VulnsInNamespace(namespace string) ([]VulnInfo, error) {
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
