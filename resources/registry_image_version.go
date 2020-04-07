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

	"github.com/heroku/docker-registry-client/registry"
)

type RegistryImageVersion struct{}

func (RegistryImageVersion) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (RegistryImageVersion) Versions(source map[string]interface{}) (map[Version]string, error) {
	rep, ok := source["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository must be specified")
	}

	var (
		url   string
		image string
	)
	re := regexp.MustCompile(`^([^/]+)/(.+)$`)
	if s := re.FindStringSubmatch(rep); s != nil {
		url = fmt.Sprintf("https://%s", s[1])
		image = s[2]
	}

	reg := registry.Registry{
		URL:    url,
		Client: http.DefaultClient,
		Logf:   registry.Quiet,
	}

	if u, ok := source["username"].(string); ok {
		p, ok := source["password"].(string)
		if !ok {
			return nil, fmt.Errorf("password must be specified")
		}

		reg.Client.Transport = registry.WrapTransport(http.DefaultTransport, url, u, p)
	}

	raw, err := reg.Tags(image)
	if err != nil {
		panic(err)
	}

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+)[.-]?(.*)`)
	versions := make(map[Version]string, len(raw))
	for _, t := range raw {
		if p := cp.FindStringSubmatch(t); p != nil {
			ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
			if p[4] != "" {
				ref = fmt.Sprintf("%s-%s", ref, p[4])
			}

			versions[Version(ref)] = ""
		}
	}

	return versions, nil
}
