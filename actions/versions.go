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

package actions

import (
	"fmt"
	"regexp"
	"sort"

	"github.com/Masterminds/semver/v3"
)

type Versions struct {
	Contents        map[string]string
	Inputs          Inputs
	SemverConverter SemverConverter
}

func (v Versions) GetLatest() Outputs {
	var err error

	var vp *regexp.Regexp
	if s, ok := v.Inputs["version-pattern"]; ok {
		vp, err = regexp.Compile(s)
		if err != nil {
			panic(err)
		}
	}

	var keys []string
	for k := range v.Contents {
		if vp == nil || vp.MatchString(k) {
			keys = append(keys, k)
		}
	}

	if len(keys) == 0 {
		panic(fmt.Errorf("no candidate version"))
	}

	sort.Slice(keys, func(i, j int) bool {
		return v.SemverConverter(keys[i]).LessThan(v.SemverConverter(keys[j]))
	})

	version := keys[len(keys)-1]
	uri := v.Contents[version]
	sha256 := SHA256FromURI(uri)

	sv := v.SemverConverter(version)

	return Outputs{
		"sha256":  sha256,
		"uri":     uri,
		"version": fmt.Sprintf("%d.%d.%d", sv.Major(), sv.Minor(), sv.Patch()),
	}
}

type SemverConverter func(string) *semver.Version

var IdentitySemverConverter = func(raw string) *semver.Version {
	sv, err := semver.NewVersion(raw)
	if err != nil {
		panic(fmt.Errorf("unable to convert %s to semver\n%w", raw, err))
	}

	return sv
}

var MetadataVersionPattern = regexp.MustCompile(`^v?([\d]+)\.?([\d]+)?\.?([\d]+)?[+-.]?(.*)$`)
var MetadataSemverConverter = func(raw string) *semver.Version {
	if p := MetadataVersionPattern.FindStringSubmatch(raw); p != nil {
		for i := 1; i < 4; i++ {
			if p[i] == "" {
				p[i] = "0"
			}
		}

		s := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
		if p[4] != "" {
			s = fmt.Sprintf("%s+%s", s, p[4])
		}

		sv, err := semver.NewVersion(s)
		if err != nil {
			panic(fmt.Errorf("unable to convert %s to semver\n%w", s, err))
		}

		return sv
	}

	panic(fmt.Errorf("unable to parse %s as a metadata version (%s)", raw, MetadataVersionPattern))
}
