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
	"strings"

	"github.com/TencentBlueKing/iam-go-sdk/resource"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/pkg/component"
)

// ClusterProvider is a cluster provider
type ClusterProvider struct {
}

func init() {
	dispatcher.RegisterProvider(Cluster, ClusterProvider{})
}

// ListAttr implements the list_attr
func (p ClusterProvider) ListAttr(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: []interface{}{},
	}
}

// ListAttrValue implements the list_attr_value
func (p ClusterProvider) ListAttrValue(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// ListInstance implements the list_instance
func (p ClusterProvider) ListInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if filter.Parent.ID == "" {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	result, err := component.GetClustersByProjectID(req.Context, filter.Parent.ID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range result {
		ins := Instance{r.ClusterID, combineNameID(r.ClusterName, r.ClusterID), nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}

// FetchInstanceInfo implements the fetch_instance_info
func (p ClusterProvider) FetchInstanceInfo(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if len(filter.IDs) == 0 {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "ids is empty",
			Data:    []interface{}{},
		}
	}

	clusterMap, err := component.GetClusterMap()
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}

	results := make([]interface{}, 0)
	for _, v := range filter.IDs {
		cls, ok := clusterMap[v]
		if !ok {
			continue
		}
		results = append(results, Instance{cls.ClusterID, combineNameID(cls.ClusterName, cls.ClusterID),
			[]string{cls.Creator, cls.Updater}})
	}
	return resource.Response{
		Code: 0,
		Data: results,
	}
}

// ListInstanceByPolicy implements the list_instance_by_policy
func (p ClusterProvider) ListInstanceByPolicy(req resource.Request) resource.Response {
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: 0, Results: []interface{}{}},
	}
}

// SearchInstance implements the search_instance
func (p ClusterProvider) SearchInstance(req resource.Request) resource.Response {
	filter := convertFilter(req.Filter)
	if filter.Parent.ID == "" {
		return resource.Response{
			Code:    NotFoundCode,
			Message: "parent id is empty",
		}
	}
	clusters, err := component.GetClustersByProjectID(req.Context, filter.Parent.ID)
	if err != nil {
		return resource.Response{
			Code:    SystemErrCode,
			Message: err.Error(),
		}
	}
	results := make([]interface{}, 0)
	for _, r := range clusters {
		// 模糊搜索集群 ID 和集群名称
		if filter.Keyword != "" && !strings.Contains(r.ClusterName, filter.Keyword) &&
			!strings.Contains(r.ClusterID, filter.Keyword) {
			continue
		}
		ins := Instance{r.ClusterID, combineNameID(r.ClusterName, r.ClusterID), nil}
		results = append(results, ins)
	}
	return resource.Response{
		Code: 0,
		Data: ListResult{Count: len(results), Results: results},
	}
}
