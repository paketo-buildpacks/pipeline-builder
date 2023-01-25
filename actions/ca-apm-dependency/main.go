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

package main

import (
	"fmt"
	"regexp"

	"github.com/gocolly/colly"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	uri := "https://packages.broadcom.com/artifactory/apm-agents"

	c := colly.NewCollector()
	versions := make(actions.Versions)

	var cp *regexp.Regexp
	switch t {
	case "dotnet":
		cp = regexp.MustCompile(`^dotnet-agent-linux_([\d]+)\.([\d]+)\.([\d]+)_([\d]+).zip$`)
	case "java":
		cp = regexp.MustCompile(`^agent-default-([\d]+)\.([\d]+)\.([\d]+)_([\d]+).tar$`)
	case "php":
		cp = regexp.MustCompile(`^CA-APM-PHPAgent-([\d]+)\.([\d]+)\.([\d]+)_linux.tar.gz$`)
	default:
		panic(fmt.Errorf("unsupported type %s", t))
	}

	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		h := element.Attr("href")
		if p := cp.FindStringSubmatch(h); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
			if len(p) == 5 {
				v = fmt.Sprintf("%s-%s", v, p[4])
			}

			versions[v] = fmt.Sprintf("%s/%s", uri, p[0])
		}
	})

	if err := c.Visit(uri); err != nil {
		panic(err)
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
