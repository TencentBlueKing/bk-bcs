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

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// GetAction action to add subnet
type GetAction struct {
	req  *pb.GetIPQuotaReq
	resp *pb.GetIPQuotaResp

	ctx context.Context

	storeIf store.Interface
	ipQuota *types.IPQuota
}

// NewGetAction create GetAction
func NewGetAction(ctx context.Context,
	req *pb.GetIPQuotaReq, resp *pb.GetIPQuotaResp,
	storeIf store.Interface) *GetAction {

	action := &GetAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *GetAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *GetAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input do something before Do function
func (a *GetAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *GetAction) Output() error {
	if a.ipQuota != nil {
		a.resp.Quota = &pbcommon.IPQuota{
			Cluster: a.ipQuota.Cluster,
			Limit:   uint32(a.ipQuota.Limit),
		}
	}
	return nil
}

func (a *GetAction) getQuota(cluster string) (pbcommon.ErrCode, string) {
	quota, err := a.storeIf.GetIPQuota(a.ctx, cluster)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	a.ipQuota = quota
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do get action
func (a *GetAction) Do() error {
	// get quota in storage
	if errCode, errMsg := a.getQuota(a.req.Cluster); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
