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
 *
 */

package manager

import (
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-process-daemon/process-daemon/config"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-process-executor/process-executor/types"
)

// Manager xxx
type Manager interface {
	// Init func
	Init() error

	// Start func
	Start()

	// GetConfig xxx
	// get config
	GetConfig() *config.Config

	// HeartBeat xxx
	// heartbeat
	HeartBeat(heartbeat *types.HeartBeat)

	// CreateProcess xxx
	// Create process
	CreateProcess(processInfo *types.ProcessInfo) error

	// InspectProcessStatus xxx
	// inspect process status info
	// processId = types.ProcessInfo.Id
	InspectProcessStatus(processId string) (*types.ProcessStatusInfo, error)

	// StopProcess xxx
	// Stop process
	// processId = types.ProcessInfo.Id
	// process will be killed when timeout seconds
	StopProcess(processId string, timeout int) error

	// DeleteProcess xxx
	// delete process
	// processId = types.ProcessInfo.Id
	DeleteProcess(processId string) error
}
