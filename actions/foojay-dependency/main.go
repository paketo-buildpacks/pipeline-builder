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
	"os"
	"regexp"
	"strconv"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

const userAgent = "Paketo_Buildpacks_CI/1.0"

var client *http.Client = http.DefaultClient

func main() {
	inputs := actions.NewInputs()

	// See list of distro's -> https://api.foojay.io/swagger-ui#/default/getDistributionsV2
	d, ok := inputs["distro"]
	if !ok {
		panic(fmt.Errorf("distro must be specified"))
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

	versions := LoadPackages(d, t, v)

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

	outputs.Write(os.Stdout)
}

func LoadPackages(d string, t string, v int) actions.Versions {
	uri := fmt.Sprintf(
		"https://api.foojay.io/disco/v3.0/packages?"+
			"distro=%s&"+
			"architecture=x64&"+
			"archive_type=tar.gz&"+
			"package_type=%s&"+
			"operating_system=linux&"+
			"javafx_bundled=false&"+
			"version=%d..%%3C%d", // 11..<12
		url.PathEscape(d), t, v, v+1)

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

	re := regexp.MustCompile(`(\d+\.\d+\.\d+|\d+)\+?.*`)

	versions := make(actions.Versions)
	for _, result := range raw.Result {
		if ver := re.FindStringSubmatch(result.JavaVersion); ver != nil {
			versions[ver[1]] = LoadDownloadURI(result.Links.URI)
		} else {
			fmt.Println(result.JavaVersion, "failed to parse")
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

type PackagesResponse struct {
	Result []struct {
		JavaVersion string `json:"java_version"`
		Links       struct {
			URI string `json:"pkg_info_uri"`
		}
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
