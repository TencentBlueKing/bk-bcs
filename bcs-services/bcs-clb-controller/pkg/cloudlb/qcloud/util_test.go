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

package qcloud

import (
	"reflect"
	"testing"

	loadbalance "bk-bcs/bcs-services/bcs-clb-controller/pkg/apis/network/v1"
)

func TestGetClusterIDPostfix(t *testing.T) {
	cases := []struct {
		info   string
		inStr  string
		outStr string
	}{
		{
			info:   "",
			inStr:  "BCS-MESOS-3344",
			outStr: "3344",
		},
		{
			info:   "",
			inStr:  "BCS-MESOS-",
			outStr: "",
		},
		{
			info:   "",
			inStr:  "xxxxxxx",
			outStr: "xxxxxxx",
		},
	}
	for i, c := range cases {
		t.Logf("index %d, info %s", i, c.info)
		tmpStr := GetClusterIDPostfix(c.inStr)
		if tmpStr != c.outStr {
			t.Errorf("expect %s, but get %s", c.outStr, tmpStr)
		} else {
			t.Log("success")
		}
	}
}

func TestGetBackendSegement(t *testing.T) {
	cases := []struct {
		info         string
		inSlice      []*loadbalance.Backend
		inCur        int
		inSegmentLen int
		outSlice     []*loadbalance.Backend
	}{
		{
			info:         "nil slice",
			inSlice:      nil,
			inCur:        1,
			inSegmentLen: 10,
			outSlice:     nil,
		},
		{
			info: "normal test",
			inSlice: []*loadbalance.Backend{
				&loadbalance.Backend{
					IP: "a",
				},
				&loadbalance.Backend{
					IP: "b",
				},
				&loadbalance.Backend{
					IP: "c",
				},
				&loadbalance.Backend{
					IP: "d",
				},
				&loadbalance.Backend{
					IP: "e",
				},
			},
			inCur:        1,
			inSegmentLen: 3,
			outSlice: []*loadbalance.Backend{
				&loadbalance.Backend{
					IP: "b",
				},
				&loadbalance.Backend{
					IP: "c",
				},
				&loadbalance.Backend{
					IP: "d",
				},
			},
		},
		{
			info: "normal test",
			inSlice: []*loadbalance.Backend{
				&loadbalance.Backend{
					IP: "a",
				},
				&loadbalance.Backend{
					IP: "b",
				},
				&loadbalance.Backend{
					IP: "c",
				},
				&loadbalance.Backend{
					IP: "d",
				},
				&loadbalance.Backend{
					IP: "e",
				},
			},
			inCur:        2,
			inSegmentLen: 5,
			outSlice: []*loadbalance.Backend{
				&loadbalance.Backend{
					IP: "c",
				},
				&loadbalance.Backend{
					IP: "d",
				},
				&loadbalance.Backend{
					IP: "e",
				},
			},
		},
	}

	for i, c := range cases {
		t.Logf("index %d, info %s", i, c.info)
		ret := GetBackendsSegment(c.inSlice, c.inCur, c.inSegmentLen)
		if !reflect.DeepEqual(ret, c.outSlice) {
			t.Errorf("expect %v, but get %v", c.outSlice, ret)
		} else {
			t.Log("success")
		}
	}
}
