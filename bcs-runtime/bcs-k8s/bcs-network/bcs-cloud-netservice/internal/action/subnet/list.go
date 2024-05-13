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

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
)

// ListAction action to list subnets
type ListAction struct {
	req  *pb.ListSubnetReq
	resp *pb.ListSubnetResp

	ctx context.Context

	storeIf store.Interface

	subnets []*types.CloudSubnet
}

// NewListAction create list action
func NewListAction(ctx context.Context,
	req *pb.ListSubnetReq, resp *pb.ListSubnetResp,
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

// validate input parameters
func (a *ListAction) validate() error {
	return nil
}

// Input do something before Do function
func (a *ListAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ListAction) Output() error {
	for _, sn := range a.subnets {
		a.resp.Subnets = append(a.resp.Subnets, &pbcommon.CloudSubnet{
			VpcID:          sn.VpcID,
			Region:         sn.Region,
			Zone:           sn.Zone,
			SubnetID:       sn.SubnetID,
			SubnetCidr:     sn.SubnetCidr,
			State:          sn.State,
			MinIPNumPerEni: sn.MinIPNumPerEni,
			CreateTime:     sn.CreateTime,
			UpdateTime:     sn.UpdateTime,
		})
	}
	return nil
}

func (a *ListAction) listSubnets() (pbcommon.ErrCode, string) {
	labelsMap := make(map[string]string)
	if len(a.req.VpcID) != 0 {
		labelsMap[kube.CrdNameLabelsVpcID] = a.req.VpcID
	}
	if len(a.req.Region) != 0 {
		labelsMap[kube.CrdNameLabelsRegion] = a.req.Region
	}
	subnets, err := a.storeIf.ListSubnet(a.ctx, labelsMap)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	a.subnets = subnets
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do list action
func (a *ListAction) Do() error {
	if errCode, errMsg := a.listSubnets(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
