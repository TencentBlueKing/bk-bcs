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

// Package templatespace template space
package templatespace

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
	tableName = "templatespace"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyName, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelTemplateSpace provides handling template space operations to database
type ModelTemplateSpace struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelTemplateSpace instance
func New(db drivers.DB) *ModelTemplateSpace {
	return &ModelTemplateSpace{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

// ensure object database table and table indexes
func (m *ModelTemplateSpace) ensureTable(ctx context.Context) error {
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

// GetTemplateSpace get a specific entity.TemplateSpace from database
func (m *ModelTemplateSpace) GetTemplateSpace(ctx context.Context, id string) (
	*entity.TemplateSpace, error) {
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

	templateSpace := &entity.TemplateSpace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, templateSpace); err != nil {
		return nil, err
	}

	return templateSpace, nil
}

// ListTemplateSpace get a list of entity.TemplateSpace by condition from database
func (m *ModelTemplateSpace) ListTemplateSpace(
	ctx context.Context, cond *operator.Condition) ([]*entity.TemplateSpace, error) {

	t := make([]*entity.TemplateSpace, 0)
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &t)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// CreateTemplateSpace create a new entity.TemplateSpace into database
func (m *ModelTemplateSpace) CreateTemplateSpace(
	ctx context.Context, templateSpace *entity.TemplateSpace) (string, error) {
	if templateSpace == nil {
		return "", fmt.Errorf("can not create empty templatespace")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	// 没有id的情况下生成
	now := time.Now()
	if templateSpace.ID.IsZero() {
		templateSpace.ID = primitive.NewObjectIDFromTimestamp(now)
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{templateSpace}); err != nil {
		return "", err
	}
	return templateSpace.ID.Hex(), nil
}

// CreateTemplateSpaceBatch create many new entity.TemplateSpace into database
func (m *ModelTemplateSpace) CreateTemplateSpaceBatch(
	ctx context.Context, templateSpaces []*entity.TemplateSpace) error {
	if len(templateSpaces) == 0 {
		return nil
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	insertValues := make([]interface{}, 0)
	now := time.Now()
	for _, templateSpace := range templateSpaces {
		// id 覆盖
		templateSpace.ID = primitive.NewObjectIDFromTimestamp(now)
		insertValues = append(insertValues, templateSpace)
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, insertValues); err != nil {
		return err
	}
	return nil
}

// UpdateTemplateSpace update an entity.TemplateSpace into database
func (m *ModelTemplateSpace) UpdateTemplateSpace(ctx context.Context, id string, templateSpace entity.M) error {
	if id == "" {
		return fmt.Errorf("can not update with empty id")
	}

	if templateSpace == nil {
		return fmt.Errorf("can not update empty templateSpace")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	old := &entity.TemplateSpace{}
	if err = m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": templateSpace}); err != nil {
		return err
	}

	return nil
}

// DeleteTemplateSpace delete a specific entity.TemplateSpace from database
func (m *ModelTemplateSpace) DeleteTemplateSpace(ctx context.Context, id string) error {
	if id == "" {
		return fmt.Errorf("can not delete with empty id")
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
