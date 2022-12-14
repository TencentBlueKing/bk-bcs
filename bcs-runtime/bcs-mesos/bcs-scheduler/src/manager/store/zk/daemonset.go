/*
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
	"fmt"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/scheduler/schetypes"
)

// SaveDaemonset xxx
// save agent
func (store *managerStore) SaveDaemonset(daemon *types.BcsDaemonset) error {
	return fmt.Errorf("zookeeper store don't support Daemonset, Please switch to the etcd store")
}

// FetchDaemonset xxx
// fetch agent for agent InnerIP
func (store *managerStore) FetchDaemonset(ns, name string) (*types.BcsDaemonset, error) {
	return nil, fmt.Errorf("zookeeper store don't support Daemonset, Please switch to the etcd store")
}

// ListAllDaemonset xxx
// list all agent list
func (store *managerStore) ListAllDaemonset() ([]*types.BcsDaemonset, error) {
	return nil, fmt.Errorf("zookeeper store don't support Daemonset, Please switch to the etcd store")
}

// DeleteDaemonset xxx
// delete daemonset for innerip
func (store *managerStore) DeleteDaemonset(ns, name string) error {
	return fmt.Errorf("zookeeper store don't support Daemonset, Please switch to the etcd store")
}

// ListDaemonsetTaskGroups xxx
// ListTaskGroups show us all the task group on line
func (store *managerStore) ListDaemonsetTaskGroups(namespace, name string) ([]*types.TaskGroup, error) {
	return nil, fmt.Errorf("zookeeper store don't support Daemonset, Please switch to the etcd store")
}
