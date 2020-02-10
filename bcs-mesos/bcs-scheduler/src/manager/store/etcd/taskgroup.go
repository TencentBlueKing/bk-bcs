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
	schStore "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (store *managerStore) CheckTaskGroupExist(taskGroup *types.TaskGroup) (string, bool) {
	obj, err := store.FetchTaskGroup(taskGroup.ID)
	if err == nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

//SaveTaskGroup save task group to store
func (store *managerStore) SaveTaskGroup(taskGroup *types.TaskGroup) error {
	client := store.BkbcsClient.TaskGroups(taskGroup.RunAs)
	v2Taskgroup := &v2.TaskGroup{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdTaskGroup,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        taskGroup.ID,
			Namespace:   taskGroup.RunAs,
			Labels:      store.filterSpecialLabels(taskGroup.ObjectMeta.Labels),
			Annotations: taskGroup.ObjectMeta.Annotations,
		},
		Spec: v2.TaskGroupSpec{
			TaskGroup: *taskGroup,
		},
	}

	var err error
	rv, exist := store.CheckTaskGroupExist(taskGroup)
	if exist {
		v2Taskgroup.ResourceVersion = rv
		v2Taskgroup, err = client.Update(v2Taskgroup)
	} else {
		v2Taskgroup, err = client.Create(v2Taskgroup)
	}
	if err != nil {
		return err
	}

	taskGroup.ResourceVersion = v2Taskgroup.ResourceVersion
	saveCacheTaskGroup(taskGroup)
	//save task
	if taskGroup.Taskgroup != nil {
		for _, task := range taskGroup.Taskgroup {
			err := store.SaveTask(task)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (store *managerStore) listTaskgroupsInDB(runAs, appID string) ([]*types.TaskGroup, error) {
	client := store.BkbcsClient.TaskGroups(runAs)
	v2Taskgroups, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	taskgroups := make([]*types.TaskGroup, 0, len(v2Taskgroups.Items))
	for _, taskgroup := range v2Taskgroups.Items {
		if taskgroup.Spec.AppID == appID {
			obj := taskgroup.Spec.TaskGroup
			obj.ResourceVersion = taskgroup.ResourceVersion
			//get tasks
			taskIds := make([]string, 0, len(obj.Taskgroup))
			for _, task := range obj.Taskgroup {
				taskIds = append(taskIds, task.ID)
			}
			obj.Taskgroup = make([]*types.Task, len(taskIds))
			for index, taskID := range taskIds {
				task, err := store.FetchDBTask(taskID)
				if err != nil {
					blog.Error("fail to get task by ID(%s), err:%s", taskID, err.Error())
					return nil, err
				}

				obj.Taskgroup[index] = task
			}

			taskgroups = append(taskgroups, &obj)
		}
	}

	return taskgroups, nil
}

//ListTaskGroups show us all the task group on line
func (store *managerStore) ListTaskGroups(runAs, appID string) ([]*types.TaskGroup, error) {
	if cacheMgr.isOK {
		cacheList, _ := listCacheTaskGroups(runAs, appID)
		return cacheList, nil
	}

	return store.listTaskgroupsInDB(runAs, appID)
}

//DeleteTaskGroup delete a task group with executor id is taskGroupID
func (store *managerStore) DeleteTaskGroup(taskGroupID string) error {
	taskgroup, err := store.FetchTaskGroup(taskGroupID)
	if err != nil {
		return err
	}

	taskIds := make([]string, 0, len(taskgroup.Taskgroup))
	for _, task := range taskgroup.Taskgroup {
		taskIds = append(taskIds, task.ID)
	}

	//delte tasks
	for _, taskID := range taskIds {
		if err := store.DeleteTask(taskID); err != nil {
			blog.Error("fail to delete task(ID:%s), err:%s", taskID, err.Error())
			return err
		}
	}

	runAs, _ := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	client := store.BkbcsClient.TaskGroups(runAs)
	err = client.Delete(taskGroupID, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	deleteCacheTaskGroup(taskGroupID)

	return nil
}

//FetchTaskGroup fetch a types.TaskGroup
func (store *managerStore) FetchTaskGroup(taskGroupID string) (*types.TaskGroup, error) {
	if cacheMgr.isOK {
		cacheTaskgroup, _ := fetchCacheTaskGroup(taskGroupID)
		if cacheTaskgroup == nil {
			return nil, schStore.ErrNoFound
		}
		return cacheTaskgroup, nil
	}

	return store.FetchDBTaskGroup(taskGroupID)
}

//FetchTaskGroup fetch a types.TaskGroup
func (store *managerStore) FetchDBTaskGroup(taskGroupID string) (*types.TaskGroup, error) {
	runAs, _ := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	client := store.BkbcsClient.TaskGroups(runAs)
	v2Taskgroup, err := client.Get(taskGroupID, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	taskGroup := v2Taskgroup.Spec.TaskGroup
	taskIds := make([]string, 0, len(taskGroup.Taskgroup))
	for _, task := range taskGroup.Taskgroup {
		taskIds = append(taskIds, task.ID)
	}

	taskGroup.Taskgroup = make([]*types.Task, len(taskIds))
	for index, taskID := range taskIds {
		task, err := store.FetchDBTask(taskID)
		if err != nil {
			blog.Error("fail to get task by ID(%s), err:%s", taskID, err.Error())
			return nil, err
		}

		taskGroup.Taskgroup[index] = task
	}
	taskGroup.ResourceVersion = v2Taskgroup.ResourceVersion
	return &taskGroup, nil
}
