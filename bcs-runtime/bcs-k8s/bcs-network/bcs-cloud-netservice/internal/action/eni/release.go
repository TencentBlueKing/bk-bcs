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

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
)

// ReleaseAction action for release eni record
type ReleaseAction struct {
	req     *pb.ReleaseEniReq
	resp    *pb.ReleaseEniResp
	ctx     context.Context
	storeIf store.Interface
}

// NewReleaseAction create action for release eni record
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseEniReq, resp *pb.ReleaseEniResp,
	storeIf store.Interface) *ReleaseAction {

	action := &ReleaseAction{
		req:     req,
		resp:    resp,
		ctx:     ctx,
		storeIf: storeIf,
	}
	action.resp.Seq = req.Seq
	return action
}

// Err set err info
func (a *ReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *ReleaseAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.InstanceID, "instanceID"); !isValid {
		return errors.New(errMsg)
	}
	return nil
}

// Input do something before Do function
func (a *ReleaseAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ReleaseAction) Output() error {
	return nil
}

func (a *ReleaseAction) releaseEniPrimary() (pbcommon.ErrCode, string) {
	primaryIPObj, err := a.storeIf.GetIPObject(a.ctx, a.req.EniPrimaryIP)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	if primaryIPObj.Status != types.IPStatusENIPrimary {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("ip %s is not %s", primaryIPObj.Address, types.IPStatusENIPrimary)
	}
	eniName := utils.GenerateEniName(a.req.InstanceID, a.req.Index)
	if primaryIPObj.EniID != eniName {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_INVALID_PARAMS,
			fmt.Sprintf("host %s of ip %s is not %s", primaryIPObj.Host, primaryIPObj.Address, eniName)
	}
	primaryIPObj.Status = types.IPStatusFree
	primaryIPObj.EniID = ""
	primaryIPObj.Host = ""
	primaryIPObj.Cluster = ""
	if _, err := a.storeIf.UpdateIPObject(a.ctx, primaryIPObj); err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do release eni record
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.releaseEniPrimary(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
