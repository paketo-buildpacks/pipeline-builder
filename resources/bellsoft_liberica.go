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

type BellSoftLiberica struct{}

func (BellSoftLiberica) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (b BellSoftLiberica) Versions(source map[string]interface{}) (map[Version]string, error) {
	t, ok := source["type"]
	if !ok {
		return nil, fmt.Errorf("type must be specified")
	}

	v, ok := source["version"]
	if !ok {
		return nil, fmt.Errorf("version must be specified")
	}

	uri := fmt.Sprintf("https://api.bell-sw.com/v1/liberica/releases"+
		"?arch=x86"+
		"&bitness=64"+
		"&bundle-type=%s"+
		"&os=linux"+
		"&package-type=tar.gz"+
		"&version-feature=%s"+
		"&version-modifier=latest",
		t, v)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	var raw []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	versions := make(map[Version]string, len(raw))
	for _, r := range raw {
		v := fmt.Sprintf("%.0f.%.0f.%.0f-%.0f",
			r["featureVersion"].(float64),
			r["interimVersion"].(float64),
			r["updateVersion"].(float64),
			r["buildVersion"].(float64))
		versions[Version(v)] = r["downloadUrl"].(string)
	}

	return versions, nil
}
