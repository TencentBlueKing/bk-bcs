/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package ip

import (
	"context"
	"errors"

	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

// ReleaseAction action for release ip
type ReleaseAction struct {
	// request for releasing ip
	req *pb.ReleaseIPReq
	// response for releasing ip
	resp *pb.ReleaseIPResp

	ctx context.Context

	// client for store ip object and subnet
	storeIf store.Interface

	// ip object get from store
	ipObj *types.IPObject
}

// NewReleaseAction create ReleaseAction
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseIPReq, resp *pb.ReleaseIPResp,
	storeIf store.Interface) *ReleaseAction {

	action := &ReleaseAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *ReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *ReleaseAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if len(a.req.Host) == 0 {
		return errors.New("host cannot be empty")
	}
	if len(a.req.PodName) == 0 {
		return errors.New("podname cannot be empty")
	}
	if len(a.req.Namespace) == 0 {
		return errors.New("namespaces cannot be empty")
	}
	if len(a.req.EniID) == 0 {
		return errors.New("eniID cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *ReleaseAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ReleaseAction) Output() error {
	return nil
}

func (a *ReleaseAction) getIPObjectFromStore() (pbcommon.ErrCode, string) {
	ipObj, err := a.storeIf.GetIPObject(a.ctx, a.req.Address)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if ipObj.VpcID != a.req.VpcID ||
		ipObj.Region != a.req.Region ||
		ipObj.EniID != a.req.EniID {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "vpcID, region or eniID not match"
	}

	if ipObj.Cluster != a.req.Cluster ||
		ipObj.Namespace != a.req.Namespace ||
		ipObj.PodName != a.req.PodName {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "container info not match"
	}
	if ipObj.IsFixed {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "ip is fixed, cannot be normally released"
	}

	a.ipObj = ipObj
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) changeIPObjectToAvailable() (pbcommon.ErrCode, string) {
	a.ipObj.Status = types.IP_STATUS_AVAILABLE
	err := a.storeIf.UpdateIPObject(a.ctx, a.ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.getIPObjectFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// when ip object already not active, do not need to be release
	if a.ipObj.Status != types.IP_STATUS_ACTIVE {
		return nil
	}
	if errCode, errMsg := a.changeIPObjectToAvailable(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
