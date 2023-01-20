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
	"regexp"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	v, ok := inputs["version"]
	if !ok {
		panic(fmt.Errorf("version must be specified"))
	}

	// should be `liberica` (OpenJDK) or `nik` (Native Image)
	p, ok := inputs["product"]
	if !ok {
		panic(fmt.Errorf("product must be specified"))
	}

	uri := fmt.Sprintf("https://api.bell-sw.com/v1/%s/releases"+
		"?arch=x86"+
		"&bitness=64"+
		"&bundle-type=%s"+
		"&os=linux"+
		"&package-type=tar.gz"+
		"&version-feature=%s"+
		"&version-modifier=latest",
		p, t, v)
	if p == "nik" {
		uri = fmt.Sprintf("https://api.bell-sw.com/v1/%s/releases"+
			"?arch=x86"+
			"&bitness=64"+
			"&bundle-type=%s"+
			"&os=linux"+
			"&package-type=tar.gz"+
			"&component-version=liberica%%40%s"+
			"&version-modifier=latest",
			p, t, v)
	}

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw []Release
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	versions := make(actions.Versions)
	additionalOutputs := make(actions.Outputs)

	for _, r := range raw {
		key := fmt.Sprintf("%d.%d.%d-%d", r.FeatureVersion, r.InterimVersion, r.UpdateVersion, r.BuildVersion)
		if p == "nik" {
			if v := actions.ExtendedVersionPattern.FindStringSubmatch(r.Components[0].Version); v != nil {
				s := fmt.Sprintf("%s.%s.%s", v[1], v[2], v[3])
				if v[4] != "" {
					s = fmt.Sprintf("%s-%s", s, v[4])
				}
				key = s

				// Use NIK version for CPE/PURL
				re := regexp.MustCompile(`\/vm/([\d]+\.[\d]+\.[\d]+\.?[\d]?)\/`)
				matches := re.FindStringSubmatch(r.DownloadURL)
				if matches == nil || len(matches) != 2 {
					panic(fmt.Errorf("unable to parse NIK version: %s", matches))
				}
				additionalOutputs["purl"] = matches[1]
			}
		}
		versions[key] = r.DownloadURL
	}

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(fmt.Errorf("unable to get latest version\n%w", err))
	}

	outputs, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion, additionalOutputs)
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

type Release struct {
	FeatureVersion int    `json:"featureVersion"`
	InterimVersion int    `json:"interimVersion"`
	UpdateVersion  int    `json:"updateVersion"`
	BuildVersion   int    `json:"buildVersion"`
	DownloadURL    string `json:"downloadUrl"`
	Components     []struct {
		Version string `json:"version"`
	}
}
