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

// Package envmanage environment manage
package envmanage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/cluster-resources/pkg/store/utils"
)

const (
	tableName = "envmanage"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyEnv, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelEnvManage provides handling EnvManage operations to database
type ModelEnvManage struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelEnvManage instance
func New(db drivers.DB) *ModelEnvManage {
	return &ModelEnvManage{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelEnvManage) ensureTable(ctx context.Context) error {
	if m.isTableEnsured {
		return nil
	}

	m.isTableEnsuredMutex.Lock()
	defer m.isTableEnsuredMutex.Unlock()
	if m.isTableEnsured {
		return nil
	}

	if err := utils.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		return err
	}
	m.isTableEnsured = true
	return nil
}

// GetEnvManage get a specific entity.EnvManage from database
func (m *ModelEnvManage) GetEnvManage(ctx context.Context, id string) (
	*entity.EnvManage, error) {
	if id == "" {
		return nil, fmt.Errorf("can not get with empty id")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})

	envManage := &entity.EnvManage{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, envManage); err != nil {
		return nil, err
	}

	return envManage, nil
}

// ListEnvManages get a list of entity.EnvManage by condition and option from database
func (m *ModelEnvManage) ListEnvManages(ctx context.Context, cond *operator.Condition) ([]*entity.EnvManage, error) {

	l := make([]*entity.EnvManage, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &l); err != nil {
		return nil, err
	}

	return l, nil
}

// CreateEnvManage create a new entity.EnvManage into database
func (m *ModelEnvManage) CreateEnvManage(ctx context.Context, envManage *entity.EnvManage) (string, error) {
	if envManage == nil {
		return "", fmt.Errorf("can not create empty envManage")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	now := time.Now()
	if envManage.ID.IsZero() {
		envManage.ID = primitive.NewObjectIDFromTimestamp(now)
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{envManage}); err != nil {
		return "", err
	}
	return envManage.ID.Hex(), nil
}

// UpdateEnvManage update an entity.EnvManage into database
func (m *ModelEnvManage) UpdateEnvManage(ctx context.Context, id string, envManage entity.M) error {
	if id == "" {
		return fmt.Errorf("can not update with empty id")
	}

	if envManage == nil {
		return fmt.Errorf("can not update empty envManage")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	old := &entity.EnvManage{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": envManage}); err != nil {
		return err
	}

	return nil
}

// DeleteEnvManage delete a specific entity.EnvManage from database
func (m *ModelEnvManage) DeleteEnvManage(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("can not get with empty id")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}
