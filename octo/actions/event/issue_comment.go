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

const IssueCommentType Type = "issue_comment"

type IssueCommentActivityType string

const (
	IssueCommentCreated IssueCommentActivityType = "created"
	IssueCommentEdited  IssueCommentActivityType = "edited"
	IssueCommentDeleted IssueCommentActivityType = "deleted"
)

type IssueComment struct {
	Types []IssueCommentActivityType `yaml:",omitempty"`
}

func (IssueComment) Type() Type {
	return IssueCommentType
}
