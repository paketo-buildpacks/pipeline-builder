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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	b, ok := inputs["bucket"]
	if !ok {
		panic(fmt.Errorf("bucket must be specified"))
	}

	g, ok := inputs["glob"]
	if !ok {
		panic(fmt.Errorf("glob must be specified"))
	}

	client, err := storage.NewClient(context.Background(), option.WithoutAuthentication())
	if err != nil {
		panic(err)
	}

	bucket := client.Bucket(b)
	objects := bucket.Objects(context.Background(), &storage.Query{Prefix: filepath.Dir(g)})

	cp, err := regexp.Compile(g)
	if err != nil {
		panic(fmt.Errorf("unable to compile %s as a regexp", g))
	}

	versions := make(actions.Versions)
	for {
		o, err := objects.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			panic(err)
		}

		if g := cp.FindStringSubmatch(o.Name); g != nil {
			versions[g[1]] = o.MediaLink
		}
	}

	versions.GetLatest(inputs).Write(os.Stdout)
}
