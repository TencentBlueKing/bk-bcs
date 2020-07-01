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

package app

import (
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/zkclient"

	"github.com/samuel/go-zookeeper/zk"
)

//ZkClient interface to define zk operation
//interface is only use for dependency injection
type ZkClient interface {
	Get(path string) ([]byte, *zk.Stat, error)
	GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error)
	Children(path string) ([]string, *zk.Stat, error)
	ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error)
	Exists(path string) (bool, *zk.Stat, error)
	ExistsW(path string) (bool, *zk.Stat, <-chan zk.Event, error)
}

//AdapterZkClient to adapt common zk interface and go-zookeeper zk
type AdapterZkClient struct {
	commZKClient *zkclient.ZkClient
}

//NewAdapterZkClient to new a AdapterZkClient and connect zk
func NewAdapterZkClient(host []string, sessionTimeOut time.Duration) (*AdapterZkClient, error) {
	azc := &AdapterZkClient{
		commZKClient: zkclient.NewZkClient(host),
	}
	err := azc.commZKClient.ConnectEx(sessionTimeOut)
	return azc, err
}

//Get to get a node value
func (azc *AdapterZkClient) Get(path string) ([]byte, *zk.Stat, error) {
	data, err := azc.commZKClient.Get(path)
	return []byte(data), nil, err
}

//GetW to get a node value and watch its change
func (azc *AdapterZkClient) GetW(path string) ([]byte, *zk.Stat, <-chan zk.Event, error) {
	return azc.commZKClient.GetW(path)
}

//Children to get path children node
func (azc *AdapterZkClient) Children(path string) ([]string, *zk.Stat, error) {
	data, err := azc.commZKClient.GetChildren(path)
	return data, nil, err
}

//ChildrenW to get path children node and watch their change
func (azc *AdapterZkClient) ChildrenW(path string) ([]string, *zk.Stat, <-chan zk.Event, error) {
	return azc.commZKClient.ChildrenW(path)
}

//Exists to judge node exist or not
func (azc *AdapterZkClient) Exists(path string) (bool, *zk.Stat, error) {
	exist, err := azc.commZKClient.Exist(path)
	return exist, nil, err
}

//ExistsW to judge node exist or not and watch it
func (azc *AdapterZkClient) ExistsW(path string) (bool, *zk.Stat, <-chan zk.Event, error) {
	return azc.commZKClient.ExistW(path)
}
