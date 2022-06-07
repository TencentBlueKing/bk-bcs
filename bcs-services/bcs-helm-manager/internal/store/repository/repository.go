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

package repository

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/common"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-helm-manager/internal/store/utils"
	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "repository"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectID, Value: 1},
				bson.E{Key: entity.FieldKeyName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelRepository provides handling repository-related operations to database
type ModelRepository struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelRepository instance
func New(db drivers.DB) *ModelRepository {
	return &ModelRepository{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelRepository) ensureTable(ctx context.Context) error {
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

// CreateRepository create a new entity.Repository into database
func (m *ModelRepository) CreateRepository(ctx context.Context, repository *entity.Repository) error {
	if repository == nil {
		return fmt.Errorf("can not create empty repository")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	timestamp := time.Now().UTC().Unix()
	repository.CreateTime = timestamp
	repository.UpdateTime = timestamp
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{repository}); err != nil {
		return err
	}
	return nil
}

// UpdateRepository update an entity.Repository into database
func (m *ModelRepository) UpdateRepository(ctx context.Context, projectID, name string, repository entity.M) error {
	if projectID == "" || name == "" {
		return fmt.Errorf("can not update with empty projectID or name")
	}

	if repository == nil {
		return fmt.Errorf("can not update empty repository")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyName:      name,
	})
	old := &entity.Repository{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, old); err != nil {
		return err
	}

	repository[entity.FieldKeyUpdateTime] = time.Now().UTC().Unix()
	if err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": repository}); err != nil {
		return err
	}

	return nil
}

// GetRepository get a specific entity.Repository from database
func (m *ModelRepository) GetRepository(ctx context.Context, projectID, name string) (*entity.Repository, error) {
	if projectID == "" || name == "" {
		return nil, fmt.Errorf("can not get with empty projectID or name")
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyName:      name,
	})
	repository := &entity.Repository{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, repository); err != nil {
		return nil, err
	}

	return repository, nil
}

// ListRepository get a list of entity.Repository by condition and option from database
func (m *ModelRepository) ListRepository(ctx context.Context, cond *operator.Condition, opt *utils.ListOption) (
	int64, []*entity.Repository, error) {

	l := make([]*entity.Repository, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(common.MapInt2MapIf(opt.Sort))
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

// DeleteRepository delete a specific entity.Repository from database
func (m *ModelRepository) DeleteRepository(ctx context.Context, projectID, name string) error {
	if projectID == "" || name == "" {
		return fmt.Errorf("can not delete with empty projectID or name")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectID: projectID,
		entity.FieldKeyName:      name,
	})

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return err
	}

	return nil
}

// DeleteRepositories delete a batch of entity.Repository by some projectID and multiple names from database
func (m *ModelRepository) DeleteRepositories(ctx context.Context, projectID string, names []string) error {
	if projectID == "" || len(names) == 0 {
		return fmt.Errorf("can not delete with empty projectID or names")
	}

	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewBranchCondition(operator.And,
		operator.NewLeafCondition(operator.Eq, operator.M{
			entity.FieldKeyProjectID: projectID,
		}),
		operator.NewLeafCondition(operator.In, operator.M{
			entity.FieldKeyName: names,
		}),
	)

	if _, err := m.db.Table(m.tableName).Delete(ctx, cond); err != nil {
		return nil
	}

	return nil
}
