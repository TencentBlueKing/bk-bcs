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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
)

// AddAction action to add subnet
type AddAction struct {
	req  *pb.AddSubnetReq
	resp *pb.AddSubnetResp

	ctx context.Context

	storeIf store.Interface
	cloudIf cloud.Interface

	cloudSubnet *types.CloudSubnet
	eniList     []*types.EniObject
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
	if a.req.MinIPNumPerEni < 0 {
		return fmt.Errorf("minIPNumPerEni cannot be negative")
	} else if a.req.MinIPNumPerEni == 0 {
		a.req.MinIPNumPerEni = constant.DefaultMinIPNumPerEni
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
	a.cloudSubnet = subnet
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AddAction) queryEnis() (pbcommon.ErrCode, string) {
	eniList, err := a.cloudIf.QueryEniList(a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_FAILED,
			fmt.Sprintf("cloud QueryEniList failed, err %s", err.Error())
	}
	a.eniList = eniList
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AddAction) createReservedIPs() (pbcommon.ErrCode, string) {
	reservedIPSet := make(map[string]struct{})
	for _, eni := range a.eniList {
		for _, ip := range eni.IPs {
			if err := a.storeIf.CreateIPObject(a.ctx, &types.IPObject{
				Address:    ip.IP,
				VpcID:      eni.VpcID,
				Region:     eni.Region,
				SubnetID:   eni.SubnetID,
				SubnetCidr: a.cloudSubnet.SubnetCidr,
				EniID:      eni.EniID,
				IsFixed:    false,
				Status:     types.IPStatusReserved,
			}); err != nil {
				return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
			}
			reservedIPSet[ip.IP] = struct{}{}
		}
	}
	allIPList, err := utils.GetIPListFromCidr(a.cloudSubnet.SubnetCidr)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error()
	}
	for _, ip := range allIPList {
		if _, ok := reservedIPSet[ip]; !ok {
			if err := a.storeIf.CreateIPObject(a.ctx, &types.IPObject{
				Address:    ip,
				VpcID:      a.cloudSubnet.VpcID,
				Region:     a.cloudSubnet.Region,
				SubnetID:   a.cloudSubnet.SubnetID,
				SubnetCidr: a.cloudSubnet.SubnetCidr,
				IsFixed:    false,
				Status:     types.IPStatusFree,
			}); err != nil {
				return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
			}
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AddAction) createSubnet() (pbcommon.ErrCode, string) {
	newSubnet := &types.CloudSubnet{
		SubnetID:       a.req.SubnetID,
		VpcID:          a.req.VpcID,
		Region:         a.req.Region,
		Zone:           a.req.Zone,
		SubnetCidr:     a.req.SubnetCidr,
		AvailableIPNum: a.cloudSubnet.AvailableIPNum,
		MinIPNumPerEni: a.req.MinIPNumPerEni,
		State:          types.SubnetStatusDisabled,
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
	// query enis
	if errCode, errMsg := a.queryEnis(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// record subnet in storage
	if errCode, errMsg := a.createSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// record reserved ip in storage
	if errCode, errMsg := a.createReservedIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
