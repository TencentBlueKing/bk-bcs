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

package conflicthandler

import (
	"testing"
)

func TestIsConflict(t *testing.T) {
	testCases := []struct {
		res           resource
		otherRes      resource
		isTCPUDPReuse bool
		hasConflict   bool
	}{
		{
			res: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"UDP"},
					},
				},
			},
			isTCPUDPReuse: false,
			hasConflict:   true,
		},
		{
			res: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"UDP"},
					},
				},
			},
			isTCPUDPReuse: true,
			hasConflict:   false,
		},
		{
			res: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPortSegment: []portSegment{
					{
						Start:     49080,
						End:       49090,
						Protocols: []string{"UDP"},
					},
				},
			},
			isTCPUDPReuse: false,
			hasConflict:   true,
		},
		{
			res: resource{
				usedPort: map[int]portStruct{
					49085: {
						val:       49085,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPortSegment: []portSegment{
					{
						Start:     49080,
						End:       49090,
						Protocols: []string{"UDP"},
					},
				},
			},
			isTCPUDPReuse: true,
			hasConflict:   false,
		},
		{
			res: resource{
				usedPort: map[int]portStruct{
					49080: {
						val:       49080,
						Protocols: []string{"TCP"},
					},
				},
				usedPortSegment: []portSegment{
					{
						Start:     30000,
						End:       30001,
						Protocols: []string{"TCP"},
					},
					{
						Start:     30002,
						End:       30003,
						Protocols: []string{"TCP"},
					},
					{
						Start:     30004,
						End:       30005,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPortSegment: []portSegment{
					{
						Start:     29998,
						End:       30002,
						Protocols: []string{"TCP"},
					},
				},
			},
			isTCPUDPReuse: true,
			hasConflict:   true,
		},
		{
			res: resource{
				usedPort: map[int]portStruct{
					49080: {
						val:       49080,
						Protocols: []string{"TCP"},
					},
				},
				usedPortSegment: []portSegment{
					{
						Start:     30000,
						End:       30001,
						Protocols: []string{"TCP"},
					},
					{
						Start:     30002,
						End:       30003,
						Protocols: []string{"TCP"},
					},
					{
						Start:     30004,
						End:       30005,
						Protocols: []string{"TCP"},
					},
				},
			},
			otherRes: resource{
				usedPortSegment: []portSegment{
					{
						Start:     29998,
						End:       30002,
						Protocols: []string{"HTTP"},
					},
				},
			},
			isTCPUDPReuse: true,
			hasConflict:   true,
		},
	}

	for idx, testCase := range testCases {
		res := &testCase.res
		otherRes := &testCase.otherRes
		if res.IsConflict(testCase.isTCPUDPReuse, otherRes) != testCase.hasConflict {
			t.Errorf("idx %d, should conflict %v", idx, testCase.hasConflict)
		}
	}
}
