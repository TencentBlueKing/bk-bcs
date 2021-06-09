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

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
	schStore "github.com/Tencent/bk-bcs/bcs-mesos/bcs-scheduler/src/manager/store"
	"github.com/Tencent/bk-bcs/bcs-mesos/kubebkbcsv2/apis/bkbcs/v2"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CheckDaemonsetExist check agent whether exist
func (store *managerStore) CheckDaemonsetExist(daemon *types.BcsDaemonset) (string, bool) {
	obj, _ := store.FetchDaemonset(daemon.NameSpace, daemon.Name)
	if obj != nil {
		return obj.ResourceVersion, true
	}

	return "", false
}

// SaveDaemonset save agent
func (store *managerStore) SaveDaemonset(daemon *types.BcsDaemonset) error {
	client := store.BkbcsClient.BcsDaemonsets(daemon.NameSpace)
	v2Daemonset := &v2.BcsDaemonset{
		TypeMeta: metav1.TypeMeta{
			Kind:       CrdBcsDaemonset,
			APIVersion: ApiversionV2,
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      daemon.Name,
			Namespace: daemon.NameSpace,
		},
		Spec: v2.BcsDaemonsetSpec{
			BcsDaemonset: *daemon,
		},
	}

	var err error
	rv, exist := store.CheckDaemonsetExist(daemon)
	//if exist, then update
	if exist {
		v2Daemonset.ResourceVersion = rv
		v2Daemonset, err = client.Update(context.Background(), v2Daemonset, metav1.UpdateOptions{})
		//else not exist, then create it
	} else {
		v2Daemonset, err = client.Create(context.Background(), v2Daemonset, metav1.CreateOptions{})
	}
	if err != nil {
		return err
	}

	//update kube-apiserver ResourceVersion
	daemon.ResourceVersion = v2Daemonset.ResourceVersion
	//save daemonset in cache
	saveCacheDaemonset(daemon)
	return nil
}

// FetchDaemonset fetch agent for agent InnerIP
func (store *managerStore) FetchDaemonset(ns, name string) (*types.BcsDaemonset, error) {
	//fetch agent in cache
	agent := getCacheDaemonset(ns, name)
	if agent == nil {
		return nil, schStore.ErrNoFound
	}
	return agent, nil
}

// ListAllDaemonset list all agent list
func (store *managerStore) ListAllDaemonset() ([]*types.BcsDaemonset, error) {
	if cacheMgr.isOK {
		return listCacheDaemonsets()
	}

	client := store.BkbcsClient.BcsDaemonsets("")
	v2Daemonsets, err := client.List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	daemonsets := make([]*types.BcsDaemonset, 0, len(v2Daemonsets.Items))
	for _, v2 := range v2Daemonsets.Items {
		obj := v2.Spec.BcsDaemonset
		obj.ResourceVersion = v2.ResourceVersion
		daemonsets = append(daemonsets, &obj)
	}
	return daemonsets, nil
}

// DeleteDaemonset delete daemonset for innerip
func (store *managerStore) DeleteDaemonset(ns, name string) error {
	client := store.BkbcsClient.BcsDaemonsets(ns)
	err := client.Delete(context.Background(), name, metav1.DeleteOptions{})
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	//delete daemonset in cache
	deleteCacheDaemonset(ns, name)
	return nil
}

// ListDaemonsetTaskGroups show us all the task group on line
func (store *managerStore) ListDaemonsetTaskGroups(namespace, name string) ([]*types.TaskGroup, error) {
	taskgroups := make([]*types.TaskGroup, 0)
	daemonset, err := store.FetchDaemonset(namespace, name)
	//if err!=nil, show application not found
	//then return empty
	if err != nil {
		return taskgroups, nil
	}

	for podId := range daemonset.Pods {
		taskgroup, err := store.FetchTaskGroup(podId)
		if err != nil {
			return nil, err
		}

		taskgroups = append(taskgroups, taskgroup)
	}
	return taskgroups, nil
}
