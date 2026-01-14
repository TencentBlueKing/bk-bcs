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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/bcsproject"
	"github.com/golang/protobuf/ptypes/wrappers"

	projectrmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
)

// ProjectAction project action interface
type ProjectAction interface { // nolint
	ListProject(ctx context.Context, req *bcsproject.ListProjectsRequest) (*types.ListProjectsResp, error)
	GetProject(ctx context.Context, projectIDOrCode string) (*types.GetProjectsResp, error)
	UpdateProject(ctx context.Context, req *types.UpdateProjectReq) (bool, error)
	UpdateProjectManagers(ctx context.Context, req *types.UpdateProjectManagersReq) (bool, error)
	UpdateProjectBusiness(ctx context.Context, req *types.UpdateProjectBusinessReq) (bool, error)
	UpdateProjectIsOffline(ctx context.Context, req *types.UpdateProjectIsOfflineReq) (bool, error)
}

// Action action for project
type Action struct{}

// NewProjectAction new project action
func NewProjectAction() ProjectAction {
	return &Action{}
}

// ListProject list project
func (a *Action) ListProject(ctx context.Context, req *bcsproject.ListProjectsRequest) (
	*types.ListProjectsResp, error) {
	businesses, err := cmdb.GetCmdbClient().GetBusiness()
	if err != nil {
		return nil, utils.SystemError(err)
	}

	data, err := projectrmgr.ListProject(ctx, req)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	result := types.ListProjectsResp{
		Total:   data.Total,
		Results: make([]*types.ListProjectsData, 0),
	}
	for _, project := range data.Results {
		result.Results = append(result.Results, &types.ListProjectsData{
			CreateTime:  project.CreateTime,
			Creator:     project.Creator,
			ProjectID:   project.ProjectID,
			Name:        project.Name,
			ProjectCode: project.ProjectCode,
			Description: project.Description,
			IsOffline:   project.IsOffline,
			BusinessID:  project.BusinessID,
			BusinessName: func() string {
				for _, business := range *businesses {
					if fmt.Sprint(business.BkBizID) == project.BusinessID {
						return business.BkBizName
					}
				}

				return ""
			}(),
			Managers: project.Managers,
			Link: fmt.Sprintf("%s/bcs/projects/%s/project-info",
				config.G.BCS.Server, project.ProjectCode),
		})
	}

	return &result, nil
}

// GetProject get project
func (a *Action) GetProject(ctx context.Context, projectIDOrCode string) (*types.GetProjectsResp, error) {
	project, err := projectrmgr.GetProject(ctx, projectIDOrCode)
	if err != nil {
		return nil, utils.SystemError(err)
	}

	return &types.GetProjectsResp{
		CreateTime:   project.CreateTime,
		UpdateTime:   project.UpdateTime,
		Creator:      project.Creator,
		Updater:      project.Updater,
		Managers:     project.Managers,
		ProjectID:    project.ProjectID,
		Name:         project.Name,
		ProjectCode:  project.ProjectCode,
		UseBKRes:     project.UseBKRes,
		Description:  project.Description,
		IsOffline:    project.IsOffline,
		Kind:         project.Kind,
		BusinessID:   project.BusinessID,
		BusinessName: project.BusinessName,
		Labels:       project.Labels,
		Annotations:  project.Annotations,
	}, nil
}

// UpdateProject update project
func (a *Action) UpdateProject(ctx context.Context, req *types.UpdateProjectReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID:   req.ProjectID,
		Managers:    req.Managers,
		BusinessID:  req.BusinessID,
		Name:        req.Name,
		ProjectCode: req.ProjectCode,
		UseBKRes:    &wrappers.BoolValue{Value: req.UseBKRes},
		IsOffline:   &wrappers.BoolValue{Value: req.IsOffline},
		Description: req.Description,
		Kind:        req.Kind,
		Labels:      req.Labels,
		Annotations: req.Annotations,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateProjectManagers update project managers
func (a *Action) UpdateProjectManagers(ctx context.Context, req *types.UpdateProjectManagersReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID: req.ProjectID,
		Managers:  req.Managers,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateProjectBusiness update project business
func (a *Action) UpdateProjectBusiness(ctx context.Context, req *types.UpdateProjectBusinessReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID:  req.ProjectID,
		BusinessID: req.BusinessID,
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}

// UpdateProjectIsOffline update project isoffline
func (a *Action) UpdateProjectIsOffline(ctx context.Context, req *types.UpdateProjectIsOfflineReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID: req.ProjectID,
		IsOffline: &wrappers.BoolValue{Value: req.IsOffline},
	})
	if err != nil {
		return false, utils.SystemError(err)
	}

	return result, nil
}
