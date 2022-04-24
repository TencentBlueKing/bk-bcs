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

var lightDeployManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Deployment",
	"spec": map[string]interface{}{
		"progressDeadlineSeconds": int64(600),
		"replicas":                int64(3),
		"minReadySeconds":         int64(60),
		"strategy": map[string]interface{}{
			"rollingUpdate": map[string]interface{}{
				"maxSurge":       int64(2),
				"maxUnavailable": "25%",
			},
			"type": "RollingUpdate",
		},
	},
}

var exceptedDeployReplicas = model.DeployReplicas{
	Cnt:                  3,
	UpdateStrategy:       "RollingUpdate",
	MaxSurge:             2,
	MSUnit:               "cnt",
	MaxUnavailable:       25,
	MUAUnit:              "percent",
	MinReadySecs:         60,
	ProgressDeadlineSecs: 600,
}

func TestParseDeployReplicas(t *testing.T) {
	replicas := model.DeployReplicas{}
	ParseDeployReplicas(lightDeployManifest, &replicas)
	assert.Equal(t, exceptedDeployReplicas, replicas)
}
