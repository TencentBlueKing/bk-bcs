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

package project

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bkmonitor"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/cmdb"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/stringx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/tenant"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
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
	if req.GetBusinessID() != "" && (oldProject.BusinessID == "" || oldProject.BusinessID == "0") {
		// 开启容器服务
		// 1. 在监控创建对应的容器项目空间
		if err := bkmonitor.CreateSpace(ctx, p); err != nil {
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
		if _, err := cmdb.IsMaintainer(ua.ctx, authUser.Username, ua.req.BusinessID); err != nil {
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
	if authUser, err := tenant.GetAuthUserInfoFromCtx(ua.ctx); err == nil {
		p.Updater = authUser.GetUsername()
		// 更新管理员，并且去重
		if ua.req.GetManagers() != "" {
			p.Managers = stringx.JoinString(ua.req.GetManagers())
		}
		if ua.req.GetBusinessID() != "" && (p.BusinessID == "" || p.BusinessID == "0") {
			// 更新请求为开启容器服务，增加当前用户为项目管理员
			p.Managers = stringx.JoinString(p.Managers, authUser.GetUsername())
		}
		p.Managers = strings.Join(stringx.RemoveDuplicateValues(stringx.SplitString(p.Managers)), ",")
	}

	req := ua.req

	if req.Name != "" {
		p.Name = req.Name
	}
	if req.BusinessID != "" {
		p.BusinessID = req.BusinessID
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
	if ua.req.Creator != "" {
		p.Creator = req.Creator
	}
	if ua.req.Labels != nil {
		p.Labels = req.Labels
	}
	if ua.req.Annotations != nil {
		p.Annotations = req.Annotations
	}

	return ua.model.UpdateProject(ua.ctx, p)
}
