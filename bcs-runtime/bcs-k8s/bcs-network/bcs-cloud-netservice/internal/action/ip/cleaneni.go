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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// CleanEniAction action for cleaning all unused ip on one eni
type CleanEniAction struct {
	req  *pb.CleanEniReq
	resp *pb.CleanEniResp

	ctx context.Context

	storeIf store.Interface
	cloudIf cloud.Interface

	ipObjs         []*types.IPObject
	deletingIPObjs []*types.IPObject
}

// NewCleanEniAction create CleanEniAction
func NewCleanEniAction(ctx context.Context,
	req *pb.CleanEniReq, resp *pb.CleanEniResp,
	storeIf store.Interface, cloudIf cloud.Interface) *CleanEniAction {

	action := &CleanEniAction{
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
func (a *CleanEniAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameter
func (a *CleanEniAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.EniID, "eniID"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input accept input data
func (a *CleanEniAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output generate output
func (a *CleanEniAction) Output() error {
	return nil
}

// get all ip on this eni
func (a *CleanEniAction) getEniIPs() (pbcommon.ErrCode, string) {
	existedObjects, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsEni: a.req.EniID,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("list node ips failed, err %s", err.Error())
	}
	a.ipObjs = existedObjects
	return pbcommon.ErrCode_ERROR_OK, ""
}

// check active ips
func (a *CleanEniAction) checkActiveIPs() (pbcommon.ErrCode, string) {
	for _, ip := range a.ipObjs {
		if ip.Status == types.IPStatusActive {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLEAN_ENI_WITH_ACTIVE_IPS,
				fmt.Sprintf("eni %s has active ips", a.req.EniID)
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// update all ips to deleting status
func (a *CleanEniAction) transIPToDeleting() (pbcommon.ErrCode, string) {
	var deletingIPList []*types.IPObject
	for _, ip := range a.ipObjs {
		if ip.Status == types.IPStatusENIPrimary {
			blog.V(3).Infof("skip trans %s to deleting", ip.Address)
		}
		ip.Status = types.IPStatusDeleting
		deletingIP, err := a.storeIf.UpdateIPObject(a.ctx, ip)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("trans ip %s to deleting failed, err %s", ip.Address, err.Error())
		}
		blog.V(3).Infof("trans ip %v to deleting", ip)
		deletingIPList = append(deletingIPList, deletingIP)
	}
	a.deletingIPObjs = deletingIPList
	return pbcommon.ErrCode_ERROR_OK, ""
}

// update all deleting ip to free status
func (a *CleanEniAction) transIPToFree() (pbcommon.ErrCode, string) {
	for _, ip := range a.deletingIPObjs {
		ip.EniID = ""
		ip.Host = ""
		ip.ContainerID = ""
		ip.Cluster = ""
		if ip.IsFixed {
			ip.Status = types.IPStatusAvailable
		} else {
			ip.Status = types.IPStatusFree
		}
		_, err := a.storeIf.UpdateIPObject(a.ctx, ip)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("trans ip %s to deleting failed, err %s", ip.Address, err.Error())
		}
		blog.V(3).Infof("trans deleting %v to free", ip)
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// do clean
func (a *CleanEniAction) cleanEniIPs() (pbcommon.ErrCode, string) {
	eni, err := a.cloudIf.QueryEni(a.req.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_FAILED,
			fmt.Sprintf("query eni %s info failed, err %s", a.req.EniID, err.Error())
	}
	var delIPs []string
	for _, eniIP := range eni.IPs {
		if !eniIP.IsPrimary {
			delIPs = append(delIPs, eniIP.IP)
		}
	}
	if len(delIPs) != 0 {
		if err := a.cloudIf.UnassignIPFromEni(delIPs, a.req.EniID); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_UNASSIGNIP_FAILED,
				fmt.Sprintf("unassign %v from %s failed, err %s", delIPs, a.req.EniID, err.Error())
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do clean node action
func (a *CleanEniAction) Do() error {
	if errCode, errMsg := a.getEniIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if !a.req.IsForce {
		if errCode, errMsg := a.checkActiveIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
			return a.Err(errCode, errMsg)
		}
	}
	if errCode, errMsg := a.transIPToDeleting(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.cleanEniIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.transIPToFree(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
