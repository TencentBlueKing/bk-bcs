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
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"

	"github.com/samuel/go-zookeeper/zk"
)

var appLocks map[string]*sync.Mutex
var appRWlock sync.RWMutex

func (store *managerStore) InitLockPool() {
	if appLocks == nil {
		blog.Info("init application lock pool")
		appLocks = make(map[string]*sync.Mutex)
	}
}

func (store *managerStore) LockApplication(appID string) {

	appRWlock.RLock()
	myLock, ok := appLocks[appID]
	appRWlock.RUnlock()
	if ok {
		myLock.Lock()
		return
	}

	appRWlock.Lock()
	myLock, ok = appLocks[appID]
	if !ok {
		blog.Info("create application lock(%s), current locknum(%d)", appID, len(appLocks))
		appLocks[appID] = new(sync.Mutex)
		myLock, _ = appLocks[appID]
	}
	appRWlock.Unlock()

	myLock.Lock()
	return
}

func (store *managerStore) UnLockApplication(appID string) {
	appRWlock.RLock()
	myLock, ok := appLocks[appID]
	appRWlock.RUnlock()

	if !ok {
		blog.Error("application lock(%s) not exist when do unlock", appID)
		return
	}
	myLock.Unlock()
}

func getApplicationRootPath() string {
	return "/" + bcsRootNode + "/" + applicationNode + "/"
}

//SaveApplication save application data into db.
func (store *managerStore) SaveApplication(application *types.Application) error {

	data, err := json.Marshal(application)
	if err != nil {
		return err
	}

	path := getApplicationRootPath() + application.RunAs + "/" + application.ID
	return store.Db.Insert(path, string(data))
}

func (store *managerStore) ListRunAs() ([]string, error) {

	//rootPath := getApplicationRootPath()
	rootPath := "/" + bcsRootNode + "/" + applicationNode

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

func (store *managerStore) ListApplicationNodes(runAs string) ([]string, error) {

	path := getApplicationRootPath() + runAs

	appIDs, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list path:%s, err:%s", path, err.Error())
		return nil, err
	}

	return appIDs, nil
}

//FetchApplication is used to fetch application by appID
func (store *managerStore) FetchApplication(runAs, appID string) (*types.Application, error) {

	path := getApplicationRootPath() + runAs + "/" + appID

	data, err := store.Db.Fetch(path)
	if err != nil {
		if err == zk.ErrNoNode {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	app := &types.Application{}
	if err := json.Unmarshal(data, app); err != nil {
		blog.Error("fail to unmarshal application(%s:%s) data(%s) err:%s", runAs, appID, string(data), err.Error())
		return nil, err
	}

	return app, nil
}

//ListApplications is used to get all applications
func (store *managerStore) ListApplications(runAs string) ([]*types.Application, error) {

	path := getApplicationRootPath() + runAs //defaultRunAs

	appIDs, err := store.Db.List(path)
	if err != nil {
		blog.Error("fail to list application ids, err:%s", err.Error())
		return nil, err
	}

	if nil == appIDs {
		blog.V(3).Infof("no application in (%s)", runAs)
		return nil, nil
	}

	var apps []*types.Application

	for _, appID := range appIDs {
		app, err := store.FetchApplication(runAs, appID)
		if err != nil {
			blog.Error("fail to fetch application by appID(%s:%s)", runAs, appID)
			continue
		}

		apps = append(apps, app)
	}

	return apps, nil
}

//DeleteApplication remove the application from db by appID
func (store *managerStore) DeleteApplication(runAs, appID string) error {

	path := getApplicationRootPath() + runAs + "/" + appID
	blog.V(3).Infof("will delete applcation,path(%s)", path)

	if err := store.Db.Delete(path); err != nil {
		blog.Error("fail to delete application, application id(%s), err:%s", appID, err.Error())
		return err
	}
	deleteAppCacheNode(runAs, appID)

	return nil
}

func (store *managerStore) ListAllApplications() ([]*types.Application, error) {
	nss, err := store.ListObjectNamespaces(applicationNode)
	if err != nil {
		return nil, err
	}

	var objs []*types.Application
	for _, ns := range nss {
		obj, err := store.ListApplications(ns)
		if err != nil {
			blog.Error("fail to fetch application by ns(%s)", ns)
			continue
		}

		objs = append(objs, obj...)
	}

	return objs, nil
}

/*
func (store *managerStore) CleanApplication(runAs, appId string) error {

	blog.V(3).Infof("destroy application: runAs(%s), appId(%s)", runAs, appId)

	groupIDs, err := store.ListTaskGroupNodes(runAs, appId)
	if err != nil {
		blog.Error("fail to list taskGroups (%s %s), err:%s", runAs, appId, err.Error())
		return err
	}

	if nil != groupIDs {
		for _, groupID := range groupIDs {
			err = store.CleanTaskGroup(runAs, appId, groupID)
			if err != nil {
				blog.Error("CleanTaskGroup(%s %s %s) err: %s", runAs, appId, groupID, err.Error())
				continue
			} else {
				blog.V(3).Infof("CleanTaskGroup(%s %s %s) succ", runAs, appId, groupID)
			}
		}
	}

	versions, err := store.ListVersions(runAs, appId)
	if err != nil {
		blog.Error("ListVersions(%s %s) err:%s", runAs, appId, err.Error())
		return err
	}

	if versions != nil {
		for _, version := range versions {
			if err = store.DeleteVersion(runAs, appId, version); err != nil {
				blog.Error("DeleteVersion(%s %s %s) err:%s", runAs, appId, version, err.Error())
				continue
			}
		}
	}

	if err = store.DeleteVersionNode(runAs, appId); err != nil {
		blog.Error("delete version node(%s %s) err:%s", runAs, appId, err.Error())
	}

	err = store.DeleteApplication(runAs, appId)
	if err != nil {
		blog.Error("DeleteApplication(%s %s) err:%s", runAs, appId, err.Error())
		return err
	}

	return nil
}
*/
