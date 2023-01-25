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
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/google/go-github/v43/github"
	"golang.org/x/oauth2"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

const (
	JDKProductType       = "jdk"
	NIKProductType       = "nik"
	UnknownProductType   = "unknown"
	JDKProductTypePrefix = "graalvm-ce-java"
	NIKProductTypePrefix = "native-image-installable-svm"
)

func main() {
	inputs := actions.NewInputs()

	var (
		err         error
		globRegex   *regexp.Regexp
		productType string
	)
	if s, ok := inputs["glob"]; ok {
		if strings.HasPrefix(s, JDKProductTypePrefix) {
			productType = JDKProductType
		} else if strings.HasPrefix(s, NIKProductTypePrefix) {
			productType = NIKProductType
		} else {
			productType = UnknownProductType
		}

		globRegex, err = regexp.Compile(s)
		if err != nil {
			panic(fmt.Errorf("unable to compile %s as a regexp\n%w", s, err))
		}
	} else {
		globRegex = regexp.MustCompile(".+")
	}

	var c *http.Client
	if s, ok := inputs["token"]; ok {
		c = oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(&oauth2.Token{AccessToken: s}))
	}
	gh := github.NewClient(c)

	tagRegex := regexp.MustCompile(`vm-(.+)`)
	versions := actions.Versions{}
	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, rsp, err := gh.Repositories.ListReleases(context.Background(), "graalvm", "graalvm-ce-builds", opt)
		if err != nil {
			panic(fmt.Errorf("unable to list existing releases for graalvm/graalvm-ce-builds\n%w", err))
		}

		for _, r := range rel {
			if tag := tagRegex.FindStringSubmatch(*r.TagName); tag != nil {
				for _, a := range r.Assets {
					if globRegex.MatchString(*a.Name) {
						version, err := actions.NormalizeVersion(tag[1])
						if err != nil {
							panic(fmt.Errorf("unable to normalize version %s\n%w", version, err))
						}
						versions[version] = *a.BrowserDownloadURL
						break
					}
				}
			}
		}

		if rsp.NextPage == 0 {
			break
		}
		opt.Page = rsp.NextPage
	}

	latestVersion, err := versions.GetLatestVersion(inputs)
	if err != nil {
		panic(fmt.Errorf("unable to get latest version\n%w", err))
	}

	url := versions[latestVersion.Original()]

	latestVersion = semver.MustParse(GetVersion(url))

	outputs, err := actions.NewOutputs(url, latestVersion, nil)
	if err != nil {
		panic(fmt.Errorf("unable to create outputs\n%w", err))
	}

	if productType == NIKProductType {
		// NIK/Substrate VM uses a different version, not the Java version
		//
		// This adjusts the update job to set the PURL & CPE in this way instead
		// of using the standard version format
		urlRegex := regexp.MustCompile(`\/vm-([\d]+\.[\d]+\.[\d]+\.?[\d]?)\/`)
		matches := urlRegex.FindStringSubmatch(url)
		if matches == nil || len(matches) != 2 {
			panic(fmt.Errorf("unable to parse NIK version: %s", matches))
		}
		outputs["cpe"] = matches[1]
		outputs["purl"] = matches[1]
	}

	outputs.Write()
}

func GetVersion(uri string) string {
	// for native-image installer, we convert the URL to the graalvm-ce download URL to fetch the version
	uri = strings.Replace(strings.Replace(uri, ".jar", ".tar.gz", 1), "native-image-installable-svm-", "graalvm-ce-", 1)

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		panic(fmt.Errorf("unable to create GZIP reader\n%w", err))
	}
	defer gz.Close()

	t := tar.NewReader(gz)

	re := regexp.MustCompile(`graalvm-ce-java[^/]+/release`)
	for {
		f, err := t.Next()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			panic(fmt.Errorf("unable to read TAR file\n%w", err))
		}

		if !re.MatchString(f.Name) {
			continue
		}

		b, err := ioutil.ReadAll(t)
		if err != nil {
			panic(fmt.Errorf("unable to read %s\n%w", f.Name, err))
		}

		var v string

		re = regexp.MustCompile(`JAVA_VERSION="([\d]+)\.([\d]+)\.([\d]+)[_]?([\d]+)?"`)
		if g := re.FindStringSubmatch(string(b)); g != nil {
			v = fmt.Sprintf("%s.%s.%s", g[1], g[2], g[3])
		}

		return v
	}

	panic(fmt.Errorf("unable to find file that matches %s", re.String()))
}
