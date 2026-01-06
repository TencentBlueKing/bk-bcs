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

package cloudvpc

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/golang/protobuf/ptypes/wrappers"

	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// UpdateAction update action for cluster vpc
type UpdateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.UpdateCloudVPCRequest
	resp  *cmproto.UpdateCloudVPCResponse
}

// NewUpdateAction create update action for cluster vpc
func NewUpdateAction(model store.ClusterManagerModel) *UpdateAction {
	return &UpdateAction{
		model: model,
	}
}

func (ua *UpdateAction) updateCloudVPC(destCloudVPC *cmproto.CloudVPC) error {
	timeStr := time.Now().UTC().Format(time.RFC3339)
	destCloudVPC.UpdateTime = timeStr
	destCloudVPC.Updater = ua.req.Updater

	if len(ua.req.Region) != 0 {
		destCloudVPC.Region = ua.req.Region
	}
	if len(ua.req.RegionName) != 0 {
		destCloudVPC.RegionName = ua.req.RegionName
	}
	if len(ua.req.NetworkType) != 0 {
		destCloudVPC.NetworkType = ua.req.NetworkType
	}
	if len(ua.req.Available) > 0 {
		destCloudVPC.Available = ua.req.Available
	}
	if ua.req.ReservedIPNum != nil {
		destCloudVPC.ReservedIPNum = ua.req.ReservedIPNum.GetValue()
	}
	if ua.req.BusinessID != nil {
		destCloudVPC.BusinessID = ua.req.GetBusinessID().GetValue()
	}
	if ua.req.GetOverlay() != nil {
		destCloudVPC.Overlay = ua.req.GetOverlay()
	}
	if ua.req.GetUnderlay() != nil {
		destCloudVPC.Underlay = ua.req.GetUnderlay()
	}

	return ua.model.UpdateCloudVPC(ua.ctx, destCloudVPC)
}

func (ua *UpdateAction) validate() error {
	err := ua.req.Validate()
	if err != nil {
		return err
	}

	if ua.req.ReservedIPNum.GetValue() <= 0 {
		ua.req.ReservedIPNum = &wrappers.UInt32Value{Value: 0}
	}

	return nil
}

func (ua *UpdateAction) setResp(code uint32, msg string) {
	ua.resp.Code = code
	ua.resp.Message = msg
	ua.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle handle update cluster vpc
func (ua *UpdateAction) Handle(
	ctx context.Context, req *cmproto.UpdateCloudVPCRequest, resp *cmproto.UpdateCloudVPCResponse) {

	if req == nil || resp == nil {
		blog.Errorf("update cloudVPC failed, req or resp is empty")
		return
	}
	ua.ctx = ctx
	ua.req = req
	ua.resp = resp

	if err := ua.validate(); err != nil {
		ua.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	destCloudVPC, err := ua.model.GetCloudVPC(ua.ctx, req.CloudID, req.VpcID)
	if err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		blog.Errorf("find cloudVPC %s failed when pre-update checking, err %s", req.VpcID, err.Error())
		return
	}
	if err = ua.updateCloudVPC(destCloudVPC); err != nil {
		ua.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err = ua.model.CreateOperationLog(ua.ctx, &cmproto.OperationLog{
		ResourceType: common.CloudVPC.String(),
		ResourceID:   req.VpcID,
		TaskID:       "",
		Message:      fmt.Sprintf("更新云[%s]vpc网络[%s]信息", req.CloudID, req.VpcID),
		OpUser:       req.Updater,
		CreateTime:   time.Now().UTC().Format(time.RFC3339),
		ResourceName: destCloudVPC.GetVpcName(),
	})
	if err != nil {
		blog.Errorf("UpdateCloudVPC[%s] CreateOperationLog failed: %v", req.VpcID, err)
	}

	ua.resp.Data = destCloudVPC
	ua.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
}
