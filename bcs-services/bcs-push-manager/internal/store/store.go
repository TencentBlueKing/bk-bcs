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

// Package store provides a unified data access layer for the application.
package store

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	opt "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/options"
	mg "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
)

// Store aggregates all data store interfaces.
type Store struct {
	client        *mongo.Client
	PushEvent     mg.PushEventStore
	PushWhitelist mg.PushWhitelistStore
	PushTemplate  mg.PushTemplateStore
}

// NewStore creates a new Store instance and initializes the database connection.
func NewStore(mongoOpt *opt.MongoOption) (*Store, error) {
	if mongoOpt == nil {
		return nil, fmt.Errorf("mongoOpt is nil")
	}
	clientOptions := options.Client().ApplyURI(mongoOpt.Endpoints)

	if mongoOpt.Username != "" && mongoOpt.Password != "" {
		credential := options.Credential{
			Username: mongoOpt.Username,
			Password: mongoOpt.Password,
		}
		clientOptions.SetAuth(credential)
	}

	if mongoOpt.ConnectTimeout > 0 {
		timeout := time.Duration(mongoOpt.ConnectTimeout) * time.Second
		clientOptions.SetConnectTimeout(timeout)
	}

	client, err := mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}
	db := client.Database(mongoOpt.Database)

	return &Store{
		client:        client,
		PushEvent:     mg.NewPushEventStore(db),
		PushWhitelist: mg.NewPushWhitelistStore(db),
		PushTemplate:  mg.NewPushTemplateStore(db),
	}, nil
}

// Close disconnects the MongoDB client.
func (s *Store) Close() error {
	if s.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.client.Disconnect(ctx)
	}
	return nil
}
