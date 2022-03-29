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
	"encoding/hex"
	"fmt"
	"time"

	uuidTool "github.com/satori/go.uuid"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store"
	pm "github.com/Tencent/bk-bcs/bcs-services/bcs-project/internal/store/project"
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

// Handle create project request
func (ca *CreateAction) Handle(ctx context.Context, req *proto.CreateProjectRequest, resp *proto.ProjectResponse) {
	if req == nil || resp == nil {
		return
	}
	ca.ctx = ctx
	ca.req = req

	if err := ca.validate(); err != nil {
		setResp(resp, common.BcsProjectParamErr, common.BcsProjectParamErrMsg, err.Error(), nil)
		return
	}

	// 如果有传递项目ID，则以传递的为准，否则动态生成32位的字符串作为项目ID
	if req.ProjectID == "" {
		ca.req.ProjectID = GenProjectID()
	}

	if err := ca.createProject(); err != nil {
		setResp(resp, common.BcsProjectDbErr, common.BcsProjectDbErrMsg, err.Error(), nil)
		return
	}

	p, err := ca.model.GetProject(ca.ctx, ca.req.ProjectID)
	if err != nil {
		setResp(resp, common.BcsProjectDbErr, common.BcsProjectDbErrMsg, err.Error(), nil)
		return
	}
	// 返回项目信息
	setResp(resp, common.BcsProjectSuccess, "", common.BcsProjectSuccessMsg, p)
	return
}

func (ca *CreateAction) createProject() error {
	timeStr := time.Now().Format(time.RFC3339)
	p := &proto.Project{
		ProjectID:   ca.req.ProjectID,
		Name:        ca.req.Name,
		ProjectCode: ca.req.ProjectCode,
		Creator:     ca.req.Creator,
		ProjectType: ca.req.ProjectType,
		UseBKRes:    ca.req.UseBKRes,
		Description: ca.req.Description,
		IsOffline:   ca.req.IsOffline,
		Kind:        ca.req.Kind,
		BusinessID:  ca.req.BusinessID,
		DeployType:  ca.req.DeployType,
		BgID:        ca.req.BgID,
		BgName:      ca.req.BgName,
		DeptID:      ca.req.DeptID,
		DeptName:    ca.req.DeptName,
		CenterID:    ca.req.CenterID,
		CenterName:  ca.req.CenterName,
		IsSecret:    ca.req.IsSecret,
		CreateTime:  timeStr,
		UpdateTime:  timeStr,
		Manager:     ca.req.Creator,
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

// GenProjectID generate project ID
func GenProjectID() string {
	uuid := uuidTool.NewV4()
	buf := make([]byte, 32)
	hex.Encode(buf[:], uuid[:])
	return string(buf[:])
}
