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

package release

type Drafter struct {
	Template           string          `yaml:",omitempty"`
	NameTemplate       string          `yaml:"name-template,omitempty"`
	TagTemplate        string          `yaml:"tag-template,omitempty"`
	VersionTemplate    string          `yaml:"version-template,omitempty"`
	ChangeTemplate     string          `yaml:"change-template,omitempty"`
	ChangeTitleEscapes string          `yaml:"change-title-escapes,omitempty"`
	NoChangesTemplate  string          `yaml:"no-changes-template,omitempty"`
	References         []string        `yaml:",omitempty"`
	Categories         []Category      `yaml:",omitempty"`
	ExcludeLabels      []string        `yaml:"exclude-labels,omitempty"`
	IncludeLabels      []string        `yaml:"include-labels,omitempty"`
	Replacers          []Replacer      `yaml:",omitempty"`
	SortBy             SortBy          `yaml:"sort-by,omitempty"`
	SortDirection      SortDirection   `yaml:"sort-direction,omitempty"`
	PreRelease         bool            `yaml:",omitempty"`
	VersionResolver    VersionResolver `yaml:"version-resolver,omitempty"`
}
