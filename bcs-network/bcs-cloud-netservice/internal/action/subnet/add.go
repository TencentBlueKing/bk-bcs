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

package subnet

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

// AddAction action to add subnet
type AddAction struct {
	req  *pb.AddSubnetReq
	resp *pb.AddSubnetResp

	ctx context.Context

	storeIf store.Interface

	cloudIf cloud.Interface
}

// NewAddAction create AddAction
func NewAddAction(ctx context.Context,
	req *pb.AddSubnetReq, resp *pb.AddSubnetResp,
	storeIf store.Interface, cloudIf cloud.Interface) *AddAction {

	action := &AddAction{
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
func (a *AddAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *AddAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIPv4Cidr(a.req.SubnetCidr); !isValid {
		return errors.New(errMsg)
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

func (a *AddAction) queryCloudSubnet() (pbcommon.ErrCode, string) {
	subnet, err := a.cloudIf.DescribeSubnet(a.req.VpcID, a.req.Region, a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_FAILED,
			fmt.Sprintf("cloud DescribeSubnet failed, err %s", err.Error())
	}

	if subnet.SubnetCidr != a.req.SubnetCidr {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("inconsistent cidr, input cidr %s, cloud cidr %s", a.req.SubnetCidr, subnet.SubnetCidr)
	}

	if subnet.Zone != a.req.Zone {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("inconsistent cidr, input zone %s, cloud zone %s", a.req.Zone, subnet.Zone)
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AddAction) createSubnet() (pbcommon.ErrCode, string) {
	newSubnet := &types.CloudSubnet{
		SubnetID:   a.req.SubnetID,
		VpcID:      a.req.VpcID,
		Region:     a.req.Region,
		Zone:       a.req.Zone,
		SubnetCidr: a.req.SubnetCidr,
		State:      types.SUBNET_STATUS_DISABLED,
	}

	err := a.storeIf.CreateSubnet(a.ctx, newSubnet)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, 
			fmt.Sprintf("store CreateSubnet failed, err %s", err.Error())
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do add action
func (a *AddAction) Do() error {
	// query subnet to cloud
	if errCode, errMsg := a.queryCloudSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// record subnet in storage
	if errCode, errMsg := a.createSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
