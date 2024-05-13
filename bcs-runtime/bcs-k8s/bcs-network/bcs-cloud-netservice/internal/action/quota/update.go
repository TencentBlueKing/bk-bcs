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

// UpdateAction action to add subnet
type UpdateAction struct {
	req  *pb.UpdateIPQuotaReq
	resp *pb.UpdateIPQuotaResp

	ctx context.Context

	storeIf store.Interface
}

// NewUpdateAction create UpdateAction
func NewUpdateAction(ctx context.Context,
	req *pb.UpdateIPQuotaReq, resp *pb.UpdateIPQuotaResp,
	storeIf store.Interface) *UpdateAction {

	action := &UpdateAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *UpdateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *UpdateAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if a.req.Limit < 0 {
		return fmt.Errorf("limit cannot be negative")
	}
	return nil
}

// Input do something before Do function
func (a *UpdateAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *UpdateAction) Output() error {
	return nil
}

func (a *UpdateAction) updateQuota() (pbcommon.ErrCode, string) {
	newQuota := &types.IPQuota{
		Cluster: a.req.Cluster,
		Limit:   int64(a.req.Limit),
	}

	err := a.storeIf.UpdateIPQuota(a.ctx, newQuota)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("store update quota failed, err %s", err.Error())
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do update action
func (a *UpdateAction) Do() error {
	// update quota in storage
	if errCode, errMsg := a.updateQuota(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
