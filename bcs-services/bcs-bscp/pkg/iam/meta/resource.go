/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package meta

// ResourceType 表示 bscp 这一侧的资源类型， 对应的有 client.TypeID 表示 iam 一侧的资源类型
// 两者之间有映射关系，详情见 AdaptAuthOptions
type ResourceType string

// String convert ResourceType to string.
func (r ResourceType) String() string {
	return string(r)
}

const (
	// Biz 业务
	Biz ResourceType = "biz"
	// App resource's bscp auth resource type
	App ResourceType = "app"
	// Commit resource's bscp auth resource type
	Commit ResourceType = "commit"
	// ConfigItem resource's bscp auth resource type
	ConfigItem ResourceType = "config_item"
	// Content resource's bscp auth resource type
	Content ResourceType = "content"
	// CRInstance means current released instance resource's bscp auth resource type
	CRInstance ResourceType = "current_released_instance"
	// ReleasedCI resource's bscp auth resource type
	ReleasedCI ResourceType = "released_config_item"
	// Release resource's bscp auth resource type
	Release ResourceType = "release"
	// Strategy resource's bscp auth resource type
	Strategy ResourceType = "strategy"
	// StrategySet resource's bscp auth resource type
	StrategySet ResourceType = "strategy_set"
	// Hook resource's bscp auth resource type
	Hook ResourceType = "hook"
	// TemplateSpace resource's bscp auth resource type
	TemplateSpace ResourceType = "template_space"
	// Group resource's bscp auth resource type
	Group ResourceType = "group"
	// PSH resource's bscp auth resource type
	PSH ResourceType = "published_strategy_history"
	// Repo represents repository resource's related bscp auth resource type
	Repo ResourceType = "repository"
	// Sidecar represent requests form sidecar
	Sidecar ResourceType = "sidecar"
	// Credential resource's bscp auth resource type
	Credential ResourceType = "credential"
	// CredentialScope resource's bscp auth resource type
	CredentialScope ResourceType = "credential_scope"
)
