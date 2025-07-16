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

package k8s

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"strings"
	"sync"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/ssl"
	netservicetypes "github.com/Tencent/bk-bcs/bcs-common/pkg/bcsapi/netservice"
	"github.com/parnurzeal/gorequest"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/pkg/metrics"
)

const (
	// defaultWatcherQueueTime for watcher queue metrics collect
	// defaultWatcherQueueTime = 3 * time.Second
	eventQueueBackoffBaseDuration = 1 * time.Second
	eventQueueBackoffMaxDuration  = 32 * time.Second

	// defaultSyncInterval is default sync interval.
	defaultSyncInterval = 30 * time.Second

	// defaultNetServiceTimeout is default netservice timeout.
	defaultNetServiceTimeout = 20 * time.Second

	// defaultHTTPRetryerCount is default http request retry count.
	defaultHTTPRetryerCount = 2

	// defaultHTTPRetryerTime is default http request retry time.
	defaultHTTPRetryerTime = time.Second

	// defaultQueueTimeout is default timeout of queue.
	defaultQueueTimeout = 1 * time.Second
)

// Watcher watchs target type resource metadata from k8s cluster,
// and write to storage by synchronizer with series actions.
type Watcher struct {
	resourceType       string
	resourceNamespaced bool
	// queue              *queue.Queue
	eventQueue       workqueue.RateLimitingInterface
	controller       cache.Controller
	store            cache.Store
	writer           *output.Writer
	sharedWatchers   map[string]WatcherInterface
	stopChan         chan struct{}
	namespace        string
	labelSelector    string
	labelMap         map[string]string
	namespaceFilters map[string]struct{}
	nameFilters      map[string]struct{}
	dataMaskers      []Masker
	storageSynced    bool
}

// WatcherOptions provide options for create Watcher
type WatcherOptions struct {
	DynamicClient    *dynamic.DynamicClient
	Namespace        string
	ResourceType     string
	GroupVersion     string
	ResourceName     string
	ObjType          runtime.Object
	Writer           *output.Writer
	SharedWatchers   map[string]WatcherInterface
	IsNameSpaced     bool
	LabelSelector    string
	NamespaceFilters []string
	NameFilters      []string
	MaskerConfigs    []options.MaskerConfig
}

// Validate validate WatcherOptions
func (wo *WatcherOptions) Validate() error {
	if wo.DynamicClient == nil {
		return fmt.Errorf("DynamicClient is nil in WatcherOptions")
	}

	if wo.Writer == nil {
		return fmt.Errorf("Writer is nil in WatcherOptions")
	}

	if wo.SharedWatchers == nil {
		return fmt.Errorf("SharedWatchers is nil in WatcherOptions")
	}

	return nil
}

// NewWatcher creates a new watcher of target type resource.
// nolint
func NewWatcher(wo *WatcherOptions) (*Watcher, error) {
	if wo == nil {
		return nil, fmt.Errorf("WatcherOptions can not be nil pointer")
	}

	if err := wo.Validate(); err != nil {
		return nil, err
	}

	labelSet, err := labels.ConvertSelectorToLabelsMap(wo.LabelSelector)
	if err != nil {
		return nil, err
	}
	watcher := &Watcher{
		resourceType:       wo.ResourceType,
		writer:             wo.Writer,
		sharedWatchers:     wo.SharedWatchers,
		resourceNamespaced: wo.IsNameSpaced,
		// queue:              queue.New(),
		eventQueue: workqueue.NewRateLimitingQueue(
			workqueue.NewItemExponentialFailureRateLimiter(
				eventQueueBackoffBaseDuration,
				eventQueueBackoffMaxDuration)),
		namespace:        wo.Namespace,
		labelSelector:    wo.LabelSelector,
		labelMap:         labelSet,
		namespaceFilters: map[string]struct{}{},
		nameFilters:      map[string]struct{}{},
		dataMaskers:      make([]Masker, 0),
		storageSynced:    false,
	}
	for _, ns := range wo.NamespaceFilters {
		watcher.namespaceFilters[ns] = struct{}{}
	}
	for _, name := range wo.NameFilters {
		watcher.nameFilters[name] = struct{}{}
	}
	for _, mc := range wo.MaskerConfigs {
		// watcher 保留与自己相关的masker
		if mc.Kind == wo.ResourceType {
			path := make([]string, len(mc.Path))
			copy(path, mc.Path)
			watcher.dataMaskers = append(watcher.dataMaskers, Masker{
				Namespace: mc.Namespace,
				Path:      path,
			})
		}
	}

	glog.Infof("NewWatcher with resource type: %s, resource name: %s, namespace: %s, labelSelector: %s",
		wo.ResourceType, wo.ResourceName, wo.Namespace, wo.LabelSelector)

	gv, err := schema.ParseGroupVersion(wo.GroupVersion)
	if err != nil {
		return nil, err
	}
	gvr := schema.GroupVersionResource{Group: gv.Group, Version: gv.Version, Resource: wo.ResourceName}

	var listWatch *cache.ListWatch
	if !wo.IsNameSpaced {
		// unnamespaced resource
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = watcher.labelSelector
				return (*wo.DynamicClient).Resource(gvr).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = watcher.labelSelector
				timeoutSeconds := int64(5 * time.Minute.Seconds() * (rand.Float64() + 1.0)) // nolint
				options.TimeoutSeconds = &timeoutSeconds
				return (*wo.DynamicClient).Resource(gvr).Watch(context.TODO(), options)
			},
		}
	} else {
		// wo.Namespace specified namespace, if "" watch all namespace
		listWatch = &cache.ListWatch{
			ListFunc: func(options metav1.ListOptions) (runtime.Object, error) {
				options.LabelSelector = watcher.labelSelector
				return (*wo.DynamicClient).Resource(gvr).Namespace(wo.Namespace).List(context.TODO(), options)
			},
			WatchFunc: func(options metav1.ListOptions) (watch.Interface, error) {
				options.LabelSelector = watcher.labelSelector
				timeoutSeconds := int64(5 * time.Minute.Seconds() * (rand.Float64() + 1.0)) // nolint
				options.TimeoutSeconds = &timeoutSeconds
				return (*wo.DynamicClient).Resource(gvr).Namespace(wo.Namespace).Watch(context.TODO(), options)
			},
		}
	}

	// register event handler.
	eventHandler := cache.ResourceEventHandlerFuncs{
		AddFunc:    watcher.AddEvent,
		UpdateFunc: watcher.UpdateEvent,
		DeleteFunc: watcher.DeleteEvent,
	}

	// build informer.
	store, controller := cache.NewInformer(listWatch, wo.ObjType, 0, eventHandler)
	watcher.store = store
	watcher.controller = controller

	return watcher, nil
}

// GetTriggerQueue returns queue for requeue retry object
func (w *Watcher) GetTriggerQueue() workqueue.RateLimitingInterface {
	return w.eventQueue
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
	// do with handler data
	// go w.handleQueueData(stopCh)

	// metrics collect watcher fifo queue length
	go wait.NonSlidingUntil(func() {
		metrics.ReportK8sWatcherQueueLength(w.resourceType, float64(w.eventQueue.Len()))
	}, time.Second*1, stopCh)

	// metrics collect watcher cache keys length
	go wait.NonSlidingUntil(func() {
		metrics.ReportK8sWatcherCacheKeys(w.resourceType, float64(len(w.ListKeys())))
	}, time.Second*1, stopCh)

	wg := &sync.WaitGroup{}
	wg.Add(5)
	for i := 0; i < 5; i++ {
		go func() {
			defer wg.Done()
			// Run a worker thread that just dequeues items, processes them, and marks them done.
			// It enforces that the reconcileHandler is never invoked concurrently with the same object.
			for w.processNextWorkItem() {
			}
		}()
	}
	go func() {
		<-stopCh
		w.eventQueue.ShutDown()
		glog.Warnf("event queue shut downed")
	}()

	// run controller.
	w.controller.Run(stopCh)
}

// distributeDataToHandler xxx
// distribute data to handler at watcher handlers.
func (w *Watcher) distributeDataToHandler(data *action.SyncData) {
	handlerKey := w.writer.GetHandlerKeyBySyncData(data)
	if handlerKey == "" {
		glog.Errorf("get handler key failed, resource: %s, namespace: %s, name: %s", data.Kind,
			data.Namespace, data.Name)
		return
	}

	if handler, ok := w.writer.Handlers[handlerKey]; ok {
		handler.HandleWithTimeout(data, defaultQueueTimeout)
	} else {
		glog.Errorf("can't distribute the normal metadata, unknown DataType[%+v]", data.Kind)
	}
}

// AddEvent is event handler for add resource event.
func (w *Watcher) AddEvent(obj interface{}) {
	dMeta, isObj := obj.(metav1.Object)
	if !isObj {
		glog.Errorf("Error casting to k8s metav1 object, new obj: %+v", obj)
		return
	}

	// ignore managedFields field
	if !options.IsWatchManagedFields {
		dMeta.SetManagedFields(nil)
	}

	item := types.NamespacedName{
		Name:      dMeta.GetName(),
		Namespace: dMeta.GetNamespace(),
	}
	w.eventQueue.Forget(item)
	w.eventQueue.Add(item)
}

// DeleteEvent is event handler for delete resource event.
func (w *Watcher) DeleteEvent(obj interface{}) {
	// Deal with tombstone events by pulling the object out.  Tombstone events wrap the object in a
	// DeleteFinalStateUnknown struct, so the object needs to be pulled out.
	// Copied from sample-controller
	// This should never happen if we aren't missing events, which we have concluded that we are not
	// and made decisions off of this belief.  Maybe this shouldn't be here?
	var ok bool
	if _, ok = obj.(metav1.Object); !ok {
		// If the object doesn't have Metadata, assume it is a tombstone object of type DeletedFinalStateUnknown
		tombstone, ok := obj.(cache.DeletedFinalStateUnknown)
		if !ok {
			glog.Errorf("Error decoding objects. Expected cache.DeletedFinalStateUnknown, obj type %s, obj %v",
				fmt.Sprintf("%T", obj), obj)
			return
		}
		// Set obj to the tombstone obj
		obj = tombstone.Obj
	}
	dMeta, isObj := obj.(metav1.Object)
	if !isObj {
		glog.Errorf("Error casting to k8s metav1 object, new obj: %+v", obj)
		return
	}

	// ignore managedFields field
	if !options.IsWatchManagedFields {
		dMeta.SetManagedFields(nil)
	}

	item := types.NamespacedName{
		Name:      dMeta.GetName(),
		Namespace: dMeta.GetNamespace(),
	}
	w.eventQueue.Forget(item)
	w.eventQueue.Add(item)
}

// UpdateEvent is event handler for update resource event.
func (w *Watcher) UpdateEvent(oldObj, newObj interface{}) {
	// convert to unstructured object
	oldUnstructuredObj, oOk := oldObj.(*unstructured.Unstructured)
	newUnstructuredObj, nOk := newObj.(*unstructured.Unstructured)
	if !oOk || !nOk {
		glog.Errorf("Error casting to k8s metav1 unstructured object, new obj: %+v", newObj)
		return
	}

	// compare the object changes for update.
	if reflect.DeepEqual(oldObj, newObj) {
		// there is no changes, no need to update.
		glog.V(2).Infof("watcher got the same ResourceType[%s]: %s/%s",
			w.resourceType, newUnstructuredObj.GetNamespace(), newUnstructuredObj.GetName())
		return
	}

	// skip unnecessary node update event to reduce writer-queues pressure.
	if w.resourceType == "Node" {
		// convert to corev1 object
		oldNode, newNode := &v1.Node{}, &v1.Node{}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			oldUnstructuredObj.UnstructuredContent(), oldNode); err != nil {
			glog.Errorf("Error casting to k8s corev1 object, old obj: %+v", oldObj)
			return
		}

		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(
			newUnstructuredObj.UnstructuredContent(), newNode); err != nil {
			glog.Errorf("Error casting to k8s corev1 object, new obj: %+v", newObj)
			return
		}

		if len(newNode.Status.Conditions) == len(oldNode.Status.Conditions) {
			// NOTE: a best way is to use deepcopy function, save the common fields,
			// update the change fields.
			var tempLastTimes = make([]metav1.Time, len(newNode.Status.Conditions))
			tempVersion := newNode.ResourceVersion
			newNode.ResourceVersion = oldNode.ResourceVersion

			for i := range newNode.Status.Conditions {
				tempLastTimes[i] = newNode.Status.Conditions[i].LastHeartbeatTime
				newNode.Status.Conditions[i].LastHeartbeatTime = oldNode.Status.Conditions[i].LastHeartbeatTime
			}

			// the first DeepEqual skips in obj level, the second DeepEqual skips
			// the node data after save common fields.
			if reflect.DeepEqual(oldNode, newNode) {
				glog.V(2).Infof("skip unnecessary node %s update event", newNode.GetName())
				return
			}
			// recover new node metadata after DeepEqual finally.
			newNode.ResourceVersion = tempVersion
			for i := range newNode.Status.Conditions {
				newNode.Status.Conditions[i].LastHeartbeatTime = tempLastTimes[i]
			}
		}
	}
	item := types.NamespacedName{
		Name:      newUnstructuredObj.GetName(),
		Namespace: newUnstructuredObj.GetNamespace(),
	}
	w.eventQueue.Forget(item)
	w.eventQueue.Add(item)
}

// processNextWorkItem will read a single work item off the workqueue and
// attempt to process it, by calling the reconcileHandler.
func (w *Watcher) processNextWorkItem() bool {
	obj, shutdown := w.eventQueue.Get()
	if shutdown {
		// Stop working
		return false
	}

	// We call Done here so the workqueue knows we have finished
	// processing this item. We also must remember to call Forget if we
	// do not want this work item being re-queued. For example, we do
	// not call Forget if a transient error occurs, instead the item is
	// put back on the workqueue and attempted again after a back-off
	// period.
	defer w.eventQueue.Done(obj)

	tObj := obj.(types.NamespacedName)
	key := tObj.Name
	if len(tObj.Namespace) > 0 {
		key = tObj.Namespace + "/" + tObj.Name
	}
	storeObj, isExisted, err := w.store.GetByKey(key)
	if err != nil {
		glog.Errorf("get store obj by key %s failed, requeue, err %s", key, err.Error())
		w.eventQueue.AddRateLimited(obj)
		return true
	}
	if !isExisted {
		data := w.genSyncData(tObj, nil, action.SyncDataActionDelete)
		if data == nil {
			// event should be filtered
			return true
		}
		w.distributeDataToHandler(data)
		w.eventQueue.Forget(obj)
		return true
	}
	data := w.genSyncData(tObj, storeObj, action.SyncDataActionUpdate)
	if data == nil {
		// event should be filtered
		return true
	}
	w.distributeDataToHandler(data)
	w.eventQueue.Forget(obj)
	return true
}

// isEventShouldFilter filters k8s system events.
func (w *Watcher) isEventShouldFilter(meta types.NamespacedName, eventAction string) bool {
	// NOTE: event not support delete
	// bugfix here: must in top of this func, in case of Name or Namespace return true.
	if eventAction == action.SyncDataActionDelete && w.resourceType == ResourceTypeEvent {
		// Event not support delete.
		return true
	}

	if meta.Namespace == "kube-system" && w.resourceType == ResourceTypeEvent {
		// kubeops start pod with those prefix.
		name := meta.Name
		if strings.HasPrefix(name, "kube-") ||
			strings.HasPrefix(name, "kubedns-") ||
			strings.HasPrefix(name, "nginx-proxy") ||
			strings.HasPrefix(name, "bcs-") {
			return false
		}
		return true
	}

	if _, isFilter := w.namespaceFilters[meta.Namespace]; isFilter {
		return true
	}
	if _, isFilter := w.nameFilters[meta.Name]; isFilter {
		return true
	}
	return false
}

func (w *Watcher) genSyncData(nsedName types.NamespacedName, obj interface{}, eventAction string) *action.SyncData {
	namespace := nsedName.Namespace
	name := nsedName.Name

	if w.isEventShouldFilter(nsedName, eventAction) {
		glog.V(2).Infof("watcher metadata is filtered %s %s: %s/%s", eventAction, w.resourceType, namespace, name)
		return nil
	}

	var dMeta *unstructured.Unstructured
	var isObj bool
	if obj != nil {
		dMeta, isObj = obj.(*unstructured.Unstructured)
		if !isObj {
			glog.Errorf("Error casting to unstructured Object, obj: %+v", obj)
			return nil
		}
		// must deepcopy obj before modify it, otherwise will panic
		dMeta = dMeta.DeepCopy()

		// don't remove this code
		// in a specific scenario, when using label selector to watch multiple sub-clusters of a karmada federated cluster,
		// returned data may not carry the label selector, so we add label selector into object returned.
		if len(w.labelMap) != 0 {
			tmpLabels := dMeta.GetLabels()
			if tmpLabels == nil {
				tmpLabels = make(map[string]string)
			}
			for k, v := range w.labelMap {
				tmpLabels[k] = v
			}
			dMeta.SetLabels(tmpLabels)
		}

		if !options.IsWatchManagedFields {
			dMeta.SetManagedFields(nil)
		}

		// mask data
		w.dataMasking(dMeta)
	}

	ownerUID := ""
	glog.Infof("Ready to sync: %s %s: %s/%s", eventAction, w.resourceType, namespace, name)
	syncData := &action.SyncData{
		Kind:      w.resourceType,
		Namespace: namespace,
		Name:      name,
		Action:    eventAction,
		Data:      dMeta,
		OwnerUID:  ownerUID,
		RequeueQ:  w.GetTriggerQueue(),
	}

	return syncData
}

// mask data by masker
func (w *Watcher) dataMasking(dMeta *unstructured.Unstructured) {
	if len(w.dataMaskers) == 0 {
		return
	}
	for _, m := range w.dataMaskers {
		m.MaskData(dMeta)
	}
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
		index := rand.Intn(serversCount) // nolint
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
		index := rand.Intn(serversCount) // nolint
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

// SyncIPResource syncs target ip resources to storages.
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
	_ = w.action.Update(metadata)
}

// SyncIPResourceDetail syncs target ip resource detail to storages.
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
	_ = w.action.Update(metadata)
}

// Run starts the netservice watcher.
func (w *NetServiceWatcher) Run(stopCh <-chan struct{}) {
	// sync ip resource.
	go wait.NonSlidingUntil(w.SyncIPResource, defaultSyncInterval*2, stopCh)

	// sync ip resource detail.
	go wait.NonSlidingUntil(w.SyncIPResourceDetail, defaultSyncInterval*2, stopCh)

	// Note: add more resource-sync logics here.
}
