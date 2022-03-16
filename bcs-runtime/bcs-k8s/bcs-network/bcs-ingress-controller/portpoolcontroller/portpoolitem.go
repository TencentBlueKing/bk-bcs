/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package portpoolcontroller

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/generator"
	netextv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// PortPoolItemHandler port pool item
type PortPoolItemHandler struct {
	PortPoolName string
	// namespace for port pool item
	Namespace string
	// default region for ingress converter
	DefaultRegion string
	ListenerAttr  *netextv1.IngressListenerAttribute
	// cloud loadbalance client
	LbClient cloud.LoadBalance
	// client for k8s
	K8sClient client.Client
}

// do something when new port pool item is added
// the returned bool value show that whether it should retry
func (ppih *PortPoolItemHandler) ensurePortPoolItem(
	item *netextv1.PortPoolItem, itemStatus *netextv1.PortPoolItemStatus) (*netextv1.PortPoolItemStatus, bool) {
	var retItemStatus *netextv1.PortPoolItemStatus
	if itemStatus == nil {
		retItemStatus = &netextv1.PortPoolItemStatus{
			ItemName:        item.ItemName,
			LoadBalancerIDs: item.LoadBalancerIDs,
			StartPort:       item.StartPort,
			EndPort:         item.EndPort,
			SegmentLength:   item.SegmentLength,
		}
	} else {
		// endport can be increased
		retItemStatus = itemStatus.DeepCopy()
		retItemStatus.EndPort = item.EndPort
	}
	// check loadbalanceIDs
	lbIDs := make([]string, len(item.LoadBalancerIDs))
	copy(lbIDs, item.LoadBalancerIDs)
	sort.Strings(lbIDs)

	lbObjList, err := ppih.getCloudListenersByRegionIDs(lbIDs)
	if err != nil {
		blog.Infof("port pool item %s become %s, get loadbalancer info of %v failed, err %s",
			item.ItemName, constant.PortPoolItemStatusNotReady, lbIDs, err.Error())
		retItemStatus.Status = constant.PortPoolItemStatusNotReady
		retItemStatus.Message = fmt.Sprintf("get loadbalancer info of %v failed, err %s", lbIDs, err.Error())
		return retItemStatus, true
	}
	retItemStatus.PoolItemLoadBalancers = lbObjList

	// check listeners belong to this item
	var errMsgs []string
	for _, lbObj := range lbObjList {
		segmentLen := item.SegmentLength
		if segmentLen == 0 {
			segmentLen = 1
		}
		if err := ppih.ensureListeners(
			lbObj.Region, lbObj.LoadbalancerID, item.StartPort, item.EndPort, segmentLen, item.Protocol); err != nil {
			blog.Warnf("listeners of loadbalance %s not all ready, err %s", lbObj.LoadbalancerID, err.Error())
			errMsgs = append(errMsgs, fmt.Sprintf("lb %s: %s", lbObj.LoadbalancerID, err.Error()))
		}
	}
	if len(errMsgs) != 0 {
		retItemStatus.Message = strings.Join(errMsgs, ";")
		retItemStatus.Status = constant.PortPoolItemStatusNotReady
		return retItemStatus, true
	}
	retItemStatus.Message = constant.PortPoolItemMessageReady
	retItemStatus.Status = constant.PortPoolItemStatusReady
	return retItemStatus, false
}

// check an port pool item can be deleted from status
// If the returned error is empty, it is considered that it can be deleted normally
func (ppih *PortPoolItemHandler) checkPortPoolItemDeletion(itemStatus *netextv1.PortPoolItemStatus) error {
	if itemStatus.Status != constant.PortPoolItemStatusDeleting {
		return fmt.Errorf("wait item %s of pool %s/%s status becoming %s",
			itemStatus.ItemName, ppih.PortPoolName, ppih.Namespace, constant.PortPoolItemStatusDeleting)
	}
	// check whether there is port bind object related to this port pool item
	set := k8slabels.Set(map[string]string{
		fmt.Sprintf(netextv1.PortPoolBindingLabelKeyFromat, ppih.PortPoolName, ppih.Namespace): itemStatus.ItemName,
	})
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(set))
	if err != nil {
		return fmt.Errorf("get selector from set %v failed, err %s", set, err.Error())
	}
	bindingList := &netextv1.PortBindingList{}
	if err := ppih.K8sClient.List(
		context.Background(), bindingList, &client.ListOptions{LabelSelector: selector}); err != nil {
		return fmt.Errorf("failed to list port bind list, err %s", err.Error())
	}
	if len(bindingList.Items) != 0 {
		return fmt.Errorf("port binding object found! cannot delete port pool item %s of pool %s/%s",
			itemStatus.ItemName, ppih.PortPoolName, ppih.Namespace)
	}

	found := false
	// check whether there is listener related to this port pool item
	for _, lbObj := range itemStatus.PoolItemLoadBalancers {
		listenerList, err := ppih.getListenerList(lbObj.LoadbalancerID)
		if err != nil {
			return err
		}
		if len(listenerList.Items) != 0 {
			found = true
			for _, listener := range listenerList.Items {
				if listener.DeletionTimestamp == nil {
					if err := ppih.K8sClient.Delete(
						context.Background(), &listener, &client.DeleteOptions{}); err != nil {
						blog.Warnf("delete listener %s/%s failed, err %s",
							listener.GetName(), listener.GetNamespace(), err.Error())
					}
				}
			}
		}
	}
	if found {
		return fmt.Errorf("wait listener of item %s of pool %s/%s to delete",
			itemStatus.ItemName, ppih.PortPoolName, ppih.Namespace)
	}
	return nil
}

func (ppih *PortPoolItemHandler) getCloudListenersByRegionIDs(regionIDs []string) (
	[]*netextv1.IngressLoadBalancer, error) {
	var retLbs []*netextv1.IngressLoadBalancer
	for _, lbID := range regionIDs {
		var tmpRegion string
		var tmpID string
		var lbObj *cloud.LoadBalanceObject
		var err error
		tmpRegion, tmpID, err = common.GetLbRegionAndName(lbID)
		if err != nil {
			return nil, err
		}
		if len(tmpRegion) == 0 {
			tmpRegion = ppih.DefaultRegion
			tmpID = lbID
		}
		if ppih.LbClient.IsNamespaced() {
			lbObj, err = ppih.LbClient.DescribeLoadBalancerWithNs(ppih.Namespace, tmpRegion, tmpID, "")
		} else {
			lbObj, err = ppih.LbClient.DescribeLoadBalancer(tmpRegion, tmpID, "")
		}
		if err != nil {
			return nil, fmt.Errorf("describe lb %s info failed, err %s", lbID, err.Error())
		}

		retLbs = append(retLbs, &netextv1.IngressLoadBalancer{
			LoadbalancerID:   tmpID,
			LoadbalancerName: lbObj.Name,
			Region:           lbObj.Region,
			Type:             lbObj.Type,
			IPs:              lbObj.IPs,
			DNSName:          lbObj.DNSName,
			Scheme:           lbObj.Scheme,
			AWSLBType:        lbObj.AWSLBType,
		})
	}
	return retLbs, nil
}

// get listener from k8s apiserver
func (ppih *PortPoolItemHandler) getListenerList(lbID string) (*netextv1.ListenerList, error) {
	set := k8slabels.Set(map[string]string{
		netextv1.LabelKeyForPortPoolListener: netextv1.LabelValueTrue,
		netextv1.LabelKeyForLoadbalanceID:    generator.GetLabelLBId(lbID),
	})
	selector, err := k8smetav1.LabelSelectorAsSelector(k8smetav1.SetAsLabelSelector(set))
	if err != nil {
		return nil, fmt.Errorf("get selector from set %v failed, err %s", set, err.Error())
	}
	listenerList := &netextv1.ListenerList{}
	if err := ppih.K8sClient.List(context.Background(), listenerList, &client.ListOptions{
		LabelSelector: selector,
	}); err != nil {
		return nil, fmt.Errorf("get listener by labelSelector %s failed, err %s", selector.String(), err.Error())
	}
	return listenerList, nil
}

// ensure listeners about this port pool item
func (ppih *PortPoolItemHandler) ensureListeners(region, lbID string, startPort, endPort,
	segment uint32, protocol string) error {
	listenerList, err := ppih.getListenerList(lbID)
	if err != nil {
		return err
	}

	listenerMap := make(map[string]*netextv1.Listener)
	for i, listenerItem := range listenerList.Items {
		tmpKey := common.GetListenerNameWithProtocol(
			lbID, listenerItem.Spec.Protocol, listenerItem.Spec.Port, listenerItem.Spec.EndPort)
		listenerMap[tmpKey] = &listenerList.Items[i]
	}

	notReady := false
	for p := startPort; p < endPort; p += segment {
		protocolList := make([]string, 0)
		if len(protocol) == 0 {
			protocolList = []string{constant.PortPoolPortProtocolTCP, constant.PortPoolPortProtocolUDP}
		} else {
			protocolList = strings.Split(protocol, ",")
		}
		for _, protocol := range protocolList {
			tmpStartPort := p
			tmpEndPort := 0
			if segment > 1 {
				tmpEndPort = int(p + segment - 1)
			}
			tmpName := common.GetListenerNameWithProtocol(lbID, protocol, int(tmpStartPort), int(tmpEndPort))
			listener, ok := listenerMap[tmpName]
			if !ok {
				notReady = true
				if err := ppih.K8sClient.Create(context.Background(), ppih.generateListener(
					region, lbID, protocol, tmpStartPort, uint32(tmpEndPort),
				), &client.CreateOptions{}); err != nil {
					blog.Warnf("create listener %s failed, err %s", tmpName, err.Error())
				}
			} else {
				if len(listener.Status.ListenerID) == 0 {
					notReady = true
					blog.Warnf("listener %s is not ready", tmpName)
				}
			}
		}
	}
	if notReady {
		return fmt.Errorf("some listener of %s was not ready", lbID)
	}

	return nil
}

func (ppih *PortPoolItemHandler) generateListener(
	region, lbID, protocol string, startPort, endPort uint32) *netextv1.Listener {
	li := &netextv1.Listener{}
	segLabelValue := netextv1.LabelValueTrue
	listenerName := common.GetListenerNameWithProtocol(lbID, protocol, int(startPort), int(endPort))
	if endPort == 0 {
		segLabelValue = netextv1.LabelValueFalse
	}
	li.SetName(listenerName)
	li.SetNamespace(ppih.Namespace)
	li.SetLabels(map[string]string{
		netextv1.LabelKeyForPortPoolListener:  netextv1.LabelValueTrue,
		netextv1.LabelKeyForIsSegmentListener: segLabelValue,
		netextv1.LabelKeyForLoadbalanceID:     generator.GetLabelLBId(lbID),
		netextv1.LabelKeyForLoadbalanceRegion: region,
	})
	li.Finalizers = append(li.Finalizers, constant.FinalizerNameBcsIngressController)
	li.Spec.Port = int(startPort)
	li.Spec.EndPort = int(endPort)
	li.Spec.Protocol = protocol
	li.Spec.LoadbalancerID = lbID
	li.Spec.ListenerAttribute = ppih.ListenerAttr
	return li
}
