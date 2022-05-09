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
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for online cluster credential
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateProjectRequest
	resp  *cmproto.UpdateProjectResponse
}

// NewUpdateAction create update action for online cluster credential
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateProject(pro *cmproto.Project) error {
	timeStr := time.Now().Format(time.RFC3339)
	// update field if required
	pro.UpdateTime = timeStr
	pro.Updater = ua.req.Updater

	if len(ua.req.Name) != 0 {
		pro.Name = ua.req.Name
	}
	if ua.req.ProjectType > 0 {
		pro.ProjectType = ua.req.ProjectType
	}

	if ua.req.UseBKRes != nil && ua.req.UseBKRes.GetValue() != pro.UseBKRes {
		pro.UseBKRes = ua.req.UseBKRes.GetValue()
	}

	if len(ua.req.Description) != 0 {
		pro.Description = ua.req.Description
	}
	if ua.req.IsOffline != nil && ua.req.IsOffline.GetValue() != pro.IsOffline {
		pro.IsOffline = ua.req.IsOffline.GetValue()
	}
	if len(ua.req.Kind) != 0 {
		pro.Kind = ua.req.Kind
	}
	if ua.req.DeployType > 0 {
		pro.DeployType = ua.req.DeployType
	}
	if len(ua.req.BgID) != 0 {
		pro.BgID = ua.req.BgID
	}
	if len(ua.req.BgName) != 0 {
		pro.BgName = ua.req.BgName
	}
	if len(ua.req.DeptID) != 0 {
		pro.DeptID = ua.req.DeptID
	}
	if len(ua.req.DeptName) != 0 {
		pro.DeptName = ua.req.DeptName
	}
	if len(ua.req.CenterID) != 0 {
		pro.CenterID = ua.req.CenterID
	}
	if len(ua.req.CenterName) != 0 {
		pro.CenterName = ua.req.CenterName
	}
	if ua.req.IsSecret != nil && ua.req.IsSecret.GetValue() != pro.IsSecret {
		pro.IsSecret = ua.req.IsSecret.GetValue()
	}
	if ua.req.Credentials != nil {
		pro.Credentials = ua.req.Credentials
	}
	if len(ua.req.BusinessID) != 0 {
		pro.BusinessID = ua.req.BusinessID
	}
	return ua.model.UpdateProject(ua.ctx, pro)
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster credential
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateProjectRequest, resp *cmproto.UpdateProjectResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update Project failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := req.Validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}

	//get old project information, update fields if required
	destPro, err := ua.model.GetProject(ua.ctx, req.ProjectID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find project %s failed when pre-update checking, err %s", req.ProjectID, err.Error())
		return
	}
	if err := ua.updateProject(destPro); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新项目%s信息", destPro.Name),
		OpUser:       req.Updater,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("UpdateProject[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	ua.resp.Data = destPro
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
