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

	actions "github.com/paketo-buildpacks/pipeline-builder/v2/actions"
)

func main() {
	inputs := actions.NewInputs()

	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	c := colly.NewCollector()
	versions := make(actions.Versions)

	var uri string
	switch t {
	case "php":
		uri = "https://download.newrelic.com/php_agent/archive/index.html"

		cp := regexp.MustCompile(`^/php_agent/archive/([\d]+)\.([\d]+)\.([\d]+)\.([\d]+)$`)
		c.OnHTML("a[href]", func(element *colly.HTMLElement) {
			h := element.Attr("href")
			if p := cp.FindStringSubmatch(h); p != nil {
				v := fmt.Sprintf("%s.%s.%s-%s", p[1], p[2], p[3], p[4])

				versions[v] = fmt.Sprintf(
					"https://download.newrelic.com%s/newrelic-php5-%s.%s.%s.%s-linux.tar.gz",
					h, p[1], p[2], p[3], p[4])
			}
		})
	case "dotnet":
		uri = "https://download.newrelic.com/dot_net_agent/previous_releases/"

		cp := regexp.MustCompile(`^/dot_net_agent/previous_releases/([\d]+)\.([\d]+)\.([\d]+)$`)
		c.OnHTML("a[href]", func(element *colly.HTMLElement) {
			h := element.Attr("href")
			if p := cp.FindStringSubmatch(h); p != nil {
				v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

				versions[v] = fmt.Sprintf(
					"https://download.newrelic.com%s/newrelic-dotnet-agent_%s.%s.%s_amd64.tar.gz",
					h, p[1], p[2], p[3])
			}
		})
	default:
		panic(fmt.Errorf("unsupported type %s", t))
	}

	if err := c.Visit(uri); err != nil {
		panic(err)
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write()
	}
}
