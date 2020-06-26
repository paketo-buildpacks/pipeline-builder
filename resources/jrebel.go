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
	"fmt"
	"regexp"

	"github.com/gocolly/colly"
)

type JRebel struct{}

func (j JRebel) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, j)
}

func (j JRebel) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, j)
}

func (JRebel) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (JRebel) Versions(source map[string]interface{}) (map[Version]string, error) {
	c := colly.NewCollector()

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+).*$`)
	versions := make(map[Version]string)
	c.OnHTML("td.views-field-name", func(element *colly.HTMLElement) {
		if p := cp.FindStringSubmatch(element.Text); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

			versions[Version(v)] = fmt.Sprintf(
				"https://dl.zeroturnaround.com/jrebel/releases/jrebel-%s.%s.%s-nosetup.zip", p[1], p[2], p[3])
		}
	})

	err := c.Visit("https://www.jrebel.com/products/jrebel/download/prev-releases")

	return versions, err
}
