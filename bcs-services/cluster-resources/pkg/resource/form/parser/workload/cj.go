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

// ParseCJ ...
func ParseCJ(manifest map[string]interface{}) map[string]interface{} {
	cj := model.CJ{}
	cj.APIVersion, cj.Kind = common.ParseAPIVersionKind(manifest)
	common.ParseMetadata(manifest, &cj.Metadata)
	ParseCJSpec(manifest, &cj.Spec)
	ParseWorkloadVolume(manifest, &cj.Volume)
	ParseContainerGroup(manifest, &cj.ContainerGroup)
	return structs.Map(cj)
}

// ParseCJSpec ...
func ParseCJSpec(manifest map[string]interface{}, spec *model.CJSpec) {
	ParseCJJobManage(manifest, &spec.JobManage)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.jobTemplate.spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseCJJobManage ...
func ParseCJJobManage(manifest map[string]interface{}, jm *model.CJJobManage) {
	jm.Schedule = mapx.Get(manifest, "spec.schedule", "").(string)
	jm.ConcurrencyPolicy = mapx.Get(manifest, "spec.concurrencyPolicy", "").(string)
	jm.Suspend = mapx.Get(manifest, "spec.suspend", false).(bool)
	jm.Completions = mapx.Get(manifest, "spec.jobTemplate.spec.completions", int64(0)).(int64)
	jm.Parallelism = mapx.Get(manifest, "spec.jobTemplate.spec.parallelism", int64(0)).(int64)
	jm.BackoffLimit = mapx.Get(manifest, "spec.jobTemplate.spec.backoffLimit", int64(0)).(int64)
	jm.ActiveDDLSecs = mapx.Get(manifest, "spec.jobTemplate.spec.activeDeadlineSeconds", int64(0)).(int64)
	jm.SuccessfulJobsHistoryLimit = mapx.Get(manifest, "spec.successfulJobsHistoryLimit", int64(0)).(int64)
	jm.FailedJobsHistoryLimit = mapx.Get(manifest, "spec.failedJobsHistoryLimit", int64(0)).(int64)
	jm.StartingDDLSecs = mapx.Get(manifest, "spec.startingDeadlineSeconds", int64(0)).(int64)
}
