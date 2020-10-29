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
	"strings"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"

	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
	pb "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-cloud-netagent/internal/inspector"
	"github.com/Tencent/bk-bcs/bcs-network/internal/constant"
	common "github.com/Tencent/bk-bcs/bcs-network/pkg/common"
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

	fixedIPWorkloadMap map[string]bool

	pod              *k8sv1.Pod
	nodeNetwork      *cloudv1.NodeNetwork
	ipFromNetService *pbcommon.IPObject
	mask             int
}

// NewAllocateAction create AllocateAction
func NewAllocateAction(ctx context.Context,
	req *pb.AllocIPReq, resp *pb.AllocIPResp,
	k8sClient k8score.CoreV1Interface,
	k8sIPClient cloudv1set.CloudV1Interface,
	cloudNetClient pbcloudnet.CloudNetserviceClient,
	inspector *inspector.NodeNetworkInspector,
	fixedIPWorkloads []string) *AllocateAction {

	fixedIPWorkloadMap := make(map[string]bool)
	for _, workload := range fixedIPWorkloads {
		fixedIPWorkloadMap[strings.ToLower(workload)] = true
	}

	action := &AllocateAction{
		req:                req,
		resp:               resp,
		ctx:                ctx,
		k8sClient:          k8sClient,
		k8sIPClient:        k8sIPClient,
		cloudNetClient:     cloudNetClient,
		inspector:          inspector,
		fixedIPWorkloadMap: fixedIPWorkloadMap,
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
	if len(a.req.PodName) == 0 {
		return errors.New("PodName cannot be empty")
	}
	if len(a.req.PodNamespace) == 0 {
		return errors.New("PodNamespace cannot be empty")
	}
	if len(a.req.ContainerID) == 0 {
		return errors.New("pod containerID cannot be empty")
	}
	return nil
}

// Input do something before Do function
func (a *AllocateAction) Input() error {
	if err := a.validate(); err != nil {
		return a.Err(pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_INVALID_PARAMS, err.Error())
	}
	return nil
}

// Output do something after Do function
func (a *AllocateAction) Output() error {
	if a.nodeNetwork != nil && a.ipFromNetService != nil {
		a.resp.IpAddr = a.ipFromNetService.Address
		a.resp.MacAddr = a.nodeNetwork.Status.FloatingIPEni.Eni.MacAddress
		a.resp.Mask = int32(a.mask)
		a.resp.Gateway = "169.254.1.1"
	}
	return nil
}

func (a *AllocateAction) getPodInfo() (pbcommon.ErrCode, string) {
	pod, err := a.k8sClient.Pods(a.req.PodNamespace).Get(a.ctx, a.req.PodName, metav1.GetOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_POD_NOT_FOUND, fmt.Sprintf("get pod failed, err %s", err.Error())
	}
	a.pod = pod
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) isNodeReadyForAllocate() (pbcommon.ErrCode, string) {
	if !a.inspector.CanAllocate() {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE,
			fmt.Sprintf("agent not ready for allocating")
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) getNodeInfo() (pbcommon.ErrCode, string) {
	nodeNetwork := a.inspector.GetNodeNetwork()
	if nodeNetwork == nil ||
		nodeNetwork.Status.FloatingIPEni == nil ||
		nodeNetwork.Status.Status != cloudv1.NodeNetworkStatusReady {

		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE, fmt.Sprintf("node eni not ready")
	}
	a.nodeNetwork = nodeNetwork
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) allocateFromCloudNetservice() (pbcommon.ErrCode, string) {
	if len(a.pod.OwnerReferences) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_POD_WORKLOAD_NOT_FOUND, fmt.Sprintf("pod no owner")
	}
	workloadRef := a.pod.OwnerReferences[0]

	annotationValue, ok := a.pod.ObjectMeta.Annotations[constant.POD_ANNOTATION_KEY_FOR_ENI]
	if ok && annotationValue == constant.POD_ANNOTATION_VALUE_FOR_FIXED_IP {
		if _, ok := a.fixedIPWorkloadMap[strings.ToLower(workloadRef.Kind)]; !ok {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_WORKLOAD_NOT_SUPPORT_FIXED_IP_FEATURE,
				fmt.Sprintf("workload %s not support fixed ip feature", workloadRef.Kind)
		}
		newReq := &pbcloudnet.AllocateFixedIPReq{
			Seq:          common.TimeSequence(),
			VpcID:        a.nodeNetwork.Spec.VM.NodeVpcID,
			Region:       a.nodeNetwork.Spec.VM.NodeRegion,
			SubnetID:     a.nodeNetwork.Status.FloatingIPEni.Eni.EniSubnetID,
			Cluster:      a.nodeNetwork.Spec.Cluster,
			Namespace:    a.pod.GetNamespace(),
			PodName:      a.pod.GetName(),
			WorkloadName: workloadRef.Name,
			WorkloadKind: workloadRef.Kind,
			ContainerID:  a.req.ContainerID,
			Host:         a.nodeNetwork.Spec.NodeAddress,
			EniID:        a.nodeNetwork.Status.FloatingIPEni.Eni.EniID,
		}
		requestIP, ok := a.pod.ObjectMeta.Annotations[constant.POD_ANNOTATION_KEY_FOR_ENI_REQUEST_IP]
		if ok {
			newReq.Address = requestIP
		}
		ipResult, err := a.cloudNetClient.AllocateFixedIP(a.ctx, newReq)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED,
				fmt.Sprintf("call AllocateFixedIP failed, err %s", err.Error())
		}
		if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
			return ipResult.ErrCode, ipResult.ErrMsg
		}
		a.ipFromNetService = ipResult.Ip
		return pbcommon.ErrCode_ERROR_OK, ""
	}

	newReq := &pbcloudnet.AllocateIPReq{
		Seq:          common.TimeSequence(),
		VpcID:        a.nodeNetwork.Spec.VM.NodeVpcID,
		Region:       a.nodeNetwork.Spec.VM.NodeRegion,
		SubnetID:     a.nodeNetwork.Status.FloatingIPEni.Eni.EniSubnetID,
		Cluster:      a.nodeNetwork.Spec.Cluster,
		Namespace:    a.pod.GetNamespace(),
		PodName:      a.pod.GetName(),
		WorkloadName: workloadRef.Name,
		WorkloadKind: workloadRef.Kind,
		ContainerID:  a.req.ContainerID,
		Host:         a.nodeNetwork.Spec.NodeAddress,
		EniID:        a.nodeNetwork.Status.FloatingIPEni.Eni.EniID,
	}

	ipResult, err := a.cloudNetClient.AllocateIP(a.ctx, newReq)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED,
			fmt.Sprintf("call AllocateIP failed, err %s", err.Error())
	}
	if ipResult.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return ipResult.ErrCode, ipResult.ErrMsg
	}
	a.ipFromNetService = ipResult.Ip
	return pbcommon.ErrCode_ERROR_OK, ""
}

// record IP object to cluster k8s apiserver, for cleaning fixed ips and scheduler
func (a *AllocateAction) storeIPObjectToAPIServer() (pbcommon.ErrCode, string) {
	ipObj, err := a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).
		Get(a.ctx, a.ipFromNetService.Address, metav1.GetOptions{})
	timeNow := time.Now()
	if err != nil {
		if k8serrors.IsNotFound(err) {
			newIPObj := &cloudv1.CloudIP{
				TypeMeta: metav1.TypeMeta{
					Kind:       constant.CLOUD_CRD_NAME_CLOUD_IP,
					APIVersion: constant.CLOUD_CRD_VERSION_V1,
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      a.ipFromNetService.Address,
					Namespace: a.ipFromNetService.Namespace,
					Labels: map[string]string{
						constant.IP_LABEL_KEY_FOR_WORKLOAD_KIND:    a.ipFromNetService.WorkloadKind,
						constant.IP_LABEL_KEY_FOR_HOST:             a.ipFromNetService.Host,
						constant.IP_LABEL_KEY_FOR_IS_FIXED:         strconv.FormatBool(a.ipFromNetService.IsFixed),
						constant.IP_LABEL_KEY_FOR_STATUS:           constant.IP_STATUS_ACTIVE,
						constant.IP_LABEL_KEY_FOR_IS_CLUSTER_LAYER: strconv.FormatBool(true),
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
					WorkloadKind: a.ipFromNetService.WorkloadKind,
					ContainerID:  a.ipFromNetService.ContainerID,
					Host:         a.ipFromNetService.Host,
					EniID:        a.ipFromNetService.EniID,
					IsFixed:      a.ipFromNetService.IsFixed,
				},
				Status: cloudv1.CloudIPStatus{
					Status:     constant.IP_STATUS_ACTIVE,
					CreateTime: common.FormatTime(timeNow),
					UpdateTime: common.FormatTime(timeNow),
				},
			}
			_, inErr := a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).
				Create(a.ctx, newIPObj, metav1.CreateOptions{})
			if inErr != nil {
				return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, inErr.Error()
			}
			return pbcommon.ErrCode_ERROR_OK, ""
		}
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}

	ipObj.Labels[constant.IP_LABEL_KEY_FOR_WORKLOAD_KIND] = a.ipFromNetService.WorkloadKind
	ipObj.Labels[constant.IP_LABEL_KEY_FOR_HOST] = a.ipFromNetService.Host
	ipObj.Labels[constant.IP_LABEL_KEY_FOR_IS_FIXED] = strconv.FormatBool(a.ipFromNetService.IsFixed)
	ipObj.Labels[constant.IP_LABEL_KEY_FOR_STATUS] = constant.IP_STATUS_ACTIVE
	ipObj.Spec.Address = a.ipFromNetService.Address
	ipObj.Spec.VpcID = a.ipFromNetService.VpcID
	ipObj.Spec.Region = a.ipFromNetService.Region
	ipObj.Spec.SubnetID = a.ipFromNetService.SubnetID
	ipObj.Spec.SubnetCidr = a.ipFromNetService.SubnetCidr
	ipObj.Spec.Cluster = a.ipFromNetService.Cluster
	ipObj.Spec.Namespace = a.ipFromNetService.Namespace
	ipObj.Spec.PodName = a.ipFromNetService.PodName
	ipObj.Spec.WorkloadName = a.ipFromNetService.WorkloadName
	ipObj.Spec.WorkloadKind = a.ipFromNetService.WorkloadKind
	ipObj.Spec.ContainerID = a.ipFromNetService.ContainerID
	ipObj.Spec.IsFixed = a.ipFromNetService.IsFixed
	ipObj.Spec.Host = a.ipFromNetService.Host
	ipObj.Spec.EniID = a.ipFromNetService.EniID
	ipObj.Status.Status = constant.IP_STATUS_ACTIVE
	ipObj.Status.UpdateTime = common.FormatTime(timeNow)
	_, err = a.k8sIPClient.CloudIPs(a.ipFromNetService.Namespace).Update(a.ctx, ipObj, metav1.UpdateOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	return pbcommon.ErrCode_ERROR_OK, ""
}

// parseIP
func (a *AllocateAction) parseIP() (pbcommon.ErrCode, string) {
	_, mask, err := common.ParseCIDR(a.ipFromNetService.SubnetCidr)
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_INVALID_IP_INFO, err.Error()
	}
	a.mask = mask
	return pbcommon.ErrCode_ERROR_OK, ""
}

// Do do allocate action
func (a *AllocateAction) Do() error {
	if errCode, errMsg := a.getPodInfo(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.getNodeInfo(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.allocateFromCloudNetservice(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.parseIP(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.storeIPObjectToAPIServer(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
