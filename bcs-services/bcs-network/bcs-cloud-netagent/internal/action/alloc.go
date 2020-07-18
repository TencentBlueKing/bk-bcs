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

	k8sv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"

	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	bcsclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	pb "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/internal/inspector"
	common "github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netagent/pkg/common"
)

// AllocateAction action for allocate ip
type AllocateAction struct {
	req  *pb.AllocIPReq
	resp *pb.AllocIPResp

	ctx context.Context

	k8sClient k8score.CoreV1Interface

	k8sIPClient cloudv1set.CloudV1Interface

	cloudNetClient pbcloudnet.CloudNetserviceClient

	inspector *inspector.NodeNetworkInspector

	pod              *k8sv1.Pod
	ipFromNetService *pbcommon.IPObject
}

// NewAllocateAction create AllocateAction
func NewAllocateAction(ctx context.Context,
	req *pb.AllocIPReq, resp *pb.AllocIPResp,
	k8sClient k8score.CoreV1Interface,
	cloudNetClient pbcloudnet.CloudNetserviceClient,
	inspector *inspector.NodeNetworkInspector) *AllocateAction {

	action := &AllocateAction{
		req:            req,
		resp:           resp,
		ctx:            ctx,
		k8sClient:      k8sClient,
		cloudNetClient: cloudNetClient,
		inspector:      inspector,
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
	if len(a.req.Podname) == 0 {
		return errors.New("Podname cannot be empty")
	}
	if len(a.req.Podnamespace) == 0 {
		return errors.New("Podnamespace cannot be empty")
	}
	if len(a.req.ContainerID) == 0 {
		return errors.New("pod containerID cannot be empty")
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

	return nil
}

func (a *AllocateAction) getPodInfo() (pbcommon.ErrCode, string) {
	pod, err := a.k8sClient.Pods(a.req.Podnamespace).Get(a.ctx, a.req.Podname, metav1.GetOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_POD_NOT_FOUND, fmt.Sprintf("get pod failed, err %s", err.Error())
	}
	a.pod = pod
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) allocateFromCloudNetservice() (pbcommon.ErrCode, string) {
	nodeNetwork := a.inspector.GetNodeNetwork()
	if nodeNetwork == nil || nodeNetwork.Status.FloatingIPEni == nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE, fmt.Sprintf("node eni not ready")
	}

	if len(a.pod.OwnerReferences) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_POD_WORKLOAD_NOT_FOUND, fmt.Sprintf("pod no owner")
	}
	workloadRef := a.pod.OwnerReferences[0]

	annotationValue, ok := a.pod.ObjectMeta.Annotations[constant.ANNOTATION_POD_ENI_KEY]
	if annotationValue == constant.ANNOTATION_POD_ENI_VALUE_FIXED {
		newReq := &pbcloudnet.AllocateFixedIPReq{
			Seq:          common.TimeSequence(),
			VpcID:        nodeNetwork.Spec.VM.NodeVpcID,
			Region:       nodeNetwork.Spec.VM.NodeRegion,
			SubnetID:     nodeNetwork.Status.FloatingIPEni.Eni.EniSubnetID,
			Cluster:      nodeNetwork.ClusterName,
			Namespace:    a.pod.GetNamespace(),
			PodName:      a.pod.GetName(),
			WorkloadName: workloadRef.Name,
			WorkloadKind: workloadRef.Kind,
			ContainerID:  a.req.ContainerID,
			Host:         nodeNetwork.Spec.NodeAddress,
			EniID:        nodeNetwork.Status.FloatingIPEni.Eni.EniID,
		}
		requestIP, ok := a.pod.ObjectMeta.Annotations[constant.ANNOTATION_POD_REQUESTIP_KEY]
		if ok {
			newReq.Address = requestIP
		}
		ipResult, err := a.cloudNetClient.AllocateFixedIP(a.ctx, newReq)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED, fmt.Sprintf("call AllocateFixedIP failed, err %s", err.Error())
		}
		if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
			return ipResult.ErrCode, ipResult.ErrMsg
		}
		a.ipFromNetService = ipResult.Ip
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	newReq := &pbcloudnet.AllocateIPReq{
		Seq:          common.TimeSequence(),
		VpcID:        nodeNetwork.Spec.VM.NodeVpcID,
		Region:       nodeNetwork.Spec.VM.NodeRegion,
		SubnetID:     nodeNetwork.Status.FloatingIPEni.Eni.EniSubnetID,
		Cluster:      nodeNetwork.ClusterName,
		Namespace:    a.pod.GetNamespace(),
		PodName:      a.pod.GetName(),
		WorkloadName: workloadRef.Name,
		WorkloadKind: workloadRef.Kind,
		ContainerID:  a.req.ContainerID,
		Host:         nodeNetwork.Spec.NodeAddress,
		EniID:        nodeNetwork.Status.FloatingIPEni.Eni.EniID,
	}

	ipResult, err := a.cloudNetClient.AllocateIP(a.ctx, newReq)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED, fmt.Sprintf("call AllocateIP failed, err %s", err.Error())
	}
	if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return ipResult.ErrCode, ipResult.ErrMsg
	}
	a.ipFromNetService = ipResult.Ip
	return pbcommon.ErrCode_ERROR_OK, ""
}

// record IP object to cluster k8s apiserver, for cleaning fixed ips and scheduler
func (a *AllocateAction) storeIPObjectToAPIServer() (pbcommon.ErrCode, string) {
	ipObj, err := a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).Get(a.ctx, a.ipFromNetService.Address, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			newIPObj := &cloudv1.CloudIP{
				TypeMeta: metav1.TypeMeta{
					Kind:       constant.CRD_NAME_CLOUD_IP,
					APIVersion: constant.CRD_VERSION_V1,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      a.ipFromNetService.Address,
					Namespace: a.ipFromNetService.Namespace,
					Labels: map[string]string{
						constant.ANNOTATION_CLOUD_IP_HOST: a.ipFromNetService.Host,
					},
				},
				Spec: cloudv1.CloudIPSpec{
					Address:      a.ipFromNetService.Address,
					VpcID:        a.ipFromNetService.VpcID,
					Region:       a.ipFromNetService.Region,
					SubnetID:     a.ipFromNetService.SubnetID,
					SubnetCidr:   a.ipFromNetService.SubnetCidr,
					Cluster:      a.ipFromNetService.Cluster,
					Namespace:    a.ipFromNetService.Namespace,
					PodName:      a.ipFromNetService.PodName,
					WorkloadName: a.ipFromNetService.WorkloadName,
					ContainerID:  a.ipFromNetService.ContainerID,
					Host:         a.ipFromNetService.Host,
					EniID:        a.ipFromNetService.EniID,
					IsFixed:      a.ipFromNetService.IsFixed,
				},
				Status: cloudv1.CloudIPStatus{
					Status:     a.ipFromNetService.Status,
					CreateTime: a.ipFromNetService.CreateTime,
					UpdateTime: a.ipFromNetService.UpdateTime,
				},
			}
			_, err := a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).Create(a.ctx, newIPObj, metav1.CreateOptions{})
			if err != nil {
				pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
			}
			return pbcommon.ErrCode_ERROR_OK, ""
		}
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	ipObj.Spec.Host = a.ipFromNetService.Host
	ipObj.Spec.EniID = a.ipFromNetService.EniID
	ipObj.Status.Status = a.ipFromNetService.Status
	ipObj.Status.CreateTime = a.ipFromNetService.CreateTime
	ipObj.Status.UpdateTime = a.ipFromNetService.UpdateTime
	_, err := a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).Update(a.ctx, ipObj, metav1.UpdateOptions{})
	if err != nil {
		pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *AllocateAction) Do() error {
	if errCode, errMsg := a.getPodInfo(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.allocateFromCloudNetservice(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.storeIPObjectToAPIServer(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
