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
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/samuel/go-zookeeper/zk"
	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"
	schedtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster"
	clusteretcd "github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/cluster/etcd"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/service"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/storage"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
)

//ZkClient interface to define zk operation
//interface is only use for dependency injection
type ZkClient interface {
	ConnectEx(sessionTimeOut time.Duration) error
	GetEx(path string) ([]byte, *zk.Stat, error)
	GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error)
	GetChildrenEx(path string) ([]string, *zk.Stat, error)
	ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error)
	ExistEx(path string) (bool, *zk.Stat, error)
	State() zk.State
	Close()
}

//WatchInterface define interface for watch
type WatchInterface interface {
	cluster.DataExister
	cluster.EventHandler
	addWatch(path string) error       //post path for watch
	cleanWatch(path string) error     //post path to clean path
	setClient(client ZkClient)        //setting new client
	postData(data *types.BcsSyncData) //post single data manually/fix error data
	sync()                            //force sync all data
	stop()                            //stop all watch, reset all data
}

var (
	// SyncDefaultTimeOut default timeout
	SyncDefaultTimeOut = time.Second * 1
)

var (
	//ApplicationThreadNum goroutine number for Application channel
	ApplicationThreadNum int
	//TaskgroupThreadNum goroutine number for taskgroup channel
	TaskgroupThreadNum int
	//ExportserviceThreadNum goroutine number for exportservice channel
	ExportserviceThreadNum int
	//DeploymentThreadNum goroutine number for deployment channel
	DeploymentThreadNum int
)

//NewMesosCluster create mesos cluster
func NewMesosCluster(cfg *types.CmdConfig, st storage.Storage, netservice *service.InnerService) cluster.Cluster {

	blog.Info("mesos cluster(%s) will be created ...", cfg.ClusterID)

	ApplicationThreadNum = cfg.ApplicationThreadNum
	TaskgroupThreadNum = cfg.TaskgroupThreadNum
	ExportserviceThreadNum = cfg.ExportserviceThreadNum
	DeploymentThreadNum = cfg.DeploymentThreadNum

	linkItems := strings.Split(cfg.ClusterInfo, "/")
	mesos := &MesosCluster{
		zkLinks:        linkItems[0],
		watchPath:      "/" + strings.Join(linkItems[1:], "/"),
		clusterID:      cfg.ClusterID,
		storage:        st,
		netservice:     netservice,
		reportCallback: make(map[string]cluster.ReportFunc),
		existCallback:  make(map[string]cluster.DataExister),
	}
	//mesos watch initialize
	if err := mesos.initialize(); err != nil {
		blog.Error("mesos cluster initialize failed, %s", err.Error())
		return nil
	}
	return mesos
}

//MesosCluster cluster implements all cluster interface
type MesosCluster struct {
	zkLinks           string                         //zk connection link, like 127.0.0.1:2181,127.0.0.2:2181
	watchPath         string                         //zk watching path, like /blueking
	clusterID         string                         //watch cluster id
	client            ZkClient                       //client for zookeeper
	retry             bool                           //flag for reconnect zookeeper
	connCxt           context.Context                //context for client disconnected
	cancel            context.CancelFunc             //cancel func when client disconnected
	reportCallback    map[string]cluster.ReportFunc  //report map for handling registery data type
	existCallback     map[string]cluster.DataExister //check data exist in local cache
	storage           storage.Storage                //storage interface for remote Storage, like CC
	netservice        *service.InnerService
	app               *AppWatch           //watch for application
	taskGroup         *TaskGroupWatch     //watch for taskgroup
	exportSvr         *ExportServiceWatch //watch for exportservice
	Status            string              //curr status
	configmap         *ConfigMapWatch
	secret            *SecretWatch
	service           *ServiceWatch
	deployment        *DeploymentWatch
	endpoint          *EndpointWatch
	netServiceWatcher *clusteretcd.NetServiceWatcher
	stopCh            chan struct{}
}

//createZkConn create zookeeper connection with cluster
func (ms *MesosCluster) createZkConn() error {

	blog.Info("mesos cluster create ZK connection ...")
	servers := strings.Split(ms.zkLinks, ",")
	ms.client = zkclient.NewZkClient(servers)
	conErr := ms.client.ConnectEx(time.Second * 5)
	if conErr != nil {
		blog.Error("cluster connect zookeeper failed: %s", conErr.Error())
		return conErr
	}
	blog.Info("mesos cluster link zookeeper %s success! base watch path: %s ", ms.zkLinks, ms.watchPath)
	return nil
}

// GenerateRandnum just for test
func GenerateRandnum() int {
	rand.Seed(time.Now().Unix())
	rnd := rand.Intn(100)
	return rnd
}

//zkConnMonitor monitor zookeeper connection state
func (ms *MesosCluster) zkConnMonitor() {

	state := ms.client.State()
	if state == zk.StateConnected {
		//connection ok
		blog.V(3).Infof("############ZK Connection status is connected#####")
		return
	}

	if state == zk.StateHasSession {
		blog.V(3).Infof("############ZK Connection status is hassession#####")
		return
	}

	blog.V(3).Infof("zookeeper connection under %s", state.String())

	ms.Status = "zkfail"
}

func (ms *MesosCluster) registerReportHandler() error {

	blog.Info("mesos cluster register report handler...")

	ms.reportCallback["Application"] = ms.reportApplication
	for i := 0; i == 0 || i < ApplicationThreadNum; i++ {
		applicationChannel := types.ApplicationChannelPrefix + strconv.Itoa(i)
		ms.reportCallback[applicationChannel] = ms.reportApplication
	}

	ms.reportCallback["TaskGroup"] = ms.reportTaskGroup
	for i := 0; i == 0 || i < TaskgroupThreadNum; i++ {
		taskGroupChannel := types.TaskgroupChannelPrefix + strconv.Itoa(i)
		ms.reportCallback[taskGroupChannel] = ms.reportTaskGroup
	}

	ms.reportCallback["ExportService"] = ms.reportExportService
	for i := 0; i == 0 || i < ExportserviceThreadNum; i++ {
		exportserviceChannel := types.ExportserviceChannelPrefix + strconv.Itoa(i)
		ms.reportCallback[exportserviceChannel] = ms.reportExportService
	}

	ms.reportCallback["ConfigMap"] = ms.reportConfigMap
	ms.reportCallback["Service"] = ms.reportService
	ms.reportCallback["Secret"] = ms.reportSecret

	ms.reportCallback["Deployment"] = ms.reportDeployment
	for i := 0; i == 0 || i < DeploymentThreadNum; i++ {
		deploymentChannel := types.DeploymentChannelPrefix + strconv.Itoa(i)
		ms.reportCallback[deploymentChannel] = ms.reportDeployment
	}

	ms.reportCallback["Endpoint"] = ms.reportEndpoint

	// report ip pool static resource data callback.
	ms.reportCallback["IPPoolStatic"] = ms.reportIPPoolStatic

	// report ip pool static resource detail data callback.
	ms.reportCallback["IPPoolStaticDetail"] = ms.reportIPPoolStaticDetail

	return nil
}

func (ms *MesosCluster) registerExistHandler() error {

	blog.Info("mesos cluster register exist handler...")
	return nil
}

func (ms *MesosCluster) createDatTypeWatch() error {

	blog.Info("mesos cluster create exportservice watcher...")
	epSvrCxt, _ := context.WithCancel(ms.connCxt)
	ms.exportSvr = NewExportServiceWatch(epSvrCxt, ms.client, ms, ms.clusterID, ms.watchPath)

	blog.Info("mesos cluster create taskgroup watcher...")
	taskCxt, _ := context.WithCancel(ms.connCxt)
	ms.taskGroup = NewTaskGroupWatch(taskCxt, ms.client, ms)

	blog.Info("mesos cluster create app watcher...")
	appCxt, _ := context.WithCancel(ms.connCxt)
	ms.app = NewAppWatch(appCxt, ms.client, ms)

	ms.ProcessAppPathes()

	configmapCxt, _ := context.WithCancel(ms.connCxt)
	ms.configmap = NewConfigMapWatch(configmapCxt, ms.client, ms, ms.watchPath)
	go ms.configmap.Work()

	secretCxt, _ := context.WithCancel(ms.connCxt)
	ms.secret = NewSecretWatch(secretCxt, ms.client, ms, ms.watchPath)
	go ms.secret.Work()

	serviceCxt, _ := context.WithCancel(ms.connCxt)
	ms.service = NewServiceWatch(serviceCxt, ms.client, ms, ms.watchPath)
	go ms.service.Work()

	deploymentCxt, _ := context.WithCancel(ms.connCxt)
	ms.deployment = NewDeploymentWatch(deploymentCxt, ms.client, ms, ms.watchPath)
	go ms.deployment.Work()

	endpointCxt, _ := context.WithCancel(ms.connCxt)
	ms.endpoint = NewEndpointWatch(endpointCxt, ms.client, ms, ms.watchPath)
	go ms.endpoint.Work()

	ms.netServiceWatcher = clusteretcd.NewNetServiceWatcher(ms.clusterID, ms, ms.netservice)
	go ms.netServiceWatcher.Run(ms.stopCh)

	return nil
}

//ProcessAppPathes handle all Application datas
func (ms *MesosCluster) ProcessAppPathes() error {

	appPath := ms.watchPath + "/application"
	blog.V(3).Infof("Mesos cluster process path(%s)", appPath)
	nmList, _, err := ms.client.GetChildrenEx(appPath)
	if err != nil {
		blog.Error("get path(%s) children err: %s", appPath, err.Error())
		return err
	}
	if len(nmList) == 0 {
		blog.V(3).Infof("get empty namespace list under path(%s)", appPath)
		return nil
	}
	for _, nmNode := range nmList {
		blog.V(3).Infof("get node(%s) under path(%s)", nmNode, appPath)
		nmPath := appPath + "/" + nmNode
		ms.app.addWatch(nmPath)
	}

	return nil
}

func (ms *MesosCluster) initialize() error {

	blog.Info("mesos cluster(%s) initialize ...", ms.clusterID)

	//register ReporterHandler
	if err := ms.registerReportHandler(); err != nil {

		blog.Error("Mesos cluster register report handler err:%s", err.Error())
		return err
	}
	//register ExistHandler
	if err := ms.registerExistHandler(); err != nil {
		blog.Error("Mesos cluster register exist handler err:%s", err.Error())
		return err
	}

	//create connection to zookeeper
	if err := ms.createZkConn(); err != nil {
		blog.Error("Mesos cluster create ZK err:%s", err.Error())
		return err
	}

	ms.connCxt, ms.cancel = context.WithCancel(context.Background())
	ms.stopCh = make(chan struct{})

	//create datatype watch
	if err := ms.createDatTypeWatch(); err != nil {
		blog.Error("Mesos cluster create wathers err:%s", err.Error())
		return err
	}

	ms.Status = "running"

	return nil
}

//ReportData report data to reportHandler, handle all data independently
func (ms *MesosCluster) ReportData(data *types.BcsSyncData) error {
	callBack, ok := ms.reportCallback[data.DataType]
	if !ok {
		blog.Error("ReportHandler for %s do not register", data.DataType)
		return fmt.Errorf("ReportHandler for %s do not register", data.DataType)
	}
	return callBack(data)
}

func (ms *MesosCluster) reportService(data *types.BcsSyncData) error {
	dataType := data.Item.(*commtypes.BcsService)
	blog.V(3).Infof("mesos cluster report service(%s.%s) for action(%s)",
		dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action)

	ms.exportSvr.postData(data)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("service(%s.%s) sync(%s) dispatch failed: %+v",
			dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action, err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportConfigMap(data *types.BcsSyncData) error {
	dataType := data.Item.(*commtypes.BcsConfigMap)
	blog.V(3).Infof("mesos cluster report configmap(%s.%s) for action(%s)",
		dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("configmap(%s.%s) sync(%s) dispatch failed: %+v",
			dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action, err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportDeployment(data *types.BcsSyncData) error {
	dataType := data.Item.(*schedtypes.Deployment)
	blog.V(3).Infof("mesos cluster report deployment(%s.%s) for action(%s)",
		dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("deployment(%s.%s) sync(%s) dispatch failed: %+v",
			dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action, err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportSecret(data *types.BcsSyncData) error {
	dataType := data.Item.(*commtypes.BcsSecret)
	blog.V(3).Infof("mesos cluster report secret(%s.%s) for action(%s)",
		dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("secret(%s.%s) sync(%s) dispatch failed: %+v",
			dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action, err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportEndpoint(data *types.BcsSyncData) error {
	dataType := data.Item.(*commtypes.BcsEndpoint)
	blog.V(3).Infof("mesos cluster report endpoint(%s.%s) for action(%s)",
		dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("endpoint(%s.%s) sync(%s) dispatch failed: %+v",
			dataType.ObjectMeta.NameSpace, dataType.ObjectMeta.Name, data.Action, err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportIPPoolStatic(data *types.BcsSyncData) error {
	blog.V(3).Infof("mesos cluster report netservice ip pool static resource[%+v]", data.Item)

	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Errorf("mesos cluster report netservice ip pool static resource failed, %+v", err)
		return err
	}
	return nil
}

func (ms *MesosCluster) reportIPPoolStaticDetail(data *types.BcsSyncData) error {
	blog.V(3).Infof("mesos cluster report netservice ip pool static resource detail[%+v]", data.Item)

	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Errorf("mesos cluster report netservice ip pool static resource detail failed, %+v", err)
		return err
	}
	return nil
}

//reportTaskGroup
func (ms *MesosCluster) reportTaskGroup(data *types.BcsSyncData) error {

	taskgroup := data.Item.(*schedtypes.TaskGroup)
	blog.V(3).Infof("mesos cluster report taskgroup(%s) for action(%s)", taskgroup.ID, data.Action)

	//post to ExportService first
	ms.exportSvr.postData(data)
	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("TaskGroup sync dispatch failed: %+v", err)
		return err
	}
	return nil
}

//reportApplication
func (ms *MesosCluster) reportApplication(data *types.BcsSyncData) error {

	app := data.Item.(*schedtypes.Application)
	blog.V(3).Infof("mesos cluster report app(%s %s) for action(%s)", app.RunAs, app.ID, data.Action)

	path := ms.watchPath + "/application/" + app.RunAs + "/" + app.ID
	//check acton for Add & Delete
	if data.Action == types.ActionAdd {
		blog.Info("app(%s.%s) added, so add taskgroup pathwatch(%s)", app.RunAs, app.ID, path)
		ms.taskGroup.addWatch(path)
	} else if data.Action == types.ActionDelete {
		blog.Info("app(%s.%s) deleted, so clean taskgroup pathwatch(%s)", app.RunAs, app.ID, path)
		ms.taskGroup.cleanWatch(path)
	} else if data.Action == types.ActionUpdate {
		blog.V(3).Infof("app(%s.%s) updated, try to add taskgroup pathwatch(%s)", app.RunAs, app.ID, path)
		ms.taskGroup.addWatch(path)
	}

	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("Application sync dispatch failed: %+v", err)
		return err
	}
	return nil
}

//reportExportService
func (ms *MesosCluster) reportExportService(data *types.BcsSyncData) error {

	blog.V(3).Infof("mesos cluster report service for action(%s)", data.Action)

	if err := ms.storage.SyncTimeout(data, SyncDefaultTimeOut); err != nil {
		blog.Error("ExportService sync dispatch failed: %+v", err)
		return err
	}
	return nil
}

//Run running cluster watch
func (ms *MesosCluster) Run(cxt context.Context) {

	blog.Info("mesos cluster run ...")

	//ready to start zk connection monitor
	tick := time.NewTicker(5 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-cxt.Done():
			blog.Info("MesosCluster asked to exit...")
			return
		case <-tick.C:
			blog.V(3).Infof("tick: mesos cluster is alive ...")
			if ms.client != nil {
				ms.zkConnMonitor()
				ms.ProcessAppPathes()
			}
		}
	}
}

//Sync ask cluster to sync data to local cache
func (ms *MesosCluster) Sync(tp string) error {

	blog.Info("mesos cluster sync ...")

	return nil
}

//Stop ask cluster stopped
func (ms *MesosCluster) Stop() {

	blog.Info("mesos cluster stop ...")
	blog.Info("MesosCluster stop app watcher...")
	ms.app.stop()
	blog.Info("MesosCluster stop taskgroup watcher...")
	ms.taskGroup.stop()
	blog.Info("MesosCluster stop exportsvr watcher...")
	ms.exportSvr.stop()

	if ms.cancel != nil {
		ms.cancel()
	}
	close(ms.stopCh)
	time.Sleep(2 * time.Second)

	if ms.client != nil {
		ms.client.Close()
		ms.client = nil
	}
}

//GetClusterStatus get synchronization status
func (ms *MesosCluster) GetClusterStatus() string {
	return ms.Status
}

// GetClusterID get mesos clusterID
func (ms *MesosCluster) GetClusterID() string {
	return ms.clusterID
}
