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
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"golang.org/x/oauth2"
)

const (
	ORG  = "clojure"
	REPO = "brew-install"
)

func main() {
	inputs := actions.NewInputs()

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	versions := make(actions.Versions)

	// fetch all the clojure-tools versions from tags
	opt := &github.ListOptions{PerPage: 100}
	for {
		tags, rsp, err := gh.Repositories.ListTags(context.Background(), ORG, REPO, opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing tags for %s/%s\n%w", ORG, REPO, err))
		}

		for _, t := range tags {
			normalVersion, err := actions.NormalizeVersion(*t.Name)
			if err != nil {
				panic(err)
			}

			versions[normalVersion] = fmt.Sprintf("https://download.clojure.org/install/linux-install-%s.sh", *t.Name)
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	// pick the most recent, this gives us the most recent major.minor.patch
	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		version := o["version"]

		uri := fmt.Sprintf("https://raw.githubusercontent.com/clojure/brew-install/%s/stable.properties",
			version)

		resp, err := http.Get(uri)
		if err != nil {
			panic(fmt.Errorf("unable to get %s\n%w", uri, err))
		}
		defer resp.Body.Close()

		if resp.StatusCode != 200 {
			panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
		}

		b, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		origVersion := strings.Split(string(b), " ")[0]
		normalVersion, err := actions.NormalizeVersion(origVersion)
		if err != nil {
			panic(err)
		}

		versions = make(actions.Versions)
		versions[normalVersion] = fmt.Sprintf("https://download.clojure.org/install/linux-install-%s.sh", origVersion)

		if o, err := versions.GetLatest(inputs); err != nil {
			panic(err)
		} else {
			o.Write(os.Stdout)
		}
	}
}
