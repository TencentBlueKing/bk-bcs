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

import (
	"time"

	"golang.org/x/net/context"

	"github.com/Tencent/bk-bcs/bcs-mesos/bcs-mesos-watch/types"
)

//DataLister is interface for request data from storage
//type DataLister interface {
//	ListData(key string, sample interface{}) map[string]interface{}
//}

//InfoHandler is interface for handle detail data
type InfoHandler interface {
	//List(key interface{}) map[string]interface{}
	Add(data interface{}) error
	Update(data interface{}) error
	Delete(data interface{}) error
	GetType() string

	CheckDirty() error
}

//Storage is writer interface for all database
//such as zookeeper, Mysql, redis, CC and etc.
type Storage interface {
	//DataLister                          //embed DataLister interface for requesting data from Storage
	DataOperator                        //CREATE, DELETE, LIST operation for storage
	Sync(data *types.BcsSyncData) error //sync data
	SyncTimeout(data *types.BcsSyncData, timeout time.Duration) error
	Run(cxt context.Context) error //start point for StorageWriter
	Worker()                       //storage writer worker goroutine
	SetDCAddress(address []string) //set storage server address
	GetDCAddress() string
}

//DataOperator is Interface to Operator data for storage
type DataOperator interface {
	CreateDCNode(node string, value interface{}, action string) error
	DeleteDCNode(node, action string) error
	DeleteDCNodes(node string, value interface{}, action string) error
}
