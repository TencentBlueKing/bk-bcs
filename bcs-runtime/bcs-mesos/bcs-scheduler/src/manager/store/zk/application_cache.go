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
	"errors"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// at present, we just
type applicationCacheNode struct {
	Taskgroups []*types.TaskGroup
}

// Cache Manager
type cacheManager struct {
	// Current cached application
	Applications map[string]*applicationCacheNode
	// Manager currently is OK
	isOK    bool
	mapLock *sync.RWMutex
}

var cacheMgr *cacheManager

// Init cache manager, use cache or not
func (store *managerStore) InitCacheMgr(isUsed bool) error {

	//zk store always don't use cache
	isUsed = false
	if isUsed {
		blog.Infof("init cache begin: use cache for storage")
	} else {
		blog.Infof("init cache begin: not use cache for storage")
	}

	cacheMgr = new(cacheManager)
	cacheMgr.mapLock = new(sync.RWMutex)

	cacheMgr.mapLock.Lock()
	cacheMgr.Applications = make(map[string]*applicationCacheNode)
	cacheMgr.isOK = isUsed
	cacheMgr.mapLock.Unlock()

	blog.Infof("init cache end")

	return nil
}

// Unint cache manager
func (store *managerStore) UnInitCacheMgr() error {

	blog.Infof("uninit cache begin")

	if cacheMgr != nil {
		cacheMgr.mapLock.Lock()
		cacheMgr.Applications = nil
		cacheMgr.mapLock = nil
		cacheMgr.isOK = false
		cacheMgr.mapLock.Unlock()
	}
	cacheMgr = nil

	blog.Infof("uninit cache end")
	return nil
}

func deleteAppCacheNode(runAs, appID string) error {

	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	blog.Infof("delete app(%s) in cache", key)

	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Applications, key)
	cacheMgr.mapLock.Unlock()

	return nil
}

func taskgroupDeepCopy(src, dst *types.TaskGroup) {
	dataBytes, _ := json.Marshal(src)
	json.Unmarshal(dataBytes, dst)
}

func taskDeepCopy(src, dst *types.Task) {
	dataBytes, _ := json.Marshal(src)
	json.Unmarshal(dataBytes, dst)
}

func syncAppCacheNode(runAs, appID string, taskGroups []*types.TaskGroup) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	app := getCacheAppNode(runAs, appID)
	if app != nil {
		blog.Infof("app(%s) already in cache", key)
	} else {
		blog.Infof("app(%s) not in cache, to create", key)
		cacheMgr.mapLock.Lock()
		appNode := new(applicationCacheNode)
		cacheMgr.Applications[key] = appNode
		cacheMgr.mapLock.Unlock()
	}
	if app == nil {
		app = getCacheAppNode(runAs, appID)
	}
	if app == nil {
		blog.Warnf("app(%s) not in cache", key)
		return nil
	}

	blog.Infof("app(%s) sync taskgroups in cache", key)
	//set empty list
	app.Taskgroups = nil
	for _, taskGroup := range taskGroups {
		blog.Infof("app(%s) sync taskgroup(%s) in cache", key, taskGroup.ID)
		tmpData := new(types.TaskGroup)
		taskgroupDeepCopy(taskGroup, tmpData)
		app.Taskgroups = append(app.Taskgroups, tmpData)
	}

	return nil
}

func getCacheAppNode(runAs, appID string) *applicationCacheNode {

	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	cacheMgr.mapLock.RLock()
	app, ok := cacheMgr.Applications[key]
	cacheMgr.mapLock.RUnlock()
	if ok {
		//blog.V(3).Infof("app(%s.%s) is in cache", runAs, appID)
		return app
	}

	//blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
	return nil
}

func listCacheTaskGroups(runAs, appID string) ([]*types.TaskGroup, error) {
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil, nil
	}

	var taskgroupsList []*types.TaskGroup
	for _, taskGroup := range app.Taskgroups {
		blog.V(3).Infof("app(%s.%s) list taskgroups, get %s in cache", runAs, appID, taskGroup.ID)
		tmpData := new(types.TaskGroup)
		taskgroupDeepCopy(taskGroup, tmpData)
		taskgroupsList = append(taskgroupsList, tmpData)
	}

	return taskgroupsList, nil
}

func saveCacheTaskGroup(taskGroup *types.TaskGroup) error {
	app := getCacheAppNode(taskGroup.RunAs, taskGroup.AppID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", taskGroup.RunAs, taskGroup.AppID)
		return nil
	}

	tmpData := new(types.TaskGroup)
	taskgroupDeepCopy(taskGroup, tmpData)

	isExist := false
	for index, currTaskGroup := range app.Taskgroups {
		if currTaskGroup.ID == taskGroup.ID {
			blog.V(3).Infof("update taskgroup(%s) in cache", taskGroup.ID)
			app.Taskgroups[index] = tmpData
			isExist = true
		}
	}

	if !isExist {
		blog.Infof("insert taskgroup(%s) in cache", tmpData.ID)
		app.Taskgroups = append(app.Taskgroups, tmpData)
	}

	return nil
}

func fetchCacheTaskGroup(taskGroupID string) (*types.TaskGroup, error) {

	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil, nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("fetched taskgroup(%s) in cache", taskGroupID)
			tmpData := new(types.TaskGroup)
			taskgroupDeepCopy(taskGroup, tmpData)
			return tmpData, nil
		}
	}

	blog.Warnf("fetch taskgroup(%s) in cache return nil", taskGroupID)
	return nil, nil
}

func deleteCacheTaskGroup(taskGroupID string) error {

	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil
	}

	for index, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.Infof("delete taskgroup(%s) in cache", taskGroupID)
			app.Taskgroups = append(app.Taskgroups[:index], app.Taskgroups[index+1:]...)
		}
	}

	return nil
}

func saveCacheTask(task *types.Task) error {

	if !cacheMgr.isOK {
		return nil
	}

	taskGroupID := types.GetTaskGroupID(task.ID)
	if taskGroupID == "" {
		return errors.New("task id error")
	}
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("taskgroup(%s) in cache", taskGroupID)

			cacheData := new(types.Task)
			taskDeepCopy(task, cacheData)
			isExist := false
			for index, cacheTask := range taskGroup.Taskgroup {
				if task.ID == cacheTask.ID {
					blog.V(3).Infof("update task(%s) in cache", task.ID)
					taskGroup.Taskgroup[index] = cacheData
					isExist = true
				}
			}

			if !isExist {
				blog.Warnf("insert task(%s) in cache", task.ID)
				taskGroup.Taskgroup = append(taskGroup.Taskgroup, cacheData)
			}
		}
	}

	return nil
}

func deleteCacheTask(taskId string) error {
	if !cacheMgr.isOK {
		return nil
	}

	taskGroupID := types.GetTaskGroupID(taskId)
	if taskGroupID == "" {
		return errors.New("task id error")
	}
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("taskgroup(%s) in cache", taskGroupID)
			for index, cacheTask := range taskGroup.Taskgroup {
				if taskId == cacheTask.ID {
					blog.Infof("delete task(%s) in cache", taskId)
					taskGroup.Taskgroup = append(taskGroup.Taskgroup[:index], taskGroup.Taskgroup[index+1:]...)
				}
			}
		}
	}
	return nil
}

func fetchCacheTask(taskId string) (*types.Task, error) {

	if !cacheMgr.isOK {
		return nil, nil
	}

	taskGroupID := types.GetTaskGroupID(taskId)
	if taskGroupID == "" {
		return nil, errors.New("task id error")
	}
	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil, nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("taskgroup(%s) in cache", taskGroupID)
			for _, cacheTask := range taskGroup.Taskgroup {
				if taskId == cacheTask.ID {
					blog.V(3).Infof("fetched task(%s) in cache", taskId)
					tmpData := new(types.Task)
					taskDeepCopy(cacheTask, tmpData)
					return tmpData, nil
				}
			}
		}
	}

	blog.Warnf("fetch task(%s) in cache return nil", taskId)

	return nil, nil
}
