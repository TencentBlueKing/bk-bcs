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

package gcp

import (
	"fmt"
	"os"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
	k8scorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// GCLB client to operate GCLB instance
type GCLB struct {
	project    string
	sdkWrapper *SdkWrapper
	client     client.Client
}

// NewGclb create GCLB client
func NewGclb(k8sClient client.Client) (*GCLB, error) {
	sdkWrapper, err := NewSdkWrapper()
	if err != nil {
		return nil, err
	}
	return &GCLB{
		sdkWrapper: sdkWrapper,
		client:     k8sClient,
		project:    os.Getenv("GCP_PROJECT_ID"),
	}, nil
}

// NewGclbWithSecret create gclb client with k8s secret
func NewGclbWithSecret(secret *k8scorev1.Secret, k8sClient client.Client) (cloud.LoadBalance, error) {
	credentials, ok := secret.Data[EnvNameGCPCredentials]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", EnvNameGCPCredentials,
			secret.Namespace, secret.Name)
	}
	sdkWrapper, err := NewSdkWrapperWithSecretIDKey(credentials)
	if err != nil {
		return nil, err
	}
	return &GCLB{
		sdkWrapper: sdkWrapper,
		client:     k8sClient,
		project:    os.Getenv("GCP_PROJECT_ID"),
	}, nil
}

var _ cloud.LoadBalance = &GCLB{}

// DescribeLoadBalancer get loadbalancer object by id or name
func (e *GCLB) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	out, err := e.sdkWrapper.GetAddress(e.project, region, lbID)
	if err != nil {
		blog.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
		return nil, fmt.Errorf("DescribeLoadBalancers failed, err %s", err.Error())
	}

	if out == nil {
		blog.Errorf("lb %s not found", lbID)
		return nil, cloud.ErrLoadbalancerNotFound
	}

	retlb := &cloud.LoadBalanceObject{
		LbID:   out.Name,
		Region: region,
		Name:   out.Name,
		IPs:    make([]string, 0),
		Type:   out.AddressType,
	}
	retlb.IPs = append(retlb.IPs, out.Address)
	return retlb, nil
}

// DescribeLoadBalancerWithNs get loadbalancer object by id or name with namespace specified
func (e *GCLB) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	return e.DescribeLoadBalancer(region, lbID, name)
}

// IsNamespaced if client is namespaced
func (e *GCLB) IsNamespaced() bool {
	return false
}

// EnsureListener ensure listener to cloud
func (e *GCLB) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	if listener.Spec.LoadbalancerID == "" {
		return "", fmt.Errorf("loadbalancer id is empty")
	}

	switch listener.Spec.Protocol {
	case ProtocolHTTP, ProtocolHTTPS:
		return e.ensureApplicationLBListener(region, listener)
	case ProtocolTCP, ProtocolUDP:
		return e.ensureNetworkLBListener(region, listener)
	default:
		blog.Errorf("invalid protocol %s", listener.Spec.Protocol)
		return "", fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}
}

// DeleteListener delete listener by name
func (e *GCLB) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	switch listener.Spec.Protocol {
	case ProtocolHTTP, ProtocolHTTPS:
		return e.deleteL7Listener(listener)
	case ProtocolTCP, ProtocolUDP:
		return e.deleteL4Listener(listener)
	default:
		blog.Errorf("invalid protocol %s", listener.Spec.Protocol)
		return fmt.Errorf("invalid protocol %s", listener.Spec.Protocol)
	}
}

// EnsureMultiListeners ensure multiple listeners to cloud
func (e *GCLB) EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	type listenerMap struct {
		name string
		id   string
	}
	retMap := make(map[string]string)
	errCh := make(chan error, len(listeners))
	retCh := make(chan listenerMap, len(listeners))
	wg := sync.WaitGroup{}
	wg.Add(len(listeners))

	// ensure listener
	for _, listener := range listeners {
		go func(listener *networkextensionv1.Listener) {
			liID, err := e.EnsureListener(region, listener)
			defer wg.Done()
			if err != nil {
				errCh <- err
				return
			}
			retCh <- listenerMap{name: listener.Name, id: liID}
		}(listener)
	}

	// wait for listener ensured
	wg.Wait()
	close(errCh)
	close(retCh)
	for e := range errCh {
		blog.Errorf("ensure listener failed, err %s", e.Error())
	}
	for li := range retCh {
		retMap[li.name] = li.id
	}
	return retMap, nil
}

// DeleteMultiListeners delete multiple listeners from cloud
func (e *GCLB) DeleteMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) error {
	errCh := make(chan error, 1)
	wg := sync.WaitGroup{}
	wg.Add(len(listeners))

	// ensure listener
	for _, listener := range listeners {
		go func(listener *networkextensionv1.Listener) {
			defer wg.Done()
			err := e.DeleteListener(region, listener)
			if err != nil {
				errCh <- err
			}
		}(listener)
	}

	// wait for listener ensured
	wg.Wait()
	close(errCh)
	for e := range errCh {
		return e
	}
	return nil
}

// EnsureSegmentListener ensure listener with port segment
func (e *GCLB) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	return e.EnsureListener(region, listener)
}

// EnsureMultiSegmentListeners ensure multi segment listeners
func (e *GCLB) EnsureMultiSegmentListeners(region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	type listenerMap struct {
		name string
		id   string
	}
	retMap := make(map[string]string)
	errCh := make(chan error, len(listeners))
	retCh := make(chan listenerMap, len(listeners))
	wg := sync.WaitGroup{}
	wg.Add(len(listeners))

	// ensure listener
	for _, listener := range listeners {
		go func(listener *networkextensionv1.Listener) {
			liID, err := e.EnsureSegmentListener(region, listener)
			defer wg.Done()
			if err != nil {
				errCh <- err
				return
			}
			retCh <- listenerMap{name: listener.Name, id: liID}
		}(listener)
	}

	// wait for listener ensured
	wg.Wait()
	close(errCh)
	close(retCh)
	for e := range errCh {
		blog.Errorf("ensure listener failed, err %s", e.Error())
	}
	for li := range retCh {
		retMap[li.name] = li.id
	}
	return retMap, nil
}

// DeleteSegmentListener delete segment listener
func (e *GCLB) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	return e.DeleteListener(region, listener)
}

// DescribeBackendStatus describe GCLB backend status, the input ns is no use here, only effects in namespaced cloud client
func (e *GCLB) DescribeBackendStatus(region, ns string, lbIDs []string) (map[string][]*cloud.BackendHealthStatus, error) {
	return nil, nil
}
