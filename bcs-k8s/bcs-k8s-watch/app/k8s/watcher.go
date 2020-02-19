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
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/ssl"
	"github.com/parnurzeal/gorequest"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
	netservicetypes "bk-bcs/bcs-services/bcs-netservice/pkg/netservice/types"
)

const (
	// defaultSyncInterval is default sync interval.
	defaultSyncInterval = 30 * time.Second

	// defaultNetServiceTimeout is default netservice timeout.
	defaultNetServiceTimeout = 2 * time.Second

	// defaultHTTPRetryerCount is default http request retry count.
	defaultHTTPRetryerCount = 2

	// defaultHTTPRetryerTime is default http request retry time.
	defaultHTTPRetryerTime = time.Second
)

// Watcher watchs target type resource metadata from k8s cluster,
// and write to storage by synchronizer with series actions.
type Watcher struct {
	resourceType       string
	resourceNamespaced bool
	controller         cache.Controller
	store              cache.Store
	writer             *output.Writer
	sharedWatchers     map[string]WatcherInterface
	stopChan           chan struct{}
}

// NewWatcher creates a new watcher of target type resource.
func NewWatcher(client *rest.Interface, resourceType string, resourceName string, objType runtime.Object,
	writer *output.Writer, sharedWatchers map[string]WatcherInterface, resourceNamespaced bool) *Watcher {

	watcher := &Watcher{
		resourceType:       resourceType,
		writer:             writer,
		sharedWatchers:     sharedWatchers,
		resourceNamespaced: resourceNamespaced,
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

// GetByKey returns object data by target key.
func (w *Watcher) GetByKey(key string) (interface{}, bool, error) {
	return w.store.GetByKey(key)
}

// ListKeys returns all keys in local store.
func (w *Watcher) ListKeys() []string {
	return w.store.ListKeys()
}

// Run starts the watcher.
func (w *Watcher) Run(stopCh <-chan struct{}) {
	// run controller.
	w.controller.Run(stopCh)
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
	// compare the object changes for update.
	if reflect.DeepEqual(oldObj, newObj) {
		newObjMetadata := newObj.(metav1.Object)

		// there is no changes, no need to update.
		glog.V(2).Infof("watcher got the same ResourceType[%s]: %s/%s",
			w.resourceType, newObjMetadata.GetNamespace(), newObjMetadata.GetName())
		return
	}

	// skip unnecessary node update event to reduce writer-queues pressure.
	if w.resourceType == "Node" {
		oldNode := oldObj.(*v1.Node)
		newNode := newObj.(*v1.Node)

		// NOTE: a best way is to use deepcopy function, save the common fields,
		// update the change fields.

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

// isEventShouldFilter filters k8s system events.
func (w *Watcher) isEventShouldFilter(meta metav1.Object, eventAction string) bool {
	// NOTE: event not support delete
	// bugfix here: must in top of this func, in case of Name or Namespace return true.
	if eventAction == action.SyncDataActionDelete && w.resourceType == ResourceTypeEvent {
		// Event not support delete.
		return true
	}

	if meta.GetNamespace() == "kube-system" && w.resourceType == ResourceTypeEvent {
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

	// construct and send
	dMeta := obj.(metav1.Object)
	namespace := dMeta.GetNamespace()
	name := dMeta.GetName()

	if w.isEventShouldFilter(dMeta, eventAction) {
		glog.V(2).Infof("watcher metadata is filtered %s %s: %s/%s", eventAction, w.resourceType, namespace, name)
		return nil
	}

	ownerUID := ""
	// NOTE: 生成时, 就确认了是否会告警, 即 ownerUID != "", 告警
	// if Event, get OwnerReference uid
	if w.resourceType == ResourceTypeEvent {
		event, ok := obj.(*v1.Event)
		if !ok {
			glog.Infof("watcher parse object to Event failed, %v", obj)
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
		Data:      obj,
		OwnerUID:  ownerUID,
	}

	return syncData
}

// NetServiceWatcher watchs resources in netservice, and sync to storage.
type NetServiceWatcher struct {
	clusterID      string
	storageService *bcs.InnerService
	netservice     *bcs.InnerService
	action         *action.StorageAction
}

// NewNetServiceWatcher creates a new NetServiceWatcher instance.
func NewNetServiceWatcher(clusterID string, storageService, netservice *bcs.InnerService) *NetServiceWatcher {
	w := &NetServiceWatcher{
		clusterID:      clusterID,
		storageService: storageService,
		netservice:     netservice,
		action:         action.NewStorageAction(clusterID, "", storageService),
	}
	return w
}

func (w *NetServiceWatcher) httpClient(httpConfig *bcs.HTTPClientConfig) (*gorequest.SuperAgent, error) {
	request := gorequest.New().Set("Accept", "application/json").Set("BCS-ClusterID", w.clusterID)

	if httpConfig.Scheme == "https" {
		tlsConfig, err := ssl.ClientTslConfVerity(httpConfig.CAFile, httpConfig.CertFile,
			httpConfig.KeyFile, httpConfig.Password)

		if err != nil {
			return nil, fmt.Errorf("init tls fail [clientConfig=%v, errors=%s]", tlsConfig, err)
		}
		request = request.TLSClientConfig(tlsConfig)
	}

	return request, nil
}

func (w *NetServiceWatcher) queryIPResource() (*netservicetypes.NetResponse, error) {
	targets := w.netservice.Servers()
	serversCount := len(targets)

	if serversCount == 0 {
		return nil, errors.New("netservice server list is empty, there is no available services now")
	}

	var httpClientConfig *bcs.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = targets[0]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = targets[index]
	}

	request, err := w.httpClient(httpClientConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create netservice client, %+v, %+v", httpClientConfig, err)
	}

	url := fmt.Sprintf("%s/v1/pool/%s", httpClientConfig.URL, w.clusterID)
	response := &netservicetypes.NetResponse{}

	if _, _, err := request.
		Timeout(defaultNetServiceTimeout).
		Get(url).
		Retry(defaultHTTPRetryerCount, defaultHTTPRetryerTime, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(response); err != nil {
		return nil, fmt.Errorf("request to netservice, get ip resource failed, %+v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("request to netservice, get ip resource failed, code[%d], message[%s]",
			response.Code, response.Message)
	}
	return response, nil
}

func (w *NetServiceWatcher) queryIPResourceDetail() (*netservicetypes.NetResponse, error) {
	targets := w.netservice.Servers()
	serversCount := len(targets)

	if serversCount == 0 {
		return nil, errors.New("netservice server list is empty, there is no available services now")
	}

	var httpClientConfig *bcs.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = targets[0]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = targets[index]
	}

	request, err := w.httpClient(httpClientConfig)
	if err != nil {
		return nil, fmt.Errorf("can't create netservice client, %+v, %+v", httpClientConfig, err)
	}

	url := fmt.Sprintf("%s/v1/pool/%s?info=detail", httpClientConfig.URL, w.clusterID)
	response := &netservicetypes.NetResponse{}

	if _, _, err := request.
		Timeout(defaultNetServiceTimeout).
		Get(url).
		Retry(defaultHTTPRetryerCount, defaultHTTPRetryerTime, http.StatusBadRequest, http.StatusInternalServerError).
		EndStruct(response); err != nil {
		return nil, fmt.Errorf("request to netservice, get ip resource detail failed, %+v", err)
	}

	if response.Code != 0 {
		return nil, fmt.Errorf("request to netservice, get ip resource detail failed, code[%d], message[%s]",
			response.Code, response.Message)
	}
	return response, nil
}

// Sync syncs target ip resources to storages.
func (w *NetServiceWatcher) SyncIPResource() {
	// query resource from netservice.
	resource, err := w.queryIPResource()
	if err != nil {
		glog.Warnf("sync netservice ip resource, query from netservice failed, %+v", err)
		return
	}

	// only sync ip pool static information.
	if resource.Type != netservicetypes.ResponseType_PSTATIC {
		glog.Warnf("sync netservice ip resource, query from netservice, invalid response type[%+v]", resource.Type)
		return
	}

	// sync ip resource.
	metadata := &action.SyncData{
		Name:   "IPPoolStatic-" + w.clusterID,
		Kind:   "IPPoolStatic",
		Action: action.SyncDataActionUpdate,
		Data:   resource.Data,
	}
	w.action.Update(metadata)
}

// Sync syncs target ip resource detail to storages.
func (w *NetServiceWatcher) SyncIPResourceDetail() {
	// query resource detail from netservice.
	resource, err := w.queryIPResourceDetail()
	if err != nil {
		glog.Warnf("sync netservice ip resource detail, query from netservice failed, %+v", err)
		return
	}

	// only sync ip pool detail information.
	if resource.Type != netservicetypes.ResponseType_POOL {
		glog.Warnf("sync netservice ip resource detail, query from netservice, invalid response type[%+v]", resource.Type)
		return
	}

	// sync ip resource detail.
	metadata := &action.SyncData{
		Name:   "IPPoolStaticDetail-" + w.clusterID,
		Kind:   "IPPoolStaticDetail",
		Action: action.SyncDataActionUpdate,
		Data:   resource.Data,
	}
	w.action.Update(metadata)
}

// Run starts the netservice watcher.
func (w *NetServiceWatcher) Run(stopCh <-chan struct{}) {
	// sync ip resource.
	go wait.NonSlidingUntil(w.SyncIPResource, defaultSyncInterval, stopCh)

	// sync ip resource detail.
	go wait.NonSlidingUntil(w.SyncIPResourceDetail, defaultSyncInterval, stopCh)

	// TODO: add more resource-sync logics here.
}
