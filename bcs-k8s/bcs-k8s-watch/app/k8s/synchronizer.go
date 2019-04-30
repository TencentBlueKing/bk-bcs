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

	glog "bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/bcs"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/action"
	"bk-bcs/bcs-k8s/bcs-k8s-watch/app/output/http"
)

type Synchronizer struct {
	watchers       map[string]WatcherInterface
	StorageService *bcs.StorageService
	ClusterID      string
}

func (sync *Synchronizer) Run(stop <-chan struct{}) {
	glog.Info("wait for datawatcher to be ready")
	duration := 5 * 60 * time.Second

	for {
		time.Sleep(duration)

		select {
		case <-stop:
			glog.Warn("Synchronizer is stopped by signal....")
			return
		default:
		}

		synced := true
		for _, watcher := range sync.watchers {
			w := watcher.(*Watcher)
			if !w.controller.HasSynced() {
				glog.Warn("Synchronizer is waiting all the watch controller synced. skip now...")
				synced = false
				break
			}
		}

		if !synced {
			continue
		}

		// sync cluster resource
		clusterResourceTypes := []string{"Namespace", "Node"}
		for _, clusterResourceType := range clusterResourceTypes {
			glog.Info("begin to sync %s", clusterResourceType)
			sync.SyncClusterResource(clusterResourceType, sync.watchers[clusterResourceType].(*Watcher))
			glog.Info("sync %s done", clusterResourceType)
		}
		namespaces := sync.watchers["Namespace"].(*Watcher).store.ListKeys()

		// sync  namespace resource
		namespaceResourceTypes := []string{"Pod", "ReplicationController", "Service", "EndPoints", "ConfigMap", "Secret",
			"Deployment", "Ingress", "ReplicaSet", "DaemonSet", "Job", "StatefulSet"}

		for _, namespaceResourceType := range namespaceResourceTypes {
			glog.Info("begin to sync %s", namespaceResourceType)
			sync.SyncNamespaceResource(namespaceResourceType, namespaces, sync.watchers[namespaceResourceType].(*Watcher))
			glog.Info("sync %s done", namespaceResourceType)
		}
		// ignore event
	}
}

func (sync *Synchronizer) SyncNamespaceResource(kind string, namespaces []string, watcher *Watcher) {

	// get all pods local
	localKeys := watcher.store.ListKeys()
	//glog.Infof("Sync %s got pod list from local: len=%d, data=%v", kind, len(localKeys), localKeys)
	glog.Infof("Sync %s got pod list from local: len=%d", kind, len(localKeys))

	totalData := []map[string]string{}

	for _, namespace := range namespaces {

		//innerData := []map[string]string{}
		data, err := sync.doRequest(namespace, kind)
		//glog.Infof("Sync %s: namespace=%s, data len=%d, total len=%d", kind, namespace, len(data), len(totalData))
		if err != nil {
			glog.Errorf("Sync %s fail: namespace=%s, type=Pod, err=%s", kind, namespace, err)
			continue
		}
		d, err := sync.doTransData(data)
		if err != nil {
			glog.Errorf("Sync %s fail, transData err: namespace=%s, type=Pod, err=%s", kind, namespace, err)
			continue
		}

		for _, record := range d {
			resourceName := record["resourceName"]
			totalData = append(totalData, map[string]string{"resourceName": resourceName, "namespace": namespace})
		}
	}
	//glog.Infof("Sync %s got list from storage: len=%d, data=%v", kind, len(totalData), totalData)
	glog.Infof("Sync %s got list from storage: len=%d", kind, len(totalData))

	sync.doSync(localKeys, totalData, watcher)

}

func (sync *Synchronizer) SyncClusterResource(kind string, watcher *Watcher) {
	data, err := sync.doRequest("", kind)
	if err != nil {
		glog.Errorf("SyncClusterResource %s fail: err=%s", kind, err)
		return
	}

	localKeys := watcher.store.ListKeys()

	d, err := sync.doTransData(data)
	if err != nil {
		glog.Errorf("SyncClusterResource %s fail, transData error: err=%s", kind, err)
		return
	}

	glog.Infof("SyncClusterResource got %s list from storage: %v", kind, data)
	sync.doSync(localKeys, d, watcher)

}

func (sync *Synchronizer) doTransData(data interface{}) (d []map[string]string, err error) {
	//d := []map[string]string{}
	j, err := jsoniter.Marshal(data)
	if err != nil {
		glog.Errorf("doSync fail: err=%s", err)
		return
	}
	err = jsoniter.Unmarshal(j, &d)
	if err != nil {
		glog.Errorf("doSync fail: err=%s", err)
		return
	}
	return
}

func (sync *Synchronizer) doSync(localKeys []string, data []map[string]string, watcher *Watcher) {
	/*
	   data = [
	   {"resourceName": "n1"},
	   {"resourceName": "n2"},
	   ]
	*/
	localKeysMap := map[string]string{}
	for _, localKey := range localKeys {
		//  locakKey = namespace/name  or name
		localKeysMap[localKey] = ""
	}

	storageKeysMap := map[string]string{}
	for _, record := range data {
		resourceName := record["resourceName"]
		namespace := record["namespace"]

		// format to  storageKey = namespace/name or name
		if namespace == "" {
			storageKeysMap[resourceName] = namespace
		} else {
			key := fmt.Sprintf("%s/%s", namespace, resourceName)
			storageKeysMap[key] = namespace
		}
		// key = namespace/resourceName  value = namespace
	}

	glog.Infof("sync got %s [local=%s, storage=%s]", watcher.resourceType, localKeysMap, storageKeysMap)

	// sync from local to storage
	for key := range localKeysMap {
		_, ok := storageKeysMap[key]
		if !ok {
			// to update
			item, exists, err := watcher.store.GetByKey(key)

			if exists && err == nil {
				syncData := watcher.genSyncData(item, "Add")
				// maybe filtered
				if syncData == nil {
					continue
				}
				watcher.writer.Sync(syncData)
			}
		}
	}

	// sync from storage with local
	for key, namespace := range storageKeysMap {
		_, ok := localKeysMap[key]
		if !ok {
			// to delete
			name := key
			if namespace != "" {
				namespaceNameList := strings.Split(key, "/")
				name = namespaceNameList[1]
			}
			glog.Infof("sync: %s: %s (name=%s) not on local, do delete", watcher.resourceType, key, name)

			syncData := &action.SyncData{
				Kind:      watcher.resourceType,
				Namespace: namespace,
				Name:      name,
				Action:    "Delete",
				Data:      "",
			}
			watcher.writer.Sync(syncData)
		}
	}
}

// get resource from storage, namespace can be empty
func (sync *Synchronizer) doRequest(namespace string, kind string) (data []interface{}, err error) {
	serversCount := len(sync.StorageService.Servers)
	if serversCount == 0 {
		// the process get address from zk not finished yet or there is no storage server on zk
		glog.Errorf("storage server list is empty! got no address yet")
		err = fmt.Errorf("storage server list is empty! got no address yet")
		return
	}

	keys := make([]string, 0, serversCount)
	for key := range sync.StorageService.Servers {
		keys = append(keys, key)
	}

	var httpClientConfig *bcs.HTTPClientConfig
	if serversCount == 1 {
		httpClientConfig = sync.StorageService.Servers[keys[0]]
	} else {
		index := rand.Intn(serversCount)
		httpClientConfig = sync.StorageService.Servers[keys[index]]
	}

	client := http.StorageClient{
		HTTPClientConfig: httpClientConfig,
		ClusterID:        sync.ClusterID,
		Namespace:        namespace,
		ResourceType:     kind,
		ResourceName:     "",
	}

	glog.V(2).Infof("sync request: namespace=%s, kind=%s, client.ResourceType=%s", namespace, kind, client.ResourceType)
	if namespace != "" {
		data, err = client.ListNamespaceResource()
	} else {
		data, err = client.ListClusterResource()
	}
	return

}
