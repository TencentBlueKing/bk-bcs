/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.,
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package controller

import (
	"context"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/common/types"
	cmdb "github.com/Tencent/bk-bcs/bcs-common/pkg/esb/cmdbv3"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/config"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/discovery"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/reconciler"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/storage"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskinformer"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-bkcmdb-synchronizer/taskmanager"
)

// Controller controller for controlling reconciler life cycle
type Controller struct {
	ops           *config.SyncOption
	disc          *discovery.Client
	storageClient storage.Interface
	cmdbClient    cmdb.ClientInterface
	informer      *taskinformer.Informer
	manager       *taskmanager.Manager
	isMaster      bool

	reconcilerMap map[string]*reconciler.Reconciler
	cancelFuncMap map[string]context.CancelFunc
}

// NewController create controller
func NewController(ops *config.SyncOption,
	serverInfo *types.ServerInfo,
	disc *discovery.Client,
	storageClient storage.Interface,
	cmdbClient cmdb.ClientInterface,
	informer *taskinformer.Informer,
	manager *taskmanager.Manager) *Controller {

	c := &Controller{
		ops:           ops,
		disc:          disc,
		informer:      informer,
		manager:       manager,
		storageClient: storageClient,
		cmdbClient:    cmdbClient,
		reconcilerMap: make(map[string]*reconciler.Reconciler),
		cancelFuncMap: make(map[string]context.CancelFunc),
	}

	informer.RegisterHandler(c)

	return c
}

// OnAdd implements informer event handler
func (c *Controller) OnAdd(add common.Cluster) {
	blog.Infof("cluster %+v add", add)
	// add new reconciler there is no reconciler for the cluster
	if _, ok := c.reconcilerMap[add.ClusterID]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		newReconciler, err := reconciler.NewReconciler(add, c.storageClient, c.cmdbClient, c.ops.FullSyncInterval)
		if err != nil {
			blog.Errorf("failed, to create new reconciler, err %s", err.Error())
			cancel()
			return
		}
		blog.Infof("add reconciler for cluster %+v", add)
		c.reconcilerMap[add.ClusterID] = newReconciler
		c.cancelFuncMap[add.ClusterID] = cancel
		go newReconciler.Run(ctx)
	} else {
		blog.Warnf("duplicated add cluster")
	}
}

// OnUpdate implements informer event handler
func (c *Controller) OnUpdate(old, new common.Cluster) {
	blog.Infof("cluster old %+v new %+v", old, new)
	if _, ok := c.reconcilerMap[new.ClusterID]; !ok {
		ctx, cancel := context.WithCancel(context.Background())
		newReconciler, err := reconciler.NewReconciler(new, c.storageClient, c.cmdbClient, c.ops.FullSyncInterval)
		if err != nil {
			blog.Errorf("failed, to create new reconciler, err %s", err.Error())
			cancel()
			return
		}
		blog.Infof("add reconciler for cluster %+v", new)
		c.reconcilerMap[new.ClusterID] = newReconciler
		c.cancelFuncMap[new.ClusterID] = cancel
		go newReconciler.Run(ctx)
	} else {
		blog.Infof("delete old reconciler for %+v", old)
		// call cancel function
		c.cancelFuncMap[old.ClusterID]()
		delete(c.cancelFuncMap, old.ClusterID)
		delete(c.reconcilerMap, old.ClusterID)

		blog.Infof("add new reconciler for %+v", new)
		ctx, cancel := context.WithCancel(context.Background())
		newReconciler, err := reconciler.NewReconciler(new, c.storageClient, c.cmdbClient, c.ops.FullSyncInterval)
		if err != nil {
			blog.Errorf("failed, to create new reconciler, err %s", err.Error())
			cancel()
			return
		}
		c.reconcilerMap[new.ClusterID] = newReconciler
		c.cancelFuncMap[new.ClusterID] = cancel
		go newReconciler.Run(ctx)
	}
}

// OnDelete implements informer event handler
func (c *Controller) OnDelete(del common.Cluster) {
	blog.Infof("cluster %+v delete", del)
	if _, ok := c.reconcilerMap[del.ClusterID]; ok {
		blog.Infof("delete del reconciler for %+v", del)
		// call cancel function
		c.cancelFuncMap[del.ClusterID]()
		delete(c.cancelFuncMap, del.ClusterID)
		delete(c.reconcilerMap, del.ClusterID)
	} else {
		blog.Infof("no reconciler for cluster %+v, need to delete", del)
	}
}

// Run run the controller
func (c *Controller) Run(ctx context.Context) {

	go c.informer.Run(ctx)

	c.masterLoop()
}

// masterLoop
func (c *Controller) masterLoop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			isMaster := c.disc.IsMaster()
			if isMaster && !c.isMaster {
				c.isMaster = isMaster
				blog.Infof("I become master, start task manager")
				go c.manager.Run()

			} else if !isMaster && c.isMaster {
				c.isMaster = isMaster
				blog.Infof("I become slave, stop task manager")
				c.manager.Stop()
			}
		}
	}
}
