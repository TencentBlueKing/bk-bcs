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

	projectrmgr "github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/component/bcs/projectmanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/types"
)

// ProjectAction project action interface
type ProjectAction interface { // nolint
	ListProject(ctx context.Context, req *bcsproject.ListProjectsRequest) ([]*types.ListProjectsResp, error)
	GetProject(ctx context.Context, projectIDOrCode string) (*bcsproject.Project, error)
	UpdateProjectManagers(ctx context.Context, req *types.UpdateProjectManagersReq) (bool, error)
	UpdateProjectBusiness(ctx context.Context, req *types.UpdateProjectBusinessReq) (bool, error)
}

// Action action for project
type Action struct{}

// NewProjectAction new project action
func NewProjectAction() ProjectAction {
	return &Action{}
}

// ListProject list project
func (a *Action) ListProject(ctx context.Context, req *bcsproject.ListProjectsRequest) (
	[]*types.ListProjectsResp, error) {
	projects, err := projectrmgr.ListProject(ctx, req)
	if err != nil {
		return nil, err
	}

	result := make([]*types.ListProjectsResp, 0)
	for _, project := range projects {
		result = append(result, &types.ListProjectsResp{
			CreateTime:  project.CreateTime,
			Creator:     project.Creator,
			ProjectID:   project.ProjectID,
			Name:        project.Name,
			ProjectCode: project.ProjectCode,
			Description: project.Description,
			IsOffline:   project.IsOffline,
			BusinessID:  project.BusinessID,
			Link: fmt.Sprintf("%s/bcs/projects/%s/project-info",
				config.G.BCS.Host, project.ProjectCode),
		})
	}

	return result, nil
}

// GetProject get project
func (a *Action) GetProject(ctx context.Context, projectIDOrCode string) (*bcsproject.Project, error) {
	project, err := projectrmgr.GetProject(ctx, projectIDOrCode)
	if err != nil {
		return nil, err
	}

	return project, nil
}

// UpdateProjectManagers update project managers
func (a *Action) UpdateProjectManagers(ctx context.Context, req *types.UpdateProjectManagersReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID: req.ProjectID,
		Managers:  req.Managers,
	})

	return result, err
}

// UpdateProjectBusiness update project business
func (a *Action) UpdateProjectBusiness(ctx context.Context, req *types.UpdateProjectBusinessReq) (bool, error) {
	result, err := projectrmgr.UpdateProjectV2(ctx, &bcsproject.UpdateProjectV2Request{
		ProjectID:  req.ProjectID,
		BusinessID: req.BusinessID,
	})

	return result, err
}
