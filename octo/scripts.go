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
	"os"
	"path/filepath"

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

	c.Content = []byte(StatikString("/build-script.sh"))
	c.Structure = gotree.New("scripts/ [project scripts]")
	c.Structure.Add("scripts/build.sh [build]")

	return []Contribution{c}, nil
}
