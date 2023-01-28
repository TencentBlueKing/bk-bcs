/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package mongo xxx
package mongo

import (
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

const (
	tableNamePrefix    = "nodegroup_manager_"
	strategyTableName  = "strategy"
	actionTableName    = "action"
	eventTableName     = "event"
	nodeGroupTableName = "node_group"
	taskTableName      = "task"
)

const (
	nameKey         = "name"
	nodeGroupIDKey  = "node_group_id"
	clusterIDKey    = "cluster_id"
	eventKey        = "event"
	eventTimeKey    = "event_time"
	isDeletedKey    = "is_deleted"
	taskIDKey       = "task_id"
	strategyTypeKey = "strategy.type"
	strategyKey     = "node_group_strategy"
)

type server struct {
	*ModelStrategy
	*ModelGroup
	*ModelAction
	*ModelEvent
	*ModelTask
}

// NewServer new db server
func NewServer(db drivers.DB) storage.Storage {
	return &server{
		ModelStrategy: NewModelStrategy(db),
		ModelGroup:    NewModelGroup(db),
		ModelAction:   NewModelAction(db),
		ModelEvent:    NewModelEvent(db),
		ModelTask:     NewModelTask(db),
	}
}

// Public public model set
type Public struct {
	TableName           string
	Indexes             []drivers.Index
	DB                  drivers.DB
	IsTableEnsured      bool
	IsTableEnsuredMutex sync.RWMutex
}
