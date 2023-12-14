/*
 * Copyright 2018-2023 the original author or authors.
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
	"strings"

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

const (
	ORG  = "rust-lang"
	REPO = "rust"
)

func main() {
	inputs := actions.NewInputs()

	target, ok := inputs["target"]
	if !ok {
		panic(fmt.Errorf("target must be specified"))
	}

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := make(actions.Versions)
	sources := actions.Outputs{}

	// pull latest version tag
	opt := &github.ListOptions{PerPage: 100}
	for {
		tags, rsp, err := gh.Repositories.ListTags(context.Background(), ORG, REPO, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing tags for %s/%s\n%w", ORG, REPO, err))
		}

		for _, t := range tags {
			normalVersion, err := actions.NormalizeVersion(*t.Name)
			if err != nil && strings.HasPrefix(err.Error(), "unable to parse") && strings.Contains(err.Error(), "as a extended version") {
				fmt.Println("Skipping unrecognized tag", *t.Name, "as it does not match the expected pattern")
				continue
			} else if err != nil {
				panic(fmt.Errorf("unable to normalize version [%s]\n%w", *t.Name, err))
			}

			versions[normalVersion] = fmt.Sprintf("https://static.rust-lang.org/dist/"+"rust-%s-%s.tar.gz", *t.Name, target)
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
	fmt.Println("fetch latest version as:", latestVersion)

	sources["source"] = fmt.Sprintf("https://static.rust-lang.org/dist/rustc-%s-src.tar.gz", latestVersion.Original())

	o, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion, sources)
	if err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
