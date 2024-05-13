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

package aws

import (
	"reflect"
	"testing"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
)

func TestSplitPortMappings(t *testing.T) {
	support := &AgaSupporter{}
	testCases := []struct {
		title        string
		portPairs    []portPair
		expectedData []*portMapping
	}{
		{
			title: "单个连续Port段",
			portPairs: []portPair{
				{
					CloudPort: 1081,
					LocalPort: 81,
				},
				{
					CloudPort: 1080,
					LocalPort: 80,
				},
				{
					CloudPort: 1082,
					LocalPort: 82,
				},
			},
			expectedData: []*portMapping{
				{
					CloudStartPort: 1080,
					CloudEndPort:   1082,
					LocalStartPort: 80,
					LocalEndPort:   82,
				},
			},
		},
		{
			title: "多个连续Port段",
			portPairs: []portPair{
				{
					CloudPort: 1081,
					LocalPort: 81,
				},
				{
					CloudPort: 1080,
					LocalPort: 80,
				},
				{
					CloudPort: 1082,
					LocalPort: 82,
				},
				// ---
				{
					CloudPort: 2091,
					LocalPort: 91,
				},
				{
					CloudPort: 2090,
					LocalPort: 90,
				},
				{
					CloudPort: 2092,
					LocalPort: 92,
				},
				// ---
				{
					CloudPort: 1085,
					LocalPort: 85,
				},
			},
			expectedData: []*portMapping{
				{
					CloudStartPort: 1080,
					CloudEndPort:   1082,
					LocalStartPort: 80,
					LocalEndPort:   82,
				},
				{
					CloudStartPort: 1085,
					CloudEndPort:   1085,
					LocalStartPort: 85,
					LocalEndPort:   85,
				},
				{
					CloudStartPort: 2090,
					CloudEndPort:   2092,
					LocalStartPort: 90,
					LocalEndPort:   92,
				},
			},
		},
		{
			title: "单独端口",
			portPairs: []portPair{
				{
					CloudPort: 1081,
					LocalPort: 81,
				},
			},
			expectedData: []*portMapping{
				{
					CloudStartPort: 1081,
					CloudEndPort:   1081,
					LocalStartPort: 81,
					LocalEndPort:   81,
				},
			},
		},
	}

	for index, tCase := range testCases {
		t.Logf("test %d - title: %s", index, tCase.title)
		portMappings := support.splitPortMappings(tCase.portPairs)
		if !reflect.DeepEqual(portMappings, tCase.expectedData) {
			t.Errorf("test %d - title: %s failed, expectedData: %s, result: %s", index, tCase.title,
				common.ToJsonString(tCase.expectedData), common.ToJsonString(portMappings))
		}
	}
}
