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

package dynamicquery

import (
	"testing"
)

func TestGetTime(t *testing.T) {
	timeStr := "1516849200"
	layout := timestampsLayout

	r, err := getTime(timeStr, layout)
	if err != nil {
		t.Errorf("getTime() failed! err: %v", err)
		return
	}
	if unix, ok := r.(int64); !ok || unix != 1516849200 {
		t.Errorf("getTime() failed! \nresult:\n%v\nexpect:\n1516849200\n", r)
	}

	layout = "2006-01-02 15:04:05"
	r, err = getTime(timeStr, layout)
	if err != nil {
		t.Errorf("getTime() failed! err: %v", err)
		return
	}

	if unix, ok := r.(string); !ok || unix != "2018-01-25 11:00:00" {
		t.Errorf("getTime() failed! \nresult:\n%v\nexpect:\n1516849200\n", r)
	}
}
