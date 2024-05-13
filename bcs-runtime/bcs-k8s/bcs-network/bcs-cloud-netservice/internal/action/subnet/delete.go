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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store/kube"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// DeleteAction action for deleting action
type DeleteAction struct {
	req  *pb.DeleteSubnetReq
	resp *pb.DeleteSubnetResp

	ctx context.Context

	storeIf store.Interface
}

// NewDeleteAction create DeleteAction
func NewDeleteAction(ctx context.Context,
	req *pb.DeleteSubnetReq, resp *pb.DeleteSubnetResp,
	storeIf store.Interface) *DeleteAction {

	action := &DeleteAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *DeleteAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *DeleteAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.SubnetID, "subnetID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.VpcID, "vpcID"); !isValid {
		return errors.New(errMsg)
	}
	if isValid, errMsg := utils.ValidateIDName(a.req.Region, "region"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input do something before Do function
func (a *DeleteAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *DeleteAction) Output() error {
	return nil
}

func (a *DeleteAction) querySubnet() (pbcommon.ErrCode, string) {
	sn, err := a.storeIf.GetSubnet(a.ctx, a.req.SubnetID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if sn.VpcID != a.req.VpcID ||
		sn.Region != a.req.Region {

		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, "vpcID or region not match"
	}
	if sn.State == types.SubnetStatusEnabled {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_TRY_TO_DELETE_ENABLED_SUBNET, "try delete enabled subnet"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *DeleteAction) checkIPObject(status string) (pbcommon.ErrCode, string) {
	ipList, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsStatus:   status,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "check ip object failed"
	}
	if len(ipList) != 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_TRY_TO_DELETE_ACTIVE_SUBNET, "try to delete a active subnet"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *DeleteAction) deleteIPObjectsInStatus(status string) (pbcommon.ErrCode, string) {
	ipList, err := a.storeIf.ListIPObject(a.ctx, map[string]string{
		kube.CrdNameLabelsSubnetID: a.req.SubnetID,
		kube.CrdNameLabelsStatus:   status,
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("list %s ip object failed", status)
	}
	for _, ip := range ipList {
		if err := a.storeIf.DeleteIPObject(a.ctx, ip.Address); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
				fmt.Sprintf("delete %s ip %s failed, err %s", status, ip.Address, err.Error())
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *DeleteAction) deleteSubnet() (pbcommon.ErrCode, string) {
	if err := a.storeIf.DeleteSubnet(a.ctx, a.req.SubnetID); err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED,
			fmt.Sprintf("store DeleteSubnet failed, err %s", err.Error())
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do delete action
func (a *DeleteAction) Do() error {
	if errCode, errMsg := a.querySubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.checkIPObject(types.IPStatusActive); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.checkIPObject(types.IPStatusAvailable); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.checkIPObject(types.IPStatusDeleting); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.checkIPObject(types.IPStatusApplying); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.checkIPObject(types.IPStatusENIPrimary); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.deleteIPObjectsInStatus(types.IPStatusReserved); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.deleteIPObjectsInStatus(types.IPStatusFree); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.deleteSubnet(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
