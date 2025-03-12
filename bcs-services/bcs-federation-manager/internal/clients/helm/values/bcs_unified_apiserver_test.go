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

// Package values xxx
package values

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

// Test the NewBcsUnifiedAPIServerValues function
func TestNewBcsUnifiedAPIServerValues(t *testing.T) {
	values := NewBcsUnifiedAPIServerValues()
	if values == nil {
		t.Fatal("Expected non-nil BcsUnifiedAPIServerValues instance")
	}
}

// Test the Yaml method of BcsUnifiedAPIServerValues
func TestBcsUnifiedAPIServerValues_Yaml(t *testing.T) {
	values := NewBcsUnifiedAPIServerValues()
	values.Config.BcsConf.Host = "127.0.0.1"
	values.Config.BcsConf.Token = "secret-token"
	values.Config.BcsConf.JwtPublicKey = "public-key"
	values.Config.Apiserver.FederationHostClusterId = "BCS-K8S-00000"
	values.Config.Apiserver.StoreMode = "bcs_storage"
	values.Config.Apiserver.WebhookAddress = "http://webhook1"
	values.SetLoadbalancerId("subnet-12345")
	values.SetUserToken("token")
	values.SetFederationClusterId("BCS-K8S-99999")

	yamlStr := values.Yaml()

	// Print yamlStr to console
	t.Logf("Generated YAML:\n%s", yamlStr)

	var unmarshalledValues BcsUnifiedAPIServerValues
	if err := yaml.Unmarshal([]byte(yamlStr), &unmarshalledValues); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	if !reflect.DeepEqual(values, &unmarshalledValues) {
		t.Errorf("Expected unmarshalled values to be equal, got %+v, want %+v", unmarshalledValues, values)
	}
}
