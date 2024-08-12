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

// Package k8s xxx
package k8s

import (
	"fmt"
	"time"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/capabilities"
	apiextensionsV1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsV1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	crdClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/k8s/resources"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/action"
)

const (
	// CRDVersionSupportV1 is cluster capabilities for CustomResourceVersion V1
	CRDVersionSupportV1 = "v1"
	// CRDVersionSupportV1Beta1 is cluster capabilities for CustomResourceVersion V1Beta1
	CRDVersionSupportV1Beta1 = "v1beta1"
)

// WatcherManager is resource watcher manager.
type WatcherManager struct {
	// normal k8s resource watchers.
	watchers map[string]WatcherInterface

	// k8s kubefed watchers
	crdWatchers map[string]WatcherInterface

	// synchronizer syncs normal metadata got by watchers to storage.
	synchronizer *Synchronizer

	// special resource watchers.
	netserviceWatcher *NetServiceWatcher

	stopChan <-chan struct{}

	writer *output.Writer

	// id of current cluster.
	clusterID string
	// namespaces 指定监听的namespace资源，暂时单个
	watchResource *options.WatchResource
	// target storage service.
	storageService *bcs.InnerService
}

// WatcherManagerOptions options for WatcherManager
type WatcherManagerOptions struct {
	ClusterID      string
	WatchResource  *options.WatchResource
	FilterConfig   *options.FilterConfig
	Writer         *output.Writer
	K8sConfig      *options.K8sConfig
	StorageService *bcs.InnerService
	Netservice     *bcs.InnerService
	StopChan       <-chan struct{}
}

// clusterID string, watchResource *options.WatchResource, filterConfig *options.FilterConfig, writer *output.Writer,
// k8sConfig *options.K8sConfig,
// storageService, netservice *bcs.InnerService, sc <-chan struct{}

// NewWatcherManager creates a new WatcherManager instance.
func NewWatcherManager(wmo *WatcherManagerOptions) (*WatcherManager, error) {

	mgr := &WatcherManager{
		watchers:       make(map[string]WatcherInterface),
		crdWatchers:    make(map[string]WatcherInterface),
		stopChan:       wmo.StopChan,
		writer:         wmo.Writer,
		clusterID:      wmo.ClusterID,
		storageService: wmo.StorageService,
		watchResource:  wmo.WatchResource,
	}
	mgr.initWatchers(wmo.ClusterID, wmo.K8sConfig, wmo.FilterConfig, wmo.StorageService, wmo.Netservice)

	mgr.synchronizer = NewSynchronizer(wmo.ClusterID, wmo.WatchResource.Namespace, wmo.WatchResource.LabelSelectors,
		mgr.watchers, mgr.crdWatchers, wmo.StorageService)
	return mgr, nil
}

func (mgr *WatcherManager) initWatchers(clusterID string,
	k8sconfig *options.K8sConfig, filterConfig *options.FilterConfig, storageService, netservice *bcs.InnerService) {

	restConfig, err := resources.GetRestConfig(k8sconfig)
	if err != nil {
		panic(err)
	}

	// init k8s normal resource watchers
	for name, resourceObjType := range resources.K8sWatcherConfigList {
		labelSelector := ""
		// get labelSelector for the resourceType
		if val, ok := mgr.watchResource.LabelSelectors[name]; ok {
			labelSelector = val
		}
		watcher, err := NewWatcher(&WatcherOptions{
			DynamicClient:    resourceObjType.Client,
			Namespace:        mgr.watchResource.Namespace,
			ResourceType:     name,
			GroupVersion:     resourceObjType.GroupVersion,
			ResourceName:     resourceObjType.ResourceName,
			ObjType:          resourceObjType.ObjType,
			Writer:           mgr.writer,
			SharedWatchers:   mgr.watchers,
			IsNameSpaced:     resourceObjType.Namespaced,
			LabelSelector:    labelSelector,
			NamespaceFilters: filterConfig.NamespaceFilters,
			NameFilters:      filterConfig.NameFilters,
			MaskerConfigs:    filterConfig.DataMaskConfigList,
		})
		if err != nil {
			panic(err)
		}
		mgr.watchers[name] = watcher
	}

	if !mgr.watchResource.DisableCRD {
		// begin to watch crd to init crd watchers
		crdClient, err := crdClientSet.NewForConfig(restConfig)
		if err != nil {
			panic(err)
		}
		crdInformerFactory := externalversions.NewSharedInformerFactory(crdClient, time.Second*30)

		// check crd version supported in cluster
		crdVersion, err := mgr.capabilities(restConfig)
		if err != nil {
			if filterConfig.CrdVersionSupport == "" {
				glog.Errorf("get crd version from cluster failed and not set crd version supported in cluster, err: %s", err)
				panic(err)
			}
			glog.Warnf("get crd version from cluster failed, use config from file, err: %s", err)
			crdVersion = filterConfig.CrdVersionSupport
		}

		hasSyncedFunc := mgr.addCrdInformer(crdVersion, crdInformerFactory, restConfig)

		go crdInformerFactory.Start(mgr.stopChan)

		glog.Infof("Waiting for informer caches to sync")
		if ok := cache.WaitForCacheSync(mgr.stopChan, hasSyncedFunc); !ok {
			err := fmt.Errorf("failed to wait for crd caches to sync")
			panic(err)
		}
	}

	if !mgr.watchResource.DisableNetservice {
		// init netservice watcher.
		mgr.netserviceWatcher = NewNetServiceWatcher(clusterID, storageService, netservice)
	}
}

func (mgr *WatcherManager) capabilities(restConfig *rest.Config) (string, error) {
	discoveryClient, err := apiextensionsclient.NewForConfig(restConfig)
	if err != nil {
		return "", err
	}

	capabilities, err := capabilities.GetCapabilities(discoveryClient.Discovery())
	if err != nil {
		return "", fmt.Errorf("get kubernetes capabilities failed, err %s", err.Error())
	}

	glog.Infof("kubernetes capabilities %+v", capabilities.APIVersions)

	if !capabilities.APIVersions.Has("apiextensions.k8s.io/v1beta1") {
		glog.Infof("capabilities does not has apiextensions.k8s.io/v1beta1, create v1 version")
		return CRDVersionSupportV1, nil
	}
	glog.Infof("capabilities has apiextensions.k8s.io/v1beta1, create v1 beta version")
	return CRDVersionSupportV1Beta1, nil
}

func (mgr *WatcherManager) addCrdInformer(crdVersion string,
	crdInformerFactory externalversions.SharedInformerFactory, restConfig *rest.Config) cache.InformerSynced { // nolint
	glog.Infof("add crd informer with crd version: %s", crdVersion)
	// CustomResourceDefinition V1Beta1
	if crdVersion == CRDVersionSupportV1Beta1 {
		crdInformer := crdInformerFactory.Apiextensions().V1beta1().CustomResourceDefinitions()
		crdInformer.Lister()

		crdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
			AddFunc:    mgr.AddEventV1beta1,
			UpdateFunc: mgr.UpdateEventV1beta1,
			DeleteFunc: mgr.DeleteEventV1beta1,
		})
		return crdInformer.Informer().HasSynced
	}

	// CustomResourceDefinition V1
	crdInformer := crdInformerFactory.Apiextensions().V1().CustomResourceDefinitions()
	crdInformer.Lister()

	crdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.AddEventV1,
		UpdateFunc: mgr.UpdateEventV1,
		DeleteFunc: mgr.DeleteEventV1,
	})
	return crdInformer.Informer().HasSynced
}

// AddEventV1beta1 handles add event.
func (mgr *WatcherManager) AddEventV1beta1(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1beta1.CustomResourceDefinition)
	if !ok {
		return
	}

	crdObjCore := &CustomResourceDefinitionCore{
		TypeMeta:   crdObj.TypeMeta,
		ObjectMeta: crdObj.ObjectMeta,
		Spec: CustomResourceDefinitionCoreSpec{
			Group:       crdObj.Spec.Group,
			Version:     crdObj.Spec.Version,
			NamesPlural: crdObj.Spec.Names.Plural,
			NamesKind:   crdObj.Spec.Names.Kind,
			Scope:       ResourceScope(crdObj.Spec.Scope),
		},
	}

	glog.Infof("run watcher for crd: %s", crdObjCore.Name)
	mgr.runCrdWatcher(crdObjCore)

}

// AddEventV1 handles add event.
func (mgr *WatcherManager) AddEventV1(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1.CustomResourceDefinition)
	if !ok {
		return
	}

	if len(crdObj.Spec.Versions) < 1 {
		glog.Warnf("crdObj %s/%s , len(obj.Spec.Versions) is less than 1, obj: %+v\n", crdObj.Namespace, crdObj.Name, crdObj)
		return
	}

	for _, version := range crdObj.Spec.Versions {
		if version.Storage {
			crdObjCore := &CustomResourceDefinitionCore{
				TypeMeta:   crdObj.TypeMeta,
				ObjectMeta: crdObj.ObjectMeta,
				Spec: CustomResourceDefinitionCoreSpec{
					Group:       crdObj.Spec.Group,
					Version:     version.Name,
					NamesPlural: crdObj.Spec.Names.Plural,
					NamesKind:   crdObj.Spec.Names.Kind,
					Scope:       ResourceScope(crdObj.Spec.Scope),
				},
			}

			glog.Infof("run watcher for crd: %s", crdObjCore.Name)
			mgr.runCrdWatcher(crdObjCore)
		}
	}

}

// UpdateEventV1beta1 handles update event.
func (mgr *WatcherManager) UpdateEventV1beta1(oldObj, newObj interface{}) {
}

// UpdateEventV1 handles update event.
func (mgr *WatcherManager) UpdateEventV1(oldObj, newObj interface{}) {
}

// DeleteEventV1beta1 handles delete event.
func (mgr *WatcherManager) DeleteEventV1beta1(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1beta1.CustomResourceDefinition)
	if !ok {
		return
	}
	key := mgr.getCrdWatcherKey(crdObj.Spec.Group, crdObj.Spec.Version, crdObj.Spec.Names.Kind)
	mgr.stopCrdWatcher(key)
}

// DeleteEventV1 handles delete event.
func (mgr *WatcherManager) DeleteEventV1(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1.CustomResourceDefinition)
	if !ok {
		return
	}
	for _, version := range crdObj.Spec.Versions {
		if version.Storage {
			key := mgr.getCrdWatcherKey(crdObj.Spec.Group, version.Name, crdObj.Spec.Names.Kind)
			mgr.stopCrdWatcher(key)
		}
	}
}

// runCrdWatcher run a crd watcher and writer handler
func (mgr *WatcherManager) runCrdWatcher(obj *CustomResourceDefinitionCore) {
	groupVersion := obj.Spec.Group + "/" + obj.Spec.Version
	kind := obj.Spec.NamesKind
	crdName := obj.Name

	crdClient, ok := resources.CrdClientList[groupVersion]
	if !ok {
		glog.Infof("can not get client for CRD: %s, GVK: %s/%s", crdName, groupVersion, kind)
		return
	}
	glog.Infof("create watcher for CRD: %s, GVK: %s/%s", crdName, groupVersion, kind)

	var runtimeObject k8sruntime.Unstructured
	var namespaced bool
	if obj.Spec.Scope == "Cluster" {
		namespaced = false
	} else if obj.Spec.Scope == "Namespaced" {
		namespaced = true
	}

	// init and run writer handler
	if _, ok := mgr.writer.Handlers[obj.Spec.NamesKind]; !ok {
		// not close handler once create , to support multi groupversion but same kind
		action := action.NewStorageAction(mgr.clusterID, obj.Spec.NamesKind, mgr.storageService)
		mgr.writer.Handlers[obj.Spec.NamesKind] = output.NewHandler(mgr.clusterID, obj.Spec.NamesKind, action)
		mgr.writer.Handlers[obj.Spec.NamesKind].Run(make(chan struct{}))
	}

	labelSelector := ""
	// get labelSelector for the resourceType
	if val, ok := mgr.watchResource.LabelSelectors[obj.Spec.NamesKind]; ok {
		labelSelector = val
	}
	// init and run watcher
	gv := fmt.Sprintf("%s/%s", obj.Spec.Group, obj.Spec.Version)
	watcher, err := NewWatcher(&WatcherOptions{
		DynamicClient:    crdClient,
		Namespace:        mgr.watchResource.Namespace,
		ResourceType:     obj.Spec.NamesKind,
		GroupVersion:     gv,
		ResourceName:     obj.Spec.NamesPlural,
		ObjType:          runtimeObject,
		Writer:           mgr.writer,
		SharedWatchers:   mgr.watchers,
		IsNameSpaced:     namespaced,
		LabelSelector:    labelSelector,
		NamespaceFilters: []string{},
		NameFilters:      []string{},
	})
	if err != nil {
		panic(err)
	}
	stopChan := make(chan struct{})
	watcher.stopChan = stopChan
	key := mgr.getCrdWatcherKey(obj.Spec.Group, obj.Spec.Version, obj.Spec.NamesKind)
	mgr.crdWatchers[key] = watcher
	glog.Infof("watcher manager, start list-watcher[%+v]", obj.Spec.NamesKind)
	go watcher.Run(watcher.stopChan)

}

// stopCrdWatcher stop watcher and writer handler
func (mgr *WatcherManager) stopCrdWatcher(watcherKey string) {
	if wc, ok := mgr.crdWatchers[watcherKey]; ok {
		watcher := wc.(*Watcher)
		glog.Infof("watcher manager, stop list-watcher[%+v]", watcherKey)
		close(watcher.stopChan)
		delete(mgr.crdWatchers, watcherKey)
		// multi groupversion but same kind, should not delete writer.Handlers by kind when crd is deleted
		// delete(mgr.writer.Handlers, kind)
	}
}

// Run starts the watcher manager, and runs all watchers.
func (mgr *WatcherManager) Run(stopCh <-chan struct{}) {
	// run normal resource watchers.
	for resourceName, watcher := range mgr.watchers {
		glog.Infof("watcher manager, start list-watcher[%+v]", resourceName)
		go watcher.Run(stopCh)
	}

	if !mgr.watchResource.DisableNetservice {
		// run netservice watcher.
		go mgr.netserviceWatcher.Run(stopCh)
	}

	// synchronizer run once
	var count = 0
	for {
		if count >= 5 {
			panic("synchronizer run failed")
		}
		if err := mgr.synchronizer.RunOnce(); err != nil { // nolint
			glog.Errorf("synchronizer sync failed: %v", err)
			time.Sleep(5 * time.Minute)
		} else {
			glog.Infof("synchronizer sync done.")
			break
		}
		count++
	}
}

// StopCrdWatchers stop all crd watcher and writer handler
func (mgr *WatcherManager) StopCrdWatchers() {
	for _, wc := range mgr.crdWatchers {
		watcher := wc.(*Watcher)
		close(watcher.stopChan)
	}
}

func (mgr *WatcherManager) getCrdWatcherKey(group, version, kind string) string {
	return group + "/" + version + "/" + kind
}
