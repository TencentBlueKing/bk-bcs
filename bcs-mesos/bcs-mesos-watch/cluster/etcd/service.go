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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions/bkbcs/v2"

	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/labels"
)

//ServiceInfo wrapper for BCSService
type ServiceInfo struct {
	data       *commtypes.BcsService
	syncTime   int64
	reportTime int64
}

//NewServiceWatch create watch for Service
func NewServiceWatch(cxt context.Context, informer bkbcsv2.BcsServiceInformer, reporter cluster.Reporter) *ServiceWatch {

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
		informer:  informer,
		dataCache: cache.NewCache(keyFunc),
	}
}

//ServiceWatch watch all event for Service and store in local cache
type ServiceWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	dataCache cache.Store      //cache for all app data
	watchPath string
	informer  bkbcsv2.BcsServiceInformer
}

//Work list all Service data periodically
func (watch *ServiceWatch) Work() {
	blog.Infof("ServiceWatch start work")

	watch.ProcessAllServices()
	tick := time.NewTicker(8 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.Infof("ServiceWatch asked to exit")
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
	blog.V(3).Infof("sync all services, currTime(%d)", currTime)

	v2Svc, err := watch.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("list configmaps error %s", err.Error())
		return err
	}

	var numNode, numDel int
	for _, svc := range v2Svc {
		numNode++
		service := &svc.Spec.BcsService
		key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
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
			if reflect.DeepEqual(cacheDataInfo.data, service) {
				if cacheDataInfo.reportTime > currTime {
					cacheDataInfo.reportTime = currTime
				}
				if currTime-cacheDataInfo.reportTime > 180 {
					blog.Info("service %s data not changed, but long time not report, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, service)
					cacheDataInfo.reportTime = currTime
				}
			} else {
				blog.Info("service %s data changed, do report", key)
				watch.UpdateEvent(cacheDataInfo.data, service)
				cacheDataInfo.reportTime = currTime
			}
			cacheDataInfo.syncTime = currTime
			cacheDataInfo.data = service
		} else {
			blog.Info("service %s is not in cache, add, time(%d)", key, currTime)
			watch.AddEvent(service)
			dataInfo := new(ServiceInfo)
			dataInfo.data = service
			dataInfo.syncTime = currTime
			dataInfo.reportTime = currTime
			watch.dataCache.Add(dataInfo)
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

	blog.Info("sync %d services from etcd, delete %d cache services", numNode, numDel)
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
