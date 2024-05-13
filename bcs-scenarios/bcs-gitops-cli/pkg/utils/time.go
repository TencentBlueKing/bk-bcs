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
	"fmt"
	"math"
	"time"
)

// FriendTimeFormat output the time
// nolint
func FriendTimeFormat(timeCreate time.Time, timeEnd time.Time) string {
	subTime := int(timeEnd.Sub(timeCreate).Seconds())
	if subTime < 60 {
		return fmt.Sprintf("%ds", subTime)
	}
	if subTime < 60*60 {
		minute := int(math.Floor(float64(subTime / 60)))
		second := subTime % 60
		return fmt.Sprintf("%dm%ds", minute, second)
	}
	if subTime < 60*60*24 {
		hour := int(math.Floor(float64(subTime / (60 * 60))))
		tail := subTime % (60 * 60)
		minute := int(math.Floor(float64(tail / 60)))
		return fmt.Sprintf("%dh%dm", hour, minute)
	}
	day := int(math.Floor(float64(subTime / (60 * 60 * 24))))
	tail := subTime % (60 * 60 * 24)
	hour := int(math.Floor(float64(tail / (60 * 60))))
	tail = subTime % (60 * 60)
	if day >= 10 {
		return fmt.Sprintf("%dd", day)
	}
	return fmt.Sprintf("%dd%dh", day, hour)
}
