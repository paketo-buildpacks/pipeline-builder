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
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"sync"

	"github.com/BurntSushi/toml"
)

type SpringGenerations struct{}

func (SpringGenerations) Check(request CheckRequest) (CheckResult, error) {
	u, ok := request.Source["uri"].(string)
	if !ok {
		return CheckResult{}, fmt.Errorf("uri must be specified")
	}

	resp, err := http.Head(u)
	if err != nil {
		return CheckResult{}, fmt.Errorf("unable to HEAD %s\n%w", u, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return CheckResult{}, fmt.Errorf("unable to download %s", u)
	}

	return CheckResult{Version(resp.Header.Get("Date"))}, nil
}

func (SpringGenerations) In(request InRequest, destination string) (InResult, error) {
	u, ok := request.Source["uri"].(string)
	if !ok {
		return InResult{}, fmt.Errorf("uri must be specified")
	}

	uri := fmt.Sprintf("%s/projects", u)
	resp, err := http.Get(uri)
	if err != nil {
		return InResult{}, fmt.Errorf("unable to GET %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return InResult{}, fmt.Errorf("unable to download %s", uri)
	}

	raw := struct {
		Embedded map[string][]json.RawMessage `json:"_embedded"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return InResult{}, fmt.Errorf("unable to decode %s\n%w", uri, err)
	}

	results := make(chan result)
	var wg sync.WaitGroup

	for _, raw := range raw.Embedded["projects"] {

		wg.Add(1)
		go func(raw json.RawMessage) {
			defer wg.Done()

			var project Project
			if err := json.Unmarshal(raw, &project); err != nil {
				results <- result{err: fmt.Errorf("unable to decode %s\n%w", uri, err)}
				return
			}

			uri := fmt.Sprintf("%s/projects/%s/generations", u, project.Slug)
			resp, err := http.Get(uri)
			if err != nil {
				results <- result{err: fmt.Errorf("unable to GET %s\n%w", uri, err)}
				return
			}
			defer resp.Body.Close()

			if resp.StatusCode != 200 {
				results <- result{err: fmt.Errorf("unable to download %s", uri)}
				return
			}

			raw2 := struct {
				Embedded map[string][]json.RawMessage `json:"_embedded"`
			}{}
			if err := json.NewDecoder(resp.Body).Decode(&raw2); err != nil {
				results <- result{err: fmt.Errorf("unable to decode %s\n%w", uri, err)}
				return
			}

			for _, raw := range raw2.Embedded["generations"] {
				var generation Generation
				if err := json.Unmarshal(raw, &generation); err != nil {
					results <- result{err: fmt.Errorf("unable to decode %s\n%w", uri, err)}
					return
				}

				project.Generations = append(project.Generations, generation)
			}

			results <- result{value: project}
		}(raw)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	content := &Content{}
	for r := range results {
		if r.err != nil {
			return InResult{}, fmt.Errorf("unable to get generations\n%w", r.err)
		}
		content.Projects = append(content.Projects, r.value)
	}

	sort.Slice(content.Projects, func(i, j int) bool {
		return content.Projects[i].Slug < content.Projects[j].Slug
	})

	file := filepath.Join(destination, "spring-generations.toml")
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return InResult{}, fmt.Errorf("unable to open file %s\n%w", file, err)
	}
	defer f.Close()

	s := sha256.New()
	out := io.MultiWriter(f, s)

	if err := toml.NewEncoder(out).Encode(content); err != nil {
		return InResult{}, err
	}
	hash := hex.EncodeToString(s.Sum(nil))

	file = filepath.Join(destination, "sha256")
	if err := ioutil.WriteFile(file, []byte(hash), 0644); err != nil {
		return InResult{}, fmt.Errorf("unable to write %s\n%w", file, err)
	}

	result := InResult{
		Version:  request.Version,
		Metadata: Metadata{"sha256": hash},
	}

	return result, nil
}

func (SpringGenerations) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

type result struct {
	err   error
	value Project
}

type Content struct {
	Projects []Project `toml:"projects"`
}

type Project struct {
	Name        string       `json:"name" toml:"name"`
	Slug        string       `json:"slug" toml:"slug"`
	Status      string       `json:"status" toml:"status"`
	Generations []Generation `toml:"generations"`
}

type Generation struct {
	Name       string `json:"name" toml:"name"`
	OSS        string `json:"ossSupportEndDate" toml:"oss"`
	Commercial string `json:"commercialSupportEndDate" toml:"commercial"`
}
