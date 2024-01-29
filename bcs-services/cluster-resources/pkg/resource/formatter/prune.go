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

import "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"

// DefaultPruneFunc 默认的 PruneFunc
func DefaultPruneFunc(manifest map[string]interface{}) map[string]interface{} {
	return manifest
}

// CommonPrune 裁剪
func CommonPrune(manifest map[string]interface{}) map[string]interface{} {
	name, _ := mapx.GetItems(manifest, "metadata.name")
	namespace, _ := mapx.GetItems(manifest, "metadata.namespace")
	uid, _ := mapx.GetItems(manifest, "metadata.uid")
	labels, _ := mapx.GetItems(manifest, "metadata.labels")
	annotations, _ := mapx.GetItems(manifest, "metadata.annotations")
	newManifest := map[string]interface{}{
		"apiVersion": mapx.GetStr(manifest, "apiVersion"),
		"kind":       mapx.GetStr(manifest, "kind"),
		"metadata": map[string]interface{}{
			"name": name, "namespace": namespace, "uid": uid, "labels": labels, "annotations": annotations},
	}
	return newManifest
}

// PrunePod 裁剪 Pod
func PrunePod(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	hostIP, _ := mapx.GetItems(manifest, "status.hostIP")
	nodeName, _ := mapx.GetItems(manifest, "spec.nodeName")
	ret["status"] = map[string]interface{}{"hostIP": hostIP}
	ret["spec"] = map[string]interface{}{"nodeName": nodeName}
	return ret
}

// PruneConfig 裁剪 Configmap 和 secret
func PruneConfig(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonPrune(manifest)
	return ret
}
