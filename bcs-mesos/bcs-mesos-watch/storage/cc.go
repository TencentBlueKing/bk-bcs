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

package storage

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/http/httpclient"
	commtypes "github.com/Tencent/bk-bcs/bcs-common/common/types"
	lbtypes "github.com/Tencent/bk-bcs/bcs-common/pkg/loadbalance/v2"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
	schedtypes "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

//CCResponse response struct from CC
type CCResponse struct {
	Result  bool   `json:"result"`
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    string `json:"data,omitempty"`
}

//CCListData hold all data
type CCListData struct {
	Apps          []map[string]*schedtypes.Application `json:"application,omitempty"`
	TaskGroup     []map[string]*schedtypes.TaskGroup   `json:"taskgroup,omitempty"`
	ExportService []map[string]*lbtypes.ExportService  `json:"exportservice,omitempty"`
}

//NewCCStorage create storage for cc
func NewCCStorage(config *types.CmdConfig) (Storage, error) {
	//init zookeeper writer
	ccStorage := &CCStorage{
		rwServers: new(sync.RWMutex),
		//dcServer:  "",
		queue:                  make(chan *types.BcsSyncData, 10240),
		handlers:               make(map[string]*ChannelProxy),
		ClusterID:              config.ClusterID,
		applicationThreadNum:   config.ApplicationThreadNum,
		taskgroupThreadNum:     config.TaskgroupThreadNum,
		exportserviceThreadNum: config.ExportserviceThreadNum,
		deploymentThreadNum:    config.DeploymentThreadNum,
	}
	if err := ccStorage.init(); err != nil {
		return nil, err
	}

	ccStorage.client = httpclient.NewHttpClient()
	if "" != config.CAFile || "" != config.CertFile {
		ccStorage.client.SetTlsVerity(config.CAFile, config.CertFile, config.KeyFile, config.PassWord)
	}
	ccStorage.client.SetHeader("Content-Type", "application/json")
	ccStorage.client.SetHeader("Accept", "application/json")

	return ccStorage, nil
}

//CCStorage writing data to CC
type CCStorage struct {
	rwServers   *sync.RWMutex
	dcServer    []string
	dcServerIdx int

	//zkServer  string
	//zkClient  *zkclient.ZkClient
	queue                  chan *types.BcsSyncData  //queue for handling data
	exitCxt                context.Context          //context for exit
	handlers               map[string]*ChannelProxy //channel proxy
	ClusterID              string                   //
	applicationThreadNum   int
	taskgroupThreadNum     int
	exportserviceThreadNum int
	deploymentThreadNum    int
	client                 *httpclient.HttpClient // http client to do with request.
}

//init init CCStorage
func (cc *CCStorage) init() error {
	cc.handlers["Application"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 10240),
		actionHandler: &AppHandler{
			oper:         cc,
			dataType:     "application",
			ClusterID:    cc.ClusterID,
			DoCheckDirty: true,
		},
	}

	for i := 0; i == 0 || i < cc.applicationThreadNum; i++ {
		applicationChannel := types.ApplicationChannelPrefix + strconv.Itoa(i)
		cc.handlers[applicationChannel] = &ChannelProxy{
			dataQueue: make(chan *types.BcsSyncData, 10240),
			actionHandler: &AppHandler{
				oper:         cc,
				dataType:     "application",
				ClusterID:    cc.ClusterID,
				DoCheckDirty: false,
			},
		}
	}

	//
	for i := 0; i == 0 || i < cc.taskgroupThreadNum; i++ {
		taskGroupChannel := types.TaskgroupChannelPrefix + strconv.Itoa(i)
		cc.handlers[taskGroupChannel] = &ChannelProxy{
			dataQueue: make(chan *types.BcsSyncData, 10240),
			actionHandler: &TaskGroupHandler{
				oper:         cc,
				dataType:     "taskgroup",
				ClusterID:    cc.ClusterID,
				DoCheckDirty: false,
			},
		}
	}

	for i := 0; i == 0 || i < cc.exportserviceThreadNum; i++ {
		exportserviceChannel := types.ExportserviceChannelPrefix + strconv.Itoa(i)
		cc.handlers[exportserviceChannel] = &ChannelProxy{
			dataQueue: make(chan *types.BcsSyncData, 10240),
			actionHandler: &ExpServiceHandler{
				oper:      cc,
				dataType:  "exportservice",
				ClusterID: cc.ClusterID,
			},
		}
	}

	for i := 0; i == 0 || i < cc.deploymentThreadNum; i++ {
		deploymentChannel := types.DeploymentChannelPrefix + strconv.Itoa(i)
		cc.handlers[deploymentChannel] = &ChannelProxy{
			dataQueue: make(chan *types.BcsSyncData, 10240),
			actionHandler: &DeploymentHandler{
				oper:         cc,
				dataType:     "deployment",
				ClusterID:    cc.ClusterID,
				DoCheckDirty: false,
			},
		}
	}

	cc.handlers["TaskGroup"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 10240),
		actionHandler: &TaskGroupHandler{
			oper:         cc,
			dataType:     "taskgroup",
			ClusterID:    cc.ClusterID,
			DoCheckDirty: true,
		},
	}

	//
	cc.handlers["ExportService"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 10240),
		actionHandler: &ExpServiceHandler{
			oper:      cc,
			dataType:  "exportservice",
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers["Service"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &ServiceHandler{
			oper:      cc,
			dataType:  "service",
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers["ConfigMap"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &ConfigMapHandler{
			oper:      cc,
			dataType:  "configmap",
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers["Secret"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &SecretHandler{
			oper:      cc,
			dataType:  "secret",
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers["Deployment"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &DeploymentHandler{
			oper:         cc,
			dataType:     "deployment",
			ClusterID:    cc.ClusterID,
			DoCheckDirty: true,
		},
	}

	cc.handlers["Endpoint"] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &EndpointHandler{
			oper:      cc,
			dataType:  "endpoint",
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers[dataTypeIPPoolStatic] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &NetServiceHandler{
			oper:      cc,
			dataType:  dataTypeIPPoolStatic,
			ClusterID: cc.ClusterID,
		},
	}

	cc.handlers[dataTypeIPPoolStaticDetail] = &ChannelProxy{
		dataQueue: make(chan *types.BcsSyncData, 1024),
		actionHandler: &NetServiceHandler{
			oper:      cc,
			dataType:  dataTypeIPPoolStaticDetail,
			ClusterID: cc.ClusterID,
		},
	}

	return nil
}

//SetDCAddress had better add rwlock
func (cc *CCStorage) SetDCAddress(address []string) {
	blog.Info("CCStorage set DC address: %s", address)
	cc.rwServers.Lock()
	cc.dcServer = address
	cc.dcServerIdx = 0
	cc.rwServers.Unlock()
	return
}

//GetDCAddress get bcs-storage address
func (cc *CCStorage) GetDCAddress() string {

	address := ""
	cc.rwServers.Lock()

	if cc.dcServerIdx < 0 || cc.dcServerIdx >= len(cc.dcServer) {
		cc.dcServerIdx = 0
	}

	if cc.dcServerIdx < len(cc.dcServer) {
		address = cc.dcServer[cc.dcServerIdx]
		cc.dcServerIdx++
	}

	cc.rwServers.Unlock()
	return address
}

//Sync sync data to storage
func (cc *CCStorage) Sync(data *types.BcsSyncData) error {
	if data == nil {
		blog.Error("CCWriter get nil BcsInstance pointer")
		return nil
	}

	cc.queue <- data
	return nil
}

//Run start point for StorageWriter
func (cc *CCStorage) Run(cxt context.Context) error {
	cc.exitCxt = cxt
	for name, handler := range cc.handlers {
		blog.Info("CCStorage starting %s data channel", name)
		hCxt, _ := context.WithCancel(cxt)
		go handler.Run(hCxt)
	}
	go cc.Worker()
	return nil
}

// Worker storage writer worker goroutine
func (cc *CCStorage) Worker() {
	blog.Info("CCStorage ready to go into worker!")
	tick := time.NewTicker(120 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-tick.C:
			blog.Info("tick: ccStorage walker is alive, current task queue(%d/%d)", len(cc.queue), cap(cc.queue))
		case <-cc.exitCxt.Done():
			blog.Info("CCStorage Get exit signal, ready to exit, current task queue(%d/%d)", len(cc.queue), cap(cc.queue))
			return
		case data := <-cc.queue:

			if len(cc.queue)+1024 > cap(cc.queue) {
				blog.Warnf("CCStorage task busy, current task queue(%d/%d)", len(cc.queue), cap(cc.queue))
			} else {
				blog.V(3).Infof("CCStorage recieve task, current queue(%d/%d)", len(cc.queue), cap(cc.queue))
			}

			if handler, ok := cc.handlers[data.DataType]; ok {
				handler.Handle(data)
			} else {
				blog.Error("Get unkown DataType: %s", data.DataType)
			}
		}
	}
}

//CreateDCNode bcs-storage create operation
func (cc *CCStorage) CreateDCNode(node string, value interface{}, action string) error {

	if len(node) == 0 || value == nil {
		blog.Error("CCStorage Get empty node or value")
		return fmt.Errorf("CCStorage Get empty node or value")
	}

	path := cc.GetDCAddress() + node
	reportData := &commtypes.BcsStorageDynamicIf{
		Data: value,
	}

	valueBytes, err := json.Marshal(reportData)
	if err != nil {
		blog.Error("marsha1 json for %s failed: %+v", path, err)
		return err
	}

	begin := time.Now().UnixNano() / 1e6

	resp, rerr := cc.client.Request(path, action, nil, valueBytes)
	if rerr != nil {
		blog.Warn("DC [%s %s] err: %+v, retry", action, path, rerr)

		//do retry
		resp, rerr = cc.client.Request(path, action, nil, valueBytes)
		if rerr != nil {
			blog.Error("retry DC [%s %s] err: %+v", action, path, rerr)
			return rerr
		}
	}
	bodyStr := string(resp)

	end := time.Now().UnixNano() / 1e6
	useTime := end - begin
	if useTime > 100 {
		blog.Warnf("DC %d ms [%s %s] response: %s , slow query", useTime, action, path, bodyStr)
	} else {
		blog.Infof("DC %d ms [%s %s] response: %s ", useTime, action, path, bodyStr)
	}

	return nil
}

//DeleteDCNode storage delete operation
func (cc *CCStorage) DeleteDCNode(node, action string) error {
	if len(node) == 0 {
		blog.Error("CCStorage Get empty node")
		return fmt.Errorf("Get empty node")
	}

	path := cc.GetDCAddress() + node

	//blog.V(3).Infof("DC [%s %s] begin", action, path)

	begin := time.Now().UnixNano() / 1e6

	resp, rerr := cc.client.Request(path, action, nil, nil)
	if rerr != nil {
		blog.Warn("DC [%s %s] err: %+v, retry", action, path, rerr)

		//do retry
		resp, rerr = cc.client.Request(path, action, nil, nil)
		if rerr != nil {
			blog.Error("retry DC [%s %s] err: %+v", action, path, rerr)
			return rerr
		}
	}

	bodyStr := string(resp)

	end := time.Now().UnixNano() / 1e6
	useTime := end - begin
	if useTime > 100 {
		blog.Warnf("DC %d ms [%s %s] response: %s , slow query", useTime, action, path, bodyStr)
	} else {
		blog.Infof("DC %d ms [%s %s] response: %s ", useTime, action, path, bodyStr)
	}

	return nil
}

//DeleteDCNodes bcs-storage delete operation
func (cc *CCStorage) DeleteDCNodes(node string, value interface{}, action string) error {

	if len(node) == 0 || value == nil {
		blog.Error("CCStorage Get empty node or value")
		return fmt.Errorf("CCStorage Get empty node or value")
	}

	path := cc.GetDCAddress() + node

	valueBytes, err := json.Marshal(value)
	if err != nil {
		blog.Error("marsha1 json for %s failed: %+v", node, err)
		return err
	}

	//blog.V(3).Infof("DC [%s %s] begin", action, path)

	resp, rerr := cc.client.Request(path, action, nil, valueBytes)
	if rerr != nil {
		blog.Warn("DC %s %s err: %+v, retry", action, path, rerr)
		//do retry
		resp, rerr = cc.client.Request(path, action, nil, valueBytes)
		if rerr != nil {
			blog.Error("retry DC %s %s err: %+v", action, path, rerr)
			return rerr
		}
	}
	bodyStr := string(resp)
	blog.Info("DC [%s %s] response: %s, req-data: %s", action, path, bodyStr, string(valueBytes))
	return nil
}
