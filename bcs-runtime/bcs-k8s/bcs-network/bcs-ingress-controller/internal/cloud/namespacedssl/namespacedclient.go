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

package namespacedssl

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8scorev1 "k8s.io/api/core/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/internal/cloud/tencentcloud"
	networkextensionv1 "github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1"
)

const (
	// IDKeySecretName secret name to store secret id and secret key.
	IDKeySecretName = "ingress-secret.networkextension.bkbcs.tencent.com"
	// IDKeyConfigName config name to store controller config, include cloud secret.
	IDKeyConfigName = "ingress-config.networkextension.bkbcs.tencent.com"
)

// NamespacedSSL routes SSL client credentials by namespace.
type NamespacedSSL struct {
	ctx context.Context

	k8sClient client.Client

	nsClientSet map[string]tencentcloud.SSLClient

	ncClientResourceVersionMap map[string]string

	newSSLFunc func(map[string][]byte) (tencentcloud.SSLClient, error)

	defaultClient tencentcloud.SSLClient

	exemptNamespaces map[string]struct{}

	clientLock sync.Mutex
}

// NewNamespacedSSLForTest constructs NamespacedSSL without background reload for tests.
func NewNamespacedSSLForTest(k8sClient client.Client,
	newSSLFunc func(map[string][]byte) (tencentcloud.SSLClient, error),
	defaultClient tencentcloud.SSLClient, exempt map[string]struct{}) *NamespacedSSL {
	return &NamespacedSSL{
		k8sClient:                  k8sClient,
		nsClientSet:                make(map[string]tencentcloud.SSLClient),
		ncClientResourceVersionMap: make(map[string]string),
		newSSLFunc:                 newSSLFunc,
		defaultClient:              defaultClient,
		exemptNamespaces:           exempt,
	}
}

// NewNamespacedSSL creates a namespaced SSL client.
// defaultClient and exemptNamespaces are optional; when both are provided, namespaces in
// exemptNamespaces reuse defaultClient directly (built from the controller's global credentials)
// and skip the per-namespace secret/controllerconfig lookup. Passing a nil defaultClient OR a
// nil/empty exemptNamespaces keeps the original per-namespace behavior.
func NewNamespacedSSL(ctx context.Context, k8sClient client.Client,
	defaultClient tencentcloud.SSLClient, exemptNamespaces map[string]struct{}) *NamespacedSSL {
	ns := &NamespacedSSL{
		ctx:                        ctx,
		k8sClient:                  k8sClient,
		nsClientSet:                make(map[string]tencentcloud.SSLClient),
		ncClientResourceVersionMap: make(map[string]string),
		newSSLFunc:                 tencentcloud.NewSSLClientWithSecret,
		defaultClient:              defaultClient,
		exemptNamespaces:           exemptNamespaces,
	}
	go ns.reloadLoop()
	return ns
}

func (nc *NamespacedSSL) reloadLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			nc.reloadNsClient()
		case <-nc.ctx.Done():
			return
		}
	}
}

func (nc *NamespacedSSL) isExempt(ns string) bool {
	if nc.defaultClient == nil || len(nc.exemptNamespaces) == 0 {
		return false
	}
	_, ok := nc.exemptNamespaces[ns]
	return ok
}

func (nc *NamespacedSSL) reloadNsClient() {
	nc.clientLock.Lock()
	defer nc.clientLock.Unlock()
	for ns := range nc.nsClientSet {
		if nc.isExempt(ns) {
			continue
		}
		cloudSecret, resourceVersion, err := nc.getNsSecret(ns)
		if err != nil {
			blog.Errorf("get namespace[%s] cloud secret failed, err: %s", ns, err.Error())
			continue
		}
		if resourceVersion == nc.ncClientResourceVersionMap[ns] {
			continue
		}
		newClient, err := nc.newSSLFunc(cloudSecret)
		if err != nil {
			blog.Errorf("create ssl client for namespace [%s] failed, err %s", ns, err.Error())
			continue
		}
		nc.nsClientSet[ns] = newClient
		nc.ncClientResourceVersionMap[ns] = resourceVersion
		blog.Infof("namespace[%s] cloud secret changed, reload ssl client", ns)
	}
}

func (nc *NamespacedSSL) getNsSecret(ns string) (map[string][]byte, string, error) {
	tmpSecret := &k8scorev1.Secret{}
	err := nc.k8sClient.Get(context.TODO(), k8stypes.NamespacedName{
		Name:      IDKeySecretName,
		Namespace: ns,
	}, tmpSecret)
	foundSecret := true
	if err != nil {
		foundSecret = false
		if !k8serrors.IsNotFound(err) {
			return nil, "", fmt.Errorf("get secret %s/%s failed, err %s", IDKeySecretName, ns, err.Error())
		}
	}

	controllerConfig := &networkextensionv1.ControllerConfig{}
	err = nc.k8sClient.Get(context.TODO(), k8stypes.NamespacedName{
		Name:      IDKeyConfigName,
		Namespace: ns,
	}, controllerConfig)
	foundConfig := true
	if err != nil {
		foundConfig = false
		if !k8serrors.IsNotFound(err) {
			return nil, "", fmt.Errorf("get controller config %s/%s failed, err %s", IDKeyConfigName, ns, err.Error())
		}
	}

	if foundSecret {
		return tmpSecret.Data, tmpSecret.GetResourceVersion(), nil
	}
	if foundConfig {
		return controllerConfig.Spec.Secret, controllerConfig.GetResourceVersion(), nil
	}
	return nil, "", fmt.Errorf("not found secret or controllerConfig in namespace %s, "+
		"please create secret '%s' or controllerConfig %s in namespace", ns, IDKeySecretName, IDKeyConfigName)
}

func (nc *NamespacedSSL) initNsClient(ns string) (tencentcloud.SSLClient, string, error) {
	cloudSecret, resourceVersion, err := nc.getNsSecret(ns)
	if err != nil {
		return nil, "", err
	}
	newClient, err := nc.newSSLFunc(cloudSecret)
	if err != nil {
		return nil, "", fmt.Errorf("create ssl client for ns %s failed, err %s", ns, err.Error())
	}
	return newClient, resourceVersion, nil
}

// GetNsClient returns the SSL client for the given namespace.
func (nc *NamespacedSSL) GetNsClient(ns string) (tencentcloud.SSLClient, error) {
	if nc.isExempt(ns) {
		return nc.defaultClient, nil
	}
	tmpClient, ok := nc.nsClientSet[ns]
	if ok {
		return tmpClient, nil
	}
	newClient, resourceVersion, err := nc.initNsClient(ns)
	if err != nil {
		return nil, err
	}
	nc.clientLock.Lock()
	nc.nsClientSet[ns] = newClient
	nc.ncClientResourceVersionMap[ns] = resourceVersion
	nc.clientLock.Unlock()
	return newClient, nil
}
