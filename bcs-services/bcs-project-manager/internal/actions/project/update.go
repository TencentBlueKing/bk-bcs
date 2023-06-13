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
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
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
		return nil, errorx.NewReadableErr(errorx.ParamErr, err.Error())
	}
	// 获取要更新的项目信息
	p, err := ua.model.GetProject(ua.ctx, req.ProjectID)
	if err != nil {
		logging.Error("project: %s not found", req.ProjectID)
		return nil, errorx.NewParamErr(err.Error())
	}
	oldProject := *p
	if err := ua.updateProject(p); err != nil {
		return nil, errorx.NewDBErr(err.Error())
	}
	if req.GetBusinessID() != "" && oldProject.BusinessID == "" || oldProject.BusinessID == "0" {
		// 开启容器服务
		// 1. 在监控创建对应的容器项目空间
		if err := bkmonitor.CreateSpace(p); err != nil {
			logging.Error("[ALARM-BK-MONITOR] create space for %s/%s in bkmonitor failed, err: %s",
				p.ProjectID, p.ProjectCode, err.Error())
		}
	}

	// 更新 bcs cc 中的数据
	go func() {
		if err := bcscc.UpdateProject(p); err != nil {
			logging.Error("[ALARM-CC-PROJECT] update project %s/%s in paas-cc failed, err: %s",
				p.ProjectID, p.ProjectCode, err.Error())
		}
	}()

	return p, nil
}

func (ua *UpdateAction) validate() error {
	// 当业务ID不为空时，校验当前用户是否为要绑定业务的业务运维
	if ua.req.BusinessID != "" {
		authUser, ok := ua.ctx.Value(middleware.AuthUserKey).(middleware.AuthUser)
		if !ok || authUser.Username == "" {
			return errorx.NewAuthErr("invalid user")
		}
		if _, err := cmdb.IsMaintainer(authUser.Username, ua.req.BusinessID); err != nil {
			return err
		}
	}
	name := ua.req.Name
	if name == "" {
		return nil
	}
	if len(strings.TrimSpace(name)) == 0 {
		return fmt.Errorf("name cannot contains only spaces")
	}
	// check name unique
	if p, _ := ua.model.GetProjectByField(ua.ctx, &pm.ProjectField{Name: name}); p != nil {
		// 如果是同一个项目，忽略名称校验
		if p.ProjectID == ua.req.ProjectID {
			return nil
		}
		return fmt.Errorf("name: %s is already exists", name)
	}
	return nil
}

func (ua *UpdateAction) updateProject(p *pm.Project) error {
	p.UpdateTime = time.Now().Format(time.RFC3339)
	// 从 context 中获取 username
	if authUser, err := middleware.GetUserFromContext(ua.ctx); err == nil {
		p.Updater = authUser.GetUsername()
		// 更新管理员，添加项目更新者，并且去重
		var managers string
		if ua.req.GetManagers() != "" {
			managers = stringx.JoinString(ua.req.GetManagers(), authUser.GetUsername())
		} else {
			managers = stringx.JoinString(p.Managers, authUser.GetUsername())
		}
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
	// 更新bool型，判断是否为nil
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
	if ua.req.Creator != "" {
		p.Creator = req.Creator
	}
	return ua.model.UpdateProject(ua.ctx, p)
}
