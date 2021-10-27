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
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/mesosproto/mesos"
	bcstype "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schedTypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

type ProcessTaskInfo struct {
	TaskId     string
	LocalFiles []*LocalFile

	ProcInfo *ProcessInfo
	Status   TaskStatus

	//bcs executor report to bcs scheduler process info
	bcsInfo *BcsProcessInfo
}

//process task status type
type TaskStatus string

const (
	TaskStatusStaging  TaskStatus = "staging"
	TaskStatusStarting TaskStatus = "starting"
	TaskStatusRunning  TaskStatus = "running"
	TaskStatusKilling  TaskStatus = "killing"
	TaskStatusFailed   TaskStatus = "failed"
	TaskStatusFinish   TaskStatus = "finish"
	TaskStatusError    TaskStatus = "error"
)

//BcsKV key/value structure for anywhere necessary
type BcsKV struct {
	Key   string
	Value string
}

//local file
type LocalFile struct {
	To    string
	Right string
	User  string
	Value string
}

type ProcessInfo struct {
	Id string

	//process info
	WorkDir          string //进程工作目录
	ProcessName      string //进程名，pid文件所对应的名称
	Uris             []*Uri //process packages uris
	PidFile          string //process pid file path
	StartCmd         string //process start command
	StartGracePeriod int64  //start process grace period seconds
	StopCmd          string //process stop command
	StopTimeout      int
	KillCmd          string //kill -9
	Resource         *bcstype.Resource
	Envs             []string // in the form "key=value".
	Argv             []string
	User             string

	//status info
	StatusInfo *ProcessStatusInfo

	//exexutor
	ExecutorId            string //process executor id
	ExecutorHeartBeatTime int64  //process daemon & process executor last heartbeat time
}

type Uri struct {
	Value        string //process package registry uri, example for "http://xxx.artifactory.xxx.com/xxx/v1/pack.tar.gz"
	User         string //package registry user
	Pwd          string //package registry password, example for curl -u 'user:pwd' -X GET "http://xxx.artifactory.xxx.com/xxx/v1/pack.tar.gz"
	OutputDir    string
	ExtractDir   string
	PackagesFile string
}

type ProcessStatusInfo struct {
	Id            string
	Status        ProcessStatusType
	ExitCode      int    //'0' show finish, >'0' show failed
	Message       string //if failed,then message
	Pid           int
	RegisterTime  int64
	LastStartTime int64
}

type ProcessStatusType string

const (
	ProcessStatusStaging  ProcessStatusType = "staging"
	ProcessStatusStarting ProcessStatusType = "starting"
	ProcessStatusRunning  ProcessStatusType = "running"
	ProcessStatusStopping ProcessStatusType = "stopping"
	ProcessStatusStopped  ProcessStatusType = "stopped"
)

//CallbackFuncType
type CallbackFuncType string

const (
	CallbackFuncUpdateTask CallbackFuncType = "UpdateTaskFunc"
)

type UpdateTaskFunc func(*mesos.TaskStatus) error

//ExecutorStatus type
type ExecutorStatus string

const (
	ExecutorStatusUnknown   ExecutorStatus = "unknown"
	ExecutorStatusLaunching ExecutorStatus = "launching"
	ExecutorStatusRunning   ExecutorStatus = "running"
	ExecutorStatusShutdown  ExecutorStatus = "shutdown"
	ExecutorStatusFinish    ExecutorStatus = "finish"
)

//BcsProcessInfo only for BcsExecutor
type BcsProcessInfo struct {
	ID          string                 `json:"ID,omitempty"`          //container ID
	Name        string                 `json:"Name,omitempty"`        //container name
	Pid         int                    `json:"Pid,omitempty"`         //container pid
	StartAt     time.Time              `json:"StartAt,omitempty"`     //startting time
	FinishAt    time.Time              `json:"FinishAt,omitempty"`    //Exit time
	Status      string                 `json:"Status,omitempty"`      //status string, paused, restarting, running, dead, created, exited
	Healthy     bool                   `json:"Healthy,omitempty"`     //Container healthy
	ExitCode    int                    `json:"ExitCode,omitempty"`    //container exit code
	Hostname    string                 `json:"Hostname,omitempty"`    //container host name
	NetworkMode string                 `json:"NetworkMode,omitempty"` //Network mode for container
	IPAddress   string                 `json:"IPAddress,omitempty"`   //Contaienr IP address
	NodeAddress string                 `json:"NodeAddress,omitempty"` //node host address
	Ports       []BcsPort              `json:"Ports,omitempty"`       //ports info for report
	Message     string                 `json:"Message,omitempty"`     //status message for container
	Resource    *schedTypes.Resource   `json:"Resource,omitempty"`
	BcsMessage  *schedTypes.BcsMessage `json:",omitempty"`
}

//BcsPort port service for process port reflection
type BcsPort struct {
	Name          string `json:"name,omitempty"`
	ContainerPort string `json:"containerPort,omitempty"`
	HostPort      string `json:"hostPort,omitempty"`
	Protocol      string `json:"protocol,omitempty"`
	HostIP        string `json:"hostIP,omitempty"` //use for host has multiple ip address
}

//executor & process daemon heartbeat mechanism
type HeartBeat struct {
	ProcessId  string
	ExecutorId string
	Type       HeartBeatType
}

type HeartBeatType string

const (
	HeartBeatPing HeartBeatType = "ping"
	HeartBeatPong HeartBeatType = "pong"
)

type JfrogRegistry struct {
	DownloadUri string
	Checksums   *CheckSum
}

type CheckSum struct {
	Md5 string
}
