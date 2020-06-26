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

type Gradle struct{}

func (g Gradle) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, g)
}

func (g Gradle) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, g)
}

func (Gradle) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (Gradle) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri := "https://raw.githubusercontent.com/gradle/gradle/master/released-versions.json"

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		FinalReleases []struct {
			Version string `json:"version"`
		} `json:"finalReleases"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	versions := make(map[Version]string, len(raw.FinalReleases))
	for _, v := range raw.FinalReleases {
		versions[Version(v.Version)] = fmt.Sprintf("https://downloads.gradle.org/distributions/gradle-%s-bin.zip", v.Version)
	}

	return versions, nil
}
