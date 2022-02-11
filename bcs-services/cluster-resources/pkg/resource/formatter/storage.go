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

package formatter

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"
)

// PVAccessMode2ShortMap PersistentVolume AccessMode 缩写映射表
var PVAccessMode2ShortMap = map[string]string{
	"ReadWriteOnce": "RWO",
	"ReadOnlyMany":  "ROX",
	"ReadWriteMany": "RWX",
}

// FormatStorageRes ...
func FormatStorageRes(manifest map[string]interface{}) map[string]interface{} {
	return CommonFormatRes(manifest)
}

// FormatPV ...
func FormatPV(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)

	// accessModes
	ret["accessModes"] = parseShortAccessModes(manifest)

	// claim
	claimInfo, _ := util.GetItems(manifest, "spec.claimRef")
	if c, ok := claimInfo.(map[string]interface{}); ok {
		ret["claim"] = fmt.Sprintf("%s/%s", c["namespace"], c["name"])
	} else {
		ret["claim"] = nil
	}

	return ret
}

// FormatPVC ...
func FormatPVC(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)
	ret["accessModes"] = parseShortAccessModes(manifest)
	return ret
}

// 工具方法

// 解析 AccessModes (缩写)
func parseShortAccessModes(manifest map[string]interface{}) (shortAccessModes []string) {
	accessModes, _ := util.GetItems(manifest, "spec.accessModes")
	for _, am := range accessModes.([]interface{}) {
		shortAccessModes = append(shortAccessModes, PVAccessMode2ShortMap[am.(string)])
	}
	return shortAccessModes
}
