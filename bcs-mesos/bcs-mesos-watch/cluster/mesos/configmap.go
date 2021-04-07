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

package mesos

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
	//schedulertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

//NSControlInfo store all app info under one namespace
//type NSControlInfo struct {
//	path   string             //parent zk node, namespace absolute path
//	cxt    context.Context    //context for creating sub context
//	cancel context.CancelFunc //for cancel sub goroutine
//}

//ConfigMapInfo wrapper for BCS ConfigMap
type ConfigMapInfo struct {
	data       *commtypes.BcsConfigMap
	syncTime   int64
	reportTime int64
}

//NewConfigMapWatch create watch for BCS ConfigMap
func NewConfigMapWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter, watchPath string) *ConfigMapWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*ConfigMapInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}

	/*
		nsKeyFunc := func(data interface{}) (string, error) {
			ns, ok := data.(*NSControlInfo)
			if !ok {
				return "", fmt.Errorf("NSControlInfo type Assert failed")
			}
			return ns.path, nil
		}*/

	return &ConfigMapWatch{
		report:    reporter,
		cancelCxt: cxt,
		client:    client,
		watchPath: watchPath,
		dataCache: cache.NewCache(keyFunc),
		//nsCache:   cache.NewCache(nsKeyFunc),
	}
}

//ConfigMapWatch watch for configmap, watch all detail and store to local cache
type ConfigMapWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	client    ZkClient         //client for zookeeper
	dataCache cache.Store      //cache for all app data
	//nsCache   cache.Store      //all namespace path / namespace goroutine control info
	watchPath string
}

//Work to add path and node watch
func (watch *ConfigMapWatch) Work() {
	watch.ProcessAllConfigmaps()
	tick := time.NewTicker(12 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.V(3).Infof("ConfigMapWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("ConfigMapWatch is running")
			watch.ProcessAllConfigmaps()
		}
	}
}

//ProcessAllConfigmaps handle all configmap under all namespace
func (watch *ConfigMapWatch) ProcessAllConfigmaps() error {

	currTime := time.Now().Unix()
	basePath := watch.watchPath + "/configmap"
	blog.V(3).Infof("sync all configmaps under(%s), currTime(%d)", basePath, currTime)

	nmList, _, err := watch.client.GetChildrenEx(basePath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", basePath, err.Error())
		return err
	}
	if len(nmList) == 0 {
		blog.V(3).Infof("get empty namespace list under path(%s)", basePath)
		return nil
	}

	// sync all configmaps from zk and update cache, create add and update events
	numZk := 0
	numDel := 0
	for _, nmNode := range nmList {
		blog.V(3).Infof("get namespace node(%s) under path(%s)", nmNode, basePath)
		nmPath := basePath + "/" + nmNode
		nodeList, _, err := watch.client.GetChildrenEx(nmPath)
		if err != nil {
			blog.Error("get children nodes under %s err: %s", nmPath, err.Error())
			continue
		}
		for _, oneNode := range nodeList {
			numZk++
			blog.V(3).Infof("get node(%s) under path(%s)", oneNode, nmPath)
			nodePath := nmPath + "/" + oneNode
			byteData, _, err := watch.client.GetEx(nodePath)
			if err != nil {
				blog.Error("Get %s data err: %s", nodePath, err.Error())
				continue
			}
			data := new(commtypes.BcsConfigMap)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

			key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
			cacheData, exist, err := watch.dataCache.GetByKey(key)
			if err != nil {
				blog.Error("get configmap %s from cache return err:%s", key, err.Error())
				continue
			}
			if exist == true {
				cacheDataInfo, ok := cacheData.(*ConfigMapInfo)
				if !ok {
					blog.Error("convert cachedata to ConfigMapInfo fail, key(%s)", key)
					continue
				}
				blog.V(3).Infof("configmap %s is in cache, update sync time(%d)", key, currTime)
				//watch.UpdateEvent(cacheDataInfo.data, data)
				if reflect.DeepEqual(cacheDataInfo.data, data) {
					if cacheDataInfo.reportTime > currTime {
						cacheDataInfo.reportTime = currTime
					}
					if currTime-cacheDataInfo.reportTime > 180 {
						blog.Info("configmap %s data not changed, but long time not report, do report", key)
						watch.UpdateEvent(cacheDataInfo.data, data)
						cacheDataInfo.reportTime = currTime
					}
				} else {
					blog.Info("configmap %s data changed, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, data)
					cacheDataInfo.reportTime = currTime
				}

				cacheDataInfo.syncTime = currTime
				cacheDataInfo.data = data
			} else {
				blog.Info("configmap %s is not in cache, add, time(%d)", key, currTime)
				watch.AddEvent(data)
				dataInfo := new(ConfigMapInfo)
				dataInfo.data = data
				dataInfo.syncTime = currTime
				dataInfo.reportTime = currTime
				watch.dataCache.Add(dataInfo)
			}
		}
	}

	// check cache, create delete events
	keyList := watch.dataCache.ListKeys()
	for _, key := range keyList {
		blog.V(3).Infof("to check cache configmap %s", key)
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("configmap %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("configmap %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*ConfigMapInfo)
		if !ok {
			blog.Error("convert cachedata to ConfigMapInfo fail, key(%s)", key)
			continue
		}

		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Info("configmap %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			watch.DeleteEvent(cacheDataInfo.data)
			watch.dataCache.Delete(cacheDataInfo)
		}
	}

	blog.Info("sync %d configmaps from zk, delete %d cache configmaps", numZk, numDel)

	return nil
}

//AddEvent call when data added
func (watch *ConfigMapWatch) AddEvent(obj interface{}) {
	configmapData, ok := obj.(*commtypes.BcsConfigMap)
	if !ok {
		blog.Error("can not convert object to BcsConfigMap in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for BcsConfigMap %s.%s", configmapData.ObjectMeta.NameSpace, configmapData.ObjectMeta.Name)

	data := &types.BcsSyncData{
		DataType: "ConfigMap",
		Action:   types.ActionAdd,
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionAdd, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionAdd, cluster.SyncSuccess)
	}
}

//DeleteEvent when delete
func (watch *ConfigMapWatch) DeleteEvent(obj interface{}) {
	configmapData, ok := obj.(*commtypes.BcsConfigMap)
	if !ok {
		blog.Error("can not convert object to BcsConfigMap in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for BcsConfigMap %s.%s", configmapData.ObjectMeta.NameSpace, configmapData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "ConfigMap",
		Action:   types.ActionDelete,
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionDelete, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionDelete, cluster.SyncSuccess)
	}
}

//UpdateEvent when update
func (watch *ConfigMapWatch) UpdateEvent(old, cur interface{}) {
	configmapData, ok := cur.(*commtypes.BcsConfigMap)
	if !ok {
		blog.Error("can not convert object to BcsConfigMap in UpdateEvent, object %v", cur)
		return
	}

	//if reflect.DeepEqual(old, cur) {
	//	blog.V(3).Infof("BcsConfigMap %s.%s data do not changed", configmapData.ObjectMeta.NameSpace, configmapData.ObjectMeta.Name)
	//	return
	//}

	blog.V(3).Infof("EVENT:: Update Event for BcsConfigMap %s.%s", configmapData.ObjectMeta.NameSpace, configmapData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "ConfigMap",
		Action:   types.ActionUpdate,
		Item:     cur,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionUpdate, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeCfg, types.ActionUpdate, cluster.SyncSuccess)
	}
}
