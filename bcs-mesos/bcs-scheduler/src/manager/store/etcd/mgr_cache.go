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
	"errors"
	"fmt"
	"sync"

	"bk-bcs/bcs-common/common/blog"
	commtypes "bk-bcs/bcs-common/common/types"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
)

// at present, we just
type applicationCacheNode struct {
	Taskgroups  []*types.TaskGroup
	Application *types.Application
	Versions    []*types.Version
}

// Cache Manager
type cacheManager struct {
	// Current cached application
	Applications map[string]*applicationCacheNode
	Namespaces   map[string]struct{}
	Configmaps   map[string]*commtypes.BcsConfigMap
	Secrets      map[string]*commtypes.BcsSecret
	// Manager currently is OK
	isOK    bool
	mapLock *sync.RWMutex
}

var cacheMgr *cacheManager

// Init cache manager, use cache or not
func (store *managerStore) InitCacheMgr(isUsed bool) error {

	if isUsed {
		blog.Infof("init cache begin: use cache for storage")
	} else {
		blog.Infof("init cache begin: not use cache for storage")
	}

	cacheMgr = new(cacheManager)
	cacheMgr.mapLock = new(sync.RWMutex)

	cacheMgr.mapLock.Lock()
	cacheMgr.Applications = make(map[string]*applicationCacheNode)
	cacheMgr.Namespaces = make(map[string]struct{})
	cacheMgr.Configmaps = make(map[string]*commtypes.BcsConfigMap)
	cacheMgr.Secrets = make(map[string]*commtypes.BcsSecret)
	if isUsed {
		// init namespace in cache
		err := store.initCacheNamespaces()
		if err != nil {
			return err
		}
		//init application in cache
		err = store.initCacheApplications()
		if err != nil {
			return err
		}
		//init configmap in cache
		err = store.initCacheConfigmaps()
		if err != nil {
			return err
		}
		//init secret in cache
		err = store.initCacheSecrets()
		if err != nil {
			return err
		}
	}
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
		cacheMgr.Namespaces = nil
		cacheMgr.Configmaps = nil
		cacheMgr.Secrets = nil
		cacheMgr.mapLock = nil
		cacheMgr.isOK = false
		cacheMgr.mapLock.Unlock()
	}
	cacheMgr = nil

	blog.Infof("uninit cache end")
	return nil
}

func (store *managerStore) initCacheNamespaces() error {
	nss, err := store.ListRunAs()
	if err != nil {
		blog.Errorf("cacheManager init namespace failed: %s", err.Error())
		return err
	}

	for _, ns := range nss {
		cacheMgr.Namespaces[ns] = struct{}{}
		blog.V(3).Infof("cacheManager sync namespace %s in cache", ns)
	}
	blog.Infof("cacheMgr init cache namespaces done")
	return nil
}

func (store *managerStore) initCacheConfigmaps() error {
	cfgs, err := store.ListAllConfigmaps()
	if err != nil {
		blog.Errorf("cacheManager init configmaps failed: %s", err.Error())
		return err
	}

	for _, cfg := range cfgs {
		key := fmt.Sprintf("%s.%s", cfg.NameSpace, cfg.Name)
		cacheMgr.Configmaps[key] = cfg.DeepCopy()
		blog.V(3).Infof("cacheManager sync configmap %s in cache", key)
	}
	blog.Infof("cacheMgr init cache configmaps done")
	return nil
}

func (store *managerStore) initCacheSecrets() error {
	secs, err := store.ListAllSecrets()
	if err != nil {
		blog.Errorf("cacheManager init secrets failed: %s", err.Error())
		return err
	}

	for _, sec := range secs {
		key := fmt.Sprintf("%s.%s", sec.NameSpace, sec.Name)
		cacheMgr.Secrets[key] = sec.DeepCopy()
		blog.V(3).Infof("cacheManager sync secret %s in cache", key)
	}
	blog.Infof("cacheMgr init cache secrets done")
	return nil
}

func (store *managerStore) initCacheApplications() error {
	apps, err := store.ListAllApplications()
	if err != nil {
		blog.Errorf("cacheManager init application failed: %s", err.Error())
		return err
	}

	for _, app := range apps {
		appNode := new(applicationCacheNode)
		//cache application
		appNode.Application = app.DeepCopy()
		blog.V(3).Infof("cacheManager app(%s:%s) sync Application in cache", app.RunAs, app.ID)

		//cache versions
		versions, err := store.listVersions(app.RunAs, app.ID)
		if err != nil {
			return err
		}
		for _, version := range versions {
			blog.V(3).Infof("cacheManager app(%s:%s) sync version(%s) in cache", app.RunAs, app.ID, version.Name)
			tmpData := version.DeepCopy()
			appNode.Versions = append(appNode.Versions, tmpData)
		}

		//cache taskgroups
		taskGroups, err := store.listTaskgroupsInDB(app.RunAs, app.ID)
		if err != nil {
			return err
		}
		for _, taskGroup := range taskGroups {
			blog.V(3).Infof("cacheManager app(%s:%s) sync taskgroup(%s) in cache", app.RunAs, app.ID, taskGroup.ID)
			tmpData := taskGroup.DeepCopy()
			appNode.Taskgroups = append(appNode.Taskgroups, tmpData)
		}

		key := app.RunAs + "." + app.ID
		cacheMgr.Applications[key] = appNode
	}
	blog.Infof("cacheMgr init cache applications done")
	return nil
}

func checkCacheNamespaceExist(ns string) bool {
	if !cacheMgr.isOK {
		return false
	}

	cacheMgr.mapLock.RLock()
	_, ok := cacheMgr.Namespaces[ns]
	cacheMgr.mapLock.RUnlock()
	return ok
}

func syncCacheNamespace(ns string) {
	if !cacheMgr.isOK {
		return
	}

	cacheMgr.mapLock.Lock()
	cacheMgr.Namespaces[ns] = struct{}{}
	cacheMgr.mapLock.Unlock()
}

func deleteAppCacheNode(runAs, appID string) error {

	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	blog.V(3).Infof("delete app(%s) in cache", key)

	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Applications, key)
	cacheMgr.mapLock.Unlock()

	return nil
}

func saveCacheConfigmap(obj *commtypes.BcsConfigMap) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := fmt.Sprintf("%s.%s", obj.NameSpace, obj.Name)
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Configmaps[key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

func getCacheConfigmap(ns, name string) *commtypes.BcsConfigMap {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Configmaps[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

func deleteCacheConfigmap(ns, name string) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Configmaps, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

func saveCacheSecret(obj *commtypes.BcsSecret) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := fmt.Sprintf("%s.%s", obj.NameSpace, obj.Name)
	tmpData := obj.DeepCopy()
	cacheMgr.mapLock.Lock()
	cacheMgr.Secrets[key] = tmpData
	cacheMgr.mapLock.Unlock()
	return nil
}

func getCacheSecret(ns, name string) *commtypes.BcsSecret {
	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.RLock()
	obj, ok := cacheMgr.Secrets[key]
	cacheMgr.mapLock.RUnlock()
	if !ok {
		return nil
	}

	return obj.DeepCopy()
}

func deleteCacheSecret(ns, name string) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := fmt.Sprintf("%s.%s", ns, name)
	cacheMgr.mapLock.Lock()
	delete(cacheMgr.Secrets, key)
	cacheMgr.mapLock.Unlock()
	return nil
}

func saveCacheVersion(runAs, appID string, obj *types.Version) error {
	if !cacheMgr.isOK {
		return nil
	}

	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil
	}

	tmpData := obj.DeepCopy()
	app.Versions = append(app.Versions, tmpData)
	return nil
}

func getCacheVersion(runAs, versionId, versionNo string) (*types.Version, error) {
	app := getCacheAppNode(runAs, versionId)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, versionId)
		return nil, nil
	}

	for _, version := range app.Versions {
		if version.Name == versionNo {
			return version.DeepCopy(), nil
		}
	}

	return nil, nil
}

func listCacheVersions(runAs, versionId string) ([]*types.Version, error) {
	app := getCacheAppNode(runAs, versionId)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, versionId)
		return nil, nil
	}

	var versions []*types.Version
	for _, version := range app.Versions {
		tmpData := version.DeepCopy()
		versions = append(versions, tmpData)
	}

	return versions, nil
}

func saveCacheApplication(runAs, appID string, obj *types.Application) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	app := getCacheAppNode(runAs, appID)
	if app != nil {
		blog.V(3).Infof("app(%s) already in cache", key)
	} else {
		blog.V(3).Infof("app(%s) not in cache, to create", key)
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

	app.Application = obj.DeepCopy()
	return nil
}

func getCacheApplication(runAs, appID string) (*types.Application, error) {
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil, nil
	}
	if app.Application == nil {
		blog.V(3).Infof("app(%s.%s) Application not in cache", runAs, appID)
		return nil, nil
	}

	return app.Application.DeepCopy(), nil
}

func syncAppCacheNode(runAs, appID string, taskGroups []*types.TaskGroup) error {
	if !cacheMgr.isOK {
		return nil
	}

	key := runAs + "." + appID
	app := getCacheAppNode(runAs, appID)
	if app != nil {
		blog.V(3).Infof("app(%s) already in cache", key)
	} else {
		blog.V(3).Infof("app(%s) not in cache, to create", key)
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

	//set empty list
	app.Taskgroups = nil
	for _, taskGroup := range taskGroups {
		blog.V(3).Infof("app(%s) sync taskgroup(%s) in cache", key, taskGroup.ID)
		tmpData := taskGroup.DeepCopy()
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
		return app
	}

	return nil
}

func listCacheTaskGroups(runAs, appID string) ([]*types.TaskGroup, error) {
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil, nil
	}
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
		return nil, nil
	}

	var taskgroupsList []*types.TaskGroup
	for _, taskGroup := range app.Taskgroups {
		blog.V(3).Infof("app(%s.%s) list taskgroups, get %s in cache", runAs, appID, taskGroup.ID)
		tmpData := taskGroup.DeepCopy()
		taskgroupsList = append(taskgroupsList, tmpData)
	}

	return taskgroupsList, nil
}

func saveCacheTaskGroup(taskGroup *types.TaskGroup) error {
	if !cacheMgr.isOK {
		return nil
	}

	app := getCacheAppNode(taskGroup.RunAs, taskGroup.AppID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", taskGroup.RunAs, taskGroup.AppID)
		return nil
	}

	tmpData := taskGroup.DeepCopy()
	if app.Taskgroups == nil {
		blog.V(3).Infof("insert taskgroup(%s) in cache", tmpData.ID)
		app.Taskgroups = append(app.Taskgroups, tmpData)
		return nil
	}

	isExist := false
	for index, currTaskGroup := range app.Taskgroups {
		if currTaskGroup.ID == taskGroup.ID {
			blog.V(3).Infof("update taskgroup(%s) in cache", taskGroup.ID)
			app.Taskgroups[index] = tmpData
			isExist = true
		}
	}

	if !isExist {
		blog.V(3).Infof("insert taskgroup(%s) in cache", tmpData.ID)
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
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
		return nil, nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("fetched taskgroup(%s) in cache", taskGroupID)
			tmpData := taskGroup.DeepCopy()
			return tmpData, nil
		}
	}

	blog.V(3).Infof("fetch taskgroup(%s) not in cache return nil", taskGroupID)
	return nil, nil
}

func deleteCacheTaskGroup(taskGroupID string) error {
	if !cacheMgr.isOK {
		return nil
	}

	runAs, appID := types.GetRunAsAndAppIDbyTaskGroupID(taskGroupID)
	app := getCacheAppNode(runAs, appID)
	if app == nil {
		blog.V(3).Infof("app(%s.%s) not in cache", runAs, appID)
		return nil
	}
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
		return nil
	}

	for index, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("delete taskgroup(%s) in cache", taskGroupID)
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
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
		return nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("taskgroup(%s) in cache", taskGroupID)

			cacheData := task.DeepCopy()
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
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
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
	if app.Taskgroups == nil {
		blog.V(3).Infof("app(%s.%s) taskgroups not in cache", runAs, appID)
		return nil, nil
	}

	for _, taskGroup := range app.Taskgroups {
		if taskGroup.ID == taskGroupID {
			blog.V(3).Infof("taskgroup(%s) in cache", taskGroupID)
			for _, cacheTask := range taskGroup.Taskgroup {
				if taskId == cacheTask.ID {
					blog.V(3).Infof("fetched task(%s) in cache", taskId)
					tmpData := cacheTask.DeepCopy()
					return tmpData, nil
				}
			}
		}
	}
	blog.Warnf("fetch task(%s) in cache return nil", taskId)
	return nil, nil
}
