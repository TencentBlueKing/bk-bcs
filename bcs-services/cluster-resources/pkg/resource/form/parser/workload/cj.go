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

// Package workload xxx
package workload

import (
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseCJ xxx
func ParseCJ(manifest map[string]interface{}) map[string]interface{} {
	cj := model.CJ{}
	common.ParseMetadata(manifest, &cj.Metadata)
	ParseCJSpec(manifest, &cj.Spec)
	ParseWorkloadVolume(manifest, &cj.Volume)
	ParseContainerGroup(manifest, &cj.ContainerGroup)
	return structs.Map(cj)
}

// ParseCJSpec xxx
func ParseCJSpec(manifest map[string]interface{}, spec *model.CJSpec) {
	jobTemplatelabels, _ := mapx.GetItems(manifest, "spec.jobTemplate.metadata.labels")
	common.ParseLabels(jobTemplatelabels, &spec.Labels.JobTemplatelabels)
	templateLables, _ := mapx.GetItems(manifest, "spec.jobTemplate.spec.template.metadata.labels")
	common.ParseLabels(templateLables, &spec.Labels.TemplateLabels)
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

// ParseCJJobManage xxx
func ParseCJJobManage(manifest map[string]interface{}, jm *model.CJJobManage) {
	jm.Schedule = mapx.GetStr(manifest, "spec.schedule")
	jm.ConcurrencyPolicy = mapx.GetStr(manifest, "spec.concurrencyPolicy")
	jm.Suspend = mapx.GetBool(manifest, "spec.suspend")
	jm.Completions = mapx.GetInt64(manifest, "spec.jobTemplate.spec.completions")
	jm.Parallelism = mapx.GetInt64(manifest, "spec.jobTemplate.spec.parallelism")
	jm.BackoffLimit = mapx.GetInt64(manifest, "spec.jobTemplate.spec.backoffLimit")
	jm.ActiveDDLSecs = mapx.GetInt64(manifest, "spec.jobTemplate.spec.activeDeadlineSeconds")
	jm.SuccessfulJobsHistoryLimit = mapx.GetInt64(manifest, "spec.successfulJobsHistoryLimit")
	jm.FailedJobsHistoryLimit = mapx.GetInt64(manifest, "spec.failedJobsHistoryLimit")
	jm.StartingDDLSecs = mapx.GetInt64(manifest, "spec.startingDeadlineSeconds")
}
