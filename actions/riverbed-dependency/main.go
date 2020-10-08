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
	"os"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"

	"github.com/paketo-buildpacks/pipeline-builder/actions"
)

func main() {
	inputs := actions.NewInputs()

	sess := session.Must(session.NewSession())
	svc := s3.New(sess, &aws.Config{Region: aws.String("us-east-1")})
	result, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String("appint-pcf-instrumentation-rpm-master")})
	if err != nil {
		panic(err)
	}

	cp := regexp.MustCompile(`riverbed-appinternals-agent-([\d]+)\.([\d]+)\.([\d]+)_(.+)\.zip`)

	versions := make(actions.Versions)
	for _, o := range result.Contents {
		if p := cp.FindStringSubmatch(*o.Key); p != nil {
			v := fmt.Sprintf("%s.%s.%s-%s", p[1], p[2], p[3], p[4])

			versions[v] = fmt.Sprintf("https://pcf-instrumentation-download.steelcentral.net/%s", *o.Key)
		}
	}

	versions.GetLatest(inputs).Write(os.Stdout)
}
