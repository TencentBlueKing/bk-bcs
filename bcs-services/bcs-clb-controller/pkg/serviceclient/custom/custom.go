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

package custom

import (
	"fmt"
	"reflect"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	v1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/apis/mesh/v1"
	informers "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/informers/externalversions"
	informermeshv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/informers/externalversions/mesh/v1"
	internalclientset "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/clientset/versioned"
	listerv1 "github.com/Tencent/bk-bcs/bcs-k8s/kubedeprecated/generated/listers/mesh/v1"
	svcclient "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

	"github.com/prometheus/client_golang/prometheus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	// event for prometheus
	eventAdd    = "add"
	eventUpdate = "update"
	eventDelete = "delete"
	eventGet    = "get"
	eventList   = "list"
	// type for custom service client
	typeAppSvc     = "appsvc"
	typeAppNode    = "appnode"
	typeAppService = "appservice"
	//state for event
	statusSuccess   = "success"
	statusFailure   = "failure"
	statusNotFinish = "notfinish"
)

var (
	// metric for event
	customEvent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_custom_events",
		Help: "Events for custom service client.",
	}, []string{"type", "event", "status"})
	// metric for critical err
	customCritical = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_custom_critical_err",
		Help: "logic error for custom service client(mesos adapter)",
	}, []string{"type", "event"})
)

func init() {
	prometheus.MustRegister(customEvent)
	prometheus.MustRegister(customCritical)
}

// handler for AppSvc event
type innerAppSvcEventHandler struct {
	manager *CustomizedManager
}

func newInnerAppSvcEventHandler() *innerAppSvcEventHandler {
	return &innerAppSvcEventHandler{}
}

func (h *innerAppSvcEventHandler) RegisterManager(manager *CustomizedManager) {
	h.manager = manager
}

// AppSvc add event
// when AppSvc add or update, do update AppService in cache
func (h *innerAppSvcEventHandler) OnAdd(obj interface{}) {
	svc, ok := obj.(*v1.AppSvc)
	if !ok {
		blog.Warnf("AppSvc event handler get unknown type obj %v OnAdd", obj)
		return
	}
	blog.Infof("AppSvc %s/%s add, ready to refresh", svc.GetNamespace(), svc.GetName())
	h.manager.updateAppService(svc)
}

// AppSvc update event
func (h *innerAppSvcEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldSvc, okOld := oldObj.(*v1.AppSvc)
	if !okOld {
		blog.Warnf("AppSvc event handler get unknown type oldObj %v OnUpdate", oldObj)
		return
	}
	newSvc, okNew := newObj.(*v1.AppSvc)
	if !okNew {
		blog.Warnf("AppSvc event handler get unknown type newObj %v OnUpdate", newObj)
		return
	}
	//protection: bcs-scheduler will update timestamp every 3 minutes
	//            event if nothing is changed
	newSvc.SetCreationTimestamp(oldSvc.GetCreationTimestamp())
	if reflect.DeepEqual(oldObj, newObj) {
		customEvent.WithLabelValues(typeAppSvc, eventAdd, statusNotFinish).Inc()
		blog.Warnf("AppSvc %s/%s No changed, skip update event", newSvc.GetNamespace(), newSvc.GetName())
		return
	}
	blog.Infof("AppSvc %s/%s update, ready to refresh", newSvc.GetNamespace(), newSvc.GetName())
	h.manager.updateAppService(newSvc)
	return
}

// AppSvc delete event
func (h *innerAppSvcEventHandler) OnDelete(obj interface{}) {
	svc, ok := obj.(*v1.AppSvc)
	if !ok {
		blog.Warnf("AppSvc event handler get unknown type obj %v OnDelete", obj)
		return
	}
	blog.Infof("AppSvc %s/%s delete, ready to refresh", svc.GetNamespace(), svc.GetName())
	h.manager.deleteAppService(svc)
	return
}

// handler for AppNode event
type innerAppNodeEventHandler struct {
	manager *CustomizedManager
}

func newInnerAppNodeEventHandler() *innerAppNodeEventHandler {
	return &innerAppNodeEventHandler{}
}

func (h *innerAppNodeEventHandler) RegisterManager(manager *CustomizedManager) {
	h.manager = manager
}

// AppNode add event
// when AppNode add or update, find each associated service and update corresponding AppService in cache
func (h *innerAppNodeEventHandler) OnAdd(obj interface{}) {
	node, ok := obj.(*v1.AppNode)
	if !ok {
		blog.Warnf("AppNode event handler get unknown type obj %v OnAdd", obj)
		return
	}
	blog.Infof("AppNode %s/%s add, ready to refresh", node.GetNamespace(), node.GetName())
	tmpSelector := labels.NewSelector()
	svcs, err := h.manager.getAppSvcInformer().Lister().List(tmpSelector)
	if err != nil {
		customCritical.WithLabelValues(typeAppNode, eventAdd).Inc()
		blog.Infof("svc informer list appsvcs failed, err %s", err.Error())
		return
	}
	for _, svc := range svcs {
		if isMatch(svc, node) {
			h.manager.updateAppService(svc)
		}
	}
	return
}

// AppNode update event
func (h *innerAppNodeEventHandler) OnUpdate(oldObj, newObj interface{}) {
	oldNode, okOld := oldObj.(*v1.AppNode)
	if !okOld {
		blog.Warnf("AppNode event handler get unknown type oldObj %v OnUpdate", oldObj)
		return
	}
	newNode, okNew := newObj.(*v1.AppNode)
	if !okNew {
		blog.Warnf("AppNode event handler get unknown type newObj %v OnUpdate", newObj)
		return
	}
	//TODO: GetCreationTimestamp() will cause process crash??? Jim leave this
	newNode.SetCreationTimestamp(oldNode.GetCreationTimestamp())
	if reflect.DeepEqual(oldObj, newObj) {
		customEvent.WithLabelValues(typeAppNode, eventUpdate, statusNotFinish).Inc()
		blog.Warnf("AppNode %s/%s No changed, skip update event", newNode.GetNamespace(), newNode.GetName())
		return
	}
	blog.Infof("AppNode %s/%s update, ready to refresh", newNode.GetNamespace(), newNode.GetName())
	tmpSelector := labels.NewSelector()
	svcs, err := h.manager.getAppSvcInformer().Lister().List(tmpSelector)
	if err != nil {
		customCritical.WithLabelValues(typeAppNode, eventUpdate).Inc()
		blog.Infof("svc informer list appsvcs failed, err %s", err.Error())
		return
	}
	for _, svc := range svcs {
		if isMatch(svc, newNode) {
			h.manager.updateAppService(svc)
			break
		}
	}
}

// AppNode delete event
func (h *innerAppNodeEventHandler) OnDelete(obj interface{}) {
	node, ok := obj.(*v1.AppNode)
	if !ok {
		blog.Warnf("AppNode event handler get unknown type oldObj %v OnDelete", obj)
		return
	}
	blog.Infof("AppNode %s/%s delete, ready to refresh", node.GetNamespace(), node.GetName())
	tmpSelector := labels.NewSelector()
	svcs, err := h.manager.getAppSvcInformer().Lister().List(tmpSelector)
	if err != nil {
		customCritical.WithLabelValues(typeAppNode, eventDelete).Inc()
		blog.Infof("svc informer list appsvcs failed, err %s", err.Error())
		return
	}
	for _, svc := range svcs {
		if isMatch(svc, node) {
			h.manager.updateAppService(svc)
		}
	}
	return
}

// NewClient create custom service client
func NewClient(config string, handler cache.ResourceEventHandler, syncPeriod time.Duration) (svcclient.Client, error) {
	var restConfig *rest.Config
	var err error
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

	cliset, err := internalclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("create client set failed, with rest config %v, err %s", restConfig, err.Error())
		return nil, err
	}

	blog.Infof("start create informer factory")
	factory := informers.NewSharedInformerFactory(cliset, syncPeriod)
	// informer and lister for AppSvc
	appSvcInformer := factory.Mesh().V1().AppSvcs()
	appSvcLister := appSvcInformer.Lister()
	// informer and lister for AppNode
	appNodeInformer := factory.Mesh().V1().AppNodes()
	appNodeLister := appNodeInformer.Lister()
	blog.Infof("create AppService cache")
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)

	cus := &CustomizedManager{
		factory:           factory,
		svcInformer:       appSvcInformer,
		svcLister:         appSvcLister,
		nodeInformer:      appNodeInformer,
		nodeLister:        appNodeLister,
		appServiceCache:   store,
		appServiceHandler: handler,
		stopCh:            make(chan struct{}),
	}
	blog.Infof("start running informer")
	go appSvcInformer.Informer().Run(cus.stopCh)
	go appNodeInformer.Informer().Run(cus.stopCh)
	blog.Infof("wait for cache sync successfully")

	svcHandler := newInnerAppSvcEventHandler()
	svcHandler.RegisterManager(cus)
	appHandler := newInnerAppNodeEventHandler()
	appHandler.RegisterManager(cus)
	blog.Infof("add app svc handler to informer")
	appSvcInformer.Informer().AddEventHandler(svcHandler)
	blog.Infof("add app node handler to informer")
	appNodeInformer.Informer().AddEventHandler(appHandler)

	return cus, nil
}

// CustomizedManager client for custom service discovery
type CustomizedManager struct {
	factory           informers.SharedInformerFactory
	svcInformer       informermeshv1.AppSvcInformer
	svcLister         listerv1.AppSvcLister
	nodeInformer      informermeshv1.AppNodeInformer
	nodeLister        listerv1.AppNodeLister
	appServiceCache   cache.Store
	appServiceHandler cache.ResourceEventHandler
	stopCh            chan struct{}
}

// GetAppService get AppService from local cache
func (cm *CustomizedManager) GetAppService(ns, name string) (*svcclient.AppService, error) {
	svc, isExisted, err := cm.appServiceCache.GetByKey(fmt.Sprintf("%s/%s", ns, name))
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventGet).Inc()
		blog.Errorf("get AppService by key %s/%s from cache failed, err %s", ns, name, err.Error())
		return nil, fmt.Errorf("get AppService by key %s/%s from cache failed, err %s", ns, name, err.Error())
	}
	if !isExisted {
		customEvent.WithLabelValues(typeAppService, eventGet, statusNotFinish).Inc()
		blog.Warnf("get no AppService by key %s/%s", ns, name)
		return nil, fmt.Errorf("get no AppService by key %s/%s", ns, name)
	}
	appService, ok := svc.(*svcclient.AppService)
	if !ok {
		customCritical.WithLabelValues(typeAppService, eventGet).Inc()
		blog.Errorf("get obj %v from cache is not type AppService", svc)
		return nil, fmt.Errorf("get obj %v from cache is not type AppService", svc)
	}
	customEvent.WithLabelValues(typeAppService, eventGet, statusSuccess).Inc()
	return appService, nil
}

// ListAppService list AppService
func (cm *CustomizedManager) ListAppService(l map[string]string) ([]*svcclient.AppService, error) {
	selector := labels.Set(l).AsSelector()
	//getting all specified datas from local cache with ListOptions
	svcs, err := cm.svcLister.List(selector)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventList).Inc()
		blog.Errorf("BcsManager list all mesh.V1.AppSvc local cache failed, %s", err.Error())
		return nil, err
	}
	if len(svcs) == 0 {
		customEvent.WithLabelValues(typeAppService, eventList, statusNotFinish).Inc()
		blog.Warnf("BcsManager list no mesh.V1.AppSvc in local cache with specified selector %s", selector.String())
		return nil, nil
	}
	var internalAppSvcs []*svcclient.AppService
	for _, svc := range svcs {
		if len(svc.Spec.Selector) == 0 {
			blog.Warnf("BcsManager get empty Selector for mesh.V1.AppSvc %s/%s", svc.GetNamespace(), svc.GetName())
			continue
		}
		nodes, err := cm.getAppNodesBySvc(svc)
		if err != nil {
			blog.Errorf("BcsManager get mesh.V1.AppNode by AppSvc %s/%s failed, %s. skip", svc.GetNamespace(), svc.GetName(), err.Error())
			continue
		}
		localSvc, err := cm.convert(svc, nodes)
		if err != nil {
			blog.Errorf("BcsManager convert mesh.V1.AppSvc %s/%s to local AppService failed, %s", svc.GetNamespace(), svc.GetName(), err.Error())
			continue
		}
		internalAppSvcs = append(internalAppSvcs, localSvc)
	}
	customEvent.WithLabelValues(typeAppService, eventList, statusSuccess).Inc()
	return internalAppSvcs, nil
}

// Close close client
func (cm *CustomizedManager) Close() {
	close(cm.stopCh)
}

// ListAppServiceFromStatefulSet list app service from stateful set
// not implemented
func (cm *CustomizedManager) ListAppServiceFromStatefulSet(ns, name string) ([]*svcclient.AppService, error) {
	blog.Warnf("ListAppServiceFromStatefulSet is not implemented for CustomizedManager")
	return nil, nil
}

// getAppNodesBySvc get all AppNodes by specified v1.AppSvc
func (cm *CustomizedManager) getAppNodesBySvc(svc *v1.AppSvc) ([]*v1.AppNode, error) {
	nodeSelector := labels.Set(svc.Spec.Selector)
	return cm.nodeLister.AppNodes(svc.GetNamespace()).List(nodeSelector.AsSelector())
}

// convert internal function for Discovery data conversion
func (cm *CustomizedManager) convert(svc *v1.AppSvc, nodes []*v1.AppNode) (*svcclient.AppService, error) {
	localSvc := &svcclient.AppService{
		TypeMeta:   svc.TypeMeta,
		ObjectMeta: svc.ObjectMeta,
		Type:       svc.Spec.Type,
		Frontend:   svc.Spec.Frontend,
		Alias:      svc.Spec.Alias,
		WANIP:      svc.Spec.WANIP,
	}
	localSvc.ServicePorts = servicePortConvert(svc.Spec.ServicePorts)
	if len(nodes) == 0 {
		blog.Warnf("AppSvc %s/%s select no AppNodes.", svc.GetNamespace(), svc.GetName())
	}
	for _, node := range nodes {
		if len(node.Spec.Ports) == 0 {
			// some container in a pod has no port, log it in a high level
			blog.V(3).Infof("mesh.V1.AppNode %s/%s get no port definition, #lost expected data#.", node.GetNamespace(), node.GetName())
			continue
		}
		localNode := svcclient.AppNode{
			TypeMeta:   node.TypeMeta,
			ObjectMeta: node.ObjectMeta,
			Index:      node.Spec.Index,
			Version:    node.Spec.Version,
			Weight:     node.Spec.Weight,
			NodeIP:     node.Spec.NodeIP,
			ProxyIP:    node.Spec.ProxyIP,
		}
		localNode.Ports = nodePortConvert(node.Spec.Ports)
		localSvc.Nodes = append(localSvc.Nodes, localNode)
	}
	blog.V(5).Infof("get AppService %v", localSvc)
	return localSvc, nil
}

// isMatch
func isMatch(svc *v1.AppSvc, node *v1.AppNode) bool {
	if svc == nil || node == nil {
		return false
	}
	nodeSelector := labels.Set(svc.Spec.Selector).AsSelector()
	if nodeSelector.Matches(labels.Set(node.Labels)) {
		return true
	}
	return false
}

// convert ServicePort of AppSvc to ServicePort of AppService
func servicePortConvert(ports []v1.ServicePort) (ret []svcclient.ServicePort) {
	for _, p := range ports {
		ret = append(ret, svcclient.ServicePort{
			Name:        p.Name,
			Protocol:    p.Protocol,
			Domain:      p.Domain,
			Path:        p.Path,
			ServicePort: p.ServicePort,
			ProxyPort:   p.ProxyPort,
		})
	}
	return ret
}

// convert NodePort of AppNode to NodePort of AppService
func nodePortConvert(ports []v1.NodePort) (ret []svcclient.NodePort) {
	for _, p := range ports {
		ret = append(ret, svcclient.NodePort{
			Name:      p.Name,
			Protocol:  p.Protocol,
			NodePort:  p.NodePort,
			ProxyPort: p.ProxyPort,
		})
	}
	return ret
}

//TODO: 考虑建立node到svc的反向索引
func (cm *CustomizedManager) updateAppService(svc *v1.AppSvc) {
	// get all AppNode object by AppSvc
	nodes, err := cm.getAppNodesBySvc(svc)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		blog.Warnf("get AppNode by AppSvc %s/%s failed, err %s", svc.GetName(), svc.GetNamespace(), err.Error())
		return
	}
	// get new AppService
	newAppService, err := cm.convert(svc, nodes)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		blog.Warnf("convert %v with its nodes %v to AppService failed, err %s", svc, nodes, err.Error())
		return
	}
	// get old AppService
	oldAppService, isExisted, err := cm.appServiceCache.Get(newAppService)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		blog.Warnf("get old AppService by newAppService %v failed, err %s", newAppService, err.Error())
		return
	}
	if !isExisted {
		err = cm.appServiceCache.Add(newAppService)
		if err != nil {
			customCritical.WithLabelValues(typeAppService, eventAdd).Inc()
			blog.Warnf("add AppService %s/%s in AppService cache failed, err %s", newAppService.GetName(), newAppService.GetNamespace(), err.Error())
			return
		}
		cm.appServiceHandler.OnAdd(newAppService)
		blog.V(5).Infof("add %v to AppService cache successfully", newAppService)
	} else {
		err = cm.appServiceCache.Update(newAppService)
		if err != nil {
			customCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
			blog.Warnf("update AppService %s/%s in AppService cache failed, err %s", newAppService.GetName(), newAppService.GetNamespace(), err.Error())
			return
		}
		cm.appServiceHandler.OnUpdate(oldAppService, newAppService)
		blog.V(5).Infof("update %v to AppService cache successfully", newAppService)
	}
	customEvent.WithLabelValues(typeAppService, eventUpdate, statusSuccess).Inc()
}

// deleteAppService delete AppService from cache by AppService name and Namespace
func (cm *CustomizedManager) deleteAppService(svc *v1.AppSvc) {
	key := fmt.Sprintf("%s/%s", svc.GetNamespace(), svc.GetName())
	oldAppService, isExisted, err := cm.appServiceCache.GetByKey(key)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		blog.Warnf("get old AppService by key %s failed, err %s", key, err.Error())
		return
	}
	if !isExisted {
		customEvent.WithLabelValues(typeAppService, eventDelete, statusNotFinish).Inc()
		blog.Warnf("no old AppService %s/%s in cache, no need to delete", svc.GetName(), svc.GetNamespace())
		return
	}
	err = cm.appServiceCache.Delete(oldAppService)
	if err != nil {
		customCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		blog.Warnf("delete AppService %s/%s in AppService cache failed, err %s")
		return
	}
	cm.appServiceHandler.OnDelete(oldAppService)
	blog.V(5).Infof("delete %v from AppService cache successfully", oldAppService)
	customEvent.WithLabelValues(typeAppService, eventDelete, statusSuccess).Inc()
}

func (cm *CustomizedManager) getAppNodeInformer() informermeshv1.AppNodeInformer {
	return cm.nodeInformer
}

func (cm *CustomizedManager) getAppSvcInformer() informermeshv1.AppSvcInformer {
	return cm.svcInformer
}
