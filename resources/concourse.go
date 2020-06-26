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
	"log"
	"os"
	"sort"
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
	Check(request CheckRequest) (CheckResult, error)
	In(request InRequest, destination string) (InResult, error)
	Out(request OutRequest, source string) (OutResult, error)
}

func Check(resource Resource) {
	var request CheckRequest
	if err := json.NewDecoder(os.Stdin).Decode(&request); err != nil {
		log.Fatal(err)
	}

	result, err := resource.Check(request)
	if err != nil {
		log.Fatal(err)
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

	result, err := resource.In(request, os.Args[1])
	if err != nil {
		log.Fatal(err)
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
