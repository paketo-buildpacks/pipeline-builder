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
	"net/url"
)

type AdoptOpenJDK struct{}

func (a AdoptOpenJDK) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, a)
}

func (a AdoptOpenJDK) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, a)
}

func (AdoptOpenJDK) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (AdoptOpenJDK) Versions(source map[string]interface{}) (map[Version]string, error) {
	v, ok := source["version"].(string)
	if !ok {
		return nil, fmt.Errorf("version must be specified")
	}

	i, ok := source["implementation"].(string)
	if !ok {
		return nil, fmt.Errorf("implementation must be specified")
	}

	t, ok := source["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type must be specified")
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
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := make([]struct {
		Binaries []struct {
			Package struct {
				Link string `json:"link"`
			} `json:"package"`
		} `json:"binaries"`
		VersionData struct {
			SemVer string `json:"semver"`
		} `json:"version_data"`
	}, 1)
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	versions := make(map[Version]string, len(raw))
	for _, r := range raw {
		versions[Version(r.VersionData.SemVer)] = r.Binaries[0].Package.Link
	}

	return versions, nil
}
