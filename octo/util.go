/*
 * Copyright 2018-2024 the original author or authors.
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
	"io"
	"os"
	"path/filepath"

	"github.com/rakyll/statik/fs"

	_ "github.com/paketo-buildpacks/pipeline-builder/octo/statik"
)

func Exists(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	} else if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

type Predicate func(string) bool

func Find(path string, predicate Predicate) ([]string, error) {
	var matches []string

	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if predicate(path) {
			matches = append(matches, path)
		}

		return nil
	})

	return matches, err
}

var statik, _ = fs.New()

func StatikString(path string) string {
	in, err := statik.Open(path)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	b, err := io.ReadAll(in)
	if err != nil {
		panic(err)
	}

	return string(b)
}
