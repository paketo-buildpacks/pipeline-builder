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
	"bytes"
	"context"
	_ "embed"
	"net/http"
	"os"
	"strconv"
	"strings"

	"fmt"

	"github.com/google/go-github/v43/github"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/paketo-buildpacks/pipeline-builder/drafts"
	"golang.org/x/oauth2"
)

//go:embed "draft.template"
var templateContents string

func main() {
	inputs := actions.NewInputs()

	var c *http.Client
	if s, ok := inputs["github_token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	drafter := drafts.Drafter{
		Loader: drafts.GithubBuildpackLoader{
			GithubClient: gh,
			RegexMappers: parseMappers(inputs),
		},
	}
        bpPath := "."
        if val, ok := inputs["file_path"]; ok {
	    bpPath = val
        }
	payload, err := drafter.CreatePayload(inputs, bpPath)
	if err != nil {
		panic(err)
	}

	buf := &bytes.Buffer{}
	err = drafter.BuildAndWriteReleaseDraftFromTemplate(buf, templateContents, payload)
	if err != nil {
		panic(err)
	}
	body := buf.String()

	name := payload.Release.Name
	if payload.PrimaryBuildpack.Info.Name != "" {
		name = fmt.Sprintf("%s %s", payload.PrimaryBuildpack.Info.Name, payload.Release.Name)
	}

	fullRepo, found := os.LookupEnv("GITHUB_REPOSITORY")
	if !found {
		panic(fmt.Errorf("unable to find GITHUB_REPOSITORY"))
	}

	owner := strings.SplitN(fullRepo, "/", 2)[0]
	repo := strings.SplitN(fullRepo, "/", 2)[1]
	releaseId, err := strconv.ParseInt(payload.Release.ID, 10, 32)
	if err != nil {
		panic(fmt.Errorf("unable to parse %s\n%w", payload.Release.ID, err))
	}

	repoRelease := github.RepositoryRelease{
		TagName: &payload.Release.Tag,
		Name:    &name,
		Body:    &body,
	}

	if _, dryRun := inputs["dry_run"]; dryRun {
		fmt.Println("Title:", name)
		fmt.Println("Body:", body)
		fmt.Println("Would execute EditRelease with:")
		fmt.Println("    ", owner)
		fmt.Println("    ", repo)
		fmt.Println("    ", releaseId)
		fmt.Println("    ", repoRelease)
	} else {
		_, _, err := gh.Repositories.EditRelease(
			context.Background(),
			owner,
			repo,
			releaseId,
			&repoRelease)
		if err != nil {
			panic(fmt.Errorf("unable to execute EditRelease %s/%s/%d with %q\n%w", owner, repo, releaseId, repoRelease, err))
		}
	}
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
