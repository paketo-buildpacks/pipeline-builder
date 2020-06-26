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

type SkyWalking struct{}

func (s SkyWalking) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, s)
}

func (s SkyWalking) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, s)
}

func (SkyWalking) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (SkyWalking) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri := "https://archive.apache.org/dist/skywalking"

	c := colly.NewCollector()

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+)/$`)
	versions := make(map[Version]string)
	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		if p := cp.FindStringSubmatch(element.Attr("href")); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

			versions[Version(v)] = fmt.Sprintf("%s/%[2]s/apache-skywalking-apm-%[2]s.tar.gz", uri, v)
		}
	})

	err := c.Visit(uri)

	return versions, err
}
