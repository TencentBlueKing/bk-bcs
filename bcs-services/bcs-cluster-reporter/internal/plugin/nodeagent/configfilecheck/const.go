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

// Package configfilecheck xxx
package configfilecheck

const (
	pluginName                 = "configfilecheck"
	NormalStatus               = "ok"
	ProcessTarget              = "process"
	processStatusCheckItemType = "process status check"
	processNotFoundStatus      = "process notfound"
	configFileNotFoundStatus   = "configfile_notfound"
	configNotFoundStatus       = "config_notfound"
	errorConfigMatchedStatus   = "errconfig matched"
	zStatus                    = "z"
	dStatus                    = "d"
	otherErrorStatus           = "error"
	initContent                = `interval: 600`
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:                 "配置文件检查",
		configFileNotFoundStatus:   "配置文件不存在",
		configNotFoundStatus:       "配置不存在",
		errorConfigMatchedStatus:   "存在错误配置",
		processStatusCheckItemType: "进程状态检查",
		NormalStatus:               "正常",
		otherErrorStatus:           otherErrorStatus,
		zStatus:                    zStatus,
		dStatus:                    dStatus,
		ProcessTarget:              "进程",
	}

	EnglishStringMap = map[string]string{
		pluginName:                 pluginName,
		processNotFoundStatus:      processNotFoundStatus,
		configFileNotFoundStatus:   configFileNotFoundStatus,
		configNotFoundStatus:       configNotFoundStatus,
		errorConfigMatchedStatus:   errorConfigMatchedStatus,
		processStatusCheckItemType: processStatusCheckItemType,
		NormalStatus:               NormalStatus,
		otherErrorStatus:           otherErrorStatus,
		zStatus:                    zStatus,
		dStatus:                    dStatus,
		ProcessTarget:              ProcessTarget,
	}

	StringMap = ChinenseStringMap
)
