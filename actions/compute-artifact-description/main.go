/*
 * Copyright 2018-2022 the original author or authors.
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
	_ "embed"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v43/github"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/paketo-buildpacks/pipeline-builder/drafts"
	"golang.org/x/oauth2"
)

func main() {
	inputs := actions.NewInputs()

	if _, found := inputs["package"]; !found {
		panic(fmt.Errorf("unable to read package, package must be set"))
	}

	if _, found := inputs["version"]; !found {
		panic(fmt.Errorf("unable to read version, version must be set"))
	}

	imgUri := fmt.Sprintf("%s:%s", inputs["package"], inputs["version"])
	fmt.Println("Loading image URI:", imgUri)

	var c *http.Client
	if s, ok := inputs["github_token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	loader := drafts.GithubBuildModuleLoader{
		GithubClient: gh,
		RegexMappers: parseMappers(inputs),
	}

	mainBp, err := loader.LoadBuildpack(imgUri)
	if err != nil {
		panic(err)
	}

	depsMap := map[string]bool{}
	if len(mainBp.OrderGroups) > 0 {
		pkg, err := loader.LoadPackages(imgUri)
		if err != nil {
			panic(err)
		}

		bps := []drafts.Buildpack{}
		for _, depBp := range pkg.Dependencies {
			bp, err := loader.LoadBuildpack(depBp.URI)
			if err != nil {
				panic(fmt.Errorf("unable to load %s\n%w", depBp.URI, err))
			}
			bps = append(bps, bp)
		}

		for _, bp := range bps {
			for _, dep := range bp.Dependencies {
				depsMap[fmt.Sprintf("%s %s", dep.Name, dep.Version)] = true
			}
		}
	} else {
		for _, dep := range mainBp.Dependencies {
			depsMap[fmt.Sprintf("%s %s", dep.Name, dep.Version)] = true
		}
	}

	deps := []string{}
	for dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Strings(deps)

	actions.Outputs{
		"artifact-reference-description": strings.Join(deps, "\n"),
	}.Write()
}

func parseMappers(inputs actions.Inputs) []string {
	mappers := []string{}

	for key, val := range inputs {
		if strings.HasPrefix(key, "mapper_") {
			mappers = append(mappers, val)
		}
	}

	return mappers
}
