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

package task

import (
	"fmt"
	"testing"
)

func TestRetryIn(t *testing.T) {
	tests := []struct {
		count    int
		expected int
	}{
		{-1, 1},
		{0, 1},
		{1, 2},
		{2, 3},
		{3, 5},
		{4, 8},
		{5, 13},
		{6, 21},
		{7, 34},
		{8, 55},
		{9, 89},
		{10, 144},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("count=%d", tt.count), func(t *testing.T) {
			result := retryNext(tt.count)
			if result != tt.expected {
				t.Errorf("retryNext(%d) = %d; want %d", tt.count, result, tt.expected)
			}
		})
	}
}
