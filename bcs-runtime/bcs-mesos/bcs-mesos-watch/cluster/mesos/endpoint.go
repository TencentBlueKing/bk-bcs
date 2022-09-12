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
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-runtime/bcs-mesos/bcs-mesos-watch/types"
)

// EndpointInfo wrapper for BCSEndpoint
type EndpointInfo struct {
	data       *commtypes.BcsEndpoint
	syncTime   int64
	reportTime int64
}

// NewEndpointWatch create endpoint watch
func NewEndpointWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter,
	watchPath string) *EndpointWatch {

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
		client:    client,
		watchPath: watchPath,
		dataCache: cache.NewCache(keyFunc),
	}
}

// EndpointWatch watch for Endpoint and store all datas to local cache
type EndpointWatch struct {
	eventLock sync.Mutex       // lock for event
	report    cluster.Reporter // reporter
	cancelCxt context.Context  // context for cancel
	client    ZkClient         // client for zookeeper
	dataCache cache.Store      // cache for all app data
	watchPath string
}

// Work handle all Endpoint datas periodically
func (watch *EndpointWatch) Work() {
	watch.ProcessAllEndpoints()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
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

// ProcessAllEndpoints handle all namespace Endpoint data
func (watch *EndpointWatch) ProcessAllEndpoints() error {

	currTime := time.Now().Unix()
	basePath := watch.watchPath + "/endpoint"
	blog.V(3).Infof("sync all endpoints under(%s), currTime(%d)", basePath, currTime)

	nmList, _, err := watch.client.GetChildrenEx(basePath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", basePath, err.Error())
		return err
	}
	if len(nmList) == 0 {
		blog.V(3).Infof("get empty namespace list under path(%s)", basePath)
		return nil
	}

	// sync all endpoints from zk and update cache, create add and update events
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
			data := new(commtypes.BcsEndpoint)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

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
				// watch.UpdateEvent(cacheDataInfo.data, data)
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

	blog.Info("sync %d endpoints from zk, delete %d cache endpoints", numZk, numDel)

	return nil
}

// AddEvent call when data added
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

// DeleteEvent when delete
func (watch *EndpointWatch) DeleteEvent(obj interface{}) {
	endpointData, ok := obj.(*commtypes.BcsEndpoint)
	if !ok {
		blog.Error("can not convert object to Bcsendpoint in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for Bcsendpoint %s.%s", endpointData.ObjectMeta.NameSpace,
		endpointData.ObjectMeta.Name)
	// report to cluster
	data := &types.BcsSyncData{
		DataType: "Endpoint",
		Action:   "Delete",
		Item:     obj,
	}
	watch.report.ReportData(data)
}

// UpdateEvent when update
func (watch *EndpointWatch) UpdateEvent(old, cur interface{}) {
	endpointData, ok := cur.(*commtypes.BcsEndpoint)
	if !ok {
		blog.Error("can not convert object to Bcsendpoint in UpdateEvent, object %v", cur)
		return
	}

	// if reflect.DeepEqual(old, cur) {
	//	blog.V(3).Infof("Bcsendpoint %s.%s data do not changed", endpointData.ObjectMeta.NameSpace, endpointData.ObjectMeta.Name)
	//	return
	// }
	blog.V(3).Infof("EVENT:: Update Event for Bcsendpoint %s.%s", endpointData.ObjectMeta.NameSpace,
		endpointData.ObjectMeta.Name)

	// report to cluster
	data := &types.BcsSyncData{
		DataType: "Endpoint",
		Action:   "Update",
		Item:     cur,
	}
	watch.report.ReportData(data)
}
