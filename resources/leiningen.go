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
	"context"
	"fmt"

	"github.com/google/go-github/v30/github"
)

type Leiningen struct{}

func (l Leiningen) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, l)
}

func (l Leiningen) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, l)
}

func (Leiningen) Out(request OutRequest, source string) (OutResult, error) {
	panic("implement me")
}

func (Leiningen) Versions(source map[string]interface{}) (map[Version]string, error) {
	u, ok := source["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username must be specified")
	}

	at, ok := source["access_token"].(string)
	if !ok {
		return nil, fmt.Errorf("access_token must be specified")
	}

	gh := github.NewClient((&github.BasicAuthTransport{
		Username: u,
		Password: at,
	}).Client())

	versions := make(map[Version]string)

	opt := &github.ListOptions{PerPage: 100}
	for {
		rel, r, err := gh.Repositories.ListReleases(context.Background(), "technomancy", "leiningen", opt)
		if err != nil {
			return nil, fmt.Errorf("unable to list existing releases for technomancy/leiningen\n%w", err)
		}

		for _, r := range rel {
			versions[Version(*r.TagName)] = fmt.Sprintf("https://raw.githubusercontent.com/technomancy/leiningen/%s/bin/lein", *r.TagName)
		}

		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	return versions, nil
}
