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

package eni

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	actionutils "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/action/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/option"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/lock"
)

const (
	// locker's ttl for subnet
	lockTTL = 10 * time.Second
)

// AllocateAction action for create eni record
type AllocateAction struct {
	req              *pb.AllocateEniReq
	resp             *pb.AllocateEniResp
	cfg              *option.Config
	ctx              context.Context
	storeIf          store.Interface
	lockIf           lock.DistributedLock
	existedIP        *types.IPObject
	candidateSubnets []*types.CloudSubnet
	selectedSubnet   *types.CloudSubnet
	selectedIP       *types.IPObject
	newEniRecord     *types.EniRecord
}

// NewAllocateAction create action for creating eni record
func NewAllocateAction(ctx context.Context,
	cfg *option.Config,
	req *pb.AllocateEniReq, resp *pb.AllocateEniResp,
	storeIf store.Interface, lockIf lock.DistributedLock) *AllocateAction {

	action := &AllocateAction{
		req:     req,
		resp:    resp,
		cfg:     cfg,
		ctx:     ctx,
		storeIf: storeIf,
		lockIf:  lockIf,
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
	if isValid, errMsg := utils.ValidateIDName(a.req.InstanceID, "instanceID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Zone, "zone"); !isValid {
		return errors.New(errMsg)
	}
	if len(a.req.Cluster) == 0 {
		return fmt.Errorf("cluster cannot be empty")
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
	if a.selectedIP != nil {
		a.resp.EniPrimaryIP = &pbcommon.IPObject{
			Address:    a.selectedIP.Address,
			VpcID:      a.selectedIP.VpcID,
			Region:     a.selectedIP.Region,
			SubnetID:   a.selectedIP.SubnetID,
			SubnetCidr: a.selectedIP.SubnetCidr,
			EniID:      utils.GenerateEniName(a.req.InstanceID, a.req.Index),
			Cluster:    a.req.Cluster,
			Host:       a.req.InstanceID,
			IsFixed:    false,
		}
	}
	return nil
}

func (a *AllocateAction) queryAllocatedENIIPObject() (pbcommon.ErrCode, string) {
	ipObjs, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsEni:    utils.GenerateEniName(a.req.InstanceID, a.req.Index),
		kube.CrdNameLabelsStatus: types.IPStatusENIPrimary,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if len(ipObjs) == 0 {
		return pbcommon.ErrCode_ERROR_OK, ""
	}
	if len(ipObjs) != 1 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("get more than one eni ip object for instance %s, index %d", a.req.InstanceID, a.req.Index)
	}
	a.existedIP = ipObjs[0]
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) querySubnetListFromStoreByZone(zone string) (pbcommon.ErrCode, string) {
	subnets, err := a.storeIf.ListSubnet(a.ctx, map[string]string{
		kube.CrdNameLabelsZone: zone,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_LIST_SUBNET_FROM_STORE_FAILED, err.Error()
	}
	// sort by create time
	sort.Slice(subnets, func(i, j int) bool {
		return subnets[i].CreateTime < subnets[j].CreateTime
	})
	for _, subnet := range subnets {
		if subnet.State == types.SubnetStatusEnabled {
			a.candidateSubnets = append(a.candidateSubnets, subnet)
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) findOneSubnet() (pbcommon.ErrCode, string) {
	for _, subnet := range a.candidateSubnets {
		// lock subnet
		if err := a.lockIf.Lock(subnet.SubnetID, []lock.LockOption{lock.LockTTL(lockTTL)}...); err != nil {
			blog.Warnf("lock subnet %s failed, err %s", subnet.SubnetID, err.Error())
			continue
		}
		eniPrimaryList, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
			kube.CrdNameCloudSubnet:  subnet.SubnetID,
			kube.CrdNameLabelsStatus: types.IPStatusENIPrimary,
		})
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_LIST_ENI_FROM_STORE_FAILED, err.Error()
		}
		// need to take this eni into account when calculating, prevent zero dividend
		eniRecordNum := len(eniPrimaryList) + 1
		if float64(subnet.AvailableIPNum)/float64(eniRecordNum) > float64(subnet.MinIPNumPerEni) {
			// a.selectedSubnet = subnet
			if errCode, errMsg := a.findOneEniIP(subnet.SubnetID); errCode != pbcommon.ErrCode_ERROR_OK {
				return errCode, errMsg
			}
			// unlock
			a.lockIf.Unlock(subnet.SubnetID)
			return pbcommon.ErrCode_ERROR_OK, ""
		}
		// unlock
		a.lockIf.Unlock(subnet.SubnetID)
	}
	return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_NO_SUBNET_HAS_ENOUGH_IPS, "no subnet has enough ips"
}

func (a *AllocateAction) findOneEniIP(subnetID string) (pbcommon.ErrCode, string) {
	freeIPList, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsSubnetID: subnetID,
		kube.CrdNameLabelsStatus:   types.IPStatusFree,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if len(freeIPList) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_NO_ENOUGH_IPS,
			fmt.Sprintf("no enough ip in subnet %s", subnetID)
	}
	rand.Seed(time.Now().Unix())
	selectedIP := freeIPList[rand.Intn(len(freeIPList))]
	selectedIP.EniID = utils.GenerateEniName(a.req.InstanceID, a.req.Index)
	selectedIP.Status = types.IPStatusENIPrimary
	selectedIP.Host = a.req.InstanceID
	if _, err = a.storeIf.UpdateIPObject(a.ctx, selectedIP); err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	a.selectedIP = selectedIP
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate eni record
func (a *AllocateAction) Do() error {
	// get existed ip for
	if errCode, errMsg := a.queryAllocatedENIIPObject(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// use existed ip
	if a.existedIP != nil {
		a.selectedIP = a.existedIP
		return nil
	}
	if err := actionutils.CheckIPQuota(a.ctx, a.storeIf, a.req.Cluster); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_NO_ENOUGH_QUOTA, err.Error())
	}
	// query subnets from store
	if errCode, errMsg := a.querySubnetListFromStoreByZone(a.req.Zone); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	// find one available subnet
	if errCode, errMsg := a.findOneSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
