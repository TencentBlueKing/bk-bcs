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

package mongo

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	modelEventIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: nodeGroupIDKey, Value: 1},
			},
			Name: nodeGroupIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: clusterIDKey, Value: 1},
			},
			Name: clusterIDKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: eventKey, Value: 1},
			},
			Name: eventKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: eventKey, Value: 1},
			},
			Name: eventKey + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: eventTimeKey, Value: 1},
			},
			Name: eventTimeKey + "_1",
		},
	}
)

// ModelEvent defines model event
type ModelEvent struct {
	Public
}

// NewModelEvent new modelEvent
func NewModelEvent(db drivers.DB) *ModelEvent {
	return &ModelEvent{Public{
		TableName: tableNamePrefix + eventTableName,
		Indexes:   modelEventIndexes,
		DB:        db,
	}}
}

// ListNodeGroupEvent list NodeGroupEvent by nodeGroupID, if id is empty, return all
func (m *ModelEvent) ListNodeGroupEvent(nodeGroupID string, opt *storage.ListOptions) ([]*storage.NodeGroupEvent,
	error) {
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	page := opt.Page
	limit := opt.Limit

	cond := make([]*operator.Condition, 0)
	if nodeGroupID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			nodeGroupIDKey: nodeGroupID,
		}))
	}
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	if !opt.DoPagination && opt.Limit == 0 {
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get event count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	nodeEventList := make([]*storage.NodeGroupEvent, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nodeGroupIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &nodeEventList)
	if err != nil {
		return nil, fmt.Errorf("list nodeGroups err:%v", err)
	}
	return nodeEventList, nil
}

// CreateNodeGroupEvent create NodeGroupEvent, nodegroupID, clusterID and event cannot be empty
func (m *ModelEvent) CreateNodeGroupEvent(event *storage.NodeGroupEvent, opt *storage.CreateOptions) error {
	if opt == nil || event.NodeGroupID == "" || event.ClusterID == "" || event.Event == "" {
		return fmt.Errorf("CreateOption is nil or nodegroupID/clusterID/event is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{event})
	if err != nil {
		return fmt.Errorf("insert nodeGroupEvent error: %v", err)
	}
	return nil
}
