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
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-data-manager/pkg/types"
)

func TestFormatTimeIgnoreMin(t *testing.T) {
	original := "2022-03-08 22:00:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.HourTimeFormat, original, local)
	result := formatTimeIgnoreMin(originalTime)
	assert.Equal(t, "2022-03-08 22:00:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreSec(t *testing.T) {
	original := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.MinuteTimeFormat, original, local)
	result := formatTimeIgnoreSec(originalTime)
	assert.Equal(t, "2022-03-08 22:11:00 +0800 CST", result.String())
}

func TestFormatTimeIgnoreHour(t *testing.T) {
	original := "2022-03-08"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.DayTimeFormat, original, local)
	result := formatTimeIgnoreHour(originalTime)
	assert.Equal(t, "2022-03-08 00:00:00 +0800 CST", result.String())
}

func TestGetIndex(t *testing.T) {
	origin := "2022-03-08 22:11:00"
	local := time.Local
	originalTime, _ := time.ParseInLocation(types.MinuteTimeFormat, origin, local)
	day := GetIndex(originalTime, "day")
	hour := GetIndex(originalTime, "hour")
	minute := GetIndex(originalTime, "minute")
	assert.Equal(t, 8, day)
	assert.Equal(t, 22, hour)
	assert.Equal(t, 11, minute)
}
