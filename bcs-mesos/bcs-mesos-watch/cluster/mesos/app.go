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
	"path"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/cache"
	schedulertypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/util"
)

//NSControlInfo store all app info under one namespace
type NSControlInfo struct {
	path   string             //parent zk node, namespace absolute path
	cxt    context.Context    //context for creating sub context
	cancel context.CancelFunc //for cancel sub goroutine
}

func reportAppMetrics(clusterID, action, status string) {
	util.ReportSyncTotal(clusterID, cluster.DataTypeApp, action, status)
}

//NewAppWatch return a new application watch
func NewAppWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter) *AppWatch {
	keyFunc := func(data interface{}) (string, error) {
		meta, ok := data.(*schedulertypes.Application)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return meta.RunAs + "." + meta.ID, nil
	}
	nsKeyFunc := func(data interface{}) (string, error) {
		ns, ok := data.(*NSControlInfo)
		if !ok {
			return "", fmt.Errorf("NSControlInfo type Assert failed")
		}
		return ns.path, nil
	}

	return &AppWatch{
		report:    reporter,
		cancelCxt: cxt,
		client:    client,
		dataCache: cache.NewCache(keyFunc),
		nsCache:   cache.NewCache(nsKeyFunc),
	}
}

//AppWatch for app data in zookeeper, app wath is base on namespace.
//AppWatch will record all namespace path,
type AppWatch struct {
	eventLock sync.Mutex       //lock for event
	report    cluster.Reporter //reporter
	cancelCxt context.Context  //context for cancel
	client    ZkClient         //client for zookeeper
	dataCache cache.Store      //cache for all app data
	nsCache   cache.Store      //all namespace path / namespace goroutine control info
}

//run AppWatch running with namespace node, watch all app node under it
func (app *AppWatch) addWatch(appPath string) error {
	app.eventLock.Lock()
	defer app.eventLock.Unlock()
	_, ok, _ := app.nsCache.GetByKey(appPath)
	if ok {
		blog.V(3).Infof("appwatch add wath path %s, but already in cache", appPath)
		return fmt.Errorf("Path %s is Under watch", appPath)
	}

	//ready to watch path
	nsCxt, nsCancel := context.WithCancel(app.cancelCxt)
	blog.Info("WATCH:: app pathwatch(%s) begin", appPath)
	ns := &NSControlInfo{
		path:   appPath,
		cxt:    nsCxt,
		cancel: nsCancel,
	}
	app.nsCache.Add(ns)
	go app.pathWatch(nsCxt, appPath)
	return nil
}

//cleanWatch clean wath by path
func (app *AppWatch) cleanWatch(appPath string) error {
	blog.V(3).Infof("appwatch clean watch path %s", appPath)
	app.eventLock.Lock()
	defer app.eventLock.Unlock()
	data, ok, _ := app.nsCache.GetByKey(appPath)
	if !ok {
		blog.V(3).Infof("appwatch clean wath path %s, but not in cache", appPath)
		return fmt.Errorf("Path %s is not Under watch", appPath)
	}
	app.nsCache.Delete(data)
	control := data.(*NSControlInfo)
	blog.Info("WATCH:: app pathwatch(%s) clean", appPath)
	control.cancel()
	return nil
}

//pathWatch watch app list under namespace path
func (app *AppWatch) pathWatch(cxt context.Context, path string) {

	// Get all children node & setting watch
	// if fail or error, clean the path, patchWatch will be retried by ProcessAppPathes later
	children, state, eventChan, wErr := app.client.ChildrenW(path)
	if wErr != nil {
		blog.Warnf("WATCH::  app pathwatch(%s) exit : %s", path, wErr.Error())
		app.cleanWatch(path)
		return
	} else if state == nil {
		blog.Errorf("WATCH::  app pathwatch(%s) exit for state nil", path)
		app.cleanWatch(path)
		return
	}

	blog.V(3).Infof("watch app path(%s), handle applist under it", path)
	app.handleAppList(cxt, path, children)

	tick := time.NewTicker(240 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("WATCH: tick app pathwatch(%s) alive", path)

		case <-cxt.Done():
			blog.Info("WATCH:: app pathwatch(%s) finish", path)
			app.cleanWatch(path)
			return

		case event := <-eventChan:
			blog.V(3).Infof("***********TRIGER:: app path %s zk event(%s)", path, event.Type)
			if event.Type == zk.EventSession {
				blog.Infof("WATCH:: app pathwatch(%s) rcv event zk.EventSession", path)
				continue
			}
			blog.Infof("WATCH:: app pathwatch(%s) redo for event(%s)", path, event.Type)
			go app.pathWatch(cxt, path)
			return
		}
	}
}

//updateList check app node list in zookeeper, force update if we read data from zookeeper
//add goroutine for new app node and listening node DEELTE & UPDATE events
func (app *AppWatch) handleAppList(cxt context.Context, base string, appList []string) error {

	for _, appNode := range appList {
		apppath := base + "/" + appNode
		byteData, _, err := app.client.GetEx(apppath)
		if err != nil {
			blog.Error("Get node %s data err: %s", apppath, err.Error())
			continue
		}
		appData := new(schedulertypes.Application)
		if jsonErr := json.Unmarshal(byteData, appData); jsonErr != nil {
			blog.Error("Parse %s json data(%s) failed: %s", apppath, string(byteData), jsonErr.Error())
			continue
		}
		oldData, exist, _ := app.dataCache.Get(appData)
		if exist {
			app.dataCache.Update(appData)
			app.UpdateEvent(oldData, appData, false)
		} else {
			app.dataCache.Add(appData)
			app.AddEvent(appData)
			appCxt, _ := context.WithCancel(cxt)
			blog.Info("WATCH:: app nodewatch(%s) begin", apppath)
			go app.appNodeWatch(appCxt, apppath, appData.RunAs)
		}
	}
	return nil
}

func (app *AppWatch) appNodeWatch(cxt context.Context, apppath string, ns string) error {

	blog.V(3).Infof("appwatch watch node(%s)", apppath)

	_, state, eventChan, err := app.client.GetW(apppath)
	if err != nil {
		blog.Warnf("WATCH:: app nodewatch(%s) exit: %s", apppath, err.Error())
		ID := ns + "." + path.Base(apppath)
		if deleteItem, exist, _ := app.dataCache.GetByKey(ID); exist {
			app.dataCache.Delete(deleteItem)
			if err == zk.ErrNoNode {
				app.DeleteEvent(deleteItem)
			}
		}
		return err
	} else if state == nil {
		blog.Warnf("WATCH:: app nodewatch(%s) exit for state nil", apppath)
		ID := ns + "." + path.Base(apppath)
		if deleteItem, exist, _ := app.dataCache.GetByKey(ID); exist {
			app.dataCache.Delete(deleteItem)
		}
		return fmt.Errorf("zk state empty")
	}

	ID := ns + "." + path.Base(apppath)
	blog.V(3).Infof("appwatch wath app ID(%s)", ID)

	tick := time.NewTicker(240 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Infof("WATCH: tick app nodewatch(%s) alive", apppath)
			byteData, _, err := app.client.GetEx(apppath)
			if err != nil {
				blog.Warn("WATCH: tick app nodewatch(%s) err: %s", apppath, err.Error())
				go app.appNodeWatch(cxt, apppath, ns)
				return err
			}
			appData := new(schedulertypes.Application)
			if jsonErr := json.Unmarshal(byteData, appData); jsonErr != nil {
				blog.Error("WATCH: tick app nodewatch(%s) parse json data Err: %s", apppath, jsonErr.Error())
				go app.appNodeWatch(cxt, apppath, ns)
				return jsonErr
			}

			oldData, exist, _ := app.dataCache.Get(appData)
			if exist {
				app.dataCache.Update(appData)
				app.UpdateEvent(oldData, appData, true)
			} else {
				blog.Warn("WATCH: tick app nodewatch(%s), data not in cache", apppath)
				app.dataCache.Add(appData)
				app.AddEvent(appData)
			}

		case <-cxt.Done():
			blog.Infof("WATCH:: app nodewatch(%s) finish", apppath)
			if deleteItem, exist, _ := app.dataCache.GetByKey(ID); exist {
				app.dataCache.Delete(deleteItem)
			}
			return nil

		case event := <-eventChan:
			blog.V(3).Infof("***********TRIGER:: app node %s zk event(%s)", apppath, event.Type)
			if event.Type == zk.EventSession {
				blog.Infof("WATCH:: app nodewatch(%s) rcv event zk.EventSession", apppath)
				continue
			}

			if event.Type == zk.EventNodeDataChanged {
				byteData, _, err := app.client.GetEx(apppath)
				if err != nil {
					blog.Warn("WATCH:: app nodewatch(%s) err: %s", apppath, err.Error())
					go app.appNodeWatch(cxt, apppath, ns)
					return err
				}

				appData := new(schedulertypes.Application)
				if jsonErr := json.Unmarshal(byteData, appData); jsonErr != nil {
					blog.Error("WATCH:: app nodewatch(%s) parse json data Err: %s", apppath, jsonErr.Error())
					go app.appNodeWatch(cxt, apppath, ns)
					return err
				}

				oldData, exist, _ := app.dataCache.Get(appData)
				if exist {
					app.dataCache.Update(appData)
					app.UpdateEvent(oldData, appData, false)
				} else {
					blog.Warn("WATCH: app nodewatch(%s) recv event, data not in cache", apppath)
					app.dataCache.Add(appData)
					app.AddEvent(appData)
				}
			} else {
				blog.Infof("WATCH:: app nodewatch(%s) rcv event(%s)", apppath, event.Type)
			}

			blog.V(3).Infof("WATCH:: app nodewatch(%s) redo for event(%s)", apppath, event.Type)
			go app.appNodeWatch(cxt, apppath, ns)
			return nil
		}
	}
}

//stop ask appwatch stop, clean all data
func (app *AppWatch) stop() {

	blog.Info("appwatch stop...")

	keys := app.nsCache.ListKeys()
	for _, key := range keys {
		app.cleanWatch(key)
	}
	app.dataCache.Clear()
}

//IsExist check data exist in local dataCache
func (app *AppWatch) IsExist(data interface{}) bool {
	appData, ok := data.(*schedulertypes.Application)
	if !ok {
		return false
	}
	_, exist, _ := app.dataCache.Get(appData)
	if exist {
		return true
	}
	return false
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

	if !force && reflect.DeepEqual(old, cur) {
		blog.V(3).Infof("App %s data do not changed", appData.ID)
		return
	}
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
