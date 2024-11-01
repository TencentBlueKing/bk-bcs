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

// Package dnscheck xxx
package dnscheck

const (
	pluginName       = "dnscheck"
	NormalStatus     = "ok"
	ResolvFailStauts = "resolvefailed"
	initContent      = `interval: 600`

	clusterDNSType        = "pod dns check"
	clusterDNSclusterType = "cluster"
	clusterDNShostType    = "node"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:            "dns检查",
		clusterDNSType:        "节点DNS检查",
		clusterDNSclusterType: "容器域名解析",
		clusterDNShostType:    "节点域名解析",

		ResolvFailStauts: "解析失败",
		NormalStatus:     "正常",
	}

	EnglishStringMap = map[string]string{
		pluginName:            pluginName,
		clusterDNSType:        clusterDNSType,
		clusterDNSclusterType: clusterDNSclusterType,
		clusterDNShostType:    clusterDNShostType,

		ResolvFailStauts: ResolvFailStauts,
		NormalStatus:     NormalStatus,
	}

	StringMap = ChinenseStringMap
)
