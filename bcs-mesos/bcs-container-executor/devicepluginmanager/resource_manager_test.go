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

package devicepluginmanager

import (
	"reflect"
	"testing"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestGetAllocateDeviceIDs(t *testing.T) {
	testCases := []struct {
		allIDs      []string
		existMap    map[string]string
		resourceNum int
		result      []string
		hasErr      bool
	}{
		{
			allIDs: []string{"1", "2", "3", "4", "5"},
			existMap: map[string]string{
				"2": "container1",
				"3": "container2",
			},
			resourceNum: 2,
			result:      []string{"1", "4"},
			hasErr:      false,
		},
	}
	for _, test := range testCases {
		tmpRes, err := getAllocateDeviceIDs(test.allIDs, test.existMap, test.resourceNum)
		if err != nil && !test.hasErr {
			t.Errorf("expect no err, but get %s err", err.Error())
		}
		if err == nil && test.hasErr {
			t.Errorf("expect err, but get no err")
		}
		if !reflect.DeepEqual(tmpRes, test.result) {
			t.Errorf("expect %v, but get %v", test.result, tmpRes)
		}
	}
}

func TestGetAllocateDeviceIDsByTopology(t *testing.T) {
	testCases := []struct {
		devices     []*pluginapi.Device
		existMap    map[string]string
		resourceNum int
		result      []string
		hasErr      bool
	}{
		{
			devices: []*pluginapi.Device{
				{
					ID: "1",
					Topology: &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: 0,
							},
						},
					},
				},
				{
					ID: "2",
					Topology: &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: 0,
							},
						},
					},
				},
				{
					ID: "3",
					Topology: &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: 0,
							},
						},
					},
				},
				{
					ID: "4",
					Topology: &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: 1,
							},
						},
					},
				},
				{
					ID: "5",
					Topology: &pluginapi.TopologyInfo{
						Nodes: []*pluginapi.NUMANode{
							{
								ID: 1,
							},
						},
					},
				},
			},
			existMap: map[string]string{
				"2": "container1",
				"3": "container2",
			},
			resourceNum: 2,
			result:      []string{"4", "5"},
			hasErr:      false,
		},
	}

	for _, test := range testCases {
		tmpRes, err := getAllocateDeviceIDsByTopology(test.devices, test.existMap, test.resourceNum)
		if err != nil && !test.hasErr {
			t.Errorf("expect no err, but get %s err", err.Error())
		}
		if err == nil && test.hasErr {
			t.Errorf("expect err, but get no err")
		}
		if !reflect.DeepEqual(tmpRes, test.result) {
			t.Errorf("expect %v, but get %v", test.result, tmpRes)
		}
	}
}
