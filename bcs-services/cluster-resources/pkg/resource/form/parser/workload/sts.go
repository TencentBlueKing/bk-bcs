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

// ParseSTS ...
func ParseSTS(manifest map[string]interface{}) map[string]interface{} {
	sts := model.STS{}
	common.ParseMetadata(manifest, &sts.Metadata)
	ParseSTSSpec(manifest, &sts.Spec)
	ParseWorkloadVolume(manifest, &sts.Volume)
	ParseContainerGroup(manifest, &sts.ContainerGroup)
	return structs.Map(sts)
}

// ParseSTSSpec ...
func ParseSTSSpec(manifest map[string]interface{}, spec *model.STSSpec) {
	selectorLables, _ := mapx.GetItems(manifest, "spec.selector.matchLabels")
	common.ParseLabels(selectorLables, &spec.Labels.SelectorLabels)
	templateLables, _ := mapx.GetItems(manifest, "spec.template.metadata.labels")
	common.ParseLabels(templateLables, &spec.Labels.TemplateLabels)
	ParseSTSReplicas(manifest, &spec.Replicas)
	ParseSTSVolumeClaimTmpl(manifest, &spec.VolumeClaimTmpl)
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

// ParseSTSReplicas ...
func ParseSTSReplicas(manifest map[string]interface{}, replicas *model.STSReplicas) {
	replicas.SVCName = mapx.GetStr(manifest, "spec.serviceName")
	replicas.Cnt = mapx.GetIntStr(manifest, "spec.replicas")
	replicas.UpdateStrategy = mapx.Get(
		manifest, "spec.updateStrategy.type", resCsts.DefaultUpdateStrategy,
	).(string)
	replicas.PodManPolicy = mapx.Get(manifest, "spec.podManagementPolicy", "OrderedReady").(string)
	replicas.Partition = mapx.GetInt64(manifest, "spec.updateStrategy.rollingUpdate.partition")
}

// ParseSTSVolumeClaimTmpl ...
func ParseSTSVolumeClaimTmpl(manifest map[string]interface{}, claimTmpl *model.STSVolumeClaimTmpl) {
	for _, c := range mapx.GetList(manifest, "spec.volumeClaimTemplates") {
		claimTmpl.Claims = append(claimTmpl.Claims, parseVolumeClaim(c.(map[string]interface{})))
	}
}

// 解析卷声明结构
func parseVolumeClaim(raw map[string]interface{}) model.VolumeClaim {
	vc := model.VolumeClaim{
		PVCName:     mapx.GetStr(raw, "metadata.name"),
		ClaimType:   resCsts.PVCTypeUseExistPV,
		PVName:      mapx.GetStr(raw, "spec.volumeName"),
		SCName:      mapx.GetStr(raw, "spec.storageClassName"),
		StorageSize: util.ConvertStorageUnit(mapx.GetStr(raw, "spec.resources.requests.storage")),
		AccessModes: []string{},
	}
	// 如果没有指定 PVName，则认为是要根据 StorageClass 创建
	if vc.PVName == "" {
		vc.ClaimType = resCsts.PVCTypeCreateBySC
	}
	if accessModes := mapx.GetList(raw, "spec.accessModes"); len(accessModes) != 0 {
		for _, am := range accessModes {
			vc.AccessModes = append(vc.AccessModes, am.(string))
		}
	}
	return vc
}
