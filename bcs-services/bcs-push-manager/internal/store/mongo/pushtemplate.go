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

	"github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/types"
)

const (
	templateDomainKey = "domain"
	templateIDKey     = "template_id"
)

// PushTemplateStore defines the storage interface for push templates.
type PushTemplateStore interface {
	CreatePushTemplate(ctx context.Context, template *types.PushTemplate) error
	DeletePushTemplate(ctx context.Context, templateID string) error
	GetPushTemplate(ctx context.Context, templateID string) (*types.PushTemplate, error)
	ListPushTemplates(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushTemplate, int64, error)
	UpdatePushTemplate(ctx context.Context, templateID string, update bson.M) error
}

// pushTemplateStore is a MongoDB-based implementation of PushTemplateStore.
type pushTemplateStore struct {
	collection *mongo.Collection
}

// NewPushTemplateStore creates a new PushTemplateStore instance.
func NewPushTemplateStore(db *mongo.Database) PushTemplateStore {
	coll := db.Collection(types.CollectionPushTemplate)
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: templateDomainKey, Value: 1},
			{Key: templateIDKey, Value: 1},
		},
		Options: options.Index().SetUnique(true).SetName(templateDomainKey + "_" + templateIDKey + "_1"),
	}
	_, err := coll.Indexes().CreateOne(context.Background(), indexModel)
	if err != nil {
		blog.Error("failed to create index: %v\n", err)
	}
	return &pushTemplateStore{
		collection: coll,
	}
}

// CreatePushTemplate inserts a new push template into the database.
func (s *pushTemplateStore) CreatePushTemplate(ctx context.Context, template *types.PushTemplate) error {
	template.ID = primitive.NewObjectID()
	template.CreatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, template)
	return err
}

// DeletePushTemplate deletes a push template from the database by its ID.
func (s *pushTemplateStore) DeletePushTemplate(ctx context.Context, templateID string) error {
	filter := bson.M{"template_id": templateID}
	_, err := s.collection.DeleteOne(ctx, filter)
	return err
}

// GetPushTemplate retrieves a single push template from the database by its ID.
func (s *pushTemplateStore) GetPushTemplate(ctx context.Context, templateID string) (*types.PushTemplate, error) {
	var template types.PushTemplate
	filter := bson.M{"template_id": templateID}
	err := s.collection.FindOne(ctx, filter).Decode(&template)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &template, err
}

// ListPushTemplates retrieves a list of push templates from the database with filtering and pagination.
func (s *pushTemplateStore) ListPushTemplates(ctx context.Context, filter bson.M, page, pageSize int64) ([]*types.PushTemplate, int64, error) {

	if page < 1 {
		return nil, 0, fmt.Errorf("invalid page: must be greater than or equal to 1")
	}
	if pageSize <= 0 {
		return nil, 0, fmt.Errorf("invalid pageSize: must be greater than 0")
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

	var templates []*types.PushTemplate
	if err = cursor.All(ctx, &templates); err != nil {
		return nil, 0, err
	}

	total, err := s.collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	return templates, total, nil
}

// UpdatePushTemplate updates a push template in the database.
func (s *pushTemplateStore) UpdatePushTemplate(ctx context.Context, templateID string, update bson.M) error {
	filter := bson.M{"template_id": templateID}
	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}
