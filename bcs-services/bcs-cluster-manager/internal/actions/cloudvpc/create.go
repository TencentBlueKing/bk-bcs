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
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
)

// CreateAction action for create namespace
type CreateAction struct {
	ctx   context.Context
	model store.ClusterManagerModel
	req   *cmproto.CreateCloudVPCRequest
	resp  *cmproto.CreateCloudVPCResponse
}

// NewCreateAction create cloudVPC action
func NewCreateAction(model store.ClusterManagerModel) *CreateAction {
	return &CreateAction{
		model: model,
	}
}

func (ca *CreateAction) createCloudVPC() error {
	timeStr := time.Now().Format(time.RFC3339)
	cloudVPC := &cmproto.CloudVPC{
		CloudID:       ca.req.CloudID,
		Region:        ca.req.Region,
		RegionName:    ca.req.RegionName,
		NetworkType:   ca.req.NetworkType,
		VpcID:         ca.req.VpcID,
		VpcName:       ca.req.VpcName,
		Available:     ca.req.Available,
		Extra:         ca.req.Extra,
		Creator:       ca.req.Creator,
		Updater:       ca.req.Creator,
		ReservedIPNum: ca.req.ReservedIPNum,
		BusinessID:    ca.req.BusinessID,
		CreatTime:     timeStr,
		UpdateTime:    timeStr,
	}
	return ca.model.CreateCloudVPC(ca.ctx, cloudVPC)
}

func (ca *CreateAction) setResp(code uint32, msg string) {
	ca.resp.Code = code
	ca.resp.Message = msg
	ca.resp.Result = (code == common.BcsErrClusterManagerSuccess)
}

// Handle create namespace request
func (ca *CreateAction) Handle(ctx context.Context,
	req *cmproto.CreateCloudVPCRequest, resp *cmproto.CreateCloudVPCResponse) {
	if req == nil || resp == nil {
		blog.Errorf("create cloudVPC failed, req or resp is empty")
		return
	}
	ca.ctx = ctx
	ca.req = req
	ca.resp = resp

	if err := req.Validate(); err != nil {
		ca.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := ca.createCloudVPC(); err != nil {
		if errors.Is(err, drivers.ErrTableRecordDuplicateKey) {
			ca.setResp(common.BcsErrClusterManagerDatabaseRecordDuplicateKey, err.Error())
			return
		}
		ca.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}

	err := ca.model.CreateOperationLog(ca.ctx, &cmproto.OperationLog{
		ResourceType: common.CloudVPC.String(),
		ResourceID:   req.VpcID,
		TaskID:       "",
		Message:      fmt.Sprintf("创建云[%s]vpc网络[%s]", req.CloudID, req.VpcID),
		OpUser:       req.Creator,
		CreateTime:   time.Now().String(),
	})
	if err != nil {
		blog.Errorf("CreateCloudVPC[%s] CreateOperationLog failed: %v", req.VpcID, err)
	}

	ca.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
