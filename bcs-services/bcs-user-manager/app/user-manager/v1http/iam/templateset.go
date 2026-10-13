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

package iam

import (
	"github.com/TencentBlueKing/iam-go-sdk/resource"
)

// TemplateSetProvider is an template set provider
type TemplateSetProvider struct {
}

func init() {
	dispatcher.RegisterProvider(TemplateSet, TemplateSetProvider{})
}

// ListAttr implements the list_attr
func (p TemplateSetProvider) ListAttr(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// ListAttrValue implements the list_attr_value
func (p TemplateSetProvider) ListAttrValue(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// ListInstance implements the list_instance
func (p TemplateSetProvider) ListInstance(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// FetchInstanceInfo implements the fetch_instance_info
func (p TemplateSetProvider) FetchInstanceInfo(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// ListInstanceByPolicy implements the list_instance_by_policy
func (p TemplateSetProvider) ListInstanceByPolicy(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// SearchInstance implements the search_instance
func (p TemplateSetProvider) SearchInstance(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// FetchInstanceList implements the fetch_instance_list
func (p TemplateSetProvider) FetchInstanceList(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// FetchResourceTypeSchema implements the fetch_resource_type_schema
func (p TemplateSetProvider) FetchResourceTypeSchema(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}
