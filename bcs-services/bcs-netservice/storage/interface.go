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

package storage

//Event for callback register
type Event interface {
	AddEvent(key string, value []byte)
	UpdateEvent(key string, oldData, curData []byte)
	DeleteEvent(key string, value []byte)
}

//Locker for storage, basic use
//for pool Delete & IP lease/release
type Locker interface {
	Lock() error
	Unlock() error
}

//Storage interface for key/value storage
type Storage interface {
	GetLocker(path string) (Locker, error)           //get locker with new connection
	Register(path string, data []byte) error         //register self node
	RegisterAndWatch(path string, data []byte) error //register and watch self node
	Add(key string, value []byte) error              //add new node data
	Delete(key string) ([]byte, error)               //delete node
	Update(key string, data []byte) error            //update node data
	Get(key string) ([]byte, error)                  //get node data
	List(key string) ([]string, error)               //list all children nodes
	Exist(key string) (bool, error)                  //check node exist
	Stop()                                           //close connection
}
