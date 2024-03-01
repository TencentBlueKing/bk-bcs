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

package podpolicy

import (
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller/podpolicy/ipt"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/controller/podpolicy/np"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/datainformer"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/iptables"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/metrics"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-network/bcs-networkpolicy/options"
	docker "github.com/fsouza/go-dockerclient"
	"github.com/prometheus/client_golang/prometheus"
	api "k8s.io/api/core/v1"
	networking "k8s.io/api/networking/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

const (
	needSyncDuration = 2
)

// podPolicyController start by daemonset in every node to handle pod
// network policy, controller switchs to pod network namespace and refresh pod
// iptables. Ipset do not support virtualization, controller only sets
// them in node.
type podPolicyController struct {
	mu sync.Mutex

	needSync bool

	MetricsEnabled bool

	ipSetHandler *iptables.IPSet

	globalSyncPeriod time.Duration

	podEventHandler           cache.ResourceEventHandler
	namespaceEventHandler     cache.ResourceEventHandler
	networkPolicyEventHandler cache.ResourceEventHandler

	dataInformer       datainformer.Interface
	dataInformerSynced bool

	// dockerClient for events
	dockerClient *docker.Client
}

// SetDataInformerSynced used to set the data informer is already synced.
// If data informer not synced, no need to handle resource change event.
func (pc *podPolicyController) SetDataInformerSynced() {
	pc.dataInformerSynced = true
}

// init cleanup all iptables for every container, and sync iptables.
func (pc *podPolicyController) init() error {
	pc.Cleanup()
	if err := pc.Sync(); err != nil {
		return err
	}
	return nil
}

// Run runs forever till we receive notification on stopCh
func (pc *podPolicyController) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) error {
	defer wg.Done()
	if err := pc.init(); err != nil {
		blog.Errorf("Init networkPolicy controller failed.")
		return err
	}
	blog.Info("NetworkPolicy Controller is started.")

	needSyncTicker := time.NewTicker(time.Duration(needSyncDuration) * time.Second)
	defer needSyncTicker.Stop()
	globalSyncTicker := time.NewTicker(pc.globalSyncPeriod)
	defer globalSyncTicker.Stop()

	// loop forever till notified to stop on stopCh
	for {
		select {
		case <-stopCh:
			blog.Info("Shutting down network policies controller")
			pc.Cleanup()
			return nil
		case <-globalSyncTicker.C:
			{
				blog.Infof("Global sync period start syncing.")
				if err := pc.Sync(); err != nil {
					blog.Errorf("Sync failed.")
				}
			}
		case <-needSyncTicker.C:
			// needSync will determine whether need sync in current.
			// in the past time period, needSync should be set true
			// if event updated.
			{
				prevNeedSync := pc.setNeedSync(false)
				if !prevNeedSync {
					continue
				}
				if err := pc.Sync(); err != nil {
					blog.Errorf("Sync failed.")
				}
			}
		}
	}
}

// setNeedSync thread-safe, set whether need-sync and return pev need-sync
// if received event update, set needSync true
// before sync iptables, set needSync false
func (pc *podPolicyController) setNeedSync(needSync bool) (prev bool) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	tmp := pc.needSync
	pc.needSync = needSync
	return tmp
}

// Sync synchronizes iptables to desired state of network policies
func (pc *podPolicyController) Sync() (err error) {
	pc.mu.Lock()
	defer pc.mu.Unlock()

	start := time.Now()
	syncVersion := strconv.FormatInt(start.UnixNano(), 10)
	defer func() {
		endTime := time.Since(start)
		if pc.MetricsEnabled {
			metrics.ControllerPolicyChainsSyncTime.Observe(endTime.Seconds())
		}
		if err != nil {
			metrics.ControllerIPTablesSyncError.Inc()
		}
		blog.Infof("Sync networkPolicy took %v", endTime)
	}()

	networkPolicyInfos, err := np.NewNetworkPolicyHandler(pc.dataInformer).Build()
	if err != nil {
		blog.Errorf("Synced failed, build networkPolicies occurred an error.")
		return err
	}
	blog.Infof("Build network policies successfully, version: %s", syncVersion)

	iptHandler := ipt.NewHandler(pc.dockerClient, pc.ipSetHandler, networkPolicyInfos, syncVersion)
	if err := iptHandler.Refresh(); err != nil {
		blog.Errorf("Sync failed with version: %s.", syncVersion)
		return err
	}
	blog.Infof("Synced successfully.")
	return nil
}

// Cleanup cleanup all ipSets and iptables
func (pc *podPolicyController) Cleanup() {
	iptHandler := ipt.NewCleanupHandler(pc.dockerClient, pc.ipSetHandler)
	if err := iptHandler.Cleanup(); err != nil {
		blog.Errorf("Cleanup occurred an error.")
		return
	}
	blog.Infof("Cleanup successfully.")
}

func (pc *podPolicyController) sendEvent(event resourceEvent) {
	// If data informer not synced already, no need to
	// handle the data event
	if !pc.dataInformerSynced {
		return
	}

	blog.Infof("Received event: %v", event)
	pc.setNeedSync(true)
}

func (pc *podPolicyController) newPodEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pod := obj.(*api.Pod)
			pc.sendEvent(resourceEvent{Type: PodUpdate, Namespace: pod.Namespace, Name: pod.Name})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newPoObj := newObj.(*api.Pod)
			oldPoObj := oldObj.(*api.Pod)

			// for pod change, we are only focused on labels/podStatus/podIP
			if !compareLabels(newPoObj.Labels, oldPoObj.Labels) ||
				newPoObj.Status.Phase != oldPoObj.Status.Phase ||
				newPoObj.Status.PodIP != oldPoObj.Status.PodIP {
				pc.sendEvent(resourceEvent{Type: PodUpdate, Namespace: newPoObj.Namespace, Name: newPoObj.Name})
			}
		},
		DeleteFunc: func(obj interface{}) {
			pod := obj.(*api.Pod)
			pc.sendEvent(resourceEvent{Type: PodUpdate, Namespace: pod.Namespace, Name: pod.Name})
		},
	}
}

func (pc *podPolicyController) newNamespaceEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			pc.sendEvent(resourceEvent{Type: NamespaceUpdate, Namespace: obj.(*api.Namespace).Name})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			newNs := newObj.(*api.Namespace)
			oldNs := oldObj.(*api.Namespace)

			// for namespace change, we are only focused on labels
			if !compareLabels(newNs.Labels, oldNs.Labels) {
				pc.sendEvent(resourceEvent{Type: NamespaceUpdate, Namespace: newNs.Name})
			}
		},
		DeleteFunc: func(obj interface{}) {
			pc.sendEvent(resourceEvent{Type: NamespaceUpdate, Namespace: obj.(*api.Namespace).Name})
		},
	}
}

func (pc *podPolicyController) newNetworkPolicyEventHandler() cache.ResourceEventHandler {
	return cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			np := obj.(*networking.NetworkPolicy)
			pc.sendEvent(resourceEvent{Type: NetworkPolicyUpdate, Namespace: np.Namespace, Name: np.Name})
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			np := newObj.(*networking.NetworkPolicy)
			pc.sendEvent(resourceEvent{Type: NetworkPolicyUpdate, Namespace: np.Namespace, Name: np.Name})
		},
		DeleteFunc: func(obj interface{}) {
			np := obj.(*networking.NetworkPolicy)
			pc.sendEvent(resourceEvent{Type: NetworkPolicyUpdate, Namespace: np.Namespace, Name: np.Name})
		},
	}
}

func compareLabels(m1, m2 map[string]string) bool {
	if len(m1) != len(m2) {
		return false
	}

	isSame := true
	for k, v := range m1 {
		if ov, ok := m2[k]; !ok || ov != v {
			isSame = false
			break
		}
	}
	return isSame
}

// GetPodEventHandler get pod event handler
func (pc *podPolicyController) GetPodEventHandler() cache.ResourceEventHandler {
	return pc.podEventHandler
}

// GetNamespaceEventHandler get namespace event handler
func (pc *podPolicyController) GetNamespaceEventHandler() cache.ResourceEventHandler {
	return pc.namespaceEventHandler
}

// GetNetworkPolicyEventHandler get network policy event handler
func (pc *podPolicyController) GetNetworkPolicyEventHandler() cache.ResourceEventHandler {
	return pc.networkPolicyEventHandler
}

// OnPodUpdate handles updates to pods from the Kubernetes api server
func (pc *podPolicyController) OnPodUpdate(obj interface{}) {}

// OnNetworkPolicyUpdate handles updates to network policy from the kubernetes api server
func (pc *podPolicyController) OnNetworkPolicyUpdate(obj interface{}) {}

// OnNamespaceUpdate handles updates to namespace from kubernetes api server
func (pc *podPolicyController) OnNamespaceUpdate(obj interface{}) {}

// NewPodPolicyController returns new NetworkPolicyController object
// add data informer for pod, namespace and network policy discovery
// add iptables sync error metric
func NewPodPolicyController(
	clientset kubernetes.Interface,
	informer datainformer.Interface,
	config *options.NetworkPolicyOption) (controller.Controller, error) {

	//Register the metrics for this controller
	prometheus.MustRegister(metrics.ControllerIPTablesSyncTime)
	prometheus.MustRegister(metrics.ControllerPolicyChainsSyncTime)
	prometheus.MustRegister(metrics.ControllerIPTablesSyncError)

	ipSetHandler, err := iptables.NewIPSet(false)
	if err != nil {
		return nil, err
	}
	err = ipSetHandler.Save()
	if err != nil {
		return nil, err
	}

	dockerClient, err := docker.NewClient(config.DockerSock)
	if err != nil {
		return nil, err
	}

	ppc := podPolicyController{
		dataInformer:     informer,
		dockerClient:     dockerClient,
		ipSetHandler:     ipSetHandler,
		globalSyncPeriod: time.Duration(config.IPTableSyncPeriod) * time.Second,
	}
	ppc.podEventHandler = ppc.newPodEventHandler()
	ppc.namespaceEventHandler = ppc.newNamespaceEventHandler()
	ppc.networkPolicyEventHandler = ppc.newNetworkPolicyEventHandler()
	return &ppc, nil
}
