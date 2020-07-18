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

package controllers

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/cloud/v1"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-services/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller/internal/option"
	cloudAPI "github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller/pkg/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-network/bcs-cloud-netcontroller/pkg/constant"
)

// Processor node network processor
type Processor struct {
	needDo  bool
	isDoing bool

	eventChan  chan struct{}
	kubeClient client.Client

	option *option.ControllerOption

	cloudNetClient pbcloudnet.CloudNetserviceClient

	cloudClient cloudAPI.Interface

	nodeEventer record.EventRecorder
}

// NewProcessor create event processor
func NewProcessor(r client.Client,
	option *option.ControllerOption,
	cloudNetClient pbcloudnet.CloudNetserviceClient,
	cloudClient cloudAPI.Interface,
	nodeEventer record.EventRecorder) *Processor {

	return &Processor{
		kubeClient:     r,
		option:         option,
		cloudNetClient: cloudNetClient,
		cloudClient:    cloudClient,
		nodeEventer:    nodeEventer,
		eventChan:      make(chan struct{}, 10),
	}
}

// Run run processor
func (p *Processor) Run(ctx context.Context) error {
	timer := time.NewTicker(10 * time.Second)

	for {
		select {
		case <-timer.C:
			if !p.isDoing && p.needDo {
				p.needDo = false
				p.isDoing = true
				if err := p.handle(); err != nil {
					blog.Error("handle node change failed, err %s", err.Error())
					p.needDo = true
				}
				p.isDoing = false
			}

		case <-p.eventChan:
			p.needDo = true

		case <-ctx.Done():
			blog.Infof("processor context done")
			return nil
		}
	}
}

// OnEvent send event ring to processor
func (p *Processor) OnEvent() {
	p.eventChan <- struct{}{}
}

func (p *Processor) handle() error {
	var nodes corev1.NodeList
	labelSelector := labels.SelectorFromSet(labels.Set(map[string]string{"nodenetwork.bkbcs.tencent.com": "true"}))
	if err := p.kubeClient.List(context.TODO(), &nodes, &client.ListOptions{LabelSelector: labelSelector}); err != nil {
		blog.Errorf("unable to list Nodes, err %s", err.Error())
		return fmt.Errorf("unable to list Nodes, err %s", err.Error())
	}
	nodeNameMap := make(map[string]*corev1.Node)
	for _, node := range nodes.Items {
		nodeNameMap[node.GetName()] = &node
	}

	var nodeNetworks cloudv1.NodeNetworkList
	if err := p.kubeClient.List(context.TODO(), &nodeNetworks, &client.ListOptions{}); err != nil {
		blog.Errorf("unable to list NodeNetworks, err %s", err.Error())
		return fmt.Errorf("unable to list NodeNetworks, err %s", err.Error())
	}
	nodeNetworkMap := make(map[string]*cloudv1.NodeNetwork)
	for _, nodeNet := range nodeNetworks.Items {
		nodeNetworkMap[nodeNet.GetName()] = &nodeNet
	}

	// deal with new node
	for nodeName, node := range nodeNameMap {
		if _, ok := nodeNetworkMap[nodeName]; !ok {
			if err := p.addNodeNetwork(node); err != nil {
				return err
			}
		}
	}

	// TODO: deal with the re-created node

	// deal with deleted node network
	for nodeName, nodenetwork := range nodeNetworkMap {
		if _, ok := nodeNameMap[nodeName]; !ok {
			if err := p.deleteNodeNetwork(nodenetwork); err != nil {
				return err
			}
		}
	}

	return nil
}

// get available subnet from cloud netservice
func (p *Processor) getAvailableSubnet(nodeVMInfo *cloudv1.VMInfo) (string, error) {
	req := &pbcloudnet.GetAvailableSubnetReq{
		Seq:    common.TimeSequence(),
		VpcID:  nodeVMInfo.NodeVpcID,
		Region: nodeVMInfo.NodeRegion,
		Zone:   nodeVMInfo.NodeZone,
	}
	resp, err := p.cloudNetClient.GetAvailableSubnet(context.TODO(), req)
	if err != nil {
		return "", err
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return "", fmt.Errorf("get available subnet failed, err %s", resp.ErrMsg)
	}
	if resp.Subnet == nil {
		return "", fmt.Errorf("get available return empty subent")
	}
	return resp.Subnet.SubnetID, nil
}

// add new node network
func (p *Processor) addNodeNetwork(node *corev1.Node) error {
	// get vm info for node
	nodeVMInfo, err := p.cloudClient.GetVMInfo(node.Status.Addresses[0].Address)
	if err != nil {
		return err
	}

	// get available subnet used for creating eni
	var subnetID string
	if reqSubnetID, ok := node.ObjectMeta.Labels["nodenetwork.bkbcs.tencent.com/subnetId"]; ok {
		subnetID = reqSubnetID
	} else {
		subnetID, err = p.getAvailableSubnet(nodeVMInfo)
		if err != nil {
			return err
		}
	}

	// create new node network object
	newNodeNetwork := &cloudv1.NodeNetwork{
		TypeMeta: k8smetav1.TypeMeta{
			APIVersion: cloudv1.SchemeGroupVersion.Version,
		},
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      node.GetName(),
			Namespace: constant.CRD_NAMESPACES,
		},
		Spec: cloudv1.NodeNetworkSpec{
			Cluster:     p.option.Cluster,
			Hostname:    node.GetName(),
			NodeAddress: nodeVMInfo.InstanceIP,
			VM:          nodeVMInfo,
		},
	}
	newNodeNetwork.Finalizers = append(newNodeNetwork.Finalizers, constant.FINALIZER_NAME)

	eniCrdObj, err := p.reconcileEniForDynamic(nodeVMInfo, subnetID)
	if err != nil {
		return err
	}

	_, ipLimit, err := p.cloudClient.GetENILimit(nodeVMInfo.InstanceIP)
	if err != nil {
		return err
	}

	newNodeNetwork.Status.FloatingIPEni = &cloudv1.FloatingIPNetworkInterface{
		Eni:     eniCrdObj,
		IPLimit: ipLimit - 1,
	}

	if err := p.kubeClient.Create(context.TODO(), newNodeNetwork, &client.CreateOptions{}); err != nil {
		return err
	}

	//p.nodeEventer.Eventf(node, corev1.EventTypeNormal, "nodenetwork created", "eni info: %+v", newNodeNetwork)

	return nil
}

// delete node network
func (p *Processor) deleteNodeNetwork(nodenetwork *cloudv1.NodeNetwork) error {
	if nodenetwork.DeletionTimestamp.IsZero() {
		// pre delete
		if err := p.kubeClient.Delete(context.TODO(), nodenetwork, &client.DeleteOptions{}); err != nil {
			return err
		}
		return nil
	}
	if containsString(nodenetwork.Finalizers, constant.NODE_FINALIZER_NAME) {
		blog.Warnf("wait for agent to clean its finalizer")
	}
	if containsString(nodenetwork.Finalizers, constant.FINALIZER_NAME) {
		// release eni
		if nodenetwork.Status.FloatingIPEni != nil {
			if len(nodenetwork.Status.FloatingIPEni.FloatingIPs) != 0 {
				blog.Errorf("there is active ip in node %s, cannot delete", nodenetwork.GetName())
				// TODO: set node event to warn user
				return nil
			}
			fEni := nodenetwork.Status.FloatingIPEni
			err := p.cloudClient.DetachENI(fEni.Eni.Attachment)
			if err != nil {
				blog.Errorf("detach eni failed, err %s", err.Error())
				return nil
			}
			err = p.cloudClient.DeleteENI(fEni.Eni.EniID)
			if err != nil {
				blog.Errorf("delete eni failed, err %s", err.Error())
				return nil
			}
		}
		// real delete
		nodenetwork.Finalizers = removeString(nodenetwork.Finalizers, constant.FINALIZER_NAME)
		if err := p.kubeClient.Update(context.TODO(), nodenetwork, &client.UpdateOptions{}); err != nil {
			return fmt.Errorf("delete finalizers of %s failed, err %s", nodenetwork.GetName(), err.Error())
		}
	}

	return nil
}

func (p *Processor) reconcileEniForPreAllocate() error {
	return nil
}

func (p *Processor) reconcileEniForDynamic(nodeVMInfo *cloudv1.VMInfo, subnetID string) (*cloudv1.ElasticNetworkInterface, error) {
	eniCrdObj, err := p.cloudClient.CreateENI(generateEniName(nodeVMInfo.InstanceID, 99), subnetID, 0)
	if err != nil {
		return nil, err
	}
	if eniCrdObj.Attachment == nil {
		maxIndex, err := p.cloudClient.GetMaxENIIndex(nodeVMInfo.InstanceIP)
		if err != nil {
			return nil, err
		}
		attachment, err := p.cloudClient.AttachENI(
			maxIndex+1,
			eniCrdObj.EniID,
			nodeVMInfo.InstanceID,
			eniCrdObj.MacAddress)
		if err != nil {
			return nil, err
		}
		eniCrdObj.Attachment = attachment
	}
	eniCrdObj.Index = constant.ENI_FLOATING_IP_INDEX
	eniCrdObj.EniIfaceName = getEniIfaceName(constant.ENI_FLOATING_IP_INDEX)
	eniCrdObj.RouteTableID = constant.START_ROUTE_TABLE + constant.ENI_FLOATING_IP_INDEX
	return eniCrdObj, nil
}
