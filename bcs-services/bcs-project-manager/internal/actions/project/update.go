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

package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/iam"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// UpdateAction xxx
type UpdateAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateProjectRequest
}

// NewUpdateAction new update project action
func NewUpdateAction(model store.ProjectModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

// Do update project request
func (ua *UpdateAction) Do(ctx context.Context, req *proto.UpdateProjectRequest) (*pm.Project, error) {
	ua.ctx = ctx
	ua.req = req

	if err := ua.validate(); err != nil {
		return nil, errorx.NewParamErr(err)
	}
	// ??????????????????????????????
	p, err := ua.model.GetProject(ua.ctx, req.ProjectID)
	if err != nil {
		logging.Error("project: %s not found", req.ProjectID)
		return nil, errorx.NewParamErr(err)
	}
	oldProject := *p
	if err := ua.updateProject(p); err != nil {
		return nil, errorx.NewDBErr(err)
	}
	if req.GetBusinessID() != "" && oldProject.BusinessID == "" || oldProject.BusinessID == "0" {
		// ?????????????????????????????????????????????????????????
		matainers, err := cmdb.GetBusinessMaintainers(req.GetBusinessID())
		if err != nil {
			logging.Error("get business %s maintainers failed, err: %s", req.GetBusinessID(), err.Error())
		}
		if err := iam.CreateProjectPermManager(p.ProjectID, p.Name, matainers); err != nil {
			logging.Error("create project %s perm manager failed, err: %s", p.ProjectCode, err.Error())
		}
	}

	// ?????? bcs cc ????????????
	go func() {
		if err := bcscc.UpdateProject(p); err != nil {
			logging.Error("[ALARM-CC-PROJECT] update project %s/%s in paas-cc failed, err: %s",
				p.ProjectID, p.ProjectCode, err.Error())
		}
	}()

	return p, nil
}

func (ua *UpdateAction) validate() error {
	// ?????????ID????????????????????????????????????????????????????????????????????????
	if ua.req.BusinessID != "" {
		authUser, ok := ua.ctx.Value(middleware.AuthUserKey).(middleware.AuthUser)
		if !ok || authUser.Username == "" {
			return errorx.NewAuthErr()
		}
		if _, err := cmdb.IsMaintainer(authUser.Username, ua.req.BusinessID); err != nil {
			return err
		}
	}
	// check name unique
	name := ua.req.Name
	if name == "" {
		return nil
	}
	if p, _ := ua.model.GetProjectByField(ua.ctx, &pm.ProjectField{Name: name}); p != nil {
		// ?????????????????????????????????????????????
		if p.ProjectID == ua.req.ProjectID {
			return nil
		}
		return fmt.Errorf("name: %s is already exists", name)
	}
	return nil
}

func (ua *UpdateAction) updateProject(p *pm.Project) error {
	p.UpdateTime = time.Now().Format(time.RFC3339)
	// ??? context ????????? username
	if authUser, err := middleware.GetUserFromContext(ua.ctx); err == nil {
		p.Updater = authUser.GetUsername()
		// ??????????????????????????????????????????????????????
		managers := stringx.JoinString(p.Managers, authUser.GetUsername())
		managerList := stringx.RemoveDuplicateValues(stringx.SplitString(managers))
		p.Managers = strings.Join(managerList, ",")
	}

	req := ua.req

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.BusinessID != "" {
		p.BusinessID = req.BusinessID
	}
	if req.ProjectType != p.ProjectType {
		p.ProjectType = req.ProjectType
	}
	// ??????bool?????????????????????nil
	if req.UseBKRes != nil && req.UseBKRes.GetValue() != p.UseBKRes {
		p.UseBKRes = req.UseBKRes.GetValue()
	}
	if req.Description != "" {
		p.Description = req.Description
	}
	if req.Kind != "" {
		p.Kind = req.Kind
	}
	if req.IsOffline != nil && req.IsOffline.GetValue() != p.IsOffline {
		p.IsOffline = req.IsOffline.GetValue()
	}
	if req.DeployType > 0 {
		p.DeployType = ua.req.DeployType
	}
	if req.IsSecret != nil && req.IsSecret.GetValue() != p.IsSecret {
		p.IsSecret = req.IsSecret.GetValue()
	}
	if ua.req.BGID != "" {
		p.BGID = req.BGID
	}
	if ua.req.BGName != "" {
		p.BGName = req.BGName
	}
	if ua.req.DeptID != "" {
		p.DeptID = req.DeptID
	}
	if ua.req.DeptName != "" {
		p.DeptName = req.DeptName
	}
	if ua.req.CenterID != "" {
		p.CenterID = req.CenterID
	}
	if ua.req.CenterName != "" {
		p.CenterName = req.CenterName
	}
	return ua.model.UpdateProject(ua.ctx, p)
}
