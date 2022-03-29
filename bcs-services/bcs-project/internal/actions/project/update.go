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
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
)

// UpdateAction
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

// Handle update project request
func (ua *UpdateAction) Handle(ctx context.Context, req *proto.UpdateProjectRequest, resp *proto.ProjectResponse) {
	if req == nil || resp == nil {
		return
	}
	ua.ctx = ctx
	ua.req = req

	if err := ua.validate(); err != nil {
		setResp(resp, common.BcsProjectParamErr, common.BcsProjectParamErrMsg, err.Error(), nil)
		return
	}

	// 获取要更新的项目信息
	p, err := ua.model.GetProject(ua.ctx, req.ProjectID)
	if err != nil {
		setResp(resp, common.BcsProjectParamErr, common.BcsProjectParamErrMsg, err.Error(), nil)
		logging.Error("project: %s not found", req.ProjectID)
		return
	}
	if err := ua.updateProject(p); err != nil {
		setResp(resp, common.BcsProjectDbErr, common.BcsProjectDbErrMsg, err.Error(), nil)
		return
	}

	setResp(resp, common.BcsProjectSuccess, common.BcsProjectSuccessMsg, "", p)
	return
}

func (ua *UpdateAction) validate() error {
	// check name unique
	name := ua.req.Name
	if name == "" {
		return nil
	}
	if p, _ := ua.model.GetProjectByField(ua.ctx, &pm.ProjectField{Name: name}); p != nil {
		// 如果是同一个项目，忽略名称校验
		if p.ProjectID == ua.req.ProjectID {
			return nil
		}
		return fmt.Errorf("name: %s is already exists", name)
	}
	return nil
}

func (ua *UpdateAction) updateProject(p *proto.Project) error {
	timeStr := time.Now().Format(time.RFC3339)
	// 更新时间
	p.UpdateTime = timeStr
	p.Updater = ua.req.Updater
	p.Manager = util.JoinString(p.Manager, ua.req.Updater)

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
	if ua.req.BgID != "" {
		p.BgID = req.BgID
	}
	if ua.req.BgName != "" {
		p.BgName = req.BgName
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
