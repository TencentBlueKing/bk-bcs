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

package controller

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	internalclientset "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/clientset/versioned"
	informers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions"
	listers "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/listers/bkbcs/v2"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-dns/plugin/bcsscheduler/metrics"

	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/labels"
	clientGoCache "k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	watchTypeService  = "service"
	watchTypeEndpoint = "endpoint"
)

type EtcdController struct {
	conCxt         context.Context              //context for exit signal
	conCancel      context.CancelFunc           //stop all goroutine
	resyncperiod   int                          //resync all data period
	resynced       bool                         //status resynced
	kubeconfig     string                       //kubeconfig for kube-apiserver
	bkbcsClientSet *internalclientset.Clientset //kube bkbcs clientset
	serviceLister  listers.BcsServiceLister
	endpointLister listers.BcsEndpointLister
	watchType      string                                   //resource type to watch
	storage        cache.Store                              //cache storage
	nsStorage      map[string]context.CancelFunc            //storage for all namespace watcher
	nsLock         sync.Mutex                               //lock for nsStorage
	funcs          *clientGoCache.ResourceEventHandlerFuncs //funcs for callback
}

//NewEtcdController create controller according Store, Decoder end EventFuncs
func NewEtcdController(kconfig, wType string, period int, cache cache.Store, eventFunc *clientGoCache.ResourceEventHandlerFuncs) (*EtcdController, error) {
	if kconfig == "" {
		return nil, fmt.Errorf("create Controller failed, no kubeconfig provided")
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", kconfig)
	if err != nil {
		log.Printf("etcdcontroller build kubeconfig %s error %s", kconfig, err.Error())
		return nil, err
	}
	log.Printf("etcdcontroller build kubeconfig %s success", kconfig)

	clientset, err := internalclientset.NewForConfig(cfg)
	if err != nil {
		log.Printf("etcdcontroller build clientset error %s", err.Error())
		return nil, err
	}

	cxt, cancel := context.WithCancel(context.Background())
	controller := &EtcdController{
		conCxt:         cxt,
		conCancel:      cancel,
		resyncperiod:   period,
		kubeconfig:     kconfig,
		bkbcsClientSet: clientset,
		watchType:      wType,
		storage:        cache,
		nsStorage:      make(map[string]context.CancelFunc),
		funcs:          eventFunc,
	}

	return controller, nil
}

//RunController running controller, create goroutine watch kube-api data
func (ctrl *EtcdController) RunController(stopCh <-chan struct{}) error {

	if ctrl.watchType == watchTypeService {
		factory := informers.NewSharedInformerFactory(ctrl.bkbcsClientSet, time.Second*60)
		serviceInformer := factory.Bkbcs().V2().BcsServices()
		serviceLister := serviceInformer.Lister()
		ctrl.serviceLister = serviceLister
		//defer runtime.HandleCrash()

		serviceInformer.Informer().AddEventHandler(clientGoCache.ResourceEventHandlerFuncs{
			AddFunc:    ctrl.AddEvent,
			UpdateFunc: ctrl.UpdateEvent,
			DeleteFunc: ctrl.DeleteEvent,
		})

		go factory.Start(stopCh)

		log.Printf("Waiting for informer caches to sync")
		if ok := clientGoCache.WaitForCacheSync(stopCh, serviceInformer.Informer().HasSynced); !ok {
			return fmt.Errorf("failed to wait for services caches to sync")
		}

	} else if ctrl.watchType == watchTypeEndpoint {
		factory := informers.NewSharedInformerFactory(ctrl.bkbcsClientSet, time.Second*60)
		endpointInformer := factory.Bkbcs().V2().BcsEndpoints()
		endpointLister := endpointInformer.Lister()
		ctrl.endpointLister = endpointLister
		//defer runtime.HandleCrash()

		endpointInformer.Informer().AddEventHandler(clientGoCache.ResourceEventHandlerFuncs{
			AddFunc:    ctrl.AddEvent,
			UpdateFunc: ctrl.UpdateEvent,
			DeleteFunc: ctrl.DeleteEvent,
		})

		go factory.Start(stopCh)

		log.Printf("Waiting for informer caches to sync")
		if ok := clientGoCache.WaitForCacheSync(stopCh, endpointInformer.Informer().HasSynced); !ok {
			return fmt.Errorf("failed to wait for endpoints caches to sync")
		}
	}

	return nil

	////create resync event for all data
	//tick := time.NewTicker(time.Second * time.Duration(ctrl.resyncperiod))
	//for {
	//	select {
	//	case <-ctrl.conCxt.Done():
	//		log.Printf("[WARN] controller ask exist")
	//		return nil
	//	case now := <-tick.C:
	//		//resync all data from watchpath
	//		if ctrl.Resynced() {
	//			log.Printf("[WARN] controller %s under resync now [%s], drop this tick.", ctrl.watchType, now.String())
	//			continue
	//		}
	//		ctrl.resynced = true
	//		log.Printf("[INFO] resync %s tick, now: %s", ctrl.watchType, now.String())
	//		ctrl.dataResync()
	//		log.Printf("[INFO] resync %s tick, end: %s", ctrl.watchType, time.Now().String())
	//		ctrl.resynced = false
	//	}
	//}
}

func convertToService(obj interface{}) (interface{}, bool, string) {
	bcsService, ok := obj.(*v2.BcsService)
	if !ok {
		log.Printf("[ERROR] convert object to bcsservice failed")
		return nil, false, ""
	}
	srv := &bcsService.Spec.BcsService
	srv.Name = strings.ToLower(srv.Name)
	srv.NameSpace = strings.ToLower(srv.NameSpace)

	key := srv.NameSpace + "/" + srv.Name
	return srv, true, key
}

func convertToEndpoint(obj interface{}) (interface{}, bool, string) {
	bcsEndpoint, ok := obj.(*v2.BcsEndpoint)
	if !ok {
		log.Printf("[ERROR] convert object to bcsendpoint failed")
		return nil, false, ""
	}
	ep := &bcsEndpoint.Spec.BcsEndpoint
	ep.Name = strings.ToLower(ep.Name)
	ep.NameSpace = strings.ToLower(ep.NameSpace)

	key := ep.NameSpace + "/" + ep.Name
	return ep, true, key
}

func (ctrl *EtcdController) AddEvent(obj interface{}) {
	if ctrl.watchType == watchTypeService {
		srv, ok, key := convertToService(obj)
		if !ok {
			return
		}
		ctrl.updateStorageData(srv, key)
	} else if ctrl.watchType == watchTypeEndpoint {
		ep, ok, key := convertToEndpoint(obj)
		if !ok {
			return
		}
		ctrl.updateStorageData(ep, key)
	}
}

func (ctrl *EtcdController) UpdateEvent(oldObj, newObj interface{}) {
	/*if reflect.DeepEqual(oldObj, newObj) {
		dMeta := newObj.(metav1.Object)
		namespace := dMeta.GetNamespace()
		name := dMeta.GetName()
		log.Printf("got same %s: %s/%s, continue...", ctrl.watchType, namespace, name)
		return
	}*/

	if ctrl.watchType == watchTypeService {
		newSrv, ok, key := convertToService(newObj)
		if !ok {
			return
		}
		ctrl.updateStorageData(newSrv, key)
	} else if ctrl.watchType == watchTypeEndpoint {
		newEp, ok, key := convertToEndpoint(newObj)
		if !ok {
			return
		}
		ctrl.updateStorageData(newEp, key)
	}
}

func (ctrl *EtcdController) DeleteEvent(obj interface{}) {
	if ctrl.watchType == watchTypeService {
		_, ok, key := convertToService(obj)
		if !ok {
			return
		}
		ctrl.deleteStorageData(key)
	} else if ctrl.watchType == watchTypeEndpoint {
		_, ok, key := convertToEndpoint(obj)
		if !ok {
			return
		}
		ctrl.deleteStorageData(key)
	}
}

func (ctrl *EtcdController) updateStorageData(cur interface{}, key string) {
	old, exist, _ := ctrl.storage.Get(cur)
	if exist {
		ctrl.storage.Update(cur)
		ctrl.funcs.UpdateFunc(old, cur)
		log.Printf("[INFO] %s %s update content", ctrl.watchType, key)
		metrics.ZkNotifyTotal.WithLabelValues(metrics.UpdateOperation).Inc()
	} else {
		ctrl.storage.Add(cur)
		ctrl.funcs.AddFunc(cur)
		log.Printf("[INFO] %s %s add new data object", ctrl.watchType, key)
		metrics.DnsTotal.Inc()
		metrics.ZkNotifyTotal.WithLabelValues(metrics.AddOperation).Inc()
	}
}

func (ctrl *EtcdController) deleteStorageData(key string) {
	old, exist, _ := ctrl.storage.GetByKey(key)
	if exist {
		ctrl.storage.Delete(old)
		ctrl.funcs.DeleteFunc(old)
		log.Printf("[WARN] controller delete %s %s in cache", ctrl.watchType, key)
		metrics.DnsTotal.Dec()
		metrics.ZkNotifyTotal.WithLabelValues(metrics.DeleteOperation).Inc()
	} else {
		log.Printf("[ERROR] controller lost %s %s in cache, somewhere go wrong", ctrl.watchType, key)
	}
}

//StopController stop controller event, clean all data
func (ctrl *EtcdController) StopController() {
	log.Printf("[INFO] controller %s stop", ctrl.watchType)
	ctrl.conCancel()
}

//Resynced check controller is under resynced
func (ctrl *EtcdController) Resynced() bool {
	return ctrl.resynced
}

//dataResync resync all data from kube-apiserver
func (ctrl *EtcdController) dataResync() {
	if ctrl.watchType == watchTypeService {
		ctrl.serviceResync()
	} else if ctrl.watchType == watchTypeEndpoint {
		ctrl.endpointResync()
	}
}

//serviceResync resync all service data from kube-apiserver
func (ctrl *EtcdController) serviceResync() {
	//bcsServiceList, err := ctrl.bkbcsClientSet.BkbcsV2().BcsServices("").List(metav1.ListOptions{})
	bcsServiceList, err := ctrl.serviceLister.List(labels.Everything())
	if err != nil {
		log.Printf("etcdcontroller list bcs services failed, %s", err.Error())
	}

	//when we iterator all data in kube-api, we checking:
	//1. update cache with kube-api data by force
	//2. cache data is dirty or not(exist in cache but lost in kube-api)

	//step 1
	existsIndex := make(map[string]bool)
	for _, bcsService := range bcsServiceList {
		srv := &bcsService.Spec.BcsService
		srv.Name = strings.ToLower(srv.Name)
		srv.NameSpace = strings.ToLower(srv.NameSpace)

		key := srv.NameSpace + "/" + srv.Name
		old, exist, _ := ctrl.storage.Get(srv)
		if exist {
			log.Printf("update service %s/%s", srv.NameSpace, srv.Name)
			ctrl.storage.Update(srv)
			ctrl.funcs.UpdateFunc(old, srv)
			existsIndex[key] = true
		} else {
			//todo(developer): lost data in cache, we add it to cache,
			//and we still need add watch goroutine for
			//updating data from zookeeper
			log.Printf("[WARN] RESYNC found service ###%s### lost in cache", key)
			ctrl.storage.Add(srv)
			ctrl.funcs.AddFunc(srv)
			metrics.DnsTotal.Inc()
		}
	}

	//step 2
	cacheIndexs := ctrl.storage.ListKeys()
	for _, indexKey := range cacheIndexs {
		if _, ok := existsIndex[indexKey]; ok {
			continue
		}
		ns, name := strings.Split(indexKey, "/")[0], strings.Split(indexKey, "/")[1]
		//bcsService, _ := ctrl.bkbcsClientSet.BkbcsV2().BcsServices(ns).Get(name, metav1.GetOptions{})
		bcsService, _ := ctrl.serviceLister.BcsServices(ns).Get(name)
		if bcsService != nil {
			log.Printf("[WARN] %s all in cache& kube-api, but lost in tmp map, maybe added latest or trigger RESYNC warn", indexKey)
			continue
		}
		log.Printf("[WARN] %s lost in kube-api, TODO: delete in cache.", indexKey)
		//todo(developer): logs statistic for repairing this warnning
	}
}

//endpointResync resync all endpoint data from kube-apiserver
func (ctrl *EtcdController) endpointResync() {
	//bcsEndpointList, err := ctrl.bkbcsClientSet.BkbcsV2().BcsEndpoints("").List(metav1.ListOptions{})
	bcsEndpointList, err := ctrl.endpointLister.List(labels.Everything())
	if err != nil {
		log.Printf("etcdcontroller list bcs endpoints failed, %s", err.Error())
	}

	//when we iterator all data in kube-api, we checking:
	//1. update cache with kube-api data by force
	//2. cache data is dirty or not(exist in cache but lost in kube-api)

	//step 1
	existsIndex := make(map[string]bool)
	for _, bcsEndpoint := range bcsEndpointList {
		ep := &bcsEndpoint.Spec.BcsEndpoint
		ep.Name = strings.ToLower(ep.Name)
		ep.NameSpace = strings.ToLower(ep.NameSpace)

		key := ep.NameSpace + "/" + ep.Name
		old, exist, _ := ctrl.storage.Get(ep)
		if exist {
			log.Printf("update endpoint %s/%s", ep.NameSpace, ep.Name)
			ctrl.storage.Update(ep)
			ctrl.funcs.UpdateFunc(old, ep)
			existsIndex[key] = true
		} else {
			//todo(developer): lost data in cache, we add it to cache,
			//and we still need add watch goroutine for
			//updating data from zookeeper
			log.Printf("[WARN] RESYNC found endpoint ###%s### lost in cache", key)
			ctrl.storage.Add(ep)
			ctrl.funcs.AddFunc(ep)
			metrics.DnsTotal.Inc()
		}
	}

	//step 2
	cacheIndexs := ctrl.storage.ListKeys()
	for _, indexKey := range cacheIndexs {
		if _, ok := existsIndex[indexKey]; ok {
			continue
		}
		ns, name := strings.Split(indexKey, "/")[0], strings.Split(indexKey, "/")[1]
		//bcsEndpoint, _ := ctrl.bkbcsClientSet.BkbcsV2().BcsEndpoints(ns).Get(name, metav1.GetOptions{})
		bcsEndpoint, _ := ctrl.endpointLister.BcsEndpoints(ns).Get(name)
		if bcsEndpoint != nil {
			log.Printf("[WARN] %s all in cache& kube-api, but lost in tmp map, maybe added latest or trigger RESYNC warn", indexKey)
			continue
		}
		log.Printf("[WARN] %s lost in kube-api, TODO: delete in cache.", indexKey)
		//todo(developer): logs statistic for repairing this warnning
	}
}
