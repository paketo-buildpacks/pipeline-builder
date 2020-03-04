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

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/paketoio/pipeline-builder/tube"
	"github.com/spf13/pflag"
)

func main() {
	t := tube.Transformer{}

	flagSet := pflag.NewFlagSet("Pipeline Builder", pflag.ExitOnError)
	flagSet.StringVar(&t.DescriptorPath, "descriptor", "", "path to input descriptor")
	flagSet.StringVar(&t.PipelinePath, "pipeline", "", "path to output pipeline")

	flagSet.StringVar(&t.GitHubUsername, "github-username", "", "GitHub username")
	flagSet.StringVar(&t.GitHubAccessToken, "github-access-token", "", "GitHub access token")
	flagSet.StringVar(&t.WebHookSalt, "webhook-salt", "", "WebHook Salt")

	flagSet.StringVar(&t.ConcourseURI, "concourse-uri", "", "Concourse URI")
	flagSet.StringVar(&t.ConcourseTeam, "concourse-team", "", "Concourse Team")

	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Fatal(fmt.Errorf("unable to parse flags: %w", err))
	}

	if t.DescriptorPath == "" {
		log.Fatal("--descriptor is required")
	}

	if t.GitHubUsername == "" {
		log.Fatal("--github-username is required")
	}

	if t.GitHubAccessToken == "" {
		log.Fatal("--github-access-token is required")
	}

	if t.WebHookSalt == "" {
		log.Fatal("--web-hook-salt is required")
	}

	if t.ConcourseURI == "" {
		log.Fatal("--concourse-uri is required")
	}

	if t.ConcourseTeam == "" {
		log.Fatal("--concourse-team is required")
	}

	if err := t.Transform(); err != nil {
		log.Fatal(fmt.Errorf("unable to transform: %w", err))
	}
}
