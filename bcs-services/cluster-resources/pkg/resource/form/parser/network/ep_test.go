/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package network

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightEPManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "Endpoints",
	"metadata": map[string]interface{}{
		"name": "endpoints-test-8cx47uwj",
	},
	"subsets": []interface{}{
		map[string]interface{}{
			"addresses": []interface{}{
				map[string]interface{}{
					"ip": "1.0.0.1",
				},
				map[string]interface{}{
					"ip": "1.0.0.2",
				},
			},
			"ports": []interface{}{
				map[string]interface{}{
					"port":     int64(8080),
					"protocol": "TCP",
					"name":     "web",
				},
				map[string]interface{}{
					"port":     int64(8090),
					"protocol": "UDP",
					"name":     "abc",
				},
			},
		},
	},
}

var exceptedEPSpec = model.EPSpec{
	SubSets: []model.SubSet{
		{
			Addresses: []string{
				"1.0.0.1",
				"1.0.0.2",
			},
			Ports: []model.EPPort{
				{
					Name:     "web",
					Protocol: "TCP",
					Port:     8080,
				},
				{
					Name:     "abc",
					Protocol: "UDP",
					Port:     8090,
				},
			},
		},
	},
}

func TestParseEPSpec(t *testing.T) {
	actualEPSpec := model.EPSpec{}
	ParseEPSpec(lightEPManifest, &actualEPSpec)
	assert.Equal(t, exceptedEPSpec, actualEPSpec)
}
