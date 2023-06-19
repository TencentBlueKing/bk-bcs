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
	"math/rand"
	"strings"
	"time"

	jsoniter "github.com/json-iterator/go"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/cache"

	glog "github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/bcs"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/k8s/resources"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/action"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-k8s-watch/app/output/http"
)

const (
	// defaultWatcherCheckInterval is default watcher check interval.
	defaultWatcherCheckInterval = 5 * time.Minute
)

// Synchronizer syncs resource metadata to storage.
type Synchronizer struct {
	// id of current cluster.
	clusterID string

	// 指定单个要同步的namespace
	namespace string

	// watchers that products metadata.
	watchers map[string]WatcherInterface

	// watchers of crd
	crdWatchers map[string]WatcherInterface

	// labelSelectors for different object
	labelSelectors map[string]string

	// target storage service.
	storageService *bcs.InnerService
}

// NewSynchronizer creates a new Synchronizer instance.
func NewSynchronizer(clusterID, namespace string, labelSelectors map[string]string,
	watchers, crdWatchers map[string]WatcherInterface, storageService *bcs.InnerService) *Synchronizer {
	return &Synchronizer{
		clusterID:      clusterID,
		watchers:       watchers,
		crdWatchers:    crdWatchers,
		labelSelectors: labelSelectors,
		storageService: storageService,
		namespace:      namespace,
	}
}

// Run starts the Synchronizer and make it keep sync resources in period.
func (sync *Synchronizer) Run(stopCh <-chan struct{}) {
	glog.Info("synchronizer waiting for watchers to be ready")

	for {
		time.Sleep(defaultWatcherCheckInterval)

		select {
		case <-stopCh:
			glog.Warn("synchronizer is stopped by signal...")
			return
		default:
		}

		if err := sync.RunOnce(); err != nil {
			glog.Errorf("synchronizer sync failed: %v", err)
		}
	}
}

// RunOnce sync resources once.
func (sync *Synchronizer) RunOnce() error {
	// check all watchers sync-state.
	for _, watcher := range sync.watchers {
		w := watcher.(*Watcher)
		var count = 0
		for {
			if count >= 10 {
				break
			}
			if w.controller.HasSynced() {
				break
			} else {
				time.Sleep(30 * time.Second)
			}
			count++
		}

		if count >= 10 {
			glog.Errorf("watcher %s is not synced, skip sync", w.resourceType)
			return fmt.Errorf("watcher %s is not synced, skip sync", w.resourceType)
		}
	}

	var namespaces []string
	if sync.namespace == "" {
		namespacesWatcher := sync.watchers["Namespace"]
		if namespacesWatcher != nil {
			namespaces = namespacesWatcher.(*Watcher).store.ListKeys()
		}
	} else {
		// 如果指定了namespace
		namespaces = []string{sync.namespace}
	}

	for resourceType, resourceObjType := range resources.K8sWatcherConfigList {
		labelSelector := sync.labelSelectors[resourceType]
		if curSelector, ok := sync.labelSelectors[resourceType]; ok {
			labelSelector = curSelector
		}
		if resourceObjType.Namespaced {
			glog.Info("begin to sync %s", resourceType)
			sync.syncNamespaceResource(resourceType, namespaces, labelSelector, sync.watchers[resourceType].(*Watcher))
			glog.Info("sync %s done", resourceType)
		} else {
			glog.Info("begin to sync %s", resourceType)
			sync.syncClusterResource(resourceType, labelSelector, sync.watchers[resourceType].(*Watcher))
			glog.Info("sync %s done", resourceType)
		}
	}

	for _, watcher := range sync.crdWatchers {
		w := watcher.(*Watcher)
		if !w.controller.HasSynced() {
			continue
		}
		if w.resourceNamespaced {
			glog.Info("begin to sync %s", w.resourceType)
			sync.syncNamespaceResource(w.resourceType, namespaces, "", w)
			glog.Info("sync %s done", w.resourceType)
		} else {
			glog.Info("begin to sync %s", w.resourceType)
			sync.syncClusterResource(w.resourceType, "", w)
			glog.Info("sync %s done", w.resourceType)
		}
	}

	return nil
}

func (sync *Synchronizer) syncNamespaceResource(kind string, namespaces []string, selector string, watcher *Watcher) {
	// get all resources from local store.

	localKeys := watcher.store.ListKeys()
	glog.Infof("Sync %s got list from local: len=%d", kind, len(localKeys))

	totalData := []map[string]string{}

	for _, namespace := range namespaces {
		data, err := sync.doRequest(namespace, selector, kind)
		if err != nil {
			glog.Errorf("Sync %s fail: namespace=%s, type=Pod, err=%s", kind, namespace, err)
			continue
		}

		d, err := sync.transData(data)
		if err != nil {
			glog.Errorf("Sync %s fail, transData err: namespace=%s, type=Pod, err=%s", kind, namespace, err)
			continue
		}

		for _, record := range d {
			resourceName := record["resourceName"]
			totalData = append(totalData, map[string]string{"resourceName": resourceName, "namespace": namespace})
		}
	}

	glog.Infof("Sync %s got list from storage: len=%d", kind, len(totalData))
	sync.doSync(localKeys, totalData, watcher)
}

func (sync *Synchronizer) syncClusterResource(kind, selector string, watcher *Watcher) {
	data, err := sync.doRequest("", selector, kind)
	if err != nil {
		glog.Errorf("sync cluster resource %s selector %s fail: err=%s", kind, selector, err)
		return
	}

	localKeys := watcher.store.ListKeys()

	d, err := sync.transData(data)
	if err != nil {
		glog.Errorf("sync cluster resource %s fail, transData error: err=%s", kind, err)
		return
	}

	glog.Infof("sync cluster resource got %s list from storage: %v", kind, data)
	sync.doSync(localKeys, d, watcher)
}

func (sync *Synchronizer) transData(data interface{}) (d []map[string]string, err error) {
	j, err := jsoniter.Marshal(data)
	if err != nil {
		glog.Errorf("transData fail: err=%s", err)
		return
	}

	err = jsoniter.Unmarshal(j, &d)
	if err != nil {
		glog.Errorf("transData fail: err=%s", err)
		return
	}
	// NOCC:nakedret/ret(设计如此:允许空返回值)
	return
}

func (sync *Synchronizer) doSync(localKeys []string, data []map[string]string, watcher *Watcher) {

	localKeysMap := map[string]string{}
	for _, localKey := range localKeys {
		// localKey = namespace/name or name.
		localKeysMap[localKey] = ""
	}

	storageKeysMap := map[string]string{}
	for _, record := range data {
		resourceName := record["resourceName"]
		namespace := record["namespace"]

		// format to storageKey = namespace/name or name.
		if namespace == "" {
			storageKeysMap[resourceName] = namespace
		} else {
			key := fmt.Sprintf("%s/%s", namespace, resourceName)
			storageKeysMap[key] = namespace
		}
		// key = namespace/resourceName, value = namespace.
	}

	glog.Infof("sync got %s [local=%d, storage=%d]", watcher.resourceType, len(localKeysMap), len(storageKeysMap))

	// sync from local to storage service.
	for key := range localKeysMap {
		if _, ok := storageKeysMap[key]; !ok {
			/* need to update */

			item, exists, err := watcher.store.GetByKey(key)
			if exists && err == nil {
				ns, name, _ := cache.SplitMetaNamespaceKey(key)
				// build sync data.
				syncData := watcher.genSyncData(types.NamespacedName{
					Name:      name,
					Namespace: ns,
				}, item, action.SyncDataActionAdd)

				if syncData == nil {
					// maybe filtered.
					continue
				}

				// sync add event base on the reconciliation logic.
				watcher.writer.Sync(syncData)
			}
		}
	}

	// sync from storage to local.
	for key, namespace := range storageKeysMap {
		if _, ok := localKeysMap[key]; !ok {
			/* need to delete */

			name := key
			if namespace != "" {
				namespaceNameList := strings.Split(key, "/")
				name = namespaceNameList[1]
			}

			_, exists, err := watcher.store.GetByKey(key)
			if !exists && err == nil {
				// Event not to delete.
				if watcher.resourceType == ResourceTypeEvent {
					continue
				}
				glog.Infof("sync: %s: %s (name=%s) not on local, do delete", watcher.resourceType, key, name)
				syncData := &action.SyncData{
					Kind:      watcher.resourceType,
					Namespace: namespace,
					Name:      name,
					Action:    action.SyncDataActionDelete,
					Data:      "",
					RequeueQ:  watcher.GetTriggerQueue(),
				}

				// sync delete event base on the reconciliation logic.
				watcher.writer.Sync(syncData)
			}
		}
	}
}

// doRequest xxx
// get resource from storage, namespace can be empty.
func (sync *Synchronizer) doRequest(namespace, selector, kind string) (data []interface{}, err error) {
	targets := sync.storageService.Servers()
	serversCount := len(targets)

	if serversCount == 0 {
		// the process get address from zk not finished yet or there is no storage server on zk.
		err = fmt.Errorf("storage server list is empty, got no address yet")
		glog.Errorf(err.Error())
		return
	}

	var httpClientConfig *bcs.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = targets[0]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = targets[index]
	}

	client := http.StorageClient{
		HTTPClientConfig: httpClientConfig,
		ClusterID:        sync.clusterID,
		Namespace:        namespace,
		ResourceType:     kind,
		ResourceName:     "",
	}

	glog.V(2).Infof("sync request: namespace=%s, labelselector=%s, kind=%s, client.ResourceType=%s",
		namespace, selector, kind, client.ResourceType)
	if namespace != "" {
		data, err = client.ListNsResourceWithLabelSelector(selector)
	} else {
		data, err = client.ListClusterResourceWithLabelSelector(selector)
	}
	// NOCC:nakedret/ret(设计如此:允许空返回值)
	return
}
