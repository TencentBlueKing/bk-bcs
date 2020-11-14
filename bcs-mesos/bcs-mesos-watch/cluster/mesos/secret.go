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
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"reflect"
	"sync"
	"time"

	"golang.org/x/net/context"
)

//SecretInfo wrapper for BCSSecret
type SecretInfo struct {
	data       *commtypes.BcsSecret
	syncTime   int64
	reportTime int64
}

//NewSecretWatch create SecretWatch for data synchronization
func NewSecretWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter, watchPath string) *SecretWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*SecretInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}

	return &SecretWatch{
		report:    reporter,
		cancelCxt: cxt,
		client:    client,
		watchPath: watchPath,
		dataCache: cache.NewCache(keyFunc),
	}
}

//SecretWatch watch all secret data and store in local cache
type SecretWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	client    ZkClient         //client for zookeeper
	dataCache cache.Store      //cache for all app data
	watchPath string
}

//Work list all namespace secret periodically
func (watch *SecretWatch) Work() {
	watch.ProcessAllSecrets()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.V(3).Infof("SecretWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("SecretWatch is running")
			watch.ProcessAllSecrets()
		}
	}
}

//ProcessAllSecrets handle all namespaces data
func (watch *SecretWatch) ProcessAllSecrets() error {

	currTime := time.Now().Unix()
	basePath := watch.watchPath + "/secret"
	blog.V(3).Infof("sync all secrets under(%s), currTime(%d)", basePath, currTime)

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
			data := new(commtypes.BcsSecret)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

			key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
			cacheData, exist, err := watch.dataCache.GetByKey(key)
			if err != nil {
				blog.Error("get secret %s from cache return err:%s", key, err.Error())
				continue
			}
			if exist == true {
				cacheDataInfo, ok := cacheData.(*SecretInfo)
				if !ok {
					blog.Error("convert cachedata to SecretInfo fail, key(%s)", key)
					continue
				}
				blog.V(3).Infof("secret %s is in cache, update sync time(%d)", key, currTime)
				//watch.UpdateEvent(cacheDataInfo.data, data)
				if reflect.DeepEqual(cacheDataInfo.data, data) {
					if cacheDataInfo.reportTime > currTime {
						cacheDataInfo.reportTime = currTime
					}
					if currTime-cacheDataInfo.reportTime > 180 {
						blog.Info("secret %s data not changed, but long time not report, do report", key)
						watch.UpdateEvent(cacheDataInfo.data, data)
						cacheDataInfo.reportTime = currTime
					}
				} else {
					blog.Info("secret %s data changed, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, data)
					cacheDataInfo.reportTime = currTime
				}
				cacheDataInfo.syncTime = currTime
				cacheDataInfo.data = data
			} else {
				blog.Info("secret %s is not in cache, add, time(%d)", key, currTime)
				watch.AddEvent(data)
				dataInfo := new(SecretInfo)
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
		blog.V(3).Infof("to check cache secret %s", key)
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("secret %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("secret %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*SecretInfo)
		if !ok {
			blog.Error("convert cachedata to SecretInfo fail, key(%s)", key)
			continue
		}

		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Info("secret %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			watch.DeleteEvent(cacheDataInfo.data)
			watch.dataCache.Delete(cacheDataInfo)
		}
	}

	blog.Info("sync %d secrets from zk, delete %d cache secrets", numZk, numDel)

	return nil
}

//AddEvent call when data added
func (watch *SecretWatch) AddEvent(obj interface{}) {
	secretData, ok := obj.(*commtypes.BcsSecret)
	if !ok {
		blog.Error("can not convert object to BcsSecret in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for BcsSecret %s.%s", secretData.ObjectMeta.NameSpace, secretData.ObjectMeta.Name)

	data := &types.BcsSyncData{
		DataType: "Secret",
		Action:   "Add",
		Item:     obj,
	}
	watch.report.ReportData(data)
}

//DeleteEvent when delete
func (watch *SecretWatch) DeleteEvent(obj interface{}) {
	secretData, ok := obj.(*commtypes.BcsSecret)
	if !ok {
		blog.Error("can not convert object to BcsSecret in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for BcsSecret %s.%s", secretData.ObjectMeta.NameSpace, secretData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Secret",
		Action:   "Delete",
		Item:     obj,
	}
	watch.report.ReportData(data)
}

//UpdateEvent when update
func (watch *SecretWatch) UpdateEvent(old, cur interface{}) {
	secretData, ok := cur.(*commtypes.BcsSecret)
	if !ok {
		blog.Error("can not convert object to BcsSecret in UpdateEvent, object %v", cur)
		return
	}

	//if reflect.DeepEqual(old, cur) {
	//	blog.V(3).Infof("BcsSecret %s.%s data do not changed", secretData.ObjectMeta.NameSpace, secretData.ObjectMeta.Name)
	//	return
	//}
	blog.V(3).Infof("EVENT:: Update Event for BcsSecret %s.%s", secretData.ObjectMeta.NameSpace, secretData.ObjectMeta.Name)

	//report to cluster
	data := &types.BcsSyncData{
		DataType: "Secret",
		Action:   "Update",
		Item:     cur,
	}
	watch.report.ReportData(data)
}
