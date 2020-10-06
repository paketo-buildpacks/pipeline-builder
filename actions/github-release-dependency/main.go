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
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	o, ok := inputs["owner"]
	if !ok {
		panic(fmt.Errorf("owner must be specified"))
	}

	r, ok := inputs["repository"]
	if !ok {
		panic(fmt.Errorf("repository must be specified"))
	}

	g := "*"
	if s, ok := inputs["glob"]; ok {
		g = s
	}

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := actions.Versions{
		Contents:        make(map[string]string),
		Inputs:          inputs,
		SemverConverter: actions.MetadataSemverConverter,
	}

	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, rsp, err := gh.Repositories.ListReleases(context.Background(), o, r, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for %s/%s\n%w", o, r, err))
		}

		for _, r := range rel {
			for _, a := range r.Assets {
				if ok, err := filepath.Match(g, *a.Name); err != nil {
					panic(err)
				} else if ok {
					versions.Contents[*r.TagName] = *a.BrowserDownloadURL
					break
				}
			}
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	versions.GetLatest().Write(os.Stdout)
}
