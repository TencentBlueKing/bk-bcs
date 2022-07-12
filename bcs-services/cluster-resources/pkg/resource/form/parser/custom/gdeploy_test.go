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

	"github.com/stretchr/testify/assert"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

var lightGDeployManifest = map[string]interface{}{
	"apiVersion": "tkex.tencent.com/v1alpha1",
	"kind":       "GameDeployment",
	"metadata": map[string]interface{}{
		"annotations": map[string]interface{}{
			"io.tencent.bcs.editFormat": "form",
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
	},
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

var exceptedGDeployGracefulManage = model.GDeployGracefulManage{
	PreDeleteHook: model.GDeployHookSpec{
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
	PreInplaceHook: model.GDeployHookSpec{
		Enabled:  true,
		TmplName: "hooktemplate-99ifkar7",
		Args: []model.HookCallArg{
			{
				Key:   "789",
				Value: "012",
			},
		},
	},
	PostInplaceHook: model.GDeployHookSpec{
		Enabled: false,
	},
}

func TestGDeployGracefulManage(t *testing.T) {
	actualGDeployGracefulManage := model.GDeployGracefulManage{}
	ParseGDeployGracefulManage(lightGDeployManifest, &actualGDeployGracefulManage)
	assert.Equal(t, exceptedGDeployGracefulManage, actualGDeployGracefulManage)
}

func TestParseGDeployDeletionProtect(t *testing.T) {
	deletionProtect := model.GDeployDeletionProtect{}
	ParseGDeployDeletionProtect(lightGDeployManifest, &deletionProtect)
	assert.Equal(t, res.DeletionProtectPolicyCascading, deletionProtect.Policy)

	_ = mapx.SetItems(lightGDeployManifest, "metadata.labels", map[string]interface{}{})
	ParseGDeployDeletionProtect(lightGDeployManifest, &deletionProtect)
	assert.Equal(t, res.DeletionProtectPolicyNotAllow, deletionProtect.Policy)
}
