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

// Package audit audit
package audit

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
	tableName = "audit"
)

var (
	tableIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: entity.FieldKeyProjectCode, Value: 1},
				bson.E{Key: entity.FieldKeyClusterID, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelAudit provides handling audit operations to database
type ModelAudit struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.Mutex
}

// New return a new ModelLogRule instance
func New(db drivers.DB) *ModelAudit {
	return &ModelAudit{
		tableName: utils.DataTableNamePrefix + tableName,
		indexes:   tableIndexes,
		db:        db,
	}
}

func (m *ModelAudit) ensureTable(ctx context.Context) error {
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

// CreateAudit create audit
func (m *ModelAudit) CreateAudit(ctx context.Context, lc *entity.Audit) (primitive.ObjectID, error) {
	if err := m.ensureTable(ctx); err != nil {
		return primitive.NilObjectID, err
	}

	now := utils.JSONTime{Time: time.Now()}
	lc.CreatedAt = now
	lc.UpdatedAt = now
	if lc.ID.IsZero() {
		lc.ID = primitive.NewObjectIDFromTimestamp(now.Time)
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{lc}); err != nil {
		return primitive.NilObjectID, err
	}
	return lc.ID, nil
}

// UpdateAudit update audit
func (m *ModelAudit) UpdateAudit(ctx context.Context, id string, lc entity.M) error {
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

// DeleteAudit delete audit
func (m *ModelAudit) DeleteAudit(ctx context.Context, id string) error {
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

// GetAudit get audit by projectCode and clusterID
func (m *ModelAudit) GetAudit(ctx context.Context, projectCode, clusterID string) (*entity.Audit, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		entity.FieldKeyProjectCode: projectCode,
		entity.FieldKeyClusterID:   clusterID,
	})

	audit := &entity.Audit{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, audit); err != nil {
		return nil, err
	}

	return audit, nil
}

// FirstAuditOrCreate get audit by projectCode and clusterID, if not found, create a new one
func (m *ModelAudit) FirstAuditOrCreate(ctx context.Context, audit *entity.Audit) (*entity.Audit, error) {
	getAudit, err := m.GetAudit(ctx, audit.ProjectCode, audit.ClusterID)
	if err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			id, cerr := m.CreateAudit(ctx, audit)
			if cerr != nil {
				return nil, cerr
			}
			audit.ID = id
			return audit, nil
		}
		return nil, err
	}

	return getAudit, nil
}
