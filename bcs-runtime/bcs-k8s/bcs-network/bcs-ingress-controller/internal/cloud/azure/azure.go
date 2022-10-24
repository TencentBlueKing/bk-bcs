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
 *
 */

package azure

import (
	"fmt"
	"strings"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/constant"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/pkg/errors"
	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
func NewAlbWithSecret(secret *k8scorev1.Secret, _ client.Client) (cloud.LoadBalance, error) {
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
		return a.ensureLoadBalancerListener(region, listener)
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		return a.ensureApplicationGatewayListener(region, listener)
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
		return a.deleteLoadBalancerListener(region, listener)
	case AzureProtocolHTTP, AzureProtocolHTTPS:
		return a.deleteApplicationGatewayListener(region, listener)
	default:
		return fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}

}

// EnsureMultiListeners ensure multiple listeners to cloud
func (a *Alb) EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]cloud.Result,
	error) {
	retMap := make(map[string]cloud.Result)
	for _, listener := range listeners {
		liName, err := a.EnsureListener(region, listener)
		if err != nil {
			err = errors.Wrapf(err, "ensure multi listener failed in listener'%s/%s'", listener.GetNamespace(),
				listener.GetName())
			retMap[listener.Name] = cloud.Result{IsError: true, Err: err}
		} else {
			retMap[listener.Name] = cloud.Result{IsError: false, Res: liName}
		}
	}
	return retMap, nil
}

// DeleteMultiListeners delete multiple listeners from cloud
func (a *Alb) DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error {
	for _, listener := range listeners {
		err := a.DeleteListener(region, listener)
		if err != nil {
			return errors.Wrapf(err, "delete multi listener failed in listener: '%s/%s'",
				listener.GetNamespace(), listener.Name)
		}
	}
	return nil
}

// EnsureSegmentListener ensure listener with port segment
func (a *Alb) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if listener.Spec.EndPort == 0 {
		return a.EnsureListener(region, listener)
	}
	// create listener for each port
	portIndex := 0
	listenerIds := make([]string, 0)
	for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
		// generate single port listener to ensure listener
		li := listener.DeepCopy()
		li.Spec.Port = i
		li.Spec.EndPort = 0
		if li.Spec.TargetGroup != nil {
			for j := range li.Spec.TargetGroup.Backends {
				li.Spec.TargetGroup.Backends[j].Port += portIndex
			}
		}
		portIndex++
		liID, err := a.EnsureListener(region, li)
		if err != nil {
			return "", errors.Wrapf(err, "ensure listener %s(%d) failed", listener.Name, li.Spec.Port)
		}
		listenerIds = append(listenerIds, liID)
	}
	return strings.Join(listenerIds, ","), nil
}

// EnsureMultiSegmentListeners ensure multi segment listeners
func (a *Alb) EnsureMultiSegmentListeners(region, lbID string, listeners []*networkextensionv1.Listener) (
	map[string]cloud.Result, error) {
	retMap := make(map[string]cloud.Result)
	for _, listener := range listeners {
		liID, err := a.EnsureSegmentListener(region, listener)
		if err != nil {
			err = errors.Wrapf(err, "ensure multi segment listener failed in %s", listener.Name)
			retMap[listener.Name] = cloud.Result{IsError: true, Err: err}
		} else {
			retMap[listener.Name] = cloud.Result{IsError: false, Res: liID}
		}
	}
	return retMap, nil
}

// DeleteSegmentListener delete segment listener
func (a *Alb) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	if listener.Spec.EndPort == 0 {
		return a.DeleteListener(region, listener)
	}
	// delete listener for each port
	for i := listener.Spec.Port; i <= listener.Spec.EndPort; i++ {
		// generate single port listener to ensure listener
		li := listener.DeepCopy()
		li.Spec.Port = i
		li.Spec.EndPort = 0
		err := a.DeleteListener(region, li)
		if err != nil {
			return errors.Wrapf(err, "delete segment listener failed in %s(%d)", li.Name, li.Spec.Port)
		}
	}
	return nil
}

// DescribeBackendStatus describe Azure backend status, the input ns is no use here,
// only effects in namespaced cloud client
func (a *Alb) DescribeBackendStatus(region, ns string, lbIDs []string) (map[string][]*cloud.BackendHealthStatus,
	error) {
	return nil, nil
}
