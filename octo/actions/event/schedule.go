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

package event

import (
	"strings"
)

const ScheduleType Type = "schedule"

type Schedule []Cron

func (Schedule) Type() Type {
	return ScheduleType
}

type Cron struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
}

func (c Cron) MarshalYAML() (interface{}, error) {
	t := []string{"*", "*", "*", "*", "*"}

	if c.Minute != "" {
		t[0] = c.Minute
	}

	if c.Hour != "" {
		t[1] = c.Hour
	}

	if c.DayOfMonth != "" {
		t[2] = c.DayOfMonth
	}

	if c.Month != "" {
		t[3] = c.Month
	}

	if c.DayOfWeek != "" {
		t[4] = c.DayOfWeek
	}

	return map[string]string{"cron": strings.Join(t, " ")}, nil
}
