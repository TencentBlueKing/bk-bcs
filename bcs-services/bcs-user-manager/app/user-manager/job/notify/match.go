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
 *
 */

package notify

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

var overdue = time.Duration(0)
var day = 24 * time.Hour
var week = 7 * day
var month = 30 * day

// Match ensure the notified condition is matched, and return the matched level.
// the remaining is the time left to expire token, all the time unit is second.
func match(remain, expiration time.Duration) (matched bool, phase models.NotifyPhase) {
	phase = models.NonePhase
	matched = false
	// match 7 days to 30 days
	if remain > week && remain <= month {
		if expiration >= month {
			return true, models.MonthPhase
		}
		return
	}
	// match 1 day to 7 days
	if remain > day && remain <= week {
		if expiration >= week {
			return true, models.WeekPhase
		}
		return
	}
	// match 0 to 1 day
	if remain > overdue && remain <= day {
		if expiration >= day {
			return true, models.DayPhase
		}
		return
	}
	// match overdue
	if remain <= overdue {
		return true, models.OverduePhase
	}
	return
}
