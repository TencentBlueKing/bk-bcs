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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	mstore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckTaskExist check if task exists
func (store *managerStore) CheckTaskExist(task *types.Task) (string, bool) {
	obj, err := store.FetchDBTask(task.ID)
	if err == nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveTask save task to db
func (store *managerStore) SaveTask(task *types.Task) error {
	ns, _ := types.GetRunAsAndAppIDbyTaskID(task.ID)
	client := store.BkbcsClient.Tasks(ns)
	v2Task := &v2.Task{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdTask,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      task.ID,
			Namespace: ns,
		},
		Spec: v2.TaskSpec{
			Task: *task,
		},
	}

	var err error
	rv, exist := store.CheckTaskExist(task)
	if exist && rv != "" {
		v2Task.ResourceVersion = rv
		v2Task, err = client.Update(context.Background(), v2Task, metav1.UpdateOptions{})
	} else {
		v2Task, err = client.Create(context.Background(), v2Task, metav1.CreateOptions{})
	}
	if err != nil {
		if store.ObjectNotLatestErr(err) {
			store.syncTaskInCache(task.ID)
		}
		return err
	}

	task.ResourceVersion = v2Task.ResourceVersion
	saveCacheTask(task)
	return nil
}

func (store *managerStore) syncTaskInCache(taskId string) {
	task, err := store.FetchDBTask(taskId)
	if err != nil {
		blog.Errorf("fetch task(%s) in kube-apiserver failed: %s", taskId, err.Error())
		return
	}

	saveCacheTask(task)
}

// DeleteTask delete task
func (store *managerStore) DeleteTask(taskId string) error {
	ns, _ := types.GetRunAsAndAppIDbyTaskID(taskId)
	client := store.BkbcsClient.Tasks(ns)
	err := client.Delete(context.Background(), taskId, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteCacheTask(taskId)
	return nil
}

// FetchTask get task from cache
func (store *managerStore) FetchTask(taskId string) (*types.Task, error) {
	cacheTask, _ := fetchCacheTask(taskId)
	if cacheTask == nil {
		return nil, mstore.ErrNoFound
	}
	return cacheTask, nil
}

// FetchDBTask get task from db
func (store *managerStore) FetchDBTask(taskId string) (*types.Task, error) {
	ns, _ := types.GetRunAsAndAppIDbyTaskID(taskId)
	client := store.BkbcsClient.Tasks(ns)
	v2Task, err := client.Get(context.Background(), taskId, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, mstore.ErrNoFound
		}
		return nil, err
	}

	v2Task.Spec.Task.ResourceVersion = v2Task.ResourceVersion
	return &v2Task.Spec.Task, nil
}
