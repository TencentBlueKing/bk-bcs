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

// Package netcheck xxx
package netcheck

const (
	NormalStatus   = "ok"
	pluginName     = "netcheck"
	netCheckTarget = pluginName
	// 包含list namespace下全量的pod操作，不建议太过频繁
	initContent                              = `interval: 3600`
	errorStatus                              = "err"
	devDistinctStatus                        = "dev_distinct"
	devCheckItemType                         = "dev"
	NodeagentItemTarget                      = "node agent pod"
	PingFailedStatus                         = "pingfailed"
	NoTargetPodStatus                        = "notargetpod"
	ClusterApiserverCertExpirationCheckItem  = "ClusterApiserverCertExpiration"
	ApiserverTarget                          = "apiserver"
	AboutToExpireDetail                      = "AboutToExpireDetail"
	AboutToExpireStatus                      = "expire_soon"
	ClusterApiserverCertExpirationMetricName = "cluster_apiserver_cert_expiration"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:                              "网络检查",
		NodeagentItemTarget:                     NodeagentItemTarget,
		errorStatus:                             errorStatus,
		PingFailedStatus:                        "ping失败",
		NoTargetPodStatus:                       "没有可探测的pod",
		NormalStatus:                            "正常",
		devDistinctStatus:                       devDistinctStatus,
		devCheckItemType:                        devCheckItemType,
		ClusterApiserverCertExpirationCheckItem: "apiserver证书过期时间",
		ApiserverTarget:                         ApiserverTarget,
		AboutToExpireDetail:                     "Apiserver 的%s证书将在 %d 秒内过期",
	}

	EnglishStringMap = map[string]string{
		pluginName:                              pluginName,
		NodeagentItemTarget:                     NodeagentItemTarget,
		PingFailedStatus:                        PingFailedStatus,
		NoTargetPodStatus:                       NoTargetPodStatus,
		errorStatus:                             errorStatus,
		NormalStatus:                            NormalStatus,
		devDistinctStatus:                       devDistinctStatus,
		devCheckItemType:                        devCheckItemType,
		ClusterApiserverCertExpirationCheckItem: ClusterApiserverCertExpirationCheckItem,
		ApiserverTarget:                         ApiserverTarget,
		AboutToExpireDetail:                     "Apiserver %s cert is about to expiration in %d seconds, ",
	}

	StringMap = ChinenseStringMap
)
