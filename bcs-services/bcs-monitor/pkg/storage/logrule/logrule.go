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
 *
 */

package logrule

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-monitor/pkg/utils"
)

const (
	tableName = "logrule"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectID, Value: 1},
				bson.E{Key: entity.FieldKeyClusterID, Value: 1},
				bson.E{Key: entity.FieldKeyName, Value: 1},
			},
			Unique: true,
		},
		{
			Name: tableName + "_project_cluster",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectID, Value: 1},
				bson.E{Key: entity.FieldKeyClusterID, Value: 1},
			},
		},
	}
)

// ModelLogRule provides handling log rule operations to database
type ModelLogRule struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelLogRule instance
func New(db drivers.DB) *ModelLogRule {
	return &ModelLogRule{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelLogRule) ensureTable(ctx context.Context) error {
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

// CreateLogRule create log rule
func (m *ModelLogRule) CreateLogRule(ctx context.Context, lc *entity.LogRule) (string, error) {
	if err := m.ensureTable(ctx); err != nil {
		return "", err
	}

	now := utils.JSONTime{Time: time.Now()}
	lc.CreatedAt = now
	lc.UpdatedAt = now
	if lc.ID.IsZero() {
		lc.ID = primitive.NewObjectIDFromTimestamp(now.Time)
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{lc}); err != nil {
		return "", err
	}
	return lc.ID.Hex(), nil
}

// UpdateLogRule update log rule
func (m *ModelLogRule) UpdateLogRule(ctx context.Context, id string, lc entity.M) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})

	lc[entity.FieldKeyUpdatedAt] = utils.JSONTime{Time: time.Now()}
	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": lc}); err != nil {
		return err
	}

	return nil
}

// DeleteLogRule delete log rule
func (m *ModelLogRule) DeleteLogRule(ctx context.Context, id string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
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

// ListLogRules list log rules
func (m *ModelLogRule) ListLogRules(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.LogRule, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, nil, err
	}

	l := make([]*entity.LogRule, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(opt.Sort)
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

	for i := range l {
		l[i].FixStatus()
	}
	return total, l, nil
}

// GetLogRule get log rule
func (m *ModelLogRule) GetLogRule(ctx context.Context, id string) (*entity.LogRule, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyObjectID: objectID,
	})
	lc := &entity.LogRule{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, lc); err != nil {
		return nil, err
	}

	lc.FixStatus()
	return lc, nil
}

// GetIndexSetID get index set id
func (m *ModelLogRule) GetIndexSetID(ctx context.Context, projectID, clusterID string) (int, int, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, 0, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyClusterID: clusterID,
	})
	l := make([]*entity.LogRule, 0)
	if err := m.db.Table(m.tableName).Find(cond).All(ctx, &l); err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	stdIndexSetID := 0
	fileIndexSetID := 0
	for _, v := range l {
		if v.STDIndexSetID != 0 && v.FileIndexSetID != 0 {
			stdIndexSetID = v.STDIndexSetID
			fileIndexSetID = v.FileIndexSetID
			break
		}
	}
	return stdIndexSetID, fileIndexSetID, nil
}
