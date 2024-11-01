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

// Package hwcheck xxx
package hwcheck

const (
	pluginName       = "hwcheck"
	NormalStatus     = "ok"
	ffStatus         = "ff"
	logMatchedStatus = "matched"

	DeviceCheckItemType   = "device check"
	DeviceCheckItemTarget = "device"

	LogCheckItemType   = "log check"
	LogCheckItemTarget = "log"

	initContent = `interval: 600
logFileConfigs:
  - path: "/var/log/messages"
    keyWordList: 
      - "mce: [Hardware Error]"
    rule: "mce"`
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:            "硬件检查",
		NormalStatus:          "正常",
		ffStatus:              ffStatus,
		logMatchedStatus:      "匹配到异常日志",
		DeviceCheckItemType:   "设备检查",
		DeviceCheckItemTarget: "设备",

		LogCheckItemType:   "日志检查",
		LogCheckItemTarget: "日志",
	}

	EnglishStringMap = map[string]string{
		pluginName:            pluginName,
		NormalStatus:          NormalStatus,
		ffStatus:              ffStatus,
		logMatchedStatus:      logMatchedStatus,
		DeviceCheckItemType:   DeviceCheckItemType,
		DeviceCheckItemTarget: DeviceCheckItemTarget,
		LogCheckItemType:      LogCheckItemType,
		LogCheckItemTarget:    LogCheckItemTarget,
	}

	StringMap = ChinenseStringMap
)
