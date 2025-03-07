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

// Package containercheck xxx
package containercheck

const (
	pluginName              = "containercheck"
	Normalstatus            = "ok"
	runtimeTarget           = "runtime"
	wrongRootDirStatus      = "rootdir wrong"
	initContent             = `interval: 600`
	readFileFailStatus      = "read file failed"
	dnsInconsistencyStatus  = "dns inconsistency"
	inconsistentStatus      = "inconsistent"
	getProcessFailStatus    = "get process failed"
	runtimeErrorStatus      = "runtime error"
	processNotExistStatus   = "process not exist"
	containerNotFoundStatus = "container not found"
	inspectCoantainerError  = "inspect container error"
)

var (
	ChinenseStringMap = map[string]string{
		pluginName:              "容器检查",
		Normalstatus:            "正常",
		wrongRootDirStatus:      wrongRootDirStatus,
		readFileFailStatus:      readFileFailStatus,
		dnsInconsistencyStatus:  dnsInconsistencyStatus,
		inconsistentStatus:      inconsistentStatus,
		getProcessFailStatus:    getProcessFailStatus,
		runtimeErrorStatus:      runtimeErrorStatus,
		processNotExistStatus:   processNotExistStatus,
		containerNotFoundStatus: containerNotFoundStatus,
		inspectCoantainerError:  inspectCoantainerError,
	}

	EnglishStringMap = map[string]string{
		pluginName:              pluginName,
		Normalstatus:            Normalstatus,
		wrongRootDirStatus:      wrongRootDirStatus,
		readFileFailStatus:      readFileFailStatus,
		dnsInconsistencyStatus:  dnsInconsistencyStatus,
		inconsistentStatus:      inconsistentStatus,
		getProcessFailStatus:    getProcessFailStatus,
		runtimeErrorStatus:      runtimeErrorStatus,
		processNotExistStatus:   processNotExistStatus,
		containerNotFoundStatus: containerNotFoundStatus,
		inspectCoantainerError:  inspectCoantainerError,
	}

	StringMap = ChinenseStringMap
)
