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

type Tomcat struct{}

func (Tomcat) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (Tomcat) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri, ok := source["uri"].(string)
	if !ok {
		return nil, fmt.Errorf("uri must be specified")
	}

	c := colly.NewCollector()

	cp := regexp.MustCompile(`^v([\d]+)\.([\d]+)\.([\d]+)/$`)
	versions := make(map[Version]string)
	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		h := element.Attr("href")
		if p := cp.FindStringSubmatch(h); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

			versions[Version(v)] = fmt.Sprintf("%s/v%[2]s/bin/apache-tomcat-%[2]s.tar.gz", uri, v)
		}
	})

	err := c.Visit(uri)

	return versions, err
}
