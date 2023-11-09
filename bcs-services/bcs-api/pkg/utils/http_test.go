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
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRefineIPv6Addr(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"http://1111:1111:1112:15::1:443", "http://[1111:1111:1112:15::1]:443"},
		{"https://[1111:1111:1112:15::2]:8080", "https://[1111:1111:1112:15::2]:8080"},
		{"ftp://[1111:1111:1112:15::3]:21", "ftp://[1111:1111:1112:15::3]:21"},
		{"ftp://127.0.0.1/test", "ftp://127.0.0.1/test"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("input=%s", tc.input), func(t *testing.T) {
			actual := RefineIPv6Addr(tc.input)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
