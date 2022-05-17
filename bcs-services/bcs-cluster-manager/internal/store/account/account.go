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

package account

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName = "cloudaccount"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	CloudKey                      = "cloudid"
	AccountIDKey                  = "accountid"
	defaultCloudAccountListLength = 4000
)

var (
	cloudIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: CloudKey, Value: 1},
				bson.E{Key: AccountIDKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelCloudAccount database operation for cloudAccount
type ModelCloudAccount struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create cloudAccount model
func New(db drivers.DB) *ModelCloudAccount {
	return &ModelCloudAccount{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   cloudIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelCloudAccount) ensureTable(ctx context.Context) error {
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

// CreateCloudAccount insert cloudAccount
func (m *ModelCloudAccount) CreateCloudAccount(ctx context.Context, account *types.CloudAccount) error {
	if account == nil {
		return fmt.Errorf("cloudAccount to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{account}); err != nil {
		return err
	}
	return nil
}

// UpdateCloudAccount update cloudAccount with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelCloudAccount) UpdateCloudAccount(ctx context.Context, account *types.CloudAccount) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		CloudKey:     account.CloudID,
		AccountIDKey: account.AccountID,
	})

	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": account})
}

// DeleteCloudAccount delete cloudAccount
func (m *ModelCloudAccount) DeleteCloudAccount(ctx context.Context, cloudID string, accountID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		CloudKey:     cloudID,
		AccountIDKey: accountID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetCloudAccount get cloudAccount
func (m *ModelCloudAccount) GetCloudAccount(ctx context.Context, cloudID, accountID string) (*types.CloudAccount, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		CloudKey:     cloudID,
		AccountIDKey: accountID,
	})
	cloudAccount := &types.CloudAccount{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, cloudAccount); err != nil {
		return nil, err
	}

	return cloudAccount, nil
}

// ListCloudAccount list cloudAccount
func (m *ModelCloudAccount) ListCloudAccount(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.CloudAccount, error) {
	cloudAccountList := make([]types.CloudAccount, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultCloudAccountListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &cloudAccountList); err != nil {
		return nil, err
	}

	return cloudAccountList, nil
}
