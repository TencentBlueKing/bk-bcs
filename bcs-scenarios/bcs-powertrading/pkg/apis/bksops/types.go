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

package bksops

const (
	// CreateTaskUrl create task
	CreateTaskUrl = "%s/create_task/%s/%s"
	// StartTaskUrl start task
	StartTaskUrl = "%s/start_task/%s/%s/"
	// GetTaskStatusUrl get task status
	GetTaskStatusUrl = "%s/get_task_status/%s/%s/"
	// GetTaskDetailUrl get task detail
	GetTaskDetailUrl = "%s/get_task_detail/%s/%s/"
	// GetTaskNodeDetailUrl get task node detail
	GetTaskNodeDetailUrl = "%s/get_task_node_detail/%s/%s/?node_id=%s"
)

// CreateTaskReq create task req
type CreateTaskReq struct {
	TemplateSource string            `json:"template_source"`
	Name           string            `json:"name"`
	FlowType       string            `json:"flow_type"`
	Constants      map[string]string `json:"constants"`
}

// CreateTaskRsp create task rsp
type CreateTaskRsp struct {
	Message string         `json:"message"`
	Result  bool           `json:"result"`
	Code    int            `json:"code"`
	TraceID string         `json:"trace_id"`
	Data    CreateTaskData `json:"data"`
}

// CreateTaskData createTask data
type CreateTaskData struct {
	TaskID       int64       `json:"task_id"`
	TaskURL      string      `json:"task_url"`
	PipelineTree interface{} `json:"pipeline_tree"`
}

// StartTaskRsp startTask rsp
type StartTaskRsp struct {
	Message string      `json:"message"`
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	TraceID string      `json:"trace_id"`
	Data    interface{} `json:"data"`
	TaskURL string      `json:"task_url"`
}

// GetTaskStatusRsp get task status rsp
type GetTaskStatusRsp struct {
	Result  bool              `json:"result"`
	Code    int               `json:"code"`
	Message string            `json:"message"`
	TraceID string            `json:"trace_id"`
	Data    GetTaskStatusData `json:"data"`
}

// GetTaskStatusData getTaskStatus data
type GetTaskStatusData struct {
	ID             string                       `json:"id"`
	State          string                       `json:"state"`
	RootID         string                       `json:"root_id:"`
	ParentID       string                       `json:"parent_id"`
	Version        string                       `json:"version"`
	Loop           int                          `json:"loop"`
	Retry          int                          `json:"retry"`
	Skip           interface{}                  `json:"skip"`
	ErrorIgnorable bool                         `json:"error_ignorable"`
	ErrorIgnored   bool                         `json:"error_ignored"`
	Children       map[string]GetTaskStatusData `json:"children"`
	ElapsedTime    int                          `json:"elapsed_time"`
	StartTime      string                       `json:"start_time"`
	FinishTime     string                       `json:"finish_time"`
	Name           string                       `json:"name"`
}

// GetTaskNodeDetailRsp getTaskNodeDetail rsp
type GetTaskNodeDetailRsp struct {
	Result  bool                  `json:"result"`
	Code    int                   `json:"code"`
	Message string                `json:"message"`
	TraceID string                `json:"trace_id"`
	Data    GetTaskNodeDetailData `json:"data"`
}

// GetTaskNodeDetailData getTaskNodeDetail data
type GetTaskNodeDetailData struct {
	ID             string                        `json:"id"`
	State          string                        `json:"state"`
	RootID         string                        `json:"root_id:"`
	ParentID       string                        `json:"parent_id"`
	Version        string                        `json:"version"`
	Loop           int                           `json:"loop"`
	Retry          int                           `json:"retry"`
	Skip           bool                          `json:"skip"`
	ErrorIgnorable bool                          `json:"error_ignorable"`
	ErrorIgnored   bool                          `json:"error_ignored"`
	Children       map[string]interface{}        `json:"children"`
	ElapsedTime    int                           `json:"elapsed_time"`
	StartTime      string                        `json:"start_time"`
	FinishTime     string                        `json:"finish_time"`
	Histories      []interface{}                 `json:"histories"`
	HistoryID      int                           `json:"history_id"`
	Inputs         map[string]interface{}        `json:"inputs"`
	Outputs        []GetTaskNodeDetailDataOutput `json:"outputs"`
	ExData         interface{}                   `json:"ex_data"`
}

// GetTaskNodeDetailDataOutput getTaskNodeDetail output
type GetTaskNodeDetailDataOutput struct {
	Key    string      `json:"key"`
	Value  interface{} `json:"value"`
	Preset bool        `json:"preset"`
}
