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
	"bufio"
	"fmt"
	"net/http"

	actions "github.com/paketo-buildpacks/pipeline-builder/v2/actions"
)

func main() {
	inputs := actions.NewInputs()

	u, ok := inputs["username"]
	if !ok {
		panic(fmt.Errorf("username must be specified"))
	}

	p, ok := inputs["password"]
	if !ok {
		panic(fmt.Errorf("password must be specified"))
	}

	uri := "https://get.aquasec.com/scanner-releases.txt"

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	versions := make(actions.Versions)

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		v := scanner.Text()
		versions[v] = fmt.Sprintf("https://download.aquasec.com/scanner/%s/scannercli", v)
	}
	if err := scanner.Err(); err != nil {
		panic(fmt.Errorf("unable to read index\n%w", err))
	}

	if o, err := versions.GetLatest(inputs, actions.WithBasicAuth(u, p)); err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
