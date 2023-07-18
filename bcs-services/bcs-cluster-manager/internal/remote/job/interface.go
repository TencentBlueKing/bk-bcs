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

package job

import "context"

// JobInterface job interface method
type JobInterface interface {
	// ExecuteScript execute script on the server
	ExecuteScript(ctx context.Context, paras ExecuteScriptParas) (uint64, error)
	// GetJobStatus get job status
	GetJobStatus(ctx context.Context, job JobInfo) (int, error)
}

// 1.未执行; 2.正在执行; 3.执行成功; 4.执行失败; 5.跳过; 6.忽略错误; 7.等待用户; 8.手动结束; 9.状态异常; 10.步骤强制终止中; 11.步骤强制终止成功; 12.步骤强制终止失败
const (
	UnExecuted     = iota + 1 // 1
	Executing                 // 2
	ExecuteSuccess            // 3
	ExecuteFailure            // 4
	UnKnownStatus
)

// ExecuteScriptParas xxx
type ExecuteScriptParas struct {
	TaskName      string
	BizID         string
	ScriptContent string
	ScriptParas   string
	Servers       []ServerInfo
}

// ServerInfo xxx
type ServerInfo struct {
	BkCloudID uint64
	Ip        string
}

// JobInfo xxx
type JobInfo struct {
	BizID string
	JobID uint64
}
