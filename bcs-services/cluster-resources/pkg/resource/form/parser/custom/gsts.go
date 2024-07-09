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

package custom

import (
	"github.com/fatih/structs"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseGSTS GameDeployment manifest -> formData
func ParseGSTS(manifest map[string]interface{}) map[string]interface{} {
	deploy := model.GSTS{}
	common.ParseMetadata(manifest, &deploy.Metadata)
	ParseGSTSSpec(manifest, &deploy.Spec)
	workload.ParseWorkloadVolume(manifest, &deploy.Volume)
	workload.ParseContainerGroup(manifest, &deploy.ContainerGroup)
	return structs.Map(deploy)
}

// ParseGSTSSpec xxx
func ParseGSTSSpec(manifest map[string]interface{}, spec *model.GSTSSpec) {
	selectorLables, _ := mapx.GetItems(manifest, "spec.selector.matchLabels")
	common.ParseLabels(selectorLables, &spec.Labels.SelectorLabels)
	templateLables, _ := mapx.GetItems(manifest, "spec.template.metadata.labels")
	common.ParseLabels(templateLables, &spec.Labels.TemplateLabels)
	ParseGSTSReplicas(manifest, &spec.Replicas)
	ParseGWorkloadGracefulManage(manifest, &spec.GracefulManage)
	ParseGWorkloadDeletionProtect(manifest, &spec.DeletionProtect)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	workload.ParseNodeSelect(podSpec, &spec.NodeSelect)
	workload.ParseAffinity(podSpec, &spec.Affinity)
	workload.ParseToleration(podSpec, &spec.Toleration)
	workload.ParseNetworking(podSpec, &spec.Networking)
	workload.ParsePodSecurityCtx(podSpec, &spec.Security)
	workload.ParseSpecOther(podSpec, &spec.Other)
}

// ParseGSTSReplicas xxx
func ParseGSTSReplicas(manifest map[string]interface{}, replicas *model.GSTSReplicas) {
	replicas.Cnt = mapx.GetInt64(manifest, "spec.replicas")
	replicas.SVCName = mapx.GetStr(manifest, "spec.serviceName")
	replicas.UpdateStrategy = mapx.Get(
		manifest, "spec.updateStrategy.type", resCsts.DefaultUpdateStrategy,
	).(string)
	replicas.PodManPolicy = mapx.Get(manifest, "spec.podManagementPolicy", "OrderedReady").(string)
	replicas.MaxSurge, replicas.MSUnit = resCsts.DefaultGWorkloadMaxSurge, util.UnitCnt
	if maxSurge, err := mapx.GetItems(manifest, "spec.updateStrategy.rollingUpdate.maxSurge"); err == nil {
		replicas.MaxSurge, replicas.MSUnit = util.AnalyzeIntStr(maxSurge)
	}
	replicas.MaxUnavailable, replicas.MUAUnit = resCsts.DefaultGWorkloadMaxUnavailable, util.UnitPercent
	if maxUnavailable, err := mapx.GetItems(
		manifest, "spec.updateStrategy.rollingUpdate.maxUnavailable",
	); err == nil {
		replicas.MaxUnavailable, replicas.MUAUnit = util.AnalyzeIntStr(maxUnavailable)
	}
	replicas.Partition = mapx.GetInt64(manifest, "spec.updateStrategy.rollingUpdate.partition")
	replicas.GracePeriodSecs = mapx.Get(
		manifest,
		"spec.updateStrategy.inPlaceUpdateStrategy.gracePeriodSeconds",
		int64(resCsts.DefaultGWorkloadMaxSurge),
	).(int64)
}
