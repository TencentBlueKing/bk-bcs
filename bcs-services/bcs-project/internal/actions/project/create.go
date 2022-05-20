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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/auth"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/component/bcscc"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/errorx"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/util/stringx"
	proto "github.com/Tencent/bk-bcs/bcs-services/bcs-project/proto/bcsproject"
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
		return nil, errorx.NewParamErr(err)
	}

	// 如果有传递项目ID，则以传递的为准，否则动态生成32位的字符串作为项目ID
	if req.ProjectID == "" {
		ca.req.ProjectID = stringx.GenUUID()
	}

	if err := ca.createProject(); err != nil {
		return nil, errorx.NewDBErr(err)
	}

	p, err := ca.model.GetProject(ca.ctx, ca.req.ProjectID)
	if err != nil {
		return nil, errorx.NewDBErr(err)
	}
	// 向 bcs cc 写入数据
	go bcscc.CreateProject(p)
	// 返回项目信息
	return p, nil
}

func (ca *CreateAction) createProject() error {
	timeStr := time.Now().Format(time.RFC3339)
	// 从 context 中获取 username
	username := auth.GetUserFromCtx(ca.ctx)
	p := &pm.Project{
		ProjectID:   ca.req.ProjectID,
		Name:        ca.req.Name,
		ProjectCode: ca.req.ProjectCode,
		Creator:     username,
		ProjectType: ca.req.ProjectType,
		UseBKRes:    ca.req.UseBKRes,
		Description: ca.req.Description,
		IsOffline:   ca.req.IsOffline,
		Kind:        ca.req.Kind,
		BusinessID:  ca.req.BusinessID,
		DeployType:  ca.req.DeployType,
		BGID:        ca.req.BGID,
		BGName:      ca.req.BGName,
		DeptID:      ca.req.DeptID,
		DeptName:    ca.req.DeptName,
		CenterID:    ca.req.CenterID,
		CenterName:  ca.req.CenterName,
		IsSecret:    ca.req.IsSecret,
		CreateTime:  timeStr,
		UpdateTime:  timeStr,
		Managers:    username,
	}
	return ca.model.CreateProject(ca.ctx, p)
}

func (ca *CreateAction) validate() error {
	// check projectID、projectCode、name
	projectID, projectCode, name := ca.req.ProjectID, ca.req.ProjectCode, ca.req.Name
	if p, _ := ca.model.GetProjectByField(ca.ctx, &pm.ProjectField{ProjectID: projectID, ProjectCode: projectCode, Name: name}); p != nil {
		if p.ProjectID == projectID {
			return fmt.Errorf("projectID: %s is already exists", projectID)
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
