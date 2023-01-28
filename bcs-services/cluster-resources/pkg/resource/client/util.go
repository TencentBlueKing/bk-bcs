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

package client

import "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"

// filterByOwnerRefs 根据 ownerReferences 过滤子资源
func filterByOwnerRefs(subResItems []interface{}, ownerRefs []map[string]string) []interface{} {
	rets := []interface{}{}
	for _, subRes := range subResItems {
		sr, _ := subRes.(map[string]interface{})
		for _, resOwnerRef := range mapx.GetList(sr, "metadata.ownerReferences") {
			for _, ref := range ownerRefs {
				kind, name := "", ""
				if r, ok := resOwnerRef.(map[string]interface{}); ok {
					kind, name = r["kind"].(string), r["name"].(string)
				}
				if kind == ref["kind"] && name == ref["name"] {
					rets = append(rets, subRes)
					break
				}
			}
		}
	}
	return rets
}
