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

package k8s

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"

	// client-go v4.0.0
	//"k8s.io/client-go/pkg/api/v1"
	//"k8s.io/client-go/pkg/apis/extensions/v1beta1"

	// client-go v5.0.1
	appsv1beta1 "k8s.io/api/apps/v1beta1"
	//appsv1beta2 "k8s.io/api/apps/v1beta2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	//"k8s.io/api/apps/v1beta2"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
)

// =================== interface & struct ===================

// WatcherInterface : interface for all watchers
type WatcherInterface interface {
	Run(stop <-chan struct{})
	AddEvent(obj interface{})
	DeleteEvent(obj interface{})
	UpdateEvent(oldObject, newObject interface{})
	GetByKey(key string) (interface{}, bool, error)
	ListKeys() []string
}

// Watcher struct
type Watcher struct {
	controller              cache.Controller
	store                   cache.Store
	resourceType            string
	writer                  *output.Writer
	exportServiceController *ExportServiceController
	sharedWatchers          map[string]WatcherInterface
}

type OriginEvent struct {
	ResourceName string
	ResourceType string
	Namespace    string
	Action       string
}

const (
	ListWatchSyncPeriodSecond = 30
)

// =================== New & Run ===================

// NewWatcher https://github.com/kubernetes/client-go#how-to-use-it
// https://github.com/kubernetes/client-go/blob/master/examples/in-cluster-client-configuration/main.go
func NewWatcher(client *rest.Interface, resourceType string, resourceName string, objType runtime.Object, writer *output.Writer, sharedWatchers map[string]WatcherInterface, es *ExportServiceController) *Watcher {
	watcher := new(Watcher)
	// resource name in this
	listWatch := cache.NewListWatchFromClient(*client, resourceName, metav1.NamespaceAll, fields.Everything())
	store, controller := cache.NewInformer(
		listWatch,
		objType,
		time.Second*ListWatchSyncPeriodSecond,
		cache.ResourceEventHandlerFuncs{
			AddFunc:    watcher.AddEvent,
			UpdateFunc: watcher.UpdateEvent,
			DeleteFunc: watcher.DeleteEvent,
		},
	)

	watcher.resourceType = resourceType
	watcher.store = store
	watcher.controller = controller
	watcher.writer = writer
	watcher.exportServiceController = es

	watcher.sharedWatchers = sharedWatchers

	return watcher
}

func (watcher *Watcher) Run(stop <-chan struct{}) {
	watcher.controller.Run(stop)
}

// =================== Methods ===================

func (watcher *Watcher) GetByKey(key string) (interface{}, bool, error) {
	return watcher.store.GetByKey(key)

}

func (watcher *Watcher) ListKeys() []string {
	return watcher.store.ListKeys()

}

func (watcher *Watcher) AddEvent(obj interface{}) {
	syncData := watcher.genSyncData(obj, "Add")
	if syncData == nil {
		return
	}
	watcher.writer.Sync(syncData)
	// watcher.InformExportServiceChange(obj, "Add")

	// do alarm
	if syncData.OwnerUID != "" {
		watcher.writer.SyncAlarmEvent(syncData)
	}
}

func (watcher *Watcher) DeleteEvent(obj interface{}) {
	syncData := watcher.genSyncData(obj, "Delete")
	if syncData == nil {
		return
	}
	watcher.writer.Sync(syncData)
	// watcher.InformExportServiceChange(obj, "Delete")
}

func (watcher *Watcher) UpdateEvent(oldObj, newObj interface{}) {
	// check if no data change
	d, ok := watcher.convert(newObj)
	if !ok {
		glog.Errorf("Convert object to %s fail! object is %v", watcher.resourceType, newObj)
		return
	}

	//oldObject, ok := watcher.convert(oldObj)
	//if !ok {
	//	glog.Errorf("Convert old object to %s fail! object is %v", watcher.resourceType, newObj)
	//	return
	//}
	// TODO: test resource version equal
	//    newTgwIngress := new.(*tgwv1.TgwIngress)
	//    oldTgwIngress := old.(*tgwv1.TgwIngress)
	//    if newTgwIngress.ResourceVersion == oldTgwIngress.ResourceVersion {
	//        return
	//}
	if reflect.DeepEqual(oldObj, d) {
		dMeta := d.(metav1.Object)
		namespace := dMeta.GetNamespace()
		name := dMeta.GetName()
		// -v >=2 then log.info will show detail
		glog.V(2).Infof("Got same %s: %s/%s", watcher.resourceType, namespace, name)
		return
	}

	syncData := watcher.genSyncData(newObj, "Update")
	if syncData == nil {
		return
	}
	watcher.writer.Sync(syncData)
	//watcher.InformExportServiceChange(newObj, "Update")

	// do alarm
	if syncData.OwnerUID != "" {
		watcher.writer.SyncAlarmEvent(syncData)
	}
}

// =================== Helpers ===================

func (watcher *Watcher) convert(obj interface{}) (v interface{}, ok bool) {
	switch watcher.resourceType {
	case "Node":
		v, ok = obj.(*v1.Node)
	case "Pod":
		v, ok = obj.(*v1.Pod)
	case "ReplicationController":
		v, ok = obj.(*v1.ReplicationController)
	case "Service":
		v, ok = obj.(*v1.Service)
	case "EndPoints":
		v, ok = obj.(*v1.Endpoints)
	case "ConfigMap":
		v, ok = obj.(*v1.ConfigMap)
	case "Secret":
		v, ok = obj.(*v1.Secret)
	case "Namespace":
		v, ok = obj.(*v1.Namespace)
	case "Event":
		v, ok = obj.(*v1.Event)
	case "Deployment":
		v, ok = obj.(*v1beta1.Deployment)
		// NOTE: k8s v1.7 and v1.8, should use v1beta1
		//v, ok = obj.(*v1beta2.Deployment)
	case "Ingress":
		v, ok = obj.(*v1beta1.Ingress)
	case "ReplicaSet":
		v, ok = obj.(*v1beta1.ReplicaSet)
		// NOTE: k8s v1.7 and v1.8, should use v1beta1
		//v, ok = obj.(*v1beta2.ReplicaSet)
	case "DaemonSet":
		v, ok = obj.(*v1beta1.DaemonSet)
		//v, ok = obj.(*appsv1beta2.DaemonSet)
	case "StatefulSet":
		v, ok = obj.(*appsv1beta1.StatefulSet)
	case "Job":
		v, ok = obj.(*batchv1.Job)

	default:
		v = nil
		ok = false
	}
	return v, ok
}

// IsEventShouldFilter: filter kubernetes system event
func (watcher *Watcher) isEventShouldFilter(meta metav1.Object, eventAction string) bool {
	// NOTE: event not support delete
	// bugfix here: must in top of this func, in case of Name or Namespace return true
	if eventAction == "Delete" && watcher.resourceType == "Event" {
		// Event not support delete
		return true
	}

	if meta.GetNamespace() == "kube-system" && watcher.resourceType == "Event" {
		// kubeops start pod with those prefix
		name := meta.GetName()
		if strings.HasPrefix(name, "kube-") ||
			strings.HasPrefix(name, "kubedns-") ||
			strings.HasPrefix(name, "nginx-proxy") ||
			strings.HasPrefix(name, "bcs-") {
			return false
		}
		return true
	}
	if meta.GetNamespace() == "kube-system" {
		return true
	}
	if meta.GetName() == "kubernetes" {
		return true
	}

	return false
}

func (watcher *Watcher) genSyncData(obj interface{}, eventAction string) *action.SyncData {

	d, ok := watcher.convert(obj)
	if !ok {
		glog.Errorf("Convert object to %s fail! object is %v", watcher.resourceType, obj)
		return nil
	}

	// construct and send
	dMeta := d.(metav1.Object)
	namespace := dMeta.GetNamespace()
	name := dMeta.GetName()
	if watcher.isEventShouldFilter(dMeta, eventAction) {
		glog.V(2).Infof("Filtered %s %s: %s/%s", eventAction, watcher.resourceType, namespace, name)
		return nil
	}

	ownerUID := ""
	// NOTE: 生成时, 就确认了是否会告警, 即 ownerUID != "", 告警
	// if Event, get OwnerReference uid
	if watcher.resourceType == "Event" {
		event, ok := d.(*v1.Event)

		if !ok {
			glog.Infof("genSyncData parse to Event fail: %v", d)
		}

		// TODO: handle 1) not just pod 2) system component warning event

		// must be pod,
		// only handle the warning event. which may caused some errors.
		if ok && event.InvolvedObject.Kind == "Pod" && event.Type == "Warning" {
			iNamespace := event.InvolvedObject.Namespace
			iName := event.InvolvedObject.Name

			key := fmt.Sprintf("%s/%s", iNamespace, iName)

			// NOTE: 这里的store, 是watcher(Event)的store, 所以拿不到..........
			//relatedPodInterface, exists, err := watcher.store.GetByKey(key)

			relatedPodInterface, exists, err := watcher.sharedWatchers["Pod"].GetByKey(key)

			if exists && err == nil {
				relatedPod, ok := relatedPodInterface.(*v1.Pod)
				if ok {
					owners := relatedPod.GetOwnerReferences()
					if len(owners) >= 1 {
						ownerUID = string(owners[0].UID)
					}
				}
			} else {
				glog.Warnf("The owner of Event's related-object not exists! Event=%v, related-object=%s", event, key)
			}

		}
	}

	//if ownerUID != "" {
	//	glog.Infof("=========== ownerUID: %s", ownerUID)
	//}

	//var json =  jsoniter.ConfigCompatibleWithStandardLibrary
	glog.Infof("Ready: %s %s: %s/%s", eventAction, watcher.resourceType, namespace, name)
	syncData := &action.SyncData{
		Kind:      watcher.resourceType,
		Namespace: namespace,
		Name:      name,
		Action:    eventAction,
		Data:      d,
		OwnerUID:  ownerUID,
	}

	return syncData
}

func (watcher *Watcher) InformExportServiceChange(obj interface{}, originAction string) {
	// only following types will influence ingress
	// struct{} doesn't require any additional space
	var IngressRelatedResource = map[string]struct{}{
		"Ingress":   {},
		"Service":   {},
		"EndPoints": {},
		"ConfigMap": {},
		"Secret":    {},
	}
	if _, ok := IngressRelatedResource[watcher.resourceType]; !ok {
		return
	}

	d, ok := watcher.convert(obj)
	if !ok {
		glog.Errorf("Convert object to %s fail! object is %v", watcher.resourceType, obj)
		return
	}

	dMeta := d.(metav1.Object)
	name := dMeta.GetName()
	namespace := dMeta.GetNamespace()
	originEvent := OriginEvent{
		Action:       originAction,
		ResourceType: watcher.resourceType,
		ResourceName: name,
		Namespace:    namespace,
	}
	watcher.exportServiceController.SyncIngress(originEvent)
}
