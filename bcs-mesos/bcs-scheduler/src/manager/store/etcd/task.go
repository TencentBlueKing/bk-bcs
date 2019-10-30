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
	"bk-bcs/bcs-common/common/blog"
	"fmt"
	"strings"

	schStore "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func getNamespacebyTaskID(taskID string) string {
	appID := ""
	runAs := ""

	szSplit := strings.Split(taskID, ".")
	//runAs
	if len(szSplit) >= 5 {
		runAs = szSplit[4]
	}
	//appid
	if len(szSplit) >= 4 {
		appID = szSplit[3]
	}

	return fmt.Sprintf("%s-%s", runAs, appID)
}

func (store *managerStore) CheckTaskExist(task *types.Task) bool {
	_, err := store.FetchTask(task.ID)
	if err == nil {
		return true
	}

	return false
}

func (store *managerStore) SaveTask(task *types.Task) error {
	ns := getNamespacebyTaskID(task.ID)
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
	if store.CheckTaskExist(task) {
		v2Task.ResourceVersion = task.ResourceVersion
		v2Task, err = client.Update(v2Task)
	} else {
		v2Task, err = client.Create(v2Task)
	}
	if err != nil {
		return err
	}

	task.ResourceVersion = v2Task.ResourceVersion
	saveCacheTask(task)

	return nil
}

func (store *managerStore) ListTasks(runAs, appID string) ([]*types.Task, error) {
	ns := fmt.Sprintf("%s-%s", runAs, appID)
	client := store.BkbcsClient.Tasks(ns)
	v2Tasks, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	tasks := make([]*types.Task, 0, len(v2Tasks.Items))
	for _, task := range v2Tasks.Items {
		task.Spec.Task.ResourceVersion = task.ResourceVersion
		tasks = append(tasks, &task.Spec.Task)
	}
	return tasks, nil
}

func (store *managerStore) DeleteTask(taskId string) error {
	ns := getNamespacebyTaskID(taskId)
	client := store.BkbcsClient.Tasks(ns)
	err := client.Delete(taskId, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	deleteCacheTask(taskId)

	return nil
}

func (store *managerStore) FetchTask(taskId string) (*types.Task, error) {
	blog.Infof("fetch task %s", taskId)

	cacheTask, cacheErr := fetchCacheTask(taskId)
	if cacheErr == nil && cacheTask != nil {
		return cacheTask, nil
	}

	return store.FetchDBTask(taskId)
}

func (store *managerStore) FetchDBTask(taskId string) (*types.Task, error) {
	blog.Infof("fetch task %s in db", taskId)

	ns := getNamespacebyTaskID(taskId)
	client := store.BkbcsClient.Tasks(ns)
	v2Task, err := client.Get(taskId, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	v2Task.Spec.Task.ResourceVersion = v2Task.ResourceVersion
	return &v2Task.Spec.Task, nil
}
