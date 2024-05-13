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

package subnet

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// ChangeAction action to change subnet state
type ChangeAction struct {
	req  *pb.ChangeSubnetReq
	resp *pb.ChangeSubnetResp

	ctx context.Context

	storeIf store.Interface

	subnet *types.CloudSubnet
}

// NewChangeAction create ChangeAction
func NewChangeAction(ctx context.Context,
	req *pb.ChangeSubnetReq, resp *pb.ChangeSubnetResp,
	storeIf store.Interface) *ChangeAction {

	action := &ChangeAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *ChangeAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *ChangeAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input do something before Do function
func (a *ChangeAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ChangeAction) Output() error {
	return nil
}

func (a *ChangeAction) querySubnet() (pbcommon.ErrCode, string) {
	subnet, err := a.storeIf.GetSubnet(a.ctx, a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if subnet.VpcID != a.req.VpcID {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "vpcID not match"
	}
	a.subnet = subnet
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ChangeAction) changeSubnet() (pbcommon.ErrCode, string) {
	var err error
	// when minIPNum PerEni
	if a.req.MinIPNumPerEni <= 0 {
		err = a.storeIf.UpdateSubnetState(a.ctx, a.req.SubnetID, int32(a.req.State), a.subnet.MinIPNumPerEni)
	} else {
		err = a.storeIf.UpdateSubnetState(a.ctx, a.req.SubnetID, int32(a.req.State), a.req.MinIPNumPerEni)
	}
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("store CreateSubnet failed, err %s", err.Error())
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do change action
func (a *ChangeAction) Do() error {
	// query subnet in storage
	if errCode, errMsg := a.querySubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// change subnet in storage
	if errCode, errMsg := a.changeSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
