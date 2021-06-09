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

//TaskControlInfo store all app info under one namespace
type TaskControlInfo struct {
	path   string             //parent zk node, namespace absolute path
	cxt    context.Context    //context for creating sub context
	cancel context.CancelFunc //for cancel sub goroutine
}

//NewTaskGroupWatch create default taskgroup watch
func NewTaskGroupWatch(cxt context.Context, client ZkClient, reporter cluster.Reporter) *TaskGroupWatch {
	taskKeyFunc := func(data interface{}) (string, error) {
		meta, ok := data.(*schedulertypes.TaskGroup)
		if !ok {
			return "", fmt.Errorf("SchedulerMeta type Assert failed")
		}
		return meta.ID, nil
	}
	controlKeyFunc := func(data interface{}) (string, error) {
		meta, ok := data.(*TaskControlInfo)
		if !ok {
			return "", fmt.Errorf("TaskControlInfo type Assert failed")
		}
		return meta.path, nil
	}
	return &TaskGroupWatch{
		cancelCxt:    cxt,
		report:       reporter,
		client:       client,
		tasksCache:   cache.NewCache(taskKeyFunc),
		controlCache: cache.NewCache(controlKeyFunc),
	}
}

//TaskGroupWatch watch for taskGroup
type TaskGroupWatch struct {
	eventLock    sync.Mutex       //lock for event
	cancelCxt    context.Context  //context for cancel
	report       cluster.Reporter //reporter to cluster
	client       ZkClient         //zookeeper client
	tasksCache   cache.Store      //taskgroup info cache
	controlCache cache.Store      //app cache info
}

//run AppWatch running with namespace node, watch all app node under it
func (task *TaskGroupWatch) addWatch(appPath string) error {
	task.eventLock.Lock()
	defer task.eventLock.Unlock()
	_, ok, _ := task.controlCache.GetByKey(appPath)
	if ok {
		return fmt.Errorf("Path %s is Under watch", appPath)
	}
	//ready to watch path
	tskCxt, tskCancel := context.WithCancel(task.cancelCxt)
	blog.Info("WATCH:: taskgroup pathwatch(%s) begin", appPath)
	control := &TaskControlInfo{
		path:   appPath,
		cxt:    tskCxt,
		cancel: tskCancel,
	}
	task.controlCache.Add(control)
	go task.pathWatch(tskCxt, appPath)
	return nil
}

//cleanWatch clean wath by path
func (task *TaskGroupWatch) cleanWatch(appPath string) error {
	task.eventLock.Lock()
	defer task.eventLock.Unlock()
	data, ok, _ := task.controlCache.GetByKey(appPath)
	if !ok {
		return fmt.Errorf("Path %s is not Under watch", appPath)
	}
	task.controlCache.Delete(data)
	control := data.(*TaskControlInfo)
	blog.Info("WATCH:: taskgroup pathwatch(%s) clean", appPath)
	control.cancel()
	return nil
}

//pathWatch watch taskgroup list under application path
func (task *TaskGroupWatch) pathWatch(cxt context.Context, path string) {

	//Get all children node & setting watch
	children, state, eventChan, wErr := task.client.ChildrenW(path)
	if wErr != nil {
		blog.Infof("WATCH:: taskgroup pathwatch %s exit : %s", path, wErr.Error())
		task.cleanWatch(path)
		return
	} else if state == nil {
		blog.Error("WATCH:: taskgroup pathwatch %s exit for state nil", path)
		task.cleanWatch(path)
		return
	}
	//Get taskGroup detail data & store to local cache
	blog.V(3).Infof("watch taskgroup path(%s), handle grouplist under it", path)
	task.handleTaskGroupList(cxt, path, children)

	//watch children node event
	tick := time.NewTicker(240 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("WATCH: tick taskgroup pathwatch(%s) alive", path)

		case <-cxt.Done():
			blog.Info("WATCH:: taskgroup pathwatch(%s) finish", path)
			task.cleanWatch(path)
			return
		case event := <-eventChan:
			blog.V(3).Infof("***********TRIGER:: taskgroup path %s zk event(%s)", path, event.Type)
			if event.Type == zk.EventSession {
				blog.Info("WATCH:: taskgroup pathwatch(%s) rcv event zk.EventSession", path)
				continue
			}

			blog.Infof("WATCH:: taskgroup pathwatch(%s) redo for event(%s)", path, event.Type)
			go task.pathWatch(cxt, path)
			return
		}
	}
}

func (task *TaskGroupWatch) handleTaskGroupList(cxt context.Context, base string, tskGroups []string) error {

	for _, taskgroup := range tskGroups {
		taskpath := base + "/" + taskgroup
		blog.V(3).Infof("handle taskgroup node: %s", taskpath)
		byteData, _, err := task.client.GetEx(taskpath)
		if err != nil {
			blog.Error("Get TaskGroup %s data err: %s", taskpath, err.Error())
			continue
		}
		taskData := new(schedulertypes.TaskGroup)
		if jsonErr := json.Unmarshal(byteData, taskData); jsonErr != nil {
			blog.Error("Parse %s json data failed: %s", taskpath, jsonErr.Error())
			continue
		}
		oldData, exist, _ := task.tasksCache.Get(taskData)
		if exist {
			task.tasksCache.Update(taskData)
			task.UpdateEvent(oldData, taskData, false)
		} else {
			task.tasksCache.Add(taskData)
			task.AddEvent(taskData)
			tskGroupCxt, _ := context.WithCancel(cxt)
			blog.Info("WATCH:: taskgroup nodewatch(%s) begin", taskpath)
			go task.taskGroupNodeWatch(tskGroupCxt, taskpath)
		}
	}
	return nil
}

//taskGroupNodeWatch watch zookeeper taskgroup data node. focus on UPDATE & DELETE event
func (task *TaskGroupWatch) taskGroupNodeWatch(cxt context.Context, taskpath string) error {

	_, state, eventChan, err := task.client.GetW(taskpath)
	if err != nil {
		blog.Infof("WATCH:: taskgroup nodewatch(%s) exit for err: %s", taskpath, err.Error())
		ID := path.Base(taskpath)
		if deleteItem, exist, _ := task.tasksCache.GetByKey(ID); exist {
			task.tasksCache.Delete(deleteItem)
			if err == zk.ErrNoNode {
				task.DeleteEvent(deleteItem)
			}
		}
		return err
	} else if state == nil {
		blog.Warnf("WATCH:: taskgroup nodewatch(%s) exit for state nil", taskpath)
		ID := path.Base(taskpath)
		if deleteItem, exist, _ := task.tasksCache.GetByKey(ID); exist {
			task.tasksCache.Delete(deleteItem)
		}
		return fmt.Errorf("zk state empty")
	}

	ID := path.Base(taskpath)
	tick := time.NewTicker(240 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("WATCH: tick taskgroup nodewatch(%s) alive", taskpath)
			byteData, _, err := task.client.GetEx(taskpath)
			if err != nil {
				blog.Warn("WATCH: tick taskgroup nodewatch(%s) err: %s", taskpath, err.Error())
				go task.taskGroupNodeWatch(cxt, taskpath)
				return err
			}

			taskData := new(schedulertypes.TaskGroup)
			if jsonErr := json.Unmarshal(byteData, taskData); jsonErr != nil {
				blog.Error("WATCH: taskgroup nodewatch(%s) parse json data Err: %s", taskpath, jsonErr.Error())
				go task.taskGroupNodeWatch(cxt, taskpath)
				return jsonErr
			}

			oldData, exist, _ := task.tasksCache.Get(taskData)
			if exist {
				task.tasksCache.Update(taskData)
				task.UpdateEvent(oldData, taskData, true)
			} else {
				blog.Warn("WATCH: tick taskgroup nodewatch(%s), data not in cache", taskpath)
				task.tasksCache.Add(taskData)
				task.AddEvent(taskData)
			}

		case <-cxt.Done():
			blog.Info("WATCH:: taskgroup nodewatch(%s) finish", taskpath)
			if deleteItem, exist, _ := task.tasksCache.GetByKey(ID); exist {
				task.tasksCache.Delete(deleteItem)
			}
			return nil

		case event := <-eventChan:
			blog.V(3).Infof("***********TRIGER:: taskgroup node %s zk event(%s)", taskpath, event.Type)
			if event.Type == zk.EventSession {
				blog.Info("WATCH:: taskgroup nodewatch(%s) rcv event zk.EventSession", taskpath)
				continue
			}

			if event.Type == zk.EventNodeDataChanged {
				byteData, _, err := task.client.GetEx(taskpath)
				if err != nil {
					blog.Warn("WATCH: taskgroup nodewatch(%s) get data err when NodeDataChanged: %s", taskpath, err.Error())
					go task.taskGroupNodeWatch(cxt, taskpath)
					return err
				}

				taskData := new(schedulertypes.TaskGroup)
				if jsonErr := json.Unmarshal(byteData, taskData); jsonErr != nil {
					blog.Error("WATCH: taskgroup nodewatch(%s) parse json data Err: %s", taskpath, jsonErr.Error())
					go task.taskGroupNodeWatch(cxt, taskpath)
					return jsonErr
				}

				oldData, exist, _ := task.tasksCache.Get(taskData)
				if exist {
					task.tasksCache.Update(taskData)
					task.UpdateEvent(oldData, taskData, false)
				} else {
					blog.Warn("WATCH: taskgroup nodewatch(%s) recv event, data not in cache", taskpath)
					task.tasksCache.Add(taskData)
					task.AddEvent(taskData)
				}
			} else {
				blog.Infof("WATCH:: taskgroup nodewatch(%s) rcv event(%s)", taskpath, event.Type)
			}

			blog.V(3).Infof("WATCH:: taskgroup nodewatch(%s) watch redo for event(%s)", taskpath, event.Type)
			go task.taskGroupNodeWatch(cxt, taskpath)
			return nil
		}
	}
}

//stop ask appwatch stop, clean all data
func (task *TaskGroupWatch) stop() {
	keys := task.controlCache.ListKeys()
	for _, key := range keys {
		task.cleanWatch(key)
	}
	task.tasksCache.Clear()
}

//IsExist check data exist in local dataCache
func (task *TaskGroupWatch) IsExist(data interface{}) bool {
	taskData, ok := data.(*schedulertypes.TaskGroup)
	if !ok {
		return false
	}
	_, exist, _ := task.tasksCache.Get(taskData)
	if exist {
		return true
	}
	return false
}

//AddEvent call when data added
func (task *TaskGroupWatch) AddEvent(obj interface{}) {
	taskData, ok := obj.(*schedulertypes.TaskGroup)
	if !ok {
		blog.Error("can not convert object to taskgroup in AddEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Add Event for Taskgroup %s", taskData.ID)

	data := &types.BcsSyncData{
		//DataType: "TaskGroup",
		DataType: task.GetTaskGroupChannelV2(taskData),
		Action:   types.ActionAdd,
		Item:     obj,
	}
	if err := task.report.ReportData(data); err != nil {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionAdd, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionAdd, cluster.SyncSuccess)
	}
}

//DeleteEvent when delete
func (task *TaskGroupWatch) DeleteEvent(obj interface{}) {
	taskData, ok := obj.(*schedulertypes.TaskGroup)
	if !ok {
		blog.Error("can not convert object to TaskGroup in DeleteEvent, object %v", obj)
		return
	}
	blog.Info("EVENT:: Delete Event for TaskGroup %s", taskData.ID)

	//report to cluster
	data := &types.BcsSyncData{
		//DataType: "TaskGroup",
		DataType: task.GetTaskGroupChannelV2(taskData),
		Action:   types.ActionDelete,
		Item:     obj,
	}
	if err := task.report.ReportData(data); err != nil {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionDelete, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionDelete, cluster.SyncSuccess)
	}
}

//UpdateEvent when update
func (task *TaskGroupWatch) UpdateEvent(old, cur interface{}, force bool) {
	taskData, ok := cur.(*schedulertypes.TaskGroup)
	if !ok {
		blog.Error("can not convert object to TaskGroup in UpdateEvent, object %v", cur)
		return
	}
	if !force && reflect.DeepEqual(old, cur) {
		blog.V(3).Infof("TaskGroup %s data do not changed", taskData.ID)
		return
	}
	blog.V(3).Infof("EVENT:: Update Event for TaskGroup %s", taskData.ID)
	//report to cluster
	data := &types.BcsSyncData{
		//DataType: "TaskGroup",
		DataType: task.GetTaskGroupChannelV2(taskData),
		Action:   types.ActionUpdate,
		Item:     cur,
	}
	if err := task.report.ReportData(data); err != nil {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionUpdate, cluster.SyncFailure)
	} else {
		util.ReportSyncTotal(task.report.GetClusterID(), cluster.DataTypeTaskGroup, types.ActionUpdate, cluster.SyncSuccess)
	}
}

//GetTaskGroupChannel get taskgroup dispatch channel
func (task *TaskGroupWatch) GetTaskGroupChannel(taskGroup *schedulertypes.TaskGroup) string {

	return "TaskGroup_" + strconv.Itoa(int(taskGroup.InstanceID%100))

}

//GetTaskGroupChannelV2 get taskgroup dispatch channel
func (task *TaskGroupWatch) GetTaskGroupChannelV2(taskGroup *schedulertypes.TaskGroup) string {

	index := util.GetHashId(taskGroup.ID, TaskgroupThreadNum)

	return types.TaskgroupChannelPrefix + strconv.Itoa(index)

}
