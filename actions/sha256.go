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

package actions

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
)

type RequestModifierFunc func(request *http.Request) *http.Request

func SHA256FromURI(uri string, mods ...RequestModifierFunc) (string, error) {
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return "", fmt.Errorf("unable to create GET %s request\n%w", uri, err)
	}

	for _, m := range mods {
		req = m(req)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	s := sha256.New()
	if _, err := io.Copy(s, resp.Body); err != nil {
		return "", fmt.Errorf("unable to download %s\n%w", uri, err)
	}
	return hex.EncodeToString(s.Sum(nil)), nil
}

func WithBasicAuth(username, password string) RequestModifierFunc {
	return func(req *http.Request) *http.Request {
		req.SetBasicAuth(username, password)
		return req
	}
}
