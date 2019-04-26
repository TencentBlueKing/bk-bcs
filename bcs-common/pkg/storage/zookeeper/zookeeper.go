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

package zookeeper

import (
	"bk-bcs/bcs-common/common/blog"
	"bk-bcs/bcs-common/common/zkclient"
	"bk-bcs/bcs-common/pkg/storage"
	"strings"
	"time"
)

//ConLocker wrapper lock for release connection
type ConLocker struct {
	path string           //path for lock
	lock *zkclient.ZkLock //bcs common lock
}

//Lock try to Lock
func (cl *ConLocker) Lock() error {
	return cl.lock.LockEx(cl.path, time.Second*3)
}

//Unlock release lock and connection
func (cl *ConLocker) Unlock() error {
	cl.lock.UnLock()
	return nil
}

//NewStorage create etcd storage
func NewStorage(hosts string) storage.Storage {
	//create zookeeper connection
	host := strings.Split(hosts, ",")
	blog.Info("Storage create zookeeper connection with %s", hosts)
	// conn, _, conErr := zk.Connect(host, time.Second*5)
	// if conErr != nil {
	// 	blog.Error("Storage create zookeeper connection failed: %v", conErr)
	// 	return nil
	// }
	bcsClient := zkclient.NewZkClient(host)
	if conErr := bcsClient.ConnectEx(time.Second * 5); conErr != nil {
		blog.Errorf("Storage create zookeeper connection failed: %v", conErr)
		return nil
	}

	blog.Infof("Storage connect to zookeeper %s success", hosts)
	s := &zkStorage{
		zkHost:   host,
		zkClient: bcsClient,
	}
	return s
}

//eStorage storage data in etcd
type zkStorage struct {
	zkHost   []string           //zookeeper host info, for reconnection
	zkClient *zkclient.ZkClient //zookeeper client for operation
}

// Stop stop implementation
func (zks *zkStorage) Stop() {
	zks.zkClient.Close()
}

// GetLocker implementation
func (zks *zkStorage) GetLocker(key string) (storage.Locker, error) {
	blog.Infof("zkStorage create %s locker", key)
	bcsLock := zkclient.NewZkLock(zks.zkHost)
	wrap := &ConLocker{
		path: key,
		lock: bcsLock,
	}
	return wrap, nil
}

//Register register self node
func (zks *zkStorage) Register(path string, data []byte) error {
	err := zks.zkClient.CreateEphAndSeq(path, data)
	if err != nil {
		blog.Errorf("zkStorage register %s failed, %v", path, err)
		return err
	}
	return nil
}

//Add add data
func (zks *zkStorage) Add(key string, value []byte) error {
	err := zks.zkClient.Create(key, value)
	if err != nil {
		blog.Errorf("zkStorage add %s with value %s err, %v", key, string(value), err)
		return err
	}
	return nil
}

//Delete delete node by key
func (zks *zkStorage) Delete(key string) ([]byte, error) {
	//done(developerJim): get data before delete
	data, err := zks.Get(key)
	if err != nil {
		return []byte(""), err
	}
	err = zks.zkClient.Del(key, -1)
	return data, err
}

//Update update node by value
func (zks *zkStorage) Update(key string, value []byte) error {
	return zks.zkClient.Set(key, string(value), -1)
}

//Get get data of path
func (zks *zkStorage) Get(key string) ([]byte, error) {
	data, err := zks.zkClient.Get(key)
	if err != nil {
		return nil, err
	}
	return []byte(data), nil
}

//List all children nodes
func (zks *zkStorage) List(key string) ([]string, error) {
	list, err := zks.zkClient.GetChildren(key)
	if err != nil {
		return nil, err
	}

	return list, nil
}

//Exist check path exist
func (zks *zkStorage) Exist(key string) (bool, error) {
	e, err := zks.zkClient.Exist(key)
	if err != nil {
		return false, err
	}
	return e, nil
}

func (zks *zkStorage) CreateDeepNode(key string, value []byte) error {
	return zks.zkClient.CreateDeepNode(key, value)
}
