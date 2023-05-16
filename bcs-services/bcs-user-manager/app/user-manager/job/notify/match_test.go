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
	"testing"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/models"
)

func TestMatch(t *testing.T) {
	tests := []struct {
		name       string
		remain     time.Duration
		expiration time.Duration
		matched    bool
		phase      models.NotifyPhase
	}{
		{
			name:       "match 7 days to 30 days",
			remain:     day * 8,
			expiration: month * 2,
			matched:    true,
			phase:      models.MonthPhase,
		},
		{
			name:       "match 7 days to 30 days, but expiration is less than 1 month",
			remain:     day * 8,
			expiration: day * 28,
			matched:    false,
			phase:      models.NonePhase,
		},
		{
			name:       "1 years, and remain 20 days",
			remain:     day * 20,
			expiration: month * 12,
			matched:    true,
			phase:      models.MonthPhase,
		},
		{
			name:       "1 years, and remain 6 days",
			remain:     day * 6,
			expiration: month * 12,
			matched:    true,
			phase:      models.WeekPhase,
		},
		{
			name:       "1 years, and remain 1 days",
			remain:     day,
			expiration: month * 12,
			matched:    true,
			phase:      models.DayPhase,
		},
		{
			name:       "overdue",
			remain:     time.Duration(0),
			expiration: month * 12,
			matched:    true,
			phase:      models.OverduePhase,
		},
		{
			name:       "remain 5 days, expiration is 8 days",
			remain:     day * 5,
			expiration: day * 8,
			matched:    true,
			phase:      models.WeekPhase,
		},
		{
			name:       "remain 1 days, expiration is 1 days",
			remain:     day,
			expiration: day,
			matched:    true,
			phase:      models.DayPhase,
		},
		{
			name:       "no match",
			remain:     month * 2,
			expiration: month * 12,
			matched:    false,
			phase:      models.NonePhase,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if gotMatched, gotPhase := match(test.remain, test.expiration); gotMatched != test.matched || gotPhase != test.phase {
				t.Errorf("match() = (%v, %v), want (%v, %v)", gotMatched, gotPhase, test.matched, test.phase)
			}
		})
	}
}
