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

	// pull latest version
	release, _, err := gh.Repositories.GetLatestRelease(context.Background(), ORG, REPO)
	if err != nil {
		panic(fmt.Errorf("unable to list existing tags for %s/%s\n%w", ORG, REPO, err))
	}

	normalVersion, err := actions.NormalizeVersion(*release.TagName)
	if err != nil {
		panic(err)
	}

	versions[normalVersion] = fmt.Sprintf("https://static.rust-lang.org/dist/"+
		"rust-%s-%s.tar.gz", normalVersion, target)

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
