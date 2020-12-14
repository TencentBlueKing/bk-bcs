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

package cloud

import (
	"context"
	"encoding/base64"
	"fmt"

	k8scorev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubernetes/apis/networkextension/v1"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud"
	"github.com/Tencent/bk-bcs/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
)

const (
	// IDKeySecretName secret name to store secret id and secret key
	IDKeySecretName = "secret.networkextension.bkbcs.tencent.com"
)

// NamespacedClb client for cloud which is aware of listener namespace
type NamespacedClb struct {
	k8sClient client.Client

	// map for (namespace, cloud.LoadBalance)
	nsClientSet map[string]cloud.LoadBalance
}

// NewNamespacedClb create namespaced clb client
func NewNamespacedClb(k8sClient client.Client) *NamespacedClb {
	return &NamespacedClb{
		k8sClient:   k8sClient,
		nsClientSet: make(map[string]cloud.LoadBalance),
	}
}

// init client for namespace
func (nc *NamespacedClb) initNsClient(ns string) (cloud.LoadBalance, error) {
	var secretIDBytes, secretKeyBytes []byte
	var err error
	tmpSecret := &k8scorev1.Secret{}
	err = nc.k8sClient.Get(context.TODO(), k8stypes.NamespacedName{
		Name:      IDKeySecretName,
		Namespace: ns,
	}, tmpSecret)
	if err != nil {
		return nil, fmt.Errorf("get secret %s/%s failed, err %s", IDKeySecretName, ns, err.Error())
	}
	secretIDBase64Bytes, ok := tmpSecret.Data[tencentcloud.EnvNameTencentCloudAccessKeyID]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", tencentcloud.EnvNameTencentCloudAccessKeyID,
			IDKeySecretName, ns)
	}
	secretIDBytes, err = base64.StdEncoding.DecodeString(string(secretIDBase64Bytes))
	if err != nil {
		return nil, fmt.Errorf("decode %s in secret %s/%s failed, err %s", tencentcloud.EnvNameTencentCloudAccessKeyID,
			IDKeySecretName, ns, err)
	}
	secretKeyBase64Bytes, ok := tmpSecret.Data[tencentcloud.EnvNameTencentCloudAccessKey]
	if !ok {
		return nil, fmt.Errorf("lost %s in secret %s/%s", tencentcloud.EnvNameTencentCloudAccessKey,
			IDKeySecretName, ns)
	}
	secretKeyBytes, err = base64.StdEncoding.DecodeString(string(secretKeyBase64Bytes))
	if err != nil {
		return nil, fmt.Errorf("decode %s in secret %s/%s failed, err %s", tencentcloud.EnvNameTencentCloudAccessKey,
			IDKeySecretName, ns, err)
	}
	newClient, err := tencentcloud.NewClbWithSecretIDKey(string(secretIDBytes), string(secretKeyBytes))
	if err != nil {
		return nil, fmt.Errorf("create client for ns %s failed, err %s", ns, err.Error())
	}
	return newClient, nil
}

// get client for certain namespace, if it is not existed, try to create one
func (nc *NamespacedClb) getNsClient(ns string) (cloud.LoadBalance, error) {
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
func (nc *NamespacedClb) DescribeLoadBalancerWithNs(ns, region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	tmpClient, err := nc.getNsClient(ns)
	if err != nil {
		return nil, err
	}
	return tmpClient.DescribeLoadBalancer(region, lbID, name)
}

// DescribeLoadBalancer describe loadbalances with id or name
func (nc *NamespacedClb) DescribeLoadBalancer(region, lbID, name string) (*cloud.LoadBalanceObject, error) {
	return nil, fmt.Errorf("please use DescribeLoadBalancerWithNs for namespaced clb client")
}

// EnsureListener implements LoadBalance interface
func (nc *NamespacedClb) EnsureListener(region string, listener *networkextensionv1.Listener) (string, error) {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return "", err
	}
	return tmpClient.EnsureListener(region, listener)
}

// DeleteListener implements LoadBalance interface
func (nc *NamespacedClb) DeleteListener(region string, listener *networkextensionv1.Listener) error {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return err
	}
	return tmpClient.DeleteListener(region, listener)
}

// EnsureSegmentListener implements LoadBalance interface
func (nc *NamespacedClb) EnsureSegmentListener(region string, listener *networkextensionv1.Listener) (string, error) {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return "", err
	}
	return tmpClient.EnsureSegmentListener(region, listener)
}

// DeleteSegmentListener implements LoadBalance interface
func (nc *NamespacedClb) DeleteSegmentListener(region string, listener *networkextensionv1.Listener) error {
	tmpClient, err := nc.getNsClient(listener.GetNamespace())
	if err != nil {
		return err
	}
	return tmpClient.DeleteSegmentListener(region, listener)
}
