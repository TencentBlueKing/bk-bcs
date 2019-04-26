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

package mesos

import (
	"bk-bcs/bcs-mesos/bcs-check/bcscheck/types"

	"github.com/samuel/go-zookeeper/zk"
)

type WatcherType string

const (
	/*WatcherType*/
	WatcherTypeTaskgroup WatcherType = "taskgroup"
)

//Watcher for watch data
type Watcher interface {
	Run()

	// watch health data
	DataW() <-chan *types.HealthSyncData

	Stop()
}

type ZkClient interface {
	GetEx(path string) ([]byte, *zk.Stat, error)
	GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error)
	GetChildrenEx(path string) ([]string, *zk.Stat, error)
	ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error)
	Exist(path string) (bool, error)
}
