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

// Package dbm x
package dbm

import (
	"time"
)

var defaultTimeOut = time.Second * 60

const (
	taskStatusPending   = "PENDING"
	taskStatusRunning   = "RUNNING"
	taskStatusSucceeded = "SUCCEEDED"
	taskStatusFailed    = "FAILED"
	taskStatusRevoked   = "REVOKED"
)

// AuthDbmClient AuthDbmClient
type AuthDbmClient struct {
	Host        string `json:"host"`
	Environment string `json:"env"`
	Debug       bool   `json:"debug,omitempty"`
	AppCode     string `json:"-"`
	AppSecret   string `json:"-"`
	Operator    string `json:"-"`
	Task        Task   `json:"-"`
}

// AuthorizeRequest DBM authorize apply request
type AuthorizeRequest struct {
	BKBizID        int64  `json:"bk_biz_id"`
	User           string `json:"user"`
	AccessDB       string `json:"access_db"`
	SourceIPs      string `json:"source_ips"`
	TargetInstance string `json:"target_instance"`
	Operator       string `json:"operator,omitempty"`
	App            string `json:"app,omitempty"`
	Type           string `json:"type,omitempty"`
	SetName        string `json:"set_name,omitempty"`
	ModuleHostInfo string `json:"module_host_info,omitempty"`
	ModuleNameList string `json:"module_name_list,omitempty"`
	CallFrom       string `json:"call_from,omitempty"`
}

// Task DBM authorize apply task
type Task struct {
	TaskID   string `json:"task_id"`
	Platform string `json:"platform"`
}

// AuthorizeResponse DBM authorize apply response
type AuthorizeResponse struct {
	Task      *Task  `json:"data"`
	Code      int64  `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
}

// QueryResponse DBM query authorize apply result response
type QueryResponse struct {
	Status    *TaskStatus `json:"data"`
	Code      int64       `json:"code"`
	Message   string      `json:"message"`
	RequestID string      `json:"request_id"`
}

// TaskStatus DBM authorize apply task status
type TaskStatus struct {
	Status string `json:"status"`
	Msg    string `json:"msg"`
}
