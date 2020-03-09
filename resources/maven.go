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
	"encoding/xml"
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Maven struct{}

func (Maven) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (m Maven) Versions(source map[string]interface{}) (map[Version]string, error) {
	u, ok := source["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("uri must be specified")
	}

	g, ok := source["group_id"].(string)
	if !ok {
		return nil, fmt.Errorf("group_id must be specified")
	}

	a, ok := source["artifact_id"].(string)
	if !ok {
		return nil, fmt.Errorf("artifact_id must be specified")
	}

	uri := fmt.Sprintf("%s/%s/%s/maven-metadata.xml", u, strings.ReplaceAll(g, ".", "/"), a)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s: %w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		Versions []string `xml:"versioning>versions>version"`
	}{}
	if err := xml.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload: %w", err)
	}

	cp := regexp.MustCompile("^([\\d]+)\\.([\\d]+)\\.([\\d]+)[.-]?(.*)")
	versions := make(map[Version]string, len(raw.Versions))
	for _, v := range raw.Versions {
		if p := cp.FindStringSubmatch(v); p != nil {
			ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
			if p[4] != "" {
				ref = fmt.Sprintf("%s-%s", ref, p[4])
			}

			w := fmt.Sprintf("%s/%s/%s/%s", u, strings.ReplaceAll(g, ".", "/"), a, v)
			w = fmt.Sprintf("%s/%s-%s", w, a, v)
			if s, ok := source["classifier"].(string); ok {
				w = fmt.Sprintf("%s-%s", w, s)
			}
			if s, ok := source["packaging"].(string); ok {
				w = fmt.Sprintf("%s.%s", w, s)
			} else {
				w = fmt.Sprintf("%s.jar", w)
			}

			versions[Version(ref)] = w
		}
	}

	return versions, nil
}
