/*
 * Copyright 2018-2023 the original author or authors.
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

package jitter_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/actions/event"
	"github.com/paketo-buildpacks/pipeline-builder/v2/octo/jitter"
	"github.com/sclevine/spec"
)

func testJitter(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect
	)

	context("Jitterer", func() {
		it("jittering should change unset values", func() {
			jitterer := jitter.New("my-seed")
			jittered := jitterer.Jitter(event.Cron{})

			Expect(jittered.DayOfMonth).NotTo(BeEmpty())
			Expect(jittered.DayOfWeek).NotTo(BeEmpty())
			Expect(jittered.Hour).NotTo(BeEmpty())
			Expect(jittered.Minute).NotTo(BeEmpty())
			Expect(jittered.Month).NotTo(BeEmpty())
		})

		it("jittering should not change set values", func() {
			jitterer := jitter.New("my-seed")
			jittered := jitterer.Jitter(event.Cron{
				DayOfMonth: "42",
				DayOfWeek:  "42",
				Hour:       "42",
				Minute:     "42",
				Month:      "42",
			})

			Expect(jittered.DayOfMonth).To(Equal("42"))
			Expect(jittered.DayOfWeek).To(Equal("42"))
			Expect(jittered.Hour).To(Equal("42"))
			Expect(jittered.Minute).To(Equal("42"))
			Expect(jittered.Month).To(Equal("42"))
		})

		it("jittering with the same seed should produce the same result", func() {
			jitterer := jitter.New("my-seed")
			jitteredA := jitterer.Jitter(event.Cron{})
			jitterer = jitter.New("my-seed")
			jitteredB := jitterer.Jitter(event.Cron{})
			Expect(jitteredA).To(Equal(jitteredB))
		})

		it("jittering with different seeds should produce a different result (with high probability)", func() {
			jitterer := jitter.New("my-seed")
			jitteredA := jitterer.Jitter(event.Cron{})
			jitterer = jitter.New("my-other-seed")
			jitteredB := jitterer.Jitter(event.Cron{})
			Expect(jitteredA).NotTo(Equal(jitteredB))
		})

	})
}
