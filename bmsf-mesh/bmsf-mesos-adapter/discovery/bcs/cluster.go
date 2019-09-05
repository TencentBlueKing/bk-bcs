/*
Copyright (C) 2019 The BlueKing Authors. All rights reserved.

Permission is hereby granted, free of charge, to any person obtaining a copy of
this software and associated documentation files (the "Software"), to deal in
the Software without restriction, including without limitation the rights to
use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies
of the Software, and to permit persons to whom the Software is furnished to do
so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
*/

package bcs

import (
	"bk-bcs/bcs-common/common/blog"
	bcstypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-common/pkg/watch"
	"bk-bcs/bmsf-mesh/bmsf-mesos-adapter/discovery"
)

//NewCluster create cluster for bk-bcs scheduler
func NewCluster(clusterID string, hosts []string) (discovery.Cluster, error) {
	m := &bkbcsCluster{
		clusterName: clusterID,
	}
	//create service store
	svcCtl, err := newServiceCache(hosts)
	if err != nil {
		return nil, err
	}
	m.svcCtl = svcCtl
	//create taskgroup
	taskCtl, err := newTaskGroupCache(hosts)
	if err != nil {
		return nil, err
	}
	m.taskgroupsCtl = taskCtl
	return m, nil
}

//containerInfo hold info from BcsContainer
type containerInfo struct {
	IPAddress   string `json:"IPAddress"`
	NodeAddress string `json:"NodeAddress"`
}

//event inner event object
type svcEvent struct {
	EventType watch.EventType
	Old       *bcstypes.BcsService
	Cur       *bcstypes.BcsService
}

// type appEvent struct {
// 	EventType watch.EventType
// 	Old       *DiscoveryApp
// 	Cur       *DiscoveryApp
// }

type taskGroupEvent struct {
	EventType watch.EventType
	Old       *TaskGroup
	Cur       *TaskGroup
}

//bkbcsCluster bcs-scheduler cluster management
//discovery informations are based on BcsService.
type bkbcsCluster struct {
	clusterName   string               //cluster name
	svcCtl        *svcController       //service controller
	taskgroupsCtl *taskGroupController //taskgroup info controller
}

// GetName implementation for cluster
func (bcs *bkbcsCluster) GetName() string {
	return "bk-bcs"
}

// Run cluster event loop
func (bcs *bkbcsCluster) Run() {
	blog.Infof("bcs-scheduler cluster data plugin is ready to run...")
	//running backgroup recvLoop
	bcs.taskgroupsCtl.run()
	bcs.svcCtl.run()
}

// Stop close cluster event loop
func (bcs *bkbcsCluster) Stop() {
	//close all
	blog.Infof("bk-bcs cluster data plugin is ready to stop...")
	bcs.svcCtl.stop()
	bcs.taskgroupsCtl.stop()
}

// AppSvcs get controller of AppSvc
func (bcs *bkbcsCluster) AppSvcs() discovery.AppSvcController {
	return bcs.svcCtl
}

// AppNodes get controller of AppNode
func (bcs *bkbcsCluster) AppNodes() discovery.AppNodeController {
	return bcs.taskgroupsCtl
}
