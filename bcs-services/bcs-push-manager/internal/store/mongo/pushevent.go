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
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

// PushEventStore defines the storage interface for push events.
type PushEventStore interface {
	CreatePushEvent(ctx context.Context, event *types.PushEvent) error
	DeletePushEvent(ctx context.Context, eventID string) error
	GetPushEvent(ctx context.Context, eventID string) (*types.PushEvent, error)
	ListPushEvents(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushEvent, int64, error)
	UpdatePushEvent(ctx context.Context, eventID string, update bson.M) error
	UpdatePushEventStatus(ctx context.Context, eventID string, status int) error
	AppendNotificationResult(ctx context.Context, eventID string, channel, result string) error
}

// pushEventStore is a MongoDB-based implementation of PushEventStore.
type pushEventStore struct {
	collection *mongo.Collection
}

// NewPushEventStore creates a new PushEventStore instance.
func NewPushEventStore(db *mongo.Database) PushEventStore {
	return &pushEventStore{
		collection: db.Collection(types.CollectionPushEvent),
	}
}

// CreatePushEvent inserts a new push event into the database.
func (s *pushEventStore) CreatePushEvent(ctx context.Context, event *types.PushEvent) error {
	event.ID = primitive.NewObjectID()
	event.CreatedAt = time.Now()
	event.UpdatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, event)
	return err
}

// DeletePushEvent deletes a push event from the database by event_id.
func (s *pushEventStore) DeletePushEvent(ctx context.Context, eventID string) error {
	filter := bson.M{"event_id": eventID}
	_, err := s.collection.DeleteOne(ctx, filter)
	return err
}

// GetPushEvent retrieves a single push event from the database by event_id.
func (s *pushEventStore) GetPushEvent(ctx context.Context, eventID string) (*types.PushEvent, error) {
	var event types.PushEvent
	filter := bson.M{"event_id": eventID}
	err := s.collection.FindOne(ctx, filter).Decode(&event)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &event, err
}

// ListPushEvents retrieves a list of push events from the database with filtering and pagination.
func (s *pushEventStore) ListPushEvents(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushEvent, int64, error) {
	findOptions := options.Find()
	findOptions.SetSkip((page - 1) * pageSize)
	findOptions.SetLimit(pageSize)

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var events []*types.PushEvent
	if err = cursor.All(ctx, &events); err != nil {
		return nil, 0, err
	}

	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return events, total, nil
}

// UpdatePushEvent updates a push event in the database.
func (s *pushEventStore) UpdatePushEvent(ctx context.Context, eventID string, update bson.M) error {
	filter := bson.M{"event_id": eventID}
	update["$set"].(bson.M)["updated_at"] = time.Now()
	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// UpdatePushEventStatus updates the status of a specific event.
func (s *pushEventStore) UpdatePushEventStatus(ctx context.Context, eventID string, status int) error {
	update := bson.M{"$set": bson.M{"status": status}}
	return s.UpdatePushEvent(ctx, eventID, update)
}

// AppendNotificationResult appends or updates the notification_results field for an event.
func (s *pushEventStore) AppendNotificationResult(ctx context.Context, eventID string, channel, result string) error {
	event, err := s.GetPushEvent(ctx, eventID)
	if err != nil || event == nil {
		return err
	}
	if event.NotificationResults.Fields == nil {
		event.NotificationResults.Fields = make(map[string]string)
	}
	event.NotificationResults.Fields[channel] = result
	update := bson.M{"$set": bson.M{"notification_results": event.NotificationResults}}
	return s.UpdatePushEvent(ctx, eventID, update)
}
