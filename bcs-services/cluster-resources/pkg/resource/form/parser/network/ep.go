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

package network

import (
	"github.com/fatih/structs"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/model"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/form/parser/common"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// ParseEP ...
func ParseEP(manifest map[string]interface{}) map[string]interface{} {
	ep := model.EP{}
	common.ParseMetadata(manifest, &ep.Metadata)
	ParseEPSpec(manifest, &ep.Spec)
	return structs.Map(ep)
}

// ParseEPSpec ...
func ParseEPSpec(manifest map[string]interface{}, spec *model.EPSpec) {
	for _, subset := range mapx.GetList(manifest, "subsets") {
		ss := subset.(map[string]interface{})

		addresses, ports := []string{}, []model.EPPort{}
		for _, addr := range mapx.GetList(ss, "addresses") {
			addresses = append(addresses, mapx.GetStr(addr.(map[string]interface{}), "ip"))
		}
		for _, port := range mapx.GetList(ss, "ports") {
			p := port.(map[string]interface{})
			ports = append(ports, model.EPPort{
				Name:     mapx.GetStr(p, "name"),
				Port:     mapx.GetInt64(p, "port"),
				Protocol: mapx.GetStr(p, "protocol"),
			})
		}
		spec.SubSets = append(spec.SubSets, model.SubSet{Addresses: addresses, Ports: ports})
	}
}
