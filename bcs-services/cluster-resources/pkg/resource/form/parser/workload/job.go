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
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseJob xxx
func ParseJob(manifest map[string]interface{}) map[string]interface{} {
	job := model.Job{}
	common.ParseMetadata(manifest, &job.Metadata)
	ParseJobSpec(manifest, &job.Spec)
	ParseWorkloadVolume(manifest, &job.Volume)
	ParseContainerGroup(manifest, &job.ContainerGroup)
	return structs.Map(job)
}

// ParseJobSpec xxx
func ParseJobSpec(manifest map[string]interface{}, spec *model.JobSpec) {
	ParseJobManage(manifest, &spec.JobManage)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseJobManage xxx
func ParseJobManage(manifest map[string]interface{}, jm *model.JobManage) {
	jm.Completions = mapx.GetInt64(manifest, "spec.completions")
	jm.Parallelism = mapx.GetInt64(manifest, "spec.parallelism")
	jm.BackoffLimit = mapx.GetInt64(manifest, "spec.backoffLimit")
	jm.ActiveDDLSecs = mapx.GetInt64(manifest, "spec.activeDeadlineSeconds")
}
