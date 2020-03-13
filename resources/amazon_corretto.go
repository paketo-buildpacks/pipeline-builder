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
	"regexp"
	"strings"

	"github.com/google/go-github/v29/github"
)

type AmazonCorretto struct{}

func (AmazonCorretto) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (a AmazonCorretto) Versions(source map[string]interface{}) (map[Version]string, error) {
	o, ok := source["owner"].(string)
	if !ok {
		return nil, fmt.Errorf("owner must be specified")
	}

	r, ok := source["repository"].(string)
	if !ok {
		return nil, fmt.Errorf("repository must be specified")
	}

	u, ok := source["username"].(string)
	if !ok {
		return nil, fmt.Errorf("username must be specified")
	}

	p, ok := source["password"].(string)
	if !ok {
		return nil, fmt.Errorf("password must be specified")
	}

	gh := github.NewClient((&github.BasicAuthTransport{Username: u, Password: p}).Client())

	var raw []*github.RepositoryRelease
	opt := &github.ListOptions{PerPage: 100}
	for {
		s, r, err := gh.Repositories.ListReleases(context.Background(), o, r, opt)
		if err != nil {
			return nil, fmt.Errorf("unable to list existing releases for %s/%s\n%w", o, r, err)
		}

		raw = append(raw, s...)

		if r.NextPage == 0 {
			break
		}
		opt.Page = r.NextPage
	}

	cp := regexp.MustCompile("([\\d]+)\\.([\\d]+)\\.([\\d]+)\\.([^-]+)")
	versions := make(map[Version]string, len(raw))
	for _, r := range raw {
		if p := cp.FindStringSubmatch(r.GetTagName()); p != nil {
			ref := fmt.Sprintf("%s.%s.%s-%s", p[1], p[2], p[3], p[4])

			versions[Version(ref)] = fmt.Sprintf(
				"https://corretto.aws/downloads/resources/%[1]s/amazon-corretto-%[1]s-linux-x64.tar.gz",
				strings.Replace(ref, "-", ".", -1),
			)
		}
	}

	return versions, nil
}
