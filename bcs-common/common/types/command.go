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

type BcsCommand struct {
	TypeMeta `json:",inline"`
	//ObjectMeta `json:"metadata"`
	Spec *BcsCommandSpec `json:"spec"`
}

type BcsCommandSpec struct {
	CommandTargetRef *TargetRef `json:"commandTargetRef"`
	Taskgroups       []string   `json:"taskgroups"`
	Command          []string   `json:"command"` //["/bin/bash","-c","ps -ef |grep gamesvc"]
	Env              []string   `json:"env"`     //environments
	User             string     `json:"user"`    //root or others
	WorkingDir       string     `json:"workingDir"`
	Privileged       bool       `json:"privileged"`
	ReserveTime      int        `json:"reserveTime"` //minutes
}

type BcsCommandStatus struct {
	Taskgroups []*TaskgroupCommandInfo `json:"taskgroups"`
}

type BcsCommandInfo struct {
	Id         string            `json:"id"` //command inspect info id
	CreateTime int64             `json:"createTime"`
	Spec       *BcsCommandSpec   `json:"spec"`
	Status     *BcsCommandStatus `json:"status"`
}

type TaskgroupCommandInfo struct {
	TaskgroupId string             `json:"taskgroupId"`
	Tasks       []*TaskCommandInfo `json:"tasks"`
}

type TaskCommandInfo struct {
	TaskId      string              `json:"taskId"` //application taskid
	Status      TaskCommandStatus   `json:"status"`
	Message     string              `json:"message"`
	CommInspect *CommandInspectInfo `json:"commInspect"`
}

type TaskCommandStatus string

const (
	TaskCommandStatusStaging TaskCommandStatus = "staging"
	TaskCommandStatusRunning TaskCommandStatus = "running"
	TaskCommandStatusFinish  TaskCommandStatus = "finish"
	TaskCommandStatusFailed  TaskCommandStatus = "failed"
)

type CommandInspectInfo struct {
	ExitCode int    `json:"exitCode"` // command exitcode,  0: success, >0: failed
	Stdout   string `json:"stdout"`   //command stdout, maxbyte 1024
	Stderr   string `json:"stderr"`   //command stderr, maxbyte 1024
}
