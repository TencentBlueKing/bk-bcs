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

package store

import (
	"bk-bcs/bcs-common/common/zkclient"
)

//dbZk is a struct of the zookeeper client
type dbZk struct {
	ZkHost []string
	ZkCli  *zkclient.ZkClient
}

//NewDbZk create a dbZk object
func NewDbZk(host []string) Dbdrvier {
	zk := dbZk{
		ZkHost: host[:],
		ZkCli:  zkclient.NewZkClient(host),
	}

	return &zk
}

func (zk *dbZk) Connect() error {
	return zk.ZkCli.Connect()
}

func (zk *dbZk) Close() {
	zk.ZkCli.Close()
}

func (zk *dbZk) Insert(path string, value string) error {

	//path := "/registry/" + obj.object + "/" + obj.namespace + "/" + obj.id

	return zk.ZkCli.Update(path, value)
}

func (zk *dbZk) Fetch(path string) ([]byte, error) {

	//path := "/registry/" + obj.object + "/" + obj.namespace + "/" + obj.id

	data, err := zk.ZkCli.Get(path)

	return []byte(data), err
}

func (zk *dbZk) Update(path string, value string) error {

	//path := "/registry/" + obj.object + "/" + obj.namespace + "/" + obj.id

	return zk.ZkCli.Update(path, value)
}

func (zk *dbZk) Delete(path string) error {

	//path := "/registry/" + obj.object + "/" + obj.namespace + "/" + obj.id

	return zk.ZkCli.Del(path, -1)
}

func (zk *dbZk) List(path string) ([]string, error) {
	//path := "/registry/" + obj.object + "/" + obj.namespace
	b, _ := zk.ZkCli.Exist(path)
	if !b {
		return nil, nil
	}

	return zk.ZkCli.GetChildren(path)
}
