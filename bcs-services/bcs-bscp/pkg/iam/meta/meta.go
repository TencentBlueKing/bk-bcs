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

// Package meta NOTES
package meta

// UserInfo user info for authorization use.
type UserInfo struct {
	// UserName the name of this user.
	UserName string
}

// ResourceAttribute represent one iam resource
type ResourceAttribute struct {
	*Basic
	// BizID biz id of the iam resource.
	BizID uint32 `json:"biz_id,omitempty"`
	// GenApplyURL 是否生成申请链接
	GenApplyURL bool `json:"gen_apply_url,omitempty"`
}

// Basic defines the basic info for a resource.
type Basic struct {
	// Type the type of the resource.
	Type ResourceType `json:"type"`

	// Action the action that user want to do with this resource.
	Action Action `json:"action"`

	// ResourceID the instance id of this resource.
	ResourceID uint32
}

// BasicDetail add deail for auth api
type BasicDetail struct {
	Basic        `json:",inline"`
	TypeName     string `json:"type_name"`
	ActionName   string `json:"action_name"`
	ResourceName string `json:"resource_name"`
}

// Decision defines the authorization decision of a resource.
type Decision struct {
	// Authorized the authorization decision, whether a user has permission to the resource or not.
	Authorized bool
}

// IamPermission defines the iam permission, used to show user which permission to apply and generate iam apply url.
type IamPermission struct {
	SystemID   string       `json:"system_id"`
	SystemName string       `json:"system_name"`
	Actions    []*IamAction `json:"actions"`
}

// IamAction defines the iam permission with related resource info.
type IamAction struct {
	ID                   string             `json:"id"`
	Name                 string             `json:"name"`
	RelatedResourceTypes []*IamResourceType `json:"related_resource_types"`
}

// IamResourceType defines the iam resource with instance info.
type IamResourceType struct {
	SystemID   string                   `json:"system_id"`
	SystemName string                   `json:"system_name"`
	Type       string                   `json:"type"`
	TypeName   string                   `json:"type_name"`
	Instances  [][]*IamResourceInstance `json:"instances,omitempty"`
	Attributes []*IamResourceAttribute  `json:"attributes,omitempty"`
}

// IamResourceInstance defines the iam resource instance info.
type IamResourceInstance struct {
	Type     string `json:"type"`
	TypeName string `json:"type_name"`
	ID       string `json:"id"`
	Name     string `json:"name"`
}

// IamResourceAttribute defines the iam resource attribute info.
type IamResourceAttribute struct {
	ID     string                       `json:"id"`
	Values []*IamResourceAttributeValue `json:"values"`
}

// IamResourceAttributeValue defines the iam resource attribute value info.
type IamResourceAttributeValue struct {
	ID string `json:"id"`
}
