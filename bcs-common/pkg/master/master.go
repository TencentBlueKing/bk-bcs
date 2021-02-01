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

package master

import bcstypes "github.com/Tencent/bk-bcs/bcs-common/common/types"

//Master register server node in event storage, like
// zookeeper, etcd, check if local node is master
type Master interface {
	Init() error                                  //init stage, like create connection
	Finit()                                       //finit, release resource
	Register() error                              //registery information to storage
	Clean() error                                 //clean self node
	IsMaster() bool                               //check if self is master or not
	CheckSelfNode() (bool, error)                 //check self node exist, and data correct
	GetAllNodes() ([]*bcstypes.ServerInfo, error) //get all server nodes
	GetPath() string                              //get parent path
}

//Empty for test
type Empty struct{}

//Init init stage, like create connection
func (e *Empty) Init() error {
	return nil
}

//Finit init stage, like create connection
func (e *Empty) Finit() {
}

//Register registery information to storage
func (e *Empty) Register() error {
	return nil

}

//Clean clean self node
func (e *Empty) Clean() error {
	return nil
}

//IsMaster check if self is master or not
func (e *Empty) IsMaster() bool {
	return false
}

//CheckSelfNode check self node exist, and data correct
func (e *Empty) CheckSelfNode() (bool, error) {
	return false, nil
}

//GetAllNodes get all server nodes
func (e *Empty) GetAllNodes() ([]*bcstypes.ServerInfo, error) {
	return nil, nil
}

//GetPath setting self info, now is ip address & port
func (e *Empty) GetPath() string {
	return ""
}
