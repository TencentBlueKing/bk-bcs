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

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
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
	clsNamePath := []string{"metadata", "annotations", resCsts.IngClsAnnoKey}
	controller.Type = mapx.Get(manifest, clsNamePath, resCsts.IngClsNginx).(string)
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
	existLBIDPath := []string{"metadata", "annotations", resCsts.IngExistLBIDAnnoKey}
	network.ExistLBID = mapx.GetStr(manifest, existLBIDPath)

	subNetIDPath := []string{"metadata", "annotations", resCsts.IngSubNetIDAnnoKey}
	network.SubNetID = mapx.GetStr(manifest, subNetIDPath)

	// 如果已指定子网 ID，则使用模式为为自动创建新 clb，否则使用已存在的 clb
	if network.SubNetID != "" {
		network.CLBUseType = resCsts.CLBUseTypeAutoCreate
	} else {
		network.CLBUseType = resCsts.CLBUseTypeUseExists
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
	if mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.IngAutoRewriteHTTPAnnoKey}) == "true" {
		cert.AutoRewriteHTTP = true
	}
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
