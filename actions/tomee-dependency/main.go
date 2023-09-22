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

var TomeeVersionPattern = regexp.MustCompile(`^tomee-?([\d]+)\.?([\d]+)?\.?([\d]+)?([-+.])?(.*)/$`)

func main() {
	inputs := actions.NewInputs()

	uri, ok := inputs["uri"]
	if !ok {
		panic(fmt.Errorf("uri must be specified"))
	}

	dist, ok := inputs["dist"]
	if !ok {
		dist = "webprofile"
	}

	major, ok := inputs["major"]
	if !ok {
		major = ""
	}

	c := colly.NewCollector()

	versions := make(actions.Versions)
	sources := make(map[string]string)
	
	c.OnHTML("a[href]", func(element *colly.HTMLElement) {
		if p := TomeeVersionPattern.FindStringSubmatch(element.Attr("href")); p != nil {
			if major == "" || major == p[1] {
				verKey := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
				verVal := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
				if p[4] != "" {
					verKey = fmt.Sprintf("%s-%s", verKey, p[5])
					verVal = fmt.Sprintf("%s%s%s", verVal, p[4], p[5])
				}

				versions[verKey] = fmt.Sprintf("%s/tomee-%[2]s/apache-tomee-%[2]s-%s.tar.gz", uri, verVal, dist)
				sources[verKey] = fmt.Sprintf("%s/tomee-%[2]s/tomee-project-%[2]s-source-release.zip", uri, verVal, dist)
			}
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
	}

	// override the version with the full string (including prerelease information)
	if latestVersion.Prerelease() != "" {
		o["version"] = latestVersion.String()
	}
	o.Write()
}
