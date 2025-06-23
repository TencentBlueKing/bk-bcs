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

// Package iam xxx
package iam

import (
	"strings"

	"github.com/TencentBlueKing/iam-go-sdk/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
)

// CloudAccountProvider is a cloud account provider
type CloudAccountProvider struct {
}

func init() {
	dispatcher.RegisterProvider(CloudAccount, CloudAccountProvider{})
}

// ListAttr implements the list_attr
func (p CloudAccountProvider) ListAttr(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: []interface{}{},
	}
}

// ListAttrValue implements the list_attr_value
func (p CloudAccountProvider) ListAttrValue(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// ListInstance implements the list_instance
func (p CloudAccountProvider) ListInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if filter.Parent.ID == "" {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	result, err := component.ListCloudAccount(req.Context, filter.Parent.ID, nil)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		ins := Instance{r.AccountID, r.AccountName, nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}

// FetchInstanceInfo implements the fetch_instance_info
func (p CloudAccountProvider) FetchInstanceInfo(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.IDs) == 0 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "ids is empty",
			Data:    []interface{}{},
		}
	}

	result, err := component.ListCloudAccount(req.Context, "", filter.IDs)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		ins := Instance{r.AccountID, r.AccountName, []string{r.Creator, r.Updater}}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: results,
	}
}

// ListInstanceByPolicy implements the list_instance_by_policy
func (p CloudAccountProvider) ListInstanceByPolicy(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// SearchInstance implements the search_instance
func (p CloudAccountProvider) SearchInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if filter.Parent.ID == "" {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	result, err := component.ListCloudAccount(req.Context, filter.Parent.ID, nil)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		if filter.Keyword != "" && !strings.Contains(r.AccountName, filter.Keyword) {
			continue
		}
		ins := Instance{r.AccountID, r.AccountName, nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}

// FetchInstanceList implements the fetch_instance_list
func (p CloudAccountProvider) FetchInstanceList(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// FetchResourceTypeSchema implements the fetch_resource_type_schema
func (p CloudAccountProvider) FetchResourceTypeSchema(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}
