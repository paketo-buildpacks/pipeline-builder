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

package resources

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Masterminds/semver/v3"
)

type Version string

type version struct {
	Ref string `json:"ref,omitempty"`
}

func (v Version) MarshalJSON() ([]byte, error) {
	r := version{Ref: string(v)}

	return json.Marshal(r)
}

func (v *Version) UnmarshalJSON(data []byte) error {
	var r version
	if err := json.Unmarshal(data, &r); err != nil {
		return fmt.Errorf("cannot unmarshal check source\n%w", err)
	}

	*v = Version(r.Ref)
	return nil
}

type Metadata map[string]string

type metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (m Metadata) MarshalJSON() ([]byte, error) {
	var e []metadata

	var n []string
	for k, _ := range m {
		n = append(n, k)
	}
	sort.Strings(n)

	for _, k := range n {
		e = append(e, metadata{Name: k, Value: m[k]})
	}

	return json.Marshal(e)
}

type CheckRequest struct {
	Source  map[string]interface{} `json:"source"`
	Version Version                `json:"version"`
}

type CheckResult []Version

type InRequest struct {
	Source  map[string]interface{} `json:"source"`
	Version Version                `json:"version"`
	Params  map[string]interface{} `json:"params"`
}

type InResult struct {
	Version  Version  `json:"version,omitempty"`
	Metadata Metadata `json:"metadata,omitempty"`
}

type OutRequest struct {
	Source map[string]interface{} `json:"source"`
	Params map[string]interface{} `json:"params"`
}

type OutResult struct {
	Version  Version    `json:"version,omitempty"`
	Metadata []Metadata `json:"metadata,omitempty"`
}

type Resource interface {
	Out(request OutRequest, destination string) (OutResult, error)
	Versions(source map[string]interface{}) (map[Version]string, error)
}

func Check(resource Resource) {
	var request CheckRequest
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		log.Fatal(err)
	}

	versions, err := resource.Versions(request.Source)
	if err != nil {
		log.Fatal(err)
	}

	var vp *regexp.Regexp
	if s, ok := request.Source["version_pattern"].(string); ok {
		vp, err = regexp.Compile(s)
		if err != nil {
			log.Fatal(err)
		}
	}

	pr := true
	if b, ok := request.Source["pre_release"].(bool); ok {
		pr = b
	}

	var sv []*semver.Version
	for k, _ := range versions {
		if vp == nil || vp.MatchString(string(k)) {
			v, err := semver.NewVersion(string(k))
			if err != nil {
				log.Fatal(err)
			}

			if v.Prerelease() == "" || pr {
				sv = append(sv, v)
			}
		}
	}

	sort.Slice(sv, func(i, j int) bool {
		return sv[i].LessThan(sv[j])
	})

	if request.Version == "" {
		sv = sv[len(sv)-1:]
	} else {
		since, err := semver.NewVersion(string(request.Version))
		if err != nil {
			log.Fatal(err)
		}

		for i, v := range sv {
			if !v.LessThan(since) {
				sv = sv[i:]
				break
			}
		}
	}

	result := CheckResult{}
	for _, v := range sv {
		result = append(result, Version(v.Original()))
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}

func In(resource Resource) {
	var request InRequest
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		log.Fatal(err)
	}

	versions, err := resource.Versions(request.Source)
	if err != nil {
		log.Fatal(err)
	}

	uri := versions[request.Version]

	file := filepath.Join(os.Args[1], "version")
	if err := ioutil.WriteFile(file, []byte(request.Version), 0644); err != nil {
		log.Fatal(err)
	}

	result := InResult{Version: request.Version}

	if uri != "" {
		sha256, err := DownloadArtifact(uri, request.Version, os.Args[1])
		if err != nil {
			log.Fatal(err)
		}

		result.Metadata = Metadata{
			"uri":    uri,
			"sha256": sha256,
		}
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}

func Out(resource Resource) {
	var request OutRequest
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		log.Fatal(err)
	}

	result, err := resource.Out(request, os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatal(err)
	}
}
