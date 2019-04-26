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
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	comtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/pkg/cache"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/types"
	"bk-bcs/bcs-mesos/bcs-container-executor/container"
	schtypes "bk-bcs/bcs-mesos/bcs-scheduler/src/types"

	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"
)

const (
	MaxDataLength = 1024
)

type taskgroupWatcher struct {
	zk ZkClient

	cxt    context.Context
	cancel context.CancelFunc

	dataQueue chan *types.HealthSyncData

	caches cache.Store
}

func NewTaskgroupWatcher(rootCxt context.Context, zk ZkClient) Watcher {

	cxt, cancel := context.WithCancel(rootCxt)

	keyFunc := func(data interface{}) (string, error) {
		meta, ok := data.(types.WatchPathController)
		if !ok {
			return "", fmt.Errorf("WatchPathController type Assert failed")
		}

		return meta.Path, nil
	}

	watcher := &taskgroupWatcher{
		dataQueue: make(chan *types.HealthSyncData, MaxDataLength),
		caches:    cache.NewCache(keyFunc),
		cxt:       cxt,
		cancel:    cancel,
		zk:        zk,
	}

	return watcher
}

func (t *taskgroupWatcher) Run() {
	blog.Info("taskgroupWatcher running...")

	// watch zk application, like /blueking/application, for discover the change of namespace
	t.watchZkApplication()
}

func (t *taskgroupWatcher) Stop() {
	t.cancel()
}

func (t *taskgroupWatcher) syncHealthCheck(action types.SyncDataAction, check types.HealthCheck) {
	data := &types.HealthSyncData{
		Action:      action,
		HealthCheck: &check,
	}

	blog.Info("syncHealthCheck action %s checkId %s", string(action), check.ID)
	t.dataQueue <- data
}

func (t *taskgroupWatcher) DataW() <-chan *types.HealthSyncData {
	return t.dataQueue
}

func (t *taskgroupWatcher) watchZkApplication() {
	cxt, cancel := context.WithCancel(t.cxt)

	wp := &types.WatchPathController{
		Path:   "/blueking/application",
		Type:   types.ZkPathTypeApplication,
		Cxt:    cxt,
		Cancel: cancel,
	}

	for {
		exist, err := t.zk.Exist(wp.Path)
		if err != nil {
			blog.Errorf("zk Exist path %s error %s", wp.Path, err.Error())
		}

		if exist {
			break
		}

		blog.Warnf("zk path %s not exists", wp.Path)
		time.Sleep(time.Second)
	}

	err := t.watchData(wp)
	if err != nil {
		blog.Errorf("watchZkApplication path %s error %s", wp.Path, err.Error())
	}
}

func (t *taskgroupWatcher) watchData(data *types.WatchPathController) error {

	_, ok, _ := t.caches.GetByKey(data.Path)
	if ok {
		return fmt.Errorf("zk path %s is under watching", data.Path)
	}

	err := t.addWatchData(data)
	if err != nil {
		return err
	}

	//zk GetW
	go t.getW(data)

	//zk ChildrenW
	go t.childrenW(data)

	return nil
}

func (t *taskgroupWatcher) getW(wp *types.WatchPathController) {
	blog.V(3).Infof("getW zk path %s", wp.Path)

	if wp.Path == "" {
		blog.Error("getW zk path can't be empty")
		return
	}

	_, state, eventChan, err := t.zk.GetW(wp.Path)

	if err != nil {
		blog.Error("GetW zk path %s error %s", wp.Path, err.Error())

		err = t.caches.Delete(*wp)
		if err != nil {
			blog.Errorf("cache delete data error %s", err.Error())
		}
		return
	} else if state == nil {
		blog.Error("GetW zk path %s, but state response nil", wp.Path)

		err = t.caches.Delete(*wp)
		if err != nil {
			blog.Errorf("cache delete data error %s", err.Error())
		}
		return
	}

	tick := time.NewTicker(60 * time.Second)

	var continueBool bool
	var goBool bool

ForLabel:
	for {

		select {
		case <-tick.C:
			exist, err1 := t.zk.Exist(wp.Path)
			if err1 != nil {
				blog.Errorf("zk Exist path %s error %s", wp.Path, err1.Error())
				continue
			}
			if !exist {
				blog.Warnf("zk path %s not exists", wp.Path)
				wp.Cancel()
				return
			}

		case <-wp.Cxt.Done():
			blog.Info("cxt: zk path %s GetW exit", wp.Path)
			return

		case event := <-eventChan:
			blog.V(3).Infof("getW path %s rev event %s", wp.Path, event.Type.String())

			continueBool, goBool, err = t.handleZkEvent(event, wp)
			if err != nil {
				blog.Errorf("GetW handleZkEvent zk path %s error %s", wp.Path, err.Error())
			}

			//if continueBool = true, then continue
			if continueBool {
				continue
			}

			// if continueBool = false, then break
			break ForLabel

		}

	}

	// if goBool = true,then go getW()
	if goBool {
		go t.getW(wp)
	}
}

func (t *taskgroupWatcher) handleZkEvent(event zk.Event, wp *types.WatchPathController) (bool, bool, error) {
	var (
		err          error
		continueBool bool
		goBool       bool
	)
	switch event.Type {
	case zk.EventNotWatching:
		blog.Warn("watch: zk path %s rcv event zk.EventNotWatching and exit", wp.Path)
		// stop watch
		wp.Cancel()

	case zk.EventSession:
		blog.V(3).Infof("watch: zk path %s rcv event zk.EventSession", wp.Path)
		continueBool = true

	case zk.EventNodeDeleted:
		blog.V(3).Infof("watch: zk path %s rcv event zk.EventNodeDeleted", wp.Path)
		// stop watch
		wp.Cancel()
		// stop watch path,and if type = taskgroup, then delete health check task
		err = t.handleNodeDeleteEvent(wp)

	case zk.EventNodeDataChanged:
		blog.V(3).Infof("watch: zk path %s rcv event zk.EventNodeDataChanged", wp.Path)
		goBool = true
		// only type = taskgroup, and update healthcheck
		err = t.handleDataChangedEvent(wp)

	case zk.EventNodeChildrenChanged:
		blog.V(3).Infof("watch: zk path %s rcv event zk.EventNodeChildrenChanged", wp.Path)
		goBool = true
		//when children changed, watch children path
		//err = t.handleNodeChildrenChanged(wp)

	case zk.EventNodeCreated:
		blog.V(3).Infof("watch: zk path %s rcv event zk.EventNodeCreated", wp.Path)
		goBool = true
	}

	return continueBool, goBool, err
}

func (t *taskgroupWatcher) handleNodeChildrenChanged(wp *types.WatchPathController) error {

	children, state, err := t.zk.GetChildrenEx(wp.Path)
	if err != nil {
		blog.Error("childrenW zk path %s error %s", wp.Path, err.Error())

		err1 := t.caches.Delete(*wp)
		if err1 != nil {
			blog.Errorf("cache delete data error %s", err1.Error())
		}
		return err

	} else if state == nil {
		blog.Error("childrenW zk path %s, but state response nil", wp.Path)

		err1 := t.caches.Delete(*wp)
		if err1 != nil {
			blog.Errorf("cache delete data error %s", err1.Error())
		}
		return err
	}

	// watch the children
	go t.handleChildren(wp, children)

	return nil
}

func (t *taskgroupWatcher) childrenW(wp *types.WatchPathController) {
	blog.V(3).Infof("childrenW zk path %s", wp.Path)

	// if type = "taskgroup", then don't need to watch children
	if wp.Type == types.ZkPathTypeTask {
		blog.Info("childrenW zk path %s type %s, and don't need ChildrenW", wp.Path, wp.Type)
		return
	}

	if wp.Path == "" {
		blog.Error("childrenW zk path can't be empty")
		return
	}

	children, state, eventChan, err := t.zk.ChildrenW(wp.Path)

	if err != nil {
		blog.Error("childrenW zk path %s error %s", wp.Path, err.Error())

		err1 := t.caches.Delete(*wp)
		if err1 != nil {
			blog.Errorf("cache delete data error %s", err1.Error())
		}

		return

	} else if state == nil {
		blog.Error("childrenW zk path %s, but state response nil", wp.Path)

		err1 := t.caches.Delete(*wp)
		if err1 != nil {
			blog.Errorf("cache delete data error %s", err1.Error())
		}
		return
	}

	// watch the children
	go t.handleChildren(wp, children)

	//blog.Info("start waiting for zk path %s childrenW event",wp.Path)

	tick := time.NewTicker(60 * time.Second)

	var continueBool bool
	var goBool bool

ForLabel:
	for {

		select {
		case <-tick.C:
			exist, err1 := t.zk.Exist(wp.Path)
			if err1 != nil {
				blog.Errorf("zk Exist path %s error %s", wp.Path, err1.Error())
				continue
			}
			if !exist {
				blog.Warnf("zk path %s not exists", wp.Path)
				wp.Cancel()
				return
			}

		case <-wp.Cxt.Done():
			blog.Info("cxt: zk path %s GetW exit", wp.Path)
			return

		case event := <-eventChan:
			blog.V(3).Infof("childrenW path %s rev event %s", wp.Path, event.Type.String())

			//if rev event EventNodeDataChanged, then ignore it
			if event.Type == zk.EventNodeDataChanged {
				goBool = true
				break ForLabel
			}

			continueBool, goBool, err = t.handleZkEvent(event, wp)
			if err != nil {
				blog.Errorf("childrenW handleZkEvent zk path %s error %s", wp.Path, err.Error())
			}

			//if continueBool = true, then continue
			if continueBool {
				continue
			}

			// if continueBool = false, then break
			break ForLabel

		}

	}

	// if goBool = true, then go childrenW()
	if goBool {
		go t.childrenW(wp)
	}

}

func (t *taskgroupWatcher) handleChildren(wp *types.WatchPathController, children []string) {

	for _, child := range children {

		// create children WatchPathController, and watch it
		err := t.createChildWatchPath(wp, child)
		if err != nil {
			blog.Errorf("createChildWatchPath zk path %s/%s error %s", wp.Path, child, err.Error())
		}

	}
}

func (t *taskgroupWatcher) createChildWatchPath(wp *types.WatchPathController, child string) error {
	childP := fmt.Sprintf("%s/%s", wp.Path, child)

	//if zk path in caches
	_, ok, _ := t.caches.GetByKey(childP)
	if ok {
		blog.V(3).Infof("zk path %s is under watching", childP)
		return nil
	}

	blog.Info("createChildWatchPath zk path %s", childP)

	var pathType types.ZkPathType

	switch wp.Type {

	// zk path, like /blueking/application
	case types.ZkPathTypeApplication:
		pathType = types.ZkPathTypeNamespace

	// zk path, like /blueking/application/defaultGroup
	case types.ZkPathTypeNamespace:
		pathType = types.ZkPathTypeAppname

	// zk path, like /blueking/application/defaultGroup/app-001
	case types.ZkPathTypeAppname:
		pathType = types.ZkPathTypeTaskgroup

	// zk path, like /blueking/application/defaultGroup/app-001/{taskgroup}
	case types.ZkPathTypeTaskgroup:
		pathType = types.ZkPathTypeTask

	case types.ZkPathTypeTask:
		blog.Warn("ZkPathType %s not child path", types.ZkPathTypeTask)
		return nil

	default:
		return fmt.Errorf("ZkPathType %s is invalid", wp.Type)
	}

	cxt, cancel := context.WithCancel(wp.Cxt)

	taskWp := &types.WatchPathController{
		Path:   childP,
		Type:   pathType,
		Cxt:    cxt,
		Cancel: cancel,
	}

	return t.watchData(taskWp)
}

func (t *taskgroupWatcher) getZkValueByKey(key string) ([]byte, error) {
	by, state, err := t.zk.GetEx(key)
	if err != nil {
		return nil, fmt.Errorf("get zk path %s error %s", key, err.Error())
	}

	if state == nil {
		return nil, fmt.Errorf("get zk path %s state is nil", key)
	}

	return by, nil

}

func (t *taskgroupWatcher) handleDataChangedEvent(wp *types.WatchPathController) error {

	old, ok, err := t.caches.GetByKey(wp.Path)
	if err != nil {
		return err
	}

	//if ok = true,then update
	if ok {
		err = t.updateWatchData(wp, old)
		// else ok = false,then add
	} else {
		err = t.addWatchData(wp)
	}

	return err
}

func (t *taskgroupWatcher) updateWatchData(wp *types.WatchPathController, old interface{}) error {
	if wp.Type != types.ZkPathTypeTask {
		blog.V(3).Infof("if event type is zk.EventNodeDataChanged, only deal with zkPathType %s", types.ZkPathTypeTask)
		return nil
	}

	blog.V(3).Infof("updateWatchData zk path %s", wp.Path)

	newChecks, err := t.createHealthCheck(*wp)
	if err != nil {
		return err
	}

	oldWp, ok := old.(types.WatchPathController)
	if !ok {
		return fmt.Errorf("zk path %s WatchPathController type Assert failed", wp.Path)
	}

	oldChecks, ok := oldWp.Data.([]types.HealthCheck)
	if !ok {
		return fmt.Errorf("zk path %s []types.HealthCheck type Assert failed", wp.Path)
	}

	if len(newChecks) != len(oldChecks) {
		return fmt.Errorf("zk path %s old length %d != new length %d", wp.Path, len(oldChecks), len(newChecks))
	}

	var diff bool

	for i, check := range newChecks {

		if !reflect.DeepEqual(oldChecks[i], check) {
			diff = true
			break
		}
	}

	if !diff {
		return nil
	}

	wp.Data = newChecks

	err = t.caches.Update(*wp)
	if err != nil {
		blog.Error("cache add WatchPathController path %s error %s", wp.Path, err.Error())
		return err
	}

	for _, check := range newChecks {
		err = t.syncData(types.SyncDataActionUpdate, check)
		if err != nil {
			blog.Errorf("syncData error %s", err.Error())
		}
	}

	return nil
}

func (t *taskgroupWatcher) addWatchData(wp *types.WatchPathController) error {
	blog.Info("addWatchData zk path %s", wp.Path)

	if wp.Type == types.ZkPathTypeTask {
		checks, err := t.createHealthCheck(*wp)
		if err != nil {
			return err
		}

		wp.Data = checks
	}

	err := t.caches.Add(*wp)
	if err != nil {
		blog.Errorf("cache add data error %s", err.Error())
	}

	if wp.Type == types.ZkPathTypeTask {
		checks, _ := wp.Data.([]types.HealthCheck)

		for _, check := range checks {
			err = t.syncData(types.SyncDataActionAdd, check)
			if err != nil {
				blog.Errorf("syncData error %s", err.Error())
			}
		}

	}

	return nil
}

func (t *taskgroupWatcher) handleNodeDeleteEvent(wp *types.WatchPathController) error {
	_, ok, err := t.caches.GetByKey(wp.Path)
	if err != nil {
		return err
	}

	if !ok {
		blog.Warn("watch zk path %s had be deleted", wp.Path)
		return nil
	}

	// delete data in caches
	err = t.caches.Delete(*wp)
	if err != nil {
		blog.Error("deleteWatchData delete caches path %s error %s", wp.Path, err.Error())
	}

	if wp.Type == types.ZkPathTypeTask {
		checks, ok := wp.Data.([]types.HealthCheck)
		if !ok {
			return fmt.Errorf("zk path %s WatchPathController type Assert failed", wp.Path)
		}

		for _, check := range checks {
			err = t.syncData(types.SyncDataActionDelete, check)
			if err != nil {
				blog.Errorf("syncData error %s", err.Error())
			}
		}
	}

	return err
}

func (t *taskgroupWatcher) syncData(action types.SyncDataAction, data types.HealthCheck) error {
	// sync data to dataQueue
	t.syncHealthCheck(action, data)

	return nil
}

func (t *taskgroupWatcher) createHealthCheck(wp types.WatchPathController) ([]types.HealthCheck, error) {
	by, err := t.getZkValueByKey(wp.Path)
	if err != nil {
		return nil, err
	}

	var task *schtypes.Task

	err = json.Unmarshal(by, &task)
	if err != nil {
		return nil, fmt.Errorf("Unmarshal taskgroup error %s", err.Error())
	}

	checks, err := t.createHealthCheckByTask(task)
	if err != nil {
		return nil, err
	}

	return checks, nil
}

func (t *taskgroupWatcher) createHealthCheckByTask(task *schtypes.Task) ([]types.HealthCheck, error) {
	checks := make([]types.HealthCheck, 0)

	if len(task.HealthChecks) == 0 {
		return nil, fmt.Errorf("task %s have not healthcheck", task.ID)
	}

	for index, taskHealth := range task.HealthChecks {
		check, err := t.createHealthCheckByTaskHealth(taskHealth, task)
		if err != nil {
			blog.Errorf(err.Error())
			continue
		}

		check.ID = fmt.Sprintf("%d.%s", index, check.OriginID)
		checks = append(checks, check)
	}

	return checks, nil
}

func (t *taskgroupWatcher) createHealthCheckByTaskHealth(taskHealth *comtypes.HealthCheck, task *schtypes.Task) (types.HealthCheck, error) {
	check := types.HealthCheck{}

	check.OriginID = task.ID
	check.Operation = types.HealthCheckOperationStopped

	if taskHealth.Type != comtypes.BcsHealthCheckType_REMOTEHTTP && taskHealth.Type != comtypes.BcsHealthCheckType_REMOTETCP {
		return check, fmt.Errorf("task %s healtcheck type %s is invalid", task.ID, string(taskHealth.Type))
	}

	if task.StatusData == "" {
		blog.Errorf("task %s StatusData is empty", task.ID)
		return check, nil
	}

	var bcsinfo container.BcsContainerInfo

	err := json.Unmarshal([]byte(task.StatusData), &bcsinfo)
	if err != nil {
		blog.Errorf("Unmarshal task %s StatusData error %s", task.ID, err.Error())
		return check, nil
	}

	check.Type = taskHealth.Type

	check.DelaySeconds = taskHealth.DelaySeconds
	check.GracePeriodSeconds = taskHealth.GracePeriodSeconds
	check.IntervalSeconds = taskHealth.IntervalSeconds
	check.TimeoutSeconds = taskHealth.TimeoutSeconds
	check.ConsecutiveFailures = taskHealth.ConsecutiveFailures

	switch taskHealth.Type {
	//if health check type = REMOTE_HTTP
	case comtypes.BcsHealthCheckType_REMOTEHTTP:
		if taskHealth.Http == nil {
			return check, fmt.Errorf("task %s have not health check http", task.ID)
		}

		check.Http = &types.HttpHealthCheck{}

		if taskHealth.Http.Port > 0 {
			check.Http.Port = taskHealth.Http.Port
			check.Http.Ip, _ = t.getTargetIp(task)

		} else {
			check.Http.Ip, check.Http.Port, err = t.getTargetIpPort(task, taskHealth.Http.PortName)
			if err != nil {
				blog.Errorf("task %s getTargetIpPort error %s", task.ID, err.Error())
				return check, nil
			}
		}

		check.Http.Scheme = taskHealth.Http.Scheme
		check.Http.Path = taskHealth.Http.Path
		check.Http.Headers = taskHealth.Http.Headers

		// if health check type = REMOTE_TCP
	case comtypes.BcsHealthCheckType_REMOTETCP:
		if taskHealth.Tcp == nil {
			return check, fmt.Errorf("task %s have not health check tcp", task.ID)
		}

		check.Tcp = &types.TcpHealthCheck{}

		if taskHealth.Tcp.Port > 0 {
			check.Tcp.Port = taskHealth.Tcp.Port
			check.Tcp.Ip, _ = t.getTargetIp(task)

		} else {
			check.Tcp.Ip, check.Tcp.Port, err = t.getTargetIpPort(task, taskHealth.Tcp.PortName)
			if err != nil {
				blog.Errorf("task %s getTargetIpPort error %s", task.ID, err.Error())
				return check, nil
			}

		}

	}

	if task.Status == "Running" {
		check.Operation = types.HealthCheckOperationRunning
	}

	return check, nil
}

func (t *taskgroupWatcher) getTargetIp(task *schtypes.Task) (string, error) {

	var bcsinfo container.BcsContainerInfo
	var targetIp string

	err := json.Unmarshal([]byte(task.StatusData), &bcsinfo)
	if err != nil {
		return targetIp, fmt.Errorf("Unmarshal task %s StatusData error %s", task.ID, err.Error())
	}

	switch strings.ToLower(task.Network) {
	case "host":
		targetIp = bcsinfo.NodeAddress

	case "bridge":
		hostPort, _ := strconv.Atoi(bcsinfo.Ports[0].HostPort)

		if hostPort > 0 {
			targetIp = bcsinfo.NodeAddress
		} else {
			targetIp = bcsinfo.IPAddress
		}

	default:
		targetIp = bcsinfo.IPAddress
	}

	return targetIp, nil
}

func (t *taskgroupWatcher) getTargetIpPort(task *schtypes.Task, portName string) (string, int32, error) {

	var bcsinfo container.BcsContainerInfo
	var targetIp string
	var targetPort int32

	err := json.Unmarshal([]byte(task.StatusData), &bcsinfo)
	if err != nil {
		return targetIp, targetPort, fmt.Errorf("Unmarshal task %s StatusData error %s", task.ID, err.Error())
	}

	for _, onePort := range task.PortMappings {
		if onePort.Name == portName {

			switch strings.ToLower(task.Network) {
			case "host":
				targetIp = bcsinfo.NodeAddress
				targetPort = onePort.ContainerPort

			case "bridge":
				if onePort.HostPort > 0 {
					targetIp = bcsinfo.NodeAddress
					targetPort = onePort.HostPort
				} else {
					targetIp = bcsinfo.IPAddress
					targetPort = onePort.ContainerPort
				}

			default:
				targetIp = bcsinfo.IPAddress
				targetPort = onePort.ContainerPort
			}

			return targetIp, targetPort, nil
		}
	}

	return targetIp, targetPort, fmt.Errorf("Not found portName in task.PortMappings")
}
