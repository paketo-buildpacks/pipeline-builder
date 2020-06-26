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
	"context"
	"fmt"
	"regexp"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

type StackdriverProfiler struct{}

func (s StackdriverProfiler) Check(request CheckRequest) (CheckResult, error) {
	return VersionCheck(request, s)
}

func (s StackdriverProfiler) In(request InRequest, destination string) (InResult, error) {
	return VersionIn(request, destination, s)
}

func (StackdriverProfiler) Out(request OutRequest, source string) (OutResult, error) {
	return OutResult{}, nil
}

func (StackdriverProfiler) Versions(source map[string]interface{}) (map[Version]string, error) {
	client, err := storage.NewClient(context.Background(), option.WithoutAuthentication())
	if err != nil {
		return nil, fmt.Errorf("unable to create GCS client\n%w", err)
	}

	cp := regexp.MustCompile(`^.*_([\d]+)_RC([\d]+).*$`)
	versions := make(map[Version]string)

	it := client.Bucket("cloud-profiler").Objects(context.Background(),
		&storage.Query{Prefix: "java/cloud-profiler-java-agent"})
	for {
		o, err := it.Next()
		if err == iterator.Done {
			break
		} else if err != nil {
			return nil, fmt.Errorf("unable to list contents of bucket cloud-profiler\n%w", err)
		}

		if p := cp.FindStringSubmatch(o.Name); p != nil {
			versions[Version(fmt.Sprintf("%s.%s", p[1], p[2]))] = fmt.Sprintf("https://storage.googleapis.com/%s/%s", o.Bucket, o.Name)
		}
	}

	return versions, nil
}
