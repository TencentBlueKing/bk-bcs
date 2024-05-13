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

package ip

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	actionutils "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
)

// FixedAllocateAction action for allocate fixed ip
type FixedAllocateAction struct {
	// request for allocating fixed ip
	req *pb.AllocateIPReq
	// response for allocating fixed ip
	resp *pb.AllocateIPResp

	ctx context.Context

	// client for store ip object and subnet
	storeIf store.Interface

	// cloud interface for operate eni ip
	cloudIf cloud.Interface

	// ip already for this pod before
	allocatedIPObj *types.IPObject

	availableIPObj *types.IPObject

	newIPObj *types.IPObject

	ipObj *types.IPObject
}

// NewFixedAllocateAction create FixedAllocateAction
func NewFixedAllocateAction(ctx context.Context,
	req *pb.AllocateIPReq, resp *pb.AllocateIPResp,
	storeIf store.Interface, cloudIf cloud.Interface) *FixedAllocateAction {

	action := &FixedAllocateAction{
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
func (a *FixedAllocateAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *FixedAllocateAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
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
	if len(a.req.EniID) == 0 {
		return errors.New("eniID cannot be empty")
	}
	if len(a.req.KeepDuration) == 0 {
		a.req.KeepDuration = constant.DefaultFixedIPKeepDurationStr
	} else {
		duration, err := time.ParseDuration(a.req.KeepDuration)
		if err != nil {
			return fmt.Errorf("parse keep duration %s failed, err %s", a.req.KeepDuration, err.Error())
		}
		if duration > constant.MaxFixedIPKeepDuration {
			return fmt.Errorf("duration %s is bigger than %f hours",
				a.req.KeepDuration, constant.MaxFixedIPKeepDuration.Hours())
		}
	}
	return nil
}

// Input do something before Do function
func (a *FixedAllocateAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	if err := actionutils.CheckIPQuota(a.ctx, a.storeIf, a.req.Cluster); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_NO_ENOUGH_QUOTA, err.Error())
	}
	return nil
}

// Output do something
func (a *FixedAllocateAction) Output() error {
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
			IsFixed:      a.ipObj.IsFixed,
			Status:       a.ipObj.Status,
		}
	}
	return nil
}

// check allocated ip
func (a *FixedAllocateAction) checkAllocatedIP() (pbcommon.ErrCode, string) {
	ipObjs, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsCluster:   a.req.Cluster,
		kube.CrdNameLabelsSubnetID:  a.req.SubnetID,
		kube.CrdNameLabelsNamespace: a.req.Namespace,
		kube.CrdNameLabelsIsFixed:   strconv.FormatBool(true),
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "list ip object failed"
	}
	if len(ipObjs) == 0 {
		return pbcommon.ErrCode_ERROR_OK, ""
	}
	found := false
	var allocatedIPObj *types.IPObject
	for _, ipObj := range ipObjs {
		if ipObj.PodName == a.req.PodName {
			allocatedIPObj = ipObj
			found = true
			break
		}
	}
	if !found {
		return pbcommon.ErrCode_ERROR_OK, ""
	}
	if allocatedIPObj.Status != types.IPStatusAvailable {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_ALLOCATE_IP_NOT_MATCH,
			fmt.Sprintf("fixed ip %s is in status %s", allocatedIPObj.Address, allocatedIPObj.Status)
	}
	a.allocatedIPObj = allocatedIPObj
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedAllocateAction) assignIPToEni(ip string) (pbcommon.ErrCode, string) {
	_, err := a.cloudIf.AssignIPToEni(ip, a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ASSIGNIP_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// activateAllocatedIP activate allocated ip
func (a *FixedAllocateAction) activateAllocatedIP(ipObj *types.IPObject) (pbcommon.ErrCode, string) {
	var err error
	ipObj.Status = types.IPStatusActive
	ipObj.Host = a.req.Host
	ipObj.EniID = a.req.EniID
	ipObj.ContainerID = a.req.ContainerID
	ipObj.PodName = a.req.PodName
	ipObj.Namespace = a.req.Namespace
	ipObj.KeepDuration = a.req.KeepDuration
	_, err = a.storeIf.UpdateIPObject(a.ctx, ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("update ip object %+v failed, err %s", ipObj, err.Error())
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedAllocateAction) findOneAvailableIP() (pbcommon.ErrCode, string) {
	ipObjs, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsStatus:   types.IPStatusAvailable,
		kube.CrdNameLabelsIsFixed:  strconv.FormatBool(false),
		kube.CrdNameLabelsEni:      a.req.EniID,
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsCluster:  a.req.Cluster,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("query available ip failed, err %s", err.Error())
	}
	if len(ipObjs) == 0 {

	}
	if len(ipObjs) != 0 {
		a.availableIPObj = ipObjs[0]
		a.availableIPObj.Namespace = a.req.Namespace
		a.availableIPObj.PodName = a.req.PodName
		a.availableIPObj.ContainerID = a.req.ContainerID
		a.availableIPObj.Host = a.req.Host
		a.availableIPObj.EniID = a.req.EniID
		a.availableIPObj.IsFixed = true
		a.availableIPObj.Status = types.IPStatusActive
		a.availableIPObj.KeepDuration = a.req.KeepDuration
		a.availableIPObj, err = a.storeIf.UpdateIPObject(a.ctx, a.availableIPObj)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("update ip to store failed, err %s", err.Error())
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedAllocateAction) findOneFreeIP() (pbcommon.ErrCode, string) {
	freeIPList, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsStatus:   types.IPStatusFree,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if len(freeIPList) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_NO_ENOUGH_IPS,
			fmt.Sprintf("no enough ip in subnet %s", a.req.SubnetID)
	}
	rand.Seed(time.Now().Unix())
	selectedIP := freeIPList[rand.Intn(len(freeIPList))]
	selectedIP.EniID = a.req.EniID
	selectedIP.Status = types.IPStatusApplying
	selectedIP.Host = a.req.Host
	selectedIP.ContainerID = a.req.ContainerID
	selectedIP.Cluster = a.req.Cluster
	selectedIP.PodName = a.req.PodName
	selectedIP.Namespace = a.req.Namespace
	selectedIP.IsFixed = true
	selectedIP.KeepDuration = a.req.KeepDuration
	selectedIP, err = a.storeIf.UpdateIPObject(a.ctx, selectedIP)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	a.newIPObj = selectedIP
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *FixedAllocateAction) Do() error {
	// check ip that already allocated
	if errCode, errMsg := a.checkAllocatedIP(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}

	if a.allocatedIPObj != nil {
		// assign ip to eni
		if errCode, errMsg := a.assignIPToEni(
			a.allocatedIPObj.Address); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		// update ip status to active
		if errCode, errMsg := a.activateAllocatedIP(a.allocatedIPObj); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		a.ipObj = a.allocatedIPObj
		return nil
	}

	if errCode, errMsg := a.findOneAvailableIP(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if a.availableIPObj != nil {
		a.ipObj = a.availableIPObj
		return nil
	}

	if errCode, errMsg := a.findOneFreeIP(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if a.newIPObj != nil {
		if errCode, errMsg := a.assignIPToEni(a.newIPObj.Address); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		// update ip status to active
		if errCode, errMsg := a.activateAllocatedIP(a.newIPObj); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		a.ipObj = a.newIPObj
	}
	return nil
}
