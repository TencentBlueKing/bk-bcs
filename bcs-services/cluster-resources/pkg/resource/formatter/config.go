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

// FormatConfigRes ...
func FormatConfigRes(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	data := []string{}
	if cmData, ok := manifest["data"].(map[string]interface{}); ok {
		for k := range cmData {
			data = append(data, k)
		}
	}
	ret["data"] = data
	ret["immutable"] = mapx.GetBool(manifest, "immutable")
	return ret
}

// FormatBscpConfig ...
func FormatBscpConfig(manifest map[string]interface{}) map[string]interface{} {
	ret := CommonFormatRes(manifest)
	ret["releaseID"] = mapx.GetInt64(manifest, "status.releaseID")
	return ret
}
