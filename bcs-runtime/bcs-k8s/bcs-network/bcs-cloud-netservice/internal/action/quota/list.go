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
)

// ListAction action for list quotas
type ListAction struct {
	req  *pb.ListIPQuotaReq
	resp *pb.ListIPQuotaResp

	ctx context.Context

	storeIf   store.Interface
	quotaList []*types.IPQuota
}

// NewListAction create list action
func NewListAction(ctx context.Context,
	req *pb.ListIPQuotaReq, resp *pb.ListIPQuotaResp,
	storeIf store.Interface) *ListAction {

	action := &ListAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *ListAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// Input do something before Do function
func (a *ListAction) Input() error {
	return nil
}

// Output do something after Do function
func (a *ListAction) Output() error {
	if len(a.quotaList) != 0 {
		for _, quota := range a.quotaList {
			a.resp.Quotas = append(a.resp.Quotas, &pbcommon.IPQuota{
				Cluster: quota.Cluster,
				Limit:   uint32(quota.Limit),
			})
		}
	}
	return nil
}

func (a *ListAction) listQuotas() (pbcommon.ErrCode, string) {
	quotaList, err := a.storeIf.ListIPQuota(a.ctx)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	a.quotaList = quotaList
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do list action
func (a *ListAction) Do() error {
	if errCode, errMsg := a.listQuotas(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
