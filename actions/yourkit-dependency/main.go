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
	"os"
	"regexp"

	"github.com/gocolly/colly"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	c := colly.NewCollector()

	cp := regexp.MustCompile(`(?s)Version: ([\d]+)\.([\d]+).*Build: #([\d]+)`)
	versions := make(actions.Versions)
	c.OnHTML("ul", func(element *colly.HTMLElement) {
		if p := cp.FindStringSubmatch(element.ChildText("li")); p != nil {
			v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])

			versions[v] = fmt.Sprintf(
				"https://www.yourkit.com/download/YourKit-JavaProfiler-%s.%s-b%s-docker.zip", p[1], p[2], p[3])
		}
	})

	if err := c.Visit("https://www.yourkit.com/java/profiler/download"); err != nil {
		panic(err)
	}

	versions.GetLatest(inputs).Write(os.Stdout)
}
