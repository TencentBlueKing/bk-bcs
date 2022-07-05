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

package custom

import (
	"github.com/fatih/structs"

	res "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/util"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/workload"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseGDeploy GameDeployment manifest -> formData
func ParseGDeploy(manifest map[string]interface{}) map[string]interface{} {
	deploy := model.GDeploy{}
	common.ParseMetadata(manifest, &deploy.Metadata)
	ParseGDeploySpec(manifest, &deploy.Spec)
	workload.ParseWorkloadVolume(manifest, &deploy.Volume)
	workload.ParseContainerGroup(manifest, &deploy.ContainerGroup)
	return structs.Map(deploy)
}

// ParseGDeploySpec ...
func ParseGDeploySpec(manifest map[string]interface{}, spec *model.GDeploySpec) {
	ParseGDeployReplicas(manifest, &spec.Replicas)
	ParseGDeployGracefulManage(manifest, &spec.GracefulManage)
	ParseGDeployDeletionProtect(manifest, &spec.DeletionProtect)
	tmplSpec, _ := mapx.GetItems(manifest, "spec.template.spec")
	podSpec, _ := tmplSpec.(map[string]interface{})
	workload.ParseNodeSelect(podSpec, &spec.NodeSelect)
	workload.ParseAffinity(podSpec, &spec.Affinity)
	workload.ParseToleration(podSpec, &spec.Toleration)
	workload.ParseNetworking(podSpec, &spec.Networking)
	workload.ParsePodSecurityCtx(podSpec, &spec.Security)
	workload.ParseSpecOther(podSpec, &spec.Other)
}

// ParseGDeployReplicas ...
func ParseGDeployReplicas(manifest map[string]interface{}, replicas *model.GDeployReplicas) {
	replicas.Cnt = mapx.GetInt64(manifest, "spec.replicas")
	replicas.UpdateStrategy = mapx.Get(manifest, "spec.updateStrategy.type", workload.DefaultUpdateStrategy).(string)
	replicas.MaxSurge, replicas.MSUnit = DefaultGDeployMaxSurge, util.UnitCnt
	if maxSurge, err := mapx.GetItems(manifest, "spec.updateStrategy.maxSurge"); err == nil {
		replicas.MaxSurge, replicas.MSUnit = util.AnalyzeIntStr(maxSurge)
	}
	replicas.MaxUnavailable, replicas.MUAUnit = DefaultGDeployMaxUnavailable, util.UnitPercent
	if maxUnavailable, err := mapx.GetItems(manifest, "spec.updateStrategy.maxUnavailable"); err == nil {
		replicas.MaxUnavailable, replicas.MUAUnit = util.AnalyzeIntStr(maxUnavailable)
	}
	replicas.MinReadySecs = mapx.GetInt64(manifest, "spec.minReadySeconds")
	replicas.Partition = mapx.GetInt64(manifest, "spec.updateStrategy.partition")
	replicas.GracePeriodSecs = mapx.Get(
		manifest, "spec.updateStrategy.inPlaceUpdateStrategy.gracePeriodSeconds", int64(DefaultGDeployGracePeriodSecs),
	).(int64)
}

// ParseGDeployGracefulManage ...
func ParseGDeployGracefulManage(manifest map[string]interface{}, man *model.GDeployGracefulManage) {
	if hook, err := mapx.GetItems(manifest, "spec.preDeleteUpdateStrategy.hook"); err == nil {
		man.PreDeleteHook = genGDeployHookSpec(hook.(map[string]interface{}))
	}
	if hook, err := mapx.GetItems(manifest, "spec.preInplaceUpdateStrategy.hook"); err == nil {
		man.PreInplaceHook = genGDeployHookSpec(hook.(map[string]interface{}))
	}
	if hook, err := mapx.GetItems(manifest, "spec.postInplaceUpdateStrategy.hook"); err == nil {
		man.PostInplaceHook = genGDeployHookSpec(hook.(map[string]interface{}))
	}
}

func genGDeployHookSpec(hook map[string]interface{}) model.GDeployHookSpec {
	spec := model.GDeployHookSpec{Enabled: true, TmplName: mapx.GetStr(hook, "templateName")}
	for _, arg := range mapx.GetList(hook, "args") {
		a := arg.(map[string]interface{})
		spec.Args = append(spec.Args, model.HookCallArg{
			Key: mapx.GetStr(a, "name"), Value: mapx.GetStr(a, "value"),
		})
	}
	return spec
}

// ParseGDeployDeletionProtect 解析 GameDeployment 删除保护规则
// io.tencent.bcs.dev/deletion-allow: Cascading -> 实例数量为 0 时可以删除
// io.tencent.bcs.dev/deletion-allow: Always -> 实例数量为 0 时可以删除
// label io.tencent.bcs.dev/deletion-allow key 不存在：无法删除
func ParseGDeployDeletionProtect(manifest map[string]interface{}, protect *model.GDeployDeletionProtect) {
	protect.Policy = mapx.Get(
		manifest, []string{"metadata", "labels", res.DeletionProtectLabelKey}, res.GDeployDeletionProtectPolicyNotAllow,
	).(string)
}
