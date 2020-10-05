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
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	v, ok := inputs["version"]
	if !ok {
		panic(fmt.Errorf("version must be specified"))
	}

	i, ok := inputs["implementation"]
	if !ok {
		panic(fmt.Errorf("implementation must be specified"))
	}

	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	uri := fmt.Sprintf("https://api.adoptopenjdk.net/v3/assets/version/%s"+
		"?architecture=x64"+
		"&heap_size=normal"+
		"&image_type=%s"+
		"&jvm_impl=%s"+
		"&os=linux"+
		"&release_type=ga"+
		"&vendor=adoptopenjdk",
		url.PathEscape(v), t, i)

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw []Asset
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	versions := make(actions.Versions, len(raw))
	for _, r := range raw {
		versions[r.VersionData.Semver] = r.Binaries[0].Package.Link
	}

	versions.GetLatest(inputs).Write(os.Stdout)
}

type Asset struct {
	Binaries    []Binary
	VersionData VersionData `json:"version_data"`
}

type Binary struct {
	Package Package
}

type Package struct {
	Link string
}

type VersionData struct {
	Semver string
}
