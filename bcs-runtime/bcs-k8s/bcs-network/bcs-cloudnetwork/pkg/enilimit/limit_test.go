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

package enilimit

import (
	"os"
	"testing"
)

func TestExtraLimitationGetter(t *testing.T) {
	testCases := []struct {
		env     string
		vmType  string
		hasErr  bool
		eniNum  int
		ipNum   int
		isFound bool
	}{
		{
			env:     "{\"c5a.8xlarge\": {\"maxEniNum\": 8, \"maxIPNum\": 8}}",
			vmType:  "c5a.8xlarge",
			hasErr:  false,
			eniNum:  8,
			ipNum:   8,
			isFound: true,
		},
		{
			env:     "{\"c5a.8xlarge\": {\"maxEniNum\": 8, \"maxIPNum\": 8",
			vmType:  "c5a.8xlarge",
			hasErr:  true,
			eniNum:  0,
			ipNum:   0,
			isFound: false,
		},
		{
			env:     "{\"c5a.8xlarge\": {\"maxEniNum\": 8, \"maxIPNum\": 8}}",
			vmType:  "c6i.12xlarge",
			hasErr:  false,
			eniNum:  0,
			ipNum:   0,
			isFound: false,
		},
		{
			env:     "",
			vmType:  "c6i.12xlarge",
			hasErr:  true,
			eniNum:  0,
			ipNum:   0,
			isFound: false,
		},
	}
	for _, test := range testCases {
		os.Setenv(EnvNameForExtraEniLimitation, test.env)
		getter, err := NewGetterFromEnv()
		if err != nil {
			if !test.hasErr {
				t.Fatalf("expect no error but get err %s", err.Error())
			}
			continue
		}
		if test.hasErr {
			t.Fatalf("expect error but get no err")
		}
		eniNum, ipNum, isFound := getter.GetLimit(test.vmType)
		if eniNum != test.eniNum ||
			ipNum != test.ipNum ||
			isFound != test.isFound {
			t.Fatalf("expect %d %d %v, but get %d %d %v",
				test.eniNum, test.ipNum, test.isFound, eniNum, ipNum, isFound)
		}
	}
}
