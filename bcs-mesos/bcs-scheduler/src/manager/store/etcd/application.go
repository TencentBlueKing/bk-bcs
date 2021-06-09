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
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var appLocks map[string]*sync.Mutex
var appRWlock sync.RWMutex

// InitLockPool init application pool
func (store *managerStore) InitLockPool() {
	if appLocks == nil {
		blog.Info("init application lock pool")
		appLocks = make(map[string]*sync.Mutex)
	}
}

// LockApplication lock application
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

// UnLockApplication unlock application
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

// CheckApplicationExist check if an application exists
func (store *managerStore) CheckApplicationExist(application *types.Application) (string, bool) {
	app, _ := store.fetchApplicationInDB(application.RunAs, application.ID)
	if app != nil {
		return app.ResourceVersion, true
	}

	return "", false
}

// SaveApplication save application data into db.
func (store *managerStore) SaveApplication(application *types.Application) error {
	err := store.checkNamespace(application.RunAs)
	if err != nil {
		return err
	}

	client := store.BkbcsClient.Applications(application.RunAs)
	v2Application := &v2.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdApplication,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        application.ID,
			Namespace:   application.RunAs,
			Labels:      store.filterSpecialLabels(application.ObjectMeta.Labels),
			Annotations: application.ObjectMeta.Annotations,
		},
		Spec: v2.ApplicationSpec{
			Application: *application,
		},
	}

	rv, exist := store.CheckApplicationExist(application)
	if exist {
		v2Application.ResourceVersion = rv
		v2Application, err = client.Update(context.Background(), v2Application, metav1.UpdateOptions{})
	} else {
		v2Application, err = client.Create(context.Background(), v2Application, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	application.ResourceVersion = v2Application.ResourceVersion
	saveCacheApplication(application.RunAs, application.ID, application)
	return nil
}

// ListApplicationNodes list application nodes
func (store *managerStore) ListApplicationNodes(runAs string) ([]string, error) {
	apps, err := store.ListApplications(runAs)
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(apps))
	for _, app := range apps {
		nodes = append(nodes, app.ID)
	}

	return nodes, nil
}

// FetchApplication is used to fetch application by appID
func (store *managerStore) FetchApplication(runAs, appID string) (*types.Application, error) {
	cacheApp, _ := getCacheApplication(runAs, appID)
	if cacheApp == nil {
		return nil, schStore.ErrNoFound
	}

	return cacheApp, nil
}

func (store *managerStore) fetchApplicationInDB(runAs, appID string) (*types.Application, error) {
	client := store.BkbcsClient.Applications(runAs)
	v2App, err := client.Get(context.Background(), appID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	app := &v2App.Spec.Application
	app.ResourceVersion = v2App.ResourceVersion
	return app, nil
}

// ListApplications is used to get all applications
func (store *managerStore) ListApplications(runAs string) ([]*types.Application, error) {
	if cacheMgr.isOK {
		return listCacheRunAsApplications(runAs)
	}

	client := store.BkbcsClient.Applications(runAs)
	v2Apps, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	apps := make([]*types.Application, 0, len(v2Apps.Items))
	for _, app := range v2Apps.Items {
		obj := app.Spec.Application
		obj.ResourceVersion = app.ResourceVersion
		apps = append(apps, &obj)
	}
	return apps, nil
}

// DeleteApplication remove the application from db by appID
func (store *managerStore) DeleteApplication(runAs, appID string) error {
	client := store.BkbcsClient.Applications(runAs)
	err := client.Delete(context.Background(), appID, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	deleteAppCacheNode(runAs, appID)
	return nil
}

// ListAllApplications list all applications
func (store *managerStore) ListAllApplications() ([]*types.Application, error) {
	if cacheMgr.isOK {
		return listCacheApplications()
	}

	client := store.BkbcsClient.Applications("")
	v2Apps, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	apps := make([]*types.Application, 0, len(v2Apps.Items))
	for _, app := range v2Apps.Items {
		obj := app.Spec.Application
		obj.ResourceVersion = app.ResourceVersion
		apps = append(apps, &obj)
	}
	return apps, nil
}

// ListTaskGroups show us all the task group on line
func (store *managerStore) ListTaskGroups(runAs, appID string) ([]*types.TaskGroup, error) {
	taskgroups := make([]*types.TaskGroup, 0)
	app, err := store.FetchApplication(runAs, appID)
	// if err!=nil, show application not found
	// then return empty
	if err != nil {
		return taskgroups, nil
	}

	for _, podId := range app.Pods {
		taskgroup, err := store.FetchTaskGroup(podId.Name)
		if err != nil {
			return nil, err
		}

		taskgroups = append(taskgroups, taskgroup)
	}
	return taskgroups, nil
}
