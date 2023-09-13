/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 * Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under,
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Package view xxx
package view

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
	tableName = "view"
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

// ModelView provides handling view-related operations to database
type ModelView struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelView instance
func New(db drivers.DB) *ModelView {
	return &ModelView{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelView) ensureTable(ctx context.Context) error {
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

// CreateView create a new entity.View into database
func (m *ModelView) CreateView(ctx context.Context, view *entity.View) (string, error) {
	if view == nil {
		return "", fmt.Errorf("can not create empty view")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	now := time.Now()
	if view.ID.IsZero() {
		view.ID = primitive.NewObjectIDFromTimestamp(now)
	}
	if view.CreateAt == 0 {
		view.CreateAt = now.UTC().Unix()
	}
	if view.UpdateAt == 0 {
		view.UpdateAt = now.UTC().Unix()
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{view}); err != nil {
		return "", err
	}
	return view.ID.Hex(), nil
}

// UpdateView update an entity.View into database
func (m *ModelView) UpdateView(ctx context.Context, id string, view entity.M) error {
	if id == "" {
		return fmt.Errorf("can not update with empty id")
	}

	if view == nil {
		return fmt.Errorf("can not update empty view")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	old := &entity.View{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	if view[entity.FieldKeyUpdateAt] == nil {
		view.Update(entity.FieldKeyUpdateAt, time.Now().UTC().Unix())
	}
	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": view}); err != nil {
		return err
	}

	return nil
}

// GetViewByName get a specific entity.sView from database
func (m *ModelView) GetViewByName(ctx context.Context, projectCode, name string) (
	*entity.View, error) {
	if projectCode == "" {
		return nil, fmt.Errorf("can not get with empty projectCode")
	}
	if name == "" {
		return nil, fmt.Errorf("can not get with empty name")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: projectCode,
		entity.FieldKeyName:        name,
	})

	view := &entity.View{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, view); err != nil {
		return nil, err
	}

	return view, nil
}

// GetView get a specific entity.sView from database
func (m *ModelView) GetView(ctx context.Context, id string) (
	*entity.View, error) {
	if id == "" {
		return nil, fmt.Errorf("can not get with empty id")
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: primitive.ObjectID(objectID),
	})

	view := &entity.View{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, view); err != nil {
		return nil, err
	}

	return view, nil
}

// ListViews get a list of entity.View by condition and option from database
func (m *ModelView) ListViews(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.View, error) {

	l := make([]*entity.View, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(utils.MapInt2MapIf(opt.Sort))
	}
	if opt.Page != 0 {
		finder = finder.WithStart(opt.Page * opt.Size)
	}
	if opt.Size != 0 {
		finder = finder.WithLimit(opt.Size)
	}

	if err := finder.All(ctx, &l); err != nil {
		return 0, nil, err
	}

	total, err := finder.Count(ctx)
	if err != nil {
		return 0, nil, err
	}

	return total, l, nil
}

// DeleteView delete a specific entity.View from database
func (m *ModelView) DeleteView(ctx context.Context, id string) error {
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
