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

// Package business provide business mongo store
package business

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-project-manager/internal/store/dbtable"
)

const (
	tableName = "business"
	// FieldKeyBusinessID businessID
	FieldKeyBusinessID = "businessID"
	// maxBusinessIDLen CMDB bk_biz_id 转十进制字符串的上限
	maxBusinessIDLen = 20
)

var (
	businessIndexes = []drivers.Index{
		{
			Name: tableName + "_businessID_idx",
			Key: bson.D{
				bson.E{Key: FieldKeyBusinessID, Value: 1},
			},
			Unique: true,
		},
	}
)

// Business CMDB 业务镜像，供其他程序直连 Mongo 消费
type Business struct {
	BusinessID string `json:"businessID" bson:"businessID"`
	Name       string `json:"name" bson:"name"`
	Default    int    `json:"default" bson:"default"`
	Maintainer string `json:"maintainer" bson:"maintainer"`
	Bs2NameID  int    `json:"bs2NameID" bson:"bs2NameID"`
	Productor  string `json:"productor" bson:"productor"`
	Tester     string `json:"tester" bson:"tester"`
	Developer  string `json:"developer" bson:"developer"`
	UpdateTime string `json:"updateTime" bson:"updateTime"`
	SyncTime   string `json:"syncTime" bson:"syncTime"`
}

// ModelBusiness provide business db
type ModelBusiness struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New return a new business model instance
func New(db drivers.DB) *ModelBusiness {
	return &ModelBusiness{
		tableName: dbtable.DataTableNamePrefix + tableName,
		indexes:   businessIndexes,
		db:        db,
	}
}

func (m *ModelBusiness) ensureTable(ctx context.Context) error {
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

// UpsertBusiness upsert business by businessID
func (m *ModelBusiness) UpsertBusiness(ctx context.Context, biz *Business) error {
	if biz == nil {
		return fmt.Errorf("business cannot be empty")
	}
	if err := validateBusinessID(biz.BusinessID); err != nil {
		return err
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		FieldKeyBusinessID: biz.BusinessID,
	})
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": biz})
}

// DeleteBusinessesNotIn hard-delete businesses whose businessID is not in keepIDs
func (m *ModelBusiness) DeleteBusinessesNotIn(ctx context.Context, keepIDs []string) (int64, error) {
	if err := m.ensureTable(ctx); err != nil {
		return 0, err
	}
	if keepIDs == nil {
		keepIDs = []string{}
	}
	cond := operator.NewLeafCondition(operator.Nin, operator.M{
		FieldKeyBusinessID: keepIDs,
	})
	return m.db.Table(m.tableName).Delete(ctx, cond)
}

func validateBusinessID(businessID string) error {
	if businessID == "" {
		return fmt.Errorf("businessID cannot be empty")
	}
	if len(businessID) > maxBusinessIDLen {
		return fmt.Errorf("businessID length exceeds %d", maxBusinessIDLen)
	}
	for _, c := range businessID {
		if c < '0' || c > '9' {
			return fmt.Errorf("businessID must be numeric")
		}
	}
	return nil
}
