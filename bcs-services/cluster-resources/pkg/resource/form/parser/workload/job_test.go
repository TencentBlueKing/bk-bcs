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

var lightJobManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "Job",
	"spec": map[string]interface{}{
		"completions":           int64(3),
		"parallelism":           int64(1),
		"backoffLimit":          int64(5),
		"activeDeadlineSeconds": int64(720),
		"template": map[string]interface{}{
			"spec": map[string]interface{}{
				"initContainers": containerConf4Test,
				"containers":     containerConf4Test,
				"volumes":        volumeConf4Test,
			},
		},
	},
}

func TestParseJob(t *testing.T) {
	formData := ParseJob(lightJobManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedJobManage = model.JobManage{
	Completions:   3,
	Parallelism:   1,
	BackoffLimit:  5,
	ActiveDDLSecs: 720,
}

func TestParseJobManage(t *testing.T) {
	jm := model.JobManage{}
	ParseJobManage(lightJobManifest, &jm)
	assert.Equal(t, exceptedJobManage, jm)
}
