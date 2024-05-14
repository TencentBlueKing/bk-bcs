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
	"testing"
)

// TestGetIPListFromCidr test GetIPListFromCidr function
func TestGetIPListFromCidr(t *testing.T) {
	testCases := []struct {
		cidr   string
		ipList []string
		err    bool
	}{
		{
			cidr: "127.0.0.0/29",
			ipList: []string{
				"127.0.0.2",
				"127.0.0.3",
				"127.0.0.4",
				"127.0.0.5",
				"127.0.0.6",
			},
			err: false,
		},
		{
			cidr: "127.0.1.0/28",
			ipList: []string{
				"127.0.1.2",
				"127.0.1.3",
				"127.0.1.4",
				"127.0.1.5",
				"127.0.1.6",
				"127.0.1.7",
				"127.0.1.8",
				"127.0.1.9",
				"127.0.1.10",
				"127.0.1.11",
				"127.0.1.12",
				"127.0.1.13",
				"127.0.1.14",
			},
			err: false,
		},
		{
			cidr: "127.0.1.0",
			err:  true,
		},
	}
	for index, test := range testCases {
		t.Logf("test %d", index)
		tmpIPList, err := GetIPListFromCidr(test.cidr)
		if err != nil {
			if !test.err {
				t.Errorf("expect no err, but get err %s", err.Error())
				return
			}
			continue
		}
		if test.err {
			t.Error("expect err, but get no err")
			return
		}
		if len(tmpIPList) != len(test.ipList) {
			t.Errorf("expect %v, but get %v", test.ipList, tmpIPList)
			return
		}
		for i, tmpIP := range tmpIPList {
			if test.ipList[i] != tmpIP {
				t.Errorf("expect %s, but get %s", test.ipList[i], tmpIP)
				return
			}
		}
	}
}
