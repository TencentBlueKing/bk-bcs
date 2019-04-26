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
	"os"
	"strings"
	"time"

	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/task/cluster"
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/types"

	"bk-bcs/bcs-common/common/zkclient"
	"golang.org/x/net/context"
)

const (
	MaxDataQueueLength = 1024
)

type mesosCluster struct {
	conf config.HealthCheckConfig

	zk ZkClient

	cxt    context.Context
	cancel context.CancelFunc

	watchers map[WatcherType]Watcher

	dataQueue chan *types.HealthSyncData
}

func NewMesosCluster(rootCxt context.Context, conf config.HealthCheckConfig) cluster.Cluster {
	cxt, cancel := context.WithCancel(rootCxt)
	zkservs := strings.Split(conf.SchedDiscvSvr, ",")

	zk := zkclient.NewZkClient(zkservs)

	//conn,_,err := zk.Connect(zkservs,time.Second*5)

	err := zk.ConnectEx(time.Second * 5)
	if err != nil {
		blog.Error("Connect schedzk %s error %s", conf.SchedDiscvSvr, err.Error())
		os.Exit(1)
	}

	m := &mesosCluster{
		cxt:       cxt,
		cancel:    cancel,
		conf:      conf,
		zk:        zk,
		dataQueue: make(chan *types.HealthSyncData, MaxDataQueueLength),
	}

	return m
}

func (m *mesosCluster) initWatcher() {
	m.watchers = make(map[WatcherType]Watcher)

	// init taskgroup watcher
	taskgroup := NewTaskgroupWatcher(m.cxt, m.zk)
	m.watchers[WatcherTypeTaskgroup] = taskgroup
}

func (m *mesosCluster) Run() {
	blog.Info("mesosCluster running...")
	//TODO
	go m.runWatcher()
}

func (m *mesosCluster) runWatcher() {
	m.initWatcher()

	for k, watcher := range m.watchers {
		blog.Info("mesosCluster run Watcher %s", string(k))

		go watcher.Run()
		go m.clusterW(watcher)
	}
}

func (m *mesosCluster) clusterW(watcher Watcher) {
	tick := time.NewTicker(time.Second * 60)

	for {

		select {
		case <-tick.C:
			blog.V(3).Infof("TaskManager waiting for clusterW watcher DataW")

		case <-m.cxt.Done():
			blog.Warn("TaskManager stop clusterW watcher DataW")
			return

		case data := <-watcher.DataW():
			blog.V(3).Infof("mesosCluster rev data action %s checkId %s", string(data.Action), data.HealthCheck.ID)
			//TODO
			m.sync(data)

		}

	}
}

func (m *mesosCluster) sync(data *types.HealthSyncData) {
	//TODO

	data.HealthCheck.TaskMode = types.HealthTaskModeMesos

	m.dataQueue <- data
}

func (m *mesosCluster) DataW() <-chan *types.HealthSyncData {
	//TODO
	return m.dataQueue
}

/*func (m *mesosCluster) DataW() <-chan *types.HealthSyncData{
	//TODO
	data := make(chan *types.HealthSyncData,1)

	return data
}*/

func (m *mesosCluster) Stop() {
	//TODO
	blog.Info("mesosCluster stopped")
	m.cancel()
}
