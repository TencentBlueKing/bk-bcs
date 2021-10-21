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

/*
 * [Design]
 * For ordinary Service and Pod,
 * we use Endpoint data to structure AppService,
 * because k8s has helped to select ready Pods;
 * for Statefulset pods, we use pod data directly;
 */

package kubernetes

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	svcclient "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8slabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/informers"
	k8sinfappv1 "k8s.io/client-go/informers/apps/v1"
	k8sinfcorev1 "k8s.io/client-go/informers/core/v1"
	k8scorecliset "k8s.io/client-go/kubernetes"
	k8slistappv1 "k8s.io/client-go/listers/apps/v1"
	k8slistcorev1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	// NSKubeSystem namespace for kubernetes service,
	// event from this namespace shoud be ignored
	NSKubeSystem = "kube-system"

	// event for prometheus
	eventAdd    = "add"
	eventUpdate = "update"
	eventDelete = "delete"
	eventGet    = "get"
	eventList   = "list"
	//state for event
	statusSuccess   = "success"
	statusFailure   = "failure"
	statusNotFinish = "notfinish"
)

var (
	k8sEvent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_k8s_events",
		Help: "Events for k8s service client.",
	}, []string{"type", "event", "status"})
	k8sCritical = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_k8s_critical_err",
		Help: "logic error for k8s service client",
	}, []string{"type", "event"})
)

func init() {
	prometheus.MustRegister(k8sEvent)
	prometheus.MustRegister(k8sCritical)
}

// service handler to deal with event from service informer
type kubeSvcEventHandler struct {
	manager *KubeManager
}

func newKubeSvcEventHandler() *kubeSvcEventHandler {
	return &kubeSvcEventHandler{}
}

func (k *kubeSvcEventHandler) RegisterManager(manager *KubeManager) {
	k.manager = manager
}

// when getting add event and update event for service, call updateAppService
// service add event
func (k *kubeSvcEventHandler) OnAdd(obj interface{}) {
	svc, ok := obj.(*k8scorev1.Service)
	if !ok {
		k8sCritical.WithLabelValues("service", eventAdd).Inc()
		blog.Warnf("k8s Service event handler get unknown type obj %v OnAdd", obj)
		return
	}
	if svc.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(svc.Kind, eventAdd, statusNotFinish).Inc()
		return
	}
	blog.V(3).Infof("k8s Service %s/%s add", svc.GetName(), svc.GetNamespace())
	blog.V(5).Infof("data: %s", svc.String())
	k.manager.updateAppService(svc)
}

// service update event
func (k *kubeSvcEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldSvc, okOld := oldObj.(*k8scorev1.Service)
	if !okOld {
		k8sCritical.WithLabelValues("service", eventUpdate).Inc()
		blog.Warnf("k8s Service event handler get unknown type old obj %v OnUpdate", oldObj)
		return
	}
	newSvc, okNew := newObj.(*k8scorev1.Service)
	if !okNew {
		k8sCritical.WithLabelValues("service", eventUpdate).Inc()
		blog.Warnf("k8s Service event handler get unknown type new obj %v OnUpdate", newObj)
		return
	}
	if newSvc.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(newSvc.Kind, eventUpdate, statusNotFinish).Inc()
		return
	}
	k8sEvent.WithLabelValues(newSvc.Kind, eventUpdate, statusSuccess).Inc()
	blog.V(3).Infof("k8s Service %s/%s update", newSvc.GetName(), newSvc.GetNamespace())
	blog.V(5).Infof("new %s,\n old %s", newSvc.String(), oldSvc.String())
	k.manager.updateAppService(newSvc)
}

// service delete event
func (k *kubeSvcEventHandler) OnDelete(obj interface{}) {
	svc, ok := obj.(*k8scorev1.Service)
	if !ok {
		k8sCritical.WithLabelValues("service", eventDelete).Inc()
		blog.Warnf("k8s Service event handler get unknown type obj %v OnDelete", obj)
		return
	}
	if svc.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(svc.Kind, eventDelete, statusNotFinish).Inc()
		return
	}
	k8sEvent.WithLabelValues(svc.Kind, eventDelete, statusSuccess).Inc()
	blog.V(3).Infof("k8s Service %s/%s delete", svc.GetName(), svc.GetNamespace())
	blog.V(5).Infof("data: %s", svc.String())
	k.manager.deleteAppService(svc)
}

// endpoints event handler
type kubeEpsEventHandler struct {
	manager *KubeManager
}

func newKubeEpsEventHandler() *kubeEpsEventHandler {
	return &kubeEpsEventHandler{}
}

func (k *kubeEpsEventHandler) RegisterManager(manager *KubeManager) {
	k.manager = manager
}

// when getting add event or update event for endpoints, first find Service that the endpoints belongs to,
// then call updateAppService
// endpoint add event
func (k *kubeEpsEventHandler) OnAdd(obj interface{}) {
	eps, ok := obj.(*k8scorev1.Endpoints)
	if !ok {
		k8sCritical.WithLabelValues("service", eventAdd).Inc()
		blog.Warnf("k8s Endpoint event handler get unknown type obj %v add", eps)
		return
	}
	if eps.GetNamespace() == "kube-system" {
		k8sEvent.WithLabelValues(eps.Kind, eventAdd, statusNotFinish).Inc()
		return
	}
	svcKey := fmt.Sprintf("%s/%s", eps.GetNamespace(), eps.GetName())
	blog.V(3).Infof("k8s Endpoint %s add", svcKey)
	blog.V(5).Infof("data: %s", eps.String())
	svc, err := k.manager.svcsLister.Services(eps.GetNamespace()).Get(eps.GetName())
	if err != nil {
		k8sEvent.WithLabelValues(eps.Kind, eventAdd, statusFailure).Inc()
		blog.V(4).Infof("get k8s Service from cache by key %s from Endpoints %s failed, err %s",
			svcKey, eps.String(), err.Error())
		return
	}
	k.manager.updateAppService(svc)
	k8sEvent.WithLabelValues(eps.Kind, eventAdd, statusSuccess).Inc()
}

// endpoints update event
func (k *kubeEpsEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldEps, okOld := oldObj.(*k8scorev1.Endpoints)
	if !okOld {
		k8sCritical.WithLabelValues("endpoints", eventUpdate).Inc()
		blog.Warnf("kube Endpoints handler get unknown type old obj %v OnUpdate", oldObj)
		return
	}
	newEps, okNew := newObj.(*k8scorev1.Endpoints)
	if !okNew {
		k8sCritical.WithLabelValues("endpoints", eventUpdate).Inc()
		blog.Warnf("kube Endpoints handler get unknown type new obj %v OnUpdate", newObj)
		return
	}
	if newEps.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(newEps.Kind, eventUpdate, statusNotFinish).Inc()
		return
	}
	if reflect.DeepEqual(oldObj, newObj) {
		k8sEvent.WithLabelValues(newEps.Kind, eventUpdate, statusNotFinish).Inc()
		blog.V(5).Infof("kube Endpoints obj has no changes, skip update event\n%v", newObj)
		return
	}
	svcKey := fmt.Sprintf("%s/%s", newEps.GetNamespace(), newEps.GetName())
	blog.V(3).Infof("k8s Endpoints %s update", svcKey)
	blog.V(5).Infof("old %s,\n new %s", oldEps.String(), newEps.String())
	svc, err := k.manager.svcsLister.Services(newEps.GetNamespace()).Get(newEps.GetName())
	if err != nil {
		k8sEvent.WithLabelValues(newEps.Kind, eventUpdate, statusFailure).Inc()
		blog.V(4).Infof("get k8s Service from cache by key %s from Endpoints %s failed, err %s",
			svcKey, newEps.String(), err.Error())
		return
	}
	k.manager.updateAppService(svc)
	k8sEvent.WithLabelValues(newEps.Kind, eventUpdate, statusSuccess).Inc()
}

// endpoints delete event
func (k *kubeEpsEventHandler) OnDelete(obj interface{}) {
	eps, ok := obj.(*k8scorev1.Endpoints)
	if !ok {
		blog.Warnf("k8s Endpoint event handler get unknown type obj %v delete", eps)
		return
	}
	// skip kube-system
	if eps.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(eps.Kind, eventDelete, statusNotFinish).Inc()
		return
	}
	svcKey := fmt.Sprintf("%s/%s", eps.GetNamespace(), eps.GetName())
	blog.V(3).Infof("k8s Endpoints %s delete", svcKey)
	blog.V(5).Infof("data: %s", eps.String())
	svc, err := k.manager.svcsLister.Services(eps.GetNamespace()).Get(eps.GetName())
	if err != nil {
		k8sEvent.WithLabelValues(eps.Kind, eventDelete, statusFailure).Inc()
		blog.V(4).Infof("get k8s Service from cache by key %s from Endpoints %s failed, err %s",
			svcKey, eps.String(), err.Error())
		return
	}
	k.manager.updateAppService(svc)
	k8sEvent.WithLabelValues(eps.Kind, eventDelete, statusSuccess).Inc()
}

type kubePodEventHandler struct {
	manager *KubeManager
}

func newKubePodEventHandler() *kubePodEventHandler {
	return &kubePodEventHandler{}
}

func (k *kubePodEventHandler) RegisterManager(manager *KubeManager) {
	k.manager = manager
}

// pod add
func (k *kubePodEventHandler) OnAdd(obj interface{}) {
	pod, ok := obj.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("k8s Pod event handler get unknown type obj %v OnAdd", obj)
		return
	}
	if pod.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(pod.Kind, eventAdd, statusNotFinish).Inc()
		return
	}
	k8sEvent.WithLabelValues(pod.Kind, eventAdd, statusSuccess).Inc()
	blog.V(5).Infof("k8s Pod add\n%s", pod.String())
}

// pod update
func (k *kubePodEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldPod, okOld := oldObj.(*k8scorev1.Pod)
	if !okOld {
		blog.Warnf("k8s Pod event handler get unknown type old obj %v OnUpdate", oldObj)
		return
	}
	newPod, okNew := newObj.(*k8scorev1.Pod)
	if !okNew {
		blog.Warnf("k8s Pod event handler get unknown type new obj %v OnUpdate", newObj)
		return
	}
	if newPod.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(newPod.Kind, eventUpdate, statusNotFinish).Inc()
		return
	}
	if reflect.DeepEqual(oldPod.Spec, newPod.Spec) {
		k8sEvent.WithLabelValues(newPod.Kind, eventUpdate, statusFailure).Inc()
		blog.V(5).Infof("k8s Pod has no changes, skip update event\n%s", newPod.String())
		return
	}
	k8sEvent.WithLabelValues(newPod.Kind, eventUpdate, statusSuccess).Inc()
	blog.V(5).Infof("k8s Pod update\nnew %s,\n old %s", newPod.String(), oldPod.String())
}

// pod delete
func (k *kubePodEventHandler) OnDelete(obj interface{}) {
	pod, ok := obj.(*k8scorev1.Pod)
	if !ok {
		blog.Warnf("k8s Service event handler get unknown type obj %v OnDelete", obj)
		return
	}
	// skip kube-system
	if pod.GetNamespace() == NSKubeSystem {
		k8sEvent.WithLabelValues(pod.Kind, eventDelete, statusNotFinish).Inc()
		return
	}
	k8sEvent.WithLabelValues(pod.Kind, eventDelete, statusSuccess).Inc()
	blog.V(5).Infof("k8s Pod delete\n %s", pod.String())
}

// NewClient create internal service client
func NewClient(config string, handler cache.ResourceEventHandler, syncPeriod time.Duration) (svcclient.Client, error) {
	var restConfig *rest.Config
	var err error
	// use incluster config by default
	if len(config) == 0 {
		blog.Infof("use in-cluster kube config")
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("get incluster config failed, err %s", err.Error())
			return nil, err
		}
	} else {
		//parse configuration
		restConfig, err = clientcmd.BuildConfigFromFlags("", config)
		if err != nil {
			blog.Errorf("create internal client with kubeconfig %s failed, err %s", config, err.Error())
			return nil, err
		}
	}
	//initialize specified client implementation
	cliset, err := k8scorecliset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("create clientset failed, with rest config %v, err %s", restConfig, err.Error())
		return nil, err
	}
	blog.Infof("start create informer factory")
	factory := informers.NewSharedInformerFactory(cliset, syncPeriod)
	// informer and lister for k8s service
	svcsInformer := factory.Core().V1().Services()
	svcsLister := svcsInformer.Lister()
	// informer and lister for k8s endpoint
	epsInformer := factory.Core().V1().Endpoints()
	epsLister := epsInformer.Lister()
	// informer and lister for k8s pods
	podsInformer := factory.Core().V1().Pods()
	podsLister := podsInformer.Lister()
	// informer and lister for k8s statefulset
	statefulSetInformer := factory.Apps().V1().StatefulSets()
	statefulSetLister := statefulSetInformer.Lister()
	blog.Infof("create AppService cache")
	// cache for app service
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	//create informer, cache, reflector
	kube := &KubeManager{
		factory:             factory,
		svcsInformer:        svcsInformer,
		epsInformer:         epsInformer,
		podsInformer:        podsInformer,
		podsLister:          podsLister,
		svcsLister:          svcsLister,
		epsLister:           epsLister,
		statefulSetInformer: statefulSetInformer,
		statefulSetLister:   statefulSetLister,
		appServiceCache:     store,
		appServiceHandler:   handler,
		stopCh:              make(chan struct{}),
	}

	// start sync data to local cache
	factory.Start(kube.stopCh)
	results := factory.WaitForCacheSync(kube.stopCh)
	for key, value := range results {
		blog.Infof("MesosManager Wait For Cache %s Sync, result: %v", key, value)
	}
	blog.Infof("wait for cache sync successfully")

	// register handler
	svcHandler := newKubeSvcEventHandler()
	svcHandler.RegisterManager(kube)
	epsHandler := newKubeEpsEventHandler()
	epsHandler.RegisterManager(kube)
	podHandler := newKubePodEventHandler()
	podHandler.RegisterManager(kube)
	// add handler to informer to get events
	svcsInformer.Informer().AddEventHandler(svcHandler)
	epsInformer.Informer().AddEventHandler(epsHandler)
	podsInformer.Informer().AddEventHandler(podHandler)

	return kube, nil
}

// KubeManager manage kubernetes Service/Pod info AppService discovery
type KubeManager struct {
	factory             informers.SharedInformerFactory
	svcsInformer        k8sinfcorev1.ServiceInformer
	svcsLister          k8slistcorev1.ServiceLister
	epsInformer         k8sinfcorev1.EndpointsInformer
	epsLister           k8slistcorev1.EndpointsLister
	podsInformer        k8sinfcorev1.PodInformer
	podsLister          k8slistcorev1.PodLister
	statefulSetInformer k8sinfappv1.StatefulSetInformer
	statefulSetLister   k8slistappv1.StatefulSetLister
	appServiceCache     cache.Store
	appServiceHandler   cache.ResourceEventHandler
	stopCh              chan struct{}
}

// GetAppService get service by specified name
func (m *KubeManager) GetAppService(namespace, name string) (*svcclient.AppService, error) {
	key := fmt.Sprintf("%s/%s", namespace, name)
	obj, isExisted, err := m.appServiceCache.GetByKey(key)
	if err != nil {
		k8sCritical.WithLabelValues("appservice", eventGet).Inc()
		blog.Errorf("get AppService by key %s from cache failed, err %s", key, err.Error())
		return nil, fmt.Errorf("get AppService by key %s from cache failed, err %s", key, err.Error())
	}
	if !isExisted {
		blog.Warnf("get no AppService by key %s", key)
		return nil, fmt.Errorf("get no AppService by key %s", key)
	}
	appSvc, ok := obj.(*svcclient.AppService)
	if !ok {
		k8sCritical.WithLabelValues("appservice", eventGet).Inc()
		blog.Errorf("get obj %v from cache by key %s is not type AppService", obj, key)
		return nil, fmt.Errorf("get obj %v from cache by key %s is not type AppService", obj, key)
	}
	return appSvc, nil
}

// ListAppService list all service in cache
// TODO: list all AppService from AppService cache
func (m *KubeManager) ListAppService(labels map[string]string) ([]*svcclient.AppService, error) {
	var selector k8slabels.Selector
	if len(labels) == 0 || labels == nil {
		selector = k8slabels.Everything()
	} else {
		set := k8slabels.Set(labels)
		selector = set.AsSelector()
	}
	svcs, err := m.svcsInformer.Lister().List(selector)
	if err != nil {
		k8sCritical.WithLabelValues("appservice", eventList).Inc()
		blog.Errorf("KubeManager list all k8s Services by selector %s failed, err %s", selector.String(), err.Error())
		return nil, fmt.Errorf("KubeManager list all k8s Services by selector %s failed, err %s",
			selector.String(), err.Error())
	}
	if len(svcs) == 0 {
		blog.Warnf("KubeManager list no k8s Services by selector %s", selector.String())
		return nil, nil
	}
	var internalAppSvcs []*svcclient.AppService
	for _, svc := range svcs {
		if len(svc.Spec.Selector) == 0 {
			blog.Warnf("KubeManager get empty selector for k8s Service %s, skip", svc.String())
			continue
		}
		eps, err := m.getEpsByService(svc)
		if err != nil {
			blog.Warnf("KubeManager get eps by service %s failed, err %s", svc.String(), err.Error())
			continue
		}
		if eps == nil {
			continue
		}
		localAppSvc, err := m.convert(svc, eps)
		if err != nil {
			blog.Warnf("KubeManager convert svc %s and eps %s to AppService failed, err %s",
				svc.String(), eps.String(), err.Error())
			continue
		}
		internalAppSvcs = append(internalAppSvcs, localAppSvc)
		blog.V(5).Infof("list get AppService %v", localAppSvc)
	}
	return internalAppSvcs, nil
}

// ListAppServiceFromStatefulSet list app service from stateful set by service name
// generate a AppService object for each statefulset pod
func (m *KubeManager) ListAppServiceFromStatefulSet(ns, name string) ([]*svcclient.AppService, error) {
	statefulSets, err := m.statefulSetLister.StatefulSets(ns).List(k8slabels.Everything())
	if err != nil {
		blog.Warnf("list all stateful set failed, err %s", err.Error())
		return make([]*svcclient.AppService, 0), nil
	}
	for _, set := range statefulSets {
		if set.Spec.ServiceName == name {
			pods, err := m.podsLister.Pods(ns).List(k8slabels.Set(set.Spec.Selector.MatchLabels).AsSelector())
			if err != nil {
				return nil, fmt.Errorf("failed to find pods by set labelSelector %v, err %s",
					set.Spec.Selector, err.Error())
			}
			// if statefulset has no pod, do not return error
			if len(pods) == 0 {
				blog.Warnf("stateful %s has no pod")
				return make([]*svcclient.AppService, 0), nil
			}
			// for the pod of stateful set, port must be created in order.
			sortStatefulSetPod(pods)
			appSvcList, err := m.convertStatefulSet(pods)
			if err != nil {
				return nil, fmt.Errorf("convert stateful set to endpoints to AppServices failed, err %s", err.Error())
			}
			return appSvcList, nil
		}
	}
	blog.Warnf("can not find statefulSet with serviceName %s in ns %s", name, ns)
	return make([]*svcclient.AppService, 0), nil
}

// get index from k8s statefulset pod name
// example test-0, test-1
func getIndexFromStatefulSetName(name string) (int, error) {
	nameStrs := strings.Split(name, "-")
	podNumberStr := nameStrs[len(nameStrs)-1]
	podIndex, err := strconv.Atoi(podNumberStr)
	if err != nil {
		blog.Errorf("get stateful set pod index failed from podName %s, err %s", name, err.Error())
		return -1, fmt.Errorf("get stateful set pod index failed from podName %s, err %s", name, err.Error())
	}
	return podIndex, nil
}

// sort statefulset pod by pod name
// we should always keep statefulset pods in order
func sortStatefulSetPod(pods []*k8scorev1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		indexI, _ := getIndexFromStatefulSetName(pods[i].GetName())
		indexJ, _ := getIndexFromStatefulSetName(pods[j].GetName())
		return indexI < indexJ
	})
}

// pod is ready when all the containers in pod is ready
func checkPodReady(pod *k8scorev1.Pod) bool {
	if len(pod.Spec.Containers) != len(pod.Status.ContainerStatuses) {
		return false
	}
	for _, containerStatus := range pod.Status.ContainerStatuses {
		if !containerStatus.Ready {
			return false
		}
	}
	return true
}

func (m *KubeManager) convertStatefulSet(pods []*k8scorev1.Pod) ([]*svcclient.AppService, error) {
	var appServiceList []*svcclient.AppService
	if len(pods) == 0 {
		blog.Errorf("stateful to be converted set has no pods")
		return nil, fmt.Errorf("stateful to be converted set has no pods")
	}
	for _, pod := range pods {
		newAppSvc := &svcclient.AppService{
			TypeMeta: k8smetav1.TypeMeta{
				Kind:       "AppService",
				APIVersion: pod.APIVersion,
			},
			ObjectMeta:   pod.ObjectMeta,
			Frontend:     make([]string, 0),
			ServicePorts: make([]svcclient.ServicePort, 0),
			Nodes:        make([]svcclient.AppNode, 0),
		}
		podName := pod.GetName()
		nameStrs := strings.Split(podName, "-")
		podNumberStr := nameStrs[len(nameStrs)-1]
		podIndex, err := strconv.Atoi(podNumberStr)
		if err != nil {
			blog.Errorf("get stateful set pod index failed from podName %s, err %s", podName, err.Error())
			return nil, fmt.Errorf("get stateful set pod index failed from podName %s, err %s", podName, err.Error())
		}
		// use pod index to set service port and node port
		newServicePort := svcclient.ServicePort{
			Name:        pod.GetName(),
			ServicePort: podIndex,
			TargetPort:  podIndex,
		}
		newPort := svcclient.NodePort{
			Name:     podName + strconv.Itoa(podIndex),
			Protocol: "",
			NodePort: podIndex,
		}
		newAppNode := svcclient.AppNode{
			TypeMeta:   pod.TypeMeta,
			ObjectMeta: pod.ObjectMeta,
			Index:      pod.GetName(),
			Weight:     10,
			NodeIP:     pod.Status.PodIP,
			ProxyIP:    pod.Status.HostIP,
			Ports:      make([]svcclient.NodePort, 0),
		}
		newAppNode.Ports = append(newAppNode.Ports, newPort)
		// if pod is not ready, we don't add it into AppService struct
		if checkPodReady(pod) {
			newAppSvc.Nodes = append(newAppSvc.Nodes, newAppNode)
		}
		newAppSvc.ServicePorts = append(newAppSvc.ServicePorts, newServicePort)
		appServiceList = append(appServiceList, newAppSvc)
	}

	return appServiceList, nil
}

// getPodByService get all pod by specified k8s Service
func (m *KubeManager) getEpsByService(svc *k8scorev1.Service) (*k8scorev1.Endpoints, error) {
	key := fmt.Sprintf("%s/%s", svc.GetNamespace(), svc.GetName())
	eps, err := m.epsLister.Endpoints(svc.GetNamespace()).Get(svc.GetName())
	if err != nil {
		k8sCritical.WithLabelValues("endpoints", eventGet).Inc()
		blog.Errorf("get eps by key %s failed, err %s", key, err.Error())
		return nil, fmt.Errorf("get eps by key %s failed, err %s", key, err.Error())
	}
	return eps, nil
}

func (m *KubeManager) getPod(ns, name string) (*k8scorev1.Pod, error) {
	key := fmt.Sprintf("%s/%s", ns, name)
	pod, err := m.podsLister.Pods(ns).Get(name)
	if err != nil {
		k8sCritical.WithLabelValues("pod", eventGet).Inc()
		blog.Errorf("get pod by key %s failed, err %s", key, err.Error())
		return nil, fmt.Errorf("get pod by key %s failed, err %s", key, err.Error())
	}
	return pod, nil
}

// convert k8s service and corresponding endpoints into AppService struct
func (m *KubeManager) convert(svc *k8scorev1.Service, eps *k8scorev1.Endpoints) (*svcclient.AppService, error) {
	localAppSvc := &svcclient.AppService{
		TypeMeta:     svc.TypeMeta,
		ObjectMeta:   svc.ObjectMeta,
		Frontend:     []string{svc.Spec.ClusterIP},
		ServicePorts: make([]svcclient.ServicePort, 0),
		Nodes:        make([]svcclient.AppNode, 0),
	}
	for _, k8sSvcPort := range svc.Spec.Ports {
		newPort := svcclient.ServicePort{
			Name:        k8sSvcPort.Name,
			Protocol:    string(k8sSvcPort.Protocol),
			ServicePort: int(k8sSvcPort.Port),
			ProxyPort:   int(k8sSvcPort.NodePort),
			TargetPort:  k8sSvcPort.TargetPort.IntValue(),
		}
		localAppSvc.ServicePorts = append(localAppSvc.ServicePorts, newPort)
	}
	if eps != nil {
		if len(eps.Subsets) == 0 {
			blog.V(4).Infof("converted AppService %s has no AppNode", localAppSvc.ObjectMeta.String())
			return localAppSvc, nil
		}
		for _, subset := range eps.Subsets {
			for _, addr := range subset.Addresses {
				if addr.TargetRef != nil {
					// get pod info from endpoint addresses
					pod, err := m.getPod(addr.TargetRef.Namespace, addr.TargetRef.Name)
					if err != nil {
						blog.Errorf("get pod by ns %s name %s failed, err %s",
							addr.TargetRef.Namespace, addr.TargetRef.Name, err.Error())
						continue
					}
					if pod == nil {
						blog.Errorf("get no pod by ns %s name %s",
							addr.TargetRef.Namespace, addr.TargetRef.Name)
						continue
					}
					newAppNode := svcclient.AppNode{
						TypeMeta:   pod.TypeMeta,
						ObjectMeta: pod.ObjectMeta,
						Index:      pod.GetName(),
						//TODO: to set weight for different app in same service
						Weight:  10,
						NodeIP:  addr.IP,
						ProxyIP: pod.Status.HostIP,
						Ports:   make([]svcclient.NodePort, 0),
					}
					for _, port := range subset.Ports {
						newPort := svcclient.NodePort{
							Name:     port.Name,
							Protocol: string(port.Protocol),
							NodePort: int(port.Port),
						}
						newAppNode.Ports = append(newAppNode.Ports, newPort)
					}
					localAppSvc.Nodes = append(localAppSvc.Nodes, newAppNode)
				}
			}
		}
	}
	return localAppSvc, nil
}

// updateAppService update AppService according to k8s service
// 1. get endpoints from cache according the k8s Service
// 2. convert k8s service and endpoints into AppService struct
// 3. update cache
func (m *KubeManager) updateAppService(svc *k8scorev1.Service) {
	eps, err := m.getEpsByService(svc)
	if err != nil {
		blog.V(4).Infof("get endpoints by svc %s failed, err %s", svc.String(), err.Error())
		return
	}
	if eps == nil {
		blog.Warnf("Get no endpoints for service %s/%s", svc.GetNamespace(), svc.GetName())
		k8sEvent.WithLabelValues("appservice", eventUpdate, statusNotFinish).Inc()
		return
	}
	newAppService, err := m.convert(svc, eps)
	if err != nil {
		blog.Warnf("convert svc %s with eps %v to AppService failed, err %s", svc, eps, err.Error())
		return
	}
	oldAppService, isExisted, err := m.appServiceCache.Get(newAppService)
	if err != nil {
		k8sCritical.WithLabelValues("appservice", eventUpdate)
		blog.V(4).Infof("get AppService from cache by svc %v failed, err %s", newAppService, err.Error())
		return
	}
	if !isExisted {
		err = m.appServiceCache.Add(newAppService)
		if err != nil {
			k8sCritical.WithLabelValues("appservice", eventUpdate)
			blog.Warnf("add AppService %v to cache failed, err %s", newAppService, err.Error())
			return
		}
		blog.V(4).Infof("add AppService %v to cache successfully", newAppService)
		m.appServiceHandler.OnAdd(newAppService)
		k8sEvent.WithLabelValues("appservice", eventUpdate, statusSuccess).Inc()
		return
	}
	// if new AppService is equal to old AppService, just return
	if reflect.DeepEqual(newAppService, oldAppService) {
		k8sEvent.WithLabelValues("appservice", eventUpdate, statusNotFinish).Inc()
		blog.V(4).Infof("new AppService %v is equal to the old, no need to update", newAppService)
		return
	}
	err = m.appServiceCache.Update(newAppService)
	if err != nil {
		k8sCritical.WithLabelValues("appservice", eventUpdate).Inc()
		blog.Warnf("update AppService %v to cache failed, err %s", newAppService, err.Error())
		return
	}
	k8sEvent.WithLabelValues("appservice", eventUpdate, statusSuccess).Inc()
	blog.V(4).Infof("update AppService %s/%s to cache successfully",
		newAppService.GetNamespace(), newAppService.GetName())
	m.appServiceHandler.OnUpdate(oldAppService, newAppService)
}

// delete AppService from AppService cache
// do nothing when AppService is not in cache
func (m *KubeManager) deleteAppService(svc *k8scorev1.Service) {
	key := fmt.Sprintf("%s/%s", svc.GetNamespace(), svc.GetName())
	oldAppServiceObj, isExisted, err := m.appServiceCache.GetByKey(key)
	if err != nil {
		k8sCritical.WithLabelValues("appservice", eventDelete).Inc()
		blog.V(4).Infof("get AppService by key %s from cache failed, err %s", key, err.Error())
		return
	}
	if !isExisted {
		k8sEvent.WithLabelValues("appservice", eventDelete, statusSuccess).Inc()
		blog.V(4).Infof("no old AppService with key %s in cache, no need to delete", key)
		return
	}
	oldAppService, ok := oldAppServiceObj.(*svcclient.AppService)
	if !ok {
		blog.V(4).Infof("get old obj from cache:\n %v is not type AppService", oldAppServiceObj)
		return
	}
	err = m.appServiceCache.Delete(oldAppService)
	if err != nil {
		k8sEvent.WithLabelValues("appservice", eventDelete, statusFailure).Inc()
		blog.V(4).Infof("delete AppService %v in cache failed, err %s", oldAppService, err.Error())
		return
	}
	k8sEvent.WithLabelValues("appservice", eventDelete, statusSuccess).Inc()
	blog.V(5).Infof("delete AppService %s%s from cache successfully",
		oldAppService.GetNamespace(), oldAppService.GetName())
}

// Close client, clean resource
func (m *KubeManager) Close() {
	close(m.stopCh)
}
