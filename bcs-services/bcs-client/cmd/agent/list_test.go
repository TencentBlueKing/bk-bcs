/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package agent

import (
	"testing"
)

// Test Label validation test label validation
func TestLabelValidation(t *testing.T) {
	testCases := []struct {
		str    string
		result bool
	}{
		{
			"a",
			true,
		},
		{
			"1",
			true,
		},
		{
			"a_a_k00293",
			true,
		},
		{
			"a???",
			false,
		},
	}
	for _, test := range testCases {
		if isLabelKeyValid(test.str) != test.result {
			t.Fatalf("%v", test)
		}
		if isLabelValueValid(test.str) != test.result {
			t.Fatalf("%v", test)
		}
	}
}
