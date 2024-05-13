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

package common

import (
	"testing"
	"time"
)

// TestParseAndFormatTime test time parse and format
func TestParseAndFormatTime(t *testing.T) {
	now := time.Now()
	nowStr := FormatTime(now)
	t.Log(nowStr)
	now2, err := ParseTimeString(nowStr)
	if err != nil {
		t.Errorf("parse utc string failed, err %s", err.Error())
	}
	nowStr2 := FormatTime(now2)
	if nowStr != nowStr2 {
		t.Errorf("inconsistent time string, %s, %s", nowStr, nowStr2)
	}
}
