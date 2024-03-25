/*
 * Copyright 2018-2024 the original author or authors.
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

package octo

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/disiqueira/gotree"
)

func ContributeScripts(descriptor Descriptor) ([]Contribution, error) {
	scriptPath := filepath.Join(descriptor.Path, "scripts", "build.sh")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		return nil, nil
	}

	c := Contribution{
		Path:        "scripts/build.sh",
		Permissions: 0755,
	}

	templateContents := []byte(StatikString("/build-script.sh"))
	tmpl, err := template.New("build.sh").Parse(string(templateContents))
	if err != nil {
		return nil, fmt.Errorf("unable to parse template %q\n%w", templateContents, err)
	}

	output := &bytes.Buffer{}
	err = tmpl.Execute(output, descriptor.Helpers)
	if err != nil {
		return nil, fmt.Errorf("unable to execute template %q\n%w", templateContents, err)
	}

	c.Content = output.Bytes()
	c.Structure = gotree.New("scripts/ [project scripts]")
	c.Structure.Add("scripts/build.sh [build]")

	return []Contribution{c}, nil
}
