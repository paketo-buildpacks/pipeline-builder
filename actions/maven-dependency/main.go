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
	"encoding/xml"
	"fmt"
	"net/http"
	"os"
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

	a, ok := inputs["artifact_id"]
	if !ok {
		panic(fmt.Errorf("artifact_id must be specified"))
	}

	versionRegex, ok := inputs["version_regex"]
	if !ok {
		fmt.Println(`No version_regex set, using default: ^[\d]+\.[\d]+\.[\d]+/$`)
		versionRegex = `^[\d]+\.[\d]+\.[\d]+$`
	}
	versionPattern := regexp.MustCompile(versionRegex)

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

	versions := make(actions.Versions)
	for _, v := range raw.Versioning.Versions {
		if p := versionPattern.MatchString(v); p {
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

			n, err := actions.NormalizeVersion(v)
			if err != nil {
				panic(err)
			}

			versions[n] = w
		}
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write(os.Stdout)
	}
}

type Metadata struct {
	Versioning Versioning `xml:"versioning"`
}

type Versioning struct {
	Versions []string `xml:"versions>version"`
}
