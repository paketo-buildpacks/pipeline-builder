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
	"fmt"
	"net/http"
	"os"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
	"gopkg.in/yaml.v3"
)

func main() {
	inputs := actions.NewInputs()

    // same repository_root as documented here: https://github.com/cloudfoundry/java-buildpack/blob/main/docs/extending-repositories.md
	repositoryRoot, ok := inputs["repository_root"]
	if !ok {
		panic(fmt.Errorf("repository_root must be specified"))
	}

	uri := fmt.Sprintf("%s/index.yml", repositoryRoot)
	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	raw := make(map[string]string)
	if err := yaml.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	versions := make(actions.Versions)
	for k, v := range raw {
		versions[k] = v
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write(os.Stdout)
	}
}
