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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// FindAvailableAction action to find available subnet
type FindAvailableAction struct {
	req  *pb.GetAvailableSubnetReq
	resp *pb.GetAvailableSubnetResp

	ctx context.Context

	storeIf store.Interface

	cloudIf cloud.Interface

	subnetsFromStore []*types.CloudSubnet

	subnet *types.CloudSubnet
}

// NewFindAvailableAction create FindAvailableAction
func NewFindAvailableAction(ctx context.Context,
	req *pb.GetAvailableSubnetReq, resp *pb.GetAvailableSubnetResp,
	storeIf store.Interface, cloudIf cloud.Interface) *FindAvailableAction {

	action := &FindAvailableAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
		cloudIf: cloudIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *FindAvailableAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *FindAvailableAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Zone, "zone"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input do somthing before Do function
func (a *FindAvailableAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *FindAvailableAction) Output() error {
	if a.subnet != nil {
		a.resp.Subnet = &pbcommon.CloudSubnet{
			VpcID:          a.subnet.VpcID,
			Region:         a.subnet.Region,
			Zone:           a.subnet.Zone,
			SubnetID:       a.subnet.SubnetID,
			SubnetCidr:     a.subnet.SubnetCidr,
			AvailableIPNum: uint64(a.subnet.AvailableIPNum),
			MinIPNumPerEni: a.subnet.MinIPNumPerEni,
			State:          a.subnet.State,
			CreateTime:     a.subnet.CreateTime,
			UpdateTime:     a.subnet.UpdateTime,
		}
	}
	return nil
}

// query subnet list from store
func (a *FindAvailableAction) querySubnetListFromStore() (pbcommon.ErrCode, string) {
	labelsMap := make(map[string]string)
	if len(a.req.VpcID) != 0 {
		labelsMap[kube.CrdNameLabelsVpcID] = a.req.VpcID
	}
	if len(a.req.Region) != 0 {
		labelsMap[kube.CrdNameLabelsRegion] = a.req.Region
	}
	if len(a.req.Zone) != 0 {
		labelsMap[kube.CrdNameLabelsZone] = a.req.Zone
	}
	subnets, err := a.storeIf.ListSubnet(a.ctx, labelsMap)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("list subnet from store failed err %s", err.Error())
	}
	if len(subnets) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("no subnet in store")
	}
	a.subnetsFromStore = subnets
	return pbcommon.ErrCode_ERROR_OK, ""
}

// query available subnet
func (a *FindAvailableAction) querySubnetListFromCloud() (pbcommon.ErrCode, string) {
	var selectedSubnet *types.CloudSubnet
	for _, subnet := range a.subnetsFromStore {
		subnetCloud, err := a.cloudIf.DescribeSubnet(subnet.VpcID, subnet.Region, subnet.SubnetID)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_FAILED,
				fmt.Sprintf("describe subnet from cloud failed, err %s", err.Error())
		}
		err = a.storeIf.UpdateSubnetAvailableIP(a.ctx, subnet.SubnetID, subnetCloud.AvailableIPNum)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("update subnet available ip failed, err %s", err.Error())
		}
		if subnet.State == types.SubnetStatusDisabled {
			continue
		}
		if selectedSubnet == nil {
			selectedSubnet = subnetCloud
		} else {
			if subnetCloud.AvailableIPNum > selectedSubnet.AvailableIPNum &&
				subnetCloud.AvailableIPNum > types.SubnetLeastIPNum {
				selectedSubnet = subnetCloud
			}
		}
	}
	a.subnet = selectedSubnet
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do find available action
func (a *FindAvailableAction) Do() error {
	// list all subnet
	if errCode, errMsg := a.querySubnetListFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// query all subnet from cloud, and update available ip number
	if errCode, errMsg := a.querySubnetListFromCloud(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
