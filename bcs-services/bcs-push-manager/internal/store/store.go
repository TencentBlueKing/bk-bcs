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

	orimongo "go.mongodb.org/mongo-driver/mongo"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"

	mg "github.com/Tencent/bk-bcs/bcs-services/bcs-push-manager/internal/store/mongo"
)

// Store aggregates all data store interfaces.
type Store struct {
	client        *orimongo.Client
	PushEvent     mg.PushEventStore
	PushWhitelist mg.PushWhitelistStore
	PushTemplate  mg.PushTemplateStore
}

// NewStore creates a new Store instance and initializes the database connection.
func NewStore(mongoOpt *mongo.Options) (*Store, error) {
	if mongoOpt == nil {
		return nil, fmt.Errorf("mongoOpt is nil")
	}
	mongoDB, err := mongo.NewDB(mongoOpt)
	if err != nil {
		blog.Errorf("init mongo db failed, err %s", err.Error())
		return nil, err
	}
	if err = mongoDB.Ping(); err != nil {
		blog.Errorf("ping mongo db failed, err %s", err.Error())
		return nil, err
	}
	client := mongoDB.Client()
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
