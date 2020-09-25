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

type Schedule struct {
	Minute     string
	Hour       string
	DayOfMonth string
	Month      string
	DayOfWeek  string
}

func (Schedule) Type() Type {
	return ScheduleType
}

func (s Schedule) MarshalYAML() (interface{}, error) {
	t := []string{"*", "*", "*", "*", "*"}

	if s.Minute != "" {
		t[0] = s.Minute
	}

	if s.Hour != "" {
		t[1] = s.Hour
	}

	if s.DayOfMonth != "" {
		t[2] = s.DayOfMonth
	}

	if s.Month != "" {
		t[3] = s.Month
	}

	if s.DayOfWeek != "" {
		t[4] = s.DayOfWeek
	}

	return map[string]string{"cron": strings.Join(t, " ")}, nil
}
