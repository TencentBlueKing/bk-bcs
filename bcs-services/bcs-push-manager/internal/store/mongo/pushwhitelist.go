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

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/constant"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

const (
	whitelistDomainKey = "domain"
	whitelistIDKey     = "whitelist_id"
)

// PushWhitelistStore defines the storage interface for push whitelists.
type PushWhitelistStore interface {
	CreatePushWhitelist(ctx context.Context, whitelist *types.PushWhitelist) error
	DeletePushWhitelist(ctx context.Context, whitelistID string) error
	GetPushWhitelist(ctx context.Context, whitelistID string) (*types.PushWhitelist, error)
	ListPushWhitelists(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushWhitelist, int64, error)
	UpdatePushWhitelist(ctx context.Context, whitelistID string, update bson.M) error
	IsDimensionWhitelisted(ctx context.Context, domain string, dimension types.Dimension) (bool, error)
}

// pushWhitelistStore is a MongoDB-based implementation of PushWhitelistStore.
type pushWhitelistStore struct {
	collection *mongo.Collection
}

// NewPushWhitelistStore creates a new PushWhitelistStore instance.
func NewPushWhitelistStore(db *mongo.Database) PushWhitelistStore {
	coll := db.Collection(types.CollectionPushWhitelist)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: whitelistDomainKey, Value: 1},
			{Key: whitelistIDKey, Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(whitelistDomainKey + "_" + whitelistIDKey + "_1"),
	}
	_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		blog.Error("failed to create index: %v\n", err)
	}
	return &pushWhitelistStore{
		collection: coll,
	}
}

// CreatePushWhitelist inserts a new push whitelist into the database.
func (s *pushWhitelistStore) CreatePushWhitelist(ctx context.Context, whitelist *types.PushWhitelist) error {
	whitelist.ID = primitive.NewObjectID()
	whitelist.CreatedAt = time.Now()
	whitelist.UpdatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, whitelist)
	return err
}

// DeletePushWhitelist soft-deletes a push whitelist from the database by its ID.
func (s *pushWhitelistStore) DeletePushWhitelist(ctx context.Context, whitelistID string) error {
	filter := bson.M{"whitelist_id": whitelistID}
	now := time.Now()
	update := bson.M{
		"$set": bson.M{
			"deleted_at": &now,
			"updated_at": now,
		},
	}
	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// GetPushWhitelist retrieves a single push whitelist from the database by its ID.
func (s *pushWhitelistStore) GetPushWhitelist(ctx context.Context, whitelistID string) (*types.PushWhitelist, error) {
	var whitelist types.PushWhitelist
	filter := bson.M{
		"whitelist_id": whitelistID,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": nil},
		},
	}
	err := s.collection.FindOne(ctx, filter).Decode(&whitelist)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &whitelist, err
}

// ListPushWhitelists retrieves a list of push whitelists from the database with filtering and pagination.
func (s *pushWhitelistStore) ListPushWhitelists(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushWhitelist, int64, error) {
	filter["$or"] = []bson.M{
		{"deleted_at": bson.M{"$exists": false}},
		{"deleted_at": nil},
	}
	findOptions := options.Find()
	findOptions.SetSkip((page - 1) * pageSize)
	findOptions.SetLimit(pageSize)
	findOptions.SetSort(bson.M{"created_at": -1})

	cursor, err := s.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var whitelists []*types.PushWhitelist
	if err = cursor.All(ctx, &whitelists); err != nil {
		return nil, 0, err
	}

	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return whitelists, total, nil
}

// UpdatePushWhitelist updates a push whitelist in the database.
func (s *pushWhitelistStore) UpdatePushWhitelist(ctx context.Context, whitelistID string, update bson.M) error {
	filter := bson.M{
		"whitelist_id": whitelistID,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": nil},
		},
	}
	if update["$set"] == nil {
		update["$set"] = bson.M{}
	}
	update["$set"].(bson.M)["updated_at"] = time.Now()
	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

// IsDimensionWhitelisted checks if a given domain and dimension are whitelisted, active, and approved.
func (s *pushWhitelistStore) IsDimensionWhitelisted(ctx context.Context, domain string, dimension types.Dimension) (bool, error) {
	now := time.Now()
	filter := bson.M{
		"domain":           domain,
		"approval_status":  constant.ApprovalStatusApproved,
		"whitelist_status": constant.WhitelistStatusActive,
		"$or": []bson.M{
			{"deleted_at": bson.M{"$exists": false}},
			{"deleted_at": nil},
		},
	}
	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return false, fmt.Errorf("find whitelist failed: %w", err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var wl types.PushWhitelist
		if err := cursor.Decode(&wl); err != nil {
			blog.Errorf("decode push whitelist failed, document may be corrupted: %v", err)
			continue
		}
		match, _ := mapEqualsDetail(wl.Dimension.Fields, dimension.Fields)
		if !match {
			continue
		}
		if !wl.StartTime.IsZero() && wl.StartTime.After(now) {
			continue
		}
		if !wl.EndTime.IsZero() && wl.EndTime.Before(now) {
			update := bson.M{"$set": bson.M{"whitelist_status": constant.WhitelistStatusExpired, "updated_at": now}}
			_, _ = s.collection.UpdateOne(ctx, bson.M{"_id": wl.ID}, update)
			continue
		}
		return true, nil
	}
	if err := cursor.Err(); err != nil {
		return false, fmt.Errorf("cursor error: %w", err)
	}
	return false, nil
}

// mapEqualsDetail xxx
func mapEqualsDetail(a, b map[string]string) (bool, string) {
	if len(a) != len(b) {
		return false, fmt.Sprintf("length not equal: a=%d, b=%d", len(a), len(b))
	}
	var missing, extra, diff []string
	for k, v := range a {
		if bv, ok := b[k]; !ok {
			missing = append(missing, k)
		} else if bv != v {
			diff = append(diff, fmt.Sprintf("key=%s, a=%s, b=%s", k, v, bv))
		}
	}
	for k := range b {
		if _, ok := a[k]; !ok {
			extra = append(extra, k)
		}
	}
	if len(missing) == 0 && len(extra) == 0 && len(diff) == 0 {
		return true, "maps equal"
	}
	return false, fmt.Sprintf("missing in b: %v, extra in b: %v, value diff: %v", missing, extra, diff)
}
