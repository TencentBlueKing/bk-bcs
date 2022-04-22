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

// ParseSTS ...
func ParseSTS(manifest map[string]interface{}) map[string]interface{} {
	sts := model.STS{}
	sts.APIVersion, sts.Kind = common.ParseAPIVersionKind(manifest)
	common.ParseMetadata(manifest, &sts.Metadata)
	ParseSTSSpec(manifest, &sts.Spec)
	ParseWorkloadVolume(manifest, &sts.Volume)
	ParseContainerGroup(manifest, &sts.ContainerGroup)
	return structs.Map(sts)
}

// ParseSTSSpec ...
func ParseSTSSpec(manifest map[string]interface{}, spec *model.STSSpec) {
	ParseSTSReplicas(manifest, &spec.Replicas)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseSTSReplicas ...
func ParseSTSReplicas(manifest map[string]interface{}, replicas *model.STSReplicas) {
	replicas.Cnt = mapx.Get(manifest, "spec.replicas", int64(0)).(int64)
	replicas.UpdateStrategy = mapx.Get(manifest, "spec.strategy.type", "RollingUpdate").(string)
	replicas.PodManPolicy = mapx.Get(manifest, "spec.podManagementPolicy", "").(string)
}
