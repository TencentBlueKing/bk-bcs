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

package aws

import (
	"testing"
	"time"
)

func TestIncreaseTimeSeries(t *testing.T) {
	it := NewIncreseSeries(time.Second*2, 0.15, 0.15)
	for i := 0; i < 20; i++ {
		tmp := it.Next()
		t.Log(tmp)
		if tmp/time.Second < 2 {
			t.Errorf("invalid time duration %+v", tmp)
		}
	}
}
