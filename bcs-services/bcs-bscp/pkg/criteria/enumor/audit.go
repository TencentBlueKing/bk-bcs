/*
Tencent is pleased to support the open source community by making Basic Service Configuration Platform available.
Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
Licensed under the MIT License (the "License"); you may not use this file except
in compliance with the License. You may obtain a copy of the License at
http://opensource.org/licenses/MIT
Unless required by applicable law or agreed to in writing, software distributed under
the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
either express or implied. See the License for the specific language governing permissions and
limitations under the License.
*/

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
	CredentialScope AuditResourceType = "credential_scope"
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
