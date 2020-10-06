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

		versions[actions.NormalizeVersionWithMetadata(v)] = w
	}

	versions.GetLatest().Write(os.Stdout)
}

type Metadata struct {
	Versioning Versioning `xml:"versioning"`
}

type Versioning struct {
	Versions []string `xml:"versions>version"`
}

/**
<?xml version="1.0" encoding="UTF-8"?>
<metadata modelVersion="1.1.0">
  <groupId>org.springframework.experimental</groupId>
  <artifactId>spring-graalvm-native</artifactId>
  <version>0.8.1</version>
  <versioning>
    <latest>0.8.1</latest>
    <release>0.8.1</release>
    <versions>
      <version>0.7.0</version>
      <version>0.7.1</version>
      <version>0.8.0-RC1</version>
      <version>0.8.0</version>
      <version>0.8.1-RC1</version>
      <version>0.8.1</version>
    </versions>
    <lastUpdated>20200925075534</lastUpdated>
  </versioning>
</metadata>

*/
