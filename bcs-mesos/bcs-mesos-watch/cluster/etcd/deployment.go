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
	"strconv"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	schedulertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions/bkbcs/v2"

	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/labels"
)

//DeploymentInfo wrapper for BCS Deployment
type DeploymentInfo struct {
	data       *schedulertypes.Deployment
	syncTime   int64
	reportTime int64
}

//NewDeploymentWatch create deployment watch
func NewDeploymentWatch(cxt context.Context, informer bkbcsv2.DeploymentInformer, reporter cluster.Reporter) *DeploymentWatch {

	keyFunc := func(data interface{}) (string, error) {
		dataType, ok := data.(*DeploymentInfo)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return dataType.data.ObjectMeta.NameSpace + "." + dataType.data.ObjectMeta.Name, nil
	}

	return &DeploymentWatch{
		report:    reporter,
		cancelCxt: cxt,
		informer:  informer,
		dataCache: cache.NewCache(keyFunc),
		//nsCache:   cache.NewCache(nsKeyFunc),
	}
}

//DeploymentWatch watch all deployment data and store to local cache
type DeploymentWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	dataCache cache.Store      //cache for all app data
	//nsCache   cache.Store     //all namespace path / namespace goroutine control info
	watchPath string
	informer  bkbcsv2.DeploymentInformer
}

//Work to add path and node watch
func (watch *DeploymentWatch) Work() {
	blog.Infof("DeploymentWatch start work")

	watch.ProcessAllDeployments()
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-watch.cancelCxt.Done():
			blog.Infof("DeploymentWatch asked to exit")
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
	blog.V(3).Infof("sync all deployments, currTime(%d)", currTime)

	v2Deps, err := watch.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("list configmaps error %s", err.Error())
		return err
	}

	var numNode, numDel int
	for _, dep := range v2Deps {
		numNode++
		deployment := &dep.Spec.Deployment
		key := deployment.ObjectMeta.NameSpace + "." + deployment.ObjectMeta.Name
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
			if reflect.DeepEqual(cacheDataInfo.data, deployment) {
				if cacheDataInfo.reportTime > currTime {
					cacheDataInfo.reportTime = currTime
				}
				if currTime-cacheDataInfo.reportTime > 360 {
					blog.Info("deployment %s data not changed, but long time not report, do report", key)
					watch.UpdateEvent(cacheDataInfo.data, deployment)
					cacheDataInfo.reportTime = currTime
				}
			} else {
				blog.Info("deployment %s data changed, do report", key)
				watch.UpdateEvent(cacheDataInfo.data, deployment)
				cacheDataInfo.reportTime = currTime
			}

			cacheDataInfo.syncTime = currTime
			cacheDataInfo.data = deployment
		} else {
			blog.Info("deployment %s is not in cache, add, time(%d)", key, currTime)
			watch.AddEvent(deployment)
			dataInfo := new(DeploymentInfo)
			dataInfo.data = deployment
			dataInfo.syncTime = currTime
			dataInfo.reportTime = currTime
			watch.dataCache.Add(dataInfo)
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

	blog.Info("sync %d deployments from etcd, delete %d cache deployments", numNode, numDel)

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

// GetDeploymentChannel return random channel by hash algorithm
func (watch *DeploymentWatch) GetDeploymentChannel(deployment *schedulertypes.Deployment) string {
	index := util.GetHashId(deployment.ObjectMeta.Name, DeploymentThreadNum)

	return types.DeploymentChannelPrefix + strconv.Itoa(index)
}
