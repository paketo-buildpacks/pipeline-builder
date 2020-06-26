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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
)

func DownloadArtifact(uri string, destination string) (string, error) {
	if err := os.MkdirAll(destination, 0755); err != nil {
		return "", fmt.Errorf("unable to make directory %s\n%w", destination, err)
	}

	resp, err := http.Get(uri)
	if err != nil {
		return "", fmt.Errorf("unable to GET %s\n%w", uri, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("unable to download %s", uri)
	}

	_, _ = fmt.Fprintf(os.Stderr, "Downloading %s\n", uri)

	file := filepath.Join(destination, filepath.Base(uri))
	f, err := os.OpenFile(file, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return "", fmt.Errorf("unable to open file %s\n%w", file, err)
	}
	defer f.Close()

	s := sha256.New()
	out := io.MultiWriter(f, s)

	if _, err := io.Copy(out, resp.Body); err != nil {
		return "", fmt.Errorf("unable to copy data\n%w", err)
	}
	hash := hex.EncodeToString(s.Sum(nil))

	file = filepath.Join(destination, "sha256")
	if err := ioutil.WriteFile(file, []byte(hash), 0644); err != nil {
		return "", fmt.Errorf("unable to write %s\n%w", file, err)
	}

	file = filepath.Join(destination, "uri")
	if err := ioutil.WriteFile(file, []byte(uri), 0644); err != nil {
		return "", fmt.Errorf("unable to write %s\n%w", file, err)
	}

	return hash, nil
}
