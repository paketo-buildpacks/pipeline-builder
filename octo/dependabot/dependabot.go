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

package dependabot

const Version = 2

type Dependabot struct {
	Version int      `yaml:",omitempty"`
	Updates []Update `yaml:",omitempty"`
}

type Update struct {
	PackageEcosystem      PackageEcosystem      `yaml:"package-ecosystem,omitempty"`
	Directory             string                `yaml:",omitempty"`
	Schedule              Schedule              `yaml:",omitempty"`
	Allow                 []Dependency          `yaml:",omitempty"`
	Assignees             []string              `yaml:",omitempty"`
	CommitMessage         CommitMessage         `yaml:"commit-message,omitempty"`
	Ignore                []Dependency          `yaml:",omitempty"`
	Labels                []string              `yaml:",omitempty"`
	Milestone             int                   `yaml:",omitempty"`
	OpenPullRequestsLimit int                   `yaml:"open-pull-requests-limit,omitempty"`
	PullRequestBranchName PullRequestBranchName `yaml:"pull-request-branch-name,omitempty"`
	RebaseStrategy        RebaseStrategy        `yaml:"rebase-strategy,omitempty"`
	Reviewers             []string              `yaml:",omitempty"`
	TargetBranch          string                `yaml:"target-branch,omitempty"`
	VersioningStrategy    VersioningStrategy    `yaml:"versioning-strategy,omitempty"`
}
