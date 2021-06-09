/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by tasklicable law or agreed to in writing, software distributed under
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

//TaskControlInfo store all task info under one namespace
type TaskControlInfo struct {
	cxt    context.Context    //context for creating sub context
	cancel context.CancelFunc //for cancel sub goroutine
}

//NewTaskGroupWatch create default taskgroup watch
func NewTaskGroupWatch(cxt context.Context, informer bkbcsv2.TaskGroupInformer, reporter cluster.Reporter) *TaskGroupWatch {
	return &TaskGroupWatch{
		cancelCxt: cxt,
		report:    reporter,
		informer:  informer,
	}
}

//TaskGroupWatch watch for taskGroup
type TaskGroupWatch struct {
	eventLock sync.Mutex       //lock for event
	cancelCxt context.Context  //context for cancel
	report    cluster.Reporter //reporter to cluster
	informer  bkbcsv2.TaskGroupInformer
}

// Work main work init for taskgroup
func (task *TaskGroupWatch) Work() {
	blog.Infof("TaskGroupWatch start work")
	task.syncAlltaskgroups()
	blog.Infof("TaskGroupWatch syncAlltaskgroups done")

	task.informer.Informer().AddEventHandler(
		cache.ResourceEventHandlerFuncs{
			AddFunc:    task.addNodeToCache,
			UpdateFunc: task.updateNodeInCache,
			DeleteFunc: task.deleteNodeFromCache,
		},
	)
}

func (task *TaskGroupWatch) addNodeToCache(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		blog.Errorf("cannot convert to *v2.taskgroup: %v", obj)
		return
	}

	task.AddEvent(&taskgroup.Spec.TaskGroup)
}

func (task *TaskGroupWatch) updateNodeInCache(oldObj, newObj interface{}) {
	oldtask, ok := oldObj.(*v2.TaskGroup)
	if !ok {
		blog.Errorf("cannot convert oldObj to *v2.taskgroup: %v", oldObj)
		return
	}
	newtask, ok := newObj.(*v2.TaskGroup)
	if !ok {
		blog.Errorf("cannot convert newObj to *v2.taskgroup: %v", newObj)
		return
	}

	task.UpdateEvent(&oldtask.Spec.TaskGroup, &newtask.Spec.TaskGroup, false)
}

func (task *TaskGroupWatch) deleteNodeFromCache(obj interface{}) {
	taskgroup, ok := obj.(*v2.TaskGroup)
	if !ok {
		blog.Errorf("cannot convert to *v2.taskgroup: %v", obj)
		return
	}

	task.DeleteEvent(&taskgroup.Spec.TaskGroup)
}

func (task *TaskGroupWatch) syncAlltaskgroups() {
	v2tasks, err := task.informer.Lister().List(labels.Everything())
	if err != nil {
		blog.Errorf("TaskGroupWatch syncAlltaskgroups error %s", err.Error())
		os.Exit(1)
	}

	for _, obj := range v2tasks {
		task.AddEvent(&obj.Spec.TaskGroup)
	}
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
	/*if !force && reflect.DeepEqual(old, cur) {
		blog.V(3).Infof("TaskGroup %s data do not changed", taskData.ID)
		return
	}*/
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
