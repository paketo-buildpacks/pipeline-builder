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
)

type Takipi struct{}

func (t Takipi) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, t)
}

func (t Takipi) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, t)
}

func (Takipi) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (Takipi) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri := "https://app.overops.com/app/download?t=sa-tgz"

	client := http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	resp, err := client.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to create GET %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 300 || resp.StatusCode > 399 {
		return nil, fmt.Errorf("unable to determine redirect URI")
	}

	cp := regexp.MustCompile(`^.*/takipi-agent-([\d]+)\.([\d]+)\.([\d]+)\.tar\.gz`)
	versions := make(map[Version]string, 1)

	location := resp.Header.Get("Location")
	if p := cp.FindStringSubmatch(location); p != nil {
		ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
		versions[Version(ref)] = location
	}

	return versions, nil
}
