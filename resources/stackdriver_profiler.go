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
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
)

type StackdriverProfiler struct{}

func (StackdriverProfiler) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (StackdriverProfiler) Versions(source map[string]interface{}) (map[Version]string, error) {
	uri := "https://storage.googleapis.com/cloud-profiler/java/latest/profiler_java_agent.tar.gz"

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	gz, err := gzip.NewReader(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to decompress %s\n%w", uri, err)
	}
	defer gz.Close()

	versions := make(map[Version]string, 1)

	t := tar.NewReader(gz)
	for {
		f, err := t.Next()
		if err != nil && err == io.EOF {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unable to untar %s\n%w", uri, err)
		}

		if f.Name == "version.txt" {
			b, err := ioutil.ReadAll(t)
			if err != nil {
				return nil, fmt.Errorf("unable to read version.txt from %s\n%w", uri, err)
			}

			s := string(b)
			g := regexp.MustCompile("cloud-profiler-java-agent_([^_]+).*").FindStringSubmatch(s)
			if len(g) < 1 {
				return nil, fmt.Errorf("unable to extract version from %s", s)
			}

			versions[Version(g[1])] = uri
			break
		}
	}

	return versions, nil
}
