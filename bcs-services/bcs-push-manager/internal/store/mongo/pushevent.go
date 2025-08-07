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
 */

// Package mongo provides a MongoDB-based implementation of the data store interfaces.
package mongo

import (
	"context"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

var (
	modelPushEventIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: pushDomainKey, Value: 1},
				bson.E{Key: pushEventUniqueKey, Value: 1},
			},
			Name:   pushEventTableName + "_1",
			Unique: true,
		},
		{
			Key: bson.D{
				bson.E{Key: pushEventUniqueKey, Value: 1},
			},
			Name:   pushEventUniqueKey + "_1",
			Unique: true,
		},
	}
)

// ModelPushEvent is a MongoDB-based implementation of PushEventStore.
type ModelPushEvent struct {
	Public
}

// NewModelPushEvent creates a new PushEventStore instance.
func NewModelPushEvent(db drivers.DB) *ModelPushEvent {
	return &ModelPushEvent{
		Public: Public{
			TableName: tableNamePrefix + pushEventTableName,
			Indexes:   modelPushEventIndexes,
			DB:        db,
		}}
}

// CreatePushEvent inserts a new push event into the database.
func (m *ModelPushEvent) CreatePushEvent(ctx context.Context, event *types.PushEvent) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if event == nil {
		return fmt.Errorf("push event is nil")
	}
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	if _, err := m.DB.Table(m.TableName).Insert(ctx, []interface{}{event}); err != nil {
		return fmt.Errorf("create push event failed: %v", err)
	}
	return nil
}

// DeletePushEvent deletes a push event from the database by event_id.
func (m *ModelPushEvent) DeletePushEvent(ctx context.Context, eventID string) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if eventID == "" {
		return fmt.Errorf("eventID cannot be empty")
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushEventUniqueKey: eventID,
	})

	if _, err := m.DB.Table(m.TableName).Delete(ctx, cond); err != nil {
		return fmt.Errorf("delete push event failed: %v", err)
	}
	return nil
}

// GetPushEvent retrieves a single push event from the database by event_id.
func (m *ModelPushEvent) GetPushEvent(ctx context.Context, eventID string) (*types.PushEvent, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, fmt.Errorf("ensure table failed: %v", err)
	}
	if eventID == "" {
		return nil, fmt.Errorf("eventID cannot be empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushEventUniqueKey: eventID,
	})

	var event types.PushEvent
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, &event); err != nil {
		return nil, fmt.Errorf("get push event failed: %v", err)
	}
	return &event, nil
}

// ListPushEvents retrieves a list of push events from the database with filtering and pagination.
func (m *ModelPushEvent) ListPushEvents(ctx context.Context, filter operator.M, page, pageSize int64) ([]*types.PushEvent, int64, error) {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return nil, 0, fmt.Errorf("ensure table failed: %v", err)
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{})
	if filter != nil {
		cond = operator.NewBranchCondition(operator.And, cond, operator.NewLeafCondition(operator.Eq, filter))
	}

	var events []*types.PushEvent
	finder := m.DB.Table(m.TableName).Find(cond)
	if page > 1 {
		finder = finder.WithStart((page - 1) * pageSize)
	}
	if pageSize > 0 {
		finder = finder.WithLimit(pageSize)
	}
	if err := finder.All(ctx, &events); err != nil {
		return nil, 0, fmt.Errorf("list push events failed: %v", err)
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count push events failed: %v", err)
	}

	return events, total, nil
}

// UpdatePushEvent updates a push event in the database.
func (m *ModelPushEvent) UpdatePushEvent(ctx context.Context, eventID string, update operator.M) error {
	if err := ensureTable(ctx, &m.Public); err != nil {
		return fmt.Errorf("ensure table failed: %v", err)
	}
	if eventID == "" {
		return fmt.Errorf("eventID cannot be empty")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		pushEventUniqueKey: eventID,
	})
	if update == nil {
		return fmt.Errorf("update cannot be nil")
	}
	set, ok := update["$set"].(operator.M)
	if !ok {
		return fmt.Errorf("invalid update format: $set must be operator.M type")
	}
	if set == nil {
		set = operator.M{}
		update["$set"] = set
	}
	set["updated_at"] = time.Now()

	if err := m.DB.Table(m.TableName).Update(ctx, cond, update); err != nil {
		return fmt.Errorf("update push event failed: %v", err)
	}
	return nil
}

// UpdatePushEventStatus updates the status of a specific event.
func (m *ModelPushEvent) UpdatePushEventStatus(ctx context.Context, eventID string, status int) error {
	return m.UpdatePushEvent(ctx, eventID, operator.M{
		"$set": operator.M{"status": status},
	})
}

// AppendNotificationResult appends or updates the notification_results field for an event.
func (m *ModelPushEvent) AppendNotificationResult(ctx context.Context, eventID string, channel, result string) error {
	event, err := m.GetPushEvent(ctx, eventID)
	if err != nil {
		return fmt.Errorf("get push event failed: %v", err)
	}
	if event == nil {
		return fmt.Errorf("push event not found")
	}

	if event.NotificationResults.Fields == nil {
		event.NotificationResults.Fields = make(map[string]string)
	}
	event.NotificationResults.Fields[channel] = result

	return m.UpdatePushEvent(ctx, eventID, operator.M{
		"$set": operator.M{"notification_results": event.NotificationResults},
	})
}
