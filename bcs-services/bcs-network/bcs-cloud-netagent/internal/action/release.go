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

package action

import (
	"context"
	"errors"
	"fmt"

	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/common"
	common "github.com/Tencent/bk-bcs/bcs-services/bcs-network/pkg/common"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReleaseAction action for releasing ip
type ReleaseAction struct {
	req  *pb.ReleaseIPReq
	resp *pb.ReleaseIPResp

	ctx context.Context

	k8sIPClient cloudv1set.CloudV1Interface

	cloudNetClient pbcloudnet.CloudNetserviceClient

	nodeNetwork *cloudv1.NodeNetwork

	ipObj *cloudv1.CloudIP
}

// NewReleaseAction create ReleaseAction
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseIPReq, resp *pb.ReleaseIPResp,
	k8sIPClient cloudv1set.CloudV1Interface,
	cloudNetClient pbcloudnet.CloudNetserviceClient) *ReleaseAction {

	action := &ReleaseAction{
		req:            req,
		resp:           resp,
		ctx:            ctx,
		k8sIPClient:    k8sIPClient,
		cloudNetClient: cloudNetClient,
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
	if len(a.req.PodName) == 0 {
		return errors.New("Podname cannot be empty")
	}
	if len(a.req.PodNamespace) == 0 {
		return errors.New("Podnamespace cannot be empty")
	}
	if len(a.req.ContainerID) == 0 {
		return errors.New("pod containerID cannot be empty")
	}
	if len(a.req.IpAddr) == 0 {
		return errors.New("ipAddr cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *ReleaseAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *ReleaseAction) Output() error {

	return nil
}

// get ip object from api server
func (a *ReleaseAction) getIPObjectFromAPIServer() (pbcommon.ErrCode, string) {
	ipObj, err := a.k8sIPClient.CloudIPs(a.req.PodNamespace).Get(a.ctx, a.req.IpAddr, metav1.GetOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	a.ipObj = ipObj
	return pbcommon.ErrCode_ERROR_OK, ""
}

// release ip to cloud netservice
func (a *ReleaseAction) releaseToCloudNetservice() (pbcommon.ErrCode, string) {
	if a.ipObj.Spec.IsFixed {
		newReq := &pbcloudnet.ReleaseFixedIPReq{
			Seq:         common.TimeSequence(),
			VpcID:       a.ipObj.Spec.VpcID,
			Region:      a.ipObj.Spec.Region,
			SubnetID:    a.ipObj.Spec.SubnetID,
			Cluster:     a.ipObj.Spec.Cluster,
			Namespace:   a.ipObj.Spec.Namespace,
			PodName:     a.ipObj.Spec.PodName,
			ContainerID: a.ipObj.Spec.ContainerID,
			EniID:       a.ipObj.Spec.EniID,
			Host:        a.ipObj.Spec.Host,
			Address:     a.ipObj.Spec.Address,
		}

		ipResult, err := a.cloudNetClient.ReleaseFixedIP(a.ctx, newReq)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_RELEASE_IP_FAILED, fmt.Sprintf("call ReleaseFixedIP failed, err %s", err.Error())
		}
		if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
			return ipResult.ErrCode, ipResult.ErrMsg
		}
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	newReq := &pbcloudnet.ReleaseIPReq{
		Seq:         common.TimeSequence(),
		VpcID:       a.ipObj.Spec.VpcID,
		Region:      a.ipObj.Spec.Region,
		SubnetID:    a.ipObj.Spec.SubnetID,
		Cluster:     a.ipObj.Spec.Cluster,
		Namespace:   a.ipObj.Spec.Namespace,
		PodName:     a.ipObj.Spec.PodName,
		ContainerID: a.ipObj.Spec.ContainerID,
		EniID:       a.ipObj.Spec.EniID,
		Host:        a.ipObj.Spec.Host,
		Address:     a.ipObj.Spec.Address,
	}

	ipResult, err := a.cloudNetClient.ReleaseIP(a.ctx, newReq)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_RELEASE_IP_FAILED, fmt.Sprintf("call ReleaseIP failed, err %s", err.Error())
	}
	if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return ipResult.ErrCode, ipResult.ErrMsg
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) deleteIPObjFromAPIServer() (pbcommon.ErrCode, string) {
	// only delete ip related to non-fixed ip
	if !a.ipObj.Spec.IsFixed {
		if err := a.k8sIPClient.CloudIPs(a.ipObj.GetNamespace()).Delete(a.ctx, a.ipObj.GetName(), metav1.DeleteOptions{}); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED,
				fmt.Sprintf("delete ip %s/%s from api server failed, err %s", a.ipObj.GetNamespace(), a.ipObj.GetName(), err.Error())
		}
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do release action
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.getIPObjectFromAPIServer(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.releaseToCloudNetservice(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.deleteIPObjFromAPIServer(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
