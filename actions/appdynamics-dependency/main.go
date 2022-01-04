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
	"bytes"
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
		panic(fmt.Errorf("type must be specified [`sun-jvm` or `php-tar`]"))
	}

	user, ok := inputs["user"]
	if !ok {
		panic(fmt.Errorf("user must be specified"))
	}

	pass, ok := inputs["password"]
	if !ok {
		panic(fmt.Errorf("password must be specified"))
	}

	versions, err := FetchLatestVersions(t)
	if err != nil {
		panic(fmt.Errorf("unable to fetch versions\n%w", err))
	}

	addToken := func(request *http.Request) *http.Request {
		token, err := FetchAPIToken(user, pass)
		if err != nil {
			panic(fmt.Errorf("unable to fetch token\n%w", err))
		}

		request.Header.Add("Authorization", token)

		return request
	}

	if o, err := versions.GetLatest(inputs, addToken); err != nil {
		panic(fmt.Errorf("unable to get latest\n%w", err))
	} else {
		o.Write(os.Stdout)
	}
}

func FetchAPIToken(user, password string) (string, error) {
	uri := "https://identity.msrv.saas.appdynamics.com/v2.0/oauth/token"

	resp, err := http.Post(uri, "application/json",
		bytes.NewBufferString(
			fmt.Sprintf(`{"username": "%s","password": "%s","scopes": ["download"]}`, user, password)))
	if err != nil {
		return "", fmt.Errorf("unable to post %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to read token %s: %d", uri, resp.StatusCode)
	}

	var raw struct {
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
		AccessToken string `json:"access_token"`
		Scope       string `json:"scope"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return "", fmt.Errorf("unable to decode payload\n%w", err)
	}
	return fmt.Sprintf("%s %s", raw.TokenType, raw.AccessToken), nil
}

func FetchLatestVersions(t string) (actions.Versions, error) {
	uri := "https://download.appdynamics.com/download/downloadfilelatest"

	resp, err := http.Get(uri)
	if err != nil {
		return actions.Versions{}, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return actions.Versions{}, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	var raw []struct {
		DownloadPath string `json:"download_path"`
		FileType     string `json:"filetype"`
		Version      string `json:"version"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return actions.Versions{}, fmt.Errorf("unable to decode payload\n%w", err)
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

				versions[v] = r.DownloadPath
			}
			break
		}
	}

	return versions, nil
}
