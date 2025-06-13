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
	"os"
	"reflect"
	"strings"
	"testing"

	"gopkg.in/yaml.v2"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/pkg/common"
	meshmanager "github.com/Tencent/bk-bcs/bcs-services/bcs-mesh-manager/proto/bcs-mesh-manager"
)

func TestMergeValues(t *testing.T) {
	cases := []struct {
		name          string
		defaultValues string
		customValues  string
		expect        string
	}{
		{
			name: "simple merge",
			defaultValues: `a: 1
b: 2`,
			customValues: `b: 3
c: 4`,
			expect: "a: 1\nb: 3\nc: 4\n",
		},
		{
			name: "nested merge",
			defaultValues: `a:
  b: 1
  c: 2
x: 5`,
			customValues: `a:
  c: 3
d: 4`,
			expect: "a:\n  b: 1\n  c: 3\nx: 5\nd: 4\n",
		},
		{
			name: "custom overrides default",
			defaultValues: `foo: bar
bar: baz`,
			customValues: `foo: newbar`,
			expect:       "foo: newbar\nbar: baz\n",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			result, err := MergeValues(c.defaultValues, c.customValues)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			var expectMap, resultMap map[string]interface{}
			if err := yaml.Unmarshal([]byte(c.expect), &expectMap); err != nil {
				t.Fatalf("unmarshal expect failed: %v", err)
			}
			if err := yaml.Unmarshal([]byte(result), &resultMap); err != nil {
				t.Fatalf("unmarshal result failed: %v", err)
			}
			if !reflect.DeepEqual(expectMap, resultMap) {
				t.Errorf("merge failed.\nExpected:\n%v\nGot:\n%v", expectMap, resultMap)
			}
		})
	}
}

func TestGenIstiodValues(t *testing.T) {
	dir := t.TempDir()
	// mock values.yaml 文件
	istiodYaml := "global:\n  bar: baz\nmeshConfig:\n  outboundTrafficPolicy:\n    mode: ALLOW_ANY"
	os.MkdirAll(dir+"/1.24", 0755)
	os.WriteFile(dir+"/1.24/istiod-values.yaml", []byte(istiodYaml), 0644)

	featureConfigs := map[string]*meshmanager.FeatureConfig{
		"outboundTrafficPolicy": {
			Name:  "outboundTrafficPolicy",
			Value: "REGISTRY_ONLY",
		},
	}
	result, err := GenIstiodValues(
		dir,
		common.IstioInstallModePrimary,
		"1.24",
		"primary-cluster",
		"",
		"test-cluster",
		"mesh-123",
		"net-456",
		featureConfigs,
	)
	t.Logf("result: %s", result)
	if err != nil {
		t.Fatalf("GenIstiodValues error: %v", err)
	}
	if !strings.Contains(result, "mesh-123") || !strings.Contains(result, "net-456") {
		t.Errorf("result missing meshID or networkID: %s", result)
	}
	if !strings.Contains(result, "REGISTRY_ONLY") {
		t.Errorf("feature config not merged: %s", result)
	}
}
