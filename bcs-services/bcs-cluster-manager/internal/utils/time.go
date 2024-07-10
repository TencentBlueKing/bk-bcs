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

package utils

import (
	"time"
)

const (
	format = "2006-01-02 15:04:05"
	// DefaultTimeZone 默认时区
	DefaultTimeZone = "Asia/Shanghai"
)

// TransTimeFormat trans time.RFC3339 to format that user easy to read
func TransTimeFormat(input string) string {
	formatTime, err := time.Parse(time.RFC3339, input)
	if err != nil {
		return input
	}

	return formatTime.Format(format)
}

// TransTsToTime trans timestamp to time
func TransTsToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// TransTsToStr trans timestamp to time string
func TransTsToStr(timestamp int64) string {
	t := TransTsToTime(timestamp)
	formattedTime := t.Format(time.RFC3339)
	return formattedTime
}

// TransStrToTs trans  time string totimestamp
func TransStrToTs(input string) int64 {
	t, err := time.Parse(time.RFC3339Nano, input)
	if err != nil {
		return 0
	}

	return t.UnixMilli()
}
