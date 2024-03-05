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

package sync

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v2"
)

// TestConfigmapParse test function to parse configmap
func TestConfigmapParse(t *testing.T) {
	testCases := []struct {
		rawData      string
		expectedData string
	}{
		{
			rawData: `nonMasqueradeCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
nonMasqueradeSrcCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
masqLinkLocal: true
resyncInterval: 1m0s`,
			expectedData: `nonMasqueradeCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
- 127.3.0.0/19
nonMasqueradeSrcCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
- 127.3.0.0/19
masqLinkLocal: true
resyncInterval: 1m0s
`,
		},
		{
			rawData: `nonMasqueradeCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
masqLinkLocal: true
resyncInterval: 1m0s`,
			expectedData: `nonMasqueradeCIDRs:
- 127.1.0.0/22
- 127.2.0.0/19
- 127.3.0.0/19
nonMasqueradeSrcCIDRs:
- 127.3.0.0/19
masqLinkLocal: true
resyncInterval: 1m0s
`,
		},
	}
	for _, test := range testCases {
		config := &IPMasqConfig{}
		if err := yaml.UnmarshalStrict([]byte(test.rawData), config); err != nil {
			t.Fatalf(err.Error())
		}
		config.NonMasqueradeCIDRs = append(config.NonMasqueradeCIDRs, "127.3.0.0/19")
		config.NonMasqueradeSrcCIDRs = append(config.NonMasqueradeSrcCIDRs, "127.3.0.0/19")
		outData, err := yaml.Marshal(config)
		if err != nil {
			t.Fatalf(err.Error())
		}
		t.Logf(string(outData))
		if !reflect.DeepEqual(test.expectedData, string(outData)) {
			t.Fatalf("expect %s, but get %s", test.expectedData, string(outData))
		}
	}

}
