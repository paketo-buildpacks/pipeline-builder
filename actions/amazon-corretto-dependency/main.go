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
	"regexp"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	repo, ok := inputs["repository"]
	if !ok {
		panic(fmt.Errorf("repository must be specified"))
	}

	var (
		err error
		g   *regexp.Regexp
	)
	if s, ok := inputs["glob"]; ok {
		g, err = regexp.Compile(s)
		if err != nil {
			panic(fmt.Errorf("unable to compile %s as a regexp\n%w", s, err))
		}
	} else {
		g = regexp.MustCompile(".+")
	}

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := make(actions.Versions)
	sources := make(map[string]string)

	md := regexp.MustCompile(`(?U)\[(.+)]\((.+)\)`)

	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, rsp, err := gh.Repositories.ListReleases(context.Background(), "corretto", repo, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for %s/%s\n%w", "corretto", repo, err))
		}

		for _, r := range rel {
			if l := md.FindAllStringSubmatch(*r.Body, -1); l != nil {
				for _, p := range l {
					if g.MatchString(p[1]) {
						if v := actions.ExtendedVersionPattern.FindStringSubmatch(*r.TagName); v != nil {
							var s string

							if v[1] == "8" {
								s = fmt.Sprintf("%s.0.%s-%s", v[1], v[2], v[3])
								if v[4] != "" {
									s = fmt.Sprintf("%s-%s", s, v[4])
								}
							} else {
								s = fmt.Sprintf("%s.%s.%s", v[1], v[2], v[3])
								if v[4] != "" {
									s = fmt.Sprintf("%s-%s", s, v[4])
								}
							}

							versions[s] = p[2]
							sources[s] = fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.tar.gz", "corretto", repo, *r.TagName)
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

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(fmt.Errorf("unable to get latest version\n%w", err))
	}

	latestSource := actions.Outputs{}
	if sources != nil {
		latestSource["source"] = sources[latestVersion.Original()]
	}

	outputs, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion, latestSource)
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
