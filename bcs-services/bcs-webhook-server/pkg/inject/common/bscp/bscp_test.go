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

package bscp

import (
	"testing"
)

func TestAddPathIntoAppInfoMode(t *testing.T) {
	tests := []struct {
		inputStr  string
		inputPath string
		outputStr string
		isError   bool
	}{
		{
			"[{\"business\":\"mars\",\"app\":\"app\",\"cluster\":\"sz\",\"zone\":\"zone1\",\"dc\":\"szdc\"}]",
			"/aaa/bbb",
			"[{\"business\":\"mars\",\"app\":\"app\",\"cluster\":\"sz\",\"zone\":\"zone1\",\"dc\":\"szdc\",\"path\":\"/aaa/bbb\"}]",
			false,
		},
		{
			"[{\"business\":\"mars\",\"app\":\"app\"}]",
			"/a/a/b",
			"",
			true,
		},
	}
	for index, test := range tests {
		t.Logf("%d test", index)
		tmpStr, err := AddPathIntoAppInfoMode(test.inputStr, test.inputPath)
		if err != nil && !test.isError {
			t.Errorf("expect no error, but get err %s", err.Error())
		}
		if err == nil && test.isError {
			t.Errorf("expect error, but no error")
		}
		if tmpStr != test.outputStr {
			t.Errorf("expect %s, but get %s", test.outputStr, tmpStr)
		}
	}
}
