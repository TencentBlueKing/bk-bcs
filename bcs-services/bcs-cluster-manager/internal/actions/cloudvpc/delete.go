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
 *
 */

package cloudvpc

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
	req   *cmproto.DeleteCloudVPCRequest
	resp  *cmproto.DeleteCloudVPCResponse
}

// NewDeleteAction create delete action for clusterVPC
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

// Handle handle delete cluster vpc
func (da *DeleteAction) Handle(
	ctx context.Context, req *cmproto.DeleteCloudVPCRequest, resp *cmproto.DeleteCloudVPCResponse) {
	if req == nil || resp == nil {
		blog.Errorf("delete cloudVPC failed, req or resp is empty")
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
	deletedCloudVPC, err := da.model.GetCloudVPC(da.ctx, da.req.CloudID, da.req.VpcID)
	if err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("Get CloudVPC %s:%s in pre-delete checking failed, err %s", da.req.CloudID, da.req.VpcID, err.Error())
		return
	}
	da.resp.Data = deletedCloudVPC
	if err = da.model.DeleteCloudVPC(da.ctx, da.req.CloudID, da.req.VpcID); err != nil {
		da.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = da.model.CreateOperationLog(da.ctx, &cmproto.OperationLog{
		ResourceType: common.CloudVPC.String(),
		ResourceID:   req.VpcID,
		TaskID:       "",
		Message:      fmt.Sprintf("删除云[%s]vpc网络[%s]", req.CloudID, req.VpcID),
		OpUser:       deletedCloudVPC.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("DeleteCloudVPC[%s] CreateOperationLog failed: %v", req.VpcID, err)
	}

	da.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
