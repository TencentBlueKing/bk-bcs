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

package action

import (
	"context"
	"errors"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/inspector"
	common "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReleaseAction action for releasing ip
type ReleaseAction struct {
	req  *pb.ReleaseIPReq
	resp *pb.ReleaseIPResp

	ctx context.Context

	cloudNetClient pbcloudnet.CloudNetserviceClient

	k8sIPClient cloudv1set.CloudV1Interface

	inspector *inspector.NodeNetworkInspector

	nodeNetwork *cloudv1.NodeNetwork
}

// NewReleaseAction create ReleaseAction
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseIPReq, resp *pb.ReleaseIPResp,
	cloudNetClient pbcloudnet.CloudNetserviceClient, k8sIPClient cloudv1set.CloudV1Interface,
	inspector *inspector.NodeNetworkInspector) *ReleaseAction {

	action := &ReleaseAction{
		req:            req,
		resp:           resp,
		ctx:            ctx,
		cloudNetClient: cloudNetClient,
		inspector:      inspector,
		k8sIPClient:    k8sIPClient,
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

func (a *ReleaseAction) getNodeInfo() (pbcommon.ErrCode, string) {
	nodeNetwork := a.inspector.GetNodeNetwork()
	if nodeNetwork == nil || len(nodeNetwork.Status.Enis) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE, fmt.Sprintf("node eni not ready")
	}
	a.nodeNetwork = nodeNetwork
	return pbcommon.ErrCode_ERROR_OK, ""
}

// release ip to cloud netservice
func (a *ReleaseAction) releaseToCloudNetservice() (pbcommon.ErrCode, string) {
	newReq := &pbcloudnet.ReleaseIPReq{
		Seq:          common.TimeSequence(),
		Cluster:      a.inspector.GetCluster(),
		PodName:      a.req.PodName,
		PodNamespace: a.req.PodNamespace,
		ContainerID:  a.req.ContainerID,
	}

	ipResult, err := a.cloudNetClient.ReleaseIP(a.ctx, newReq)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_RELEASE_IP_FAILED,
			fmt.Sprintf("call ReleaseIP failed, err %s", err.Error())
	}
	if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return ipResult.ErrCode, ipResult.ErrMsg
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) deleteIPObjFromAPIServer() (pbcommon.ErrCode, string) {
	if err := a.k8sIPClient.CloudIPs(a.req.PodNamespace).
		Delete(a.ctx, a.req.PodName, metav1.DeleteOptions{}); err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Infof("cluster cloudIP of pod %s/%s not found, no need to delete", a.req.PodName, a.req.PodNamespace)
			return pbcommon.ErrCode_ERROR_OK, ""
		}
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED,
			fmt.Sprintf("delete ip %s/%s from api server failed, err %s",
				a.req.PodNamespace, a.req.PodName, err.Error())
	}
	a.inspector.GetIPCache().DeleteEniIPbyContainerID(a.req.ContainerID)
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do release action
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.getNodeInfo(); errCode != pbcommon.ErrCode_ERROR_OK {
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
