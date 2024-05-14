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

package ip

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// TransStatusAction action for trans ip status
type TransStatusAction struct {
	req     *pb.TransIPStatusReq
	resp    *pb.TransIPStatusResp
	ctx     context.Context
	storeIf store.Interface
}

// NewTransStatusAction create action for trans status action
func NewTransStatusAction(ctx context.Context,
	req *pb.TransIPStatusReq, resp *pb.TransIPStatusResp,
	storeIf store.Interface) *TransStatusAction {

	action := &TransStatusAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *TransStatusAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *TransStatusAction) validate() error {
	if len(a.req.Address) == 0 {
		return errors.New("address cannot be empty")
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input input parameters
func (a *TransStatusAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *TransStatusAction) Output() error {
	return nil
}

func (a *TransStatusAction) transIPStatus() (pbcommon.ErrCode, string) {
	ipObj, err := a.storeIf.GetIPObject(a.ctx, a.req.Address)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if ipObj.Status != a.req.SrcStatus {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("ip %s is not %s", a.req.Address, a.req.SrcStatus)
	}
	if ipObj.SubnetID != a.req.SubnetID {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("subnet of ip %s is not %s", a.req.Address, a.req.SubnetID)
	}
	ipObj.Status = a.req.DestStatus
	if _, err := a.storeIf.UpdateIPObject(a.ctx, ipObj); err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do trans ip status action
func (a *TransStatusAction) Do() error {
	if errCode, errMsg := a.transIPStatus(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
