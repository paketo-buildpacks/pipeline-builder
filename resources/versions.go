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
	"io/ioutil"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/Masterminds/semver/v3"
)

type VersionProvider interface {
	Versions(source map[string]interface{}) (map[Version]string, error)
}

func VersionCheck(request CheckRequest, provider VersionProvider) (CheckResult, error) {
	versions, err := provider.Versions(request.Source)
	if err != nil {
		return nil, err
	}

	var vp *regexp.Regexp
	if s, ok := request.Source["version_pattern"].(string); ok {
		vp, err = regexp.Compile(s)
		if err != nil {
			return nil, err
		}
	}

	pr := true
	if b, ok := request.Source["pre_release"].(bool); ok {
		pr = b
	}

	var sv []*semver.Version
	for k := range versions {
		if vp == nil || vp.MatchString(string(k)) {
			v, err := semver.NewVersion(string(k))
			if err != nil {
				return nil, err
			}

			if v.Prerelease() == "" || pr {
				sv = append(sv, v)
			}
		}
	}

	sort.Slice(sv, func(i, j int) bool {
		return sv[i].LessThan(sv[j])
	})

	if request.Version == "" {
		sv = sv[len(sv)-1:]
	} else {
		since, err := semver.NewVersion(string(request.Version))
		if err != nil {
			return nil, err
		}

		for i, v := range sv {
			if !v.LessThan(since) {
				sv = sv[i:]
				break
			}
		}
	}

	result := CheckResult{}
	for _, v := range sv {
		result = append(result, Version(v.Original()))
	}

	return result, nil
}

func VersionIn(request InRequest, destination string, provider VersionProvider) (InResult, error) {
	versions, err := provider.Versions(request.Source)
	if err != nil {
		return InResult{}, err
	}

	uri := versions[request.Version]

	file := filepath.Join(destination, "version")
	if err := ioutil.WriteFile(file, []byte(request.Version), 0644); err != nil {
		return InResult{}, err
	}

	result := InResult{Version: request.Version}

	if uri != "" {
		sha256, err := DownloadArtifact(uri, destination)
		if err != nil {
			return InResult{}, err
		}

		result.Metadata = Metadata{
			"uri":    uri,
			"sha256": sha256,
		}
	}

	return result, nil
}
