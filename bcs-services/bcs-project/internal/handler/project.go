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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/perm"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
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
	// 判断是否有创建权限
	if err := perm.CanCreateProject(auth.GetUserFromCtx(ctx)); err != nil {
		return err
	}
	// 创建项目
	ca := project.NewCreateAction(p.model)
	projectInfo, e := ca.Do(ctx, req)
	if e != nil {
		return e
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

// GetProject get project info
func (p *ProjectHandler) GetProject(ctx context.Context, req *proto.GetProjectRequest, resp *proto.ProjectResponse) error {
	// 查询项目信息
	ga := project.NewGetAction(p.model)
	projectInfo, err := ga.Do(ctx, req)
	if err != nil {
		return err
	}
	// 校验项目的查看权限
	if err := perm.CanViewProject(auth.GetUserFromCtx(ctx), projectInfo.ProjectID); err != nil {
		return err
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

// DeleteProject delete a project record
func (p *ProjectHandler) DeleteProject(ctx context.Context, req *proto.DeleteProjectRequest, resp *proto.ProjectResponse) error {
	// 校验项目的删除权限
	if err := perm.CanDeleteProject(auth.GetUserFromCtx(ctx), req.ProjectID); err != nil {
		return err
	}
	// 删除项目
	da := project.NewDeleteAction(p.model)
	if err := da.Do(ctx, req); err != nil {
		return err
	}
	// 处理返回数据及权限
	setResp(resp, nil)
	return nil
}

func (p *ProjectHandler) UpdateProject(ctx context.Context, req *proto.UpdateProjectRequest, resp *proto.ProjectResponse) error {
	// 校验项目的删除权限
	if err := perm.CanEditProject(auth.GetUserFromCtx(ctx), req.ProjectID); err != nil {
		return err
	}

	ua := project.NewUpdateAction(p.model)
	projectInfo, e := ua.Do(ctx, req)
	if e != nil {
		return e
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

func (p *ProjectHandler) ListProjects(ctx context.Context, req *proto.ListProjectsRequest, resp *proto.ListProjectsResponse) error {
	la := project.NewListAction(p.model)
	projects, e := la.Do(ctx, req)
	if e != nil {
		return e
	}
	// 获取权限
	permClient, err := perm.NewPermClient()
	if err != nil {
		return errorx.NewIAMClientErr(err)
	}
	// 获取 project id, 用以获取对应的权限
	ids := getProjectIDs(projects)
	perms, err := permClient.GetMultiProjectMultiActionPermission(
		auth.GetUserFromCtx(ctx), ids,
		[]string{perm.ProjectCreate, perm.ProjectView, perm.ProjectEdit, perm.ProjectDelete},
	)
	if err != nil {
		return err
	}
	// 处理返回
	setListPermsResp(resp, projects, perms)
	return nil
}

func (p *ProjectHandler) ListAuthorizedProjects(ctx context.Context, req *proto.ListAuthorizedProjReq, resp *proto.ListAuthorizedProjResp) error {
	lap := project.NewListAuthorizedProj(p.model)
	projects, e := lap.Do(ctx, req)
	if e != nil {
		return e
	}
	setListResp(resp, projects)
	return nil
}

// getProjectIDs 获取项目ID
func getProjectIDs(p *map[string]interface{}) []string {
	var ids []string
	results := (*p)["results"]
	if val, ok := results.([]*pm.Project); ok {
		for _, i := range val {
			ids = append(ids, i.ProjectID)
		}
	}
	return ids
}
