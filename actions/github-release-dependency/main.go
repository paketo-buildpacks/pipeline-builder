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
	"sort"
	"strconv"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"

	actions "github.com/paketo-buildpacks/pipeline-builder/v2/actions"
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

	var err error
	useCreationTime := false
	if b, ok := inputs["latest_by_creation_time"]; ok {
		useCreationTime, err = strconv.ParseBool(b)
		if err != nil {
			panic(fmt.Errorf("unable to parse %s as a bool\n%w", b, err))
		}
	}

	var g *regexp.Regexp
	if s, ok := inputs["glob"]; ok {
		g, err = regexp.Compile(s)
		if err != nil {
			panic(fmt.Errorf("unable to compile %s as a regexp\n%w", s, err))
		}
	} else {
		g = regexp.MustCompile(".+")
	}

	t := `v?([^v].*)`
	if s, ok := inputs["tag_filter"]; ok {
		t = s
	}

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := make(actions.Versions)

	re, err := regexp.Compile(t)
	if err != nil {
		panic(fmt.Errorf("%s is not a valid regex", t))
	}

	var releases []*github.RepositoryRelease
	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, rsp, err := gh.Repositories.ListReleases(context.Background(), o, r, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for %s/%s\n%w", o, r, err))
		}

		for _, r := range rel {
			if *r.Prerelease {
				continue
			}

			if p := re.FindStringSubmatch(*r.TagName); p != nil {
				n, err := actions.NormalizeVersion(p[1])
				if err != nil {
					panic(err)
				}
				r.TagName = github.String(n)

				releases = append(releases, r)
			}
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	if useCreationTime {
		sort.Slice(releases, func(i, j int) bool {
			return releases[i].CreatedAt.Before(releases[j].CreatedAt.Time)
		})

		releases = releases[len(releases)-1:]
	}

	for _, r := range releases {
		for _, a := range r.Assets {
			if g.MatchString(*a.Name) {
				versions[*r.TagName] = *a.BrowserDownloadURL
				break
			}
		}
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
