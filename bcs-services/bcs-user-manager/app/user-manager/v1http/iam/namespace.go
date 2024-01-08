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
	"errors"
	"strings"
	"sync"

	authUtils "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/TencentBlueKing/iam-go-sdk/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/utils"
)

// NamespaceProvider is an namespace provider
type NamespaceProvider struct {
}

func init() {
	dispatcher.RegisterProvider(Namespace, NamespaceProvider{})
}

// ListAttr implements the list_attr
func (p NamespaceProvider) ListAttr(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: []interface{}{},
	}
}

// ListAttrValue implements the list_attr_value
func (p NamespaceProvider) ListAttrValue(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// ListInstance implements the list_instance
func (p NamespaceProvider) ListInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.Ancestors) != 2 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	projectID := filter.Ancestors[0].ID
	clusterID := filter.Ancestors[1].ID
	// get project code from project id
	project, err := component.GetProjectWithCache(req.Context, projectID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}

	result, err := component.GetClusterNamespaces(req.Context, project.ProjectCode, clusterID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		ins := Instance{authUtils.CalcIAMNsID(clusterID, r.Name), r.Name, nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}

// FetchInstanceInfo implements the fetch_instance_info
func (p NamespaceProvider) FetchInstanceInfo(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.IDs) == 0 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "ids is empty",
			Data:    []interface{}{},
		}
	}

	// get namespaces
	nsChan := make(chan Instance, len(filter.IDs))
	results := make([]interface{}, 0)
	wg := sync.WaitGroup{}
	for _, v := range filter.IDs {
		wg.Add(1)
		go func(nsID string) {
			defer wg.Done()
			ctx := context.Background()
			ctx = context.WithValue(ctx, utils.ContextValueKeyRequestID, utils.GetRequestIDFromContext(req.Context))
			clusterID, err := parseNSID(nsID)
			if err != nil {
				blog.Log(ctx).Errorf("get namespace %s in cluster %s failed, err %s", nsID, clusterID, err.Error())
				return
			}
			ns, err := component.GetCachedNamespace(ctx, clusterID, nsID)
			if err != nil {
				blog.Log(ctx).Errorf("get namespace %s in cluster %s failed, err %s", nsID, clusterID, err.Error())
				return
			}
			nsChan <- Instance{nsID, ns.Name, ns.Managers}
		}(v)
	}
	wg.Wait()
	close(nsChan)

	for n := range nsChan {
		results = append(results, n)
	}
	return resource.Response{
		Code: 0,
		Data: results,
	}
}

// ListInstanceByPolicy implements the list_instance_by_policy
func (p NamespaceProvider) ListInstanceByPolicy(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// SearchInstance implements the search_instance
func (p NamespaceProvider) SearchInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.Ancestors) != 2 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	projectID := filter.Ancestors[0].ID
	clusterID := filter.Ancestors[1].ID
	// get project code from project id
	project, err := component.GetProjectWithCache(req.Context, projectID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}

	result, err := component.GetClusterNamespaces(req.Context, project.ProjectCode, clusterID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		if filter.Keyword != "" && !strings.Contains(r.Name, filter.Keyword) {
			continue
		}
		ins := Instance{authUtils.CalcIAMNsID(clusterID, r.Name), r.Name, nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}

func parseNSID(nsID string) (string, error) {
	s := strings.Split(nsID, ":")
	if len(s) != 2 {
		return "", errors.New("invalid ns id")
	}
	return "BCS-K8S-" + s[0], nil
}
