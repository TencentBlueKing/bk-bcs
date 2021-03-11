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
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	schedulertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

//NSControlInfo store all app info under one namespace
//type NSControlInfo struct {
//	path   string             //parent zk node, namespace absolute path
//	cxt    context.Context    //context for creating sub context
//	cancel context.CancelFunc //for cancel sub goroutine
//}

//DeploymentInfo wrapper for BCS Deployment
type DeploymentInfo struct {
	data       *schedulertypes.Deployment
	syncTime   int64
	reportTime int64
}

//NewDeploymentWatch create deployment watch
func NewDeploymentWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter, watchPath string) *DeploymentWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*DeploymentInfo)
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

	return &DeploymentWatch{
		report:    reporter,
		cancelCxt: cxt,
		client:    client,
		watchPath: watchPath,
		dataCache: cache.NewCache(keyFunc),
		//nsCache:   cache.NewCache(nsKeyFunc),
	}
}

//DeploymentWatch watch all deployment data and store to local cache
type DeploymentWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	client    ZkClient         //client for zookeeper
	dataCache cache.Store      //cache for all app data
	//nsCache   cache.Store     //all namespace path / namespace goroutine control info
	watchPath string
}

//Work to add path and node watch
func (watch *DeploymentWatch) Work() {
	watch.ProcessAllDeployments()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.V(3).Infof("DeploymentWatch asked to exit")
			return
		case <-tick.C:
			blog.V(3).Infof("DeploymentWatch is running")
			watch.ProcessAllDeployments()
		}
	}
}

//ProcessAllDeployments handle all namespace deployment data
func (watch *DeploymentWatch) ProcessAllDeployments() error {

	currTime := time.Now().Unix()
	basePath := watch.watchPath + "/deployment"
	blog.V(3).Infof("sync all deployments under(%s), currTime(%d)", basePath, currTime)

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
			data := new(schedulertypes.Deployment)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

			key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
			cacheData, exist, err := watch.dataCache.GetByKey(key)
			if err != nil {
				blog.Error("get deployment %s from cache return err:%s", key, err.Error())
				continue
			}
			if exist == true {
				cacheDataInfo, ok := cacheData.(*DeploymentInfo)
				if !ok {
					blog.Error("convert cachedata to DeploymentInfo fail, key(%s)", key)
					continue
				}
				blog.V(3).Infof("deployment %s is in cache, update sync time(%d)", key, currTime)
				//watch.UpdateEvent(cacheDataInfo.data, data)
				if reflect.DeepEqual(cacheDataInfo.data, data) {
					if cacheDataInfo.reportTime > currTime {
						cacheDataInfo.reportTime = currTime
					}
					if currTime-cacheDataInfo.reportTime > 180 {
						blog.Info("deployment %s data not changed, but long time not report, do report", key)
						watch.UpdateEvent(cacheDataInfo.data, data)
						cacheDataInfo.reportTime = currTime
					}
				} else {
					blog.Info("deployment %s data changed, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, data)
					cacheDataInfo.reportTime = currTime
				}

				cacheDataInfo.syncTime = currTime
				cacheDataInfo.data = data
			} else {
				blog.Info("deployment %s is not in cache, add, time(%d)", key, currTime)
				watch.AddEvent(data)
				dataInfo := new(DeploymentInfo)
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
		blog.V(3).Infof("to check cache deployment %s", key)
		cacheData, exist, err := watch.dataCache.GetByKey(key)
		if err != nil {
			blog.Error("deployment %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("deployment %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*DeploymentInfo)
		if !ok {
			blog.Error("convert cachedata to DeploymentInfo fail, key(%s)", key)
			continue
		}

		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Info("deployment %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			watch.DeleteEvent(cacheDataInfo.data)
			watch.dataCache.Delete(cacheDataInfo)
		}
	}

	blog.Info("sync %d deployments from zk, delete %d cache deployments", numZk, numDel)

	return nil
}

//AddEvent call when data added
func (watch *DeploymentWatch) AddEvent(obj interface{}) {
	deploymentData, ok := obj.(*schedulertypes.Deployment)
	if !ok {
		blog.Error("can not convert object to Deployment in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for Deployment %s.%s", deploymentData.ObjectMeta.NameSpace, deploymentData.ObjectMeta.Name)

	data := &types.BcsSyncData{
		DataType: watch.GetDeploymentChannel(deploymentData),
		Action:   "Add",
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionAdd, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionAdd, cluster.SyncSuccess)
	}
}

//DeleteEvent when delete
func (watch *DeploymentWatch) DeleteEvent(obj interface{}) {
	deploymentData, ok := obj.(*schedulertypes.Deployment)
	if !ok {
		blog.Error("can not convert object to Deployment in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for Deployment %s.%s", deploymentData.ObjectMeta.NameSpace, deploymentData.ObjectMeta.Name)
	//report to cluster
	data := &types.BcsSyncData{
		DataType: watch.GetDeploymentChannel(deploymentData),
		Action:   "Delete",
		Item:     obj,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionDelete, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionDelete, cluster.SyncSuccess)
	}
}

//UpdateEvent when update
func (watch *DeploymentWatch) UpdateEvent(old, cur interface{}) {
	deploymentData, ok := cur.(*schedulertypes.Deployment)
	if !ok {
		blog.Error("can not convert object to Deployment in UpdateEvent, object %v", cur)
		return
	}

	blog.V(3).Infof("EVENT:: Update Event for BcsSecret %s.%s", deploymentData.ObjectMeta.NameSpace, deploymentData.ObjectMeta.Name)

	//report to cluster
	data := &types.BcsSyncData{
		DataType: watch.GetDeploymentChannel(deploymentData),
		Action:   "Update",
		Item:     cur,
	}
	if err := watch.report.ReportData(data); err != nil {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionUpdate, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(watch.report.GetClusterID(), cluster.DataTypeDeploy, types.ActionUpdate, cluster.SyncSuccess)
	}
}

// GetDeploymentChannel get channel by random algorithm
func (watch *DeploymentWatch) GetDeploymentChannel(deployment *schedulertypes.Deployment) string {
	index := util.GetHashId(deployment.ObjectMeta.Name, DeploymentThreadNum)

	return types.DeploymentChannelPrefix + strconv.Itoa(index)
}
