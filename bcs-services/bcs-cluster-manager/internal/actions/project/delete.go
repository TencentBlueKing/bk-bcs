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

// DeleteAction action for delete clustercredentail
type DeleteAction struct {
	ctx context.Context

	model store.ClusterManagerModel
	req   *cmproto.DeleteProjectRequest
	resp  *cmproto.DeleteProjectResponse
}

// NewDeleteAction create delete action for online cluster credential
func NewDeleteAction(model store.ClusterManagerModel) *DeleteAction {
	return &DeleteAction{
		model: model,
	}
}

func (da *DeleteAction) setResp(code uint32, msg string) {
	da.resp.Code = code
	da.resp.Message = msg
	da.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle delete cluster credential
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteProjectRequest, resp *cmproto.DeleteProjectResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete project failed, req or resp is empty")
		return
	}
	da.ctx = ctx
	da.req = req
	da.resp = resp

	if err := req.Validate(); err != nil {
		da.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	//try to get original data for return
	deletedProject, err := da.model.GetProject(da.ctx, da.req.ProjectID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	da.resp.Data = deletedProject
	if len(deletedProject.ProjectID) == 0 {
		da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
		blog.Infof("project %s does not exist, No Deletion handle.", da.req.ProjectID)
		return
	}
	if err := da.model.DeleteProject(da.ctx, da.req.ProjectID); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.Project.String(),
		ResourceID:   req.ProjectID,
		TaskID:       "",
		Message:      fmt.Sprintf("删除项目%s", deletedProject.Name),
		OpUser:       deletedProject.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("DeleteProject[%s] CreateOperationLog failed: %v", req.ProjectID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
