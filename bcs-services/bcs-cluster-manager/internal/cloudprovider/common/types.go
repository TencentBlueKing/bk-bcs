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

package common

// TplSource template_source, 流程模版来源
type TplSource string

var (
	// BusinessTpl 业务流程,默认
	BusinessTpl TplSource = "business"
	// CommonTpl 公共流程
	CommonTpl TplSource = "common"
)

// FlowType flow_type, 任务流程类型
type FlowType string

var (
	// CommonFlow 常规流程
	CommonFlow FlowType = "common"
	// CommonFuncFlow 职能化流程
	CommonFuncFlow FlowType = "common_func"
)

// Scope bk_biz_id 检索的作用域
type Scope string

var (
	// CmdbBizScope 检索的是绑定的CMDB业务ID为bk_biz_id的项目
	CmdbBizScope Scope = "cmdb_biz"
	// ProjectBizScope 检索项目ID为bk_biz_id的项目
	ProjectBizScope Scope = "project"
)

// TaskState bkops task status
type TaskState string

// String toString
func (ts TaskState) String() string {
	return string(ts)
}

const (
	// CREATED TaskState 未执行
	CREATED TaskState = "CREATED"
	// RUNNING TaskState 执行中
	RUNNING TaskState = "RUNNING"
	// FAILED TaskState	失败
	FAILED TaskState = "FAILED"
	// SUSPENDED TaskState	暂停
	SUSPENDED TaskState = "SUSPENDED"
	// REVOKED TaskState 已终止
	REVOKED TaskState = "REVOKED"
	// FINISHED TaskState 已完成
	FINISHED TaskState = "FINISHED"
)

// CreateTaskPathParas task path paras
type CreateTaskPathParas struct {
	// BkBizID template bizID
	BkBizID string `json:"bk_biz_id"`
	// TemplateID
	TemplateID string `json:"template_id"`
	// Operator template perm user
	Operator string `json:"operator"`
}

// CreateTaskRequest create task req
type CreateTaskRequest struct {
	// attention: community version bksops parameters in requestBody
	// BusinessID template biz_id
	BusinessID string `json:"bk_biz_id,omitempty"`
	// TemplateID template_id
	TemplateID string `json:"template_id,omitempty"`
	// TemplateSource 模版来源(business/common)
	TemplateSource string `json:"template_source"`
	// TaskName 任务名称
	TaskName string `json:"name"`
	// FlowType 任务流程类型 (默认 common即可)
	FlowType string `json:"flow_type"`
	// Constants  任务全局参数
	Constants map[string]string `json:"constants"`
}

// CreateTaskResponse create task resp
type CreateTaskResponse struct {
	Result  bool     `json:"result"`
	Data    *ResData `json:"data"`
	Message string   `json:"message"`
}

// ResData resp data
type ResData struct {
	TaskID  int    `json:"task_id"`
	TaskURL string `json:"task_url"`
}

// TaskReqParas task request body
type TaskReqParas struct {
	BkBizID string `json:"bk_biz_id"`
	TaskID  string `json:"task_id"`
}

// TaskPathParas task path paras
type TaskPathParas struct {
	BkBizID  string `json:"bk_biz_id"`
	TaskID   string `json:"task_id"`
	Operator string `json:"operator"`
}

// StartTaskRequest request
type StartTaskRequest struct {
	Scope string `json:"scope"`
}

// StartTaskResponse start task response
type StartTaskResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// TaskStatusResponse task status response
type TaskStatusResponse struct {
	Result  bool        `json:"result"`
	Data    *StatusData `json:"data"`
	Message string      `json:"message"`
}

// StatusData status
type StatusData struct {
	State string `json:"state"`
}

// TemplateListPathPara path parameter
type TemplateListPathPara struct {
	// BkBizID template bizID
	BkBizID string `json:"bk_biz_id"`
	// Operator template perm user
	Operator string `json:"operator"`
}

// TemplateRequest body parameter
type TemplateRequest struct {
	// BkBizID template bizID
	BkBizID string `json:"bk_biz_id,omitempty"`
	// TemplateID template id
	TemplateID string `json:"template_id,omitempty"`
	// TemplateSource 模版来源(business/common)
	TemplateSource string `json:"template_source"`
	// Scope bk_biz_id 检索的作用域。默认为 cmdb_biz，此时检索的是绑定的 CMDB 业务 ID 为 bk_biz_id 的项目；
	// 当值为 project 时则检索项目 ID 为 bk_biz_id 的项目
	Scope Scope `json:"scope"`
}

// SetDefaultTemplateBody xxx
func (tr *TemplateRequest) SetDefaultTemplateBody() {
	// 默认业务流程
	if tr.TemplateSource == "" {
		tr.TemplateSource = string(BusinessTpl)
	}
	// 默认bk_biz_id检索作用域, 默认为 cmdb_biz，此时检索的是绑定的 CMDB 业务 ID 为 bk_biz_id 的项目
	if string(tr.Scope) == "" {
		tr.Scope = CmdbBizScope
	}
}

// TemplateListResponse templateList response
type TemplateListResponse struct {
	Result  bool            `json:"result"`
	Data    []*TemplateData `json:"data"`
	Message string          `json:"message"`
}

// TemplateData template list
type TemplateData struct {
	// BkBizID business id
	BkBizID int64 `json:"bk_biz_id"`
	// BkBizName business name
	BkBizName string `json:"bk_biz_name"`
	// ID template id
	ID int `json:"id"`
	// Name template name
	Name string `json:"name"`
	// Creator template creator
	Creator string `json:"creator"`
	// Editor template editor
	Editor string `json:"editor"`
}

// TemplateDetailPathPara path parameter
type TemplateDetailPathPara struct {
	// BkBizID template bizID
	BkBizID string `json:"bk_biz_id"`
	// TemplateID template id
	TemplateID string `json:"template_id"`
	// Operator template perm user
	Operator string `json:"operator"`
}

// TemplateDetailResponse template data response
type TemplateDetailResponse struct {
	Result  bool               `json:"result"`
	Data    TemplateDetailData `json:"data"`
	Message string             `json:"message"`
}

// TemplateDetailData template detail data
type TemplateDetailData struct {
	TemplateData
	PipeTree PipelineTree `json:"pipeline_tree"`
}

// PipelineTree constants values
type PipelineTree struct {
	Constants map[string]ConstantValue `json:"constants"`
}

const (
	custom           = "custom"
	componentInputs  = "component_inputs"
	componentOutputs = "component_outputs"
)

// ConstantValue constant value
type ConstantValue struct {
	// Key 同KEY
	Key string `json:"key"`
	// Name 变量名字
	Name string `json:"name"`
	// Index 变量在模板中的显示顺序
	Index int `json:"index"`
	// Desc 变量说明
	Desc string `json:"desc"`
	// SourceType 变量来源，取值范围 custom: 自定义变量，component_inputs: 从标准插件输入参数勾选，component_outputs：从标准插件输出结果中勾选
	SourceType string `json:"source_type"`
	// CustomType source_type=custom 时有效，自定义变量类型， 取值范围 input: 输入框，textarea: 文本框，datetime: 日期时间，int: 整数
	CustomType string `json:"custom_type"`
}

// UserProjectRequest project req
type UserProjectRequest struct {
	BizId string `json:"bk_biz_id"`
	Scope Scope  `json:"scope"`
}

// UserProjectResponse project info
type UserProjectResponse struct {
	Result  bool        `json:"result"`
	Message string      `json:"message"`
	Data    ProjectInfo `json:"data"`
}

// ProjectInfo project
type ProjectInfo struct {
	ProjectId       int    `json:"project_id"`
	ProjectName     string `json:"project_name"`
	BkBizId         int    `json:"bk_biz_id"`
	FromCmdb        bool   `json:"from_cmdb"`
	BkBizName       string `json:"bk_biz_name"`
	BkBizDeveloper  string `json:"bk_biz_developer"`
	BkBizMaintainer string `json:"bk_biz_maintainer"`
	BkBizTester     string `json:"bk_biz_tester"`
	BkBizProductor  string `json:"bk_biz_productor"`
}

// Action bkops task action
type Action string

// String toString
func (ac Action) String() string {
	return string(ac)
}

const (
	// Start action  开始任务，等效于调用 start_task 接口
	Start Action = "start"
	// Pause action 暂停任务，任务处于执行状态时调用
	Pause Action = "pause"
	// Resume action 继续任务，任务处于暂停状态时调用	失败
	Resume Action = "resume"
	// Revoke action 终止任务
	Revoke TaskState = "revoke"
)

// OperateTaskRequest operate task request
type OperateTaskRequest struct {
	BizId  string `json:"bk_biz_id,omitempty"`
	TaskId string `json:"task_id,omitempty"`
	Action string `json:"action"`
	Scope  Scope  `json:"scope"`
}

// OperateTaskResponse operate task response
type OperateTaskResponse struct {
	Result  bool   `json:"result"`
	Message string `json:"message"`
}

// SyncClusterDataRequest sync cluster data to storage request
type SyncClusterDataRequest struct {
	Data map[string]interface{} `json:"data"`
}

// StorageResponse sync cluster data to storage response
type StorageResponse struct {
	Result  bool        `json:"result"`
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}
