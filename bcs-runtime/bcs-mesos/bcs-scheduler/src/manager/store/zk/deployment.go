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

package zk

import (
	"encoding/json"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

var deploymentLocks map[string]*sync.Mutex
var deploymentRWlock sync.RWMutex

func (store *managerStore) InitDeploymentLockPool() {
	if deploymentLocks == nil {
		blog.Info("init deployment lock pool")
		deploymentLocks = make(map[string]*sync.Mutex)
	}
}

func (store *managerStore) LockDeployment(deploymentName string) {
	deploymentRWlock.RLock()
	myLock, ok := deploymentLocks[deploymentName]
	deploymentRWlock.RUnlock()
	if ok {
		myLock.Lock()
		return
	}

	deploymentRWlock.Lock()
	myLock, ok = deploymentLocks[deploymentName]
	if !ok {
		blog.Info("create deployment lock(%s)", deploymentName)
		deploymentLocks[deploymentName] = new(sync.Mutex)
		myLock, _ = deploymentLocks[deploymentName]
	}
	deploymentRWlock.Unlock()

	myLock.Lock()
	return
}

func (store *managerStore) UnLockDeployment(deploymentName string) {
	deploymentRWlock.RLock()
	myLock, ok := deploymentLocks[deploymentName]
	deploymentRWlock.RUnlock()

	if !ok {
		blog.Error("deployment lock(%s) not exist when do unlock", deploymentName)
		return
	}
	myLock.Unlock()
}

func getDeploymentRootPath() string {
	return "/" + bcsRootNode + "/" + deploymentNode + "/"
}

func (store *managerStore) SaveDeployment(deployment *types.Deployment) error {

	data, err := json.Marshal(deployment)
	if err != nil {
		return err
	}

	path := getDeploymentRootPath() + deployment.ObjectMeta.NameSpace + "/" + deployment.ObjectMeta.Name
	return store.Db.Insert(path, string(data))
}

func (store *managerStore) FetchDeployment(ns, name string) (*types.Deployment, error) {
	path := getDeploymentRootPath() + ns + "/" + name
	data, err := store.Db.Fetch(path)
	if err != nil {
		return nil, err
	}
	deployment := &types.Deployment{}
	if err := json.Unmarshal(data, deployment); err != nil {
		blog.Error("fail to unmarshal deployment(%s). err:%s", string(data), err.Error())
		return nil, err
	}

	return deployment, nil
}

func (store *managerStore) ListDeployments(ns string) ([]*types.Deployment, error) {
	path := getDeploymentRootPath() + ns
	deploymentNodes, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list deploymentNodes path(%s), err:%s", path, err.Error())
		return nil, err
	}

	if nil == deploymentNodes {
		blog.Error("no deployments in (%s)", path)
		return nil, nil
	}

	var deployments []*types.Deployment
	for _, deploymentNode := range deploymentNodes {
		deployment, err := store.FetchDeployment(ns, deploymentNode)
		if err != nil {
			blog.Error("fail to fetch deployment(%s.%s)", ns, deploymentNode)
			continue
		}

		deployments = append(deployments, deployment)
	}

	return deployments, nil
}

func (store *managerStore) DeleteDeployment(ns, name string) error {

	path := getDeploymentRootPath() + ns + "/" + name
	blog.V(3).Infof("will delete deployment,path(%s)", path)
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete deployment(%s.%s), err:%s", ns, name, err.Error())
		return err
	}

	return nil
}

func (store *managerStore) ListDeploymentRunAs() ([]string, error) {

	rootPath := "/" + bcsRootNode + "/" + deploymentNode
	runAses, err := store.Db.List(rootPath)
	if err != nil {
		return nil, err
	}

	if nil == runAses {
		blog.Error("no runAs in (%s)", rootPath)
		return nil, nil
	}

	return runAses, nil
}

func (store *managerStore) ListDeploymentNodes(runAs string) ([]string, error) {

	path := getDeploymentRootPath() + runAs

	IDs, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list path:%s, err:%s", path, err.Error())
		return nil, err
	}

	return IDs, nil
}

func (store *managerStore) ListAllDeployments() ([]*types.Deployment, error) {
	nss, err := store.ListObjectNamespaces(deploymentNode)
	if err != nil {
		return nil, err
	}

	var objs []*types.Deployment
	for _, ns := range nss {
		obj, err := store.ListDeployments(ns)
		if err != nil {
			blog.Error("fail to fetch deployment by ns(%s)", ns)
			continue
		}

		objs = append(objs, obj...)
	}

	return objs, nil
}
