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

package utils

import (
	"math"
	"testing"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"

	k8scorev1 "k8s.io/api/core/v1"
	k8sresource "k8s.io/apimachinery/pkg/api/resource"
)

// TestCalculateResourceAllocRate test calculateResourceAllocRate
func TestCalculateResourceAllocRate(t *testing.T) {
	nodeList := &k8scorev1.NodeList{
		Items: []k8scorev1.Node{
			{
				Status: k8scorev1.NodeStatus{
					Allocatable: k8scorev1.ResourceList{
						// don't use MustParse in logic code, may fatal
						"cpu": k8sresource.MustParse("2000m"),
					},
				},
			},
			{
				Status: k8scorev1.NodeStatus{
					Allocatable: k8scorev1.ResourceList{
						"cpu": k8sresource.MustParse("1000m"),
					},
				},
			},
		},
	}

	quotaList := []types.ResourceQuota{
		{
			Namespace: "test",
			ClusterID: "test-cluster",
			ResourceQuota: `
{
	"apiVersion":"v1",
	"kind":"ResourceQuota",
	"metadata": {
		"name":"test"
	},
	"spec": {
		"hard":{
			"cpu":"1",
			"memory":"2Gi",
			"limits.cpu":"2",
			"limits.memory":"2Gi"
		}
	}
}`,
		},
	}

	rate, err := CalculateResourceAllocRate(quotaList, nodeList)
	if err != nil {
		t.Error(err.Error())
	}
	if math.Abs(float64(rate-0.33333)) > 0.001 {
		t.Errorf("expected rate %f but get rate %f", 0.33333, rate)
	}
}
