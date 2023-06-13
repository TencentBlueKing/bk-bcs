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

package resp

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/i18n"
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	resScene "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/scene"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// 从 manifest 生成 selectItem，部分资源在某些场景下，需要调整 label, disabled，tips 的属性
// 如果后续需要特殊处理的资源类型增多，可以考虑参考 formatter 抽离逻辑
func genSelectItem(ctx context.Context, manifest map[string]interface{}, kind, scene string) map[string]interface{} {
	name := mapx.GetStr(manifest, "metadata.name")
	label, disabled, tips := name, false, ""

	switch kind {
	case resCsts.PV:
		// PersistentVolume 如果状态不是 Available, 则不可用于绑定
		if mapx.GetStr(manifest, "status.phase") != "Available" {
			disabled, tips = true, i18n.GetMsg(ctx, "非 Available 状态，无法绑定")
		}
	case resCsts.Node:
		// Master 节点需要给展示的名称做下标识
		if mapx.GetStr(manifest, []string{"metadata", "labels", resCsts.MasterNodeLabelKey}) == "true" {
			label += " (master)"
		}
	case resCsts.SVC:
		// 使用场景为 Ingress Target Service 时，非 NodePort, LoadBalancer 类型的 Service，需要禁用
		if scene == resScene.IngTargetSVC && !slice.StringInSlice(
			mapx.GetStr(manifest, "spec.type"), resCsts.IngTargetSVCEnabledServiceTypes,
		) {
			disabled, tips = true, i18n.GetMsg(ctx, "仅可使用 NodePort，LoadBalancer 类型的 Service")
		}
	case resCsts.Secret:
		secretType := mapx.GetStr(manifest, "type")
		// 使用场景为 Ingress tls 证书时，非 tls 证书或 Opaque 类型的，需要禁用
		// 注：qcloud 类型 ingress 比较特殊，它使用的 Secret 是 Opaque 类型的 ...
		if scene == resScene.IngTLSCert && !slice.StringInSlice(secretType, resCsts.IngTLSCertEnabledSecretTypes) {
			disabled, tips = true, i18n.GetMsg(ctx, "仅可使用 tls 证书或 Opaque 类型的 Secret")
		}
		// 使用场景为 Workload imagePullSecrets 时，非 docker config 类的，需要禁用
		if scene == resScene.WorkloadImagePullSecret && secretType != resCsts.SecretTypeDocker {
			disabled, tips = true, i18n.GetMsg(ctx, "仅可使用 dockerconfigjson 类型的 Secret")
		}
	}

	return map[string]interface{}{"label": label, "value": name, "disabled": disabled, "tips": tips}
}

// 部分资源在某些场景下，允许扩展 SelectItems
func genExtSelectItems(source []interface{}, kind, scene string) (exts []map[string]interface{}) {
	switch kind {
	case resCsts.SC:
		for _, item := range source {
			if mapx.GetStr(item.(map[string]interface{}), "label") == "local-storage" {
				return exts
			}
		}
		// StorageClass 允许使用 local-storage，即使集群中不存在
		exts = append(exts, map[string]interface{}{
			"label": "local-storage", "value": "local-storage", "disabled": false, "tips": "",
		})
	}
	return exts
}
