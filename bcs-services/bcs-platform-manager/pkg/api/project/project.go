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

// Package project project operate
package project

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	actions "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/actions/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ListProject 获取项目列表
// @Summary 获取项目列表
// @Tags    Logs
// @Produce json
// @Success 200 {array} types.ListProjectsResp
// @Router  /project [get]
func ListProject(ctx context.Context, req *types.ListProjectsReq) (*[]*types.ListProjectsResp, error) {
	result, err := actions.NewProjectAction().ListProject(ctx, &bcsproject.ListProjectsRequest{
		ProjectIDs:  req.ProjectIDs,
		Names:       req.Names,
		ProjectCode: req.ProjectCode,
		SearchName:  req.SearchName,
		Kind:        req.Kind,
		Offset:      req.Offset,
		Limit:       req.Limit,
		All:         req.All,
		BusinessID:  req.BusinessID,
	})
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// GetProject 获取项目详情
// @Summary 获取项目详情
// @Tags    Logs
// @Produce json
// @Success 200 {struct} bcsproject.Project
// @Router  /project/{projectIDOrCode} [get]
func GetProject(ctx context.Context, req *types.GetProjectsReq) (*bcsproject.Project, error) {
	result, err := actions.NewProjectAction().GetProject(ctx, req.ProjectIDOrCode)
	if err != nil {
		return nil, err
	}

	return result, nil
}

// UpdateProjectManagers 更新项目managers
// @Summary 更新项目managers
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /project/{projectID}/managers [put]
func UpdateProjectManagers(ctx context.Context, req *types.UpdateProjectManagersReq) (*bool, error) {
	result, err := actions.NewProjectAction().UpdateProjectManagers(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// UpdateProjectBusiness 更新项目business
// @Summary 更新项目business
// @Tags    Logs
// @Produce json
// @Success 200 {bool} bool
// @Router  /project/{projectID}/business [put]
func UpdateProjectBusiness(ctx context.Context, req *types.UpdateProjectBusinessReq) (*bool, error) {
	result, err := actions.NewProjectAction().UpdateProjectBusiness(ctx, req)
	if err != nil {
		return nil, err
	}

	return &result, nil
}
