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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/store"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/types"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netservice/internal/utils"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
)

// ReleaseAction action for release ip
type ReleaseAction struct {
	// request for releasing ip
	req *pb.ReleaseIPReq
	// response for releasing ip
	resp *pb.ReleaseIPResp

	ctx context.Context

	storeIf store.Interface
	cloudIf cloud.Interface

	// ip object get from store
	ipObjs []*types.IPObject
}

// NewReleaseAction create ReleaseAction
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseIPReq, resp *pb.ReleaseIPResp,
	storeIf store.Interface, cloudIf cloud.Interface) *ReleaseAction {

	action := &ReleaseAction{
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
func (a *ReleaseAction) Err(errCode pbcommon.ErrCode, errMsg string) error {
	a.resp.ErrCode = errCode
	a.resp.ErrMsg = errMsg
	return errors.New(errMsg)
}

// validate input parameters
func (a *ReleaseAction) validate() error {
	if isValid, errMsg := utils.ValidateIDName(a.req.Cluster, "cluster"); !isValid {
		return errors.New(errMsg)
	}
	if len(a.req.PodName) == 0 {
		return errors.New("podName cannot be empty")
	}
	if len(a.req.PodNamespace) == 0 {
		return errors.New("podNamespace cannot be empty")
	}
	if len(a.req.ContainerID) == 0 {
		return errors.New("containerid cannot be empty")
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

func (a *ReleaseAction) getIPObjectFromStore() (pbcommon.ErrCode, string) {
	ipList, err := a.storeIf.ListIPObjectByField(a.ctx, "spec.containerID", utils.KeyToNamespacedKey(
		constant.CloudCrdNamespaceBcsSystem, a.req.ContainerID))
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	var ipObjsFound []*types.IPObject
	for _, ipObj := range ipList {
		if ipObj.PodName == a.req.PodName &&
			ipObj.ContainerID == a.req.ContainerID &&
			ipObj.Namespace == a.req.PodNamespace &&
			ipObj.Status == types.IPStatusActive {
			ipObjsFound = append(ipObjsFound, ipObj)
		}
	}

	if len(ipObjsFound) == 0 {
		blog.Warnf("active ip not found for pod %s/%s container %s",
			a.req.PodName, a.req.PodNamespace, a.req.ContainerID)
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	a.ipObjs = ipObjsFound
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) unassignIPFromEni(ipObj *types.IPObject) (pbcommon.ErrCode, string) {
	err := a.cloudIf.UnassignIPFromEni([]string{ipObj.Address}, ipObj.EniID)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_CLOUDAPI_UNASSIGNIP_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) updateFixedIPObjectToStore(ipObj *types.IPObject) (pbcommon.ErrCode, string) {
	ipObj.Status = types.IPStatusAvailable
	ipObj.EniID = ""
	ipObj.Host = ""
	ipObj.ContainerID = ""
	_, err := a.storeIf.UpdateIPObject(a.ctx, ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, "update ip object to available failed"
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) changeNonFixedIPObjectToAvailable(ipObj *types.IPObject) (pbcommon.ErrCode, string) {
	ipObj.Status = types.IPStatusAvailable
	ipObj.ContainerID = ""
	ipObj.PodName = ""
	ipObj.Namespace = ""
	_, err := a.storeIf.UpdateIPObject(a.ctx, ipObj)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETSERVICE_STOREOPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.getIPObjectFromStore(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if len(a.ipObjs) == 0 {
		return nil
	}
	for _, ipObj := range a.ipObjs {
		if ipObj.IsFixed {
			if errCode, errMsg := a.unassignIPFromEni(ipObj); errCode != pbcommon.ErrCode_ERROR_OK {
				return a.Err(errCode, errMsg)
			}
			if errCode, errMsg := a.updateFixedIPObjectToStore(ipObj); errCode != pbcommon.ErrCode_ERROR_OK {
				return a.Err(errCode, errMsg)
			}
		} else {
			if errCode, errMsg := a.changeNonFixedIPObjectToAvailable(ipObj); errCode != pbcommon.ErrCode_ERROR_OK {
				return a.Err(errCode, errMsg)
			}
		}
	}

	return nil
}
