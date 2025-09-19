/*
 * Copyright 2023-2023 the original author or authors.
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

	typeName, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	versionBranch, ok := inputs["version"]
	if !ok {
		panic(fmt.Errorf("version must be specified"))
	}

	arch, ok := inputs["arch"]
	if !ok {
		arch = "x64"
	}

	if arch == "arm64" {
		arch = "aarch64" // cause Oracle needs it this way
	}

	var pattern, majorPattern *regexp.Regexp
	var key string
	var urlTemplate string
	switch typeName {
	case "jdk":
		pattern = regexp.MustCompile(`^Java SE Development Kit (\d+\.\d+\.\d+) downloads$`)
		majorPattern = regexp.MustCompile(`^Java SE Development Kit (\d+) downloads$`)
		key = fmt.Sprintf("java%s", versionBranch)
		urlTemplate = "https://download.oracle.com/java/%s/archive/jdk-%s_linux-%s_bin.tar.gz"
	case "graalvm":
		pattern = regexp.MustCompile(`^GraalVM for JDK (\d+\.\d+\.\d+) downloads$`)
		key = fmt.Sprintf("graalvmjava%s", versionBranch)
		urlTemplate = "https://download.oracle.com/graalvm/%s/archive/graalvm-jdk-%s_linux-%s_bin.tar.gz"
	default:
		panic(fmt.Errorf("unsupported type %s", typeName))
	}

	uri := "https://www.oracle.com/java/technologies/downloads"

	c := colly.NewCollector()
	versions := make(actions.Versions)

	c.OnHTML(fmt.Sprintf("h3[id=%s]", key), func(element *colly.HTMLElement) {
		foundVersions := pattern.FindStringSubmatch(element.Text)
		if foundVersions == nil {
			foundVersions = majorPattern.FindStringSubmatch(element.Text)
		}
		if len(foundVersions) == 2 {
			foundVersion := foundVersions[1] // there should only ever be one
			versions[foundVersion] = fmt.Sprintf(urlTemplate, versionBranch, foundVersion, arch)
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
