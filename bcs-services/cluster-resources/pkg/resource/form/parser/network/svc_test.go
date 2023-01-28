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

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightSVCManifest = map[string]interface{}{
	"apiVersion": "v1",
	"kind":       "Service",
	"metadata": map[string]interface{}{
		"name": "service-test-12345",
		"annotations": map[string]interface{}{
			resCsts.SVCExistLBIDAnnoKey: "lb-abcd",
			resCsts.SVCSubNetIDAnnoKey:  "subnet-id-1234",
		},
	},
	"spec": map[string]interface{}{
		"type": "ClusterIP",
		"ports": []interface{}{
			map[string]interface{}{
				"name":       "aaa",
				"port":       int64(80),
				"protocol":   "TCP",
				"targetPort": "http",
			},
			map[string]interface{}{
				"name":       "bbb",
				"port":       int64(81),
				"protocol":   "TCP",
				"targetPort": int64(8081),
				"nodePort":   int64(30000),
			},
			map[string]interface{}{
				"name":       "ccc",
				"port":       int64(82),
				"protocol":   "UDP",
				"targetPort": int64(8082),
			},
		},
		"sessionAffinity": "ClientIP",
		"sessionAffinityConfig": map[string]interface{}{
			"clientIP": map[string]interface{}{
				"timeoutSeconds": int64(10800),
			},
		},
		"clusterIP": "1.1.1.1",
		"externalIPs": []interface{}{
			"2.2.2.2",
			"3.3.3.3",
		},
		"selector": map[string]interface{}{
			"select-123": "456",
		},
	},
}

var exceptedSVCSpec = model.SVCSpec{
	PortConf: model.SVCPortConf{
		Type: "ClusterIP",
		LB: model.SVCLB{
			UseType:   resCsts.CLBUseTypeAutoCreate,
			ExistLBID: "lb-abcd",
			SubNetID:  "subnet-id-1234",
		},
		Ports: []model.SVCPort{
			{
				Name:       "aaa",
				Port:       80,
				Protocol:   "TCP",
				TargetPort: "http",
			},
			{
				Name:       "bbb",
				Port:       81,
				Protocol:   "TCP",
				TargetPort: "8081",
				NodePort:   30000,
			},
			{
				Name:       "ccc",
				Port:       82,
				Protocol:   "UDP",
				TargetPort: "8082",
			},
		},
	},
	Selector: model.SVCSelector{
		Labels: []model.LabelSelector{
			{
				Key:   "select-123",
				Value: "456",
			},
		},
	},
	SessionAffinity: model.SessionAffinity{
		Type:       resCsts.SessionAffinityTypeClientIP,
		StickyTime: 10800,
	},
	IP: model.IPConf{
		Address: "1.1.1.1",
		External: []string{
			"2.2.2.2",
			"3.3.3.3",
		},
	},
}

func TestParseSVCSpec(t *testing.T) {
	actualSVCSpec := model.SVCSpec{}
	ParseSVCSpec(lightSVCManifest, &actualSVCSpec)
	assert.Equal(t, exceptedSVCSpec, actualSVCSpec)
}
