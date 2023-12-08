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

package formatter

import (
	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/timex"
)

// CommonFormatRes 通用资源格式化
func CommonFormatRes(manifest map[string]interface{}) map[string]interface{} {
	rawCreateTime, _ := mapx.GetItems(manifest, "metadata.creationTimestamp")
	createTime, _ := timex.NormalizeDatetime(rawCreateTime.(string))
	ret := map[string]interface{}{
		"namespace":  mapx.GetStr(manifest, []string{"metadata", "namespace"}),
		"age":        timex.CalcAge(rawCreateTime.(string)),
		"createTime": createTime,
		"editMode": mapx.Get(
			manifest, []string{"metadata", "annotations", resCsts.EditModeAnnoKey}, resCsts.EditModeYaml,
		),
		"creator":   mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.CreatorAnnoKey}),
		"updater":   mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.UpdaterAnnoKey}),
		"immutable": parseLabelsHelm(manifest),
	}
	return ret
}

// GetFormatFunc 获取资源对应 FormatFunc
func GetFormatFunc(kind string, apiVersion string) func(manifest map[string]interface{}) map[string]interface{} {
	// 自定义Ingress，按照通用资源格式化
	if kind == resCsts.Ing && apiVersion == resCsts.BCSNetworkApiVersion {
		kind = ""
	}
	formatFunc, ok := Kind2FormatFuncMap[kind]
	if !ok {
		// 若指定资源类型没有对应的，则当作自定义资源处理
		return FormatCObj
	}
	return formatFunc
}

// FormatPodManifestRes 针对pod返回需要用到的字段
func FormatPodManifestRes(kind string, manifest map[string]interface{}) map[string]interface{} {
	// NOCC:ineffassign/assign(误报)
	// nolint
	newManifest := map[string]interface{}{}
	if kind == resCsts.Po {
		metadata, _ := mapx.GetItems(manifest, "metadata")
		newManifest = map[string]interface{}{"kind": mapx.GetStr(manifest, "kind"),
			"apiVersion": mapx.GetStr(manifest, "apiVersion"), "metadata": metadata}
		items := make([]interface{}, 0)
		for _, item := range mapx.GetList(manifest, "items") {
			name, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.name")
			namespace, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.namespace")
			uid, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.uid")
			labels, _ := mapx.GetItems(item.(map[string]interface{}), "metadata.labels")
			hostIP, _ := mapx.GetItems(item.(map[string]interface{}), "status.hostIP")
			nodeName, _ := mapx.GetItems(item.(map[string]interface{}), "spec.nodeName")
			items = append(items, map[string]interface{}{
				"metadata": map[string]interface{}{
					"name": name, "namespace": namespace, "uid": uid, "labels": labels},
				"status":     map[string]interface{}{"hostIP": hostIP},
				"spec":       map[string]interface{}{"nodeName": nodeName},
				"kind":       mapx.GetStr(item.(map[string]interface{}), "kind"),
				"apiVersion": mapx.GetStr(item.(map[string]interface{}), "apiVersion"),
			})
		}
		newManifest["items"] = items
	} else {
		newManifest = manifest
	}
	return newManifest
}

// 解析labels是否包含helm发布
func parseLabelsHelm(manifest map[string]interface{}) bool {
	labels := mapx.GetMap(manifest, "metadata.labels")
	return mapx.GetStr(labels, []string{"app.kubernetes.io/managed-by"}) == "Helm"
}
