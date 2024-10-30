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

// Package nodecheck xxx
package nodecheck

const (
	initContent          = `interval: 3600`
	nodeagentNamespace   = "nodeagent"
	pluginName           = "nodecheck"
	configErrorStatus    = "config error"
	ConfigNotFoundStatus = "confignotfound"
	ConfigFileDetail     = "%s config file"

	normalStatus = "ok"

	flagNotSetDetail = "%s is not set, which is recommanded"

	processConfigCheckItem = "process config check"
)

// 统一format格式
var (
	ChinenseStringMap = map[string]string{
		processConfigCheckItem: "进程配置检查",
		configErrorStatus:      "配置错误",
		ConfigNotFoundStatus:   "配置不存在",
		flagNotSetDetail:       "%s 参数没有配置，推荐配置",
		normalStatus:           "正常",
		pluginName:             "节点检查",
		ConfigFileDetail:       "%s 配置文件",
	}

	EnglishStringMap = map[string]string{
		processConfigCheckItem: processConfigCheckItem,
		configErrorStatus:      configErrorStatus,
		ConfigNotFoundStatus:   ConfigNotFoundStatus,
		flagNotSetDetail:       flagNotSetDetail,
		normalStatus:           normalStatus,
		pluginName:             pluginName,
		ConfigFileDetail:       ConfigFileDetail,
	}

	StringMap = ChinenseStringMap
)
