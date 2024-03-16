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
	"strings"

	"github.com/disiqueira/gotree"
)

func ContributeGit(descriptor Descriptor) ([]Contribution, error) {
	c := Contribution{
		Path:        ".gitignore",
		Permissions: 0755,
	}

	licenseHeader := `# Copyright 2018-2020 the original author or authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
`

	c.Content = []byte(strings.Join([]string{
		licenseHeader,
		"bin/",
		"linux/",
		"dependencies/",
		"package/",
		"scratch/",
		"\n",
	}, "\n"))
	c.Structure = gotree.New("Git [various git files]")
	c.Structure.Add(".gitignore")

	return []Contribution{c}, nil
}
