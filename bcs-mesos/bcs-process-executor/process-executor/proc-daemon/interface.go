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

package proc_daemon

import (
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-process-executor/process-executor/types"
)

type ProcDaemon interface {
	//create process object
	CreateProcess(*types.ProcessInfo) error

	//inspect process status
	InspectProcessStatus(procId string) (*types.ProcessStatusInfo, error)

	//stop process
	StopProcess(procId string, timeout int) error

	//Delete process
	DeleteProcess(procId string) error

	//set process envs
	//types.BcsKV: key = env.key, value = env.value
	//SetProcessEnvs([]types.BcsKV)error

	//reload process, exec reloadCmd
	ReloadProcess(procId string) error

	//restart process, exec restartCmd
	RestartProcess(procId string) error
}
