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

package custom

import (
	"testing"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
)

var lightGSTSManifest = map[string]interface{}{
	"apiVersion": "tkex.tencent.com/v1alpha1",
	"kind":       "GameStatefulSet",
	"metadata": map[string]interface{}{
		"name":      "gamestatefulset-okvggfvh",
		"namespace": "default",
	},
	"spec": map[string]interface{}{
		"replicas": int64(1),
		"updateStrategy": map[string]interface{}{
			"inPlaceUpdateStrategy": map[string]interface{}{
				"gracePeriodSeconds": int64(30),
			},
			"rollingUpdate": map[string]interface{}{
				"maxSurge":       int64(0),
				"maxUnavailable": "20%",
				"partition":      int64(1),
			},
			"type": "InplaceUpdate",
		},
		"serviceName":         "service-au8j3kel",
		"podManagementPolicy": "Parallel",
		"template":            lightPodTmpl,
	},
}

func TestParseGSTS(t *testing.T) {
	formData := ParseGSTS(lightGSTSManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedGSTSReplicas = model.GSTSReplicas{
	Cnt:             1,
	SVCName:         "service-au8j3kel",
	UpdateStrategy:  "InplaceUpdate",
	PodManPolicy:    "Parallel",
	MaxSurge:        0,
	MSUnit:          util.UnitCnt,
	MaxUnavailable:  20,
	MUAUnit:         util.UnitPercent,
	Partition:       1,
	GracePeriodSecs: 30,
}

func TestParseGSTSReplicas(t *testing.T) {
	actualGSTSReplicas := model.GSTSReplicas{}
	ParseGSTSReplicas(lightGSTSManifest, &actualGSTSReplicas)
	assert.Equal(t, exceptedGSTSReplicas, actualGSTSReplicas)
}
