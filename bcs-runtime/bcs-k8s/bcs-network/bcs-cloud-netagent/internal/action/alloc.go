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
	"strconv"
	"strings"
	"time"

	k8sv1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8score "k8s.io/client-go/kubernetes/typed/core/v1"

	pb "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetagent"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netagent/internal/inspector"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	common "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"
	cloudv1set "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/generated/clientset/versioned/typed/cloud/v1"
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

	pod                     *k8sv1.Pod
	nodeNetwork             *cloudv1.NodeNetwork
	existedIPFromNetService *pbcommon.IPObject
	allocatableEniID        string
	allocatableEniMacAddr   string
	allocatableSubnetID     string
	ipFromNetService        *pbcommon.IPObject
	mask                    int
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
		a.resp.MacAddr = a.allocatableEniMacAddr
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

func (a *AllocateAction) getFixedIPObjectFromCloudNetservice() error {
	resp, err := a.cloudNetClient.ListIP(a.ctx, &pbcloudnet.ListIPsReq{
		PodName:   a.pod.GetName(),
		Namespace: a.pod.GetNamespace(),
		Cluster:   a.nodeNetwork.Spec.Cluster,
		Status:    constant.IPStatusAvailable,
	})
	if err != nil {
		return fmt.Errorf("list cloud netservice ip, pod %s, ns %s failed, err %s",
			a.pod.GetName(), a.pod.GetNamespace(), err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return fmt.Errorf("list cloud netservice ip, pod %s, ns %s failed, errCode %s, errMsg %s",
			a.pod.GetName(), a.pod.GetNamespace(), resp.ErrCode, resp.ErrMsg)
	}
	if len(resp.Ips) == 0 {
		return nil
	}
	if len(resp.Ips) != 1 {
		return fmt.Errorf("list cloud netservice ip for pod %s, ns %s return more than 1 ip",
			a.pod.GetName(), a.pod.GetNamespace())
	}
	ipObj := resp.Ips[0]
	if !ipObj.IsFixed {
		return fmt.Errorf("list cloud netservice ip for pod %s, ns %s is not fixed",
			a.pod.GetName(), a.pod.GetNamespace())
	}
	a.existedIPFromNetService = ipObj
	return nil
}

func (a *AllocateAction) getAllocatableEni() (pbcommon.ErrCode, string) {
	annotationValue, ok := a.pod.ObjectMeta.Annotations[constant.PodAnnotationKeyForEni]
	// if request fixed ip, list fixed ip for pod from cloud netservice
	if ok && annotationValue == constant.PodAnnotationValueForFixedIP {
		if err := a.getFixedIPObjectFromCloudNetservice(); err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED, err.Error()
		}
	}
	for _, eni := range a.nodeNetwork.Status.Enis {
		if eni.Status != constant.NodeNetworkEniStatusReady {
			continue
		}
		if len(a.inspector.GetIPCache().ListEniIP(eni.EniID)) >= a.nodeNetwork.Spec.IPNumPerENI {
			continue
		}
		// if found fixed ip for pod, select eni with fixed ip subnet id
		if a.existedIPFromNetService != nil {
			if eni.EniSubnetID == a.existedIPFromNetService.SubnetID {
				a.allocatableEniID = eni.EniID
				a.allocatableEniMacAddr = eni.MacAddress
				a.allocatableSubnetID = eni.EniSubnetID
				return pbcommon.ErrCode_ERROR_OK, ""
			}
			continue
		}
		a.allocatableEniID = eni.EniID
		a.allocatableEniMacAddr = eni.MacAddress
		a.allocatableSubnetID = eni.EniSubnetID
		return pbcommon.ErrCode_ERROR_OK, ""
	}
	return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_ALLOCATE_IP_FAILED, "no suitable eni found"
}

func (a *AllocateAction) getNodeInfo() (pbcommon.ErrCode, string) {
	nodeNetwork := a.inspector.GetNodeNetwork()
	if nodeNetwork == nil || len(nodeNetwork.Status.Enis) == 0 {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_NODENETWORK_NOT_AVAILABLE, fmt.Sprintf("node eni not ready")
	}
	a.nodeNetwork = nodeNetwork
	return pbcommon.ErrCode_ERROR_OK, ""
}

func (a *AllocateAction) allocateFromCloudNetservice() (pbcommon.ErrCode, string) {
	isFixed := false
	annotationValue, ok1 := a.pod.ObjectMeta.Annotations[constant.PodAnnotationKeyForEni]
	if ok1 && annotationValue == constant.PodAnnotationValueForFixedIP {
		if len(a.pod.OwnerReferences) == 0 {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_POD_WORKLOAD_NOT_FOUND,
				fmt.Sprintf("pod no owner for fixed ip")
		}
		workloadRef := a.pod.OwnerReferences[0]
		if _, ok2 := a.fixedIPWorkloadMap[strings.ToLower(workloadRef.Kind)]; !ok2 {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_WORKLOAD_NOT_SUPPORT_FIXED_IP_FEATURE,
				fmt.Sprintf("workload %s not support fixed ip feature", workloadRef.Kind)
		}
		isFixed = true
	}

	keepDuration := constant.DefaultFixedIPKeepDurationStr
	durationStr, ok3 := a.pod.ObjectMeta.Annotations[constant.PodAnnotationKeyForFixedIPKeepDuration]
	if ok3 {
		duration, err := time.ParseDuration(durationStr)
		if err != nil {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_INVALID_FIXED_IP_KEEP_DURATION,
				fmt.Sprintf("invalid keep duration %s", durationStr)
		}
		if duration > constant.MaxFixedIPKeepDuration {
			return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_INVALID_FIXED_IP_KEEP_DURATION,
				fmt.Sprintf("keep duration %s is bigger than %f hours",
					durationStr, constant.MaxFixedIPKeepDuration.Hours())
		}
		keepDuration = durationStr
	}

	newReq := &pbcloudnet.AllocateIPReq{
		Seq:          common.TimeSequence(),
		SubnetID:     a.allocatableSubnetID,
		Cluster:      a.nodeNetwork.Spec.Cluster,
		Namespace:    a.pod.GetNamespace(),
		PodName:      a.pod.GetName(),
		ContainerID:  a.req.ContainerID,
		Host:         a.nodeNetwork.Spec.NodeAddress,
		EniID:        a.allocatableEniID,
		IsFixed:      isFixed,
		KeepDuration: keepDuration,
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
func (a *AllocateAction) storeIPObjectCache() (pbcommon.ErrCode, string) {
	timeNow := time.Now()
	_, err := a.k8sIPClient.CloudIPs(a.req.PodNamespace).Create(a.ctx, &cloudv1.CloudIP{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CloudIP",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      a.req.PodName,
			Namespace: a.req.PodNamespace,
			Labels: map[string]string{
				constant.IPLabelKeyForENI:          a.ipFromNetService.EniID,
				constant.IPLabelKeyForIsFixedKey:   strconv.FormatBool(a.ipFromNetService.IsFixed),
				constant.IPLabelKeyForClusterLayer: strconv.FormatBool(true),
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
			CreateTime: timeNow.String(),
			UpdateTime: timeNow.String(),
		},
	}, metav1.CreateOptions{})
	if err != nil {
		return pbcommon.ErrCode_ERROR_CLOUD_NETAGENT_K8S_API_SERVER_OPS_FAILED, err.Error()
	}
	a.inspector.GetIPCache().PutEniIP(a.ipFromNetService.EniID, a.ipFromNetService)
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
	// lock for cocurrent allocate action
	a.inspector.Lock()
	defer a.inspector.Unlock()
	if errCode, errMsg := a.getAllocatableEni(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.allocateFromCloudNetservice(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.parseIP(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	if errCode, errMsg := a.storeIPObjectCache(); errCode != pbcommon.ErrCode_ERROR_OK {
		return a.Err(errCode, errMsg)
	}
	return nil
}
