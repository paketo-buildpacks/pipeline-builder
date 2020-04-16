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

	"github.com/BurntSushi/toml"
	"github.com/google/go-github/v30/github"
)

type ImageType uint8

const (
	Build ImageType = iota
	Run
)

func (i ImageType) String() string {
	return []string{"build", "run"}[i]
}

type UpdateImageDependencyContributor struct {
	Descriptor     Descriptor
	Name           string
	Salt           string
	Type           ImageType
	VersionPattern string
}

func NewUpdateImageDependencyContributors(descriptor Descriptor, salt string, gh *github.Client) ([]UpdateImageDependencyContributor, error) {
	in, err := gh.Repositories.DownloadContents(context.Background(), descriptor.Owner(), descriptor.Repository(), "builder.toml", nil)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s/go.mod\n%w", descriptor.Name, err)
	}
	defer in.Close()

	d := struct {
		Stack struct {
			BuildImage string `toml:"build-image"`
			RunImage   string `toml:"run-image"`
		} `toml:"stack"`
	}{}

	if _, err := toml.DecodeReader(in, &d); err != nil {
		return nil, fmt.Errorf("unable to decode\n%w", err)
	}

	re := regexp.MustCompile(`(?m)^(.+):[^:-]+([^:]+)$`)
	var i []UpdateImageDependencyContributor

	if s := re.FindStringSubmatch(d.Stack.BuildImage); s != nil {
		i = append(i, UpdateImageDependencyContributor{
			Descriptor:     descriptor,
			Name:           s[1],
			Salt:           salt,
			Type:           Build,
			VersionPattern: fmt.Sprintf(".+%s", s[2]),
		})
	}

	if s := re.FindStringSubmatch(d.Stack.RunImage); s != nil {
		i = append(i, UpdateImageDependencyContributor{
			Descriptor:     descriptor,
			Name:           s[1],
			Salt:           salt,
			Type:           Run,
			VersionPattern: fmt.Sprintf(".+%s", s[2]),
		})
	}

	return i, nil
}

func (UpdateImageDependencyContributor) Group() string {
	return "builder-dependencies"
}

func (u UpdateImageDependencyContributor) Job() Job {
	b := NewBuildCommonResource()
	i := NewImageDependencyResource(u.Name, u.VersionPattern)
	s := NewSourceResource(u.Descriptor, u.Salt)

	return Job{
		Name:   fmt.Sprintf("update-%s-image", u.Type),
		Public: true,
		Plan: []map[string]interface{}{
			{
				"in_parallel": []map[string]interface{}{
					{
						"get":      "build-common",
						"resource": b.Name,
					},
					{
						"get":      "image",
						"resource": i.Name,
						"trigger":  true,
					},
					{
						"get":      "source",
						"resource": s.Name,
					},
				},
			},
			{
				"task": "update-image-dependency",
				"file": "build-common/update-image-dependency.yml",
				"params": map[string]interface{}{
					"TYPE": u.Type.String(),
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

func (u UpdateImageDependencyContributor) Resources() []Resource {
	return []Resource{
		NewBuildCommonResource(),
		NewImageDependencyResource(u.Name, u.VersionPattern),
		NewSourceResource(u.Descriptor, u.Salt),
	}
}
