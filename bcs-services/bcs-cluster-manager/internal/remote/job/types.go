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

import (
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/types"
)

// Options for client
type Options struct {
	AppCode    string
	AppSecret  string
	BKUserName string
	Server     string
	Debug      bool
}

// query parameters
const (
	bkScopeType   = "bk_scope_type"   // nolint
	bkScopeID     = "bk_scope_id"     // nolint
	jobInstanceID = "job_instance_id" // nolint
)

// ScopeType bizScope
type ScopeType string

var (
	// Biz 业务
	Biz ScopeType = "biz"
	// BizSet 业务集
	BizSet ScopeType = "biz_set"
)

const (
	shell = iota + 1
	bat
	perl
	python
	powershell
)

// FastExecuteScriptReq execute script request body
type FastExecuteScriptReq struct {
	BkScopeType    ScopeType    `json:"bk_scope_type"`
	BkScopeID      string       `json:"bk_scope_id"`
	ScriptContent  string       `json:"script_content"`  // 脚本内容Base64
	TaskName       string       `json:"task_name"`       // 自定义作业名称
	ScriptParam    string       `json:"script_param"`    // 脚本参数Base64
	Timeout        uint64       `json:"timeout"`         // 默认 7200s
	AccountAlias   string       `json:"account_alias"`   // 默认 root
	ScriptLanguage int          `json:"script_language"` // 1 - shell, 2 - bat, 3 - perl, 4 - python, 5 - powershell。
	TargetServer   TargetServer `json:"target_server"`
}

func buildTaskName(bizID string) string {
	return fmt.Sprintf("job 执行业务[%s]脚本", bizID)
}

func transToBkJobExecuteScriptReq(paras ExecuteScriptParas) *FastExecuteScriptReq {
	taskName := buildTaskName(paras.BizID)
	if paras.TaskName != "" {
		taskName = paras.TaskName
	}

	req := &FastExecuteScriptReq{
		BkScopeType:    Biz,
		BkScopeID:      paras.BizID,
		ScriptContent:  paras.ScriptContent,
		TaskName:       taskName,
		ScriptParam:    paras.ScriptParas,
		Timeout:        7200,
		AccountAlias:   "root",
		ScriptLanguage: shell, // 默认shell脚本
		TargetServer:   TargetServer{},
	}
	if req.TargetServer.IpList == nil {
		req.TargetServer.IpList = make([]IpInfo, 0)
	}
	// append target server
	for _, server := range paras.Servers {
		req.TargetServer.IpList = append(req.TargetServer.IpList, IpInfo(server))
	}

	return req
}

// TargetServer server info
type TargetServer struct {
	IpList []IpInfo `json:"ip_list"`
}

// IpInfo ip detailed info
type IpInfo struct {
	BkCloudID uint64 `json:"bk_cloud_id"`
	Ip        string `json:"ip"`
}

// FastExecuteScriptRsp resp xxx
type FastExecuteScriptRsp struct {
	types.BaseResponse
	Data FastExecuteScripData `json:"data"`
}

// FastExecuteScripData xxx
type FastExecuteScripData struct {
	JobInstanceName string `json:"job_instance_name"`
	JobInstanceID   uint64 `json:"job_instance_id"`
	StepInstanceID  uint64 `json:"step_instance_id"`
}

// GetJobInstanceStatusRsp xxx
type GetJobInstanceStatusRsp struct {
	types.BaseResponse
	Data GetJobInstanceStatusData `json:"data"`
}

// GetJobInstanceStatusData xxx
type GetJobInstanceStatusData struct {
	Finished    bool              `json:"finished"`
	JobInstance JobInstanceStatus `json:"job_instance"`
}

// JobInstanceStatus status
type JobInstanceStatus struct { // nolint
	Name string `json:"name"`
	// 作业状态码: 1.未执行; 2.正在执行; 3.执行成功; 4.执行失败; 5.跳过;
	//  6.忽略错误; 7.等待用户; 8.手动结束; 9.状态异常; 10.步骤强制终止中; 11.步骤强制终止成功
	Status        int    `json:"status"`
	JobInstanceID uint64 `json:"job_instance_id"`
	CreateTime    uint64 `json:"create_time"`
	StartTime     uint64 `json:"start_time"`
	EndTime       uint64 `json:"end_time"`
	TotalTime     uint64 `json:"total_time"`
}

func transJobStatus(status int) int {
	switch status {
	case 1:
		return UnExecuted
	case 2:
		return Executing
	case 3:
		return ExecuteSuccess
	case 4:
		return ExecuteFailure
	default:
	}

	return UnKnownStatus
}
