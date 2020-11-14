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

// CleanNodeAction action for cleaning all unused ip on one node
type CleanNodeAction struct {
	req  *pb.CleanNodeReq
	resp *pb.CleanNodeResp

	ctx context.Context

	storeIf store.Interface

	cloudIf cloud.Interface

	ipObjs []*types.IPObject
}

// NewCleanNodeAction create CleanNodeAction
func NewCleanNodeAction(ctx context.Context,
	req *pb.CleanNodeReq, resp *pb.CleanNodeResp,
	storeIf store.Interface, cloudIf cloud.Interface) *CleanNodeAction {

	action := &CleanNodeAction{
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
func (a *CleanNodeAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameter
func (a *CleanNodeAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if len(a.req.Host) == 0 {
		return errors.New("host cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *CleanNodeAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *CleanNodeAction) Output() error {
	return nil
}

// get all ip on this node
func (a *CleanNodeAction) getNodeIPs() (pbcommon.ErrCode, string) {
	var availableIPLabels map[string]string
	if a.req.IsForce {
		availableIPLabels = map[string]string{
			kube.CrdNameLabelsHost: a.req.Host,
		}
	} else {
		availableIPLabels = map[string]string{
			kube.CrdNameLabelsHost:     a.req.Host,
			kube.CrdNameLabelsIsFixed: strconv.FormatBool(false),
			kube.CrdNameLabelsStatus:   types.IP_STATUS_AVAILABLE,
		}
	}
	existedObjects, err := a.storeIf.ListIPObject(a.ctx, availableIPLabels)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("list node ips failed, err %s", err.Error())
	}
	a.ipObjs = existedObjects
	return pbcommon.ErrCode_ERROR_OK, ""
}

// do clean
func (a *CleanNodeAction) cleanNodeIPs() (pbcommon.ErrCode, string) {
	for _, ipObj := range a.ipObjs {
		if err := a.cloudIf.UnassignIPFromEni(ipObj.Address, ipObj.EniID); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_UNASSIGNIP_FAILED,
				fmt.Sprintf("unassign %s from %s failed, err %s", ipObj.Address, ipObj.EniID, err.Error())
		}
		if err := a.storeIf.DeleteIPObject(a.ctx, ipObj.Address); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("delete ip object %s failed, err %s", ipObj.Address, err.Error())
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do clean node action
func (a *CleanNodeAction) Do() error {
	if errCode, errMsg := a.getNodeIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.cleanNodeIPs(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
