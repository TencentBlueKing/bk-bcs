/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * 	http://opensource.org/licenses/MIT
 *
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package handler

import (
	"context"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/actions/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

type ProjectHandler struct {
	model store.ProjectModel
}

// NewProject return a project service hander
func NewProject(model store.ProjectModel) *ProjectHandler {
	return &ProjectHandler{
		model: model,
	}
}

// CreateProject implement for CreateProject interface
func (p *ProjectHandler) CreateProject(ctx context.Context, req *proto.CreateProjectRequest, resp *proto.ProjectResponse) error {
	ca := project.NewCreateAction(p.model)
	projectInfo, err := ca.Do(ctx, req)
	setResp(resp, err, projectInfo)
	return nil
}

// GetProject get project info
func (p *ProjectHandler) GetProject(ctx context.Context, req *proto.GetProjectRequest, resp *proto.ProjectResponse) error {
	ga := project.NewGetAction(p.model)
	projectInfo, err := ga.Do(ctx, req)
	setResp(resp, err, projectInfo)
	return nil
}

// DeleteProject delete a project record
func (p *ProjectHandler) DeleteProject(ctx context.Context, req *proto.DeleteProjectRequest, resp *proto.ProjectResponse) error {
	da := project.NewDeleteAction(p.model)
	err := da.Do(ctx, req)
	setResp(resp, err, nil)
	return nil
}

func (p *ProjectHandler) UpdateProject(ctx context.Context, req *proto.UpdateProjectRequest, resp *proto.ProjectResponse) error {
	ua := project.NewUpdateAction(p.model)
	projectInfo, err := ua.Do(ctx, req)
	setResp(resp, err, projectInfo)
	return nil
}

func (p *ProjectHandler) ListProjects(ctx context.Context, req *proto.ListProjectsRequest, resp *proto.ListProjectsResponse) error {
	la := project.NewListAction(p.model)
	projects, err := la.Do(ctx, req)
	setListResp(resp, err, projects)
	return nil
}
