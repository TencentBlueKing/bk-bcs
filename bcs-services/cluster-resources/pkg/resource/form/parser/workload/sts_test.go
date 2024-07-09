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

package workload

import (
	"testing"

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightSTSManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "StatefulSet",
	"spec": map[string]interface{}{
		"replicas":            int64(3),
		"podManagementPolicy": "OrderedReady",
		"updateStrategy": map[string]interface{}{
			"type": "OnDelete",
		},
		"volumeClaimTemplates": []interface{}{
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name": "pvc-123",
				},
				"spec": map[string]interface{}{
					"volumeName": "pv-123",
					"accessModes": []interface{}{
						"ROX",
						"RWX",
					},
				},
			},
			map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "PersistentVolumeClaim",
				"metadata": map[string]interface{}{
					"name": "pvc-456",
				},
				"spec": map[string]interface{}{
					"resources": map[string]interface{}{
						"requests": map[string]interface{}{
							"storage": "10Gi",
						},
					},
					"storageClassName": "sc-123",
				},
			},
		},
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"initContainers": containerConf4Test,
				"containers":     containerConf4Test,
				"volumes":        volumeConf4Test,
			},
		},
	},
}

func TestParseSTS(t *testing.T) {
	formData := ParseSTS(lightSTSManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedSTSReplicas = model.STSReplicas{
	Cnt:            "3",
	UpdateStrategy: "OnDelete",
	PodManPolicy:   "OrderedReady",
}

func TestParseSTSReplicas(t *testing.T) {
	replicas := model.STSReplicas{}
	ParseSTSReplicas(lightSTSManifest, &replicas)
	assert.Equal(t, exceptedSTSReplicas, replicas)
}

var exceptedSTSVolumeClaimTmpl = model.STSVolumeClaimTmpl{
	Claims: []model.VolumeClaim{
		{
			PVCName:     "pvc-123",
			ClaimType:   resCsts.PVCTypeUseExistPV,
			PVName:      "pv-123",
			SCName:      "",
			StorageSize: 0,
			AccessModes: []string{"ROX", "RWX"},
		},
		{
			PVCName:     "pvc-456",
			ClaimType:   resCsts.PVCTypeCreateBySC,
			PVName:      "",
			SCName:      "sc-123",
			StorageSize: 10,
			AccessModes: []string{},
		},
	},
}

func TestParseSTSVolumeClaimTmpl(t *testing.T) {
	claimTmpl := model.STSVolumeClaimTmpl{}
	ParseSTSVolumeClaimTmpl(lightSTSManifest, &claimTmpl)
	assert.Equal(t, exceptedSTSVolumeClaimTmpl, claimTmpl)
}
