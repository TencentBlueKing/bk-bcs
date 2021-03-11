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
)

//ServiceInfo wrapper for BCSService
type ServiceInfo struct {
	data       *commtypes.BcsService
	syncTime   int64
	reportTime int64
}

//NewServiceWatch create watch for Service
func NewServiceWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter, watchPath string) *ServiceWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*ServiceInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}
	return &ServiceWatch{
		report:    reporter,
		cancelCxt: cxt,
		client:    client,
		watchPath: watchPath,
		dataCache: cache.NewCache(keyFunc),
	}
}

//ServiceWatch watch all event for Service and store in local cache
type ServiceWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	client    ZkClient         //client for zookeeper
	dataCache cache.Store      //cache for all app data
	watchPath string
}

//Work list all Service data periodically
func (watch *ServiceWatch) Work() {
	watch.ProcessAllServices()
	tick := time.NewTicker(8 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.V(3).Infof("ServiceWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("ServiceWatch is running")
			watch.ProcessAllServices()
		}
	}
}

//ProcessAllServices handle all namespace service
func (watch *ServiceWatch) ProcessAllServices() error {

	currTime := time.Now().Unix()
	basePath := watch.watchPath + "/service"
	blog.V(3).Infof("sync all services under(%s), currTime(%d)", basePath, currTime)

	nmList, _, err := watch.client.GetChildrenEx(basePath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", basePath, err.Error())
		return err
	}
	if len(nmList) == 0 {
		blog.V(3).Infof("get empty namespace list under path(%s)", basePath)
		return nil
	}

	// sync all secrets from zk and update cache, create add and update events
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
			data := new(commtypes.BcsService)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

			key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
			cacheData, exist, err := watch.dataCache.GetByKey(key)
			if err != nil {
				blog.Error("get service %s from cache return err:%s", key, err.Error())
				continue
			}
			if exist == true {
				cacheDataInfo, ok := cacheData.(*ServiceInfo)
				if !ok {
					blog.Error("convert cachedata to ServiceInfo fail, key(%s)", key)
					continue
				}
				blog.V(3).Infof("service %s is in cache, update sync time(%d)", key, currTime)
				//watch.UpdateEvent(cacheDataInfo.data, data)
				if reflect.DeepEqual(cacheDataInfo.data, data) {
					if cacheDataInfo.reportTime > currTime {
						cacheDataInfo.reportTime = currTime
					}
					if currTime-cacheDataInfo.reportTime > 180 {
						blog.Info("service %s data not changed, but long time not report, do report", key)
						watch.UpdateEvent(cacheDataInfo.data, data)
						cacheDataInfo.reportTime = currTime
					}
				} else {
					blog.Info("service %s data changed, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, data)
					cacheDataInfo.reportTime = currTime
				}
				cacheDataInfo.syncTime = currTime
				cacheDataInfo.data = data
			} else {
				blog.Info("service %s is not in cache, add, time(%d)", key, currTime)
				watch.AddEvent(data)
				dataInfo := new(ServiceInfo)
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
		blog.V(3).Infof("to check cache service %s", key)
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("service %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("service %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*ServiceInfo)
		if !ok {
			blog.Error("convert cachedata to ServiceInfo fail, key(%s)", key)
			continue
		}

		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Info("service %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			watch.DeleteEvent(cacheDataInfo.data)
			watch.dataCache.Delete(cacheDataInfo)
		}
	}

	blog.Info("sync %d services from zk, delete %d cache services", numZk, numDel)

	return nil
}

//AddEvent call when data added
func (watch *ServiceWatch) AddEvent(obj interface{}) {
	serviceData, ok := obj.(*commtypes.BcsService)
	if !ok {
		blog.Error("can not convert object to BcsService in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for BcsService %s.%s", serviceData.ObjectMeta.NameSpace, serviceData.ObjectMeta.Name)

	data := &types.BcsSyncData{
		DataType: "Service",
		Action:   "Add",
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionAdd, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionAdd, cluster.SyncSuccess)
	}
}

//DeleteEvent when delete
func (watch *ServiceWatch) DeleteEvent(obj interface{}) {
	serviceData, ok := obj.(*commtypes.BcsService)
	if !ok {
		blog.Error("can not convert object to BcsService in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for BcsService %s.%s", serviceData.ObjectMeta.NameSpace, serviceData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Service",
		Action:   "Delete",
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionDelete, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionDelete, cluster.SyncSuccess)
	}
}

//UpdateEvent when update
func (watch *ServiceWatch) UpdateEvent(old, cur interface{}) {
	serviceData, ok := cur.(*commtypes.BcsService)
	if !ok {
		blog.Error("can not convert object to BcsService in UpdateEvent, object %v", cur)
		return
	}
	//if reflect.DeepEqual(old, cur) {
	//	blog.V(3).Infof("BcsService %s.%s data do not changed", serviceData.ObjectMeta.NameSpace, serviceData.ObjectMeta.Name)
	//	return
	//}
	blog.V(3).Infof("EVENT:: Update Event for BcsService %s.%s", serviceData.ObjectMeta.NameSpace, serviceData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Service",
		Action:   "Update",
		Item:     cur,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionUpdate, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeSvr, types.ActionUpdate, cluster.SyncSuccess)
	}
}
