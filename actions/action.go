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
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/Masterminds/semver/v3"
)

type Inputs map[string]string

func NewInputs() Inputs {
	re := regexp.MustCompile("^INPUT_([A-Z0-9-_]+)=((?s).+)$")

	i := make(Inputs)
	for _, s := range os.Environ() {
		if g := re.FindStringSubmatch(s); g != nil {
			i[strings.ToLower(g[1])] = g[2]
		}
	}

	return i
}

type Outputs map[string]string

func NewOutputs(uri string, latestVersion *semver.Version, additionalOutputs Outputs, mods ...RequestModifierFunc) (Outputs, error) {
	sha256, err := SHA256FromURI(uri, mods...)
	if err != nil {
		return nil, fmt.Errorf("unable to calculate sha256\n%w", err)
	}

	outputs := Outputs{
		"sha256":  sha256,
		"uri":     uri,
		"version": fmt.Sprintf("%d.%d.%d", latestVersion.Major(), latestVersion.Minor(), latestVersion.Patch()),
	}

	for k, v := range additionalOutputs {
		if k == "source" {
			sourceSha256, err := SHA256FromURI(v, mods...)
			if err != nil {
				return nil, fmt.Errorf("unable to calculate source sha256\n%w", err)
			}
			outputs["source_sha256"] = sourceSha256
		}
		outputs[k] = v
	}

	return outputs, nil
}

// createOrAppend works like os.Create, but instead of replacing (truncating) the file content, it will append to the file, if it already exits.
func createOrAppend(name string) (*os.File, error) {
	return os.OpenFile(name, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
}

func (o Outputs) Write() {
	outputFileName, ok := os.LookupEnv("GITHUB_OUTPUT")
	if !ok {
		panic(errors.New("GITHUB_OUTPUT is not set, see https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#setting-an-output-parameter"))
	}
	writer, err := createOrAppend(outputFileName)
	if err != nil {
		panic(err)
	}
	defer writer.Close()
	delimiter := generateDelimiter()
	for k, v := range o {
		_, _ = fmt.Fprintf(writer, "%s<<%s\n%s\n%s\n", k, delimiter, v, delimiter) // see https://docs.github.com/en/actions/using-workflows/workflow-commands-for-github-actions#multiline-strings
	}
}

func generateDelimiter() string {
	data := make([]byte, 16) // roughly the same entropy as uuid v4 used in https://github.com/actions/toolkit/blob/b36e70495fbee083eb20f600eafa9091d832577d/packages/core/src/file-command.ts#L28
	_, err := rand.Read(data)
	if err != nil {
		log.Fatal("could not generate random delimiter", err)
	}
	return hex.EncodeToString(data)
}
