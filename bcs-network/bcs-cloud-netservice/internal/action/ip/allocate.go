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

package ip

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

// AllocateAction action for allocate ip
type AllocateAction struct {
	// request for allocate ip
	req *pb.AllocateIPReq

	// response for allocate ip
	resp *pb.AllocateIPResp

	ctx context.Context

	// client for store ip object and subnet
	storeIf store.Interface

	// cloud interface for operate eni ip
	cloudIf cloud.Interface

	// previous applied available ip object
	availableIPObj *types.IPObject

	// applied address from cloud
	ipAddr string

	// subnet object
	subnet *types.CloudSubnet

	// final ip object which will be returned to client side
	ipObj *types.IPObject
}

// NewAllocateAction create AllocateAction
func NewAllocateAction(ctx context.Context,
	req *pb.AllocateIPReq, resp *pb.AllocateIPResp,
	storeIf store.Interface, cloudIf cloud.Interface) *AllocateAction {

	action := &AllocateAction{
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
func (a *AllocateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *AllocateAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if len(a.req.Host) == 0 {
		return errors.New("host cannot be empty")
	}
	if len(a.req.PodName) == 0 {
		return errors.New("podname cannot be empty")
	}
	if len(a.req.Namespace) == 0 {
		return errors.New("namespaces cannot be empty")
	}
	if len(a.req.WorkloadName) == 0 {
		return errors.New("workloadName cannot be empty")
	}
	if len(a.req.WorkloadKind) == 0 {
		return errors.New("workloadName cannot be empty")
	}
	if len(a.req.EniID) == 0 {
		return errors.New("eniID cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *AllocateAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *AllocateAction) Output() error {
	if a.ipObj != nil {
		a.resp.Ip = &pbcommon.IPObject{
			Address:      a.ipObj.Address,
			VpcID:        a.ipObj.VpcID,
			Region:       a.ipObj.Region,
			SubnetID:     a.ipObj.SubnetID,
			SubnetCidr:   a.ipObj.SubnetCidr,
			Cluster:      a.ipObj.Cluster,
			Namespace:    a.ipObj.Namespace,
			PodName:      a.ipObj.PodName,
			WorkloadName: a.ipObj.WorkloadName,
			WorkloadKind: a.ipObj.WorkloadKind,
			ContainerID:  a.ipObj.ContainerID,
			Host:         a.ipObj.Host,
			EniID:        a.ipObj.EniID,
			IsFixed:      false,
		}
	}
	return nil
}

func (a *AllocateAction) queryAvailableIPObjectFromStore() (pbcommon.ErrCode, string) {
	ipObjs, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsStatus:   types.IP_STATUS_AVAILABLE,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
		kube.CrdNameLabelsEni:      a.req.EniID,
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsCluster:  a.req.Cluster,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("query available ip failed, err %s", err.Error())
	}
	if len(ipObjs) != 0 {
		a.availableIPObj = ipObjs[0]
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) updateIPObjectToStore() (pbcommon.ErrCode, string) {
	newObj := &types.IPObject{
		Address:         a.availableIPObj.Address,
		VpcID:           a.availableIPObj.VpcID,
		Region:          a.availableIPObj.Region,
		SubnetID:        a.availableIPObj.SubnetID,
		SubnetCidr:      a.availableIPObj.SubnetCidr,
		Cluster:         a.availableIPObj.Cluster,
		Namespace:       a.req.Namespace,
		PodName:         a.req.PodName,
		WorkloadName:    a.req.WorkloadName,
		WorkloadKind:    a.req.WorkloadKind,
		ContainerID:     a.req.ContainerID,
		Host:            a.availableIPObj.Host,
		EniID:           a.availableIPObj.EniID,
		IsFixed:         false,
		Status:          types.IP_STATUS_ACTIVE,
		ResourceVersion: a.availableIPObj.ResourceVersion,
	}
	err := a.storeIf.UpdateIPObject(a.ctx, newObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("update ip to store failed, err %s", err.Error())
	}
	a.ipObj = newObj
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) querySubnetFromStore() (pbcommon.ErrCode, string) {
	sn, err := a.storeIf.GetSubnet(a.ctx, a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_QUERY_SUBNET_FROM_STORE_FAILED,
			fmt.Sprintf("get subnet from store failed, err %s", err.Error())
	}
	if sn.State == types.SUBNET_STATUS_DISABLED {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_SUBNET_IS_DISABLED,
			"subnet is disabled"
	}
	a.subnet = sn
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) queryEniFromCloud() (pbcommon.ErrCode, string) {
	eni, err := a.cloudIf.QueryEni(a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_QUERY_ENI_FAILED,
			fmt.Sprintf("query eni from cloud failed, err %s", err.Error())
	}
	if a.req.VpcID != eni.VpcID ||
		a.req.Region != eni.Region ||
		a.req.SubnetID != eni.SubnetID {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ENI_INFO_NOTMATCH,
			"eni info not match request info"
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) assignIPToEni() (pbcommon.ErrCode, string) {
	ipAddr, err := a.cloudIf.AssignIPToEni("", a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ASSIGNIP_FAILED,
			fmt.Sprintf("assign ip from cloud to eni %s failed, err %s", a.req.EniID, err.Error())
	}
	a.ipAddr = ipAddr
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) createIPObjectToStore() (pbcommon.ErrCode, string) {
	ipObject := &types.IPObject{
		Address:      a.ipAddr,
		VpcID:        a.req.VpcID,
		Region:       a.req.Region,
		SubnetID:     a.req.SubnetID,
		SubnetCidr:   a.subnet.SubnetCidr,
		Cluster:      a.req.Cluster,
		Namespace:    a.req.Namespace,
		PodName:      a.req.PodName,
		WorkloadName: a.req.WorkloadName,
		WorkloadKind: a.req.WorkloadKind,
		ContainerID:  a.req.ContainerID,
		Host:         a.req.Host,
		EniID:        a.req.EniID,
		IsFixed:      false,
		Status:       types.IP_STATUS_ACTIVE,
	}

	err := a.storeIf.CreateIPObject(a.ctx, ipObject)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("create ip to store failed, err %s", err.Error())
	}

	a.ipObj = ipObject
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *AllocateAction) Do() error {
	// query available ip object for request eni
	if errCode, errMsg := a.queryAvailableIPObjectFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if a.availableIPObj == nil {
		// if no found previous applied ip, apply from cloud

		// query subent from store
		if errCode, errMsg := a.querySubnetFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		// query eni info from cloud
		if errCode, errMsg := a.queryEniFromCloud(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		// assign ip to eni
		if errCode, errMsg := a.assignIPToEni(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		// record ip object in storage
		if errCode, errMsg := a.createIPObjectToStore(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		return nil
	}
	// if available previous applied ip found, than update the ip object, and return to user
	if errCode, errMsg := a.updateIPObjectToStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}

	return nil
}
