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
	"os"
	"strconv"
	"sync"

	"golang.org/x/net/context"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	schedulertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	bkbcsv2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/client/informers/externalversions/bkbcs/v2"
)

func reportAppMetrics(clusterID, action, status string) {
	util.ReportSyncTotal(clusterID, cluster.DataTypeApp, action, status)
}

//NSControlInfo store all app info under one namespace
type NSControlInfo struct {
	path   string             //parent zk node, namespace absolute path
	cxt    context.Context    //context for creating sub context
	cancel context.CancelFunc //for cancel sub goroutine
}

//NewAppWatch return a new application watch
func NewAppWatch(cxt context.Context, informer bkbcsv2.ApplicationInformer, reporter cluster.Reporter) *AppWatch {
	return &AppWatch{
		report:    reporter,
		cancelCxt: cxt,
		informer:  informer,
	}
}

//AppWatch for app data in zookeeper, app wath is base on namespace.
//AppWatch will record all namespace path,
type AppWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	informer  bkbcsv2.ApplicationInformer
}

// Work for sync application by register informer
func (app *AppWatch) Work() {
	blog.Infof("AppWatch start work")
	app.syncAllApplications()
	blog.Infof("AppWatch syncAllApplications done")

	app.informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    app.addNodeToCache,
			UpdateFunc: app.updateNodeInCache,
			DeleteFunc: app.deleteNodeFromCache,
		},
	)
}

func (app *AppWatch) addNodeToCache(obj interface{}) {
	application, ok := obj.(*v2.Application)
	if !ok {
		blog.Errorf("cannot convert to *v2.Application: %v", obj)
		return
	}
	app.AddEvent(&application.Spec.Application)
}

func (app *AppWatch) updateNodeInCache(oldObj, newObj interface{}) {
	oldApp, ok := oldObj.(*v2.Application)
	if !ok {
		blog.Errorf("cannot convert oldObj to *v2.Application: %v", oldObj)
		return
	}
	newApp, ok := newObj.(*v2.Application)
	if !ok {
		blog.Errorf("cannot convert newObj to *v2.Application: %v", newObj)
		return
	}
	app.UpdateEvent(&oldApp.Spec.Application, &newApp.Spec.Application, false)
}

func (app *AppWatch) deleteNodeFromCache(obj interface{}) {
	application, ok := obj.(*v2.Application)
	if !ok {
		blog.Errorf("cannot convert to *v2.Application: %v", obj)
		return
	}
	app.DeleteEvent(&application.Spec.Application)
}

func (app *AppWatch) syncAllApplications() {
	v2Apps, err := app.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("AppWatch syncAllApplications error %s", err.Error())
		os.Exit(1)
	}

	for _, obj := range v2Apps {
		app.AddEvent(&obj.Spec.Application)
	}
}

//AddEvent call when data added
func (app *AppWatch) AddEvent(obj interface{}) {
	appData, ok := obj.(*schedulertypes.Application)
	if !ok {
		blog.Error("can not convert object to Application in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for Application %s/%s", appData.RunAs, appData.ID)

	data := &types.BcsSyncData{
		DataType: app.GetApplicationChannel(appData),
		Action:   types.ActionAdd,
		Item:     obj,
	}
	if err := app.report.ReportData(data); err != nil {
		reportAppMetrics(app.report.GetClusterID(), types.ActionAdd, cluster.SyncFailure)
	} else {
		reportAppMetrics(app.report.GetClusterID(), types.ActionAdd, cluster.SyncSuccess)
	}
}

//DeleteEvent when delete
func (app *AppWatch) DeleteEvent(obj interface{}) {
	appData, ok := obj.(*schedulertypes.Application)
	if !ok {
		blog.Error("can not convert object to Application in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for Application %s/%s", appData.RunAs, appData.ID)

	//report to cluster
	data := &types.BcsSyncData{
		DataType: app.GetApplicationChannel(appData),
		Action:   types.ActionDelete,
		Item:     obj,
	}
	if err := app.report.ReportData(data); err != nil {
		reportAppMetrics(app.report.GetClusterID(), types.ActionDelete, cluster.SyncFailure)
	} else {
		reportAppMetrics(app.report.GetClusterID(), types.ActionDelete, cluster.SyncSuccess)
	}
}

//UpdateEvent when update
func (app *AppWatch) UpdateEvent(old, cur interface{}, force bool) {
	appData, ok := cur.(*schedulertypes.Application)
	if !ok {
		blog.Error("can not convert object to Application in UpdateEvent, object %v", cur)
		return
	}

	/*if !force && reflect.DeepEqual(old, cur) {
		blog.V(3).Infof("App %s data do not changed", appData.ID)
		return
	}*/
	blog.V(3).Infof("EVENT:: Update Event for Application %s/%s", appData.RunAs, appData.ID)

	//report to cluster
	data := &types.BcsSyncData{
		DataType: app.GetApplicationChannel(appData),
		Action:   types.ActionUpdate,
		Item:     cur,
	}
	if err := app.report.ReportData(data); err != nil {
		reportAppMetrics(app.report.GetClusterID(), types.ActionUpdate, cluster.SyncFailure)
	} else {
		reportAppMetrics(app.report.GetClusterID(), types.ActionUpdate, cluster.SyncSuccess)
	}
}

//GetApplicationChannel get distribution channel for Application
func (app *AppWatch) GetApplicationChannel(application *schedulertypes.Application) string {
	index := util.GetHashId(application.ID, ApplicationThreadNum)

	return types.ApplicationChannelPrefix + strconv.Itoa(index)
}
