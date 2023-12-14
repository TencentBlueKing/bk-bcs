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

package sfs

import (
	"testing"

	pbbase "github.com/TencentBlueking/bk-bcs/bcs-services/bcs-bscp/pkg/protocol/core/base"
)

func TestIsAPIVersionMatch(t *testing.T) {
	leastAPIVersion = &pbbase.Versioning{
		Major: 3,
		Minor: 4,
		Patch: 5,
	}

	testCases := []struct {
		name     string
		ver      *pbbase.Versioning
		expected bool
	}{
		{
			name: "Major version mismatch",
			ver: &pbbase.Versioning{
				Major: 2,
				Minor: 4,
				Patch: 5,
			},
			expected: false,
		},
		{
			name: "Minor version mismatch",
			ver: &pbbase.Versioning{
				Major: 3,
				Minor: 3,
				Patch: 5,
			},
			expected: false,
		},
		{
			name: "Patch version mismatch",
			ver: &pbbase.Versioning{
				Major: 3,
				Minor: 4,
				Patch: 4,
			},
			expected: false,
		},
		{
			name: "Major version match",
			ver: &pbbase.Versioning{
				Major: 3,
				Minor: 4,
				Patch: 5,
			},
			expected: true,
		},
		{
			name: "Minor version match",
			ver: &pbbase.Versioning{
				Major: 3,
				Minor: 4,
				Patch: 6,
			},
			expected: true,
		},
		{
			name: "Patch version match",
			ver: &pbbase.Versioning{
				Major: 3,
				Minor: 5,
				Patch: 5,
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		result := IsAPIVersionMatch(tc.ver)
		if result != tc.expected {
			t.Errorf("Test %s failed, Expected %v, got %v", tc.name, tc.expected, result)
			t.Fail()
		}
	}
}
