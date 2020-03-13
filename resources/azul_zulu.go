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
	"encoding/json"
	"fmt"
	"net/http"
)

type AzulZulu struct{}

func (AzulZulu) Out(request OutRequest, destination string) (OutResult, error) {
	return OutResult{}, nil
}

func (a AzulZulu) Versions(source map[string]interface{}) (map[Version]string, error) {
	t, ok := source["type"].(string)
	if !ok {
		return nil, fmt.Errorf("type must be specified")
	}

	v, ok := source["version"].(string)
	if !ok {
		return nil, fmt.Errorf("version must be specified")
	}

	uri := fmt.Sprintf("https://api.azul.com/zulu/download/azure-only/v1.0/bundles/latest/"+
		"?arch=x86"+
		"&ext=tar.gz"+
		"&features=%s"+
		"&hw_bitness=64"+
		"&jdk_version=%s"+
		"&os=linux",
		t, v)

	resp, err := http.Get(uri)
	if err != nil {
		return nil, fmt.Errorf("unable to get %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("unable to download %s: %d", uri, resp.StatusCode)
	}

	raw := struct {
		JDKVersion []int  `json:"jdk_version"`
		URL        string `json:"url"`
	}{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("unable to decode payload\n%w", err)
	}

	versions := map[Version]string{
		Version(fmt.Sprintf("%d.%d.%d", raw.JDKVersion[0], raw.JDKVersion[1], raw.JDKVersion[2])): raw.URL,
	}

	return versions, nil
}
