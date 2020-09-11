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

// FixedReleaseAction action for release fixed ip
type FixedReleaseAction struct {
	req  *pb.ReleaseFixedIPReq
	resp *pb.ReleaseFixedIPResp

	ctx context.Context

	storeIf store.Interface

	ipObj *types.IPObject
}

// NewFixedReleaseAction create FixedReleaseAction
func NewFixedReleaseAction(ctx context.Context,
	req *pb.ReleaseFixedIPReq, resp *pb.ReleaseFixedIPResp,
	storeIf store.Interface) *FixedReleaseAction {

	action := &FixedReleaseAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *FixedReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *FixedReleaseAction) validate() error {
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
	if len(a.req.Address) == 0 {
		return errors.New("address cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *FixedReleaseAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *FixedReleaseAction) Output() error {
	return nil
}

func (a *FixedReleaseAction) getIPObjectFromStore() (pbcommon.ErrCode, string) {
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

	if !ipObj.IsFixed {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "address is not fixed"
	}

	a.ipObj = ipObj
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedReleaseAction) updateIPObjectToStore() (pbcommon.ErrCode, string) {
	a.ipObj.Status = types.IP_STATUS_AVAILABLE
	err := a.storeIf.UpdateIPObject(a.ctx, a.ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "update ip object to available failed"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do release action
func (a *FixedReleaseAction) Do() error {
	if errCode, errMsg := a.getIPObjectFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.updateIPObjectToStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
