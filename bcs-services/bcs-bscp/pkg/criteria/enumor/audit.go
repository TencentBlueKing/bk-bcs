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

// Package enumor is enum of the audit
package enumor

/*
	audit.go store audit related enum values.
*/

// AuditResourceType audit resource type.
type AuditResourceType string

const (
	// App 应用模块资源
	App AuditResourceType = "app"
	// ConfigItem 配置项资源
	ConfigItem AuditResourceType = "config_item"
	// Commit 提交资源
	Commit AuditResourceType = "commit"
	// Content 配置内容
	Content AuditResourceType = "content"
	// Release 版本资源
	Release AuditResourceType = "release"
	// StrategySet 策略集资源
	StrategySet AuditResourceType = "strategy_set"
	// CRInstance 当前已发布的实例版本
	CRInstance AuditResourceType = "current_released_instance"
	// Strategy 策略资源
	Strategy AuditResourceType = "strategy"
	// Hook hook脚本资源
	Hook AuditResourceType = "hook"
	// TemplateSpace 模版空间
	TemplateSpace AuditResourceType = "template_space"
	// Group 分组资源
	Group AuditResourceType = "group"
	// Credential 凭据资源
	Credential AuditResourceType = "credential"
	// CredentialScope 凭据规则资源
	CredentialScope AuditResourceType = "credential_scope" //nolint:gosec

	// 新版操作记录资源类型

	// ResAppConfig 服务配置
	ResAppConfig AuditResourceType = "app_config"
	// ResGroup 分组
	ResGroup AuditResourceType = "group"
	// ResHook hook脚本资源
	ResHook AuditResourceType = "hook"
	// ResTemplate 配置模版
	ResTemplate AuditResourceType = "template"
	// ResVariable 变量
	ResVariable AuditResourceType = "variable"
	// ResCredential 客户端秘钥
	ResCredential AuditResourceType = "credential"
	// ResInstance 客户端实例
	ResInstance AuditResourceType = "instance"
)

// AuditResourceTypeEnums resource type map.
var AuditResourceTypeEnums = map[AuditResourceType]bool{
	App:             true,
	ConfigItem:      true,
	Commit:          true,
	Content:         true,
	Release:         true,
	StrategySet:     true,
	Strategy:        true,
	Hook:            true,
	TemplateSpace:   true,
	Group:           true,
	Credential:      true,
	CredentialScope: true,
}

// ActionMap enum all action
var ActionMap = map[AuditAction]AuditResourceType{
	CreateApp:            ResAppConfig,
	UpdateApp:            ResAppConfig,
	DeleteApp:            ResAppConfig,
	PublishVersionConfig: ResAppConfig, // 上线配置版本
}

// Exist judge enum value exist.
func (a AuditResourceType) Exist() bool {
	_, exist := AuditResourceTypeEnums[a]
	return exist
}

// AuditAction audit action type.
type AuditAction string

const (
	// Create 创建
	Create AuditAction = "Create"
	// Update 更新
	Update AuditAction = "Update"
	// Delete 删除
	Delete AuditAction = "Delete"
	// Publish 发布
	Publish AuditAction = "Publish"
	// FinishPublish 结束发布
	FinishPublish AuditAction = "FinishPublish"
	// Rollback 回滚版本
	Rollback AuditAction = "Rollback"
	// Reload Reload配置版本
	Reload AuditAction = "Reload"

	// CreateApp 创建服务
	CreateApp AuditAction = "CreateApp"
	// UpdateApp 更新服务
	UpdateApp AuditAction = "UpdateApp"
	// DeleteApp 删除服务
	DeleteApp AuditAction = "DeleteApp"
	// PublishVersionConfig 上线版本配置
	PublishVersionConfig AuditAction = "PublishVersionConfig"
)

// AuditActionEnums op type map.
var AuditActionEnums = map[AuditAction]bool{
	Create:        true,
	Update:        true,
	Delete:        true,
	Publish:       true,
	FinishPublish: true,
	Rollback:      true,
	Reload:        true,
}

// Exist judge enum value exist.
func (a AuditAction) Exist() bool {
	_, exist := AuditActionEnums[a]
	return exist
}

// AuditStatus audit status.
type AuditStatus string

const (
	// Success audit status
	Success AuditStatus = "Success"
	// Failure audit status
	Failure AuditStatus = "Failure"
	// PendApproval means this strategy audit status is pending.
	PendApproval AuditStatus = "PendApproval"
	// PendPublish means this strategy audit status is pending.
	PendPublish AuditStatus = "PendPublish"
	// RevokedPublish means this strategy audit status is revoked.
	RevokedPublish AuditStatus = "RevokedPublish"
	// RejectedApproval means this strategy audit status is rejected.
	RejectedApproval AuditStatus = "RejectedApproval"
	// AlreadyPublish means this strategy audit status is already publish.
	AlreadyPublish AuditStatus = "AlreadyPublish"
)

// AuditOperateWay audit operate way.
type AuditOperateWay string

const (
	// WebUI audit operate way
	WebUI AuditOperateWay = "WebUI"
	// API audit operate way
	API AuditOperateWay = "API"
)
