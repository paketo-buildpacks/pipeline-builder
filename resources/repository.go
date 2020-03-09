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
	"fmt"
	"net/http"
	"regexp"

	"gopkg.in/yaml.v3"
)

type Repository struct{}

func (Repository) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (r Repository) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri, ok := source["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("uri must be specified")
	}

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	var raw map[string]string
	if err := yaml.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload: %w", err)
	}

	cp := regexp.MustCompile("^([\\d]+)\\.([\\d]+)\\.([\\d]+)_?(.*)$")
	versions := make(map[Version]string, len(raw))
	for k, v := range raw {
		if p := cp.FindStringSubmatch(k); p != nil {
			ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
			if p[4] != "" {
				ref = fmt.Sprintf("%s-%s", ref, p[4])
			}

			versions[Version(ref)] = v
		}
	}

	return versions, nil
}
