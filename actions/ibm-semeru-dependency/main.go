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
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/google/go-github/v43/github"
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

				for _, a := range r.Assets {
					if g.MatchString(*a.Name) {
						// get the matched asset's associated json file
						versionJSON := *a.BrowserDownloadURL + ".json"
						// Decode JSON so we can extract 'version.semver'
						json := getReleaseJSON(versionJSON)
						// Record semver against the asset's tar.gz URL
						versions[strings.ReplaceAll(json.VersionData.Semver, "+", "-")] = *a.BrowserDownloadURL

						break
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

	url := versions[latestVersion.Original()]
	outputs, err := actions.NewOutputs(url, latestVersion, nil)
	if err != nil {
		panic(fmt.Errorf("unable to create outputs\n%w", err))
	}

	// IBM Semeru uses the OpenJ9 version in the CPE, ex: 0.29.1
	//
	// This adjusts the update job to set the CPE in this way instead
	// of using the standard version format
	re = regexp.MustCompile(`openj9-([\d]+\.[\d]+\.[\d]+).*?\/`)
	matches := re.FindStringSubmatch(url)
	if matches == nil || len(matches) != 2 {
		panic(fmt.Errorf("unable to parse OpenJ9 version: %s", matches))
	}
	outputs["cpe"] = matches[1]

	outputs.Write()
}

func getReleaseJSON(jsonURI string) ReleaseJSON {
	resp, err := http.Get(jsonURI)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", jsonURI, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", jsonURI, resp.StatusCode))
	}
	// Decode JSON file to get semver data
	var raw ReleaseJSON
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	return raw

}

type ReleaseJSON struct {
	VersionData VersionData `json:"version"`
}

type VersionData struct {
	Semver string
}
