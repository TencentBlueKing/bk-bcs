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

package task

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/config"
	checkManager "github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/manager"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/task/cluster"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/task/cluster/mesos"
	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-check/bcscheck/types"

	"golang.org/x/net/context"
)

const (
	MaxDataQueueLength = 1024
)

type taskManager struct {
	conf config.HealthCheckConfig

	dataQueue chan *types.HealthSyncData

	cxt    context.Context
	cancel context.CancelFunc

	manager checkManager.Manager

	clusters map[types.HealthTaskMode]cluster.Cluster
}

func NewTaskManager(rootCxt context.Context, conf config.HealthCheckConfig, manager checkManager.Manager) TaskManager {
	cxt, cancel := context.WithCancel(rootCxt)

	tm := &taskManager{
		cxt:       cxt,
		cancel:    cancel,
		manager:   manager,
		conf:      conf,
		dataQueue: make(chan *types.HealthSyncData, MaxDataQueueLength),
	}

	return tm
}

func (t *taskManager) initCluster() {
	t.clusters = make(map[types.HealthTaskMode]cluster.Cluster)

	// init mesosCluster
	mesosCluster := mesos.NewMesosCluster(t.cxt, t.conf)
	t.clusters[types.HealthTaskModeMesos] = mesosCluster
}

func (t *taskManager) Run() {
	blog.Info("taskManager running...")

	// run cluster
	go t.runClusters()

	// sync data to Manager
	go t.handleDataQueue()
}

func (t *taskManager) runClusters() {
	t.initCluster()

	for k, cluster := range t.clusters {
		blog.Info("taskManager run cluster %s", string(k))

		go cluster.Run()
		go t.clusterW(cluster)
	}
}

func (t *taskManager) clusterW(cluster cluster.Cluster) {
	tick := time.NewTicker(time.Second * 10)

	for {

		select {
		case <-tick.C:
			blog.V(3).Info("TaskManager handle data queue")

		case <-t.cxt.Done():
			blog.Warn("TaskManager stop handle data queue")
			return

		case data := <-cluster.DataW():
			blog.Info("taskManager rev data action %s checkId %s", string(data.Action), data.HealthCheck.ID)

			t.sync(data)

		}

	}
}

func (t *taskManager) sync(data *types.HealthSyncData) error {

	t.dataQueue <- data

	return nil
}

func (t *taskManager) Stop() {
	t.cancel()
}

func (t *taskManager) handleDataQueue() {

	tick := time.NewTicker(time.Second * 10)

	var err error

	for {

		select {
		case <-tick.C:
			blog.V(3).Info("TaskManager handle data queue")

		case <-t.cxt.Done():
			blog.Warn("TaskManager stop handle data queue")
			return

		case data := <-t.dataQueue:
			blog.V(3).Infof("taskManager handleDataQueue action %s checkId %s", data.Action, data.HealthCheck.ID)
			//TODO
			err = t.handleData(data)

		}

		if err != nil {
			blog.Error("TaskManager handleDataQueue error %s", err.Error())
		}
	}
}

func (t *taskManager) handleData(data *types.HealthSyncData) error {
	err := t.manager.Sync(data)

	return err
}
