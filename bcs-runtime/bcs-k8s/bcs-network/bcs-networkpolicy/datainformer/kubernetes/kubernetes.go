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

package kubernetes

import (
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/options"

	corev1 "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	listercorev1 "k8s.io/client-go/listers/core/v1"
	listernetv1 "k8s.io/client-go/listers/networking/v1"
	"k8s.io/client-go/tools/cache"
)

// KubeInformer data informer for kubernetes
type KubeInformer struct {
	opt               *options.NetworkPolicyOption
	kubeClient        kubernetes.Interface
	informerFactory   informers.SharedInformerFactory
	podInformer       cache.SharedIndexInformer
	podLister         listercorev1.PodLister
	nsInformer        cache.SharedIndexInformer
	nsLister          listercorev1.NamespaceLister
	netPolicyInformer cache.SharedIndexInformer
	netPolicyLister   listernetv1.NetworkPolicyLister
	stopCh            chan struct{}
}

// New create KubeInformer
func New(opt *options.NetworkPolicyOption) *KubeInformer {
	return &KubeInformer{
		opt:    opt,
		stopCh: make(chan struct{}),
	}
}

// Init init informer for k8s
func (ki *KubeInformer) Init(client kubernetes.Interface) {
	ki.kubeClient = client
	informerFactory := informers.NewSharedInformerFactory(client, time.Duration(ki.opt.KubeReSyncPeriod)*time.Second)
	podInformer := informerFactory.Core().V1().Pods()
	nsInformer := informerFactory.Core().V1().Namespaces()
	netPolicyInformer := informerFactory.Networking().V1().NetworkPolicies()
	ki.informerFactory = informerFactory

	ki.podInformer = podInformer.Informer()
	ki.nsInformer = nsInformer.Informer()
	ki.netPolicyInformer = netPolicyInformer.Informer()

	ki.podLister = podInformer.Lister()
	ki.nsLister = nsInformer.Lister()
	ki.netPolicyLister = netPolicyInformer.Lister()
}

// AddPodEventHandler implements DataInformer interface
func (ki *KubeInformer) AddPodEventHandler(handler cache.ResourceEventHandler) {
	ki.podInformer.AddEventHandler(handler)
}

// AddNamespaceEventHandler implements DataInformer interface
func (ki *KubeInformer) AddNamespaceEventHandler(handler cache.ResourceEventHandler) {
	ki.nsInformer.AddEventHandler(handler)
}

// AddNetworkpolicyEventHandler implements DataInformer interface
func (ki *KubeInformer) AddNetworkpolicyEventHandler(handler cache.ResourceEventHandler) {
	ki.netPolicyInformer.AddEventHandler(handler)
}

// Run implements DataInformer interface
func (ki *KubeInformer) Run() error {
	// start informer factory and wait for cache sync, when timeout, return error
	syncFlag := make(chan struct{})
	ki.informerFactory.Start(ki.stopCh)
	go func() {
		blog.Infof("wait for informer factory cache sync")
		ki.informerFactory.WaitForCacheSync(ki.stopCh)
		close(syncFlag)
	}()
	select {
	case <-time.After(time.Duration(ki.opt.KubeCacheSyncTimeout) * time.Second):
		return fmt.Errorf("wait for cache sync timeout after %d seconds", ki.opt.KubeCacheSyncTimeout)
	case <-syncFlag:
		return nil
	}
}

// Stop implements DataInformer interface
func (ki *KubeInformer) Stop() {
	blog.Infof("stop kubernetes informer")
	close(ki.stopCh)
}

// ListAllPods implements DataInformer interface
func (ki *KubeInformer) ListAllPods() ([]*corev1.Pod, error) {
	return ki.podLister.List(labels.Everything())
}

// ListPodsByNamespace implements DataInformer interface
func (ki *KubeInformer) ListPodsByNamespace(ns string, labelsToMatch labels.Set) ([]*corev1.Pod, error) {
	return ki.podLister.Pods(ns).List(labels.SelectorFromSet(labelsToMatch))
}

// ListNamespaces implements DataInformer interface
func (ki *KubeInformer) ListNamespaces(labelsToMatch labels.Set) ([]*corev1.Namespace, error) {
	return ki.nsLister.List(labels.SelectorFromSet(labelsToMatch))
}

// ListAllNetworkPolicy implements DataInformer interface
func (ki *KubeInformer) ListAllNetworkPolicy() ([]*networking.NetworkPolicy, error) {
	return ki.netPolicyLister.List(labels.Everything())
}
