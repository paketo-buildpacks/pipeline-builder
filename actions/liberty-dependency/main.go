/*
 * Copyright 2018-2022 the original author or authors.
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
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	u, ok := inputs["uri"]
	if !ok {
		panic(fmt.Errorf("uri must be specified"))
	}

	g, ok := inputs["group_id"]
	if !ok {
		panic(fmt.Errorf("group_id must be specified"))
	}
	source_available := false
	if strings.Contains(g, "openliberty"){
		source_available = true
	}

	a, ok := inputs["artifact_id"]
	if !ok {
		panic(fmt.Errorf("artifact_id must be specified"))
	}

	uri := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", u, strings.ReplaceAll(g, ".", "/"), a)

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw Metadata
	if err := xml.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	originalVersions := map[string]string{}
	versions := make(actions.Versions)
	sources := make(map[string]string)

	for _, v := range raw.Versioning.Versions {
		w := fmt.Sprintf("%s/%s/%s/%s", u, strings.ReplaceAll(g, ".", "/"), a, v)
		w = fmt.Sprintf("%s/%s-%s", w, a, v)

		if s, ok := inputs["classifier"]; ok {
			w = fmt.Sprintf("%s-%s", w, s)
		}
		if s, ok := inputs["packaging"]; ok {
			w = fmt.Sprintf("%s.%s", w, s)
		} else {
			w = fmt.Sprintf("%s.jar", w)
		}

		n, err := NormalizeVersion(v)
		if err != nil {
			panic(err)
		}
		if source_available {
			sources[v] = fmt.Sprintf("https://github.com/OpenLiberty/open-liberty/archive/refs/tags/gm-%s.tar.gz", v)
		}
		originalVersions[n] = v
		versions[n] = w
	}

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(fmt.Errorf("unable to get latest version\n%w", err))
	}
	latestSource := actions.Outputs{}
	if len(sources) != 0{
		latestSource["source"] = sources[originalVersions[latestVersion.Original()]]
	}

	outputs, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion, latestSource)
	if err != nil {
		panic(fmt.Errorf("unable to create outputs\n%w", err))
	}

	outputs["cpe"] = originalVersions[latestVersion.Original()]
	outputs["purl"] = originalVersions[latestVersion.Original()]
	outputs.Write()
}

var ExtendedVersionPattern = regexp.MustCompile(`^v?([\d]+)\.?([\d]+)?\.?([\d]+)?[-+.]?(.*)$`)

func NormalizeVersion(raw string) (string, error) {
	if p := ExtendedVersionPattern.FindStringSubmatch(raw); p != nil {
		for i := 1; i < 5; i++ {
			if p[i] == "" {
				p[i] = "0"
			}
		}

		s := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[4])

		return s, nil
	}

	return "", fmt.Errorf("unable to parse %s as a extended version (%s)", raw, ExtendedVersionPattern)
}

type Metadata struct {
	Versioning Versioning `xml:"versioning"`
}

type Versioning struct {
	Versions []string `xml:"versions>version"`
}
