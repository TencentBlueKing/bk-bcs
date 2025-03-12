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

// Package kubeconfig ...
package kubeconfig

import (
	"fmt"
	"testing"

	"gopkg.in/yaml.v2"
)

// TestKubeconfig ...
func TestKubeconfig(t *testing.T) {
	// 定义测试用例
	tests := []struct {
		address   string
		expectErr bool
	}{
		{"https://127.0.0.1:6443", false},
		{"", false}, // 即使地址为空，也应生成有效的 kubeconfig
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("Address: %s", test.address), func(t *testing.T) {
			yamlString := NewConfigForRegister(test.address).Yaml()

			t.Logf("yamlString: \n%s\n", yamlString)

			// 验证生成的 YAML 是否包含关键字段
			var config Config
			err := yaml.Unmarshal([]byte(yamlString), &config)
			if err != nil {
				t.Fatalf("Failed to unmarshal YAML: %v", err)
			}

			if config.Clusters[0].Cluster.Server != test.address {
				t.Errorf("Expected server address: %s, got: %s", test.address, config.Clusters[0].Cluster.Server)
			}

			if config.Users[0].User.Token != "xxxxxxx" {
				t.Errorf("Expected token: xxxxxxx, got: %s", config.Users[0].User.Token)
			}

			if config.Contexts[0].Context.Cluster != "fed-cluster" {
				t.Errorf("Expected cluster: fed-cluster, got: %s", config.Contexts[0].Context.Cluster)
			}

			if config.CurrentContext != "fed-context" {
				t.Errorf("Expected current context: fed-context, got: %v\n", config.CurrentContext)
			}

		})
	}
}
