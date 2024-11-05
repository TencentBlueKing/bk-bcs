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

// Package process xxx
package process

// ProcessInfo xxx
type ProcessInfo struct {
	Starter      string // the way start this process
	BinaryPath   string //
	Params       []string
	Env          []string
	ConfigFiles  map[string]string
	ServiceFiles map[string]string
	Status       string
	// 配置文件修改时间，进程启动时间，
}

// ProcessStatus xxx
type ProcessStatus struct {
	Name       string // the way start this process
	Pid        int32
	Status     string
	CreateTime int64
	CpuTime    float64

	// 配置文件修改时间，进程启动时间，
}

// NS xxx
type NS string
