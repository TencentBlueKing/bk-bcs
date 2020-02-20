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

	"k8s.io/apimachinery/pkg/labels"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/pkg/cache"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	bkbcsv2 "bk-bcs/bcs-mesos/pkg/client/informers/bkbcs/v2"

	"golang.org/x/net/context"
)

//EndpointInfo wrapper for BCSEndpoint
type EndpointInfo struct {
	data       *commtypes.BcsEndpoint
	syncTime   int64
	reportTime int64
}

//NewEndpointWatch create endpoint watch
func NewEndpointWatch(cxt context.Context, informer bkbcsv2.BcsEndpointInformer, reporter cluster.Reporter) *EndpointWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*EndpointInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}

	return &EndpointWatch{
		report:    reporter,
		cancelCxt: cxt,
		dataCache: cache.NewCache(keyFunc),
		informer:  informer,
	}
}

//EndpointWatch watch for Endpoint and store all datas to local cache
type EndpointWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	dataCache cache.Store      //cache for all app data
	watchPath string
	informer  bkbcsv2.BcsEndpointInformer
}

//Work handle all Endpoint datas periodically
func (watch *EndpointWatch) Work() {
	watch.ProcessAllEndpoints()
	tick := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.V(3).Infof("EndpointWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("EndpointWatch is running")
			watch.ProcessAllEndpoints()
		}
	}
}

//ProcessAllEndpoints handle all namespace Endpoint data
func (watch *EndpointWatch) ProcessAllEndpoints() error {

	currTime := time.Now().Unix()
	blog.V(3).Infof("sync all endpoints, currTime(%d)", currTime)

	v2Endpoints, err := watch.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("list bcsendpoints error %s", err.Error())
		return err
	}

	var numNode, numDel int
	for _, end := range v2Endpoints {
		numNode++
		data := &end.Spec.BcsEndpoint
		key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("get EndpointInfo %s from cache return err:%s", key, err.Error())
			continue
		}
		if exist == true {
			cacheDataInfo, ok := cacheData.(*EndpointInfo)
			if !ok {
				blog.Error("convert cachedata to EndpointInfo fail, key(%s)", key)
				continue
			}
			blog.V(3).Infof("endpoint %s is in cache, update sync time(%d)", key, currTime)
			//watch.UpdateEvent(cacheDataInfo.data, data)
			if reflect.DeepEqual(cacheDataInfo.data, data) {
				if cacheDataInfo.reportTime > currTime {
					cacheDataInfo.reportTime = currTime
				}
				if currTime-cacheDataInfo.reportTime > 180 {
					blog.Info("endpoint %s data not changed, but long time not report, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, data)
					cacheDataInfo.reportTime = currTime
				}
			} else {
				blog.Info("endpoint %s data changed, do report", key)
				watch.UpdateEvent(cacheDataInfo.data, data)
				cacheDataInfo.reportTime = currTime
			}
			cacheDataInfo.syncTime = currTime
			cacheDataInfo.data = data
		} else {
			blog.Info("endpoint %s is not in cache, add, time(%d)", key, currTime)
			watch.AddEvent(data)
			dataInfo := new(EndpointInfo)
			dataInfo.data = data
			dataInfo.syncTime = currTime
			dataInfo.reportTime = currTime
			watch.dataCache.Add(dataInfo)
		}
	}

	// check cache, create delete events
	keyList := watch.dataCache.ListKeys()
	for _, key := range keyList {
		blog.V(3).Infof("to check cache endpoint %s", key)
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("endpoint %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("endpoint %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*EndpointInfo)
		if !ok {
			blog.Error("convert cachedata to endpointInfo fail, key(%s)", key)
			continue
		}

		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Info("endpoint %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			watch.DeleteEvent(cacheDataInfo.data)
			watch.dataCache.Delete(cacheDataInfo)
		}
	}

	blog.Info("sync %d endpoints from etcd, delete %d cache endpoints", numNode, numDel)

	return nil
}

//AddEvent call when data added
func (watch *EndpointWatch) AddEvent(obj interface{}) {
	endpointData, ok := obj.(*commtypes.BcsEndpoint)
	if !ok {
		blog.Error("can not convert object to Bcsendpoint in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for Bcsendpoint %s.%s", endpointData.ObjectMeta.NameSpace, endpointData.ObjectMeta.Name)

	data := &types.BcsSyncData{
		DataType: "Endpoint",
		Action:   "Add",
		Item:     obj,
	}
	watch.report.ReportData(data)
}

//DeleteEvent when delete
func (watch *EndpointWatch) DeleteEvent(obj interface{}) {
	endpointData, ok := obj.(*commtypes.BcsEndpoint)
	if !ok {
		blog.Error("can not convert object to Bcsendpoint in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for Bcsendpoint %s.%s", endpointData.ObjectMeta.NameSpace, endpointData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Endpoint",
		Action:   "Delete",
		Item:     obj,
	}
	watch.report.ReportData(data)
}

//UpdateEvent when update
func (watch *EndpointWatch) UpdateEvent(old, cur interface{}) {
	endpointData, ok := cur.(*commtypes.BcsEndpoint)
	if !ok {
		blog.Error("can not convert object to Bcsendpoint in UpdateEvent, object %v", cur)
		return
	}

	//if reflect.DeepEqual(old, cur) {
	//	blog.V(3).Infof("Bcsendpoint %s.%s data do not changed", endpointData.ObjectMeta.NameSpace, endpointData.ObjectMeta.Name)
	//	return
	//}
	blog.V(3).Infof("EVENT:: Update Event for Bcsendpoint %s.%s", endpointData.ObjectMeta.NameSpace, endpointData.ObjectMeta.Name)

	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Endpoint",
		Action:   "Update",
		Item:     cur,
	}
	watch.report.ReportData(data)
}
