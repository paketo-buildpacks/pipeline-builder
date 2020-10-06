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
	"strconv"

	"github.com/Masterminds/semver/v3"
)

type Versions map[string]string

func (v Versions) GetLatest(inputs Inputs) Outputs {
	if len(v) == 0 {
		panic(fmt.Errorf("no candidate version"))
	}

	var err error
	pr := true
	if s, ok := inputs["pre_release"]; ok {
		pr, err = strconv.ParseBool(s)
		if err != nil {
			panic(fmt.Errorf("unable to parse %s as a bool\n%w", s, err))
		}
	}

	var sv []*semver.Version
	for k := range v {
		v, err := semver.NewVersion(k)
		if err != nil {
			panic(fmt.Errorf("unable to parse %s as semver\n%w", k, err))
		}

		if v.Prerelease() == "" || pr {
			sv = append(sv, v)
		}
	}

	sort.Slice(sv, func(i, j int) bool {
		return sv[i].LessThan(sv[j])
	})

	l := sv[len(sv)-1]
	uri := v[l.Original()]
	sha256 := SHA256FromURI(uri)

	return Outputs{
		"sha256":  sha256,
		"uri":     uri,
		"version": fmt.Sprintf("%d.%d.%d", l.Major(), l.Minor(), l.Patch()),
	}
}

var ExtendedVersionPattern = regexp.MustCompile(`^v?([\d]+)\.?([\d]+)?\.?([\d]+)?[-+.]?(.*)$`)

func NormalizeVersion(raw string) string {
	if p := ExtendedVersionPattern.FindStringSubmatch(raw); p != nil {
		for i := 1; i < 4; i++ {
			if p[i] == "" {
				p[i] = "0"
			}
		}

		s := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
		if p[4] != "" {
			s = fmt.Sprintf("%s-%s", s, p[4])
		}

		return s
	}

	panic(fmt.Errorf("unable to parse %s as a extended version (%s)", raw, ExtendedVersionPattern))
}
