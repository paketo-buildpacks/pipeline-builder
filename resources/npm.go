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

package resources

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type NPM struct{}

func (n NPM) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, n)
}

func (n NPM) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, n)
}

func (NPM) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (NPM) Versions(source map[string]interface{}) (map[Version]string, error) {
	p, ok := source["package"]
	if !ok {
		return nil, fmt.Errorf("package must be specified")
	}

	uri := fmt.Sprintf("https://registry.npmjs.org/%s", p)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		Versions map[string]struct {
			Dist struct {
				Tarball string `json:"tarball"`
			} `json:"dist"`
		} `json:"versions"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	versions := make(map[Version]string, len(raw.Versions))
	for k, v := range raw.Versions {
		versions[Version(k)] = v.Dist.Tarball
	}

	return versions, nil
}
