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

// Package azure load balancer related api wrapper
package azure

import (
	"fmt"

	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/eventer"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

// Alb client to operate Azure lb and appGateway instance
type Alb struct {
	sdkWrapper     *SdkWrapper
	resourceHelper *ResourceHelper
}

// NewAlb create alb client
func NewAlb() (*Alb, error) {
	sdkWrapper, err := NewSdkWrapper()
	if err != nil {
		return nil, errors.Wrapf(err, "create azure sdk wrapper failed")
	}
	resourceHelper := NewResourceHelper(sdkWrapper.subscriptionID, sdkWrapper.resourceGroupName)
	return &Alb{
		sdkWrapper:     sdkWrapper,
		resourceHelper: resourceHelper,
	}, nil
}

// NewAlbWithSecret create alb client with k8s secret
func NewAlbWithSecret(secret *k8scorev1.Secret, _ client.Client, _ eventer.WatchEventInterface) (cloud.LoadBalance, error) {
	secretIDBytes, ok := secret.Data[envNameAzureClientID]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", envNameAzureClientID,
			secret.Namespace, secret.Name)
	}
	secretKeyBytes, ok := secret.Data[envNameAzureClientSecret]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", envNameAzureClientSecret,
			secret.Namespace, secret.Name)
	}
	sdkWrapper, err := NewSdkWrapperWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
	if err != nil {
		return nil, errors.Wrapf(err, "create azure sdk wrapper with secret failed")
	}
	resourceHelper := NewResourceHelper(sdkWrapper.subscriptionID, sdkWrapper.resourceGroupName)
	return &Alb{
		sdkWrapper:     sdkWrapper,
		resourceHelper: resourceHelper,
	}, nil
}

// DescribeLoadBalancer with different protocol layer, get loadBalancer/applicationGateway object
func (a *Alb) DescribeLoadBalancer(region, lbID, name, protocolLayer string) (*cloud.LoadBalanceObject, error) {
	if lbID != "" {
		name = lbID
	}
	switch protocolLayer {
	case constant.ProtocolLayerTransport:
		loadBalancer, err := a.sdkWrapper.GetLoadBalancer(region, name)
		if err != nil {
			return nil, errors.Wrapf(err, "get load balancer'%s/%s' failed", region, name)
		}
		retLb := &cloud.LoadBalanceObject{
			LbID:        *loadBalancer.Name,
			Region:      *loadBalancer.Location,
			Name:        *loadBalancer.Name,
			AzureLBType: constant.LoadBalancerTypeLoadBalancer,
		}
		return retLb, nil
	case constant.ProtocolLayerApplication:
		appGateway, err := a.sdkWrapper.GetApplicationGateway(region, name)
		if err != nil {
			return nil, errors.Wrapf(err, "get application gateway'%s/%s' failed", region, name)
		}
		retLb := &cloud.LoadBalanceObject{
			LbID:        *appGateway.Name,
			Region:      *appGateway.Location,
			Name:        *appGateway.Name,
			AzureLBType: constant.LoadBalancerTypeApplicationGateway,
		}
		return retLb, nil
	default:
		return nil, errors.Errorf("unsupport protocol layer: %s", protocolLayer)
	}
}

// DescribeLoadBalancerWithNs with different protocol layer, get loadBalancer/applicationGateway object
func (a *Alb) DescribeLoadBalancerWithNs(ns, region, lbID, name, protocolLayer string) (*cloud.LoadBalanceObject,
	error) {
	return a.DescribeLoadBalancer(region, lbID, name, protocolLayer)
}

// IsNamespaced if client is namespaced
func (a *Alb) IsNamespaced() bool {
	return false
}

// EnsureListener ensure listener to cloud
func (a *Alb) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if listener.Spec.LoadbalancerID == "" {
		return "", fmt.Errorf("listener'%s' has empty loadbalancer id", listener.Name)
	}
	switch listener.Spec.Protocol {
	case AzureProtocolTCP, AzureProtocolUDP:
		retMap, err := a.ensureLoadBalancerListener(region, []*networkextensionv1.Listener{listener})
		if err != nil {
			return "", err
		}
		if cloudRes, ok := retMap[listener.GetName()]; !ok {
			return "", fmt.Errorf("ensure failed")
		} else if cloudRes.IsError {
			return "", cloudRes.Err
		}
		return listener.GetName(), nil
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		return listener.GetName(), a.ensureApplicationGatewayListener(region,
			[]*networkextensionv1.Listener{listener})
	default:
		return "", fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}

}

// DeleteListener delete listener by name
func (a *Alb) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.LoadbalancerID == "" {
		return fmt.Errorf("listener'%s' has empty loadbalancer id", listener.Name)
	}
	switch listener.Spec.Protocol {
	case AzureProtocolTCP, AzureProtocolUDP:
		return a.deleteLoadBalancerListener(region, []*networkextensionv1.Listener{listener})
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		return a.deleteApplicationGatewayListener(region, []*networkextensionv1.Listener{listener})
	default:
		return fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}

}

// EnsureMultiListeners ensure multiple listeners to cloud
func (a *Alb) EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]cloud.Result,
	error) {
	retMap := make(map[string]cloud.Result)
	if len(listeners) == 0 {
		return retMap, nil
	}
	listenerGroup := splitListenersToDiffProtocol(listeners)
	for _, group := range listenerGroup {
		if len(group) == 0 {
			continue
		}

		var err error
		switch group[0].Spec.Protocol {
		case AzureProtocolTCP, AzureProtocolUDP:
			l4RetMap, inErr := a.ensureLoadBalancerListener(region, group)
			if inErr != nil {
				for _, li := range group {
					retMap[li.GetName()] = cloud.Result{IsError: true, Err: inErr}
				}
				continue
			}
			for liName, res := range l4RetMap {
				retMap[liName] = res
			}
			continue
		case AzureProtocolHTTP, AzureProtocolHTTPS:
			err = a.ensureApplicationGatewayListener(region, group)
		default:
			err = fmt.Errorf("invalid protocol %s", group[0].Spec.Protocol)
		}

		if err != nil {
			err = errors.Wrapf(err, "ensure multi listener failed, protocol: %s", group[0].Spec.Protocol)
			blog.Warnf("%s", err.Error())
			for _, li := range group {
				retMap[li.GetName()] = cloud.Result{
					IsError: true,
					Err:     err,
				}
			}
			continue
		}

		for _, li := range group {
			retMap[li.GetName()] = cloud.Result{
				IsError: false,
				Res:     li.GetName(),
			}
		}
	}
	return retMap, nil
}

// DeleteMultiListeners delete multiple listeners from cloud
func (a *Alb) DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		return nil
	}
	listenerGroup := splitListenersToDiffProtocol(listeners)
	for _, group := range listenerGroup {
		if len(group) == 0 {
			continue
		}

		var err error
		switch group[0].Spec.Protocol {
		case AzureProtocolTCP, AzureProtocolUDP:
			err = a.deleteLoadBalancerListener(region, group)
		case AzureProtocolHTTP, AzureProtocolHTTPS:
			err = a.deleteApplicationGatewayListener(region, group)
		default:
			err = fmt.Errorf("invalid protocol %s", group[0].Spec.Protocol)
		}

		if err != nil {
			err = errors.Wrapf(err, "delete multi listener failed, protocol: %s", group[0].Spec.Protocol)
			blog.Warnf("%s", err.Error())
			return err
		}
	}
	return nil
}

// EnsureSegmentListener ensure listener with port segment
func (a *Alb) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	listenerList := splitSegListener([]*networkextensionv1.Listener{listener})
	resMap, err := a.ensureLoadBalancerListener(region, listenerList)
	if err != nil {
		return "", err
	}
	if res, ok := resMap[listener.GetName()]; ok {
		if res.IsError {
			return "", res.Err
		}
		return res.Res, nil
	}
	return "", fmt.Errorf("ensure failed")
}

// EnsureMultiSegmentListeners ensure multi segment listeners
func (a *Alb) EnsureMultiSegmentListeners(region, lbID string, listeners []*networkextensionv1.Listener) (
	map[string]cloud.Result, error) {
	listenerList := splitSegListener(listeners)
	return a.ensureLoadBalancerListener(region, listenerList)
}

// DeleteSegmentListener delete segment listener
func (a *Alb) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.EndPort == 0 {
		return a.DeleteListener(region, listener)
	}

	listenerList := splitSegListener([]*networkextensionv1.Listener{listener})
	err := a.deleteLoadBalancerListener(region, listenerList)
	if err != nil {
		return err
	}
	return nil
}

// DescribeBackendStatus describe Azure backend status, the input ns is no use here,
// only effects in namespaced cloud client
func (a *Alb) DescribeBackendStatus(region, ns string, lbIDs []string) (map[string][]*cloud.BackendHealthStatus,
	error) {
	return nil, nil
}
