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
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	t, ok := inputs["type"]
	if !ok {
		panic(fmt.Errorf("type must be specified"))
	}

	uri := "https://download.appdynamics.com/download/downloadfilelatest"

	resp, err := http.Get(uri)
	if err != nil {
		panic(fmt.Errorf("unable to get %s\n%w", uri, err))
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		panic(fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode))
	}

	var raw []struct {
		DownloadPath string `json:"download_path"`
		FileType     string `json:"filetype"`
		Version      string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		panic(fmt.Errorf("unable to decode payload\n%w", err))
	}

	cp := regexp.MustCompile(`^([\d]+)\.([\d]+)\.([\d]+)[.-]?(.*)`)
	versions := make(actions.Versions)
	for _, r := range raw {
		if t == r.FileType {
			if p := cp.FindStringSubmatch(r.Version); p != nil {
				v := fmt.Sprintf("%s.%s.%s", p[1], p[2], p[3])
				if p[4] != "" {
					v = fmt.Sprintf("%s-%s", v, p[4])
				}

				var uri string
				switch t {
				case "php-tar":
					uri = fmt.Sprintf("https://packages.appdynamics.com/php/%[1]s/appdynamics-php-agent-linux_x64-%[1]s.tar.bz2", r.Version)
				case "sun-jvm":
					uri = fmt.Sprintf("https://packages.appdynamics.com/java/%[1]s/AppServerAgent-%[1]s.zip", r.Version)
				default:
					panic(fmt.Errorf("unknown uri type %s\n", t))
				}

				versions[v] = uri
			}
			break
		}
	}

	if o, err := versions.GetLatest(inputs); err != nil {
		panic(err)
	} else {
		o.Write(os.Stdout)
	}
}
