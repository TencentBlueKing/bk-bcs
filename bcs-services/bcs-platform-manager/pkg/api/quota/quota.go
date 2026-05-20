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

// Package quota project quota operate
package quota

import (
	"context"

	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/quota"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// CreateProjectQuota 创建项目额度
// @Summary 创建项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota [post]
func CreateProjectQuota(ctx context.Context, req *types.CreateProjectQuotaReq) (*bool, error) {
	result, err := actions.NewQuotaAction().CreateProjectQuota(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProjectQuota 获取项目额度
// @Summary 获取项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {struct} types.ProjectQuota
// @Router  /quota/{quotaID} [get]
func GetProjectQuota(ctx context.Context, req *types.GetProjectQuotaReq) (*types.ProjectQuota, error) {
	result, err := actions.NewQuotaAction().GetProjectQuota(ctx, req.QuotaId)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateProjectQuota 更新项目额度
// @Summary 更新项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota/{quotaID} [put]
func UpdateProjectQuota(ctx context.Context, req *types.UpdateProjectQuotaReq) (*bool, error) {
	result, err := actions.NewQuotaAction().UpdateProjectQuota(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ScaleUpProjectQuota 扩容项目额度
// @Summary 扩容项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota/{quotaID}/scaleup [put]
func ScaleUpProjectQuota(ctx context.Context, req *types.ScaleUpProjectQuotaReq) (*bool, error) {
	result, err := actions.NewQuotaAction().ScaleUpProjectQuota(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ScaleDownProjectQuota 缩容项目额度
// @Summary 缩容项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota/{quotaID}/scaledown [put]
func ScaleDownProjectQuota(ctx context.Context, req *types.ScaleDownProjectQuotaReq) (*bool, error) {
	result, err := actions.NewQuotaAction().ScaleDownProjectQuota(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// DeleteProjectQuota 删除项目额度
// @Summary 删除项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota/{quotaID} [delete]
func DeleteProjectQuota(ctx context.Context, req *types.DeleteProjectQuotaReq) (*bool, error) {
	result, err := actions.NewQuotaAction().DeleteProjectQuota(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// ListProjectQuotasV2 列出项目额度
// @Summary 列出项目额度
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota [get]
func ListProjectQuotasV2(ctx context.Context, req *types.ListProjectQuotasV2Req) (*types.ListProjectQuotasData, error) {
	result, err := actions.NewQuotaAction().ListProjectQuotasV2(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetProjectQuotasStatistics 获取项目额度统计信息
// @Summary 获取项目额度统计信息
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /quota/statistics [get]
func GetProjectQuotasStatistics(ctx context.Context, req *types.GetProjectQuotasStatisticsReq) (
	*types.ProjectQuotasStatisticsData, error) {
	result, err := actions.NewQuotaAction().GetProjectQuotasStatistics(ctx, req)
	if err != nil {
		return nil, err
	}

	return result, nil
}
