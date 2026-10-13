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
	"context"
	"sync"

	"github.com/TencentBlueKing/iam-go-sdk/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
	util "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/utils"
)

// ProjectProvider is a project provider
type ProjectProvider struct {
}

func init() {
	dispatcher.RegisterProvider(Project, ProjectProvider{})
}

// ListAttr implements the list_attr
func (p ProjectProvider) ListAttr(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: []interface{}{},
	}
}

// ListAttrValue implements the list_attr_value
func (p ProjectProvider) ListAttrValue(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// ListInstance implements the list_instance
func (p ProjectProvider) ListInstance(req resource.Request) resource.Response {
	offset := req.Page.Offset / req.Page.Limit
	tenantID := req.Header.Get(util.HeaderTenantID)
	result, err := component.QueryProjects(req.Context, tenantID, req.Page.Limit, offset, nil)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result.Results {
		ins := Instance{r.ProjectID, combineNameID(r.Name, r.GetProjectCode()), nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: result.Total, Results: results},
	}
}

// FetchInstanceInfo implements the fetch_instance_info
func (p ProjectProvider) FetchInstanceInfo(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.IDs) == 0 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "ids is empty",
			Data:    []interface{}{},
		}
	}

	nsChan := make(chan Instance, len(filter.IDs))
	results := make([]interface{}, 0)
	wg := sync.WaitGroup{}
	for _, v := range filter.IDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			ctx := context.Background()
			ctx = context.WithValue(ctx, utils.ContextValueKeyRequestID, utils.GetRequestIDFromContext(req.Context))
			p, err := component.GetProject(ctx, id)
			if err != nil {
				blog.Log(ctx).Errorf("get project %s failed, err %s", id, err.Error())
				return
			}
			nsChan <- Instance{p.ProjectID, combineNameID(p.Name, p.GetProjectCode()), SplitString(p.Managers)}
		}(v)
	}
	wg.Wait()
	close(nsChan)

	for r := range nsChan {
		results = append(results, r)
	}
	return resource.Response{
		Code: 0,
		Data: results,
	}
}

// ListInstanceByPolicy implements the list_instance_by_policy
func (p ProjectProvider) ListInstanceByPolicy(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// SearchInstance implements the search_instance
func (p ProjectProvider) SearchInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	tenantID := req.Header.Get(util.HeaderTenantID)
	offset := req.Page.Offset / req.Page.Limit
	params := map[string]string{"searchName": filter.Keyword}
	result, err := component.QueryProjects(req.Context, tenantID, req.Page.Limit, offset, params)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result.Results {
		ins := Instance{r.ProjectID, combineNameID(r.Name, r.GetProjectCode()), nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: result.Total, Results: results},
	}
}

// FetchInstanceList implements the fetch_instance_list
func (p ProjectProvider) FetchInstanceList(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}

// FetchResourceTypeSchema implements the fetch_resource_type_schema
func (p ProjectProvider) FetchResourceTypeSchema(req resource.Request) resource.Response {
	return resource.Response{
		Code:    -1,
		Message: "not implemented",
	}
}
