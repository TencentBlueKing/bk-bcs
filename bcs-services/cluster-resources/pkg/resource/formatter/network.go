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
