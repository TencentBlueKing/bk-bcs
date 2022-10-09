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
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/slice"
)

// ParseIng ...
func ParseIng(manifest map[string]interface{}) map[string]interface{} {
	ing := model.Ing{}
	common.ParseMetadata(manifest, &ing.Metadata)
	ParseIngController(manifest, &ing.Controller)
	ParseIngSpec(manifest, &ing.Spec)
	return structs.Map(ing)
}

// ParseIngController ...
func ParseIngController(manifest map[string]interface{}, controller *model.IngController) {
	clsNamePath := []string{"metadata", "annotations", IngClsAnnoKey}
	controller.Type = mapx.Get(manifest, clsNamePath, IngClsNginx).(string)
}

// ParseIngSpec ...
func ParseIngSpec(manifest map[string]interface{}, spec *model.IngSpec) {
	ParseIngRuleConf(manifest, &spec.RuleConf)
	ParseIngNetwork(manifest, &spec.Network)
	ParseIngDefaultBackend(manifest, &spec.DefaultBackend)
	ParseIngCert(manifest, &spec.Cert)
}

// ParseIngRuleConf ...
func ParseIngRuleConf(manifest map[string]interface{}, ruleConf *model.IngRuleConf) {
	svcNamePath, svcPortPath := "backend.service.name", "backend.service.port.number"
	if isV1beta1Ingress(mapx.GetStr(manifest, "apiVersion")) {
		svcNamePath, svcPortPath = "backend.serviceName", "backend.servicePort"
	}
	for _, rule := range mapx.GetList(manifest, "spec.rules") {
		r := rule.(map[string]interface{})
		paths := []model.IngPath{}
		for _, path := range mapx.GetList(r, "http.paths") {
			p := path.(map[string]interface{})
			paths = append(paths, model.IngPath{
				Type:      mapx.GetStr(p, "pathType"),
				Path:      mapx.GetStr(p, "path"),
				TargetSVC: mapx.GetStr(p, svcNamePath),
				Port:      mapx.GetInt64(p, svcPortPath),
			})
		}
		ruleConf.Rules = append(ruleConf.Rules, model.IngRule{
			Domain: mapx.GetStr(r, "host"), Paths: paths,
		})
	}
}

// ParseIngNetwork ...
func ParseIngNetwork(manifest map[string]interface{}, network *model.IngNetwork) {
	existLBIDPath := []string{"metadata", "annotations", IngExistLBIDAnnoKey}
	network.ExistLBID = mapx.GetStr(manifest, existLBIDPath)
	// 如果已指定 clb，则使用模式为使用已存在的 clb，否则为自动创建新 clb
	if network.ExistLBID != "" {
		network.CLBUseType = CLBUseTypeUseExists
	} else {
		network.CLBUseType = CLBUseTypeAutoCreate
	}

	subNetIDPath := []string{"metadata", "annotations", IngSubNetIDAnnoKey}
	network.SubNetID = mapx.GetStr(manifest, subNetIDPath)
	// 如果已指定子网 ID，则认为是内网 clb，否则为外网 clb
	if network.SubNetID != "" {
		network.CLBType = CLBTypeInternal
	} else {
		network.CLBType = CLBTypeExternal
	}
}

// ParseIngDefaultBackend ...
func ParseIngDefaultBackend(manifest map[string]interface{}, bak *model.IngDefaultBackend) {
	defaultBakPath, svcNamePath, svcPortPath := "spec.defaultBackend", "service.name", "service.port.number"
	if isV1beta1Ingress(mapx.GetStr(manifest, "apiVersion")) {
		defaultBakPath, svcNamePath, svcPortPath = "spec.backend", "serviceName", "servicePort"
	}

	backend := mapx.GetMap(manifest, defaultBakPath)
	bak.TargetSVC = mapx.GetStr(backend, svcNamePath)
	bak.Port = mapx.GetInt64(backend, svcPortPath)
}

// ParseIngCert ...
func ParseIngCert(manifest map[string]interface{}, cert *model.IngCert) {
	for _, tls := range mapx.GetList(manifest, "spec.tls") {
		t := tls.(map[string]interface{})
		hosts := []string{}
		for _, host := range mapx.GetList(t, "hosts") {
			hosts = append(hosts, host.(string))
		}
		cert.TLS = append(cert.TLS, model.IngTLS{
			SecretName: mapx.GetStr(t, "secretName"),
			Hosts:      hosts,
		})
	}
}

func isV1beta1Ingress(apiVersion string) bool {
	return slice.StringInSlice(apiVersion, []string{"extensions/v1beta1", "networking.k8s.io/v1beta1"})
}
