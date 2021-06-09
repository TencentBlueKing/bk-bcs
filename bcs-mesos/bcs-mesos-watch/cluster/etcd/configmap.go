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

package etcd

import (
	"fmt"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/labels"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions/bkbcs/v2"
)

//ConfigMapInfo wrapper for BCS ConfigMap
type ConfigMapInfo struct {
	data       *commtypes.BcsConfigMap
	syncTime   int64
	reportTime int64
}

//NewConfigMapWatch create watch for BCS ConfigMap
func NewConfigMapWatch(cxt context.Context, informer bkbcsv2.BcsConfigMapInformer, reporter cluster.Reporter) *ConfigMapWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*ConfigMapInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}

	return &ConfigMapWatch{
		report:    reporter,
		cancelCxt: cxt,
		informer:  informer,
		dataCache: cache.NewCache(keyFunc),
	}
}

//ConfigMapWatch watch for configmap, watch all detail and store to local cache
type ConfigMapWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	dataCache cache.Store      //cache for all app data
	watchPath string
	informer  bkbcsv2.BcsConfigMapInformer
}

//Work to add path and node watch
func (watch *ConfigMapWatch) Work() {
	blog.Infof("ConfigMapWatch start work")

	watch.ProcessAllConfigmaps()
	tick := time.NewTicker(12 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.Infof("ConfigMapWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("ConfigMapWatch is running")
			watch.ProcessAllConfigmaps()
		}
	}
}

//ProcessAllConfigmaps handle all configmap under all namespace
func (watch *ConfigMapWatch) ProcessAllConfigmaps() error {
	blog.V(3).Infof("sync all configmaps")
	currTime := time.Now().Unix()

	v2Cfgs, err := watch.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("list configmaps error %s", err.Error())
		return err
	}

	var numNode, numDel int
	for _, cfg := range v2Cfgs {
		numNode++
		configmap := &cfg.Spec.BcsConfigMap
		key := configmap.ObjectMeta.NameSpace + "." + configmap.ObjectMeta.Name
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
			blog.V(3).Infof("configmap %s is in cache, update sync", key)

			if reflect.DeepEqual(cacheDataInfo.data, configmap) {
				if cacheDataInfo.reportTime > currTime {
					cacheDataInfo.reportTime = currTime
				}
				if currTime-cacheDataInfo.reportTime > 180 {
					blog.Info("configmap %s data not changed, but long time not report, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, configmap)
					cacheDataInfo.reportTime = currTime
				}
			} else {
				blog.Info("configmap %s data changed, do report", key)
				watch.UpdateEvent(cacheDataInfo.data, configmap)
				cacheDataInfo.reportTime = currTime
			}

			cacheDataInfo.syncTime = currTime
			cacheDataInfo.data = configmap
		} else {
			blog.Info("configmap %s is not in cache, add, time(%d)", key, currTime)
			watch.AddEvent(configmap)
			dataInfo := new(ConfigMapInfo)
			dataInfo.data = configmap
			dataInfo.syncTime = currTime
			dataInfo.reportTime = currTime
			watch.dataCache.Add(dataInfo)
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

	blog.Info("sync %d configmaps from etcd, delete %d cache configmaps", numNode, numDel)

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
