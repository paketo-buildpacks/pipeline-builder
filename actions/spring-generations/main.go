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
	"sort"
	"strings"
	"sync"

	"github.com/pelletier/go-toml"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	u, ok := inputs["uri"]
	if !ok {
		panic(fmt.Errorf("uri must be specified"))
	}

	uri := fmt.Sprintf("%s/projects", u)
	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to GET %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s", uri))
	}

	var raw RawProjectsPayload
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode %s\n%w", uri, err))
	}

	results := make(chan result)
	var wg sync.WaitGroup

	for _, p := range raw.Embedded.Projects {

		wg.Add(1)
		go func(rawProject RawProject) {
			defer wg.Done()

			project := Project{
				Name:   rawProject.Name,
				Slug:   rawProject.Slug,
				Status: rawProject.Status,
			}

			uri := rawProject.Links["generations"].Href
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

			var raw RawGenerationsPayload
			if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
				results <- result{err: fmt.Errorf("unable to decode %s\n%w", uri, err)}
				return
			}

			for _, g := range raw.Embedded.Generations {
				project.Generations = append(project.Generations, Generation{
					Name:       g.Name,
					OSS:        g.OSS,
					Commercial: g.Commercial,
				})
			}

			results <- result{value: project}
		}(p)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	content := Content{}
	for r := range results {
		if r.err != nil {
			panic(fmt.Errorf("unable to get generations\n%w", r.err))
		}
		content.Projects = append(content.Projects, r.value)
	}

	sort.Slice(content.Projects, func(i, j int) bool {
		return content.Projects[i].Slug < content.Projects[j].Slug
	})

	b, err := toml.Marshal(content)
	if err != nil {
		panic(fmt.Errorf("unable to marshal content\n%w", err))
	}

	fmt.Printf("::set-output name=content::%s\n", strings.ReplaceAll(string(b), "\n", "%0A"))
}

type result struct {
	err   error
	value Project
}

type Content struct {
	Projects []Project
}

type Project struct {
	Name        string
	Slug        string
	Status      string
	Generations []Generation
}

type Generation struct {
	Name       string
	OSS        string
	Commercial string
}

type RawLink struct {
	Href string
}

type RawProjectsPayload struct {
	Embedded RawProjectsEmbedded `json:"_embedded"`
}

type RawProjectsEmbedded struct {
	Projects []RawProject
}

type RawProject struct {
	Name   string
	Slug   string
	Status string
	Links  map[string]RawLink `json:"_links"`
}

type RawGenerationsPayload struct {
	Embedded RawGenerationsEmbedded `json:"_embedded"`
}

type RawGenerationsEmbedded struct {
	Generations []RawGeneration
}

type RawGeneration struct {
	Name       string
	OSS        string `json:"ossSupportEndDate"`
	Commercial string `json:"commercialSupportEndDate"`
}
