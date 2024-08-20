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

// Package templatespacecollect template space collect
package templatespacecollect

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
	tableName = "templatespacecollect"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyTemplateSpaceID, Value: 1},
				bson.E{Key: entity.FieldKeyCreator, Value: 1},
			},
			Unique: false,
		},
	}
)

// ModelTemplateSpaceCollect provides handling template space collect operations to database
type ModelTemplateSpaceCollect struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelTemplateSpaceCollect instance
func New(db drivers.DB) *ModelTemplateSpaceCollect {
	return &ModelTemplateSpaceCollect{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

// ensure object database table and table indexes
func (m *ModelTemplateSpaceCollect) ensureTable(ctx context.Context) error {
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

// GetTemplateSpaceCollect get a specific entity.TemplateSpaceCollect from database
func (m *ModelTemplateSpaceCollect) GetTemplateSpaceCollect(ctx context.Context, id string) (
	*entity.TemplateSpaceCollect, error) {
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

	templateSpaceCollect := &entity.TemplateSpaceCollect{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, templateSpaceCollect); err != nil {
		return nil, err
	}

	return templateSpaceCollect, nil
}

// ListTemplateSpaceCollect get a list of entity.TemplateSpaceCollect by condition from database
func (m *ModelTemplateSpaceCollect) ListTemplateSpaceCollect(
	ctx context.Context, templateSpaceID, projectCode, creator string) ([]*entity.TemplateSpaceAndCollect, error) {

	t := make([]*entity.TemplateSpaceAndCollect, 0)
	// 构建聚合管道, 文件夹名称有可能会更新，不想在这张表维护
	match := operator.M{"projectCode": projectCode, "creator": creator}
	if templateSpaceID != "" {
		spaceID, err := primitive.ObjectIDFromHex(templateSpaceID)
		if err != nil {
			return nil, err
		}
		match[entity.FieldKeyTemplateSpaceID] = spaceID
	}
	pipeline := []operator.M{
		{"$match": match},
		{"$lookup": operator.M{
			"from":         "bcsclusterresources_templatespace",
			"localField":   "templateSpaceID",
			"foreignField": "_id",
			"as":           "templatespace",
		}},
		{"$unwind": "$templatespace"},
		{"$project": operator.M{
			"_id":             1,
			"templateSpaceID": 1,
			"projectCode":     1,
			"creator":         1,
			"name":            "$templatespace.name",
		}},
	}

	err := m.db.Table(m.tableName).Aggregation(ctx, pipeline, &t)

	if err != nil {
		return nil, err
	}

	return t, nil
}

// CreateTemplateSpaceCollect create a new entity.TemplateSpaceCollect into database
func (m *ModelTemplateSpaceCollect) CreateTemplateSpaceCollect(
	ctx context.Context, templateSpaceCollect *entity.TemplateSpaceCollect) (string, error) {
	if templateSpaceCollect == nil {
		return "", fmt.Errorf("can not create empty templateSpaceCollect")
	}

	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	// 没有id的情况下生成
	now := time.Now()
	if templateSpaceCollect.ID.IsZero() {
		templateSpaceCollect.ID = primitive.NewObjectIDFromTimestamp(now)
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{templateSpaceCollect}); err != nil {
		return "", err
	}
	return templateSpaceCollect.ID.Hex(), nil
}

// DeleteTemplateSpaceCollect delete a specific entity.TemplateSpaceCollect from database
func (m *ModelTemplateSpaceCollect) DeleteTemplateSpaceCollect(ctx context.Context, id string) error {
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
