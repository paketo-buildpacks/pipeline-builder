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

package event

const PullRequestTargetType Type = "pull_request_target"

type PullRequestTargetActivityType string

const (
	PullRequestTargetAssigned             PullRequestTargetActivityType = "assigned"
	PullRequestTargetUnassigned           PullRequestTargetActivityType = "unassigned"
	PullRequestTargetLabeled              PullRequestTargetActivityType = "labeled"
	PullRequestTargetUnlabeled            PullRequestTargetActivityType = "unlabeled"
	PullRequestTargetOpened               PullRequestTargetActivityType = "opened"
	PullRequestTargetEdited               PullRequestTargetActivityType = "edited"
	PullRequestTargetClosed               PullRequestTargetActivityType = "closed"
	PullRequestTargetReopened             PullRequestTargetActivityType = "reopened"
	PullRequestTargetSynchronize          PullRequestTargetActivityType = "synchronize"
	PullRequestTargetReadyForReview       PullRequestTargetActivityType = "ready_for_review"
	PullRequestTargetLocked               PullRequestTargetActivityType = "locked"
	PullRequestTargetUnlocked             PullRequestTargetActivityType = "unlocked"
	PullRequestTargetReviewRequested      PullRequestTargetActivityType = "review_requested"
	PullRequestTargetReviewRequestRemoved PullRequestTargetActivityType = "review_request_removed"
)

type PullRequestTarget struct {
	Types          []PullRequestTargetActivityType `yaml:",omitempty"`
	Branches       []string                        `yaml:",omitempty"`
	BranchesIgnore []string                        `yaml:"branches-ignore,omitempty"`
	Tags           []string                        `yaml:"tags,omitempty"`
	TagsIgnore     []string                        `yaml:"tags-ignore,omitempty"`
	Paths          []string                        `yaml:",omitempty"`
	PathsIgnore    []string                        `yaml:"paths-ignore,omitempty"`
}

func (PullRequestTarget) Type() Type {
	return PullRequestTargetType
}
