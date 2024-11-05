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

// Package processcheck xxx
package processcheck

const (
	pluginName                      = "processcheck"
	NormalStatus                    = "ok"
	ProcessTarget                   = "process"
	processStatusCheckItemType      = "process status check"
	processNotFoundStatus           = "process notfound"
	processConfigFileNotFoundStatus = "configfile_notfound"
	zStatus                         = "z"
	dStatus                         = "d"
	processOtherErrorStatus         = "error"
	initContent                     = `interval: 600`
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:                      "进程检查",
		processNotFoundStatus:           "进程不存在",
		processConfigFileNotFoundStatus: "配置文件不存在",
		processStatusCheckItemType:      "进程状态检查",
		NormalStatus:                    "正常",
		processOtherErrorStatus:         processOtherErrorStatus,
		zStatus:                         zStatus,
		dStatus:                         dStatus,
		ProcessTarget:                   "进程",
	}

	EnglishStringMap = map[string]string{
		pluginName:                      pluginName,
		processNotFoundStatus:           processNotFoundStatus,
		processConfigFileNotFoundStatus: processConfigFileNotFoundStatus,
		processStatusCheckItemType:      processStatusCheckItemType,
		NormalStatus:                    NormalStatus,
		processOtherErrorStatus:         processOtherErrorStatus,
		zStatus:                         zStatus,
		dStatus:                         dStatus,
		ProcessTarget:                   ProcessTarget,
	}

	StringMap = ChinenseStringMap
)
