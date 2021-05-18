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
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	containertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	mesosclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	mesosinformers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions"
	informerv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions/bkbcs/v2"
	listerv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/listers/bkbcs/v2"
	svcclient "github.com/Tencent/bk-bcs/bcs-services/bcs-clb-controller/pkg/serviceclient"

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
	typeBcsService   = "bcsservice"
	typeBcsEndpoint  = "bcsendpoint"
	typeAppService   = "appservice"
	typeBcsTaskgroup = "bcstaskgroup"
	// state for event
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

// NewClient create mesos etcd storage client for AppService
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
		// parse configuration
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
	bcsEndpointInformer := factory.Bkbcs().V2().BcsEndpoints()
	bcsEndpointLister := bcsEndpointInformer.Lister()
	taskgroupInformer := factory.Bkbcs().V2().TaskGroups()
	taskgroupLister := taskgroupInformer.Lister()
	blog.Infof("MesosManager create AppService cache....")
	store := cache.NewStore(cache.DeletionHandlingMetaNamespaceKeyFunc)
	manager := &Manager{
		factory:             factory,
		svcInformer:         svcInformer,
		svcLister:           svcLister,
		bcsEndpointInformer: bcsEndpointInformer,
		bcsEndpointLister:   bcsEndpointLister,
		taskgroupInformer:   taskgroupInformer,
		taskgroupLister:     taskgroupLister,
		appSvcCache:         store,
		appSvcHandler:       handler,
		stopCh:              make(chan struct{}),
	}
	blog.Infof("MesosManager start running informer....")
	factory.Start(manager.stopCh)
	results := factory.WaitForCacheSync(manager.stopCh)
	for key, value := range results {
		blog.Infof("MesosManager Wait For Cache %s Sync, result: %v", key, value)
	}
	blog.Infof("MesosManager wait for cache sync successfully...")
	svcInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.OnBcsServiceAdd,
		UpdateFunc: manager.OnBcsServiceUpdate,
		DeleteFunc: manager.OnBcsServiceDelete,
	})
	blog.Infof("MesosManager add BcsService handler to informer")
	bcsEndpointInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.OnBcsEndpointAdd,
		UpdateFunc: manager.OnBcsEndpointUpdate,
		DeleteFunc: manager.OnBcsEndpointDelete,
	})
	blog.Infof("MesosManager add BcsEndpoint handler to informer")
	taskgroupInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    manager.OnBcsTaskgroupAdd,
		UpdateFunc: manager.OnBcsTaskgroupUpdate,
		DeleteFunc: manager.OnBcsTaskgroupDelete,
	})
	blog.Infof("MesosManager created successfully")
	return manager, nil
}

// Manager implement svcclient for mesos container meta data convertion
// all mesos data structures reference to bk-bcs/bcs-mesos/pkg/apis
type Manager struct {
	factory             mesosinformers.SharedInformerFactory
	svcInformer         informerv2.BcsServiceInformer
	svcLister           listerv2.BcsServiceLister
	bcsEndpointInformer informerv2.BcsEndpointInformer
	bcsEndpointLister   listerv2.BcsEndpointLister
	taskgroupInformer   informerv2.TaskGroupInformer
	taskgroupLister     listerv2.TaskGroupLister
	appSvcCache         cache.Store
	appSvcHandler       cache.ResourceEventHandler
	stopCh              chan struct{}
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
}

func checkTaskgroupRunning(tg *v2.TaskGroup) bool {
	if tg == nil {
		return false
	}
	if tg.Spec.Status == "Running" {
		return true
	}
	return false
}

func decodeTaskgroupStatusData(data string) (*containertypes.BcsContainerInfo, error) {
	bcsInfo := new(containertypes.BcsContainerInfo)
	if err := json.Unmarshal([]byte(data), bcsInfo); err != nil {
		return nil, err
	}
	return bcsInfo, nil
}

func getIndexRcFromTaskgroupName(name string) (int, string, error) {
	nameStrs := strings.Split(name, ".")
	if len(nameStrs) != 5 {
		return 0, "", fmt.Errorf("invalid taskgroup name %s", name)
	}
	tgNumberStr := nameStrs[0]
	tgIndex, err := strconv.Atoi(tgNumberStr)
	if err != nil {
		return 0, "", fmt.Errorf("parse %s to int failed, err %s", tgNumberStr, err.Error())
	}
	return tgIndex, nameStrs[1], nil
}

func sortTaskgroups(tgs []*v2.TaskGroup) {
	sort.Slice(tgs, func(i, j int) bool {
		indexI, rcI, _ := getIndexRcFromTaskgroupName(tgs[i].GetName())
		indexJ, rcJ, _ := getIndexRcFromTaskgroupName(tgs[j].GetName())
		if indexI < indexJ {
			return true
		} else if indexI > indexJ {
			return false
		}
		return rcI < rcJ
	})
}

func convertTaskgroupsToAppServices(tgs []*v2.TaskGroup) ([]*svcclient.AppService, error) {
	var appServiceList []*svcclient.AppService
	if len(tgs) == 0 {
		blog.Errorf("no taskgroups to be converted")
		return nil, fmt.Errorf("no taskgroups to be converted")
	}
	for _, tg := range tgs {
		newAppSvc := &svcclient.AppService{
			TypeMeta: metav1.TypeMeta{
				Kind:       "AppService",
				APIVersion: tg.APIVersion,
			},
			ObjectMeta:   tg.ObjectMeta,
			Frontend:     make([]string, 0),
			ServicePorts: make([]svcclient.ServicePort, 0),
		}
		tgName := tg.GetName()
		tgIndex, _, err := getIndexRcFromTaskgroupName(tgName)
		if err != nil {
			return nil, err
		}
		// use taskgroup index to set service port and node port
		newServicePort := svcclient.ServicePort{
			Name:        tgName,
			ServicePort: tgIndex,
			TargetPort:  tgIndex,
		}
		if checkTaskgroupRunning(tg) && len(tg.Spec.Taskgroup) != 0 {
			statusInfo, err := decodeTaskgroupStatusData(tg.Spec.Taskgroup[0].StatusData)
			if err != nil {
				blog.Warnf("decode taskgroup status data %s failed, err %s",
					tg.Spec.Taskgroup[0].StatusData, err.Error())
			} else {
				taskgroupIP := statusInfo.IPAddress
				if strings.ToLower(statusInfo.NetworkMode) == "host" {
					taskgroupIP = statusInfo.NodeAddress
				}
				newAppNode := svcclient.AppNode{
					TypeMeta:   tg.TypeMeta,
					ObjectMeta: tg.ObjectMeta,
					Index:      tgName,
					Weight:     10,
					NodeIP:     taskgroupIP,
					ProxyIP:    statusInfo.NodeAddress,
					Ports:      make([]svcclient.NodePort, 0),
				}
				newPort := svcclient.NodePort{
					Name:     tgName + "-" + strconv.Itoa(tgIndex),
					Protocol: "",
					NodePort: tgIndex,
				}
				newAppNode.Ports = append(newAppNode.Ports, newPort)
				newAppSvc.Nodes = append(newAppSvc.Nodes, newAppNode)
			}
		}
		newAppSvc.ServicePorts = append(newAppSvc.ServicePorts, newServicePort)
		appServiceList = append(appServiceList, newAppSvc)
	}
	return appServiceList, nil
}

// ListAppServiceFromStatefulSet list app services, for each stateful node, generate a AppService object
func (mm *Manager) ListAppServiceFromStatefulSet(ns, svcname string) ([]*svcclient.AppService, error) {
	stsSvc, err := mm.svcLister.BcsServices(ns).Get(svcname)
	if err != nil {
		blog.Warnf("get bcsService %s/%s failed, err %s", svcname, ns, err.Error())
		return nil, nil
	}
	if len(stsSvc.Spec.Spec.Selector) == 0 {
		blog.Warnf("selector of bcsService %s/%s is empty, err %s", stsSvc.GetName(),
			stsSvc.GetNamespace(), err.Error())
		return nil, nil
	}
	selector := labels.Set(stsSvc.Spec.Spec.Selector).AsSelector()
	taskgroups, err := mm.taskgroupLister.TaskGroups(ns).List(selector)
	if err != nil {
		blog.Warnf("list taskgroups by selector %s failed, err %s", selector.String(), err.Error())
		return nil, nil
	}
	sortTaskgroups(taskgroups)
	appSvcList, err := convertTaskgroupsToAppServices(taskgroups)
	if err != nil {
		blog.Warnf("convert taskgroups to AppService failed, err %s", err.Error())
		return nil, nil
	}
	return appSvcList, nil
}

// Close client, clean resource
func (mm *Manager) Close() {
	close(mm.stopCh)
}

// convert internal function for Discovery data conversion
// todo: not finished
func (mm *Manager) mesosConvertToAppService(svc *v2.BcsService, bcsEndpoint *v2.BcsEndpoint) (
	*svcclient.AppService, error) {

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
	internalSvc.Nodes = mm.bcsendpointConvertAppNode(bcsEndpoint)
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

func convertContainerPortMapping(ports []bcstypes.ContainerPort) []svcclient.NodePort {
	var nodeports []svcclient.NodePort
	for _, port := range ports {
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

// bcsendpointConvertAppNode convert bcsendpoint to AppNode for
// container service discovery only convert
func (mm *Manager) bcsendpointConvertAppNode(bcsEndpoint *v2.BcsEndpoint) []svcclient.AppNode {
	var nodes []svcclient.AppNode
	for _, endpoint := range bcsEndpoint.Spec.Endpoints {
		node := svcclient.AppNode{
			ObjectMeta: metav1.ObjectMeta{
				Name:      endpoint.Target.Name,
				Namespace: bcsEndpoint.GetNamespace(),
			},
			Index:   endpoint.Target.Name,
			NodeIP:  endpoint.ContainerIP,
			ProxyIP: endpoint.NodeIP,
		}
		if len(endpoint.Ports) != 0 {
			node.Ports = convertContainerPortMapping(endpoint.Ports)
		} else {
			blog.Warnf("BcsEndpoints %s/%s Endpoint %s lost Port Information.",
				bcsEndpoint.GetNamespace(), bcsEndpoint.GetName(), endpoint.Target.Name)
		}
		nodes = append(nodes, node)
	}
	return nodes

}

// updateAppService update internal AppService by BcsService, including Add
func (mm *Manager) updateAppService(svc *v2.BcsService, bcsEndpoint *v2.BcsEndpoint) {
	newAppService, err := mm.mesosConvertToAppService(svc, bcsEndpoint)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
		// should not log error, because there are so many cases in which convert AppService failed
		blog.V(4).Infof("[Critical]MesosManager convert %v with its bcsendpoint %v to AppService failed, err %s",
			svc, bcsEndpoint, err.Error())
		return
	}
	// broken here, continue later
	oldAppService, isExisted, err := mm.appSvcCache.Get(newAppService)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]get old AppService by newAppService %s/%s failed, err %s",
			newAppService.GetNamespace(), newAppService.GetName(), err.Error())
		return
	}
	if !isExisted {
		if err := mm.appSvcCache.Add(newAppService); err != nil {
			mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
			mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
			blog.Errorf("[Critical]MesosManager add AppService %s/%s in AppService cache failed, err %s",
				newAppService.GetNamespace(), newAppService.GetName(), err.Error())
			return
		}
		mm.appSvcHandler.OnAdd(newAppService)
		blog.Infof("MesosManager add %v to AppService cache successfully", newAppService)
	} else {
		if err := mm.appSvcCache.Update(newAppService); err != nil {
			mesosCritical.WithLabelValues(typeAppService, eventUpdate).Inc()
			mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusFailure).Inc()
			blog.Errorf("[Critical]MesosManager update AppService %s/%s in AppService cache failed, err %s",
				newAppService.GetNamespace(), newAppService.GetName(), err.Error())
			return
		}
		mm.appSvcHandler.OnUpdate(oldAppService, newAppService)
		blog.V(5).Infof("MesosManager update %v to AppService cache successfully", newAppService)
	}
	mesosEvent.WithLabelValues(typeAppService, eventUpdate, statusSuccess).Inc()
}

// deleteAppService delete internal AppService by BcsService
func (mm *Manager) deleteAppService(key string) {
	oldAppService, isExisted, err := mm.appSvcCache.GetByKey(key)
	if err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager get old AppService by key %s failed, err %s", key, err.Error())
		return
	}
	if !isExisted {
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusNotFinish).Inc()
		blog.Warnf("MesosManager has no old AppService %s in cache, no need to delete", key)
		return
	}
	if err := mm.appSvcCache.Delete(oldAppService); err != nil {
		mesosCritical.WithLabelValues(typeAppService, eventDelete).Inc()
		mesosEvent.WithLabelValues(typeAppService, eventDelete, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager delete AppService %s in AppService cache failed, %s", key, err.Error())
		return
	}
	mm.appSvcHandler.OnDelete(oldAppService)
	mesosEvent.WithLabelValues(typeAppService, eventDelete, statusSuccess).Inc()
	blog.Infof("MesosManager delete %s from AppService cache successfully", key)
}

// Cache EventHandler implementation for BcsService

// OnBcsServiceAdd add event implementation
func (mm *Manager) OnBcsServiceAdd(obj interface{}) {
	svc, ok := obj.(*v2.BcsService)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsService, eventAdd).Inc()
		blog.Errorf("[Critical]BcsService event handler get unknown type obj %v OnAdd", obj)
		return
	}
	blog.V(5).Infof("BcsService %s/%s add, event +1", svc.GetNamespace(), svc.GetName())
	mesosEvent.WithLabelValues(typeBcsService, eventAdd, statusSuccess).Inc()
	// BcsEndpoint event will come with all IP address information later
	// we don't need to handle service add event
	// mm.updateAppService
}

// OnBcsServiceUpdate update event implementation
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
		blog.Errorf("[Critical]MesosManager BcsService event handler get unknown type newObj %v OnBcsServiceUpdate",
			newObj)
		return
	}
	if reflect.DeepEqual(oldSvc.Spec, newSvc.Spec) {
		blog.Warnf("MesosManager Found BcsService %s/%s No changed, skip update event",
			newSvc.GetNamespace(), newSvc.GetName())
		mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusNotFinish).Inc()
		return
	}
	blog.Infof("BcsService %s/%s update, ready to refresh", newSvc.GetNamespace(), newSvc.GetName())
	bcsEndpoint, err := mm.bcsEndpointLister.BcsEndpoints(newSvc.GetNamespace()).Get(newSvc.GetName())
	if err != nil {
		mesosCritical.WithLabelValues(typeBcsService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical] MesosManager get BcsEndpoint %s/%s failed when BcsService updating, %s",
			newSvc.GetNamespace(), newSvc.GetName(), err.Error())
		return
	}
	if bcsEndpoint == nil {
		mesosCritical.WithLabelValues(typeBcsService, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical] BcsService %s/%s get no relative BcsEndpoint when updating.",
			newSvc.GetNamespace(), newSvc.GetName())
		return
	}
	mesosEvent.WithLabelValues(typeBcsService, eventUpdate, statusSuccess).Inc()
	mm.updateAppService(newSvc, bcsEndpoint)
}

// OnBcsServiceDelete delete event implementation
func (mm *Manager) OnBcsServiceDelete(obj interface{}) {
	svc, ok := obj.(*v2.BcsService)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsService, eventDelete).Inc()
		blog.Errorf("[Criticail]MesosManager BcsService event handler get unknown type obj %v OnDelete", obj)
		return
	}
	key := fmt.Sprintf("%s/%s", svc.GetNamespace(), svc.GetName())
	blog.V(5).Infof("BcsService %s delete, ready to refresh", key)
	mesosEvent.WithLabelValues(typeBcsService, eventDelete, statusSuccess).Inc()
	mm.deleteAppService(key)
}

// OnBcsEndpointAdd add event implementation
func (mm *Manager) OnBcsEndpointAdd(obj interface{}) {
	bcsendpoint, ok := obj.(*v2.BcsEndpoint)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventAdd).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type obj %v OnBcsEndpointAdd", obj)
		return
	}
	blog.Infof("BcsEndpoint %s/%s add, ready to refresh", bcsendpoint.GetNamespace(), bcsendpoint.GetName())
	svcLister := mm.svcLister.BcsServices(bcsendpoint.GetNamespace())
	svc, err := svcLister.Get(bcsendpoint.GetName())
	if err != nil {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventAdd).Inc()
		mesosEvent.WithLabelValues(typeBcsEndpoint, eventAdd, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager Get BcsService by bcsendpoint Namespace/Name %s/%s failed, %s",
			bcsendpoint.GetNamespace(), bcsendpoint.GetName(), err.Error())
		return
	}
	mesosEvent.WithLabelValues(typeBcsEndpoint, eventAdd, statusSuccess).Inc()
	mm.updateAppService(svc, bcsendpoint)
}

// OnBcsEndpointUpdate upadte event implementation
func (mm *Manager) OnBcsEndpointUpdate(oldObj, newObj interface{}) {
	oldBcsendpoint, okOld := oldObj.(*v2.BcsEndpoint)
	if !okOld {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type oldObj %v OnBcsEndpointUpdate", oldObj)
		return
	}
	newBcsendpoint, okNew := newObj.(*v2.BcsEndpoint)
	if !okNew {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type newObj %v OnBcsEndpointUpdate", newObj)
		return
	}
	if reflect.DeepEqual(oldBcsendpoint.Spec.Endpoints, newBcsendpoint.Spec.Endpoints) {
		mesosEvent.WithLabelValues(typeBcsEndpoint, eventUpdate, statusNotFinish).Inc()
		blog.Warnf("MesosManager BcsEndpoint %s/%s No changed, skip update event",
			newBcsendpoint.GetNamespace(), newBcsendpoint.GetName())
		return
	}
	blog.Infof("BcsEndpoint %s/%s update, ready to refresh", newBcsendpoint.GetNamespace(), newBcsendpoint.GetName())
	svcLister := mm.svcLister.BcsServices(newBcsendpoint.GetNamespace())
	svc, err := svcLister.Get(newBcsendpoint.GetName())
	if err != nil {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeBcsEndpoint, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]MesosManager Get BcsService by bcsendpoint Namespace %s failed, %s",
			newBcsendpoint.GetNamespace(), err.Error())
		return
	}
	if svc == nil {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventUpdate).Inc()
		mesosEvent.WithLabelValues(typeBcsEndpoint, eventUpdate, statusFailure).Inc()
		blog.Errorf("[Critical]BcsEndpoint %s/%s get no relative BcsService in Cache when updating",
			newBcsendpoint.GetNamespace(), newBcsendpoint.GetName())
		return
	}
	mesosEvent.WithLabelValues(typeBcsEndpoint, eventUpdate, statusSuccess).Inc()
	mm.updateAppService(svc, newBcsendpoint)
}

// OnBcsEndpointDelete delete event implementation
func (mm *Manager) OnBcsEndpointDelete(obj interface{}) {
	bcsendpoint, ok := obj.(*v2.BcsEndpoint)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsEndpoint, eventDelete).Inc()
		blog.Errorf("[Critical]MesosManager BcsEndpoint event handler get unknown type obj %v OnDelete", obj)
		return
	}
	key := fmt.Sprintf("%s/%s", bcsendpoint.GetNamespace(), bcsendpoint.GetName())
	blog.V(5).Infof("BcsEndpoint %s delete, ready to refresh", key)
	mesosEvent.WithLabelValues(typeBcsEndpoint, eventDelete, statusSuccess).Inc()
	mm.deleteAppService(key)
}

// find service by taskgroup
func (mm *Manager) findServiceByTaskgroup(tg *v2.TaskGroup) (*v2.BcsService, error) {
	services, err := mm.svcLister.BcsServices(tg.GetNamespace()).List(labels.Everything())
	if err != nil {
		return nil, err
	}
	for _, svc := range services {
		if labels.Set(svc.Spec.Spec.Selector).AsSelector().Matches(labels.Set(tg.GetLabels())) {
			return svc, nil
		}
	}
	return nil, nil
}

// triger process update
func (mm *Manager) trigerAppServiceHandler(tg *v2.TaskGroup) {
	svc, err := mm.findServiceByTaskgroup(tg)
	if err != nil {
		blog.Errorf("find service by taskgroup %s/%s failed, err %s",
			tg.GetNamespace(), tg.GetName(), err.Error())
		return
	}
	if svc != nil {
		// just triger AppServiceHandler event, AppServiceHandler won't use this data
		mm.appSvcHandler.OnAdd(&svcclient.AppService{
			ObjectMeta: metav1.ObjectMeta{
				Name:      svc.GetName(),
				Namespace: svc.GetNamespace(),
			},
		})
	}
}

// OnBcsTaskgroupAdd taskgroup add event handler
func (mm *Manager) OnBcsTaskgroupAdd(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsTaskgroup, eventAdd).Inc()
		blog.Errorf("[Critical]BcsTaskgroup event handler get unknown type obj %v OnAdd", obj)
		return
	}
	blog.V(5).Infof("BcsTaskgroup %s/%s add, ready to refresh", taskgroup.GetNamespace(), taskgroup.GetName())
	mm.trigerAppServiceHandler(taskgroup)
}

// OnBcsTaskgroupUpdate taskgroup update event handler
func (mm *Manager) OnBcsTaskgroupUpdate(oldObj, newObj interface{}) {
	oldBcsTaskgroup, okOld := oldObj.(*v2.TaskGroup)
	if !okOld {
		mesosCritical.WithLabelValues(typeBcsTaskgroup, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type oldObj %v OnBcsTaskgroupUpdate", oldObj)
		return
	}
	newBcsTaskgroup, okNew := newObj.(*v2.TaskGroup)
	if !okNew {
		mesosCritical.WithLabelValues(typeBcsTaskgroup, eventUpdate).Inc()
		blog.Errorf("[Critical]MesosManager event handler get unknown type newObj %v OnBcsTaskgroupUpdate", newObj)
		return
	}
	// Taskgroup crd does not have status field, only care about spec
	if reflect.DeepEqual(oldBcsTaskgroup.Spec, newBcsTaskgroup.Spec) {
		mesosEvent.WithLabelValues(typeBcsTaskgroup, eventUpdate, statusNotFinish).Inc()
		blog.Warnf("MesosManager BcsTaskgroup %s/%s No changed, skip update event",
			newBcsTaskgroup.GetNamespace(), newBcsTaskgroup.GetName())
		return
	}
	blog.V(5).Infof("BcsTaskgroup %s/%s update, ready to refresh", newBcsTaskgroup.GetNamespace(), newBcsTaskgroup.GetName())
	mm.trigerAppServiceHandler(newBcsTaskgroup)
}

// OnBcsTaskgroupDelete taskgroup delete event handler
func (mm *Manager) OnBcsTaskgroupDelete(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		mesosCritical.WithLabelValues(typeBcsTaskgroup, eventAdd).Inc()
		blog.Errorf("[Critical]BcsTaskgroup event handler get unknown type obj %v OnDelete", obj)
		return
	}
	blog.V(5).Infof("BcsTaskgroup %s/%s delete, ready to refresh", taskgroup.GetNamespace(), taskgroup.GetName())
	mm.trigerAppServiceHandler(taskgroup)
}
