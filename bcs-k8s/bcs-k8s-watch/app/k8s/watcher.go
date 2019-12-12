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

	appsv1beta1 "k8s.io/api/apps/v1beta1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
)

const (
	// defaultSyncInterval is default sync interval.
	defaultSyncInterval = 30 * time.Second
)

// Watcher watch target type resource metadata from k8s cluster,
// and write to storage by synchronizer with series actions.
type Watcher struct {
	controller     cache.Controller
	store          cache.Store
	resourceType   string
	writer         *output.Writer
	sharedWatchers map[string]WatcherInterface
}

// NewWatcher creates a new watcher of target type resource.
func NewWatcher(client *rest.Interface, resourceType string, resourceName string, objType runtime.Object,
	writer *output.Writer, sharedWatchers map[string]WatcherInterface) *Watcher {

	watcher := &Watcher{
		resourceType:   resourceType,
		writer:         writer,
		sharedWatchers: sharedWatchers,
	}

	// build list watch.
	listWatch := cache.NewListWatchFromClient(*client, resourceName, metav1.NamespaceAll, fields.Everything())

	// register event handler.
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.AddEvent,
		UpdateFunc: watcher.UpdateEvent,
		DeleteFunc: watcher.DeleteEvent,
	}

	// build informer.
	store, controller := cache.NewInformer(listWatch, objType, defaultSyncInterval, eventHandler)
	watcher.store = store
	watcher.controller = controller

	return watcher
}

func (w *Watcher) GetByKey(key string) (interface{}, bool, error) {
	return w.store.GetByKey(key)
}

func (w *Watcher) ListKeys() []string {
	return w.store.ListKeys()
}

// Run starts the watcher.
func (w *Watcher) Run(stop <-chan struct{}) {
	// run controller.
	w.controller.Run(stop)
}

// AddEvent is event handler for add resource event.
func (w *Watcher) AddEvent(obj interface{}) {
	data := w.genSyncData(obj, action.SyncDataActionAdd)
	if data == nil {
		return
	}
	w.writer.Sync(data)

	// send alarm when there is an add event with non-empty owner uid.
	if data.OwnerUID != "" {
		w.writer.SyncAlarmEvent(data)
	}
}

// DeleteEvent is event handler for delete resource event.
func (w *Watcher) DeleteEvent(obj interface{}) {
	data := w.genSyncData(obj, action.SyncDataActionDelete)
	if data == nil {
		return
	}
	w.writer.Sync(data)
}

// UpdateEvent is event handler for update resource event.
func (w *Watcher) UpdateEvent(oldObj, newObj interface{}) {
	newObjReal, ok := w.convert(newObj)
	if !ok {
		glog.Errorf("Convert object to %s fail! object is %v", w.resourceType, newObj)
		return
	}

	// compare the object changes for update.
	if reflect.DeepEqual(oldObj, newObjReal) {
		newObjMetadata := newObjReal.(metav1.Object)

		// there is no changes, no need to update.
		glog.V(2).Infof("watcher got the same ResourceType[%s]: %s/%s",
			w.resourceType, newObjMetadata.GetNamespace(), newObjMetadata.GetName())
		return
	}

	// skip unnecessary node update event to reduce writer-queues pressure.
	if w.resourceType == "Node" {
		oldNode := oldObj.(*v1.Node)
		newNode := newObj.(*v1.Node)

		// TODO: a best way is to use deepcopy function, save the common fields,
		// update the change fields.

		// NOTE: why 5 ?
		var tempLastTimes = make([]metav1.Time, 5)
		tempVersion := newNode.ResourceVersion
		newNode.ResourceVersion = oldNode.ResourceVersion

		for i := range newNode.Status.Conditions {
			tempLastTimes[i] = newNode.Status.Conditions[i].LastHeartbeatTime
			newNode.Status.Conditions[i].LastHeartbeatTime = oldNode.Status.Conditions[i].LastHeartbeatTime
		}

		// the first DeepEqual skips in obj level, the second DeepEqual skips
		// the node data after save common fields.
		if reflect.DeepEqual(oldNode, newNode) {
			glog.V(2).Infof("skip unnecessary node update event")
			return
		}

		// recover new node metadata after DeepEqual finally.
		newNode.ResourceVersion = tempVersion
		for i := range newNode.Status.Conditions {
			newNode.Status.Conditions[i].LastHeartbeatTime = tempLastTimes[i]
		}
	}

	// it's need to update finally, sync metadata now.
	data := w.genSyncData(newObj, action.SyncDataActionUpdate)
	if data == nil {
		return
	}
	w.writer.Sync(data)

	// send alarm when there is an add event with non-empty owner uid.
	if data.OwnerUID != "" {
		w.writer.SyncAlarmEvent(data)
	}
}

func (w *Watcher) convert(obj interface{}) (v interface{}, ok bool) {
	switch w.resourceType {
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

// isEventShouldFilter filters k8s system events.
func (w *Watcher) isEventShouldFilter(meta metav1.Object, eventAction string) bool {
	// NOTE: event not support delete
	// bugfix here: must in top of this func, in case of Name or Namespace return true.
	if eventAction == action.SyncDataActionDelete && w.resourceType == "Event" {
		// Event not support delete.
		return true
	}

	if meta.GetNamespace() == "kube-system" && w.resourceType == "Event" {
		// kubeops start pod with those prefix.
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

func (w *Watcher) genSyncData(obj interface{}, eventAction string) *action.SyncData {
	d, ok := w.convert(obj)
	if !ok {
		glog.Errorf("watcher convert object to %s failed, object[%v]", w.resourceType, obj)
		return nil
	}

	// construct and send
	dMeta := d.(metav1.Object)
	namespace := dMeta.GetNamespace()
	name := dMeta.GetName()

	if w.isEventShouldFilter(dMeta, eventAction) {
		glog.V(2).Infof("watcher metadata is filtered %s %s: %s/%s", eventAction, w.resourceType, namespace, name)
		return nil
	}

	ownerUID := ""
	// NOTE: 生成时, 就确认了是否会告警, 即 ownerUID != "", 告警
	// if Event, get OwnerReference uid
	if w.resourceType == "Event" {
		event, ok := d.(*v1.Event)
		if !ok {
			glog.Infof("watcher parse object to Event failed, %v", d)
		}

		// TODO: handle 1) not just pod 2) system component warning event

		// must be pod,
		// only handle the warning event. which may caused some errors.
		if ok && event.InvolvedObject.Kind == "Pod" && event.Type == "Warning" {
			iNamespace := event.InvolvedObject.Namespace
			iName := event.InvolvedObject.Name

			key := fmt.Sprintf("%s/%s", iNamespace, iName)

			// NOTE: 这里的store, 是watcher(Event)的store, 所以拿不到..........
			relatedPodInterface, exists, err := w.sharedWatchers["Pod"].GetByKey(key)
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

	glog.Infof("Ready to sync: %s %s: %s/%s", eventAction, w.resourceType, namespace, name)
	syncData := &action.SyncData{
		Kind:      w.resourceType,
		Namespace: namespace,
		Name:      name,
		Action:    eventAction,
		Data:      d,
		OwnerUID:  ownerUID,
	}

	return syncData
}
