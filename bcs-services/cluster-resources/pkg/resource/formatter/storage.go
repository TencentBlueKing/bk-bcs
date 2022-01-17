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

import "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util"

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

// FormatPVCRes ...
func FormatPVCRes(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatStorageRes(manifest)
	shortAccessModes := []string{}
	accessModes, _ := util.GetItems(manifest, "spec.accessModes")
	for _, am := range accessModes.([]interface{}) {
		shortAccessModes = append(shortAccessModes, PVAccessMode2ShortMap[am.(string)])
	}
	ret["accessModes"] = shortAccessModes
	return ret
}
