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

package scheduler

import (
	alarm "bk-bcs/bcs-common/common/bcs-health/api"
	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/common/zkclient"
	"bk-bcs/bcs-common/pkg/cache"
	"bk-bcs/bcs-mesos/bcs-container-executor/container"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"encoding/json"
	"fmt"
	"github.com/samuel/go-zookeeper/zk"
	"reflect"
	"strings"
	"time"
)

// Event for service manager
type ServiceSyncData struct {
	// TaskGroup, Service
	DataType string
	// Add, Delete, Update
	Action string
	// Taskgroup or Service point
	Item interface{}
}

func esInfoKeyFunc(data interface{}) (string, error) {
	esInfo, ok := data.(*exportServiceInfo)
	if !ok {
		return "", fmt.Errorf("Error data type for exportServiceInfo")
	}
	return esInfo.bcsService.ObjectMeta.NameSpace + "." + esInfo.bcsService.ObjectMeta.Name, nil
}

type exportServiceInfo struct {
	bcsService *commtypes.BcsService
	//exportService *lbtypes.ExportService
	endpoint   *commtypes.BcsEndpoint
	createTime int64
	syncTime   int64
}

type zkClient interface {
	ConnectEx(sessionTimeOut time.Duration) error
	GetEx(path string) ([]byte, *zk.Stat, error)
	GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error)
	GetChildrenEx(path string) ([]string, *zk.Stat, error)
	ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error)
	ExistEx(path string) (bool, *zk.Stat, error)
	//ExistsW(path string) (bool, *zk.Stat, <-chan zk.Event, error)
	State() zk.State
	Close()
}

// Control message for service manager
type ServiceMgrMsg struct {
	// open:  work
	// close:  not work
	// stop: finish
	MsgType string
}

// Service Manager
type ServiceMgr struct {
	esInfoCache cache.Store
	queue       chan *ServiceSyncData
	zklink      string
	client      zkClient
	watchPath   string
	sched       *Scheduler
	msgQueue    chan *ServiceMgrMsg
	isWork      bool
}

// Create service manager
func NewServiceMgr(zkLink string, path string, scheduler *Scheduler) *ServiceMgr {
	mgr := &ServiceMgr{
		esInfoCache: cache.NewCache(esInfoKeyFunc),
		queue:       make(chan *ServiceSyncData, 4096),
		zklink:      zkLink,
		watchPath:   path,
		sched:       scheduler,
		msgQueue:    make(chan *ServiceMgrMsg, 128),
		isWork:      false,
	}

	err := mgr.createZkConn()
	if err != nil {
		return nil
	}
	return mgr
}

// Send control message to service manager
func (mgr *ServiceMgr) SendMsg(msg *ServiceMgrMsg) error {
	blog.Info("ServiceMgr: send an msg to service manager")
	select {
	case mgr.msgQueue <- msg:
	default:
		blog.Error("ServiceMgr: send an msg to service manager fail")
		return fmt.Errorf("service mgr queue is full now")
	}

	return nil
}

func (mgr *ServiceMgr) createZkConn() error {

	blog.Info("servicemgr create ZK connection(%s) ...", mgr.zklink)
	servers := strings.Split(mgr.zklink, ",")
	//var conErr error
	//mgr.client, _, conErr = zk.Connect(servers, time.Second*5)
	mgr.client = zkclient.NewZkClient(servers)
	conErr := mgr.client.ConnectEx(time.Second * 5)
	if conErr != nil {
		blog.Error("bcs-scheduler connect zookeeper failed: %s", conErr.Error())
		return conErr
	}
	blog.Info("connect zookeeper %s success! base mgr path: %s ", mgr.zklink, mgr.watchPath)
	return nil
}

func (mgr *ServiceMgr) zkConnMonitor() {
	state := mgr.client.State()
	if state == zk.StateConnected {
		blog.V(3).Infof("############ZK Connection status is connected#####")
		return
	}
	if state == zk.StateHasSession {
		blog.V(3).Infof("############ZK Connection status is hassession#####")
		return
	}
	blog.Warn("zookeeper connection under %s", state.String())
	mgr.client.Close()

	tryTimes := 0
	for {
		tryTimes++
		err := mgr.createZkConn()
		if err != nil {
			mgr.client.Close()
			time.Sleep(3 * time.Second)
		} else {
			break
		}

		if tryTimes >= 20 {
			mgr.sched.SendHealthMsg(alarm.WarnKind, "", err.Error(), "", nil)
			tryTimes = 0
		}
	}
}

// Send taskgroup update event to servie manager
func (mgr *ServiceMgr) TaskgroupUpdate(taskgroup *types.TaskGroup) {
	data := &ServiceSyncData{
		DataType: "TaskGroup",
		Action:   "Update",
		Item:     taskgroup,
	}
	mgr.postData(data)
	return
}

// Send taskgroup add event to servie manager
func (mgr *ServiceMgr) TaskgroupAdd(taskgroup *types.TaskGroup) {
	data := &ServiceSyncData{
		DataType: "TaskGroup",
		Action:   "Add",
		Item:     taskgroup,
	}
	mgr.postData(data)
	return
}

// Send taskgroup delete event to servie manager
func (mgr *ServiceMgr) TaskgroupDelete(taskgroup *types.TaskGroup) {
	data := &ServiceSyncData{
		DataType: "TaskGroup",
		Action:   "Delete",
		Item:     taskgroup,
	}
	mgr.postData(data)
	return
}

// Send service updat event to servie manager
func (mgr *ServiceMgr) ServiceUpdate(service *commtypes.BcsService) {
	data := &ServiceSyncData{
		DataType: "Service",
		Action:   "Update",
		Item:     service,
	}
	mgr.postData(data)
	return
}

// Send service add event to servie manager
func (mgr *ServiceMgr) ServiceAdd(service *commtypes.BcsService) {
	data := &ServiceSyncData{
		DataType: "Service",
		Action:   "Add",
		Item:     service,
	}
	mgr.postData(data)
	return
}

// Send service delete event to servie manager
func (mgr *ServiceMgr) ServiceDelete(service *commtypes.BcsService) {
	data := &ServiceSyncData{
		DataType: "Service",
		Action:   "Delete",
		Item:     service,
	}
	mgr.postData(data)
	return
}
func (mgr *ServiceMgr) postData(data *ServiceSyncData) {
	if data == nil {
		return
	}
	blog.V(3).Infof("post data(type:%s, action:%s) to ServiceMgr", data.DataType, data.Action)
	mgr.queue <- data
}

// The goroutine function for service monitoring
// This function will process events of taskgrou add, delete and update
// This function will process events of service add, delete and update
func (mgr *ServiceMgr) Worker() {
	tick := time.NewTicker(300 * time.Second)
	for {
		select {
		case req := <-mgr.msgQueue:
			blog.Info("ServiceMgr: receive msg: %s, current queue(%d/%d)", req.MsgType, len(mgr.msgQueue), cap(mgr.msgQueue))
			if req.MsgType == "open" {
				mgr.isWork = true
				mgr.processAllServices()
			} else if req.MsgType == "close" {
				mgr.isWork = false
				mgr.esInfoCache.Clear()
			} else if req.MsgType == "stop" {
				mgr.stop()
				blog.Info("ServiceMgr: goroutine finish!")
				return
			}
		case <-tick.C:
			mgr.zkConnMonitor()
			blog.V(3).Infof("ServiceMgr is running, managed service num: %d", mgr.esInfoCache.Num())
			if mgr.isWork == false {
				continue
			}
			mgr.processAllServices()
			mgr.doCheck()

		case data := <-mgr.queue:
			blog.V(3).Infof("ServiceMgr: receive data: %s, current queue(%d/%d)", data.DataType, len(mgr.queue), cap(mgr.queue))
			if mgr.isWork == false {
				continue
			}
			if data.DataType == "TaskGroup" {
				switch data.Action {
				case "Add":
					mgr.addTaskGroup(data.Item.(*types.TaskGroup))
				case "Delete":
					mgr.deleteTaskGroup(data.Item.(*types.TaskGroup))
				case "Update":
					mgr.updateTaskGroup(data.Item.(*types.TaskGroup))
				}
			} else if data.DataType == "Service" {
				currTime := time.Now().Unix()
				switch data.Action {
				case "Add":
					mgr.addService(data.Item.(*commtypes.BcsService), currTime)
				case "Delete":
					mgr.deleteService(data.Item.(*commtypes.BcsService))
				case "Update":
					mgr.updateService(data.Item.(*commtypes.BcsService), currTime)
				}
			} else {
				blog.Warn("ServiceMgr recieve unknown data action(type:%s, action:%s)", data.DataType, data.Action)
			}
		}
	}
}

func (mgr *ServiceMgr) doCheck() {
	blog.Info("ServiceMgr doCheck: begin")
	checkNum := 0

	keyList := mgr.esInfoCache.ListKeys()
	for _, key := range keyList {
		blog.V(3).Infof("ServiceMgr doCheck: bcsEndpoint %s", key)
		checkNum++
		cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
		if err != nil {
			blog.Error("ServiceMgr doCheck: get %s return err:%s", key, err.Error())
			continue
		}
		if exist == false {
			blog.Error("ServiceMgr doCheck: get %s return not exist", key)
			continue
		}
		esInfo, ok := cacheData.(*exportServiceInfo)
		if !ok {
			blog.Error("ServiceMgr doCheck: convert %s fail", key)
			continue
		}

		mgr.syncEndpointInfo(esInfo)
		mgr.sched.store.SaveEndpoint(esInfo.endpoint)
	}

	blog.Info("ServiceMgr doCheck end: refresh %d services", checkNum)
	return
}

func (mgr *ServiceMgr) processAllServices() error {
	currTime := time.Now().Unix()
	basePath := mgr.watchPath + "/service"
	blog.Info("sync all services under(%s), currTime(%d)", basePath, currTime)

	nmList, _, err := mgr.client.GetChildrenEx(basePath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", basePath, err.Error())
		return err
	}
	if len(nmList) == 0 {
		blog.Warn("get empty namespace list under path(%s)", basePath)
		return nil
	}

	// sync all services from zk and update cache, create add and update events
	numZk := 0
	numDel := 0
	for _, nmNode := range nmList {
		blog.V(3).Infof("get namespace node(%s) under path(%s)", nmNode, basePath)
		nmPath := basePath + "/" + nmNode
		nodeList, _, err := mgr.client.GetChildrenEx(nmPath)
		if err != nil {
			blog.Error("get children nodes under %s err: %s", nmPath, err.Error())
			continue
		}
		for _, oneNode := range nodeList {
			numZk++
			blog.V(3).Infof("get node(%s) under path(%s)", oneNode, nmPath)
			nodePath := nmPath + "/" + oneNode
			byteData, _, err := mgr.client.GetEx(nodePath)
			if err != nil {
				blog.Error("Get %s data err: %s", nodePath, err.Error())
				continue
			}
			data := new(commtypes.BcsService)
			if jsonErr := json.Unmarshal(byteData, data); jsonErr != nil {
				blog.Error("Parse %s json data(%s) Err: %s", nodePath, string(byteData), jsonErr.Error())
				continue
			}

			key := data.ObjectMeta.NameSpace + "." + data.ObjectMeta.Name
			cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
			if err != nil {
				blog.Error("get service %s from cache return err:%s", key, err.Error())
				continue
			}
			if exist == true {
				cacheDataInfo, ok := cacheData.(*exportServiceInfo)
				if !ok {
					blog.Error("convert cachedata to exportServiceInfo fail, key(%s)", key)
					continue
				}
				if !reflect.DeepEqual(cacheDataInfo.bcsService, data) {
					blog.Warnf("service %s is changed, do update and init its endpoint", key)
					mgr.updateService(data, currTime)
				} else {
					blog.V(3).Infof("service %s is not changed, update sync time(%d)", key, currTime)
					cacheDataInfo.syncTime = currTime
				}
			} else {
				blog.Info("service %s is not in cache, to do add, time(%d)", key, currTime)
				mgr.addService(data, currTime)
				//bcsEndpoint, _ := mgr.sched.store.FetchEndpoint(data.ObjectMeta.NameSpace, data.ObjectMeta.Name)
				//if bcsEndpoint != nil {
				//	blog.Info("service %s is already has endpoint data in ZK, add to cache", key)
				//	mgr.addService(data, bcsEndpoint, currTime)
				//} else {
				//	blog.Info("service %s is has not endpoint data in ZK, create for it", key)
				//	mgr.addService(data, nil, currTime)
				//}
			}
		}
	}

	// check cache, create delete events
	keyList := mgr.esInfoCache.ListKeys()
	for _, key := range keyList {
		blog.V(3).Infof("to check cache service %s", key)
		cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
		if err != nil {
			blog.Error("service %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("service %s in cache keylist, but get return not exist", key)
			continue
		}
		cacheDataInfo, ok := cacheData.(*exportServiceInfo)
		if !ok {
			blog.Error("convert cachedata to ServiceInfo fail, key(%s)", key)
			continue
		}
		if cacheDataInfo.syncTime != currTime {
			numDel++
			blog.Warn("service %s is in cache, but syncTime(%d) != currTime(%d), to delete ",
				key, cacheDataInfo.syncTime, currTime)
			mgr.deleteService(cacheDataInfo.bcsService)
		}
	}

	blog.Info("sync %d services from zk, delete %d cache services", numZk, numDel)
	return nil
}

func (mgr *ServiceMgr) addService(service *commtypes.BcsService, tNow int64) {
	if service == nil {
		return
	}
	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	blog.Info("ServiceMgr recieve addService(%s, %+v)", key, service)

	_, exist, err := mgr.esInfoCache.GetByKey(key)
	if err != nil {
		blog.Error("when recieve addService event, get esInfo %s from cache return err:%s", key, err.Error())
	}
	if exist == true {
		blog.Warn("when receive addService event, esInfo %s is in already in cache, will be updated", key)
	}

	esInfo := mgr.createServiceInfo(service)
	if esInfo == nil {
		blog.Error("when receive addService %s, createServiceInfo fail", key)
		return
	}

	esInfo.syncTime = tNow
	esInfo.createTime = tNow
	mgr.esInfoCache.Add(esInfo)
	blog.Info("add service end, save endpoint key(%s)", key)
	mgr.sched.store.SaveEndpoint(esInfo.endpoint)
	return
}

func (mgr *ServiceMgr) updateService(service *commtypes.BcsService, tNow int64) {
	if service == nil {
		return
	}
	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	blog.Info("ServiceMgr recieve updateService(%s, %+v)", key, service)

	createTime := tNow
	cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
	if err != nil {
		blog.Warn("when receive updateService event, get esInfo %s from cache return err:%s", key, err.Error())
	}
	if exist == false {
		blog.Warn("when receive updateService event, esInfo %s is in not in cache, will be added", key)
	} else {
		esInfo, ok := cacheData.(*exportServiceInfo)
		if ok {
			createTime = esInfo.createTime
		} else {
			blog.Warn("when receive updateService %s, cache data error, to update to new service: %+v ", key, service)
		}
	}

	esInfoNew := mgr.createServiceInfo(service)
	if esInfoNew == nil {
		blog.Error("when receive updateService %s, createServiceInfo fail", key)
		return
	}

	esInfoNew.syncTime = tNow
	esInfoNew.createTime = createTime
	mgr.esInfoCache.Add(esInfoNew)
	blog.Info("update service end, save endpoint key(%s)", key)
	mgr.sched.store.SaveEndpoint(esInfoNew.endpoint)

	return
}

func (mgr *ServiceMgr) deleteService(service *commtypes.BcsService) {
	if service == nil {
		return
	}
	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	blog.Info("ServiceMgr recieve deleteService(%+v)", service)

	cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
	if exist == false || err != nil {
		blog.Warn("when receive deleteService event, esInfo %s is not in cache or error", key)
		blog.Info("delete endpoint key(%s)", key)
		mgr.sched.store.DeleteEndpoint(service.ObjectMeta.NameSpace, service.ObjectMeta.Name)

		return
	}
	esInfo, ok := cacheData.(*exportServiceInfo)
	if !ok {
		blog.Warn("when receive deleteService event, convert cachedata to exportServiceInfo fail, key(%s)", key)
		blog.Info("delete endpoint key(%s)", key)
		mgr.sched.store.DeleteEndpoint(service.ObjectMeta.NameSpace, service.ObjectMeta.Name)
		return
	}

	blog.Info("delete endpoint key(%s)", key)
	mgr.sched.store.DeleteEndpoint(service.ObjectMeta.NameSpace, service.ObjectMeta.Name)

	mgr.esInfoCache.Delete(esInfo)
	return
}

func (mgr *ServiceMgr) createServiceInfo(service *commtypes.BcsService) *exportServiceInfo {
	if service == nil {
		return nil
	}
	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	esInfo := new(exportServiceInfo)
	esInfo.bcsService = service

	esInfo.endpoint = new(commtypes.BcsEndpoint)
	esInfo.endpoint.ObjectMeta = service.ObjectMeta
	esInfo.endpoint.TypeMeta = service.TypeMeta
	mgr.syncEndpointInfo(esInfo)


	blog.Info("ServiceMgr: service(%s) endpoint created, endpoint len(%d) ", key, len(esInfo.endpoint.Endpoints))

	return esInfo
}

func (mgr *ServiceMgr) syncEndpointInfo(esInfo *exportServiceInfo) {

	key := esInfo.bcsService.ObjectMeta.NameSpace + "." + esInfo.bcsService.ObjectMeta.Name

	basePath := fmt.Sprintf("%s/application/%s", mgr.watchPath, esInfo.bcsService.ObjectMeta.NameSpace)
	blog.V(3).Infof("sync all taskgroups under(%s) for service(%s)", basePath, key)

	appList, _, err := mgr.client.GetChildrenEx(basePath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", basePath, err.Error())
		return
	}

	esInfo.endpoint.Endpoints = nil

	for _, app := range appList {
		appPath := fmt.Sprintf("%s/%s", basePath, app)

		by, _, err := mgr.client.GetEx(appPath)
		if err != nil {
			blog.Warnf("get application(%s) from zk error: %s", appPath, err.Error())
			continue
		}
		var application *types.Application
		err = json.Unmarshal(by, &application)
		if err != nil {
			blog.Warnf("json.Unmarshal application(%s) error: %s", appPath, err.Error())
			continue
		}
		label := mgr.getApplicationServiceLabel(esInfo.bcsService, application)
		if label == "" {
			blog.V(3).Infof("application(%s) not match service: %s", appPath, key)
			continue
		}

		blog.Infof("sync all taskgroups under(%s) for service(%s)", appPath, key)
		tgList, _, err := mgr.client.GetChildrenEx(appPath)
		if err != nil {
			blog.Error("get path(%s) children err: %s", appPath, err.Error())
			continue
		}

		for _, tg := range tgList {
			tgPath := fmt.Sprintf("%s/%s", appPath, tg)
			by, _, err := mgr.client.GetEx(tgPath)
			if err != nil {
				blog.Errorf("get zk path %s error %s", tgPath, err.Error())
				continue
			}

			tskgroup := new(types.TaskGroup)
			err = json.Unmarshal(by, tskgroup)
			if err != nil {
				blog.Errorf("json.Unmarshal zk path %s data failed, error %s", tgPath, err.Error())
				continue
			}
			if tskgroup.Taskgroup == nil || len(tskgroup.Taskgroup) == 0 {
				blog.Error("taskgroup(%s) has no Task Info", tskgroup.ID)
				continue
			}
			if tskgroup.Status != types.TASKGROUP_STATUS_RUNNING && tskgroup.Status != types.TASKGROUP_STATUS_LOST {
				blog.V(3).Infof("taskgroup(%s) status %s, do nothing ", tskgroup.ID, tskgroup.Status)
				continue
			}

			//label := mgr.getTaskGroupServiceLabel(esInfo.bcsService, tskgroup)
			//if label == "" {
			//	blog.V(3).Infof("taskgroup(%s) not match service(%s) ", tskgroup.ID, key)
			//	continue
			//}
			//blog.V(3).Infof("taskgroup(%s) match service(%s)", tskgroup.ID, key)

			podEndpoint := mgr.buildEndpoint(esInfo.bcsService, tskgroup)
			if podEndpoint == nil {
				blog.Error("build service(%s) endpoint(%s) return nil", key, tskgroup.ID)
				continue
			}

			changed := mgr.addEndPoint(esInfo.endpoint, podEndpoint)
			if changed == true {
				blog.Info("service(%s) and endpoint(%s), curr len(%d) ", key, tskgroup.ID, len(esInfo.endpoint.Endpoints))
			}
		}
	}

	return
}

func (mgr *ServiceMgr) getTaskGroupServiceLabel(service *commtypes.BcsService, tskgroup *types.TaskGroup) string {

	if tskgroup.ObjectMeta.NameSpace != "" && service.ObjectMeta.NameSpace != tskgroup.ObjectMeta.NameSpace {
		return ""
	}

	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	for ks, vs := range service.Spec.Selector {
		task := tskgroup.Taskgroup[0]
		if task.Labels == nil {
			return ""
		}
		for kt, vt := range task.Labels {
			//blog.V(3).Infof("check task(%s) label(%s:%s) with selector label(%s:%s)", task.Name, kt, vt, ks, vs)
			if ks == kt && vs == vt {
				blog.V(3).Infof("task label match service: task(%s) label(%s:%s) service(%s)", task.Name, kt, vt, key)
				return vt
			}
		}
	}
	return ""
}

func (mgr *ServiceMgr) getApplicationServiceLabel(service *commtypes.BcsService, app *types.Application) string {
	if service.ObjectMeta.NameSpace != app.ObjectMeta.NameSpace {
		blog.V(3).Infof("namespace of service (%s.%s) and application (%s.%s) is different",
			service.NameSpace, service.Name, app.ObjectMeta.NameSpace, app.ID)
		return ""
	}

	key := service.ObjectMeta.NameSpace + "." + service.ObjectMeta.Name
	for ks, vs := range service.Spec.Selector {

		for kt, vt := range app.ObjectMeta.Labels {
			if ks == kt && vs == vt {
				blog.V(3).Infof("application label match service: application(%s.%s) label(%s:%s) service(%s)",
					app.RunAs, app.ID, kt, vt, key)
				return vt
			}
		}
	}
	return ""
}

func (mgr *ServiceMgr) buildEndpoint(service *commtypes.BcsService, tskgroup *types.TaskGroup) *commtypes.Endpoint {
	podEndpoint := new(commtypes.Endpoint)
	podEndpoint.NodeIP = ""
	podEndpoint.ContainerIP = ""
	podEndpoint.Target.Kind = "taskgroup"
	podEndpoint.Target.ID = tskgroup.ID
	podEndpoint.Target.Name = tskgroup.Name
	podEndpoint.Target.Namespace = tskgroup.RunAs
	bcsInfo := new(container.BcsContainerInfo)
	oneEndpointPort := new(commtypes.ContainerPort)
	for _, oneTask := range tskgroup.Taskgroup {
		//if oneTask.Status != types.TASK_STATUS_RUNNING {
		//	blog.V(3).Infof("ServiceMgr: buildEndpoint, but task %s is not running", oneTask.ID)
		//	continue
		//}

		var nodeAddress string
		// added  20180815, process task do not have the statusData upload by executor, because process executor
		// do not have the hostIP and port information. So we make NodeIP, ContainerIP, HostIP directly with AgentIPAddress
		// which is got from offer
		// current running taskgroup kind maybe empty, regard them as APP.
		switch oneTask.Kind {
		case commtypes.BcsDataType_PROCESS:
			podEndpoint.NetworkMode = oneTask.Network
			podEndpoint.NodeIP = oneTask.AgentIPAddress
			podEndpoint.ContainerIP = oneTask.AgentIPAddress
			nodeAddress = oneTask.AgentIPAddress
		case commtypes.BcsDataType_APP, "":
			if len(oneTask.StatusData) == 0 {
				blog.Warn("ServiceMgr: buildEndpoint, but task %s StatusData is empty", oneTask.ID)
				continue
			}
			blog.V(3).Infof("ServiceMgr: buildEndpoint, task %s StatusData: %s", oneTask.ID, oneTask.StatusData)
			if err := json.Unmarshal([]byte(oneTask.StatusData), bcsInfo); err != nil {
				blog.Warn("ServiceMgr: buildEndpoint, task %s StatusData unmarshal err: %s, cannot add to backend",
					oneTask.ID, err.Error())
				continue
			}

			podEndpoint.NetworkMode = oneTask.Network
			if bcsInfo.NodeAddress != "" {
				podEndpoint.NodeIP = bcsInfo.NodeAddress
			}
			if bcsInfo.IPAddress != "" {
				podEndpoint.ContainerIP = bcsInfo.IPAddress
			}
			nodeAddress = bcsInfo.NodeAddress
		default:
			continue
		}

		for _, onePort := range oneTask.PortMappings {
			for _, servicePort := range service.Spec.Ports {
				if onePort.Name == servicePort.Name {
					blog.V(3).Infof("ServiceMgr: buildEndpoint, task(%s) and service(%s) match port: %s",
						oneTask.ID, service.ObjectMeta.Name, onePort.Name)
					oneEndpointPort.Name = onePort.Name
					oneEndpointPort.HostPort = int(onePort.HostPort)
					oneEndpointPort.ContainerPort = int(onePort.ContainerPort)
					oneEndpointPort.HostIP = nodeAddress
					oneEndpointPort.Protocol = onePort.Protocol
					podEndpoint.Ports = append(podEndpoint.Ports, *oneEndpointPort)
				}
			}
		}
	}

	blog.V(3).Infof("ServiceMgr: build taskgroup(%s) Endpoint(%+v)", tskgroup.ID, podEndpoint)
	return podEndpoint
}

func (mgr *ServiceMgr) addEndPoint(bcsEndpoint *commtypes.BcsEndpoint, endpoint *commtypes.Endpoint) bool {
	for index, onePoint := range bcsEndpoint.Endpoints {
		if onePoint.Target.ID == endpoint.Target.ID && onePoint.Target.Namespace == endpoint.Target.Namespace {
			if !reflect.DeepEqual(onePoint, *endpoint) {
				blog.Info("ServiceMgr: endpoint(%s %s) for %s.%s changed, real update",
					endpoint.Target.Namespace, endpoint.Target.ID, bcsEndpoint.ObjectMeta.NameSpace, bcsEndpoint.ObjectMeta.Name)
				bcsEndpoint.Endpoints = append(bcsEndpoint.Endpoints[:index], bcsEndpoint.Endpoints[index+1:]...)
				bcsEndpoint.Endpoints = append(bcsEndpoint.Endpoints, *endpoint)
				return true
			}
		
			blog.V(3).Infof("ServiceMgr: endpoint(%s %s) for %s.%s not changed, ignore it",
				endpoint.Target.Namespace, endpoint.Target.ID, bcsEndpoint.ObjectMeta.NameSpace, bcsEndpoint.ObjectMeta.Name)
			return false
		}
	}

	blog.Info("ServiceMgr: endpoint(%s %s) for %s.%s real add",
		endpoint.Target.Namespace, endpoint.Target.ID, bcsEndpoint.ObjectMeta.NameSpace, bcsEndpoint.ObjectMeta.Name)
	bcsEndpoint.Endpoints = append(bcsEndpoint.Endpoints, *endpoint)
	return true
}

func (mgr *ServiceMgr) deleteEndPoint(bcsEndpoint *commtypes.BcsEndpoint, endpoint *commtypes.Endpoint) bool {
	for index, oldPoint := range bcsEndpoint.Endpoints {
		if oldPoint.Target.ID == endpoint.Target.ID && oldPoint.Target.Namespace == endpoint.Target.Namespace {
			blog.Info("ServiceMgr: endpoint(%s %s) for %s.%s real delete",
				oldPoint.Target.Namespace, oldPoint.Target.ID, bcsEndpoint.ObjectMeta.NameSpace, bcsEndpoint.ObjectMeta.Name)
			bcsEndpoint.Endpoints = append(bcsEndpoint.Endpoints[:index], bcsEndpoint.Endpoints[index+1:]...)
			return true
		}
	}

	return false
}






func (mgr *ServiceMgr) addTaskGroup(tskgroup *types.TaskGroup) {
	blog.Info("ServiceMgr receive taskgroup add event, %s: %s", tskgroup.ID, tskgroup.Status)
	if tskgroup.Taskgroup == nil || len(tskgroup.Taskgroup) == 0 {
		blog.Error("ServiceMgr receive taskgroup add event, but TaskGroup %s has no Task Info", tskgroup.ID)
		return
	}

	if tskgroup.Status != types.TASKGROUP_STATUS_RUNNING && tskgroup.Status != types.TASKGROUP_STATUS_LOST {
		blog.V(3).Infof("ServiceMgr receive taskgroup add event, TaskGroup %s status %s, do nothing ", tskgroup.ID, tskgroup.Status)
		return
	}

	keyList := mgr.esInfoCache.ListKeys()
	for _, key := range keyList {
		cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
		if err != nil {
			blog.Error("esInfo %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("esInfo %s in cache keylist, but get return not exist", key)
			continue
		}
		esInfo, ok := cacheData.(*exportServiceInfo)
		if !ok {
			blog.Error("convert cachedata to exportServiceInfo fail, key(%s)", key)
			continue
		}

		// check matching of selector and task label
		label := mgr.getTaskGroupServiceLabel(esInfo.bcsService, tskgroup)
		if label == "" {
			continue
		}
		blog.V(3).Infof("ServiceMgr: %s, match task label(%s:%s) ", key, tskgroup.ID, label)

		podEndpoint := mgr.buildEndpoint(esInfo.bcsService, tskgroup)
		if podEndpoint == nil {
			blog.Error("ServiceMgr receive taskgroup(%s) add event, build service(%s) endpoint return nil",
				tskgroup.ID, esInfo.bcsService.ObjectMeta.Name)
			continue
		}

		changed := mgr.addEndPoint(esInfo.endpoint, podEndpoint)
		if changed == true {
			blog.Info("ServiceMgr add taskgroup: service(%s) endpoint len(%d) ", key, len(esInfo.endpoint.Endpoints))
			mgr.sched.store.SaveEndpoint(esInfo.endpoint)
		}

	}

	blog.Info("ServiceMgr receive taskgroup add event end, %s: %s", tskgroup.ID, tskgroup.Status)

	return
}

func (mgr *ServiceMgr) updateTaskGroup(tskgroup *types.TaskGroup) {
	blog.V(3).Infof("ServiceMgr receive taskgroup update event, %s: %s", tskgroup.ID, tskgroup.Status)
	if tskgroup.Taskgroup == nil || len(tskgroup.Taskgroup) == 0 {
		blog.Error("ServiceMgr receive taskgroup update event, but TaskGroup %s has no Task Info", tskgroup.ID)
		return
	}

	keyList := mgr.esInfoCache.ListKeys()
	for _, key := range keyList {
		cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
		if err != nil {
			blog.Error("esInfo %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("esInfo %s in cache keylist, but get return not exist", key)
			continue
		}
		esInfo, ok := cacheData.(*exportServiceInfo)
		if !ok {
			blog.Error("convert cachedata to exportServiceInfo fail, key(%s)", key)
			continue
		}
		// check matching of selector and task label
		label := mgr.getTaskGroupServiceLabel(esInfo.bcsService, tskgroup)
		if label == "" {
			continue
		}
		blog.V(3).Infof("ServiceMgr: %s, match task label(%s: %s) ", key, tskgroup.ID, label)

		podEndpoint := mgr.buildEndpoint(esInfo.bcsService, tskgroup)
		if podEndpoint == nil {
			blog.Error("ServiceMgr receive taskgroup(%s) update event, build service(%s) endpoint return nil",
				tskgroup.ID, esInfo.bcsService.ObjectMeta.Name)
			continue
		}

		var changed bool
		if tskgroup.Status == types.TASKGROUP_STATUS_RUNNING || tskgroup.Status == types.TASKGROUP_STATUS_LOST {
			changed = mgr.addEndPoint(esInfo.endpoint, podEndpoint)
		} else {
			changed = mgr.deleteEndPoint(esInfo.endpoint, podEndpoint)
		}
		if changed == true {
			blog.Info("ServiceMgr update taskgroup: service(%s) endpoint len(%d)", key, len(esInfo.endpoint.Endpoints))
			mgr.sched.store.SaveEndpoint(esInfo.endpoint)
		}


	}

	blog.V(3).Infof("ServiceMgr receive taskgroup update event end, %s: %s", tskgroup.ID, tskgroup.Status)

	return
}

func (mgr *ServiceMgr) deleteTaskGroup(tskgroup *types.TaskGroup) {
	blog.V(3).Infof("ServiceMgr receive taskgroup delete event, %s: %s", tskgroup.ID, tskgroup.Status)

	if tskgroup.Taskgroup == nil || len(tskgroup.Taskgroup) == 0 {
		blog.Error("ServiceMgr receive taskgroup delete event, but TaskGroup %s has no Task Info", tskgroup.ID)
		return
	}

	keyList := mgr.esInfoCache.ListKeys()
	for _, key := range keyList {
		cacheData, exist, err := mgr.esInfoCache.GetByKey(key)
		if err != nil {
			blog.Error("esInfo %s in cache keylist, but get return err:%s", err.Error())
			continue
		}
		if exist == false {
			blog.Error("esInfo %s in cache keylist, but get return not exist", key)
			continue
		}
		esInfo, ok := cacheData.(*exportServiceInfo)
		if !ok {
			blog.Error("convert cachedata to exportServiceInfo fail, key(%s)", key)
			continue
		}
		// check matching of selector and task label
		label := mgr.getTaskGroupServiceLabel(esInfo.bcsService, tskgroup)
		if label == "" {
			continue
		}

		blog.V(3).Infof("ServiceMgr: %s, match task label(%s: %s) ", key, tskgroup.ID, label)

		podEndpoint := mgr.buildEndpoint(esInfo.bcsService, tskgroup)
		if podEndpoint == nil {
			blog.Error("ServiceMgr receive taskgroup(%s) delete event, build service(%s) endpoint return nil",
				tskgroup.ID, esInfo.bcsService.ObjectMeta.Name)
			continue
		}

		changed := mgr.deleteEndPoint(esInfo.endpoint, podEndpoint)
		if changed == true {
			blog.Info("ServiceMgr delete taskgroup: service(%s) endpoint len(%d) ", key, len(esInfo.endpoint.Endpoints))
			mgr.sched.store.SaveEndpoint(esInfo.endpoint)
		}

	}

	return
}

func (mgr *ServiceMgr) stop() {
	mgr.esInfoCache.Clear()
}
