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

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
)

// FormatNetworkRes ...
func FormatNetworkRes(manifest map[string]interface{}) map[string]interface{} {
	return CommonFormatRes(manifest)
}

// FormatIng ...
func FormatIng(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatNetworkRes(manifest)

	// 根据不同 api 版本，选择不同的解析 Rules 方法
	parseRulesFunc := map[string]func(map[string]interface{}) []map[string]interface{}{
		"networking.k8s.io/v1": parseV1IngRules,
		"extensions/v1beta1":   parseV1beta1IngRules,
	}[manifest["apiVersion"].(string)]

	ret["hosts"] = parseIngHosts(manifest)
	ret["addresses"] = parseIngAddrs(manifest)
	ret["defaultPorts"] = getIngDefaultPort(manifest)
	ret["rules"] = parseRulesFunc(manifest)
	return ret
}

// FormatSVC ...
func FormatSVC(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatNetworkRes(manifest)
	ret["externalIP"] = parseSVCExternalIPs(manifest)
	ret["ports"] = parseSVCPorts(manifest)
	return ret
}

// FormatEP ...
func FormatEP(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatNetworkRes(manifest)
	ret["endpoints"] = parseEndpoints(manifest)
	return ret
}

// 工具方法

// 解析 Ingress Hosts
func parseIngHosts(manifest map[string]interface{}) (hosts []string) {
	rules := mapx.Get(manifest, "spec.rules", []interface{}{})
	for _, r := range rules.([]interface{}) {
		if h, ok := r.(map[string]interface{})["host"]; ok {
			hosts = append(hosts, h.(string))
		}
	}
	return hosts
}

// 解析 Ingress Address
func parseIngAddrs(manifest map[string]interface{}) (addrs []string) {
	ingresses := mapx.Get(manifest, "status.loadBalancer.ingress", []interface{}{})
	for _, ing := range ingresses.([]interface{}) {
		ing, _ := ing.(map[string]interface{})
		if ip, ok := ing["ip"]; ok {
			addrs = append(addrs, ip.(string))
		} else if hostname, ok := ing["hostname"]; ok {
			addrs = append(addrs, hostname.(string))
		}
	}
	return addrs
}

// 获取 Ingress 默认端口
func getIngDefaultPort(manifest map[string]interface{}) string {
	if tls, _ := mapx.GetItems(manifest, "spec.tls"); tls != nil {
		return "80, 443"
	}
	return "80"
}

// 解析 networking.k8s.io/v1 版本 Ingress Rules
func parseV1IngRules(manifest map[string]interface{}) (rules []map[string]interface{}) {
	rawRules := mapx.Get(manifest, "spec.rules", []interface{}{})
	for _, r := range rawRules.([]interface{}) {
		r, _ := r.(map[string]interface{})
		paths := mapx.Get(r, "http.paths", []interface{}{})
		for _, p := range paths.([]interface{}) {
			p, _ := p.(map[string]interface{})
			subRules := map[string]interface{}{
				"host":        r["host"],
				"path":        p["path"],
				"pathType":    p["pathType"],
				"serviceName": mapx.Get(p, "backend.service.name", "--"),
				"port":        mapx.Get(p, "backend.service.port.number", "--"),
			}
			rules = append(rules, subRules)
		}
	}
	return rules
}

// 解析 extensions/v1beta1 版本 Ingress Rules
func parseV1beta1IngRules(manifest map[string]interface{}) (rules []map[string]interface{}) {
	rawRules := mapx.Get(manifest, "spec.rules", []interface{}{})
	for _, r := range rawRules.([]interface{}) {
		r, _ := r.(map[string]interface{})
		paths := mapx.Get(r, "http.paths", []interface{}{})
		for _, p := range paths.([]interface{}) {
			p, _ := p.(map[string]interface{})
			subRules := map[string]interface{}{
				"host":        r["host"],
				"path":        p["path"],
				"pathType":    "--",
				"serviceName": mapx.Get(p, "backend.serviceName", "--"),
				"port":        mapx.Get(p, "backend.servicePort", "--"),
			}
			rules = append(rules, subRules)
		}
	}
	return rules
}

// 解析 SVC ExternalIP
func parseSVCExternalIPs(manifest map[string]interface{}) []string {
	return parseIngAddrs(manifest)
}

// 解析 SVC Ports
func parseSVCPorts(manifest map[string]interface{}) (ports []string) {
	rawPorts := mapx.Get(manifest, "spec.ports", []map[string]interface{}{})
	for _, p := range rawPorts.([]interface{}) {
		p, _ := p.(map[string]interface{})
		if nodePort, ok := p["nodePort"]; ok {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", p["port"], nodePort, p["protocol"]))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", p["port"], p["protocol"]))
		}
	}
	return ports
}

// 解析所有 Endpoints
func parseEndpoints(manifest map[string]interface{}) (endpoints []string) {
	if _, ok := manifest["subsets"]; !ok {
		return endpoints
	}
	// endpoints 为 subsets ips 与 ports 的笛卡儿积
	for _, subset := range manifest["subsets"].([]interface{}) {
		ss, _ := subset.(map[string]interface{})
		for _, addr := range ss["addresses"].([]interface{}) {
			for _, p := range ss["ports"].([]interface{}) {
				addr, _ := addr.(map[string]interface{})
				p, _ := p.(map[string]interface{})
				endpoints = append(endpoints, fmt.Sprintf("%s:%d", addr["ip"], p["port"]))
			}
		}
	}
	return endpoints
}
