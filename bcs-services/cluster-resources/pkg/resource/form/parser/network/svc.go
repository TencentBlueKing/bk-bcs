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

// ParseSVC ...
func ParseSVC(manifest map[string]interface{}) map[string]interface{} {
	svc := model.SVC{}
	common.ParseMetadata(manifest, &svc.Metadata)
	ParseSVCSpec(manifest, &svc.Spec)
	return structs.Map(svc)
}

// ParseSVCSpec ...
func ParseSVCSpec(manifest map[string]interface{}, spec *model.SVCSpec) {
	ParseSVCPortConf(manifest, &spec.PortConf)
	// spec.Selector
	if selector, _ := mapx.GetItems(manifest, "spec.selector"); selector != nil {
		for k, v := range selector.(map[string]interface{}) {
			spec.Selector.Labels = append(spec.Selector.Labels, model.LabelSelector{
				Key: k, Value: v.(string),
			})
		}
	}
	// spec.SessionAffinity
	spec.SessionAffinity.Type = mapx.Get(manifest, "spec.sessionAffinity", SessionAffinityTypeNone).(string)
	spec.SessionAffinity.StickyTime = mapx.GetInt64(manifest, "spec.sessionAffinityConfig.clientIP.timeoutSeconds")
	// spec.IP
	spec.IP.Address = mapx.GetStr(manifest, "spec.clusterIP")
	for _, ip := range mapx.GetList(manifest, "spec.externalIPs") {
		spec.IP.External = append(spec.IP.External, ip.(string))
	}
}

// ParseSVCPortConf ...
func ParseSVCPortConf(manifest map[string]interface{}, portConf *model.SVCPortConf) {
	portConf.Type = mapx.GetStr(manifest, "spec.type")
	for _, port := range mapx.GetList(manifest, "spec.ports") {
		p := port.(map[string]interface{})
		portConf.Ports = append(portConf.Ports, model.SVCPort{
			Name:       mapx.GetStr(p, "name"),
			Port:       mapx.GetInt64(p, "port"),
			Protocol:   mapx.GetStr(p, "protocol"),
			TargetPort: mapx.GetInt64(p, "targetPort"),
			NodePort:   mapx.GetInt64(p, "nodePort"),
		})
	}
}
