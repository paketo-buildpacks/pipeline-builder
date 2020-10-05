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

package tube

import (
	"context"
	"fmt"
	"regexp"
	"sync"

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/v30/github"
)

type UpdatePackageDependencyContributor struct {
	Descriptor Descriptor
	Package    string
	Salt       string
}

type result struct {
	err   error
	value string
}

func NewUpdatePackageDependencyContributors(descriptor Descriptor, salt string, gh *github.Client) ([]UpdatePackageDependencyContributor, error) {
	results := make(chan result)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		file, _, resp, err := gh.Repositories.GetContents(context.Background(), descriptor.Owner(), descriptor.Repository(), "builder.toml", nil)
		if resp != nil && resp.StatusCode == 404 {
			return
		} else if err != nil {
			results <- result{err: fmt.Errorf("unable to get %s/builder.toml\n%w", descriptor.Name, err)}
			return
		}

		s, err := file.GetContent()
		if err != nil {
			results <- result{err: fmt.Errorf("unable to get %s/builder.toml content\n%w", descriptor.Name, err)}
			return
		}

		raw := struct {
			Buildpacks []struct {
				Image string `toml:"image"`
			} `toml:"buildpacks"`
		}{}

		if _, err := toml.Decode(s, &raw); err != nil {
			results <- result{err: fmt.Errorf("unable to decode\n%w", err)}
			return
		}

		for _, b := range raw.Buildpacks {
			results <- result{value: b.Image}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	re := regexp.MustCompile(`(?m)^(.+):[^:]+$`)
	var b []UpdatePackageDependencyContributor
	for result := range results {
		if result.err != nil {
			return nil, fmt.Errorf("unable to get image listing\n%w", result.err)
		}

		if s := re.FindStringSubmatch(result.value); s != nil {
			b = append(b, UpdatePackageDependencyContributor{
				Descriptor: descriptor,
				Package:    s[1],
				Salt:       salt,
			})
		}
	}

	return b, nil
}

func (UpdatePackageDependencyContributor) Group() string {
	return "package-dependencies"
}

func (u UpdatePackageDependencyContributor) Job() Job {
	b := NewBuildCommonResource()
	p := NewPackageDependencyResource(u.Package)
	s := NewSourceResource(u.Descriptor, u.Salt)

	return Job{
		Name:         p.Name,
		Public:       true,
		SerialGroups: []string{"update-package-dependency"},
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": b.Name,
					},
					{
						"get":      "dependency",
						"resource": p.Name,
						"trigger":  true,
					},
					{
						"get":      "source",
						"resource": s.Name,
					},
				},
			},
			{
				"task": "update-package-dependency",
				"file": "build-common/update-package-dependency.yml",
				"params": map[string]interface{}{
					"DEPENDENCY": u.Package,
				},
			},
			{
				"put": s.Name,
				"params": map[string]interface{}{
					"repository": "source",
					"rebase":     true,
				},
			},
		},
	}

}

func (u UpdatePackageDependencyContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewPackageDependencyResource(u.Package),
		NewSourceResource(u.Descriptor, u.Salt),
	}
}
