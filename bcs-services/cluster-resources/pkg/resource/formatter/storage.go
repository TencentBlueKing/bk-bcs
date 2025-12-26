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
	"fmt"

	"github.com/samber/lo"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// PVAccessModeList AccessMode 列表
var PVAccessModeList = []string{
	"ReadWriteOnce",
	"ReadOnlyMany",
	"ReadWriteMany",
}

const defaultSCAnnoKey = "storageclass.kubernetes.io/is-default-class"

// FormatStorageRes xxx
func FormatStorageRes(manifest map[string]interface{}) map[string]interface{} {
	return CommonFormatRes(manifest)
}

// FormatPV xxx
func FormatPV(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)

	// accessModes
	ret["accessModes"] = parseAccessModes(manifest)

	// claim
	claimInfo, _ := mapx.GetItems(manifest, "spec.claimRef")
	if c, ok := claimInfo.(map[string]interface{}); ok {
		ret["claim"] = fmt.Sprintf("%s/%s", c["namespace"], c["name"])
	} else {
		ret["claim"] = nil
	}

	return ret
}

// FormatPVC xxx
func FormatPVC(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)
	ret["accessModes"] = parseAccessModes(manifest)
	return ret
}

// FormatSC xxx
func FormatSC(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)
	ret["isDefault"] = mapx.GetStr(manifest, []string{"metadata", "annotations", defaultSCAnnoKey}) == "true"
	return ret
}

// 工具方法

// parseAccessModes 从Kubernetes资源清单中解析访问模式，返回符合PV标准的模式列表
// 参数manifest为Kubernetes资源清单的map表示
// 返回值shortAccessModes为符合PVAccessModeList定义的访问模式字符串切片
func parseAccessModes(manifest map[string]interface{}) (shortAccessModes []string) {
	for _, am := range mapx.GetList(manifest, "spec.accessModes") {
		if accessMode, ok := am.(string); ok && lo.Contains(PVAccessModeList, accessMode) {
			shortAccessModes = append(shortAccessModes, accessMode)
		}
	}
	return shortAccessModes
}
