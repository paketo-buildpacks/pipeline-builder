/*
 * Copyright 2018-2020 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

const userAgent = "Paketo_Buildpacks_CI/1.0"

var client *http.Client = http.DefaultClient

func main() {
	inputs := actions.NewInputs()

	// Needs to be in this list
	//   (aoj, aoj_openj9, bisheng, corretto, dragonwell, graalvm_ce8, graalvm_ce11, graalvm_ce16, graalvm_ce17, graalvm_ce19, graalvm_ce20, graalvm_community, graalvm, jetbrains, kona, liberica, liberica_native, mandrel, microsoft, ojdk_build, openlogic, oracle, oracle_open_jdk, redhat, sap_machine, semeru, semeru_certified, temurin, trava, zulu, zulu_prime)
	// https://api.foojay.io/swagger-ui#/default/getPackagesV3
	d, ok := inputs["distro"]
	if !ok {
		panic(fmt.Errorf("distro must be specified"))
	}

	arch, ok := inputs["arch"]
	if !ok {
		arch = "x64"
	}

	// Needs to be in this list
	//   (aarch32, aarch64, amd64, arm, arm32, arm64, mips, ppc, ppc64el, ppc64le, ppc64, riscv64, s390, s390x, sparc, sparcv9, x64, x86-64, x86, i386, i486, i586, i686, x86-32)
	// https://api.foojay.io/swagger-ui#/default/getPackagesV3
	if arch == "amd64" {
		arch = "x64"
	} else if arch == "arm64" {
		arch = "aarch64"
	}

	distros, err := LoadDistros()
	if err != nil {
		panic(fmt.Errorf("unable to load distros\n%w", err))
	}

	found := false
	for _, distro := range distros {
		if distro == d {
			found = true
			break
		}
	}

	if !found {
		panic(fmt.Errorf("invalid distro %q, valid: %s", d, distros))
	}

	// type i.e. jre or jdk
	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	// major version number, like 8 or 11
	s, ok := inputs["version"]
	if !ok {
		panic(fmt.Errorf("version must be specified"))
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		panic(fmt.Errorf("version cannot be parsed\n%w", err))
	}

	versions := LoadPackages(d, t, v, arch)

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(fmt.Errorf("unable to get latest version\n%w", err))
	}

	outputs, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion, nil)
	if err != nil {
		panic(fmt.Errorf("unable to create outputs\n%w", err))
	}

	if latestVersion.Major() == 8 {
		// Java 8 uses `1.8.0` and `updateXX` in the CPE, instead of 8.0.x
		//
		// This adjusts the update job to set the CPE in this way instead
		// of using the standard version format
		outputs["cpe"] = fmt.Sprintf("update%d", latestVersion.Patch())
	}

	outputs.Write()
}

func LoadPackages(d string, t string, v int, a string) actions.Versions {
	uri := fmt.Sprintf(
		"https://api.foojay.io/disco/v3.0/packages?"+
			"distro=%s&"+
			"architecture=%s&"+
			"archive_type=tar.gz&"+
			"package_type=%s&"+
			"operating_system=linux&"+
			"lib_c_type=glibc&"+
			"javafx_bundled=false&"+
			"version=%d..%%3C%d", // 11..<12
		url.PathEscape(d), a, t, v, v+1)

	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	request.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(request)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw PackagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode package payload\n%w", err))
	}

	re := regexp.MustCompile(`(\d+\.\d+\.\d+\.\d+|\d+\.\d+\.\d+\+\d+|\d+\.\d+\.\d+|\d+)\+?.*`)

	versions := make(actions.Versions)
	for _, result := range raw.Result {
		if !result.LatestBuild {
			continue
		}

		if ver := re.FindStringSubmatch(result.JavaVersion); ver != nil {
			version, err := actions.NormalizeVersion(ver[1])
			if err != nil {
				panic(fmt.Errorf("unable to normalize version %s\n%w", version, err))
			}
			versions[version] = LoadDownloadURI(result.Links.URI)
		} else {
			panic(fmt.Errorf(result.JavaVersion, "failed to parse"))
		}
	}

	return versions
}

func LoadDownloadURI(uri string) string {
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	request.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(request)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw DownloadResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode download payload\n%w", err))
	}

	if len(raw.Result) != 1 {
		panic(fmt.Errorf("expected 1 result but got\n%v", raw))
	}

	return raw.Result[0].DirectDownloadURI
}

func LoadDistros() ([]string, error) {
	uri := "https://api.foojay.io/disco/v3.0/distributions?include_versions=false&include_synonyms=false"
	request, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	request.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(request)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw DistroResponse
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode download payload\n%w", err))
	}

	distros := []string{}
	for _, distro := range raw.Result {
		distros = append(distros, distro.APIParameter)
	}

	return distros, nil
}

type PackagesResponse struct {
	Result []struct {
		JavaVersion string `json:"java_version"`
		Links       struct {
			URI string `json:"pkg_info_uri"`
		}
		LatestBuild bool `json:"latest_build_available"`
	}
	Message string
}

type DownloadResponse struct {
	Result []struct {
		DirectDownloadURI string `json:"direct_download_uri"`
		SignatureURI      string `json:"signature_uri"`
	}
	Message string
}

type DistroResponse struct {
	Result []struct {
		APIParameter string `json:"api_parameter"`
		Maintained   bool   `json:"maintained"`
	}
	Message string
}
