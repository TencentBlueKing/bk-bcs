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

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

var lightPodTmpl = map[string]interface{}{
	"spec": map[string]interface{}{
		"containers": []interface{}{
			map[string]interface{}{
				"image":           "busybox:latest",
				"imagePullPolicy": "IfNotPresent",
				"name":            "busybox",
				"readinessProbe": map[string]interface{}{
					"periodSeconds":    int64(10),
					"timeoutSeconds":   int64(3),
					"successThreshold": int64(1),
					"failureThreshold": int64(3),
					"tcpSocket": map[string]interface{}{
						"port": int64(80),
					},
				},
				"livenessProbe": map[string]interface{}{
					"periodSeconds":    int64(10),
					"timeoutSeconds":   int64(3),
					"successThreshold": int64(1),
					"failureThreshold": int64(3),
					"exec": map[string]interface{}{
						"command": []interface{}{
							"echo hello",
						},
					},
				},
			},
		},
		"volumes": []interface{}{
			map[string]interface{}{
				"name": "nfs",
				"nfs": map[string]interface{}{
					"path":   "/data",
					"server": "1.1.1.1",
				},
			},
		},
	},
}

var lightGDeployManifest = map[string]interface{}{
	"apiVersion": "tkex.tencent.com/v1alpha1",
	"kind":       "GameDeployment",
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			resCsts.EditModeAnnoKey: "form",
		},
		"labels": map[string]interface{}{
			"io.tencent.bcs.dev/deletion-allow": "Cascading",
		},
		"name":      "gamedeployment-vokvggfh",
		"namespace": "default",
	},
	"spec": map[string]interface{}{
		"minReadySeconds": int64(0),
		"preDeleteUpdateStrategy": map[string]interface{}{
			"hook": map[string]interface{}{
				"args": []interface{}{
					map[string]interface{}{
						"name":  "123",
						"value": "345",
					},
					map[string]interface{}{
						"name":  "456",
						"value": "789",
					},
				},
				"templateName": "hooktemplate-4mdfd82m",
			},
		},
		"preInplaceUpdateStrategy": map[string]interface{}{
			"hook": map[string]interface{}{
				"args": []interface{}{
					map[string]interface{}{
						"name":  "789",
						"value": "012",
					},
				},
				"templateName": "hooktemplate-99ifkar7",
			},
		},
		"replicas": int64(1),
		"updateStrategy": map[string]interface{}{
			"inPlaceUpdateStrategy": map[string]interface{}{
				"gracePeriodSeconds": int64(30),
			},
			"maxSurge":       int64(0),
			"maxUnavailable": "20%",
			"partition":      int64(1),
			"type":           "InplaceUpdate",
		},
		"template": lightPodTmpl,
	},
}

var exceptedContainerGroup = model.ContainerGroup{
	Containers: []model.Container{
		{
			Basic: model.ContainerBasic{
				Name:       "busybox",
				Image:      "busybox:latest",
				PullPolicy: "IfNotPresent",
			},
			Healthz: model.ContainerHealthz{
				ReadinessProbe: model.Probe{
					Enabled:          true,
					PeriodSecs:       10,
					InitialDelaySecs: 0,
					TimeoutSecs:      3,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					Type:             "tcpSocket",
					Port:             80,
				},
				LivenessProbe: model.Probe{
					Enabled:          true,
					PeriodSecs:       10,
					InitialDelaySecs: 0,
					TimeoutSecs:      3,
					SuccessThreshold: 1,
					FailureThreshold: 3,
					Type:             "exec",
					Command:          []string{"echo hello"},
				},
			},
		},
	},
}

var exceptedVolume = model.WorkloadVolume{
	NFS: []model.NFSVolume{
		{
			Name:     "nfs",
			Path:     "/data",
			Server:   "1.1.1.1",
			ReadOnly: false,
		},
	},
}

func TestParseGDeploy(t *testing.T) {
	formData := ParseGDeploy(lightGDeployManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedGDeployReplicas = model.GDeployReplicas{
	Cnt:             1,
	UpdateStrategy:  "InplaceUpdate",
	MaxSurge:        0,
	MSUnit:          util.UnitCnt,
	MaxUnavailable:  20,
	MUAUnit:         util.UnitPercent,
	MinReadySecs:    0,
	Partition:       1,
	GracePeriodSecs: 30,
}

func TestParseGDeployReplicas(t *testing.T) {
	actualGDeployReplicas := model.GDeployReplicas{}
	ParseGDeployReplicas(lightGDeployManifest, &actualGDeployReplicas)
	assert.Equal(t, exceptedGDeployReplicas, actualGDeployReplicas)
}

var exceptedGWorkloadGracefulManage = model.GWorkloadGracefulManage{
	PreDeleteHook: model.GWorkloadHookSpec{
		Enabled:  true,
		TmplName: "hooktemplate-4mdfd82m",
		Args: []model.HookCallArg{
			{
				Key:   "123",
				Value: "345",
			},
			{
				Key:   "456",
				Value: "789",
			},
		},
	},
	PreInplaceHook: model.GWorkloadHookSpec{
		Enabled:  true,
		TmplName: "hooktemplate-99ifkar7",
		Args: []model.HookCallArg{
			{
				Key:   "789",
				Value: "012",
			},
		},
	},
	PostInplaceHook: model.GWorkloadHookSpec{
		Enabled: false,
	},
}

func TestParseGWorkloadGracefulManage(t *testing.T) {
	actualGWorkloadGracefulManage := model.GWorkloadGracefulManage{}
	ParseGWorkloadGracefulManage(lightGDeployManifest, &actualGWorkloadGracefulManage)
	assert.Equal(t, exceptedGWorkloadGracefulManage, actualGWorkloadGracefulManage)
}

func TestParseGWorkloadDeletionProtect(t *testing.T) {
	deletionProtect := model.GWorkloadDeletionProtect{}
	ParseGWorkloadDeletionProtect(lightGDeployManifest, &deletionProtect)
	assert.Equal(t, resCsts.DeletionProtectPolicyCascading, deletionProtect.Policy)

	_ = mapx.SetItems(lightGDeployManifest, "metadata.labels", map[string]interface{}{})
	ParseGWorkloadDeletionProtect(lightGDeployManifest, &deletionProtect)
	assert.Equal(t, resCsts.DeletionProtectPolicyNotAllow, deletionProtect.Policy)
}
