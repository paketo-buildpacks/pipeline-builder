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
	_ "embed"

	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/paketo-buildpacks/libpak/effect"
	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"github.com/paketo-buildpacks/pipeline-builder/drafts"
)

//go:embed "draft.template"
var templateContents string

func main() {
	inputs := actions.NewInputs()

	drafter := drafts.Drafter{Loader: drafts.RegistryBuildpackLoader{}}
	payload, err := drafter.CreatePayload(inputs, ".")
	if err != nil {
		panic(err)
	}

	err = drafter.BuildAndWriteReleaseToFileDraftFromTemplate(
		filepath.Join(".", "body"), templateContents, payload)
	if err != nil {
		panic(err)
	}

	name := payload.Release.Name
	if payload.PrimaryBuildpack.Info.Name != "" {
		name = fmt.Sprintf("%s %s", payload.PrimaryBuildpack.Info.Name, payload.Release.Name)
	}

	err = ioutil.WriteFile(filepath.Join(".", "name"), []byte(name), 0644)
	if err != nil {
		panic(err)
	}

	execution := effect.Execution{
		Command: "gh",
		Args: []string{
			"api",
			"--method", "PATCH",
			fmt.Sprintf("/repos/:owner/:repo/releases/%s", payload.Release.ID),
			"--field", fmt.Sprintf("tag_name=%s", payload.Release.Tag),
			"--field", "name=@./name",
			"--field", "body=@./body",
		},
	}
	if _, dryRun := inputs["dry_run"]; dryRun {
		bits, err := ioutil.ReadFile(filepath.Join(".", "body"))
		if err != nil {
			panic(err)
		}

		fmt.Println("Title:", name)
		fmt.Println("Body:", string(bits))
		fmt.Println("Would execute:", execution)
	} else {
		err = effect.NewExecutor().Execute(execution)
		if err != nil {
			panic(fmt.Errorf("unable to execute %s\n%w", execution, err))
		}
	}
}
