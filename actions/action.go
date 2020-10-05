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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Inputs map[string]string

func NewInputs() Inputs {
	re := regexp.MustCompile("^INPUT_([A-Z0-9-_]+)=(.+)$")

	i := make(Inputs)
	for _, s := range os.Environ() {
		if g := re.FindStringSubmatch(s); g != nil {
			i[strings.ToLower(g[1])] = g[2]
		}
	}

	return i
}

type Outputs map[string]string

func (o Outputs) Write(writer io.Writer) {
	for k, v := range o {
		_, _ = fmt.Fprintf(writer, "::set-output name=%s::%s\n", k, v)
	}
}

type Versions map[string]string

func (v Versions) GetLatest(inputs Inputs) Outputs {
	var err error

	var vp *regexp.Regexp
	if s, ok := inputs["version-pattern"]; ok {
		vp, err = regexp.Compile(s)
		if err != nil {
			panic(err)
		}
	}

	pr := false
	if b, ok := inputs["pre-release"]; ok {
		pr, err = strconv.ParseBool(b)
		if err != nil {
			panic(err)
		}
	}

	var sv []*semver.Version
	for k := range v {
		if vp == nil || vp.MatchString(k) {
			v, err := semver.NewVersion(k)
			if err != nil {
				panic(err)
			}

			if v.Prerelease() == "" || pr {
				sv = append(sv, v)
			}
		}
	}

	if len(sv) == 0 {
		panic(fmt.Errorf("no candidate version"))
	}

	sort.Slice(sv, func(i, j int) bool {
		return sv[i].LessThan(sv[j])
	})

	latest := sv[len(sv)-1]

	outputs := Outputs{
		"sha256":  GetSHA256(v[latest.Original()]),
		"uri":     v[latest.Original()],
		"version": fmt.Sprintf("%d.%d.%d", latest.Major(), latest.Minor(), latest.Patch()),
	}

	return outputs
}

func GetSHA256(uri string) string {
	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	s := sha256.New()

	if _, err := io.Copy(s, resp.Body); err != nil {
		panic(fmt.Errorf("unable to download %s\n%w", uri, err))
	}

	return hex.EncodeToString(s.Sum(nil))
}
