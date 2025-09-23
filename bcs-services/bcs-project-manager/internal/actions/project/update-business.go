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
	"time"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/logging"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/errorx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/proto/bcsproject"
	"github.com/Tencent/bk-bcs/bcs-services/pkg/bcs-auth/middleware"
)

// UpdateBusinessAction xxx
type UpdateBusinessAction struct {
	ctx   context.Context
	model store.ProjectModel
	req   *proto.UpdateProjectBusinessRequest
}

// NewUpdateBusinessAction new update project business action
func NewUpdateBusinessAction(model store.ProjectModel) *UpdateBusinessAction {
	return &UpdateBusinessAction{
		model: model,
	}
}

// Do update project request
func (ua *UpdateBusinessAction) Do(ctx context.Context, req *proto.UpdateProjectBusinessRequest) (*pm.Project, error) {
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

	if err := ua.updateProjectBusiness(p); err != nil {
		return nil, errorx.NewDBErr(err.Error())
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

func (ua *UpdateBusinessAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	return nil
}

func (ua *UpdateBusinessAction) updateProjectBusiness(p *pm.Project) error {
	p.UpdateTime = time.Now().Format(time.RFC3339)
	// 从 context 中获取 username
	if authUser, err := middleware.GetUserFromContext(ua.ctx); err == nil {
		p.Updater = authUser.GetUsername()
	}

	req := ua.req

	if req.BusinessID != "" {
		p.BusinessID = req.BusinessID
	}

	return ua.model.UpdateProject(ua.ctx, p)
}
