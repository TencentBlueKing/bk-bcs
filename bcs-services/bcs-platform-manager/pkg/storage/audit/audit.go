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
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/storage/entity"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-platform-manager/pkg/utils"
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
