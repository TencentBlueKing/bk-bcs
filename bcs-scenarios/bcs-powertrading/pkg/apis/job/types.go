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

package job

// StatusResponse jobStatus rsp
type StatusResponse struct {
	Code         int        `json:"code"`
	Result       bool       `json:"result"`
	JobRequestID string     `json:"job_request_id"`
	Data         StatusData `json:"data"`
	Message      string     `json:"message"`
}

// StatusData job data
type StatusData struct {
	Finished         bool               `json:"finished"`
	JobInstance      Instance           `json:"job_instance"`
	StepInstanceList []StepInstanceList `json:"step_instance_list"`
}

// Instance job instance
type Instance struct {
	BkBizID       int64  `json:"bk_biz_id"`
	BkScopeType   string `json:"bk_scope_type"`
	BkScopeID     string `json:"bk_scope_id"`
	JobInstanceID int64  `json:"job_instance_id"`
	Status        int    `json:"status"`
	Name          string `json:"name"`
	CreateTime    int64  `json:"create_time"`
	StartTime     int64  `json:"start_time"`
	EndTime       int64  `json:"end_time"`
	TotalTime     int    `json:"total_time"`
}

// StepIPResult step ip result
type StepIPResult struct {
	IP        string `json:"ip"`
	Status    int    `json:"status"`
	Tag       string `json:"tag"`
	BkHostID  int    `json:"bk_host_id"`
	BkCloudID int    `json:"bk_cloud_id"`
	ExitCode  int    `json:"exit_code"`
	ErrorCode int    `json:"error_code"`
	StartTime int64  `json:"start_time"`
	EndTime   int64  `json:"end_time"`
	TotalTime int    `json:"total_time"`
}

// StepInstanceList step instance list
type StepInstanceList struct {
	StepInstanceID   int64          `json:"step_instance_id"`
	Name             string         `json:"name"`
	Status           int            `json:"status"`
	Type             int            `json:"type"`
	ExecuteCount     int            `json:"execute_count"`
	StartTime        int64          `json:"start_time"`
	EndTime          int64          `json:"end_time"`
	TotalTime        int            `json:"total_time"`
	CreateTime       int64          `json:"create_time"`
	StepIPResultList []StepIPResult `json:"step_ip_result_list"`
}

// BatchLogResponse jobBatchLog response
type BatchLogResponse struct {
	Code         int          `json:"code"`
	Result       bool         `json:"result"`
	Data         BatchLogData `json:"data"`
	JobRequestID string       `json:"job_request_id"`
	Message      string       `json:"message"`
}

// BatchLogData jobBatchLog data
type BatchLogData struct {
	JobInstanceID  int64           `json:"job_instance_id"`
	StepInstanceID int64           `json:"step_instance_id"`
	LogType        int             `json:"log_type"`
	ScriptTaskLogs []ScriptTaskLog `json:"script_task_logs"`
	FileTaskLogs   interface{}     `json:"file_task_logs"`
}

// ScriptTaskLog scriptTask log
type ScriptTaskLog struct {
	HostID     int         `json:"host_id"`
	BkCloudID  int         `json:"bk_cloud_id"`
	IP         string      `json:"ip"`
	IPv6       interface{} `json:"ipv6"`
	LogContent string      `json:"log_content"`
}

// BatchLogIPRequest batchLogIP req
type BatchLogIPRequest struct {
	BkCloudID int    `json:"bk_cloud_id"`
	IP        string `json:"ip"`
}

// BatchLogRequest batchLogRequest
type BatchLogRequest struct {
	BkScopeType    string              `json:"bk_scope_type"`
	BkScopeID      string              `json:"bk_scope_id"`
	JobInstanceID  string              `json:"job_instance_id"`
	StepInstanceID string              `json:"step_instance_id"`
	IPList         []BatchLogIPRequest `json:"ip_list"`
}
