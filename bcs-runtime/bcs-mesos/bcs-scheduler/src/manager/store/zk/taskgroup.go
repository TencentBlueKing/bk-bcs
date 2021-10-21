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
	"fmt"
	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	"sort"
	"strconv"
	"strings"
)

func getTaskGroupRootPath() string {
	return "/" + bcsRootNode + "/" + applicationNode + "/"
}

func createTaskGroupPath(taskGroupId string) (string, error) {
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupId)
	if "" == runAs {
		err := fmt.Errorf("fail to get runAs from taskgroupID(%s)", taskGroupId)
		return "", err
	}

	if "" == appID {
		err := fmt.Errorf("fail to get appID from taskgroupID(%s)", taskGroupId)
		return "", err
	}

	path := getTaskGroupRootPath() + runAs + "/" + appID + "/" + taskGroupId

	return path, nil
}

// SaveTaskGroup save task group to store
func (store *managerStore) SaveTaskGroup(taskGroup *types.TaskGroup) error {

	blog.V(3).Infof("save task group(id:%s)", taskGroup.ID)

	data, err := json.Marshal(taskGroup)
	if err != nil {
		blog.Error("fail to encode object taskgroup(ID:%s) by json. err:%s", taskGroup.ID, err.Error())
		return err
	}

	path, err := createTaskGroupPath(taskGroup.ID)
	if err != nil {
		blog.Error("fail to create task group path. err(%s)", err.Error())
		return err
	}

	if err := store.Db.Insert(path, string(data)); err != nil {
		blog.Error("fail to save task group(id:%s) into db. err:%s", taskGroup.ID, err.Error())
		return err
	}

	saveCacheTaskGroup(taskGroup)

	// save task
	if taskGroup.Taskgroup != nil {
		for _, task := range taskGroup.Taskgroup {
			store.SaveTask(task)
		}
	}

	return nil
}

// ListTaskGroups show us all the task group on line
func (store *managerStore) ListTaskGroups(runAs, appID string) ([]*types.TaskGroup, error) {

	cacheList, cacheErr := listCacheTaskGroups(runAs, appID)
	if cacheErr == nil && cacheList != nil {
		return cacheList, cacheErr
	}

	path := getTaskGroupRootPath() + runAs + "/" + appID
	taskGroupIds, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to get taskgroup ids by path(%s), err:%s", path, err.Error())
		return nil, err
	}

	var taskgroupsList []*types.TaskGroup

	for _, taskGroupID := range taskGroupIds {
		taskgroup, err := store.FetchTaskGroup(taskGroupID)
		if err != nil {
			blog.Error("fail to get taskgroup(ID:%s), err:%s", taskGroupID, err.Error())
			return nil, err
		}

		taskgroupsList = append(taskgroupsList, taskgroup)
	}

	syncAppCacheNode(runAs, appID, taskgroupsList)

	return taskgroupsList, nil
}

// taskSorter bia name of []TaskGroup
type taskSorter []*types.TaskGroup

func (s taskSorter) Len() int      { return len(s) }
func (s taskSorter) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s taskSorter) Less(i, j int) bool {

	// for index of taskgroup, example for 0,1,2,3...
	a, _ := strconv.Atoi(strings.Split(s[i].ID, ".")[0])
	b, _ := strconv.Atoi(strings.Split(s[j].ID, ".")[0])

	return a < b
}

func (store *managerStore) FetchTaskgroupByIndex(runAs, appID string, index int) (*types.TaskGroup, error) {
	taskgroups, err := store.ListTaskGroups(runAs, appID)
	if err != nil {
		return nil, err
	}

	if index >= len(taskgroups) {
		return nil, fmt.Errorf("Not found index %d of taskgroup in application %s.%s", index, runAs, appID)
	}

	sort.Sort(taskSorter(taskgroups))

	return taskgroups[index], nil
}

func (store *managerStore) FetchTaskgroupByName(runAs, appID string, name string) (*types.TaskGroup, error) {
	taskgroups, err := store.ListTaskGroups(runAs, appID)
	if err != nil {
		return nil, err
	}

	var taskgroup *types.TaskGroup

	for _, v := range taskgroups {
		if v.Name == name {
			return taskgroup, nil
		}
	}

	return nil, fmt.Errorf("Not found taskgroup %s", name)
}

// DeleteTaskGroup delete a task group with executor id is taskGroupID
func (store *managerStore) DeleteTaskGroup(taskGroupID string) error {

	path, err := createTaskGroupPath(taskGroupID)
	if err != nil {
		blog.Error("fail to create taskgroup path, err:%s", err.Error())
		return err
	}

	// list taskIDs
	taskIds, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to get taskid list by path(%s), err:%s", path, err.Error())
		return err
	}

	// delte tasks
	for _, taskID := range taskIds {
		if err := store.DeleteTask(taskID); err != nil {
			blog.Error("fail to delete task(ID:%s), err:%s", taskID, err.Error())
			return err
		}
	}

	// delete task group
	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete taskgroup(id:%s), err:%s", taskGroupID, err.Error())
		return err
	}

	deleteCacheTaskGroup(taskGroupID)

	return nil
}

// FetchTaskGroup fetch a types.TaskGroup
func (store *managerStore) FetchTaskGroup(taskGroupID string) (*types.TaskGroup, error) {

	cacheTaskgroup, cacheErr := fetchCacheTaskGroup(taskGroupID)
	if cacheErr == nil && cacheTaskgroup != nil {
		return cacheTaskgroup, cacheErr
	}

	return store.FetchDBTaskGroup(taskGroupID)
}

// FetchTaskGroup fetch a types.TaskGroup
func (store *managerStore) FetchDBTaskGroup(taskGroupID string) (*types.TaskGroup, error) {

	path, err := createTaskGroupPath(taskGroupID)
	if err != nil {
		blog.Error("fail to create taskgroup path, err:%s", err.Error())
		return nil, err
	}

	data, err := store.Db.Fetch(path)
	if err != nil {
		// sometimes this is normal
		return nil, err
	}

	var taskGroup types.TaskGroup
	if err = json.Unmarshal(data, &taskGroup); err != nil {
		blog.Error("fail to unmarshal task group(ID:%s), err:%s", string(data), err.Error())
		return nil, err
	}

	// Fetch task,get taskids
	taskIds, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to get taskid list by path(%s), err:%s", path, err.Error())
		return nil, err
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

	return &taskGroup, nil
}

// list mesos cluster taskgroups, include: application、deployment...
func (store *managerStore) ListClusterTaskgroups() ([]*types.TaskGroup, error) {
	apps, err := store.ListAllApplications()
	if err != nil {
		return nil, err
	}

	taskgroups := make([]*types.TaskGroup, 0)
	for _, app := range apps {
		taskgs, err := store.ListTaskGroups(app.RunAs, app.ID)
		if err != nil {
			return nil, err
		}

		taskgroups = append(taskgroups, taskgs...)
	}
	return taskgroups, nil
}
