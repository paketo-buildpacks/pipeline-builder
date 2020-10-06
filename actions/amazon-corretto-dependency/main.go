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
	"regexp"

	"github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

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

	versions := make(actions.Versions)

	md := regexp.MustCompile(`(?U)\[(.+)]\((.+)\)`)

	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, rsp, err := gh.Repositories.ListReleases(context.Background(), "corretto", r, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for %s/%s\n%w", "corretto", r, err))
		}

		for _, r := range rel {
			if l := md.FindAllStringSubmatch(*r.Body, -1); l != nil {
				for _, p := range l {
					if ok, err := filepath.Match(g, p[1]); err != nil {
						panic(err)
					} else if ok {
						if v := actions.ExtendedVersionPattern.FindStringSubmatch(*r.TagName); v != nil {
							var s string

							if v[1] == "8" {
								s = fmt.Sprintf("%s.0.%s+%s", v[1], v[2], v[3])
								if v[4] != "" {
									s = fmt.Sprintf("%s-%s", s, v[4])
								}
							} else {
								s = fmt.Sprintf("%s.%s.%s", v[1], v[2], v[3])
								if v[4] != "" {
									s = fmt.Sprintf("%s+%s", s, v[4])
								}
							}

							versions[s] = p[2]
						}
					}
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
