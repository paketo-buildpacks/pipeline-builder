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
	"crypto/sha256"
	"fmt"
)

type WebHook struct {
	Owner      string
	Repository string
	Token      string
}

func NewWebHook(salt string, owner string, repository string) WebHook {
	w := WebHook{
		Owner:      owner,
		Repository: repository,
	}

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s%s%s", salt, owner, repository)))
	w.Token = fmt.Sprintf("%x", h.Sum(nil))

	return w
}

func (w WebHook) MarshalYAML() (interface{}, error) {
	return w.Token, nil
}
