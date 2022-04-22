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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseDS ...
func ParseDS(manifest map[string]interface{}) map[string]interface{} {
	ds := model.DS{}
	ds.APIVersion, ds.Kind = common.ParseAPIVersionKind(manifest)
	common.ParseMetadata(manifest, &ds.Metadata)
	ParseDSSpec(manifest, &ds.Spec)
	ParseWorkloadVolume(manifest, &ds.Volume)
	ParseContainerGroup(manifest, &ds.ContainerGroup)
	return structs.Map(ds)
}

// ParseDSSpec ...
func ParseDSSpec(manifest map[string]interface{}, spec *model.DSSpec) {
	ParseDSReplicas(manifest, &spec.Replicas)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	ParseNodeSelect(podSpec, &spec.NodeSelect)
	ParseAffinity(podSpec, &spec.Affinity)
	ParseToleration(podSpec, &spec.Toleration)
	ParseNetworking(podSpec, &spec.Networking)
	ParsePodSecurityCtx(podSpec, &spec.Security)
	ParseSpecOther(podSpec, &spec.Other)
}

// ParseDSReplicas ...
func ParseDSReplicas(manifest map[string]interface{}, replicas *model.DSReplicas) {
	replicas.UpdateStrategy = mapx.Get(manifest, "spec.strategy.type", "RollingUpdate").(string)
	maxUnavailable, err := mapx.GetItems(manifest, "spec.strategy.rollingUpdate.maxUnavailable")
	if err == nil {
		replicas.MaxUnavailable, replicas.MUAUnit = util.AnalyzeIntStr(maxUnavailable)
	}
	replicas.MinReadySecs = mapx.Get(manifest, "spec.minReadySeconds", int64(0)).(int64)
}
