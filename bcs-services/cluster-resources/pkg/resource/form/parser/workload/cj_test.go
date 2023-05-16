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

var lightCJManifest = map[string]interface{}{
	"apiVersion": "apps/v1",
	"kind":       "CronJob",
	"spec": map[string]interface{}{
		"schedule":                   "*/10 * * * *",
		"concurrencyPolicy":          "Forbid",
		"suspend":                    true,
		"successfulJobsHistoryLimit": int64(5),
		"failedJobsHistoryLimit":     int64(3),
		"startingDeadlineSeconds":    int64(600),
		"jobTemplate": map[string]interface{}{
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
		},
	},
}

func TestParseCJ(t *testing.T) {
	formData := ParseCJ(lightCJManifest)
	assert.Equal(t, structs.Map(exceptedContainerGroup), formData["containerGroup"])
	assert.Equal(t, structs.Map(exceptedVolume), formData["volume"])
}

var exceptedCJJobManage = model.CJJobManage{
	Schedule:                   "*/10 * * * *",
	ConcurrencyPolicy:          "Forbid",
	Suspend:                    true,
	Completions:                3,
	Parallelism:                1,
	BackoffLimit:               5,
	ActiveDDLSecs:              720,
	SuccessfulJobsHistoryLimit: 5,
	FailedJobsHistoryLimit:     3,
	StartingDDLSecs:            600,
}

func TestParseCJJobManage(t *testing.T) {
	jm := model.CJJobManage{}
	ParseCJJobManage(lightCJManifest, &jm)
	assert.Equal(t, exceptedCJJobManage, jm)
}
