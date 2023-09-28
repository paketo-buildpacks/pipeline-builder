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

	uri := "https://archive.apache.org/dist/skywalking/java-agent/"

	c := colly.NewCollector()

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+)/$`)
	versions := make(actions.Versions)
	sources := make(map[string]string)

	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		if p := cp.FindStringSubmatch(element.Attr("href")); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

			versions[v] = fmt.Sprintf("%s/%[2]s/apache-skywalking-java-agent-%[2]s.tgz", uri, v)
			sources[v] = fmt.Sprintf("%s/%[2]s/apache-skywalking-java-agent-%[2]s-src.tgz", uri, v)
		}
	})

	if err := c.Visit(uri); err != nil {
		panic(err)
	}

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(err)
	}
	latestSource := actions.Outputs{}
	if sources != nil {
		latestSource["source"] = sources[latestVersion.Original()]
	}

	o, err := actions.NewOutputs(versions[latestVersion.Original()], latestVersion,  latestSource)
	if err != nil {
		panic(err)
	}else {
		o.Write()
	}
}
