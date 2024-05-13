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

package controllers

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	pbcloudnet "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/cloudnetservice"
	pbcommon "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/api/protocol/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/internal/option"
	cloudAPI "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-cloud-netcontroller/pkg/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/pkg/common"
	cloudv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/cloud/v1"

	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// NodeNetworkEvent node network event
type NodeNetworkEvent struct {
	NodeName      string
	NodeNamespace string
	DelaySecond   int
}

// Processor node network processor
type Processor struct {
	// needDo  bool
	// isDoing bool

	eventChan  chan NodeNetworkEvent
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
		eventChan:      make(chan NodeNetworkEvent, 1000),
	}
}

// Run run processor
func (p *Processor) Run(ctx context.Context) error {
	for {
		select {
		case e := <-p.eventChan:
			if e.DelaySecond != 0 {
				blog.Warnf("event %v failed before, have a rest", e)
				time.Sleep(time.Duration(e.DelaySecond) * time.Second)
			}
			if err := p.handle(k8stypes.NamespacedName{
				Name:      e.NodeName,
				Namespace: e.NodeNamespace,
			}); err != nil {
				blog.Warnf("handle event %s/%s failed, err %s", e.NodeName, e.NodeNamespace, err.Error())
				// failed event should be delayed
				e.DelaySecond = 3
				p.eventChan <- e
				continue
			}

		case <-ctx.Done():
			blog.Infof("processor context done")
			return nil
		}
	}
}

// OnEvent send event ring to processor
func (p *Processor) OnEvent(e NodeNetworkEvent) {
	p.eventChan <- e
}

func (p *Processor) handle(nodeNamespaceName k8stypes.NamespacedName) error {
	// get node network
	tmpNodeNetwork := &cloudv1.NodeNetwork{}
	err := p.kubeClient.Get(context.Background(), nodeNamespaceName, tmpNodeNetwork)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			blog.Warnf("node network %s not found, do nothing", nodeNamespaceName.String())
			return nil
		}
		return fmt.Errorf("get node network failed, err %s", err.Error())
	}

	// not deleted
	if tmpNodeNetwork.DeletionTimestamp.IsZero() {
		// do update
		if err := p.updateNodeNetwork(tmpNodeNetwork); err != nil {
			p.nodeEventer.Eventf(tmpNodeNetwork, corev1.EventTypeWarning,
				constant.NetControllerEventReasonUpdataNodeNetworkFailed,
				"update node network failed, err %s", err.Error())
			return fmt.Errorf("update node network failed, err %s", err.Error())
		}
		return nil
	}

	// handle deleted node network
	if err := p.deleteNodeNetwork(tmpNodeNetwork); err != nil {
		p.nodeEventer.Eventf(tmpNodeNetwork, corev1.EventTypeWarning,
			constant.NetControllerEventReasonUpdataNodeNetworkFailed,
			"delete node network failed, err %s", err.Error())
		return fmt.Errorf("delete node network failed, err %s", err.Error())
	}
	return nil
}

// update node network
func (p *Processor) updateNodeNetwork(nodeNetwork *cloudv1.NodeNetwork) error {
	// add finalizer if finalizer is not found
	if !containsString(nodeNetwork.Finalizers, constant.FinalizerNameForNetController) {
		nodeNetwork.Finalizers = append(nodeNetwork.Finalizers, constant.FinalizerNameForNetController)
	}
	index := len(nodeNetwork.Status.Enis)
	if index != 0 {
		lastEniObj := nodeNetwork.Status.Enis[index-1]
		if lastEniObj.Status == constant.NodeNetworkEniStatusInitializing ||
			lastEniObj.Status == constant.NodeNetworkEniStatusCleaning {
			blog.Infof("eni %s is in status %s, wait", lastEniObj.EniName, lastEniObj.Status)
			return nil
		}
	}
	if nodeNetwork.Spec.ENINum >= len(nodeNetwork.Status.Enis) {
		if index != 0 {
			lastEniObj := nodeNetwork.Status.Enis[index-1]
			switch lastEniObj.Status {
			case constant.NodeNetworkEniStatusCleaned:
				lastEniObj.Status = constant.NodeNetworkEniStatusNotReady
				if err := p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
					return fmt.Errorf("mark eni %s to status %s failed, err %s",
						lastEniObj.EniName, lastEniObj.Status, err.Error())
				}
				blog.Infof("update last eni %s is to status %s successfully", lastEniObj.EniName, lastEniObj.Status)
				return nil
			case constant.NodeNetworkEniStatusDeleting:
				if err := p.delEni(nodeNetwork, index-1, false); err != nil {
					return err
				}
				p.nodeEventer.Eventf(
					nodeNetwork, corev1.EventTypeNormal, constant.NetControllerEventReasonDelEniSuccess,
					"last eni %s with addr %s of subnet %s deleted",
					lastEniObj.EniID, lastEniObj.Address.IP, lastEniObj.EniSubnetID)
				return nil
			case constant.NodeNetworkEniStatusNotReady:
				blog.Infof("last eni %s is in status %s, wait", lastEniObj.EniName, lastEniObj.Status)
				return nil
			}
		}
		if nodeNetwork.Spec.ENINum > len(nodeNetwork.Status.Enis) {
			// if last eni status is ready, begin add eni
			blog.Infof("add new eni %d for node network %s/%s",
				index, nodeNetwork.GetName(), nodeNetwork.GetNamespace())
			newEniObj, err := p.addNewEni(nodeNetwork, index)
			if err != nil {
				return err
			}
			nodeNetwork.Status.Enis = append(nodeNetwork.Status.Enis, newEniObj)
			if err = p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
				return fmt.Errorf("update nodenetwork %s/%s status failed, err %s",
					nodeNetwork.GetName(), nodeNetwork.GetNamespace(), err.Error())
			}
			p.nodeEventer.Eventf(nodeNetwork, corev1.EventTypeNormal, constant.NetControllerEventReasonAddEniSuccess,
				"eni %s with addr %s of subnet %s added", newEniObj.EniID, newEniObj.Address.IP, newEniObj.EniSubnetID)
		}

	} else if nodeNetwork.Spec.ENINum < len(nodeNetwork.Status.Enis) {
		lastIndex := len(nodeNetwork.Status.Enis) - 1
		eniObj := nodeNetwork.Status.Enis[lastIndex]
		switch eniObj.Status {
		case constant.NodeNetworkEniStatusNotReady, constant.NodeNetworkEniStatusCleaned:
			eniObj.Status = constant.NodeNetworkEniStatusDeleting
			if err := p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
				return fmt.Errorf("mark eni %s to status %s failed, err %s",
					eniObj.EniName, eniObj.Status, err.Error())
			}
			blog.Infof("update eni %s is to status %s successfully", eniObj.EniName, eniObj.Status)
			return nil

		case constant.NodeNetworkEniStatusReady:
			eniObj.Status = constant.NodeNetworkEniStatusCleaning
			if err := p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
				return fmt.Errorf("mark eni %s to status %s failed, err %s",
					eniObj.EniName, eniObj.Status, err.Error())
			}
			blog.Infof("update eni %s is to status %s successfully", eniObj.EniName, eniObj.Status)
			return nil
		}
		// if last eni is deleting, delete eni
		if err := p.delEni(nodeNetwork, lastIndex, false); err != nil {
			return err
		}
		p.nodeEventer.Eventf(nodeNetwork, corev1.EventTypeNormal, constant.NetControllerEventReasonDelEniSuccess,
			"eni %s with addr %s of subnet %s deleted", eniObj.EniID, eniObj.Address.IP, eniObj.EniSubnetID)
	}
	return nil
}

// delete node network
func (p *Processor) deleteNodeNetwork(nodeNetwork *cloudv1.NodeNetwork) error {
	if len(nodeNetwork.Status.Enis) == 0 {
		nodeNetwork.Finalizers = removeString(nodeNetwork.Finalizers, constant.FinalizerNameForNetController)
		if err := p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
			return fmt.Errorf("failed to remove finalizer %s from node %s/%s, err %s",
				constant.FinalizerNameForNetController, nodeNetwork.GetName(), nodeNetwork.GetNamespace(), err.Error())
		}
		blog.Infof("remove finalizer %s from node %s/%s successfully",
			constant.FinalizerNameForNetController, nodeNetwork.GetName(), nodeNetwork.GetNamespace())
		return nil
	}
	lastIndex := len(nodeNetwork.Status.Enis) - 1
	eniObj := nodeNetwork.Status.Enis[lastIndex]

	foundNode := true
	tmpNode := &corev1.Node{}
	if err := p.kubeClient.Get(context.Background(), k8stypes.NamespacedName{
		Namespace: "",
		Name:      nodeNetwork.GetName(),
	}, tmpNode); err != nil {
		if k8serrors.IsNotFound(err) {
			// node is deleted
			foundNode = false
		} else {
			return fmt.Errorf("failed to get node %s, err %s", tmpNode, err.Error())
		}
	}
	if foundNode {
		// delete eni
		if err := p.delEni(nodeNetwork, lastIndex, false); err != nil {
			return err
		}
	} else {
		// force delete eni
		if err := p.delEni(nodeNetwork, lastIndex, true); err != nil {
			return err
		}
	}

	p.nodeEventer.Eventf(nodeNetwork, corev1.EventTypeNormal, constant.NetControllerEventReasonDelEniSuccess,
		"eni %s with addr %s of subnet %s deleted, forced %v",
		eniObj.EniID, eniObj.Address.IP, eniObj.EniSubnetID, !foundNode)
	return nil
}

// add new eni
func (p *Processor) addNewEni(nodeNetwork *cloudv1.NodeNetwork, index int) (
	*cloudv1.ElasticNetworkInterface, error) {
	if nodeNetwork.Spec.VM == nil {
		return nil, fmt.Errorf("node network VMInfo is empty")
	}
	// allocate eni primary ip
	primaryIPObj, err := p.allocateEniPrimaryIP(
		nodeNetwork.Spec.VM.InstanceID, nodeNetwork.Spec.VM.NodeZone, uint64(index))
	if err != nil {
		return nil, err
	}
	blog.Infof("allocate eni primary ip %s for eni index %d", primaryIPObj.Address, index)

	// ensure new eni
	eniObj, err := p.reconcileEni(nodeNetwork.Spec.VM, primaryIPObj.SubnetID, primaryIPObj.Address, index)
	if err != nil {
		inErr := p.releaseEniPrimaryIP(nodeNetwork.Spec.VM.InstanceID, primaryIPObj.GetAddress(), uint64(index))
		if inErr != nil {
			blog.Warnf("release eni primary ip %s failed, err %s", primaryIPObj.GetAddress(), inErr)
		}
		return nil, err
	}
	return eniObj, nil
}

// delete eni
func (p *Processor) delEni(nodeNetwork *cloudv1.NodeNetwork, index int, isForce bool) error {
	if nodeNetwork == nil {
		blog.Errorf("node network is empty when delete eni")
		return fmt.Errorf("node network is empty when delete eni")
	}
	if len(nodeNetwork.Status.Enis) <= index {
		blog.Errorf("index %d exceed enis array length", index)
		return fmt.Errorf("index %d exceed enis array length", index)
	}
	eniObj := nodeNetwork.Status.Enis[index]

	if err := p.cleanIPOnEni(eniObj, isForce); err != nil {
		return err
	}

	foundInCloud := true
	remoteEniObj, err := p.cloudClient.QueryENI(eniObj.EniID)
	if err != nil {
		if errors.Is(err, cloudAPI.ErrEniNotFound) {
			foundInCloud = false
		} else {
			return err
		}
	}
	if foundInCloud {
		if remoteEniObj.Attachment != nil {
			err = p.cloudClient.DetachENI(eniObj.Attachment)
			if err != nil {
				return fmt.Errorf("detach eni failed, err %s", err.Error())
			}
		}
		err = p.cloudClient.DeleteENI(eniObj.EniID)
		if err != nil {
			return fmt.Errorf("delete eni failed, err %s", err.Error())
		}
	} else {
		blog.Infof("eni with id %s not found", eniObj.EniID)
	}

	// release eni primary ip to cloud netservice
	err = p.releaseEniPrimaryIP(nodeNetwork.Spec.VM.InstanceID, eniObj.Address.IP, uint64(index))
	if err != nil {
		return fmt.Errorf("delete eni primary ip failed from cloud netservice, err %s", err.Error())
	}
	nodeNetwork.Status.Enis = append(nodeNetwork.Status.Enis[0:index], nodeNetwork.Status.Enis[index+1:]...)
	if err := p.kubeClient.Update(context.Background(), nodeNetwork); err != nil {
		return fmt.Errorf("update nodenetwork %s/%s status failed, err %s",
			nodeNetwork.GetName(), nodeNetwork.GetNamespace(), err.Error())
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
	resp, err := p.cloudNetClient.GetAvailableSubnet(context.Background(), req)
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

// allocate eni primary ip from cloud netservice
func (p *Processor) allocateEniPrimaryIP(instanceID, zone string, index uint64) (*pbcommon.IPObject, error) {
	resp, err := p.cloudNetClient.AllocateEni(context.Background(), &pbcloudnet.AllocateEniReq{
		Seq:        common.TimeSequence(),
		InstanceID: instanceID,
		Zone:       zone,
		Cluster:    p.option.Cluster,
		Index:      index,
	})
	if err != nil {
		blog.Errorf("allocate eni primary with param %s/%s/%s to netservice failed, err %s",
			instanceID, zone, p.option.Cluster, err.Error())
		return nil, fmt.Errorf("allocate eni primary with param %s/%s/%s to netservice failed, err %s",
			instanceID, zone, p.option.Cluster, err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		blog.Errorf("allocate eni primary with param %s/%s/%s to netservice response errCode %d errMsg %s",
			instanceID, zone, p.option.Cluster, resp.ErrCode, resp.ErrMsg)
		return nil, fmt.Errorf(
			"allocate eni primary with param %s/%s/%s to netservice response errCode %d errMsg %s",
			instanceID, zone, p.option.Cluster, resp.ErrCode, resp.ErrMsg)
	}
	return resp.EniPrimaryIP, nil
}

// delete eni primary ip record from cloud netservice
func (p *Processor) releaseEniPrimaryIP(instanceID, eniPrimaryIP string, index uint64) error {
	resp, err := p.cloudNetClient.ReleaseEni(context.Background(), &pbcloudnet.ReleaseEniReq{
		Seq:          common.TimeSequence(),
		InstanceID:   instanceID,
		EniPrimaryIP: eniPrimaryIP,
		Index:        index,
	})
	if err != nil {
		blog.Errorf("release eni primary ip %s for host %s failed, err %s",
			eniPrimaryIP, instanceID, err.Error())
		return fmt.Errorf("release eni primary ip %s with host %s failed, err %s",
			eniPrimaryIP, instanceID, err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		blog.Errorf("release eni primary ip %s for host %s response errCode %s errMsg %s",
			eniPrimaryIP, instanceID, resp.ErrCode, resp.ErrMsg)
		return fmt.Errorf("release eni primary ip %s for host %s response errCode %s errMsg %s",
			eniPrimaryIP, instanceID, resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

// result true stands for there are still ips on host
func (p *Processor) checkIPOnEni(eniObj *cloudv1.ElasticNetworkInterface) (bool, error) {
	cloudips := &cloudv1.CloudIPList{}
	if err := p.kubeClient.List(context.Background(), cloudips,
		&client.MatchingLabels{
			constant.IPLabelKeyForENI:          eniObj.EniID,
			constant.IPLabelKeyForClusterLayer: strconv.FormatBool(true),
		}); err != nil {

		blog.Errorf("list cloud ip on eni %s failed, err %s", eniObj.EniID, err.Error())
		return true, fmt.Errorf("list cloud ip on node %s failed, err %s", eniObj.EniID, err.Error())
	}
	if len(cloudips.Items) == 0 {
		return false, nil
	}
	blog.Infof("found %d ips on eni %s", len(cloudips.Items), eniObj.EniID)
	return true, nil
}

// clean ip on eni
func (p *Processor) cleanIPOnEni(eniObj *cloudv1.ElasticNetworkInterface, isForce bool) error {
	// if isForce is true, clean all cloudip for this eni in cluster
	if isForce {
		cloudIPList := &cloudv1.CloudIPList{}
		if err := p.kubeClient.List(context.Background(), cloudIPList, &client.MatchingLabels{
			constant.IPLabelKeyForENI:          eniObj.EniID,
			constant.IPLabelKeyForClusterLayer: strconv.FormatBool(true),
		}); err != nil {
			blog.Errorf("list cloud ip on eni %s failed, err %s", eniObj.EniID, err.Error())
			return fmt.Errorf("list cloud ip on node %s failed, err %s", eniObj.EniID, err.Error())
		}
		for _, cloudIP := range cloudIPList.Items {
			// when bcs-cloud-netservice and bcs-cloud-netcontroller are in same cluster,
			// bcs-cloud-netservice will store all IP resources in bcs-system
			if cloudIP.GetNamespace() == constant.CloudCrdNamespaceBcsSystem {
				blog.Infof("skip cloudip in bcs-system", cloudIP.GetName())
				continue
			}
			if err := p.kubeClient.Delete(context.Background(), &cloudIP); err != nil {
				return fmt.Errorf("delete ip %s/%s when clean eni", cloudIP.GetName(), cloudIP.GetNamespace())
			}
			blog.Infof("delete cloudip %s/%s successfully", cloudIP.GetName(), cloudIP.GetNamespace())
		}
	}
	// call clean eni api to bcs-cloud-netservice
	resp, err := p.cloudNetClient.CleanEni(context.Background(), &pbcloudnet.CleanEniReq{
		Seq:     common.TimeSequence(),
		EniID:   eniObj.EniID,
		IsForce: isForce,
	})
	if err != nil {
		return fmt.Errorf("clean eni failed, err %s", err.Error())
	}
	if resp.ErrCode != pbcommon.ErrCode_ERROR_OK {
		return fmt.Errorf("clean eni failed, errCode %d errMsg %s", resp.ErrCode, resp.ErrMsg)
	}
	return nil
}

func (p *Processor) reconcileEni(
	nodeVMInfo *cloudv1.VMInfo, subnetID, addr string, index int) (*cloudv1.ElasticNetworkInterface, error) {

	eniName := generateEniName(nodeVMInfo.InstanceID, index)
	eniCrdObj, err := p.cloudClient.CreateENI(eniName, subnetID, addr, 0)
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
	eniCrdObj.Index = index
	eniCrdObj.EniIfaceName = getEniIfaceName(index)
	eniCrdObj.RouteTableID = getRouteTableID(index)
	eniCrdObj.Status = constant.NodeNetworkEniStatusNotReady
	return eniCrdObj, nil
}
