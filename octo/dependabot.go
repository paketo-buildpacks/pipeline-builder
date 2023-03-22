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

package octo

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/paketo-buildpacks/pipeline-builder/octo/dependabot"
)

func ContributeDependabot(descriptor Descriptor) (*Contribution, error) {
	if _, err := os.Stat(filepath.Join(descriptor.Path, "go.mod")); os.IsNotExist(err) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("unable to determine if %s exists\n%w", filepath.Join(descriptor.Path, "go.mod"), err)
	}

	contrib, err := NewDependabotContribution(dependabot.Dependabot{
		Version: dependabot.Version,
		Updates: []dependabot.Update{
			{
				PackageEcosystem: dependabot.GoModulesPackageEcosystem,
				Directory:        "/",
				Schedule:         dependabot.Schedule{Interval: dependabot.DailyInterval},
				Labels:           []string{"semver:patch", "type:dependency-upgrade"},
				Ignore: []dependabot.Dependency{
					{Name: "github.com/onsi/gomega"},
				},
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return &contrib, err
}
