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

package errors

import (
	"github.com/Tencent/bk-bcs/bcs-common/common"
)

type StorageError struct {
	Message string
	Code    int
}

func (se *StorageError) Error() string {
	return se.Message
}

var (
	TransactionChainBreak        = &StorageError{Code: common.AdditionErrorCode + 6303, Message: "mongodb transaction chain break"}
	UnknownOperationType         = &StorageError{Code: common.AdditionErrorCode + 6304, Message: "unknown operation type"}
	MongodbCollectionNoFound     = &StorageError{Code: common.AdditionErrorCode + 6305, Message: "mongodb collection no found"}
	MongodbDatabasesNoFound      = &StorageError{Code: common.AdditionErrorCode + 6305, Message: "mongodb databases no found"}
	StableVersionNoFound         = &StorageError{Code: common.AdditionErrorCode + 6306, Message: "cluster config stable version no found"}
	StableVersionInvalid         = &StorageError{Code: common.AdditionErrorCode + 6306, Message: "cluster config stable version is invalid"}
	ServiceConfigNoFound         = &StorageError{Code: common.AdditionErrorCode + 6306, Message: "service config no found"}
	ConfigTemplateNoFound        = &StorageError{Code: common.AdditionErrorCode + 6306, Message: "config template no found"}
	ConfigTemplateInvalid        = &StorageError{Code: common.AdditionErrorCode + 6306, Message: "config template is invalid"}
	DatabaseConfigUnknown        = &StorageError{Code: common.AdditionErrorCode + 6307, Message: "Database config unknown"}
	MongodbDriverNotExist        = &StorageError{Code: common.AdditionErrorCode + 6308, Message: "Mongodb driver does not exist"}
	MongodbTankNotInit           = &StorageError{Code: common.AdditionErrorCode + 6309, Message: "Mongodb tank does not init"}
	ZookeeperDriverNotExist      = &StorageError{Code: common.AdditionErrorCode + 6310, Message: "Zookeeper driver does not exist"}
	ZookeeperTankNotInit         = &StorageError{Code: common.AdditionErrorCode + 6311, Message: "Zookeeper tank does not init"}
	SetTableVNotSupported        = &StorageError{Code: common.AdditionErrorCode + 6312, Message: "SetTableV is not supported by this driver"}
	GetTableVNotSupported        = &StorageError{Code: common.AdditionErrorCode + 6313, Message: "GetTableV is not supported by this driver"}
	ZookeeperClientNoFound       = &StorageError{Code: common.AdditionErrorCode + 6314, Message: "Zookeeper client no found"}
	MongodbDriverAlreadyInPool   = &StorageError{Code: common.AdditionErrorCode + 6315, Message: "mongodb driver already in pool"}
	ZookeeperDriverAlreadyInPool = &StorageError{Code: common.AdditionErrorCode + 6316, Message: "zookeeper driver already in pool"}
	EventWatchAlreadyConnect     = &StorageError{Code: common.AdditionErrorCode + 6317, Message: "already connected"}
	EventWatchNoUrlAvailable     = &StorageError{Code: common.AdditionErrorCode + 6318, Message: "no url available"}
	ResourceDoesNotExist         = &StorageError{Code: common.AdditionErrorCode + 6319, Message: "resource does not exist"}
	RemoveLessThanMatch          = &StorageError{Code: common.AdditionErrorCode + 6320, Message: "remove less than match"}
	UpdateLessThanMatch          = &StorageError{Code: common.AdditionErrorCode + 6321, Message: "update less than match"}
	QueueConfigUnknown           = &StorageError{Code: common.AdditionErrorCode + 6307, Message: "queue config unknown"}
)
