/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"testing"
)

// TestVerifyFileUser test VerifyFileUser function
func TestVerifyFileUser(t *testing.T) {
	testCases := []struct {
		userInput string
		isValid bool
	} {
		{
			"user00",
			true,
		},
		{
			"root",
			true,
		},
		{
			"007root",
			false,
		},
		{
			"root+1",
			false,
		},
	}

	for i, c := range testCases {
		t.Logf("test %d", i)
		err := VerifyFileUser(c.userInput)
		if (c.isValid && err != nil) || (!c.isValid && err == nil) {
			t.Errorf("isValid %v err %v", c.isValid, err)
		}
	}
}
