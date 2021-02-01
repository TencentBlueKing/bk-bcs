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

package types

import (
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	//commtypes "bcs/bmsf-mesh/pkg/datatype/bcs/common/types"
)

//Message describe all msg from bcs scheduler to bcs executor
//Include binary file, text file, signal, env
type Msg_Type int32

const (
	Msg_UNKNOWN           Msg_Type = 0
	Msg_LOCALFILE         Msg_Type = 1
	Msg_SIGNAL            Msg_Type = 2
	Msg_ENV               Msg_Type = 3
	Msg_REMOTE            Msg_Type = 4
	Msg_SECRET            Msg_Type = 5
	Msg_TASK_STATUS_QUERY Msg_Type = 6
	Msg_ENV_REMOTE        Msg_Type = 7
	Msg_UPDATE_TASK       Msg_Type = 8
	Msg_COMMIT_TASK       Msg_Type = 9
)

const (
	Msg_UNKNOWN_STR           string = "unknown"
	Msg_LOCALFILE_STR         string = "localfile"
	Msg_SIGNAL_STR            string = "signal"
	Msg_ENV_STR               string = "env"
	Msg_REMOTE_STR            string = "remote"
	Msg_SECRET_STR            string = "secret"
	Msg_TASK_STATUS_QUERY_STR string = "task_status_query"
	Msg_ENV_REMOTE_STR        string = "env_remote"
	Msg_UPDATE_TASK_STR       string = "update_task"
	Msg_COMMIT_TASK_STR       string = "commit_task"
)

type Secret_Type int32

const (
	Secret_Unknown Secret_Type = 0
	Secret_Env     Secret_Type = 1
	Secret_File    Secret_Type = 2
)

//BcsMessage describe msg from scheduler to executor by mesos MESSAGE call
type BcsMessage struct {
	Id          int64
	Type        *Msg_Type
	TaskGroupId string
	//if TaskID is null, message should be send to all tasks in same executor instance,
	//else if TaskID is not null, message should be sendto the task specialed by TaskID.
	TaskID              *mesos.TaskID
	Local               *Msg_LocalFile           `json:",omitempty"`
	Sig                 *Msg_Signal              `json:",omitempty"`
	Env                 *Msg_Env                 `json:",omitempty"`
	EnvRemote           *Msg_EnvRemote           `json:",omitempty"`
	Remote              *Msg_Remote              `json:",omitempty"`
	Secret              *Msg_Secret              `json:",omitempty"`
	TaskStatusQuery     *Msg_TaskStatusQuery     `json:",omitempty"`
	UpdateTaskResources *Msg_UpdateTaskResources `json:",omitempty"`
	CommitTask          *Msg_CommitTask          `json:",omitempty"`

	Status MsgStatus_type
	//if status=failed, then message is failed info
	Message string
	//complete time
	CompleteTime int64
	CreateTime   int64
}

type MsgStatus_type string

const (
	Msg_Status_Staging MsgStatus_type = "staging"
	Msg_Status_Success MsgStatus_type = "success"
	Msg_Status_Failed  MsgStatus_type = "failed"
)

//Msg_BinFile describe where the file should be save, and the
type Msg_LocalFile struct {
	To     *string
	Right  *string
	User   *string
	Base64 *string
}

type Msg_Signal struct {
	Signal      *uint32
	ProcessName *string
}

type Msg_Env struct {
	Name  *string
	Value *string
	//Rep   bool
}

type Msg_EnvRemote struct {
	Name         *string
	From         *string
	Type         *string //http, https, ftp, ftps
	RemoteUser   *string
	RemotePasswd *string
}

type Msg_Remote struct {
	To           *string
	Right        *string
	User         *string
	From         *string
	Type         *string //http, https, ftp, ftps
	RemoteUser   *string
	RemotePasswd *string
}

type Msg_Secret struct {
	Name  *string
	Value *string
	Type  *Secret_Type
}

type Msg_TaskStatusQuery struct {
	Reason *string
}

type Msg_UpdateTaskResources struct {
	Resources []*TaskResources
}

type TaskResources struct {
	TaskId *string
	Cpu    *float64
	Mem    *float64
}

type Msg_CommitTask struct {
	Tasks []*CommitTask
}

type CommitTask struct {
	TaskId *string
	Image  *string
}

func (x Msg_Type) Enum() *Msg_Type {
	p := new(Msg_Type)
	*p = x
	return p
}

func (x Secret_Type) Enum() *Secret_Type {
	p := new(Secret_Type)
	*p = x
	return p
}

type TaskFail_Reason int32

const (
	TaskFail_UNKOWN   TaskFail_Reason = 0
	TaskFail_IP_SHORT TaskFail_Reason = 1
	TaskFail_IP_USED  TaskFail_Reason = 2
)

type BCSTaskFailMsg struct {
	Reason TaskFail_Reason
	Desc   string
}
