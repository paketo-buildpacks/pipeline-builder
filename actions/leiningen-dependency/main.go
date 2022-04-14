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

	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := make(actions.Versions)

	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, r, err := gh.Repositories.ListReleases(context.Background(), "technomancy", "leiningen", opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for technomancy/leiningen\n%w", err))
		}

		for _, r := range rel {
			versions[*r.TagName] = fmt.Sprintf("https://raw.githubusercontent.com/technomancy/leiningen/%s/bin/lein", *r.TagName)
		}

		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write(os.Stdout)
	}
}
