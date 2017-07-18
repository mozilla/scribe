// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.
//
// Contributor:
// - Aaron Meihm ameihm@mozilla.com

package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/lib/pq"
	"os"
)

const (
	PLATFORM_CENTOS_6 = iota
	PLATFORM_CENTOS_7
)

// Our list of platforms we will support policy generation for, this maps the
// platform constants to clair namespace identifiers
type supportedPlatform struct {
	platformId       int
	name             string
	clairNamespace   string
	clairNamespaceId int // Populated upon query of the database
}

var supportedPlatforms = []supportedPlatform{
	{PLATFORM_CENTOS_6, "centos6", "centos:6", 0},
	{PLATFORM_CENTOS_7, "centos7", "centos:7", 0},
}

// Given a clair namespace, return the supportedPlatform entry for it if it is
// supported, otherwise return an error
func getPlatform(clairNamespace string) (ret supportedPlatform, err error) {
	for _, x := range supportedPlatforms {
		if clairNamespace == x.clairNamespace {
			ret = x
			return
		}
	}
	err = fmt.Errorf("platform %v not supported", clairNamespace)
	return
}

type Policy struct {
	Vulnerabilities []Vulnerability `json:"vulnerabilities,omitempty"`
}

type Vulnerability struct {
	OS       string   `json:"os,omitempty"`
	Release  string   `json:"release,omitempty"`
	Package  string   `json:"package,omitempty"`
	Version  string   `json:"version,omitempty"`
	ID       string   `json:"id,omitempty"` // Globally unique test identifier
	Metadata Metadata `json:"metadata,omitempty"`
}

type vuln struct {
	name           string
	fixedInVersion string
	pkgName        string
	severity       string
	link           string
	description    string
}

type Metadata struct {
	Description string   `json:"description"`
	CVE         []string `json:"cve"`
	CVSS        string   `json:"cvss"`
	Category    string   `json:"category"`
}

type config struct {
	Database struct {
		DBName     string
		DBHost     string
		DBUser     string
		DBPassword string
		DBPort     string
	}
}

var dbconn *sql.DB
var cfg config

// Set any configuration based on environment variables
func configFromEnv() error {
	envvar := os.Getenv("PGHOST")
	if envvar != "" {
		cfg.Database.DBHost = envvar
	}
	envvar = os.Getenv("PGUSER")
	if envvar != "" {
		cfg.Database.DBUser = envvar
	}
	envvar = os.Getenv("PGPASSWORD")
	if envvar != "" {
		cfg.Database.DBPassword = envvar
	}
	envvar = os.Getenv("PGDATABASE")
	if envvar != "" {
		cfg.Database.DBName = envvar
	}
	envvar = os.Getenv("PGPORT")
	if envvar != "" {
		cfg.Database.DBPort = envvar
	}
	return nil
}

func dbInit() (err error) {
	connstr := fmt.Sprintf("dbname=%v host=%v user=%v password=%v port=%v sslmode=disable",
		cfg.Database.DBName, cfg.Database.DBHost, cfg.Database.DBUser,
		cfg.Database.DBPassword, cfg.Database.DBPort)
	dbconn, err = sql.Open("postgres", connstr)
	return
}

func generatePolicy(p string) error {
	var platform supportedPlatform
	// First make sure this is a supported platform, and this will also get us the namespace ID
	platforms, err := dbGetSupportedPlatforms()
	if err != nil {
		return err
	}
	supported := false
	for _, x := range platforms {
		if x.name == p {
			supported = true
			platform = x
			break
		}
	}
	if !supported {
		return fmt.Errorf("platform %v not supported for policy generation", p)
	}

	// Get all vulnerabilities for the platform from the database
	_, err = dbVulnsForPlatform(platform)
	if err != nil {
		return err
	}
	return nil
}

func main() {
	var (
		genPlatform  string
		showVersions bool
		err          error
	)
	flag.BoolVar(&showVersions, "V", false, "show distributions we can generate policies for and exit")
	flag.Parse()
	if len(flag.Args()) >= 1 {
		genPlatform = flag.Args()[0]
	}

	err = configFromEnv()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error reading config from environment: %v\n", err)
		os.Exit(1)
	}
	err = dbInit()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error initializing database: %v\n", err)
		os.Exit(1)
	}

	if showVersions {
		platforms, err := dbGetSupportedPlatforms()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error retrieving platforms: %v\n", err)
			os.Exit(1)
		}
		for _, x := range platforms {
			fmt.Printf("%v\n", x.name)
		}
		os.Exit(0)
	}

	if genPlatform == "" {
		fmt.Fprintf(os.Stderr, "error: platform to generate policy for not specified\n")
		os.Exit(1)
	}
	err = generatePolicy(genPlatform)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error generating policy: %v\n", err)
	}
}
