/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package project

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	uuidTool "github.com/satori/go.uuid"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateProjectRequest
	resp  *cmproto.CreateProjectResponse
}

// NewCreateAction create namespace action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createProject() error {
	timeStr := time.Now().Format(time.RFC3339)
	pro := &cmproto.Project{
		ProjectID:   ca.req.ProjectID,
		Name:        ca.req.Name,
		EnglishName: ca.req.EnglishName,
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
		Credentials: ca.req.Credentials,
		CreatTime:   timeStr,
		UpdateTime:  timeStr,
	}
	return ca.model.CreateProject(ca.ctx, pro)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

func (ca *CreateAction) validate() error {
	err := ca.req.Validate()
	if err != nil {
		return err
	}
	// check project validation

	return nil
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateProjectRequest, resp *cmproto.CreateProjectResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create project failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := ca.validate(); err != nil {
		errMsg := fmt.Sprintf("CreateProject validate failed: %s", err.Error())
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, errMsg)
		return
	}

	// build projectID
	if ca.req.ProjectID == "" {
		ca.req.ProjectID = GenerateProjectID()
	}

	if err := ca.createProject(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   ca.req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("创建项目%s", ca.req.Name),
		OpUser:       req.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("CreateProject[%s] CreateOperationLog failed: %v", ca.req.ProjectID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}

// GenerateProjectID build project ID
func GenerateProjectID() string {
	uuid := uuidTool.NewV4()
	buf := make([]byte, 32)
	hex.Encode(buf[:], uuid[:])
	return string(buf[:])
}
