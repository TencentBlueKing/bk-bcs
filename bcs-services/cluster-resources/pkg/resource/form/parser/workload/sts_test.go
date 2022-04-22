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

package workload

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightSTSManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "StatefulSet",
	"spec": map[string]interface{}{
		"replicas":            int64(3),
		"podManagementPolicy": "OrderedReady",
		"strategy": map[string]interface{}{
			"type": "RollingUpdate",
		},
	},
}

var exceptedSTSReplicas = model.STSReplicas{
	Cnt:            3,
	UpdateStrategy: "RollingUpdate",
	PodManPolicy:   "OrderedReady",
}

func TestParseSTSReplicas(t *testing.T) {
	replicas := model.STSReplicas{}
	ParseSTSReplicas(lightSTSManifest, &replicas)
	assert.Equal(t, exceptedSTSReplicas, replicas)
}
