/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
    10| * limitations under the License.
*/

// Package resources 实现 IAM V4 Provider 的业务查数
package resources

import (
	"context"

	iamv4 "github.com/Tencent/bk-bcs/bcs-services/bcs-user-manager/app/user-manager/v1http/iamv4"
)

// NewQuerier 默认查数实现
func NewQuerier() iamv4.Querier {
	return querier{}
}

type querier struct{}

func (querier) List(ctx context.Context, tenantID, resType string, filter iamv4.Filter, page iamv4.Page) (int, []iamv4.Instance, error) {
	switch resType {
	case iamv4.Project:
		return listProjects(ctx, tenantID, filter, page)
	case iamv4.Cluster:
		return listClusters(ctx, filter, page)
	case iamv4.Namespace:
		return listNamespaces(ctx, filter, page)
	case iamv4.CloudAccount:
		return listCloudAccounts(ctx, filter, page)
	case iamv4.TemplateSet:
		return 0, []iamv4.Instance{}, nil
	default:
		return 0, nil, iamv4.InvalidRequest("unsupported type")
	}
}

func (querier) Fetch(ctx context.Context, resType string, filter iamv4.Filter) ([]iamv4.Instance, error) {
	switch resType {
	case iamv4.Project:
		return fetchProjects(ctx, filter)
	case iamv4.Cluster:
		return fetchClusters(filter)
	case iamv4.Namespace:
		return fetchNamespaces(ctx, filter)
	case iamv4.CloudAccount:
		return fetchCloudAccounts(ctx, filter)
	case iamv4.TemplateSet:
		return []iamv4.Instance{}, nil
	default:
		return nil, iamv4.InvalidRequest("unsupported type")
	}
}
