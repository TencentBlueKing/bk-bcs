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

package cloud

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "cloud"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey               = "cloudid"
	defaultCloudListLength = 1000
)

var (
	cloudIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: tableKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelCloud database operation for cloud
type ModelCloud struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create cloud model
func New(db drivers.DB) *ModelCloud {
	return &ModelCloud{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelCloud) ensureTable(ctx context.Context) error {
	m.isTableEnsuredMutex.RLock()
	if m.isTableEnsured {
		m.isTableEnsuredMutex.RUnlock()
		return nil
	}
	if err := util.EnsureTable(ctx, m.db, m.tableName, m.indexes); err != nil {
		m.isTableEnsuredMutex.RUnlock()
		return err
	}
	m.isTableEnsuredMutex.RUnlock()

	m.isTableEnsuredMutex.Lock()
	m.isTableEnsured = true
	m.isTableEnsuredMutex.Unlock()
	return nil
}

// CreateCloud create cloud
func (m *ModelCloud) CreateCloud(ctx context.Context, cloud *types.Cloud) error {
	if cloud == nil {
		return fmt.Errorf("cloud to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if cloud.CloudCredential != nil {
		if err := util.EncryptCredential(cloud.CloudCredential); err != nil {
			blog.Errorf("encrypt cloud %s credential information failed, %s", cloud.CloudID, err.Error())
			return err
		}
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{cloud}); err != nil {
		return err
	}
	return nil
}

// UpdateCloud update cloud with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelCloud) UpdateCloud(ctx context.Context, cloud *types.Cloud) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: cloud.CloudID,
	})

	if cloud.CloudCredential != nil {
		if err := util.EncryptCredential(cloud.CloudCredential); err != nil {
			blog.Errorf("encrypt cloud %s credential information failed, %s", cloud.CloudID, err.Error())
			return err
		}
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": cloud})
}

// DeleteCloud delete cloud
func (m *ModelCloud) DeleteCloud(ctx context.Context, cloudID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: cloudID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetCloud get cloud
func (m *ModelCloud) GetCloud(ctx context.Context, cloudID string) (*types.Cloud, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: cloudID,
	})
	cloud := &types.Cloud{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, cloud); err != nil {
		return nil, err
	}

	if cloud.CloudCredential != nil {
		if err := util.DecryptCredential(cloud.CloudCredential); err != nil {
			return nil, err
		}
	}
	return cloud, nil
}

// ListCloud list clusters
func (m *ModelCloud) ListCloud(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.Cloud, error) {
	cloudList := make([]types.Cloud, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultCloudListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &cloudList); err != nil {
		return nil, err
	}
	for _, cloud := range cloudList {
		if cloud.CloudCredential != nil {
			if err := util.DecryptCredential(cloud.CloudCredential); err != nil {
				blog.Errorf("decrypt cloud %s credential failed when ListCloud, %s", cloud.CloudID, err.Error())
				return nil, err
			}
		}
	}
	return cloudList, nil
}
