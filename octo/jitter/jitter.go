/*
 * Copyright 2023 the original author or authors.
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

package jitter

import (
	"crypto/md5"
	"encoding/binary"
	"math/rand"
	"strconv"

	"github.com/paketo-buildpacks/pipeline-builder/octo/actions/event"
)

type Jitterer struct {
	rng *rand.Rand
}

// New uses a string to calculate a deterministic seed for an random number generator.
//
// The actual implementation does not really matter, we just need to condense a string into an 64-bit number.
// Also: cryptocraphic security is not important here, we just don't want all cron jobs to run at the same time.
func New(seedString string) Jitterer {
	sum := md5.Sum([]byte(seedString))
	seed := binary.LittleEndian.Uint64(sum[0:8]) ^ binary.BigEndian.Uint64(sum[8:16])
	return Jitterer{
		rng: rand.New(rand.NewSource(int64(seed))),
	}
}

func (j Jitterer) jitter(min, max int) string {
	return strconv.Itoa(min + j.rng.Intn(max-min+1))
}

// Jitter takes a Cron event and adds random values for any properties not set.
//
// The values are chosen uniformly-distributed within the validity area of the property.
// E.g. Minutes will be set to a value between 0 and 59.
// The exception is the DayOfMonth property.
// It will never use a value above 28, to ensure it also runs in short months.
func (j Jitterer) Jitter(cron event.Cron) event.Cron {
	if cron.Minute == "" {
		cron.Minute = j.jitter(0, 59)
	}
	if cron.Hour == "" {
		cron.Hour = j.jitter(0, 23)
	}
	if cron.DayOfMonth == "" {
		cron.DayOfMonth = j.jitter(1, 28)
	}
	if cron.Month == "" {
		cron.Month = j.jitter(1, 12)
	}
	if cron.DayOfWeek == "" {
		cron.DayOfWeek = j.jitter(0, 6)
	}
	return cron
}
