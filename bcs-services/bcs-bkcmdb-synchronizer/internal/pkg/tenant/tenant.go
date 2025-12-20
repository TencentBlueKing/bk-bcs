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

// Package tenant tenant related functions
package tenant

import (
	"context"
	"fmt"

	cmp "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/clustermanager"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/client/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/constants"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/internal/pkg/option"
)

// ResourceMetaData xxx
type ResourceMetaData struct {
	ProjectId string
	ClusterId string
	TenantId  string
}

// WithTenantIdByResourceForContext set tenantID by resource to context
func WithTenantIdByResourceForContext(ctx context.Context, resource ResourceMetaData) (context.Context, error) {
	if !option.GetGlobalConfig().Synchronizer.EnableMultiTenantMode {
		return context.WithValue(ctx, constants.BkTenantIdHeaderKey, constants.DefaultTenantId), nil
	}

	// 优先使用resource中的租户ID
	if resource.TenantId != "" {
		return context.WithValue(ctx, constants.BkTenantIdHeaderKey, resource.TenantId), nil
	}

	var (
		projectID = resource.ProjectId
	)

	if projectID == "" && resource.ClusterId != "" {
		clusterCli, err := clustermanager.GetClusterManagerGrpcGwClient()
		if err != nil {
			return ctx, err
		}
		resp, err := clusterCli.Cli.GetCluster(clusterCli.Ctx, &cmp.GetClusterReq{ClusterID: resource.ClusterId})
		if err != nil {
			return ctx, err
		}
		projectID = resp.Data.GetProjectID()
	}

	if projectID == "" {
		return ctx, fmt.Errorf("projectID is empty")
	}

	pro, err := projectmanager.GetProjectInfo(context.TODO(), projectID, true)
	if err != nil {
		return nil, err
	}

	// 注入租户信息
	return context.WithValue(ctx, constants.BkTenantIdHeaderKey, pro.TenantID), nil
}

// WithTenantIdFromContext set tenantID to context
func WithTenantIdFromContext(ctx context.Context, tenantId string) context.Context {
	return context.WithValue(ctx, constants.BkTenantIdHeaderKey, tenantId)
}

// GetTenantIdFromContext get tenantId from context
func GetTenantIdFromContext(ctx context.Context) string {
	tenantId := ""

	if id, ok := ctx.Value(constants.BkTenantIdHeaderKey).(string); ok {
		tenantId = id
	}

	if tenantId == "" {
		tenantId = string(constants.DefaultTenantId)
	}

	return tenantId
}
