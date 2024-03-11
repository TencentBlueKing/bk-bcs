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

package tencentcloud

import (
	"encoding/json"
	"reflect"
	"testing"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// TestSplitListenersToDiffProtocol test function splitListenersToDiffProtocol
func TestSplitListenersToDiffProtocol(t *testing.T) {
	testListeners := []*networkextensionv1.Listener{
		{
			Spec: networkextensionv1.ListenerSpec{
				Port:     8000,
				Protocol: "HTTP",
			},
		},
		{
			Spec: networkextensionv1.ListenerSpec{
				Port:     8001,
				Protocol: "HTTPS",
			},
		},
		{
			Spec: networkextensionv1.ListenerSpec{
				Port:     8002,
				Protocol: "TCP",
			},
		},
		{
			Spec: networkextensionv1.ListenerSpec{
				Port:     8003,
				Protocol: "HTTPS",
			},
		},
		{
			Spec: networkextensionv1.ListenerSpec{
				Port:     8004,
				Protocol: "TCP",
			},
		},
	}
	liGroup := splitListenersToDiffProtocol(testListeners)
	for _, list := range liGroup {
		t.Logf("%+v", list)
		tmpProtocol := make(map[string]struct{})
		for _, li := range list {
			tmpProtocol[li.Spec.Protocol] = struct{}{}
		}
		if len(tmpProtocol) != 1 {
			t.Errorf("list %v contains more than one protocol %v", list, tmpProtocol)
		}
	}
}

// TestSplitListenerToDiffBatch test splitListenersToDiffBatch
func TestSplitListenerToDiffBatch(t *testing.T) {
	testCases := []struct {
		listenerList []*networkextensionv1.Listener
		resultList   [][]*networkextensionv1.Listener
	}{
		{
			listenerList: []*networkextensionv1.Listener{
				{
					Spec: networkextensionv1.ListenerSpec{
						Port:     8000,
						Protocol: "HTTP",
					},
				},
				{
					Spec: networkextensionv1.ListenerSpec{
						Port:     8001,
						Protocol: "HTTP",
					},
				},
				{
					Spec: networkextensionv1.ListenerSpec{
						Port:              8002,
						Protocol:          "HTTP",
						ListenerAttribute: &networkextensionv1.IngressListenerAttribute{},
					},
				},
			},
			resultList: [][]*networkextensionv1.Listener{
				{
					{
						Spec: networkextensionv1.ListenerSpec{
							Port:     8000,
							Protocol: "HTTP",
						},
					},
					{
						Spec: networkextensionv1.ListenerSpec{
							Port:     8001,
							Protocol: "HTTP",
						},
					},
				},
				{
					{
						Spec: networkextensionv1.ListenerSpec{
							Port:              8002,
							Protocol:          "HTTP",
							ListenerAttribute: &networkextensionv1.IngressListenerAttribute{},
						},
					},
				},
			},
		},
	}
	for index, test := range testCases {
		t.Logf("test %d", index)
		tmpList := splitListenersToDiffBatch(test.listenerList)
		if !reflect.DeepEqual(tmpList, test.resultList) {
			tmpListStr, _ := json.Marshal(tmpList)
			resultListStr, _ := json.Marshal(test.resultList)
			t.Errorf("expect %s, but get %s", resultListStr, tmpListStr)
		}
	}
}
