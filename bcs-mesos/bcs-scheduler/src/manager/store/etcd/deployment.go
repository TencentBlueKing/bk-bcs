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

package etcd

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	v2 "github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"k8s.io/apimachinery/pkg/api/errors"
)

var deploymentLocks map[string]*sync.Mutex
var deploymentRWlock sync.RWMutex

// InitDeploymentLockPool init deployment lock pool
func (store *managerStore) InitDeploymentLockPool() {
	if deploymentLocks == nil {
		blog.Info("init deployment lock pool")
		deploymentLocks = make(map[string]*sync.Mutex)
	}
}

// LockDeployment lock deployment
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

// UnLockDeployment unlock deployment
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

// CheckDeploymentExist check if a deployment exists
func (store *managerStore) CheckDeploymentExist(deployment *types.Deployment) (string, bool) {
	obj, _ := store.fetchDeploymentInDB(deployment.ObjectMeta.NameSpace, deployment.ObjectMeta.Name)
	if obj != nil {
		return obj.ObjectMeta.ResourceVersion, true
	}

	return "", false
}

// SaveDeployment save deployment to db
func (store *managerStore) SaveDeployment(deployment *types.Deployment) error {
	err := store.checkNamespace(deployment.ObjectMeta.NameSpace)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.Deployments(deployment.ObjectMeta.NameSpace)
	v2Dep := &v2.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdDeployment,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        deployment.ObjectMeta.Name,
			Namespace:   deployment.ObjectMeta.NameSpace,
			Labels:      store.filterSpecialLabels(deployment.ObjectMeta.Labels),
			Annotations: deployment.ObjectMeta.Annotations,
		},
		Spec: v2.DeploymentSpec{
			Deployment: *deployment,
		},
	}

	rv, exist := store.CheckDeploymentExist(deployment)
	if exist {
		v2Dep.ResourceVersion = rv
		v2Dep, err = client.Update(context.Background(), v2Dep, metav1.UpdateOptions{})
	} else {
		v2Dep, err = client.Create(context.Background(), v2Dep, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	deployment.ObjectMeta.ResourceVersion = v2Dep.ResourceVersion
	saveCacheDeployment(deployment)
	return nil
}

// FetchDeployment fetch deployment by namespace and name
func (store *managerStore) FetchDeployment(ns, name string) (*types.Deployment, error) {
	dep := getCacheDeployment(ns, name)
	if dep == nil {
		return nil, schStore.ErrNoFound
	}
	return dep, nil
}

func (store *managerStore) fetchDeploymentInDB(ns, name string) (*types.Deployment, error) {
	client := store.BkbcsClient.Deployments(ns)
	v2Dep, err := client.Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	obj := v2Dep.Spec.Deployment
	obj.ObjectMeta.ResourceVersion = v2Dep.ResourceVersion
	return &obj, nil
}

// ListDeployments list deployments in one namespace
func (store *managerStore) ListDeployments(ns string) ([]*types.Deployment, error) {
	if cacheMgr.isOK {
		return listCacheRunAsDeployment(ns)
	}

	client := store.BkbcsClient.Deployments(ns)
	v2Deps, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := make([]*types.Deployment, 0, len(v2Deps.Items))
	for _, dep := range v2Deps.Items {
		obj := dep.Spec.Deployment
		obj.ObjectMeta.ResourceVersion = dep.ResourceVersion
		deployments = append(deployments, &obj)
	}

	return deployments, nil
}

// DeleteDeployment delete deployment by name and namespace
func (store *managerStore) DeleteDeployment(ns, name string) error {
	client := store.BkbcsClient.Deployments(ns)
	err := client.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheDeployment(ns, name)
	return nil
}

// ListDeploymentNodes list the node of deployment
func (store *managerStore) ListDeploymentNodes(runAs string) ([]string, error) {
	deployments, err := store.ListDeployments(runAs)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(deployments))
	for _, dep := range deployments {
		nodes = append(nodes, dep.ObjectMeta.Name)
	}
	return nodes, nil
}

// ListAllDeployments list all namespaces
func (store *managerStore) ListAllDeployments() ([]*types.Deployment, error) {
	if cacheMgr.isOK {
		return listCacheDeployments()
	}

	client := store.BkbcsClient.Deployments("")
	v2Deps, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	deployments := make([]*types.Deployment, 0, len(v2Deps.Items))
	for _, dep := range v2Deps.Items {
		obj := dep.Spec.Deployment
		obj.ObjectMeta.ResourceVersion = dep.ResourceVersion
		deployments = append(deployments, &obj)
	}

	return deployments, nil
}
