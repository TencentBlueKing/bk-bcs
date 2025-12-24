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

// Package project xxx
package project

import (
	"context"
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
)

// CreateAction action for create project
type CreateAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.CreateProjectRequest
}

// NewCreateAction new create project action
func NewCreateAction(model store.ProjectModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

// Do create project request
func (ca *CreateAction) Do(ctx context.Context, req *proto.CreateProjectRequest) (*pm.Project, error) {
	ca.ctx = ctx
	ca.req = req

	if err := ca.validate(); err != nil {
		return nil, errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}

	// 如果有传递项目ID，则以传递的为准，否则动态生成32位的字符串作为项目ID
	if req.ProjectID == "" {
		ca.req.ProjectID = stringx.GenUUID()
	}

	if err := ca.createProject(); err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}

	p, err := ca.model.GetProject(ca.ctx, ca.req.ProjectID)
	if err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}
	// 向 bcs cc 写入数据
	go func() {
		if err := bcscc.CreateProject(p); err != nil {
			logging.Error("[ALARM-CC-PROJECT] create project %s/%s in paas-cc failed, err: %s",
				p.ProjectID, p.ProjectCode, err.Error())
		}

	}()
	// 返回项目信息
	return p, nil
}

func (ca *CreateAction) createProject() error {
	tenantId := tenant.GetTenantIdFromContext(ca.ctx)

	p := &pm.Project{
		ProjectID:         ca.req.ProjectID,
		Name:              ca.req.Name,
		TenantID:          tenantId,
		TenantProjectCode: ca.req.ProjectCode,
		ProjectType:       ca.req.ProjectType,
		UseBKRes:          ca.req.UseBKRes,
		Description:       ca.req.Description,
		IsOffline:         ca.req.IsOffline,
		Kind:              ca.req.Kind,
		BusinessID:        ca.req.BusinessID,
		DeployType:        ca.req.DeployType,
		BGID:              ca.req.BGID,
		BGName:            ca.req.BGName,
		DeptID:            ca.req.DeptID,
		DeptName:          ca.req.DeptName,
		CenterID:          ca.req.CenterID,
		CenterName:        ca.req.CenterName,
		IsSecret:          ca.req.IsSecret,
		Labels:            ca.req.Labels,
		Annotations:       ca.req.Annotations,
	}
	// 从 context 中获取 username
	if authUser, err := middleware.GetUserFromContext(ca.ctx); err == nil {
		p.Creator = authUser.GetUsername()
		p.Managers = authUser.GetUsername()
	}

	if tenant.IsMultiTenantEnabled() {
		p.ProjectCode = ca.generateProjectCode(p.TenantID, p.TenantProjectCode)
	} else {
		p.ProjectCode = ca.req.ProjectCode
	}

	return ca.model.CreateProject(ca.ctx, p)
}

func (ca *CreateAction) checkProjectValidate(projectId, projectCode, name string) error {
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("name cannot contains only spaces")
	}

	if p, _ := ca.model.GetProjectByField(ca.ctx, &pm.ProjectField{ProjectID: projectId, ProjectCode: projectCode,
		Name: name}); p != nil {
		if p.ProjectID == projectId {
			return fmt.Errorf("projectID: %s is already exists", projectId)
		}
		if p.ProjectCode == projectCode {
			return fmt.Errorf("projectCode: %s is already exists", projectCode)
		}
		if p.Name == name {
			return fmt.Errorf("name: %s is already exists", name)
		}
	}

	return nil
}

// validate validate create project request
// projectCode 创建项目时 BCS接收的项目EnglishName 字段；后台自动识别
// 单租户环境 tenant_project_code = projectCode = ca.req.projectCode 且 租户ID 默认是default
// 多租户环境 tenant_project_code = ca.req.projectCode 且 projectCode = tenantID-tenantProjectCode
func (ca *CreateAction) validate() error {
	// 多租户环境下，projectId、projectCode、name 需要全局唯一 且 tenantProjectCode 租户下唯一
	if tenant.IsMultiTenantEnabled() {
		tenantId := tenant.GetTenantIdFromContext(ca.ctx)
		projectID, projectCode, name := ca.req.ProjectID,
			ca.generateProjectCode(tenantId, ca.req.ProjectCode), ca.req.Name

		err := ca.checkProjectValidate(projectID, projectCode, name)
		if err != nil {
			return err
		}

		// 校验租户下 tenantProjectCode是否唯一
		if tenantId != "" && ca.req.ProjectCode != "" {
			p, _ := ca.model.GetProjectByField(ca.ctx, &pm.ProjectField{TenantID: tenantId,
				TenantProjectCode: ca.req.ProjectCode})
			if p != nil {
				return fmt.Errorf("tenant %s tenantProjectCode: %s is already exists",
					tenantId, ca.req.ProjectCode)
			}
		}

		return nil
	}

	// check projectID、projectCode、name
	projectID, projectCode, name := ca.req.ProjectID, ca.req.ProjectCode, ca.req.Name
	err := ca.checkProjectValidate(projectID, projectCode, name)
	if err != nil {
		return err
	}

	return nil
}

func (ca *CreateAction) generateProjectCode(tenantID, tenantProjectCode string) string {
	return fmt.Sprintf("%s-%s", tenantID, tenantProjectCode)
}
