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

package quota

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

// AddAction action to add subnet
type AddAction struct {
	req  *pb.CreateIPQuotaReq
	resp *pb.CreateIPQuotaResp

	ctx context.Context

	storeIf store.Interface
}

// NewAddAction create AddAction
func NewAddAction(ctx context.Context,
	req *pb.CreateIPQuotaReq, resp *pb.CreateIPQuotaResp,
	storeIf store.Interface) *AddAction {

	action := &AddAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *AddAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *AddAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if a.req.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	return nil
}

// Input do something before Do function
func (a *AddAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *AddAction) Output() error {
	return nil
}

func (a *AddAction) createQuota() (pbcommon.ErrCode, string) {
	newQuota := &types.IPQuota{
		Cluster: a.req.Cluster,
		Limit:   int64(a.req.Limit),
	}

	err := a.storeIf.CreateIPQuota(a.ctx, newQuota)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("store CreateSubnet failed, err %s", err.Error())
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do add action
func (a *AddAction) Do() error {
	// add quota in storage
	if errCode, errMsg := a.createQuota(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
