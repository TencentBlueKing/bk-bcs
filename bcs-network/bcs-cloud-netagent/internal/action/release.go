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
	"strconv"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"

	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netagent/internal/inspector"
	"github.com/Tencent/bk-bcs/bcs-network/internal/constant"
	common "github.com/Tencent/bk-bcs/bcs-network/pkg/common"
)

// ReleaseAction action for releasing ip
type ReleaseAction struct {
	req  *pb.ReleaseIPReq
	resp *pb.ReleaseIPResp

	ctx context.Context

	k8sIPClient cloudv1set.CloudV1Interface

	cloudNetClient pbcloudnet.CloudNetserviceClient

	inspector *inspector.NodeNetworkInspector

	nodeNetwork *cloudv1.NodeNetwork

	ipObj *cloudv1.CloudIP
}

// NewReleaseAction create ReleaseAction
func NewReleaseAction(ctx context.Context,
	req *pb.ReleaseIPReq, resp *pb.ReleaseIPResp,
	k8sIPClient cloudv1set.CloudV1Interface,
	cloudNetClient pbcloudnet.CloudNetserviceClient,
	inspector *inspector.NodeNetworkInspector) *ReleaseAction {

	action := &ReleaseAction{
		req:            req,
		resp:           resp,
		ctx:            ctx,
		k8sIPClient:    k8sIPClient,
		cloudNetClient: cloudNetClient,
		inspector:      inspector,
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
	if nodeNetwork == nil ||
		nodeNetwork.Status.FloatingIPEni == nil ||
		nodeNetwork.Status.Status != cloudv1.NodeNetworkStatusReady {

		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE, fmt.Sprintf("node eni not ready")
	}
	a.nodeNetwork = nodeNetwork
	return pbcommon.ErrCode_ERROR_OK, ""
}

// get ip object from api server
func (a *ReleaseAction) getIPObjectFromAPIServer() (pbcommon.ErrCode, string) {
	selector := labels.SelectorFromSet(labels.Set(map[string]string{
		constant.IP_LABEL_KEY_FOR_HOST:             a.nodeNetwork.Spec.NodeAddress,
		constant.IP_LABEL_KEY_FOR_STATUS:           constant.IP_STATUS_ACTIVE,
		constant.IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER: strconv.FormatBool(true),
	}))
	ipObjs, err := a.k8sIPClient.CloudIPs(a.req.PodNamespace).List(a.ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	for _, ipObj := range ipObjs.Items {
		if ipObj.Spec.ContainerID == a.req.ContainerID &&
			ipObj.Spec.PodName == a.req.PodName {
			a.ipObj = &ipObj
			return pbcommon.ErrCode_ERROR_OK, ""
		}
	}
	return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED,
		fmt.Sprintf("ip for pod %s/%s container %s not found", a.req.PodName, a.req.PodNamespace, a.req.ContainerID)
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
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_RELEASE_IP_FAILED,
				fmt.Sprintf("call ReleaseFixedIP failed, err %s", err.Error())
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
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_RELEASE_IP_FAILED,
			fmt.Sprintf("call ReleaseIP failed, err %s", err.Error())
	}
	if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return ipResult.ErrCode, ipResult.ErrMsg
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *ReleaseAction) deleteIPObjFromAPIServer() (pbcommon.ErrCode, string) {
	// only delete ip related to non-fixed ip
	if !a.ipObj.Spec.IsFixed {
		if err := a.k8sIPClient.CloudIPs(a.ipObj.GetNamespace()).
			Delete(a.ctx, a.ipObj.GetName(), metav1.DeleteOptions{}); err != nil {

			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED,
				fmt.Sprintf("delete ip %s/%s from api server failed, err %s",
					a.ipObj.GetNamespace(), a.ipObj.GetName(), err.Error())
		}
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	timeNow := time.Now()
	a.ipObj.Labels[constant.IP_LABEL_KEY_FOR_STATUS] = constant.IP_STATUS_AVAILABLE
	a.ipObj.Status.Status = constant.IP_STATUS_AVAILABLE
	a.ipObj.Status.UpdateTime = common.FormatTime(timeNow)
	ipObj, err := a.k8sIPClient.CloudIPs(a.ipObj.GetNamespace()).Update(a.ctx, a.ipObj, metav1.UpdateOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED,
			fmt.Sprintf("update ip %s/%s to api server failed, err %s",
				a.ipObj.GetNamespace(), a.ipObj.GetName(), err.Error())
	}
	a.ipObj = ipObj

	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do release action
func (a *ReleaseAction) Do() error {
	if errCode, errMsg := a.getNodeInfo(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
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
