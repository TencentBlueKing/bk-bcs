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

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/utils"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/remote/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// ResourceMetaData xxx
type ResourceMetaData struct {
	ProjectId   string
	ClusterId   string
	NodeGroupId string
	TenantId    string
}

// WithTenantIdByResourceForContext set tenantID by resource to context
func WithTenantIdByResourceForContext(ctx context.Context, resource ResourceMetaData) (context.Context, error) {
	if !options.GetGlobalCMOptions().TenantConfig.EnableMultiTenantMode {
		// nolint:staticcheck
		return context.WithValue(ctx, common.BkTenantIdHeaderKey, common.BkDefaultTenantId), nil
	}

	// 优先使用resource中的租户ID
	if resource.TenantId != "" {
		// nolint:staticcheck
		return context.WithValue(ctx, common.BkTenantIdHeaderKey, resource.TenantId), nil
	}

	var (
		projectID = resource.ProjectId
	)

	if projectID == "" && resource.ClusterId != "" {
		cluster, err := store.GetStoreModel().GetCluster(context.TODO(), resource.ClusterId)
		if err != nil {
			return ctx, err
		}

		projectID = cluster.ProjectID
	}

	if projectID == "" && resource.NodeGroupId != "" {
		cluster, err := store.GetStoreModel().GetNodeGroup(context.TODO(), resource.NodeGroupId)
		if err != nil {
			return ctx, err
		}

		projectID = cluster.ProjectID
	}

	if projectID == "" {
		return ctx, fmt.Errorf("projectID is empty")
	}

	pro, err := project.GetProjectManagerClient().GetProjectInfo(context.TODO(), projectID, true)
	if err != nil {
		return ctx, err
	}

	// 注入租户信息
	// nolint:staticcheck
	return context.WithValue(ctx, common.BkTenantIdHeaderKey, pro.TenantID), nil
}

// WithTenantIdFromContext set tenantID to context
func WithTenantIdFromContext(ctx context.Context, tenantId string) context.Context {
	// nolint:staticcheck
	return context.WithValue(ctx, common.BkTenantIdHeaderKey, tenantId)
}

// GetTenantIdFromContext get tenantId from context
func GetTenantIdFromContext(ctx context.Context) string {
	tenantId := ""

	if id, ok := ctx.Value(common.BkTenantIdHeaderKey).(string); ok {
		tenantId = id
	}

	if tenantId == "" {
		tenantId = utils.DefaultTenantId
	}

	return tenantId
}
