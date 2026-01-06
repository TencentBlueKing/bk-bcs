/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package timex xxx
package timex

import "time"

const (
	// BkrepoDateTime time format
	BkrepoDateTime = "2006-01-02T15:04:05.999"
)

// TransStrToUTCStr trans time string to utc RFC3339 time string
func TransStrToUTCStr(timeType, input string) string {
	switch timeType {
	case time.RFC3339Nano:
		t, err := time.Parse(time.RFC3339Nano, input)
		if err != nil {
			return input
		}
		return t.UTC().Format(time.RFC3339)
	case BkrepoDateTime:
		t, err := time.Parse(BkrepoDateTime, input)
		if err != nil {
			return input
		}
		return t.UTC().Format(time.RFC3339)
	case time.DateTime:
		t, err := time.Parse(time.DateTime, input)
		if err != nil {
			return input
		}
		return t.UTC().Format(time.RFC3339)
	default:
		return input
	}
}
