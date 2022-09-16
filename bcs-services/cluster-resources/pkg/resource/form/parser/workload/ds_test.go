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

	"github.com/fatih/structs"
	"github.com/stretchr/testify/assert"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
)

var lightDSManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "DaemonSet",
	"spec": map[string]interface{}{
		"minReadySeconds": int64(60),
		"updateStrategy": map[string]interface{}{
			"rollingUpdate": map[string]interface{}{
				"maxUnavailable": "25%",
			},
			"type": "RollingUpdate",
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

func TestParseDS(t *testing.T) {
	formData := ParseDS(lightDSManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedDSReplicas = model.DSReplicas{
	UpdateStrategy: "RollingUpdate",
	MaxUnavailable: 25,
	MUAUnit:        "percent",
	MinReadySecs:   60,
}

func TestParseDSReplicas(t *testing.T) {
	replicas := model.DSReplicas{}
	ParseDSReplicas(lightDSManifest, &replicas)
	assert.Equal(t, exceptedDSReplicas, replicas)
}
