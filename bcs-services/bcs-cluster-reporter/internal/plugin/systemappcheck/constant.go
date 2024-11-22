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

// Package systemappcheck xxx
package systemappcheck

const (
	pluginName                         = "systemappcheck"
	SystemAppImageVersionCheckItemName = "system_app_image_version"
	SystemAppStatusCheckItemName       = "system_app_status"
	SystemAppChartVersionCheckItem     = "system_app_chart_version"
	SystemAppConfigCheckItem           = "system_app_config"

	NormalStatus             = "ok"
	ImageStatusNeedUpgrade   = "need_upgrade"
	ImageStatusNiceToUpgrade = "nice_to_upgrade"
	ImageStatusUnknown       = "unknown"

	AppStatusNotReadyStatus   = "notready"
	AppStatusMemoryHighStatus = "memoryhigh"
	AppStatusCpuHighStatus    = "cpuhigh"
	AppErrorStatus            = "error"
	AppMetricErrorStatus      = "metric_error"

	ChartVersionNormalStatus = "deployed"
	APPNotfoundAppStatus     = "appnotfound"

	ConfigErrorStatus         = "configerr"
	ConfigInconsistencyStatus = "configinconsistency "
	NolabelStatus             = "no labels"
	ConfigOtherErrorStatus    = "ConfigOtherErrorStatus"
	UnrecommandedStatus       = "UnrecommandedStatus"
	ConfigNotFoundStatus      = "confignotfound"
	initContent               = `interval: 300`

	StaticPodConfigTarget = "StaticPodConfigTarget"
	SystemAppConfigTarget = "SystemAppConfigTarget"
	ServiceConfigTarget   = "ServiceConfigTarget"

	FlagUnsetDetailFormat = "%s is not set, which is recommanded set"
	NoLabelDetailFormat   = "%s has no labels, cannot be selected"

	etcdDataDiskDetail  = "etcd used %s, recommand use data disk"
	kubeProxyIpvsDetail = "when kube-proxy is ipvs mode, --ipvs-udp-timeout=10s is recommanded"

	lbSVCNoIpDetail         = "service %s %s has no external ips."
	GetResourceFailedDetail = "get resource %s failed: %s"

	deployment  = "Deployment"
	daemonset   = "Daemonset"
	statefulset = "Statefulset"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:            "系统应用检查",
		NormalStatus:          "正常",
		StaticPodConfigTarget: "静态pod配置检查",
		SystemAppConfigTarget: "系统应用配置检查",
		ServiceConfigTarget:   "service配置检查",

		ConfigErrorStatus:         "配置错误",
		ConfigInconsistencyStatus: "配置不一致",
		ConfigNotFoundStatus:      "配置不存在",
		ConfigOtherErrorStatus:    "其它问题",
		APPNotfoundAppStatus:      "应用不存在",
		NolabelStatus:             "没有标签",
		AppErrorStatus:            "错误",
		AppStatusMemoryHighStatus: "应用内存高",
		AppStatusCpuHighStatus:    "应用cpu高",
		UnrecommandedStatus:       "非推荐值",
		GetResourceFailedDetail:   "获取 %s 失败: %s",

		SystemAppImageVersionCheckItemName: "应用镜像版本检查",
		SystemAppConfigCheckItem:           "应用配置检查",
		SystemAppStatusCheckItemName:       "应用状态检查",
		SystemAppChartVersionCheckItem:     "应用chart版本检查",

		FlagUnsetDetailFormat: "没有配置%s 参数，推荐配置",
		NoLabelDetailFormat:   "%s 没有任何标签，不能被选中",

		etcdDataDiskDetail:  "etcd存储在%s, 推荐存储在数据盘",
		kubeProxyIpvsDetail: "kube-proxy使用ipvs时，推荐设置--ipvs-udp-timeout=10s",

		lbSVCNoIpDetail: "service %s %s没有external ip",
	}

	EnglishStringMap = map[string]string{
		pluginName:            pluginName,
		NormalStatus:          NormalStatus,
		StaticPodConfigTarget: "staic pod config check",
		SystemAppConfigTarget: "system app config check",
		ServiceConfigTarget:   "service config check",

		ConfigErrorStatus:         ConfigErrorStatus,
		ConfigInconsistencyStatus: ConfigInconsistencyStatus,
		ConfigNotFoundStatus:      ConfigNotFoundStatus,
		ConfigOtherErrorStatus:    "other err",
		APPNotfoundAppStatus:      APPNotfoundAppStatus,
		NolabelStatus:             NolabelStatus,
		AppErrorStatus:            AppErrorStatus,
		AppStatusMemoryHighStatus: AppStatusMemoryHighStatus,
		AppStatusCpuHighStatus:    AppStatusCpuHighStatus,
		UnrecommandedStatus:       UnrecommandedStatus,
		GetResourceFailedDetail:   GetResourceFailedDetail,

		SystemAppImageVersionCheckItemName: SystemAppImageVersionCheckItemName,
		SystemAppConfigCheckItem:           SystemAppConfigCheckItem,
		SystemAppStatusCheckItemName:       SystemAppStatusCheckItemName,
		SystemAppChartVersionCheckItem:     SystemAppChartVersionCheckItem,

		FlagUnsetDetailFormat: FlagUnsetDetailFormat,
		NoLabelDetailFormat:   NoLabelDetailFormat,

		etcdDataDiskDetail:  etcdDataDiskDetail,
		kubeProxyIpvsDetail: kubeProxyIpvsDetail,
		lbSVCNoIpDetail:     lbSVCNoIpDetail,
	}

	StringMap = ChinenseStringMap
)
