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

package iam

import (
	"context"
	"errors"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"
	"github.com/TencentBlueKing/iam-go-sdk/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
	blog "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/log"
)

// NamespaceProvider is an namespace provider
type NamespaceProvider struct {
}

func init() {
	dispatcher.RegisterProvider("namespace", NamespaceProvider{})
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
	result, err := component.GetClusterNamespaces(context.Background(), projectID, clusterID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		ins := Instance{utils.CalcIAMNsID(clusterID, r.Name), r.Name, nil}
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
	ctx := context.Background()

	// get namespaces
	results := make([]interface{}, 0)
	for _, v := range filter.IDs {
		clusterID, err := parseNSID(v)
		if err != nil {
			return resource.Response{
				Code:    NotFoundCode,
				Message: err.Error(),
				Data:    results,
			}
		}
		ns, err := component.GetCachedNamespace(ctx, clusterID, v)
		if err != nil {
			blog.Log(ctx).Errorf("get namespace %s failed, err %s", v, err.Error())
			continue
		}

		results = append(results, Instance{v, ns.Name, nil})
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
	result, err := component.GetClusterNamespaces(context.Background(), projectID, clusterID)
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
		ins := Instance{utils.CalcIAMNsID(clusterID, r.Name), r.Name, nil}
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
