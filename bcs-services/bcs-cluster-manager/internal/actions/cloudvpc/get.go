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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	cmproto "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store"
	storeopt "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
)

// GetVPCCidrAction action for get vpc cidr info
type GetVPCCidrAction struct {
	ctx      context.Context
	model    store.ClusterManagerModel
	req      *cmproto.GetVPCCidrRequest
	resp     *cmproto.GetVPCCidrResponse
	cidrList []*cmproto.VPCCidr
}

// NewGetVPCCidrAction create list action for vpc cidr
func NewGetVPCCidrAction(model store.ClusterManagerModel) *GetVPCCidrAction {
	return &GetVPCCidrAction{
		model: model,
	}
}

func (la *GetVPCCidrAction) listVPCCidrList() error {
	condM := make(operator.M)
	condM["vpc"] = la.req.VpcID
	condM["status"] = common.TkeCidrStatusAvailable

	cond := operator.NewLeafCondition(operator.Eq, condM)
	vpcCidrList, err := la.model.ListTkeCidr(la.ctx, cond, &storeopt.ListOption{})
	if err != nil {
		return err
	}

	for _, data := range vpcCidrList {
		la.cidrList = append(la.cidrList, &cmproto.VPCCidr{
			Vpc:      data.VPC,
			Cidr:     data.CIDR,
			IPNumber: data.IPNumber,
			Status:   data.Status,
		})
	}

	return nil
}

func (la *GetVPCCidrAction) setResp(code uint32, msg string) {
	la.resp.Code = code
	la.resp.Message = msg
	la.resp.Result = (code == common.BcsErrClusterManagerSuccess)
	la.resp.Data = la.cidrList
}

// Handle handle list cloud regions
func (la *GetVPCCidrAction) Handle(
	ctx context.Context, req *cmproto.GetVPCCidrRequest, resp *cmproto.GetVPCCidrResponse) {
	if req == nil || resp == nil {
		blog.Errorf("get vpc cidr list failed, req or resp is empty")
		return
	}
	la.ctx = ctx
	la.req = req
	la.resp = resp

	if err := req.Validate(); err != nil {
		la.setResp(common.BcsErrClusterManagerInvalidParameter, err.Error())
		return
	}
	if err := la.listVPCCidrList(); err != nil {
		la.setResp(common.BcsErrClusterManagerDBOperation, err.Error())
		return
	}
	la.setResp(common.BcsErrClusterManagerSuccess, common.BcsErrClusterManagerSuccessStr)
	return
}
