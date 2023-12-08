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

package workload

import (
	"github.com/fatih/structs"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseDeploy xxx
func ParseDeploy(manifest map[string]interface{}) map[string]interface{} {
	deploy := model.Deploy{}
	common.ParseMetadata(manifest, &deploy.Metadata)
	ParseDeploySpec(manifest, &deploy.Spec)
	ParseWorkloadVolume(manifest, &deploy.Volume)
	ParseContainerGroup(manifest, &deploy.ContainerGroup)
	return structs.Map(deploy)
}

// ParseDeploySpec xxx
func ParseDeploySpec(manifest map[string]interface{}, spec *model.DeploySpec) {
	ParseDeployReplicas(manifest, &spec.Replicas)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecReadinessGates(podSpec, &spec.ReadinessGates)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseDeployReplicas xxx
func ParseDeployReplicas(manifest map[string]interface{}, replicas *model.DeployReplicas) {
	replicas.Cnt = mapx.GetInt64(manifest, "spec.replicas")
	replicas.UpdateStrategy = mapx.Get(manifest, "spec.strategy.type", resCsts.DefaultUpdateStrategy).(string)
	replicas.MaxSurge, replicas.MSUnit = resCsts.DefaultMaxSurge, util.UnitPercent
	if maxSurge, err := mapx.GetItems(manifest, "spec.strategy.rollingUpdate.maxSurge"); err == nil {
		replicas.MaxSurge, replicas.MSUnit = util.AnalyzeIntStr(maxSurge)
	}
	replicas.MaxUnavailable, replicas.MUAUnit = resCsts.DefaultMaxUnavailable, util.UnitPercent
	if maxUnavailable, err := mapx.GetItems(manifest, "spec.strategy.rollingUpdate.maxUnavailable"); err == nil {
		replicas.MaxUnavailable, replicas.MUAUnit = util.AnalyzeIntStr(maxUnavailable)
	}
	replicas.MinReadySecs = mapx.GetInt64(manifest, "spec.minReadySeconds")
	replicas.ProgressDeadlineSecs = mapx.GetInt64(manifest, "spec.progressDeadlineSeconds")
}
