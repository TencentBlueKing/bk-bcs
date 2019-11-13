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

package mesos

import (
	"bk-bcs/bcs-common/common/blog"
	bcstypes "bk-bcs/bcs-common/common/types"
	schetypes "bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	v2 "bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"
	mesosinformers "bk-bcs/bcs-mesos/pkg/client/informers"
	informerv2 "bk-bcs/bcs-mesos/pkg/client/informers/bkbcs/v2"
	mesosclientset "bk-bcs/bcs-mesos/pkg/client/internalclientset"
	listerv2 "bk-bcs/bcs-mesos/pkg/client/lister/bkbcs/v2"
	svcclient "bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	// event for prometheus
	eventAdd    = "add"
	eventUpdate = "update"
	eventDelete = "delete"
	eventGet    = "get"
	eventList   = "list"
	// type for mesos service client
	typeBcsService = "bcsservice"
	typeTaskGroup  = "taskgroup"
	typeAppService = "appservice"
	//state for event
	statusSuccess   = "success"
	statusFailure   = "failure"
	statusNotFinish = "notfinish"

	mesosEvent = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_mesos_events",
		Help: "Events for mesos service client",
	}, []string{"type", "event", "status"})
	mesosCritical = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "clb_serviceclient_mesos_critical_err",
		Help: "logic error for mesos service client",
	}, []string{"type", "event"})
)

func init() {
	prometheus.MustRegister(mesosEvent)
	prometheus.MustRegister(mesosCritical)
}

//NewClient create mesos etcd storage client for AppService
func NewClient(config string, handler cache.ResourceEventHandler, syncPeriod time.Duration) (svcclient.Client, error) {
	var restConfig *rest.Config
	var err error
	if len(config) == 0 {
		blog.Infof("MesosManager use inCluster configuration: %s", config)
		restConfig, err = rest.InClusterConfig()
		if err != nil {
			blog.Errorf("MesosManager get incluster config failed, err %s", err.Error())
			return nil, err
		}
	} else {
		//parse configuration
		restConfig, err = clientcmd.BuildConfigFromFlags("", config)
		if err != nil {
			blog.Errorf("MesosManager create internal client with kubeconfig %s failed, err %s", config, err.Error())
			return nil, err
		}
	}
	cliset, err := mesosclientset.NewForConfig(restConfig)
	if err != nil {
		blog.Errorf("MesosManager create clientset failed, with rest config %v, err %s", restConfig, err.Error())
		return nil, err
	}
	blog.Infof("MesosManager start create informer factory....")
	factory := mesosinformers.NewSharedInformerFactory(cliset, syncPeriod)
	svcInformer := factory.Bkbcs().V2().BcsServices()
	svcLister := svcInformer.Lister()
	taskGroupInformer := factory.Bkbcs().V2().TaskGroups()
	taskGroupLister := taskGroupInformer.Lister()
	blog.Infof("MesosManager create AppService cache....")
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	manager := &Manager{
		factory:           factory,
		svcInformer:       svcInformer,
		svcLister:         svcLister,
		taskGroupInformer: taskGroupInformer,
		taskGroupLister:   taskGroupLister,
		appSvcCache:       store,
		appSvcHandler:     handler,
		stopCh:            make(chan struct{}),
	}
	blog.Infof("MesosManager start running informer....")
	go svcInformer.Informer().Run(manager.stopCh)
	go taskGroupInformer.Informer().Run(manager.stopCh)
	results := factory.WaitForCacheSync(manager.stopCh)
	for key, value := range results {
		blog.Infof("MesosManager Wait For Cache %s Sync, result: %s", key, value)
	}
	blog.Infof("MesosManager wait for cache sync successfully...")
	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.OnBcsServiceAdd,
		UpdateFunc: manager.OnBcsServiceUpdate,
		DeleteFunc: manager.OnBcsServiceDelete,
	})
	blog.Infof("MesosManager add TaskGroup handler to informer")
	taskGroupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.OnTaskGroupAdd,
		UpdateFunc: manager.OnTaskGroupUpdate,
		DeleteFunc: manager.OnTaskGroupDelete,
	})
	return manager, nil
}

// Manager implement svcclient for mesos container meta data convertion
// all mesos data structures reference to bk-bcs/bcs-mesos/pkg/apis
type Manager struct {
	factory           mesosinformers.SharedInformerFactory
	svcInformer       informerv2.BcsServiceInformer
	svcLister         listerv2.BcsServiceLister
	taskGroupInformer informerv2.TaskGroupInformer
	taskGroupLister   listerv2.TaskGroupLister
	appSvcCache       cache.Store
	appSvcHandler     cache.ResourceEventHandler
	stopCh            chan struct{}
}

// GetAppService get service by specified namespace & name
func (mm *Manager) GetAppService(ns, name string) (*svcclient.AppService, error) {
	if len(ns) == 0 || len(name) == 0 {
		blog.Errorf("MesosManager lost namespace or name when GetAppService")
		mesosEvent.WithLabelValues(typeAppService, eventGet, statusFailure).Inc()
		return nil, fmt.Errorf("lost name or namespace")
	}
	key := fmt.Sprintf("%s/%s", ns, name)
	data, exist, err := mm.appSvcCache.GetByKey(key)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventGet).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventGet, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager get local cache %s failed, %s", key, err.Error())
		return nil, err
	}
	if !exist {
		blog.Warnf("MesosManager get no %s AppService in local cache", key)
		return nil, fmt.Errorf("get no AppService by key %s", key)
	}
	svc, ok := data.(*svcclient.AppService)
	if !ok {
		mesosCritical.WithLabelValues(typeAppService, eventGet).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventGet, statusFailure).Inc()
		blog.Errorf("[Critical] Mesos got unexpcted Data Object from localCache. Key: %s. Pay more attention", key)
		return nil, fmt.Errorf("")
	}
	mesosEvent.WithLabelValues(typeAppService, eventGet, statusSuccess).Inc()
	return svc, nil
}

// ListAppService list all service in cache, filter by Label
// selector comes from Set.AsSelector() see: k8s.io/apimachinery/pkg/labels.Set
func (mm *Manager) ListAppService(label map[string]string) ([]*svcclient.AppService, error) {
	selector := labels.Set(label).AsSelector()
	var svcList []*svcclient.AppService
	err := cache.ListAll(mm.appSvcCache, selector, func(m interface{}) {
		svcList = append(svcList, m.(*svcclient.AppService))
	})
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventList).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventList, statusFailure).Inc()
		blog.Errorf("[Ciritcal] MesosManager List all AppService in local cache failed, %s", err.Error())
		return nil, err
	}
	mesosEvent.WithLabelValues(typeAppService, eventList, statusSuccess).Inc()
	return svcList, nil

	//logic below is converting AppService from BcsService every time.
	//below is reserved for backup

	// //getting all specified datas from local cache with ListOptions
	// svcs, err := mm.svcLister.List(selector)
	// if err != nil {
	// 	blog.Errorf("MesosManager list BcsService by Selector %s in local cache failed, %s", selector.String(), err.Error())
	// 	return nil, err
	// }
	// if len(svcs) == 0 {
	// 	blog.Warnf("MesosManager list No BcsService in local cache with specified selector %s", selector.String())
	// 	return nil, nil
	// }
	// var internalAppSvcs []*svcclient.AppService
	// for _, svc := range svcs {
	// 	if len(svc.Spec.Spec.Selector) == 0 {
	// 		blog.Warnf("MesosManager get empty Selector for BcsService %s/%s", svc.GetNamespace(), svc.GetName())
	// 		continue
	// 	}
	// 	taskGroups, err := mm.getTaskGroupByService(svc)
	// 	if err != nil {
	// 		blog.Errorf("[Critical]MesosManager get TaskGroup by BcsService %s/%s failed, %s. skip", svc.GetNamespace(), svc.GetName(), err.Error())
	// 		continue
	// 	}
	// 	//when converting, we need to filter taskgroups that don't Running/Lost
	// 	//but there is siuation: user define service, but no Running containers.
	// 	//AppService will lack of pods information.
	// 	internalSvc, err := mm.mesosConvertToAppService(svc, taskGroups)
	// 	if err != nil {
	// 		blog.Errorf("[Critical]MesosManager convert BcsService %s/%s to local AppService failed, %s", svc.GetNamespace(), svc.GetName(), err.Error())
	// 		continue
	// 	}
	// 	internalAppSvcs = append(internalAppSvcs, internalSvc)
	// }
	// return internalAppSvcs, nil
}

// ListAppServiceFromStatefulSet list app services, for each stateful node, generate a AppService object
func (mm *Manager) ListAppServiceFromStatefulSet(ns, name string) ([]*svcclient.AppService, error) {
	return nil, fmt.Errorf("Not implemented")
}

// Close client, clean resource
func (mm *Manager) Close() {
	close(mm.stopCh)
}

// convert internal function for Discovery data conversion
// todo: not finished
func (mm *Manager) mesosConvertToAppService(svc *v2.BcsService, taskGroups []*v2.TaskGroup) (*svcclient.AppService, error) {
	if len(svc.Spec.Spec.Ports) == 0 {
		return nil, fmt.Errorf("BcsService lost ports info")
	}
	internalSvc := &svcclient.AppService{
		TypeMeta:   svc.TypeMeta,
		ObjectMeta: svc.ObjectMeta,
		Type:       svc.Spec.Spec.Type,
		Frontend:   svc.Spec.Spec.ClusterIP,
	}
	internalSvc.ServicePorts = servicePortConvert(svc.Spec.Spec.Ports)
	if len(taskGroups) == 0 {
		blog.Warnf("BcsService %s/%s select no TaskGroups.", svc.GetNamespace(), svc.GetName())
	}
	for _, taskgroup := range taskGroups {
		if !mm.isTaskGroupValid(taskgroup) {
			blog.Warnf("MesosManager check TaskGroup %s/%s not ready when converting AppService.", taskgroup.GetNamespace(), taskgroup.GetName())
			continue
		}
		internalNode := mm.taskgroupConvertAppNode(taskgroup)
		internalSvc.Nodes = append(internalSvc.Nodes, *internalNode)
	}
	blog.V(5).Infof("MesosManager convert AppService successfully: %v", internalSvc)
	return internalSvc, nil
}

func servicePortConvert(ports []bcstypes.ServicePort) []svcclient.ServicePort {
	var svcPorts []svcclient.ServicePort
	for _, p := range ports {
		port := svcclient.ServicePort{
			Name:        p.Name,
			Protocol:    strings.ToLower(p.Protocol),
			Domain:      p.DomainName,
			Path:        p.Path,
			ServicePort: p.Port,
			ProxyPort:   p.NodePort,
		}
		svcPorts = append(svcPorts, port)
	}
	return svcPorts
}

func convertPortMapping(taskports []*schetypes.PortMapping) []svcclient.NodePort {
	var nodeports []svcclient.NodePort
	for _, port := range taskports {
		nport := svcclient.NodePort{
			Name:      port.Name,
			Protocol:  strings.ToLower(port.Protocol),
			NodePort:  int(port.ContainerPort),
			ProxyPort: int(port.HostPort),
		}
		nodeports = append(nodeports, nport)
	}
	return nodeports
}

//statusData holder for task raw json data
type statusData struct {
	IPAddress   string `json:"IPAddress"`
	NodeAddress string `json:"NodeAddress"`
}

//taskgroupConvertAppNode convert taskgroup to AppNode for container service discovery
// only convert
func (mm *Manager) taskgroupConvertAppNode(taskgroup *v2.TaskGroup) *svcclient.AppNode {
	node := &svcclient.AppNode{
		TypeMeta: taskgroup.TypeMeta,
		ObjectMeta: metav1.ObjectMeta{
			Name:        taskgroup.Spec.Name,
			Namespace:   taskgroup.GetNamespace(),
			Labels:      taskgroup.GetLabels(),
			Annotations: taskgroup.GetAnnotations(),
		},
		Index: taskgroup.Spec.Name,
	}
	for index, task := range taskgroup.Spec.Taskgroup {
		if task.PortMappings != nil {
			ports := convertPortMapping(task.PortMappings)
			node.Ports = append(node.Ports, ports...)
		}
		if node.Network != "" {
			node.Network = task.Network
		}
		if len(node.NodeIP) != 0 {
			//here means NodeIP & ProxyIP already setting
			//only for iterating all PortMappings
			continue
		}
		if len(task.StatusData) == 0 {
			blog.Warnf("MesosManaget check TaskGroup %s/%s index %d task lost StatusData. try next one", taskgroup.GetNamespace(), taskgroup.GetName(), index)
			continue
		}
		info := new(statusData)
		if err := json.Unmarshal([]byte(task.StatusData), info); err != nil {
			mesosCritical.WithLabelValues(typeTaskGroup, eventGet).Inc()
			blog.Errorf("[Critical] MesosManager %s/%s decode Task Container %s Status data failed, %s", taskgroup.GetNamespace(), taskgroup.GetName(), task.ID, err)
			continue
		}
		if len(info.IPAddress) != 0 {
			node.NodeIP = info.IPAddress
			node.ProxyIP = info.NodeAddress
			// for iterating all PortMappings
			continue
		}
		if len(info.IPAddress) == 0 && len(info.NodeAddress) != 0 {
			//use for simple bridge mode
			node.NodeIP = info.NodeAddress
			node.ProxyIP = info.NodeAddress
		}
	}
	return node

}

//updateAppService update internal AppService by BcsService, including Add
func (mm *Manager) updateAppService(svc *v2.BcsService) {
	taskGroups, err := mm.getTaskGroupByService(svc)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager get taskgroup by BcsService %s/%s failed, err %s", svc.GetName(), svc.GetNamespace(), err.Error())
		return
	}
	newAppService, err := mm.mesosConvertToAppService(svc, taskGroups)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager convert %v with its taskgroup %v to AppService failed, err %s", svc, taskGroups, err.Error())
		return
	}
	//broken here, continue later
	oldAppService, isExisted, err := mm.appSvcCache.Get(newAppService)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]get old AppService by newAppService %s/%s failed, err %s", newAppService.GetNamespace(), newAppService.GetName(), err.Error())
		return
	}
	if !isExisted {
		if err := mm.appSvcCache.Add(newAppService); err != nil {
			mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
			mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
			blog.Errorf("[Critical]MesosManager add AppService %s/%s in AppService cache failed, err %s", newAppService.GetNamespace(), newAppService.GetName(), err.Error())
			return
		}
		mm.appSvcHandler.OnAdd(newAppService)
		blog.Infof("MesosManager add %v to AppService cache successfully", newAppService)
	} else {
		if err := mm.appSvcCache.Update(newAppService); err != nil {
			mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
			mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
			blog.Errorf("[Critical]MesosManager update AppService %s/%s in AppService cache failed, err %s", newAppService.GetNamespace(), newAppService.GetName(), err.Error())
			return
		}
		mm.appSvcHandler.OnUpdate(oldAppService, newAppService)
		blog.Infof("MesosManager update %v to AppService cache successfully", newAppService)
	}
	mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusSuccess).Inc()
}

//deleteAppService delete internal AppService by BcsService
func (mm *Manager) deleteAppService(svc *v2.BcsService) {
	key := fmt.Sprintf("%s/%s", svc.GetNamespace(), svc.GetName())
	oldAppService, isExisted, err := mm.appSvcCache.GetByKey(key)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager get old AppService by key %s failed, err %s", key, err.Error())
		return
	}
	if !isExisted {
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusNotFinish).Inc()
		blog.Warnf("MesosManager has no old AppService %s/%s in cache, no need to delete", svc.GetNamespace(), svc.GetName())
		return
	}
	if err := mm.appSvcCache.Delete(oldAppService); err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager delete AppService %s/%s in AppService cache failed, %s", svc.GetNamespace(), svc.GetName(), err.Error())
		return
	}
	mm.appSvcHandler.OnDelete(oldAppService)
	mesosEvent.WithLabelValues(typeAppService, eventDelete, statusSuccess).Inc()
	blog.Infof("MesosManager delete %s from AppService cache successfully", key)
}

//getTaskGroupByService get taskgroup BcsService Selector
func (mm *Manager) getTaskGroupByService(svc *v2.BcsService) ([]*v2.TaskGroup, error) {
	selector := labels.Set(svc.Spec.Spec.Selector).AsSelector()
	return mm.taskGroupLister.TaskGroups(svc.GetNamespace()).List(selector)
}

//Cache EventHandler implementation for BcsService

//OnBcsServiceAdd add event implementation
func (mm *Manager) OnBcsServiceAdd(obj interface{}) {
	svc, ok := obj.(*v2.BcsService)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsService, eventAdd).Inc()
		blog.Errorf("[Critical]BcsService event handler get unknown type obj %v OnAdd", obj)
		return
	}
	blog.Infof("BcsService %s/%s add, ready to refresh", svc.GetNamespace(), svc.GetName())
	mesosEvent.WithLabelValues(typeBcsService, eventAdd, statusSuccess).Inc()
	mm.updateAppService(svc)
}

//OnBcsServiceUpdate update event implementation
func (mm *Manager) OnBcsServiceUpdate(oldObj, newObj interface{}) {
	oldSvc, okOld := oldObj.(*v2.BcsService)
	if !okOld {
		mesosCritical.WithLabelValues(typeBcsService, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager handler get unknown type oldObj %v OnBcsServiceUpdate", oldObj)
		return
	}
	newSvc, okNew := newObj.(*v2.BcsService)
	if !okNew {
		mesosCritical.WithLabelValues(typeBcsService, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager BcsService event handler get unknown type newObj %v OnBcsServiceUpdate", newObj)
		return
	}
	if reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) {
		blog.Warnf("MesosManager Found BcsService %s/%s No changed, skip update event", newSvc.GetNamespace(), newSvc.GetName())
		mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusNotFinish).Inc()
		return
	}
	//todo(DeveloperJim): for metrics
	//customEvent.With(prometheus.Labels{"type": typeAppSvc, "event": eventUpdate})
	blog.Infof("BcsService %s/%s update, ready to refresh", newSvc.GetNamespace(), newSvc.GetName())
	mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusSuccess).Inc()
	mm.updateAppService(newSvc)
}

//OnBcsServiceDelete delete event implementation
func (mm *Manager) OnBcsServiceDelete(obj interface{}) {
	svc, ok := obj.(*v2.BcsService)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsService, eventDelete).Inc()
		blog.Errorf("[Criticail]MesosManager BcsService event handler get unknown type obj %v OnDelete", obj)
		return
	}
	//todo(DeveloperJim): for metrics
	//customEvent.With(prometheus.Labels{"type": typeAppSvc, "event": eventDelete})
	blog.Infof("AppSvc %s/%s delete, ready to refresh", svc.GetNamespace(), svc.GetName())
	mesosEvent.WithLabelValues(typeBcsService, eventDelete, statusSuccess).Inc()
	mm.deleteAppService(svc)
}

//OnTaskGroupAdd add event implementation
func (mm *Manager) OnTaskGroupAdd(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		mesosCritical.WithLabelValues(typeTaskGroup, eventAdd).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type obj %v OnTaskGroupAdd", obj)
		return
	}
	//todo(DeveloperJim): for metrics
	//customEvent.With(prometheus.Labels{"type": typeAppSvc, "event": eventAdd})
	blog.Infof("TaskGroup %s/%s add, ready to refresh", taskgroup.GetNamespace(), taskgroup.GetName())
	if !mm.isTaskGroupValid(taskgroup) {
		mesosEvent.WithLabelValues(typeTaskGroup, eventAdd, statusNotFinish).Inc()
		return
	}
	//maybe multiple services select same taskgroup
	//so we need to search all service under same namespace
	svcLister := mm.svcLister.BcsServices(taskgroup.GetNamespace())
	svcs, err := svcLister.List(labels.Everything())
	if err != nil {
		mesosCritical.WithLabelValues(typeTaskGroup, eventAdd).Inc()
		blog.Errorf("[Critical]MesosManager Get BcsService by taskgroup Namespace failed, %s", taskgroup.GetNamespace(), err.Error())
		return
	}
	if len(svcs) == 0 {
		mesosEvent.WithLabelValues(typeTaskGroup, eventAdd, statusNotFinish).Inc()
		blog.Warnf("TaskGroup %s/%s get no relative BcsService in Cache.", taskgroup.GetNamespace(), taskgroup.GetName())
		return
	}
	for index, svc := range svcs {
		if isSelected(svc, taskgroup) {
			blog.Infof(
				"TaskGroup %s/%s add, relative BcsService [%d] %s/%s updated",
				taskgroup.GetNamespace(), taskgroup.GetName(), index,
				svc.GetNamespace(), svc.GetName(),
			)
			mm.updateAppService(svc)
		}
	}
	mesosEvent.WithLabelValues(typeTaskGroup, eventAdd, statusSuccess).Inc()
}

//OnTaskGroupUpdate upadte event implementation
func (mm *Manager) OnTaskGroupUpdate(oldObj, newObj interface{}) {
	oldTaskgroup, okOld := oldObj.(*v2.TaskGroup)
	if !okOld {
		mesosCritical.WithLabelValues(typeTaskGroup, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type oldObj %v OnTaskGroupUpdate", oldObj)
		return
	}
	newTaskgroup, okNew := newObj.(*v2.TaskGroup)
	if !okNew {
		mesosCritical.WithLabelValues(typeTaskGroup, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type newObj %v OnTaskGroupUpdate", newObj)
		return
	}
	if reflect.DeepEqual(oldTaskgroup.Spec, newTaskgroup.Spec) {
		mesosEvent.WithLabelValues(typeTaskGroup, eventUpdate, statusNotFinish).Inc()
		blog.Warnf("MesosManager TaskGroup %s/%s No changed, skip update event", newTaskgroup.GetNamespace(), newTaskgroup.GetName())
		return
	}
	//todo(DeveloperJim): for metrics
	//customEvent.With(prometheus.Labels{"type": typeAppSvc, "event": eventUpdate})
	blog.Infof("TaskGroup %s/%s update, ready to refresh", newTaskgroup.GetNamespace(), newTaskgroup.GetName())
	if !mm.isTaskGroupValid(newTaskgroup) {
		mesosEvent.WithLabelValues(typeTaskGroup, eventUpdate, statusNotFinish).Inc()
		return
	}
	svcLister := mm.svcLister.BcsServices(newTaskgroup.GetNamespace())
	svcs, err := svcLister.List(labels.Everything())
	if err != nil {
		mesosCritical.WithLabelValues(typeTaskGroup, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager Get BcsService by taskgroup Namespace %s failed, %s", newTaskgroup.GetNamespace(), err.Error())
		return
	}
	if len(svcs) == 0 {
		mesosEvent.WithLabelValues(typeTaskGroup, eventUpdate, statusNotFinish).Inc()
		blog.Warnf("TaskGroup %s/%s get no relative BcsService in Cache when updating", newTaskgroup.GetNamespace(), newTaskgroup.GetName())
		return
	}
	for index, svc := range svcs {
		if isSelected(svc, newTaskgroup) {
			blog.Infof(
				"TaskGroup %s/%s updated, relative BcsService [%d] %s/%s need to updated",
				newTaskgroup.GetNamespace(), newTaskgroup.GetName(), index,
				svc.GetNamespace(), svc.GetName(),
			)
			mm.updateAppService(svc)
		}
	}
	mesosEvent.WithLabelValues(typeTaskGroup, eventUpdate, statusSuccess).Inc()
}

//OnTaskGroupDelete delete event implementation
func (mm *Manager) OnTaskGroupDelete(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		mesosCritical.WithLabelValues(typeTaskGroup, eventDelete).Inc()
		blog.Errorf("[Critical]MesosManager TaskGroup event handler get unknown type obj %v OnDelete", obj)
		return
	}
	//todo(DeveloperJim): for metrics
	//customEvent.With(prometheus.Labels{"type": typeAppSvc, "event": eventDelete})
	blog.Infof("TaskGroup %s/%s delete, ready to refresh", taskgroup.GetNamespace(), taskgroup.GetName())
	svcLister := mm.svcLister.BcsServices(taskgroup.GetNamespace())
	svcs, err := svcLister.List(labels.Everything())
	if err != nil {
		mesosCritical.WithLabelValues(typeTaskGroup, eventDelete).Inc()
		blog.Errorf("[Critical]MesosManager Get BcsService by taskgroup Namespace %s failed, %s", taskgroup.GetNamespace(), err.Error())
		return
	}
	if len(svcs) == 0 {
		mesosEvent.WithLabelValues(typeTaskGroup, eventDelete, statusNotFinish).Inc()
		blog.Warnf("TaskGroup %s/%s get no relative BcsService in Cache when updating", taskgroup.GetNamespace(), taskgroup.GetName())
		return
	}
	for index, svc := range svcs {
		if isSelected(svc, taskgroup) {
			blog.Infof(
				"TaskGroup %s/%s updated, relative BcsService [%d] %s/%s need to updated",
				taskgroup.GetNamespace(), taskgroup.GetName(), index,
				svc.GetNamespace(), svc.GetName(),
			)
			mm.updateAppService(svc)
		}
	}
	mesosEvent.WithLabelValues(typeTaskGroup, eventDelete, statusSuccess).Inc()
}

//isTaskGroupValid check TaskGroup has container IP address
func (mm *Manager) isTaskGroupValid(taskgroup *v2.TaskGroup) bool {
	if !(taskgroup.Spec.Status == schetypes.TASK_STATUS_RUNNING || taskgroup.Spec.Status == schetypes.TASK_STATUS_LOST) {
		blog.Errorf(
			"MesosManager TaskGroup %s/%s Status[%s] lost IPAddress info. Not Ready",
			taskgroup.GetNamespace(), taskgroup.GetName(), taskgroup.Spec.Status,
		)
		return false
	}
	return true
}

// isSelected check if BcsService select specified TaskGroup
func isSelected(svc *v2.BcsService, taskgroup *v2.TaskGroup) bool {
	if svc == nil || taskgroup == nil {
		return false
	}
	taskgroupSelector := labels.Set(svc.Spec.Spec.Selector).AsSelector()
	if taskgroupSelector.Matches(labels.Set(taskgroup.Labels)) {
		return true
	}
	return false
}
