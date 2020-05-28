/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package reconciler

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	jump "github.com/lithammer/go-jump-consistent-hash"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	cmdb "bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
	"bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
)

const (
	SENDER_NUMBER          = 10
	QUEUE_LENGTH           = 10000
	FETCH_MODULES_INTERVAL = 100
)

// Reconciler sync container data of one cluster to cmdb
type Reconciler struct {
	clusterInfo common.Cluster
	clusterID   string
	clusterType string
	resType     string

	// map to cache cmdb module id
	moduleIDMap map[string]int64
	// block for moduleIDMap
	moduleIDMapLock sync.Mutex

	queue chan PodEvent

	senders     []*Sender
	sendersLock sync.Mutex

	hasher *jump.Hasher

	storageClient storage.Interface
	cmdbClient    cmdb.ClientInterface

	fullSyncInterval int64
}

// NewReconciler create new reconciler
func NewReconciler(clusterInfo common.Cluster, storageClient storage.Interface,
	cmdbClient cmdb.ClientInterface, fullSyncInterval int64) (*Reconciler, error) {

	clusterID := clusterInfo.ClusterID
	// Create senders
	var senders []*Sender
	for i := 0; i < SENDER_NUMBER; i++ {
		senders = append(senders, NewSender(clusterInfo, int64(i), QUEUE_LENGTH, cmdbClient))
	}

	var clusterType string
	if strings.Contains(strings.ToLower(clusterID), common.ClusterTypeMesos) {
		clusterType = common.ClusterTypeMesos
	} else if strings.Contains(strings.ToLower(clusterID), common.ClusterTypeK8S) {
		clusterType = common.ClusterTypeK8S
	}
	if len(clusterType) == 0 {
		return nil, fmt.Errorf("invalid cluster type %s", clusterID)
	}
	resType := common.GetResTypeByClusterType(clusterType)

	reconciler := &Reconciler{
		clusterInfo:      clusterInfo,
		clusterID:        clusterID,
		clusterType:      clusterType,
		resType:          resType,
		senders:          senders,
		storageClient:    storageClient,
		cmdbClient:       cmdbClient,
		fullSyncInterval: fullSyncInterval,
		moduleIDMap:      make(map[string]int64),
		queue:            make(chan PodEvent, QUEUE_LENGTH),
		hasher:           jump.New(SENDER_NUMBER, jump.NewCRC64()),
	}
	if err := reconciler.fetchModules(); err != nil {
		return nil, err
	}
	return reconciler, nil
}

// Run run reconciler
func (r *Reconciler) Run(ctx context.Context) {
	for _, sender := range r.senders {
		go sender.Run(ctx)
	}

	go r.fetchModulesLoop(ctx)

	go r.transferLoop(ctx)

	go r.watchLoop(ctx)

	go r.fullSyncLoop(ctx)

	select {
	case <-ctx.Done():
		blog.Infof("%s context done", r.logPre())
	}
}

func (r *Reconciler) logPre() string {
	return fmt.Sprintf("[reconciler:%s]", r.clusterInfo.ClusterID)
}

func (r *Reconciler) decodeK8SPod(data json.RawMessage) (*common.Pod, error) {
	kPod := new(common.K8SPod)
	err := json.Unmarshal(data, kPod)
	if err != nil {
		return nil, err
	}

	newPod, err := common.ConvertK8SPod(r.clusterID, kPod)
	if err != nil {
		blog.Errorf("%s ConvertK8SPod failed, err %s", r.logPre(), err.Error())
		return nil, err
	}
	newPod.PodCluster = r.clusterInfo.ClusterID

	setName, okSet := kPod.ObjectMeta.Annotations[common.BCS_BKCMDB_ANNOTATIONS_SET_KEY]
	moduleName, okModule := kPod.ObjectMeta.Annotations[common.BCS_BKCMDB_ANNOTATIONS_MODULE_KEY]
	findModuleFlag := false
	if okSet && okModule {
		r.moduleIDMapLock.Lock()
		moduleID, ok := r.moduleIDMap[setName+"."+moduleName]
		r.moduleIDMapLock.Unlock()
		if ok {
			findModuleFlag = true
			newPod.ModuleID = moduleID
		}
	}
	if !findModuleFlag {
		r.moduleIDMapLock.Lock()
		moduleID, ok := r.moduleIDMap[common.BCS_BKCMDB_DEFAULT_SET_NAME+"."+common.BCS_BKCMDB_DEFAULT_MODLUE_NAME]
		r.moduleIDMapLock.Unlock()
		if ok {
			newPod.ModuleID = moduleID
		} else {
			blog.Warnf("%s no default set and module for bkbcs", r.logPre())
			return nil, fmt.Errorf("%s no default set and module for bkbcs", r.logPre())
		}
	}

	return newPod, nil
}

func (r *Reconciler) decodeMesosTaskgroup(data json.RawMessage) (*common.Pod, error) {
	taskgroup := new(commtypes.BcsPodStatus)
	err := json.Unmarshal(data, taskgroup)
	if err != nil {
		return nil, err
	}

	newPod, err := common.ConvertMesosPod(taskgroup)
	if err != nil {
		blog.Errorf("%s ConvertMesosPod failed, err %s", r.logPre(), err.Error())
		return nil, err
	}
	newPod.PodCluster = r.clusterInfo.ClusterID

	setName, okSet := taskgroup.Annotations["set.bkcmdb.bkbcs.tencent.com"]
	moduleName, okModule := taskgroup.Annotations["module.bkcmdb.bkbcs.tencent.com"]
	findModuleFlag := false
	if okSet && okModule {
		r.moduleIDMapLock.Lock()
		moduleID, ok := r.moduleIDMap[setName+"."+moduleName]
		r.moduleIDMapLock.Unlock()
		if ok {
			findModuleFlag = true
			newPod.ModuleID = moduleID
		}
	}
	if !findModuleFlag {
		r.moduleIDMapLock.Lock()
		moduleID, ok := r.moduleIDMap[common.BCS_BKCMDB_DEFAULT_SET_NAME+"."+common.BCS_BKCMDB_DEFAULT_MODLUE_NAME]
		r.moduleIDMapLock.Unlock()
		if ok {
			newPod.ModuleID = moduleID
		} else {
			blog.Warnf("%s no default set and module for bkbcs", r.logPre())
			return nil, fmt.Errorf("%s no default set and module for bkbcs", r.logPre())
		}
	}

	return newPod, nil
}

func (r *Reconciler) decodeCmdbPod(data json.RawMessage) (*common.Pod, error) {
	pod := new(common.Pod)
	err := json.Unmarshal(data, pod)
	if err != nil {
		blog.Errorf("%s Unmarshal cmdb pod %+v failed, err %s", r.logPre(), data, err.Error())
		return nil, err
	}
	return pod, nil
}

func (r *Reconciler) decodeStorageResource(data json.RawMessage) (*common.Pod, error) {
	var err error
	var pod *common.Pod
	switch r.resType {
	case common.ResourceTypePod:
		pod, err = r.decodeK8SPod(data)
		if err != nil {
			return nil, err
		}
	case common.ResourceTypeTaskgroup:
		pod, err = r.decodeMesosTaskgroup(data)
		if err != nil {
			return nil, err
		}
	}
	pod.BizID = r.clusterInfo.BizID
	return pod, nil
}

func (r *Reconciler) fetchModules() error {

	result, err := r.cmdbClient.SearchBusinessTopoWithStatistics(r.clusterInfo.BizID)
	if err != nil {
		blog.Errorf("%s fetch biz module ids failed, err %s", r.logPre(), err.Error())
		return fmt.Errorf("%s fetch biz module ids failed, err %s", r.logPre(), err.Error())
	}
	if !result.Result {
		blog.Errorf("%s fetch biz module failed, resp %#v", r.logPre(), result)
		return fmt.Errorf("%s fetch biz module failed, resp %#v", r.logPre(), result)
	}

	moduleIDMap := make(map[string]int64)
	for _, bkBiz := range result.Data {
		for _, bkSet := range bkBiz.Child {
			for _, bkModule := range bkSet.Child {
				moduleIDMap[bkSet.InstName+"."+bkModule.InstName] = bkModule.InstID
			}
		}
	}

	r.moduleIDMapLock.Lock()
	r.moduleIDMap = moduleIDMap
	r.moduleIDMapLock.Unlock()

	return nil
}

func (r *Reconciler) fetchModulesLoop(ctx context.Context) error {
	ticker := time.NewTicker(FETCH_MODULES_INTERVAL * time.Second)

	for {
		select {
		case <-ticker.C:
			if err := r.fetchModules(); err != nil {
				blog.Errorf("%s fetchModules, err %s", r.logPre(), err.Error())
			}
		case <-ctx.Done():
			blog.Infof("%s context done, stop fetch modules loop", r.logPre())
			return nil
		}
	}
}

func (r *Reconciler) doCompare() error {

	resourcesData, err := r.storageClient.ListResources(r.clusterType, r.clusterID, r.resType)
	if err != nil {
		return fmt.Errorf("%s list storage resources failed, err %s", r.logPre(), err.Error())
	}

	storagePods := make(map[string]*common.Pod)
	for _, data := range resourcesData.Data {
		pod, err := r.decodeStorageResource(data.Data)
		if err != nil {
			return err
		}
		storagePods[pod.PodUUID] = pod
	}

	cmdbRes, err := r.cmdbClient.ListClusterPods(r.clusterInfo.BizID, r.clusterID)
	cmdbPods := make(map[string]*common.Pod)
	for _, data := range cmdbRes.Data.Info {
		pod, err := r.decodeCmdbPod(data)
		if err != nil {
			return err
		}
		cmdbPods[pod.PodUUID] = pod
	}

	adds, updates, dels := common.GetDiffPods(cmdbPods, storagePods)

	for _, add := range adds {
		blog.Info("%s full sync event %d, pod %s", r.logPre(), EventAdd, add.MetadataString())
		index := r.hasher.Hash(add.PodUUID)
		r.senders[index].Push(PodEvent{
			Type: EventAdd,
			Pod:  add,
		})
	}
	for _, update := range updates {
		blog.Info("%s full sync event %d, pod %s", r.logPre(), EventUpdate, update.MetadataString())
		index := r.hasher.Hash(update.PodUUID)
		r.senders[index].Push(PodEvent{
			Type: EventUpdate,
			Pod:  update,
		})
	}

	for _, del := range dels {
		blog.Info("%s full sync event %d, pod %s", r.logPre(), EventDel, del.MetadataString())
		index := r.hasher.Hash(del.PodUUID)
		r.senders[index].Push(PodEvent{
			Type: EventDel,
			Pod:  del,
		})
	}

	return nil
}

// fullSyncLoop full sync loop
func (r *Reconciler) fullSyncLoop(ctx context.Context) {

	// lock event queue
	r.sendersLock.Lock()

	// sync all pods event to sync queue
	err := r.doCompare()

	// unlock event queue
	r.sendersLock.Unlock()

	if err != nil {
		blog.Warnf("%s do compare failed, err %s", r.logPre(), err.Error())
	}

	ticker := time.NewTicker(time.Duration(r.fullSyncInterval) * time.Second)

	for {
		select {
		case <-ticker.C:
			// lock event queue
			r.sendersLock.Lock()

			// sync all pods event to sync queue
			err := r.doCompare()

			// unlock event queue
			r.sendersLock.Unlock()

			if err != nil {
				blog.Warnf("%s do compare failed, err %s", r.logPre(), err.Error())
			}

		case <-ctx.Done():
			blog.Infof("%s reconciler %s context done, stop full sync loop", r.logPre(), r.clusterID)
			return
		}
	}
}

// transferLoop sync loop
func (r *Reconciler) transferLoop(ctx context.Context) {
	for {
		select {
		case e := <-r.queue:
			blog.Infof("%s watch event: %d, pod: %s", r.logPre(), e.Type, e.Pod.MetadataString())
			index := r.hasher.Hash(e.Pod.PodUUID)
			r.sendersLock.Lock()
			r.senders[index].Push(e)
			r.sendersLock.Unlock()

		case <-ctx.Done():
			blog.Infof("%s context done, stop transfer loop", r.logPre())
			return
		}
	}
}

// watchLoop loop for watch storage
func (r *Reconciler) watchLoop(ctx context.Context) {

	ch, err := r.storageClient.WatchClusterResources(r.clusterID, r.resType)
	if err != nil {
		blog.Warnf("%s watch cluster %s failed, err %s", r.logPre(), r.clusterID, err.Error())
		time.Sleep(2 * time.Second)
		go r.watchLoop(ctx)
		return
	}
	for {
		select {
		case e := <-ch:
			newEvent := PodEvent{}
			switch e.Type {
			case common.Add:
				pod, err := r.decodeStorageResource(e.Value.Data)
				if err != nil {
					blog.Warnf("%s decode storage resource failed, err %s", r.logPre(), err.Error())
					continue
				}
				newEvent.Type = EventAdd
				newEvent.Pod = pod
				r.queue <- newEvent
			case common.Chg:
				pod, err := r.decodeStorageResource(e.Value.Data)
				if err != nil {
					blog.Warnf("%s decode storage resource failed, err %s", r.logPre(), err.Error())
					continue
				}
				newEvent.Type = EventUpdate
				newEvent.Pod = pod
				r.queue <- newEvent
			case common.Del:
				pod, err := r.decodeStorageResource(e.Value.Data)
				if err != nil {
					blog.Warnf("%s decode storage resource failed, err %s", r.logPre(), err.Error())
					continue
				}
				newEvent.Type = EventDel
				newEvent.Pod = pod
				r.queue <- newEvent

			case common.Brk:
				blog.Warnf("%s recv Brk event from storage", r.logPre())
				go r.watchLoop(ctx)
				return
			}

			r.queue <- newEvent

		case <-ctx.Done():
			blog.Infof("%s context done, stop watch loop", r.logPre())
			return
		}
	}
}
