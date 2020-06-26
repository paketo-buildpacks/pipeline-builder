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

type CAAPM struct{}

func (c CAAPM) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, c)
}

func (c CAAPM) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, c)
}

func (CAAPM) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (CAAPM) Versions(source map[string]interface{}) (map[Version]string, error) {
	t, ok := source["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type must be specified")
	}

	uri := "https://ca.bintray.com/apm-agents"

	c := colly.NewCollector()
	versions := make(map[Version]string)

	var cp *regexp.Regexp
	switch t {
	case "dotnet":
		cp = regexp.MustCompile(`^dotnet-agent-linux_([\d]+)\.([\d]+)\.([\d]+)_([\d]+).zip$`)
	case "java":
		cp = regexp.MustCompile(`^agent-default-([\d]+)\.([\d]+)\.([\d]+)_([\d]+).tar$`)
	case "php":
		cp = regexp.MustCompile(`^CA-APM-PHPAgent-([\d]+)\.([\d]+)\.([\d]+)_linux.tar.gz$`)
	default:
		return nil, fmt.Errorf("unsupported type %s", t)
	}

	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		h := element.Attr("href")
		if p := cp.FindStringSubmatch(h); p != nil {
			ref := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
			if len(p) == 5 {
				ref = fmt.Sprintf("%s-%s", ref, p[4])
			}

			versions[Version(ref)] = fmt.Sprintf("%s/%s", uri, p[0])
		}
	})

	err := c.Visit(uri)
	return versions, err
}
