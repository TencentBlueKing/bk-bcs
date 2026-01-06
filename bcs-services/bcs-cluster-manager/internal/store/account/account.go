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

// Package account xxx
package account

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

const (
	tableName = "cloudaccount"
	// CloudKey xxx
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	CloudKey = "cloudid"
	// AccountIDKey xxx
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

	if account.Account != nil {
		if err := util.EncryptCloudAccountData(nil, account.Account); err != nil {
			blog.Errorf("encrypt cloudAccount %s credential information failed, %s", account.AccountID, err.Error())
			return err
		}
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{account}); err != nil {
		return err
	}
	return nil
}

// UpdateCloudAccount update cloudAccount with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelCloudAccount) UpdateCloudAccount(ctx context.Context,
	account *types.CloudAccount, skipEncrypt bool) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		CloudKey:     account.CloudID,
		AccountIDKey: account.AccountID,
	})
	if account.Account != nil && !skipEncrypt {
		if err := util.EncryptCloudAccountData(nil, account.Account); err != nil {
			blog.Errorf("encrypt cloudAccount %s credential information failed, %s", account.AccountID, err.Error())
			return err
		}
	}

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
func (m *ModelCloudAccount) GetCloudAccount(ctx context.Context,
	cloudID, accountID string, skipDecrypt bool) (*types.CloudAccount, error) {
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
	// 兼容旧数据
	cloudAccount.CreatTime = util.TransStrToUTCStr(time.RFC3339Nano, cloudAccount.CreatTime)
	cloudAccount.UpdateTime = util.TransStrToUTCStr(time.RFC3339Nano, cloudAccount.UpdateTime)
	if cloudAccount.Account != nil && !skipDecrypt {
		if err := util.DecryptCloudAccountData(nil, cloudAccount.Account); err != nil {
			// Compatible with older versions and only output error
			blog.Errorf("decrypt cloudAccount %s credential info failed: %v", accountID, err)
			return nil, err
		}
	}

	return cloudAccount, nil
}

// ListCloudAccount list cloudAccount
func (m *ModelCloudAccount) ListCloudAccount(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]*types.CloudAccount, error) {
	cloudAccountList := make([]*types.CloudAccount, 0)
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

	for _, account := range cloudAccountList {
		if account.Account != nil && !opt.SkipDecrypt {
			if err := util.DecryptCloudAccountData(nil, account.Account); err != nil {
				// Compatible with older versions and only output error
				blog.Errorf("decrypt cloudAccount %s credential failed when ListCloudAccount, %s",
					account.AccountID, err.Error())
				return nil, err
			}
		}
	}

	return cloudAccountList, nil
}
