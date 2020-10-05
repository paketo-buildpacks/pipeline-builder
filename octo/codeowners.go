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

package octo

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/disiqueira/gotree"
)

func ContributeCodeOwners(descriptor Descriptor) Contribution {
	var c Contribution

	c.Path = filepath.Join(".github", "CODEOWNERS")
	c.Permissions = 0644

	c.Structure = gotree.New(c.Path)

	var lines []string
	for _, o := range descriptor.CodeOwners {
		c.Structure.Add(o.Owner)
		lines = append(lines, fmt.Sprintf("%s %s", o.Path, o.Owner))
	}

	c.Content = []byte(strings.Join(lines, "\n"))

	return c
}
