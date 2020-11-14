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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	cmdb "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
)

const (
	// SENDER_NUMBER default sender number for one reconciler
	SENDER_NUMBER = 10
	// QUEUE_LENGTH queue length for storage watch event queue
	QUEUE_LENGTH = 10000
	// FETCH_MODULES_INTERVAL interval for fetch biz modules from paas-cc
	FETCH_MODULES_INTERVAL = 100
)

// Reconciler sync container data of one cluster to cmdb
type Reconciler struct {
	// cluster info
	clusterInfo common.Cluster

	// cluster id
	clusterID string

	// cluster type, k8s or mesos
	clusterType string

	// resource type, pod or taskgroup
	resType string

	// map to cache cmdb module id
	moduleIDMap map[string]int64
	// block for moduleIDMap
	moduleIDMapLock sync.Mutex

	// resource event queue
	queue chan PodEvent

	// sender array
	senders []*Sender
	// sender lock, when locked, pod event will stay at event queue
	sendersLock sync.Mutex

	// storage client
	storageClient storage.Interface

	// cmdb client
	cmdbClient cmdb.ClientInterface

	// full sync interval
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

	// check cluster type from cluster id
	// TODO: some cluster with special name won't work
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
	}

	// should fetch the related modules when new reconciler is created
	if err := reconciler.fetchModules(); err != nil {
		return nil, err
	}
	return reconciler, nil
}

func (r *Reconciler) hash(key string) int32 {
	// hasher pod event to senders, to ensure the event order
	return jump.HashString(key, SENDER_NUMBER, jump.NewCRC64())
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

// log prefix for current reconciler
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

	// get pod annotations for bk cmdb
	// if there is no annotations for bk cmdb, save pod into default cmdb module bkbcs/bkbcs
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

	// get pod annotations for bk cmdb
	// if there is no annotations for bk cmdb, save pod into default cmdb module bkbcs/bkbcs
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
		index := r.hash(add.PodUUID)
		r.senders[index].Push(PodEvent{
			Type: EventAdd,
			Pod:  add,
		})
	}
	for _, update := range updates {
		blog.Info("%s full sync event %d, pod %s", r.logPre(), EventUpdate, update.MetadataString())
		index := r.hash(update.PodUUID)
		r.senders[index].Push(PodEvent{
			Type: EventUpdate,
			Pod:  update,
		})
	}

	for _, del := range dels {
		blog.Info("%s full sync event %d, pod %s", r.logPre(), EventDel, del.MetadataString())
		index := r.hash(del.PodUUID)
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
			index := r.hash(e.Pod.PodUUID)
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
