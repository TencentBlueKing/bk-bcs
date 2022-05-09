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

package clustercredential

import (
	"context"
	"errors"
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
	clusterCredentialTableName = "clustercredential"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	credentialKeyName                  = "serverkey"
	defaultClusterCredentialListLength = 1000
)

var (
	clusterCredentialIndexes = []drivers.Index{
		{
			Name: clusterCredentialTableName + "_idx",
			Key: bson.D{
				bson.E{Key: credentialKeyName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelClusterCredential database operation for online cluster
type ModelClusterCredential struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create online cluster model
func New(db drivers.DB) *ModelClusterCredential {
	return &ModelClusterCredential{
		tableName: util.DataTableNamePrefix + clusterCredentialTableName,
		indexes:   clusterCredentialIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelClusterCredential) ensureTable(ctx context.Context) error {
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

// PutClusterCredential put online cluster
func (m *ModelClusterCredential) PutClusterCredential(
	ctx context.Context, clusterCredential *types.ClusterCredential) error {
	if clusterCredential == nil {
		return fmt.Errorf("cluster credential cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		credentialKeyName: clusterCredential.ServerKey,
	})
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": clusterCredential})
}

// DeleteClusterCredential delete online cluster
func (m *ModelClusterCredential) DeleteClusterCredential(ctx context.Context, serverKey string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		credentialKeyName: serverKey,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetClusterCredential get online cluster
func (m *ModelClusterCredential) GetClusterCredential(ctx context.Context, serverKey string) (
	*types.ClusterCredential, bool, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, false, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		credentialKeyName: serverKey,
	})
	retCluster := &types.ClusterCredential{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retCluster); err != nil {
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			return nil, false, nil
		}
		return nil, false, err
	}
	return retCluster, true, nil
}

// ListClusterCredential list online clusters
func (m *ModelClusterCredential) ListClusterCredential(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.ClusterCredential, error) {
	retClusterCredentialList := make([]types.ClusterCredential, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultClusterCredentialListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retClusterCredentialList); err != nil {
		return nil, err
	}
	return retClusterCredentialList, nil
}
