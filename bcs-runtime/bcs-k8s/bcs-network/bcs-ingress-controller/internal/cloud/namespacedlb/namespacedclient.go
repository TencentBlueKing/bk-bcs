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

package namespacedlb

import (
	"context"
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"

	k8scorev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// IDKeySecretName secret name to store secret id and secret key
	IDKeySecretName = "ingress-secret.networkextension.bkbcs.tencent.com"
)

// NamespacedLB client for cloud which is aware of listener namespace
type NamespacedLB struct {
	k8sClient client.Client

	// map for (namespace, cloud.LoadBalance)
	nsClientSet map[string]cloud.LoadBalance

	// func for create cloud.LoadBalance
	newLBFunc func(*k8scorev1.Secret, client.Client) (cloud.LoadBalance, error)
}

// NewNamespacedLB create namespaced lb client
func NewNamespacedLB(k8sClient client.Client, newLBFunc func(secret *k8scorev1.Secret, cli client.Client) (cloud.LoadBalance, error)) *NamespacedLB {
	return &NamespacedLB{
		k8sClient:   k8sClient,
		nsClientSet: make(map[string]cloud.LoadBalance),
		newLBFunc:   newLBFunc,
	}
}

// init client for namespace
func (nc *NamespacedLB) initNsClient(ns string) (cloud.LoadBalance, error) {
	var err error
	tmpSecret := &k8scorev1.Secret{}
	err = nc.k8sClient.Get(context.TODO(), k8stypes.NamespacedName{
		Name:      IDKeySecretName,
		Namespace: ns,
	}, tmpSecret)
	if err != nil {
		return nil, fmt.Errorf("get secret %s/%s failed, err %s", IDKeySecretName, ns, err.Error())
	}
	newClient, err := nc.newLBFunc(tmpSecret, nc.k8sClient)
	if err != nil {
		return nil, fmt.Errorf("create client for ns %s failed, err %s", ns, err.Error())
	}
	return newClient, nil
}

// get client for certain namespace, if it is not existed, try to create one
func (nc *NamespacedLB) getNsClient(ns string) (cloud.LoadBalance, error) {
	tmpClient, ok := nc.nsClientSet[ns]
	if !ok {
		newClient, err := nc.initNsClient(ns)
		if err != nil {
			return nil, err
		}
		nc.nsClientSet[ns] = newClient
		return newClient, nil
	}
	return tmpClient, nil
}

// DescribeLoadBalancerWithNs describe loadbalances with ns
func (nc *NamespacedLB) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	tmpClient, err := nc.getNsClient(ns)
	if err != nil {
		return nil, err
	}
	return tmpClient.DescribeLoadBalancer(region, lbID, name)
}

// DescribeLoadBalancer describe loadbalances with id or name
func (nc *NamespacedLB) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	return nil, fmt.Errorf("please use DescribeLoadBalancerWithNs for namespaced clb client")
}

// IsNamespaced if client is namespaced
func (nc *NamespacedLB) IsNamespaced() bool {
	return true
}

// EnsureListener implements LoadBalance interface
func (nc *NamespacedLB) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return "", err
	}
	return tmpClient.EnsureListener(region, listener)
}

// EnsureMultiListeners ensure multiple listeners to cloud
func (nc *NamespacedLB) EnsureMultiListeners(region, lbID string, listeners []*networkextensionv1.Listener) (
	map[string]string, error) {
	if len(listeners) == 0 {
		blog.Warnf("no listeners to be ensured")
		return nil, nil
	}
	tmpClient, err := nc.getNsClient(listeners[0].GetNamespace())
	if err != nil {
		return nil, err
	}
	return tmpClient.EnsureMultiListeners(region, lbID, listeners)
}

// DeleteListener implements LoadBalance interface
func (nc *NamespacedLB) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return err
	}
	return tmpClient.DeleteListener(region, listener)
}

// DeleteMultiListeners delete multi listeners
func (nc *NamespacedLB) DeleteMultiListeners(
	region, lbID string, listeners []*networkextensionv1.Listener) error {
	if len(listeners) == 0 {
		blog.Warnf("no listeners to be delete")
		return nil
	}
	tmpClient, err := nc.getNsClient(listeners[0].GetNamespace())
	if err != nil {
		return err
	}
	return tmpClient.DeleteMultiListeners(region, lbID, listeners)
}

// EnsureSegmentListener implements LoadBalance interface
func (nc *NamespacedLB) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return "", err
	}
	return tmpClient.EnsureSegmentListener(region, listener)
}

// EnsureMultiSegmentListeners ensure multi segment listener
func (nc *NamespacedLB) EnsureMultiSegmentListeners(
	region, lbID string, listeners []*networkextensionv1.Listener) (map[string]string, error) {
	if len(listeners) == 0 {
		blog.Warnf("no segment listeners to be ensured")
		return nil, nil
	}
	tmpClient, err := nc.getNsClient(listeners[0].GetNamespace())
	if err != nil {
		return nil, err
	}
	return tmpClient.EnsureMultiSegmentListeners(region, lbID, listeners)
}

// DeleteSegmentListener implements LoadBalance interface
func (nc *NamespacedLB) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return err
	}
	return tmpClient.DeleteSegmentListener(region, listener)
}

// DescribeBackendStatus describe clb backend status
func (nc *NamespacedLB) DescribeBackendStatus(region, ns string, lbIDs []string) (
	map[string][]*cloud.BackendHealthStatus, error) {
	tmpClient, err := nc.getNsClient(ns)
	if err != nil {
		return nil, err
	}

	return tmpClient.DescribeBackendStatus(region, ns, lbIDs)
}
