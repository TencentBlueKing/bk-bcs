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

// Package quota xxx
package quota

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/common/page"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/entity"
	utime "github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/util/time"
)

const (
	// table name
	tableName = "quota"
)

var (
	projectQuotaIndexes = []drivers.Index{
		{
			Name: tableName + "_quotaId_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyQuotaId, Value: 1},
			},
			Unique: true,
		},
		{
			Name: tableName + "_quotaName_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyQuotaName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelProjectQuota provide projectQuota db
type ModelProjectQuota struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new project model instance
func New(db drivers.DB) *ModelProjectQuota {
	return &ModelProjectQuota{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   projectQuotaIndexes,
		db:        db,
	}
}

// ensureTable ensure table
func (m *ModelProjectQuota) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := dbtable.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateProjectQuota create project quota
func (m *ModelProjectQuota) CreateProjectQuota(ctx context.Context, projectQuota *ProjectQuota) error {
	if projectQuota == nil {
		return fmt.Errorf("project cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	timestamp := utime.GetStoredTimestamp(time.Now())
	projectQuota.CreateTime = timestamp
	projectQuota.UpdateTime = timestamp

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{projectQuota}); err != nil {
		return err
	}

	return nil
}

// GetProjectQuotaById get projectQuota info by quotaId
func (m *ModelProjectQuota) GetProjectQuotaById(ctx context.Context, quotaId string) (*ProjectQuota, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyQuotaId:   quotaId,
		FieldKeyIsDeleted: false,
	})

	quota := &ProjectQuota{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, quota); err != nil {
		return nil, err
	}
	return quota, nil
}

// UpdateProjectQuota update project quota info
func (m *ModelProjectQuota) UpdateProjectQuota(ctx context.Context, projectQuota *ProjectQuota) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyQuotaId: projectQuota.QuotaId,
	})

	// update project info
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": projectQuota})
}

// DeleteProjectQuota delete projectQuota record
func (m *ModelProjectQuota) DeleteProjectQuota(ctx context.Context, quotaId string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyQuotaId:   quotaId,
		FieldKeyIsDeleted: false,
	})

	oldProjectQuota := &ProjectQuota{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldProjectQuota); err != nil {
		return err
	}

	err := m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": operator.M{
		FieldKeyIsDeleted:  true,
		FieldKeyDeleteTime: time.Now().UTC().Unix(),
	}})
	if err != nil {
		return err
	}
	return nil
}

// ListProjectQuotas query projectQuotas list
func (m *ModelProjectQuota) ListProjectQuotas(ctx context.Context, cond *operator.Condition,
	pagination *page.Pagination) ([]ProjectQuota, int64, error) {
	projectQuotaList := make([]ProjectQuota, 0)

	// 根据cond获取的总量数据
	finder := m.db.Table(m.tableName).Find(cond)
	total, err := finder.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	if len(pagination.Sort) != 0 {
		finder = finder.WithSort(dbtable.MapInt2MapIf(pagination.Sort))
	}
	if pagination.Offset != 0 {
		finder = finder.WithStart(pagination.Offset * pagination.Limit)
	}
	if pagination.Limit == 0 {
		finder = finder.WithLimit(page.DefaultPageLimit)
	} else {
		finder = finder.WithLimit(pagination.Limit)
	}

	// 设置拉取全量数据
	if pagination.All {
		finder = finder.WithLimit(0).WithStart(0)
	}

	// 获取数据
	if err := finder.All(ctx, &projectQuotaList); err != nil {
		return nil, 0, err
	}

	return projectQuotaList, total, nil
}

// UpdateProjectQuotaByField update project quota by field
func (m *ModelProjectQuota) UpdateProjectQuotaByField(
	ctx context.Context, projectQuota entity.M) error {
	if projectQuota == nil {
		return fmt.Errorf("can not update empty project quota")
	}

	if err := m.ensureTable(ctx); err != nil {
		return nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyQuotaId:   projectQuota.GetString(FieldKeyQuotaId),
		FieldKeyIsDeleted: false,
	})
	oldProjectQuota := &ProjectQuota{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldProjectQuota); err != nil {
		return err
	}

	projectQuota[FieldKeyUpdateTime] = time.Now().UTC().Unix()
	return m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": projectQuota})
}

// ListProjectQuotasByProjectId list project quotas by projectId
func (m *ModelProjectQuota) ListProjectQuotasByProjectId(ctx context.Context, projectId string) (
	[]ProjectQuota, error) {
	if projectId == "" {
		return nil, fmt.Errorf("ListProjectQuotasByProject projectId empty")
	}

	projectQuotaList := make([]ProjectQuota, 0)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyProjectId: projectId,
		FieldKeyIsDeleted: false,
	})
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &projectQuotaList)
	if err != nil {
		return nil, err
	}

	return projectQuotaList, nil
}

// ListProjectQuotasByBizId list project quotas by bizId
func (m *ModelProjectQuota) ListProjectQuotasByBizId(ctx context.Context, bizId string) ([]ProjectQuota, error) {
	if bizId == "" {
		return nil, fmt.Errorf("ListProjectQuotasByProject bizId empty")
	}

	projectQuotaList := make([]ProjectQuota, 0)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyBusinessId: bizId,
		FieldKeyIsDeleted:  false,
	})
	err := m.db.Table(m.tableName).Find(cond).All(ctx, &projectQuotaList)
	if err != nil {
		return nil, err
	}

	return projectQuotaList, nil
}
