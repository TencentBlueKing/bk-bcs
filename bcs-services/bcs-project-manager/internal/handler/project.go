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

package handler

import (
	"context"

	middleauth "github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/actions/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	projutil "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

var (
	// CacheKeyBusinessIDPrefix business prefix
	CacheKeyBusinessIDPrefix = "BUSINESS_%s"
)

// ProjectHandler xxx
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
func (p *ProjectHandler) CreateProject(ctx context.Context,
	req *proto.CreateProjectRequest, resp *proto.ProjectResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	// 创建项目
	ca := project.NewCreateAction(p.model)
	projectInfo, e := ca.Do(ctx, req)
	if e != nil {
		return e
	}
	if authUser.Username != "" {
		// 授权创建者项目编辑和查看权限
		if errLocal := iam.GrantProjectCreatorActions(ctx, authUser.Username,
			projectInfo.ProjectID, projectInfo.Name); errLocal != nil {
			logging.Error("grant project %s for creator %s permission failed, err: %s",
				projectInfo.ProjectID, authUser.Username, errLocal.Error())
		}
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

// GetProject get project info
func (p *ProjectHandler) GetProject(ctx context.Context,
	req *proto.GetProjectRequest, resp *proto.ProjectResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	user, err := middleauth.GetUserFromContext(ctx)
	if err != nil {
		return err
	}

	// 查询项目信息
	ga := project.NewGetAction(p.model)
	projectInfo, err := ga.Do(ctx, req)
	if err != nil {
		return err
	}
	businessName := ""
	if projectInfo.BusinessID != "" && projectInfo.BusinessID != "0" && !user.IsInner() {
		business, err := cmdb.GetBusinessByID(ctx, projectInfo.BusinessID, true)
		if err != nil {
			logging.Error("get business %s failed, err: %s", projectInfo.BusinessID, err.Error())
		} else {
			businessName = business.BKBizName
		}
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	resp.Data.BusinessName = businessName
	return nil
}

// DeleteProject delete a project record
func (p *ProjectHandler) DeleteProject(ctx context.Context,
	req *proto.DeleteProjectRequest, resp *proto.ProjectResponse) error {
	// // 删除项目
	// da := project.NewDeleteAction(p.model)
	// if err := da.Do(ctx, req); err != nil {
	// 	return err
	// }
	// // 处理返回数据及权限
	// setResp(resp, nil)
	return errorx.NewReadableErr(errorx.PermDeniedErr, "projects are not allowed to be deleted")
}

// UpdateProject update a project record
func (p *ProjectHandler) UpdateProject(ctx context.Context,
	req *proto.UpdateProjectRequest, resp *proto.ProjectResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	ua := project.NewUpdateAction(p.model)
	projectInfo, e := ua.Do(ctx, req)
	if e != nil {
		return e
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

// UpdateProjectV2 update a project business id
func (p *ProjectHandler) UpdateProjectV2(ctx context.Context,
	req *proto.UpdateProjectV2Request, resp *proto.ProjectResponse) error {
	ua := project.NewUpdateV2Action(p.model)
	projectInfo, e := ua.Do(ctx, req)
	if e != nil {
		return e
	}
	// 处理返回数据及权限
	setResp(resp, projectInfo)
	return nil
}

// ListProjects list projects reocrds
func (p *ProjectHandler) ListProjects(ctx context.Context,
	req *proto.ListProjectsRequest, resp *proto.ListProjectsResponse) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	logging.Info("list projects, req: %v", req)
	logging.Info("list projects, authUser: %+v", authUser)

	la := project.NewListAction(p.model)
	projects, e := la.Do(ctx, req)
	if e != nil {
		return e
	}

	projectIam, err := auth.GetProjectIamClient(authUser.GetResourceTenantId())
	if err != nil {
		return err
	}

	// TODO: 赋予所有权限，等待权限中心支持后，再
	if authUser.GetUsername() != "" {
		// with username
		// 获取 project id, 用以获取对应的权限
		ids := getProjectIDs(projects)
		perms, err := projectIam.GetMultiProjectMultiActionPerm(authUser.GetUsername(), ids,
			[]string{auth.ProjectCreate, auth.ProjectView, auth.ProjectEdit, auth.ProjectDelete})
		if err != nil {
			return err
		}

		// perms := getProjectAllPerms(ids)
		// 处理返回
		setListPermsResp(resp, projects, perms)
	} else {
		// without username
		setListPermsResp(resp, projects, nil)
	}
	projutil.PatchBusinessName(ctx, resp.Data.Results)
	return nil
}

// ListAuthorizedProjects query authorized project info list
func (p *ProjectHandler) ListAuthorizedProjects(ctx context.Context,
	req *proto.ListAuthorizedProjReq, resp *proto.ListAuthorizedProjResp) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	lap := project.NewListAuthorizedProj(p.model)
	projects, e := lap.Do(ctx, req)
	if e != nil {
		return e
	}
	if req.All {
		if authUser.Username != "" {
			ids := getProjectIDs(projects)
			projectIam, err := auth.GetProjectIamClient(authUser.GetResourceTenantId())
			if err != nil {
				return err
			}

			perms, err := projectIam.GetMultiProjectMultiActionPerm(authUser.GetUsername(), ids,
				[]string{auth.ProjectCreate, auth.ProjectView, auth.ProjectEdit, auth.ProjectDelete})
			if err != nil {
				return err
			}

			// perms := getProjectAllPerms(ids)
			// set web_annotation
			setListPermsResp(resp, projects, perms)
		}
	} else {
		// list only authorized projects, so no need to set web_annotation
		setListResp(resp, projects)
	}
	projutil.PatchBusinessName(ctx, resp.Data.Results)

	return nil
}

// ListProjectsForIAM list projects with k8s enabled for iam grant
func (p *ProjectHandler) ListProjectsForIAM(ctx context.Context,
	req *proto.ListProjectsForIAMReq, resp *proto.ListProjectsForIAMResp) error {

	authUser := tenant.GetAuthAndTenantInfoFromCtx(ctx)
	ctx = tenant.WithTenantIdFromContext(ctx, authUser.ResourceTenantId)

	lap := project.NewListForIAMActionAction(p.model)
	projects, e := lap.Do(ctx, req)
	if e != nil {
		return e
	}
	resp.Data = projects
	return nil
}

// getProjectIDs 获取项目ID
// NOCC:golint/noptr(设计如此:)
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

// GetProjectActive get projects active
func (p *ProjectHandler) GetProjectActive(ctx context.Context,
	req *proto.GetProjectActiveRequest, resp *proto.GetProjectActiveResponse) error {
	lap := project.NewGetAction(p.model)
	isActive, e := lap.Active(ctx, req)
	if e != nil {
		return e
	}
	resp.Data = &proto.ProjectActiveData{
		IsActive: isActive,
	}
	return nil
}
