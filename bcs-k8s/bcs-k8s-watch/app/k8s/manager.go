/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package k8s

import (
	"fmt"
	"strings"
	"time"

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/options"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output"
	clientGoCache "k8s.io/client-go/tools/cache"

	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/k8s/resources"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
	apiextensionsV1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crdClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions"
	k8sruntime "k8s.io/apimachinery/pkg/runtime"
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

	// target storage service.
	storageService *bcs.InnerService
}

// NewWatcherManager creates a new WatcherManager instance.
func NewWatcherManager(clusterID string, writer *output.Writer, k8sConfig *options.K8sConfig,
	storageService, netservice *bcs.InnerService, sc <-chan struct{}) (*WatcherManager, error) {

	mgr := &WatcherManager{
		watchers:       make(map[string]WatcherInterface),
		crdWatchers:    make(map[string]WatcherInterface),
		stopChan:       sc,
		writer:         writer,
		clusterID:      clusterID,
		storageService: storageService,
	}
	mgr.initWatchers(clusterID, k8sConfig, storageService, netservice)

	mgr.synchronizer = NewSynchronizer(clusterID, mgr.watchers, mgr.crdWatchers, storageService)
	return mgr, nil
}

func (mgr *WatcherManager) initWatchers(clusterID string,
	k8sconfig *options.K8sConfig, storageService, netservice *bcs.InnerService) {

	restConfig, err := resources.GetRestConfig(k8sconfig)
	if err != nil {
		panic(err)
	}

	// init k8s normal resource watchers
	for name, resourceObjType := range resources.WatcherConfigList {
		watcher := NewWatcher(resourceObjType.Client, name, resourceObjType.ResourceName, resourceObjType.ObjType, mgr.writer, mgr.watchers, resourceObjType.Namespaced) // nolint
		mgr.watchers[name] = watcher
	}

	// begin to watch kubefed to init kubefed watchers
	crdClient, err := crdClientSet.NewForConfig(restConfig)
	if err != nil {
		panic(err)
	}
	crdInformerFactory := externalversions.NewSharedInformerFactory(crdClient, time.Second*30)
	crdInformer := crdInformerFactory.Apiextensions().V1beta1().CustomResourceDefinitions()
	crdInformer.Lister()

	crdInformer.Informer().AddEventHandler(clientGoCache.ResourceEventHandlerFuncs{
		AddFunc:    mgr.AddEvent,
		UpdateFunc: mgr.UpdateEvent,
		DeleteFunc: mgr.DeleteEvent,
	})

	go crdInformerFactory.Start(mgr.stopChan)

	glog.Infof("Waiting for informer caches to sync")
	if ok := clientGoCache.WaitForCacheSync(mgr.stopChan, crdInformer.Informer().HasSynced); !ok {
		err := fmt.Errorf("failed to wait for kubefed caches to sync")
		panic(err)
	}

	// init netservice watcher.
	mgr.netserviceWatcher = NewNetServiceWatcher(clusterID, storageService, netservice)
}

func (mgr *WatcherManager) AddEvent(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1beta1.CustomResourceDefinition)
	if !ok {
		return
	}

	if strings.HasSuffix(crdObj.Spec.Group, ".kubefed.io") || crdObj.Spec.Group == resources.BkbcsGroupName {
		mgr.runCrdWatcher(crdObj)
		return
	}

}

func (mgr *WatcherManager) UpdateEvent(oldObj, newObj interface{}) {
	return
}

func (mgr *WatcherManager) DeleteEvent(obj interface{}) {
	crdObj, ok := obj.(*apiextensionsV1beta1.CustomResourceDefinition)
	if !ok {
		return
	}

	mgr.stopCrdWatcher(crdObj)
}

// runCrdWatcher run a crd watcher and writer handler
func (mgr *WatcherManager) runCrdWatcher(obj *apiextensionsV1beta1.CustomResourceDefinition) {
	groupVersion := obj.Spec.Group + "/" + obj.Spec.Version
	if kubefedClient, ok := resources.CrdClientList[groupVersion]; ok {
		var runtimeObject k8sruntime.Object
		var namespaced bool
		if obj.Spec.Scope == "Cluster" {
			namespaced = false
		} else if obj.Spec.Scope == "Namespaced" {
			namespaced = true
		}

		// init and run writer handler
		action := action.NewStorageAction(mgr.clusterID, obj.Spec.Names.Kind, mgr.storageService)
		mgr.writer.Handlers[obj.Spec.Names.Kind] = output.NewHandler(obj.Spec.Names.Kind, action)
		stopChan := make(chan struct{})
		mgr.writer.Handlers[obj.Spec.Names.Kind].Run(stopChan)

		// init and run watcher
		watcher := NewWatcher(&kubefedClient, obj.Spec.Names.Kind, obj.Spec.Names.Plural, runtimeObject, mgr.writer, mgr.watchers, namespaced) // nolint
		watcher.stopChan = stopChan
		mgr.crdWatchers[obj.Spec.Names.Kind] = watcher
		glog.Infof("watcher manager, start list-watcher[%+v]", obj.Spec.Names.Kind)
		go watcher.Run(watcher.stopChan)
	}
}

// stopCrdWatcher stop watcher and writer handler
func (mgr *WatcherManager) stopCrdWatcher(obj *apiextensionsV1beta1.CustomResourceDefinition) {

	if wc, ok := mgr.crdWatchers[obj.Spec.Names.Kind]; ok {
		watcher := wc.(*Watcher)
		glog.Infof("watcher manager, stop list-watcher[%+v]", obj.Spec.Names.Kind)
		close(watcher.stopChan)
		delete(mgr.crdWatchers, obj.Spec.Names.Kind)
		delete(mgr.writer.Handlers, obj.Spec.Names.Kind)
	}

}

// Run starts the watcher manager, and runs all watchers.
func (mgr *WatcherManager) Run(stopCh <-chan struct{}) {
	// run normal resource watchers.
	for resourceName, watcher := range mgr.watchers {
		glog.Infof("watcher manager, start list-watcher[%+v]", resourceName)
		go watcher.Run(stopCh)
	}

	// run netservice watcher.
	go mgr.netserviceWatcher.Run(stopCh)

	// run synchronizer.
	go mgr.synchronizer.Run(stopCh)
}

// StopCrdWatchers stop all crd watcher and writer handler
func (mgr *WatcherManager) StopCrdWatchers() {
	for _, wc := range mgr.crdWatchers {
		watcher := wc.(*Watcher)
		close(watcher.stopChan)
	}
}
