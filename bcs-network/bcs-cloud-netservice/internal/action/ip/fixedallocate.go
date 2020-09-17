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
	"math/rand"
	"strconv"
	"time"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netservice/internal/utils"
)

// FixedAllocateAction action for allocate fixed ip
type FixedAllocateAction struct {
	// request for allocating fixed ip
	req *pb.AllocateFixedIPReq
	// response for allocating fixed ip
	resp *pb.AllocateFixedIPResp

	ctx context.Context

	// client for store ip object and subnet
	storeIf store.Interface

	// cloud interface for operate eni ip
	cloudIf cloud.Interface

	// ip already for this pod before
	allocatedIPObj *types.IPObject

	// isSubnetDisabled
	isSubnetDisabled bool

	// eni object
	eni *types.EniObject

	// newly added ip address
	ipAddr string

	// subnet from store
	subnet *types.CloudSubnet

	ipObj *types.IPObject

	// victim ip object
	victimIPObj *types.IPObject
}

// NewFixedAllocateAction create FixedAllocateAction
func NewFixedAllocateAction(ctx context.Context,
	req *pb.AllocateFixedIPReq, resp *pb.AllocateFixedIPResp,
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
		return errors.New("workloadKind cannot be empty")
	}
	if len(a.req.EniID) == 0 {
		return errors.New("eniID cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *FixedAllocateAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
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
		}
	}
	return nil
}

func (a *FixedAllocateAction) querySubnetFromStore() (pbcommon.ErrCode, string) {
	sn, err := a.storeIf.GetSubnet(a.ctx, a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_QUERY_SUBNET_FROM_STORE_FAILED, err.Error()
	}
	// no need to check disabled
	if sn.State == types.SUBNET_STATUS_DISABLED {
		a.isSubnetDisabled = true
	}
	a.subnet = sn
	return pbcommon.ErrCode_ERROR_OK, ""
}

// query eni info, to check
func (a *FixedAllocateAction) queryEniFromCloud() (pbcommon.ErrCode, string) {
	eni, err := a.cloudIf.QueryEni(a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_QUERY_ENI_FAILED, err.Error()
	}
	if a.req.VpcID != eni.VpcID ||
		a.req.Region != eni.Region ||
		a.req.SubnetID != eni.SubnetID {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ENI_INFO_NOTMATCH, "eni info not match request info"
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

// check allocated ip
func (a *FixedAllocateAction) checkAllocatedIP() (pbcommon.ErrCode, string) {
	ipObjs, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsCluster:  a.req.Cluster,
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(true),
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "list ip object failed"
	}

	var allocatedIPObj *types.IPObject
	for _, ipObj := range ipObjs {
		if ipObj.PodName == a.req.PodName &&
			ipObj.Namespace == a.req.Namespace {

			allocatedIPObj = ipObj
		}
	}
	if allocatedIPObj == nil {
		blog.Infof("no allocated ip, try to allocate one")
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	a.allocatedIPObj = allocatedIPObj
	// check info
	if a.allocatedIPObj.VpcID != a.req.VpcID ||
		a.allocatedIPObj.Region != a.req.Region ||
		a.allocatedIPObj.SubnetID != a.req.SubnetID ||
		a.allocatedIPObj.Cluster != a.req.Cluster ||
		a.allocatedIPObj.Namespace != a.req.Namespace ||
		a.allocatedIPObj.PodName != a.req.PodName ||
		a.allocatedIPObj.WorkloadName != a.req.WorkloadName ||
		a.allocatedIPObj.WorkloadKind != a.req.WorkloadKind {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_ALLOCATE_IP_NOT_MATCH,
			"found allocated fixed ip, but info not match"
	}
	// should not happen, may be the last time release is failed
	if a.allocatedIPObj.Status == types.IP_STATUS_ACTIVE {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_TRY_TO_ALLOCATE_ACTIVE_IP,
			"dirty data, request fixed ip is active"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// check if ip is occupied
func (a *FixedAllocateAction) checkIPOccupied() (pbcommon.ErrCode, string) {
	existedIP, err := a.storeIf.GetIPObject(a.ctx, a.req.Address)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("no allocated ip, try to allocate one")
			return pbcommon.ErrCode_ERROR_OK, ""
		}
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "get ip object failed"
	}
	if !existedIP.IsFixed {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_ALLOCATE_IP_NOT_MATCH, "found ip, but it is not fixed ip"
	}
	// check info
	if existedIP.VpcID != a.req.VpcID ||
		existedIP.Region != a.req.Region ||
		existedIP.SubnetID != a.req.SubnetID ||
		existedIP.Cluster != a.req.Cluster ||
		existedIP.Namespace != a.req.Namespace ||
		existedIP.PodName != a.req.PodName ||
		existedIP.WorkloadName != a.req.WorkloadName ||
		existedIP.WorkloadKind != a.req.WorkloadKind {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_ALLOCATE_IP_NOT_MATCH,
			"found allocated fixed ip, but info not match"
	}
	// should not happen, may be the last time release is failed
	if existedIP.Status == types.IP_STATUS_ACTIVE {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_TRY_TO_ALLOCATE_ACTIVE_IP,
			"dirty data, request fixed ip is active"
	}
	a.allocatedIPObj = existedIP
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedAllocateAction) assignIPToEni(ip string) (pbcommon.ErrCode, string) {
	ipAddr, err := a.cloudIf.AssignIPToEni(ip, a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ASSIGNIP_FAILED, err.Error()
	}
	a.ipAddr = ipAddr
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *FixedAllocateAction) createIPObjectToStore() (pbcommon.ErrCode, string) {
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
		IsFixed:      true,
		Status:       types.IP_STATUS_ACTIVE,
	}

	err := a.storeIf.CreateIPObject(a.ctx, ipObject)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}

	a.ipObj = ipObject
	return pbcommon.ErrCode_ERROR_OK, ""
}

// find available ip object applied previous
func (a *FixedAllocateAction) findAvailableVictimIPObject() (pbcommon.ErrCode, string) {
	victimObjects, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsEni:      a.req.EniID,
		kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
		kube.CrdNameLabelsStatus:   types.IP_STATUS_AVAILABLE,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("find available victim ip object failed, err %s", err.Error())
	}
	if len(victimObjects) == 0 {
		a.victimIPObj = nil
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	randIndex := rand.Intn(len(victimObjects))
	a.victimIPObj = victimObjects[randIndex]
	return pbcommon.ErrCode_ERROR_OK, ""
}

// update victim ip object
func (a *FixedAllocateAction) updateVictimIPObject() (pbcommon.ErrCode, string) {
	a.ipObj = &types.IPObject{
		Address:         a.victimIPObj.Address,
		VpcID:           a.victimIPObj.VpcID,
		Region:          a.victimIPObj.Region,
		SubnetID:        a.victimIPObj.SubnetID,
		SubnetCidr:      a.victimIPObj.SubnetCidr,
		Cluster:         a.req.Cluster,
		Namespace:       a.req.Namespace,
		PodName:         a.req.PodName,
		WorkloadName:    a.req.WorkloadName,
		WorkloadKind:    a.req.WorkloadKind,
		ContainerID:     a.req.ContainerID,
		Host:            a.req.Host,
		EniID:           a.req.EniID,
		IsFixed:         true,
		Status:          types.IP_STATUS_ACTIVE,
		ResourceVersion: a.victimIPObj.ResourceVersion,
	}
	err := a.storeIf.UpdateIPObject(a.ctx, a.ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("update victim ip object %+v failed, err %s", a.victimIPObj, err.Error())
	}

	return pbcommon.ErrCode_ERROR_OK, ""
}

// delete victim ip object
func (a *FixedAllocateAction) deleteVictimIPObject() (pbcommon.ErrCode, string) {
	a.victimIPObj.Status = types.IP_STATUS_DELETING
	err := a.storeIf.UpdateIPObject(a.ctx, a.victimIPObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("trans victim ip object %+v to deleting status failed, err %s", a.victimIPObj, err.Error())
	}
	err = a.cloudIf.UnassignIPFromEni(a.victimIPObj.Address, a.victimIPObj.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_ASSIGNIP_FAILED,
			fmt.Sprintf("unassign ip %s from eni %s failed, err %s",
				a.victimIPObj.Address, a.victimIPObj.EniID, err.Error())
	}
	err = a.storeIf.DeleteIPObject(a.ctx, a.victimIPObj.Address)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("delete victim ip object %s failed, err %s", a.victimIPObj.Address, err.Error())
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// migrateIP
func (a *FixedAllocateAction) migrateIP() (pbcommon.ErrCode, string) {
	err := a.cloudIf.MigrateIP(a.allocatedIPObj.Address, a.allocatedIPObj.EniID, a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_MIGRATE_IP_FAILED,
			fmt.Sprintf("migrate ip %s from eni %s to eni %s failed",
				a.allocatedIPObj.Address, a.allocatedIPObj.EniID, a.req.EniID)
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// save ip
func (a *FixedAllocateAction) updateIPObjectToStore(ip string) (pbcommon.ErrCode, string) {
	a.ipObj = &types.IPObject{
		Address:         ip,
		VpcID:           a.allocatedIPObj.VpcID,
		Region:          a.allocatedIPObj.Region,
		SubnetID:        a.allocatedIPObj.SubnetID,
		SubnetCidr:      a.allocatedIPObj.SubnetCidr,
		Cluster:         a.req.Cluster,
		Namespace:       a.req.Namespace,
		PodName:         a.req.PodName,
		WorkloadName:    a.req.WorkloadName,
		WorkloadKind:    a.req.WorkloadKind,
		ContainerID:     a.req.ContainerID,
		Host:            a.req.Host,
		EniID:           a.req.EniID,
		IsFixed:         a.allocatedIPObj.IsFixed,
		Status:          types.IP_STATUS_ACTIVE,
		ResourceVersion: a.allocatedIPObj.ResourceVersion,
	}
	err := a.storeIf.UpdateIPObject(a.ctx, a.ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "update ip failed"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *FixedAllocateAction) Do() error {
	// query subent from store
	if errCode, errMsg := a.querySubnetFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// query eni info from cloud
	if errCode, errMsg := a.queryEniFromCloud(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}

	if len(a.req.Address) != 0 {
		// check ip occupied
		if errCode, errMsg := a.checkIPOccupied(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
	} else {
		// check ip that already allocated
		if errCode, errMsg := a.checkAllocatedIP(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
	}

	if a.allocatedIPObj != nil {
		// case: already allocated before
		// do migrate ip
		if a.allocatedIPObj.EniID != a.req.EniID {
			if errCode, errMsg := a.findAvailableVictimIPObject(); errCode != pbcommon.ErrCode_ERROR_OK {
				return a.Err(errCode, errMsg)
			}
			if a.victimIPObj != nil {
				if errCode, errMsg := a.deleteVictimIPObject(); errCode != pbcommon.ErrCode_ERROR_OK {
					return a.Err(errCode, errMsg)
				}
				// TODO: migrating ip may be failed after delete victim ip to cloud
				time.Sleep(300 * time.Millisecond)
			}
			if errCode, errMsg := a.migrateIP(); errCode != pbcommon.ErrCode_ERROR_OK {
				return a.Err(errCode, errMsg)
			}
		}
		// update to store
		if errCode, errMsg := a.updateIPObjectToStore(a.allocatedIPObj.Address); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		return nil
	}
	// first, find available non-fixed victim ip object
	if errCode, errMsg := a.findAvailableVictimIPObject(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if a.victimIPObj != nil {
		if errCode, errMsg := a.updateVictimIPObject(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
		return nil
	}

	// case: no allocated ip, apply new one and save to store
	if errCode, errMsg := a.assignIPToEni(a.req.Address); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// save to store
	if errCode, errMsg := a.createIPObjectToStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
