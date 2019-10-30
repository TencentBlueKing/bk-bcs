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
	"k8s.io/apimachinery/pkg/api/errors"
	"sync"

	"bk-bcs/bcs-common/common/blog"
	schStore "bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"bk-bcs/bcs-mesos/bcs-scheduler/src/types"
	"bk-bcs/bcs-mesos/pkg/apis/bkbcs/v2"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (store *managerStore) CheckApplicationExist(application *types.Application) (string, bool) {
	client := store.BkbcsClient.Applications(application.RunAs)
	v2App, err := client.Get(application.ID, metav1.GetOptions{})
	if err == nil {
		return v2App.ResourceVersion, true
	}

	return "", false
}

//SaveApplication save application data into db.
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
			Name:      application.ID,
			Namespace: application.RunAs,
		},
		Spec: v2.ApplicationSpec{
			Application: *application,
		},
	}

	rv, exist := store.CheckApplicationExist(application)
	if exist {
		v2Application.ResourceVersion = rv
		_, err = client.Update(v2Application)
	} else {
		_, err = client.Create(v2Application)
	}
	return err
}

func (store *managerStore) ListApplicationNodes(runAs string) ([]string, error) {
	client := store.BkbcsClient.Applications(runAs)
	v2App, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	nodes := make([]string, 0, len(v2App.Items))
	for _, app := range v2App.Items {
		nodes = append(nodes, app.Spec.ID)
	}

	return nodes, nil
}

//FetchApplication is used to fetch application by appID
func (store *managerStore) FetchApplication(runAs, appID string) (*types.Application, error) {
	client := store.BkbcsClient.Applications(runAs)
	v2App, err := client.Get(appID, metav1.GetOptions{})
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, schStore.ErrNoFound
		}
		return nil, err
	}

	return &v2App.Spec.Application, nil
}

//ListApplications is used to get all applications
func (store *managerStore) ListApplications(runAs string) ([]*types.Application, error) {
	client := store.BkbcsClient.Applications(runAs)
	v2Apps, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	apps := make([]*types.Application, 0, len(v2Apps.Items))
	for _, app := range v2Apps.Items {
		apps = append(apps, &app.Spec.Application)
	}
	return apps, nil
}

//DeleteApplication remove the application from db by appID
func (store *managerStore) DeleteApplication(runAs, appID string) error {
	client := store.BkbcsClient.Applications(runAs)
	err := client.Delete(appID, &metav1.DeleteOptions{})
	return err
}

func (store *managerStore) ListAllApplications() ([]*types.Application, error) {
	client := store.BkbcsClient.Applications("")
	v2Apps, err := client.List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	apps := make([]*types.Application, 0, len(v2Apps.Items))
	for _, app := range v2Apps.Items {
		apps = append(apps, &app.Spec.Application)
	}
	return apps, nil
}
