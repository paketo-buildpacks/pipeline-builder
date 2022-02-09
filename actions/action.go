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
	"io"
	"os"
	"regexp"
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

func NewOutputs(uri string, latestVersion *semver.Version, additionalOutputs Outputs, mods ...RequestModifierFunc) (Outputs, error) {
	sha256, err := SHA256FromURI(uri, mods...)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate sha256\n%w", err)
	}

	outputs := Outputs{
		"sha256":  sha256,
		"uri":     uri,
		"version": fmt.Sprintf("%d.%d.%d", latestVersion.Major(), latestVersion.Minor(), latestVersion.Patch()),
	}

	for k, v := range additionalOutputs {
		outputs[k] = v
	}

	return outputs, nil
}

func (o Outputs) Write(writer io.Writer) {
	for k, v := range o {
		_, _ = fmt.Fprintf(writer, "::set-output name=%s::%s\n", k, v)
	}
}
