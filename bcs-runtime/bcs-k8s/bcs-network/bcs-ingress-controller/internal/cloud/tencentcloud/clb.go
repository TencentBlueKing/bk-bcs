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

package tencentcloud

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/common"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	tclb "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/clb/v20180317"
	tcommon "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	k8scorev1 "k8s.io/api/core/v1"
)

// Clb client to operate clb instance
type Clb struct {
	sdkWrapper *SdkWrapper
	apiWrapper *APIWrapper
}

// NewClb create clb client
func NewClb() (*Clb, error) {
	sdkWrapper, err := NewSdkWrapper()
	if err != nil {
		return nil, err
	}
	apiWrapper, err := NewAPIWrapper()
	if err != nil {
		return nil, err
	}
	return &Clb{
		sdkWrapper: sdkWrapper,
		apiWrapper: apiWrapper,
	}, nil
}

// NewClbWithSecretIDKey create clb client with secret id and secret key
func NewClbWithSecretIDKey(id, key string) (*Clb, error) {
	sdkWrapper, err := NewSdkWrapperWithSecretIDKey(id, key)
	if err != nil {
		return nil, err
	}
	apiWrapper, err := NewAPIWrapperWithSecretIDKey(id, key)
	if err != nil {
		return nil, err
	}
	return &Clb{
		sdkWrapper: sdkWrapper,
		apiWrapper: apiWrapper,
	}, nil
}

// NewClbWithSecret create clb client with k8s secret
func NewClbWithSecret(secret *k8scorev1.Secret) (cloud.LoadBalance, error) {
	secretIDBytes, ok := secret.Data[EnvNameTencentCloudAccessKeyID]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", EnvNameTencentCloudAccessKeyID,
			secret.Namespace, secret.Name)
	}
	secretKeyBytes, ok := secret.Data[EnvNameTencentCloudAccessKey]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", EnvNameTencentCloudAccessKey,
			secret.Namespace, secret.Name)
	}
	sdkWrapper, err := NewSdkWrapperWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
	if err != nil {
		return nil, err
	}
	apiWrapper, err := NewAPIWrapperWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
	if err != nil {
		return nil, err
	}
	return &Clb{
		sdkWrapper: sdkWrapper,
		apiWrapper: apiWrapper,
	}, nil
}

// DescribeLoadBalancer get loadbalancer object by id
func (c *Clb) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	req := tclb.NewDescribeLoadBalancersRequest()
	if len(lbID) != 0 {
		req.LoadBalancerIds = tcommon.StringPtrs([]string{lbID})
	}
	if len(name) != 0 {
		req.LoadBalancerName = tcommon.StringPtr(name)
	}
	req.Forward = tcommon.Int64Ptr(1)

	ctime := time.Now()
	resp, err := c.sdkWrapper.DescribeLoadBalancers(region, req)
	if err != nil {
		blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		cloud.StatRequest("DescribeLoadBalancers", cloud.MetricAPIFailed, ctime, time.Now())
		return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
	}
	cloud.StatRequest("DescribeLoadBalancers", cloud.MetricAPISuccess, ctime, time.Now())

	if len(resp.Response.LoadBalancerSet) == 0 {
		return nil, cloud.ErrLoadbalancerNotFound
	}
	var resplb *tclb.LoadBalancer
	for _, lb := range resp.Response.LoadBalancerSet {
		if len(lbID) != 0 && lbID == *lb.LoadBalancerId {
			resplb = lb
			break
		}
		if len(name) != 0 && name == *lb.LoadBalancerName {
			resplb = lb
			break
		}
	}
	if resplb == nil {
		blog.Errorf("lb not found in resp %s", resp.ToJsonString())
		return nil, cloud.ErrLoadbalancerNotFound
	}
	retlb := &cloud.LoadBalanceObject{
		Region: region,
	}
	if resplb.LoadBalancerId != nil {
		retlb.LbID = *resplb.LoadBalancerId
	}
	if resplb.LoadBalancerType != nil {
		retlb.Type = *resplb.LoadBalancerType
	}
	if resplb.LoadBalancerName != nil {
		retlb.Name = *resplb.LoadBalancerName
	}
	retlb.IPs = tcommon.StringValues(resplb.LoadBalancerVips)
	return retlb, nil
}

// DescribeLoadBalancerWithNs get loadbalancer object by id or name with namespace specified
func (c *Clb) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	return c.DescribeLoadBalancer(region, lbID, name)
}

// IsNamespaced if client is namespaced
func (c *Clb) IsNamespaced() bool {
	return false
}

// EnsureListener ensure listener to cloud, and get listener info
func (c *Clb) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	cloudListener, err := c.getListenerInfoByPort(region, listener.Spec.LoadbalancerID,
		listener.Spec.Protocol, listener.Spec.Port)
	if err != nil {
		if errors.Is(err, cloud.ErrListenerNotFound) {
			// to create listener
			return c.createListner(region, listener)
		}
		return "", err
	}

	blog.V(5).Infof("new listener %+v", listener)
	blog.V(5).Infof("cloud listener %+v", cloudListener)

	if strings.ToLower(listener.Spec.Protocol) != strings.ToLower(cloudListener.Spec.Protocol) {
		// delete listener
		err := c.deleteListener(region, cloudListener.Spec.LoadbalancerID,
			cloudListener.Spec.Protocol, listener.Spec.Port)
		if err != nil {
			return "", err
		}
		// create listener
		listenerID, err := c.createListner(region, listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	if err := c.updateListener(region, listener, cloudListener); err != nil {
		return "", err
	}
	return cloudListener.Status.ListenerID, nil
}

// DeleteListener delete listener by name
func (c *Clb) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	return c.deleteListener(region, listener.Spec.LoadbalancerID,
		listener.Spec.Protocol, listener.Spec.Port)
}

// EnsureMultiListeners ensure multiple listeners to cloud
func (c *Clb) EnsureMultiListeners(
	region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	var portList []int
	for _, li := range listeners {
		portList = append(portList, li.Spec.Port)
	}
	cloudListenerMap, err := c.batchDescribeListeners(region, lbID, portList)
	if err != nil {
		return nil, err
	}
	addListeners := make([]*networkextensionv1.Listener, 0)
	updatedListeners := make([]*networkextensionv1.Listener, 0)
	deleteCloudListeners := make([]*networkextensionv1.Listener, 0)
	for _, li := range listeners {
		cloudLi, ok := cloudListenerMap[common.GetListenerNameWithProtocol(
			lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)]
		if !ok {
			addListeners = append(addListeners, li)
		} else {
			if strings.ToLower(cloudLi.Spec.Protocol) != strings.ToLower(li.Spec.Protocol) {
				deleteCloudListeners = append(deleteCloudListeners, cloudLi)
				addListeners = append(addListeners, li)
			} else {
				updatedListeners = append(updatedListeners, li)
			}
		}
	}

	if len(deleteCloudListeners) != 0 {
		var delListenerIDs []string
		for _, li := range deleteCloudListeners {
			if len(li.Status.ListenerID) != 0 {
				delListenerIDs = append(delListenerIDs, li.Status.ListenerID)
			}
		}
		if err := c.batchDeleteListener(region, lbID, delListenerIDs); err != nil {
			return nil, err
		}
	}

	retMap := make(map[string]string)
	addListenerGroups := splitListenersToDiffProtocol(addListeners)
	for _, group := range addListenerGroups {
		if len(group) != 0 {
			batches := splitListenersToDiffBatch(group)
			for _, batch := range batches {
				switch group[0].Spec.Protocol {
				case ClbProtocolHTTP, ClbProtocolHTTPS:
					liIDMap, err := c.batchCreate7LayerListener(region, batch)
					if err != nil {
						blog.Warnf("batch create 7 layer listener failed, err %s", err.Error())
						continue
					}
					for liName, liID := range liIDMap {
						retMap[liName] = liID
					}
				case ClbProtocolTCP, ClbProtocolUDP:
					liIDMap, err := c.batchCreate4LayerListener(region, batch)
					if err != nil {
						blog.Warnf("batch create 4 layer listener failed, err %s", err.Error())
						continue
					}
					for liName, liID := range liIDMap {
						retMap[liName] = liID
					}
				default:
					blog.Warnf("invalid batch protocol %s", group[0].Spec.Protocol)
					continue
				}
			}
		}
	}

	updateListenerGroups := splitListenersToDiffProtocol(updatedListeners)
	for _, group := range updateListenerGroups {
		if len(group) != 0 {
			cloudListenerGroup := make([]*networkextensionv1.Listener, 0)
			for _, li := range group {
				cloudListenerGroup = append(cloudListenerGroup, cloudListenerMap[common.GetListenerNameWithProtocol(
					lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)])
			}
			switch group[0].Spec.Protocol {
			case ClbProtocolHTTP, ClbProtocolHTTPS:
				isErrArr, err := c.batchUpdate7LayerListeners(region, group, cloudListenerGroup)
				if err != nil {
					blog.Warnf("batch update 7 layer listeners %s failed, err %s", getListenerNames(group), err.Error())
					continue
				}
				for index, isErr := range isErrArr {
					if !isErr {
						retMap[group[index].GetName()] = cloudListenerGroup[index].Status.ListenerID
					} else {
						blog.Warnf("update 7 layer listener %s failed in batch", group[index].GetName())
					}
				}
			case ClbProtocolTCP, ClbProtocolUDP:
				isErrArr, err := c.batchUpdate4LayerListener(region, group, cloudListenerGroup)
				if err != nil {
					blog.Infof("batch update 4 layer listeners %s failed, err %s", getListenerNames(group), err.Error())
					continue
				}
				for index, isErr := range isErrArr {
					if !isErr {
						retMap[group[index].GetName()] = cloudListenerGroup[index].Status.ListenerID
					} else {
						blog.Warnf("update 4 layer listener %s failed in batch", group[index].GetName())
					}
				}
			default:
				blog.Warnf("invalid batch protocol %s", group[0].Spec.Protocol)
				continue
			}
		}
	}

	return retMap, nil
}

// DeleteMultiListeners delete multiple listeners from cloud
func (c *Clb) DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return fmt.Errorf("listeners cannot be empty")
	}
	var listenerIDs []string
	for _, li := range listeners {
		if len(li.Status.ListenerID) != 0 {
			listenerIDs = append(listenerIDs, li.Status.ListenerID)
		} else {
			blog.Warnf("listener %s has no listenerID when do batch deletion", li.GetName())
		}
	}
	// when get no listenerID, means no listener was created before
	if len(listenerIDs) == 0 {
		blog.Warnf("no listenerIDs to do batch deletion")
		return nil
	}
	// when delete with listenerID which is not existed in cloud, cloud will return error
	// so here describe listener first
	req := tclb.NewDescribeListenersRequest()
	req.LoadBalancerId = tcommon.StringPtr(lbID)
	req.ListenerIds = tcommon.StringPtrs(listenerIDs)
	resp, err := c.sdkWrapper.DescribeListeners(region, req)
	if err != nil {
		return fmt.Errorf("describe listener failed, err %s", err.Error())
	}
	var cloudListenerIDs []string
	for _, li := range resp.Response.Listeners {
		cloudListenerIDs = append(cloudListenerIDs, *li.ListenerId)
	}
	// describe listener success but no existed cloudListenerIDs in cloud
	// It's possible delete all listeners when listenerIds be empty
	if len(cloudListenerIDs) == 0 {
		blog.Warnf("no cloudListenerIDs to do batch deletion")
		return nil
	}

	return c.batchDeleteListener(region, lbID, cloudListenerIDs)
}

// EnsureSegmentListener ensure listener with port segment
func (c *Clb) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	cloudListener, err := c.getListenerInfoByPort(region, listener.Spec.LoadbalancerID,
		listener.Spec.Protocol, listener.Spec.Port)
	if err != nil {
		if errors.Is(err, cloud.ErrListenerNotFound) {
			// to create listener
			listenerID, err := c.createSegmentListener(region, listener)
			if err != nil {
				return "", err
			}
			return listenerID, nil
		}
		return "", nil
	}

	blog.V(5).Infof("new segment listener %+v", listener)
	blog.V(5).Infof("cloud segment listener %+v", cloudListener)

	if strings.ToLower(listener.Spec.Protocol) != strings.ToLower(cloudListener.Spec.Protocol) {
		// delete listener
		err := c.deleteSegmentListener(region, listener.Spec.LoadbalancerID, listener.Spec.Port)
		if err != nil {
			return "", err
		}
		// create listener
		listenerID, err := c.createSegmentListener(region, listener)
		if err != nil {
			return "", err
		}
		return listenerID, nil
	}

	if err := c.updateSegmentListener(region, listener, cloudListener); err != nil {
		return "", err
	}
	return cloudListener.Status.ListenerID, nil
}

// EnsureMultiSegmentListeners ensure multi segment listeners
func (c *Clb) EnsureMultiSegmentListeners(region, lbID string, listeners []*networkextensionv1.Listener) (
	map[string]string, error) {
	var portList []int
	for _, li := range listeners {
		portList = append(portList, li.Spec.Port)
	}
	cloudListenerMap, err := c.batchDescribeListeners(region, lbID, portList)
	if err != nil {
		return nil, err
	}
	addListeners := make([]*networkextensionv1.Listener, 0)
	updatedListeners := make([]*networkextensionv1.Listener, 0)
	existedListeners := make([]*networkextensionv1.Listener, 0)
	deleteCloudListeners := make([]*networkextensionv1.Listener, 0)
	for _, li := range listeners {
		cloudLi, ok := cloudListenerMap[common.GetListenerNameWithProtocol(
			lbID, li.Spec.Protocol, li.Spec.Port, li.Spec.EndPort)]
		if !ok {
			addListeners = append(addListeners, li)
		} else {
			if strings.ToLower(cloudLi.Spec.Protocol) != strings.ToLower(li.Spec.Protocol) {
				deleteCloudListeners = append(deleteCloudListeners, cloudLi)
				addListeners = append(addListeners, li)
			} else {
				updatedListeners = append(updatedListeners, li)
				existedListeners = append(existedListeners, cloudLi)
			}
		}
	}

	if len(deleteCloudListeners) != 0 {
		var delListenerIDs []string
		for _, li := range deleteCloudListeners {
			if len(li.Status.ListenerID) != 0 {
				delListenerIDs = append(delListenerIDs, li.Status.ListenerID)
			}
		}
		if err := c.batchDeleteListener(region, lbID, delListenerIDs); err != nil {
			return nil, err
		}
	}

	retMap := make(map[string]string)
	if len(addListeners) != 0 {
		liIDMap, err := c.batchCreateSegment4LayerListener(region, addListeners)
		if err != nil {
			blog.Warnf("batch create 4 layer listener segment failed, err %s", err.Error())
		} else {
			for liName, liID := range liIDMap {
				retMap[liName] = liID
			}
		}
	}
	if len(updatedListeners) != 0 {
		isErrArr, err := c.batchUpdate4LayerListener(region, updatedListeners, existedListeners)
		if err != nil {
			blog.Warnf("batch update 4 layer listener segment failed, err %s", err.Error())
		}
		for index, li := range updatedListeners {
			if !isErrArr[index] {
				retMap[li.GetName()] = existedListeners[index].Status.ListenerID
			}
		}
	}
	return retMap, nil
}

// DeleteSegmentListener delete segment listener
func (c *Clb) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	return c.deleteListener(region, listener.Spec.LoadbalancerID,
		listener.Spec.Protocol, listener.Spec.Port)
}

// DescribeBackendStatus describe clb backend status, the input ns is no use here, only effects in namespaced cloud client
func (c *Clb) DescribeBackendStatus(region, ns string, lbIDs []string) (
	map[string][]*cloud.BackendHealthStatus, error) {
	return c.getBackendHealthStatus(region, ns, lbIDs)
}
