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

	"github.com/TencentBlueKing/gopkg/collection/set"

	resCsts "github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/resource/constants"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/mapx"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/util/stringx"
)

// FormatNetworkRes xxx
func FormatNetworkRes(manifest map[string]interface{}) map[string]interface{} {
	return CommonFormatRes(manifest)
}

// FormatIng xxx
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

	annotations := mapx.GetMap(manifest, "metadata.annotations")
	// 控制器
	ret["controller"] = mapx.Get(annotations, []string{resCsts.IngClsAnnoKey}, resCsts.IngClsNginx).(string)

	// 绑定的 CLB ID，qcloud 类型的 ingress 会取实际使用的，其他类型的取默认指定的
	existLBIDPaths := []string{resCsts.IngExistLBIDAnnoKey}
	if ret["controller"] == resCsts.IngClsQCloud {
		existLBIDPaths = []string{resCsts.IngQcloudCurLBIDAnnoKey}
	}
	ret["existLBID"] = mapx.GetStr(annotations, existLBIDPaths)

	// 内网子网 ID
	ret["subNetID"] = mapx.GetStr(annotations, []string{resCsts.IngSubNetIDAnnoKey})

	// CLB 使用方式，如果已指定 clb，则使用模式为使用已存在的 clb，否则为自动创建新 clb
	if ret["existLBID"] != "" {
		ret["clbUseType"] = resCsts.CLBUseTypeUseExists
	} else {
		ret["clbUseType"] = resCsts.CLBUseTypeAutoCreate
	}
	// 重定向 HTTP 端口到 HTTPS
	ret["autoRewrite"] = mapx.GetStr(annotations, []string{resCsts.IngAutoRewriteHTTPAnnoKey}) == "true"
	return ret
}

// FormatSVC xxx
func FormatSVC(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatNetworkRes(manifest)
	ret["externalIP"] = parseSVCExternalIPs(manifest)
	ret["ports"] = parseSVCPorts(manifest)
	ret["clbID"] = mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.SVCCurLBIDAnnoKey})
	ret["subnetID"] = mapx.GetStr(manifest, []string{"metadata", "annotations", resCsts.SVCSubNetIDAnnoKey})
	ret["stickyTime"] = mapx.Get(
		manifest, "spec.sessionAffinityConfig.clientIP.timeoutSeconds",
		resCsts.DefaultSessionAffinityStickyTime,
	).(int64)

	clusterIPSet := set.NewStringSet()
	clusterIP := mapx.GetStr(manifest, "spec.clusterIP")
	clusterIPSet.Add(clusterIP)

	// 双栈集群特有字段
	for _, ip := range mapx.GetList(manifest, "spec.clusterIPs") {
		clusterIPSet.Add(ip.(string))
	}

	// 同时兼容 ipv4 / ipv6 集群
	ret["clusterIPv4"], ret["clusterIPv6"] = "", ""
	for _, ip := range clusterIPSet.ToSlice() {
		switch {
		case stringx.IsIPv4(ip):
			ret["clusterIPv4"] = ip
		case stringx.IsIPv6(ip):
			ret["clusterIPv6"] = ip
		}
	}
	return ret
}

// FormatEP xxx
func FormatEP(manifest map[string]interface{}) map[string]interface{} {
	ret := FormatNetworkRes(manifest)
	ret["endpoints"] = parseEndpoints(manifest)
	return ret
}

// 工具方法

// parseIngHosts 解析 Ingress Hosts
func parseIngHosts(manifest map[string]interface{}) (hosts []string) {
	rules := mapx.GetList(manifest, "spec.rules")
	for _, r := range rules {
		if h, ok := r.(map[string]interface{})["host"]; ok {
			hosts = append(hosts, h.(string))
		}
	}
	return hosts
}

// parseIngAddrs 解析 Ingress Address
func parseIngAddrs(manifest map[string]interface{}) (addrs []string) {
	ingresses := mapx.GetList(manifest, "status.loadBalancer.ingress")
	for _, ing := range ingresses {
		ing, _ := ing.(map[string]interface{})
		if ip, ok := ing["ip"]; ok {
			addrs = append(addrs, ip.(string))
		} else if hostname, ok := ing["hostname"]; ok {
			addrs = append(addrs, hostname.(string))
		}
	}
	return addrs
}

// getIngDefaultPort 获取 Ingress 默认端口
func getIngDefaultPort(manifest map[string]interface{}) string {
	if tls, _ := mapx.GetItems(manifest, "spec.tls"); tls != nil {
		return "80, 443"
	}
	return "80"
}

// parseV1IngRules 解析 networking.k8s.io/v1 版本 Ingress Rules
func parseV1IngRules(manifest map[string]interface{}) (rules []map[string]interface{}) {
	rawRules := mapx.GetList(manifest, "spec.rules")
	for _, r := range rawRules {
		r, _ := r.(map[string]interface{})
		paths := mapx.GetList(r, "http.paths")
		for _, p := range paths {
			p, _ := p.(map[string]interface{})
			subRules := map[string]interface{}{
				"host":        r["host"],
				"path":        p["path"],
				"pathType":    p["pathType"],
				"serviceName": mapx.Get(p, "backend.service.name", "N/A"),
				"port":        mapx.Get(p, "backend.service.port.number", "N/A"),
			}
			rules = append(rules, subRules)
		}
	}
	return rules
}

// parseV1beta1IngRules 解析 extensions/v1beta1 版本 Ingress Rules
func parseV1beta1IngRules(manifest map[string]interface{}) (rules []map[string]interface{}) {
	rawRules := mapx.GetList(manifest, "spec.rules")
	for _, r := range rawRules {
		r, _ := r.(map[string]interface{})
		paths := mapx.GetList(r, "http.paths")
		for _, p := range paths {
			p, _ := p.(map[string]interface{})
			subRules := map[string]interface{}{
				"host":        r["host"],
				"path":        p["path"],
				"pathType":    "--",
				"serviceName": mapx.Get(p, "backend.serviceName", "N/A"),
				"port":        mapx.Get(p, "backend.servicePort", "N/A"),
			}
			rules = append(rules, subRules)
		}
	}
	return rules
}

// parseSVCExternalIPs 解析 SVC ExternalIP
func parseSVCExternalIPs(manifest map[string]interface{}) []string {
	externalIPs := parseIngAddrs(manifest)
	for _, ip := range mapx.GetList(manifest, "spec.externalIPs") {
		externalIPs = append(externalIPs, ip.(string))
	}
	return externalIPs
}

// parseSVCPorts 解析 SVC Ports
func parseSVCPorts(manifest map[string]interface{}) (ports []string) {
	rawPorts := mapx.GetList(manifest, "spec.ports")
	for _, p := range rawPorts {
		p, _ := p.(map[string]interface{})
		if nodePort, ok := p["nodePort"]; ok {
			ports = append(ports, fmt.Sprintf("%d:%d/%s", p["port"], nodePort, p["protocol"]))
		} else {
			ports = append(ports, fmt.Sprintf("%d/%s", p["port"], p["protocol"]))
		}
	}
	return ports
}

// parseEndpoints 解析所有 Endpoints
func parseEndpoints(manifest map[string]interface{}) (endpoints []string) {
	if _, ok := manifest["subsets"]; !ok {
		return endpoints
	}
	// endpoints 为 subsets ips 与 ports 的笛卡儿积
	for _, subset := range mapx.GetList(manifest, "subsets") {
		ss, _ := subset.(map[string]interface{})
		if _, exists := ss["addresses"]; !exists {
			continue
		}
		for _, addr := range mapx.GetList(ss, "addresses") {
			for _, p := range mapx.GetList(ss, "ports") {
				addr, _ := addr.(map[string]interface{})
				p, _ := p.(map[string]interface{})
				endpoints = append(endpoints, fmt.Sprintf("%s:%d", addr["ip"], p["port"]))
			}
		}
	}
	return endpoints
}
