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

package resourcequota

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
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	quotaKeyNamespace           = "namespace"
	quotaKeyFederationClusterID = "federationclusterid"
	quotaKeyClusterID           = "clusterid"
	quotaTableName              = "resourcequota"
	defaultQuotaListLength      = 1000
)

var (
	quotaIndexes = []drivers.Index{
		{
			Name: quotaTableName + "_idx",
			Key: bson.D{
				bson.E{Key: quotaKeyNamespace, Value: 1},
				bson.E{Key: quotaKeyFederationClusterID, Value: 1},
				bson.E{Key: quotaKeyClusterID, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelResourceQuota database operation for namespacequota
type ModelResourceQuota struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create namespace model
func New(db drivers.DB) *ModelResourceQuota {
	return &ModelResourceQuota{
		tableName: util.DataTableNamePrefix + quotaTableName,
		indexes:   quotaIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelResourceQuota) ensureTable(ctx context.Context) error {
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

// CreateQuota create namespace quota
func (m *ModelResourceQuota) CreateQuota(ctx context.Context, quota *types.ResourceQuota) error {
	if quota == nil {
		return fmt.Errorf("namespace quota to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{quota}); err != nil {
		return err
	}
	return nil
}

// UpdateQuota update namespace quota
func (m *ModelResourceQuota) UpdateQuota(ctx context.Context, quota *types.ResourceQuota) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		quotaKeyNamespace:           quota.Namespace,
		quotaKeyFederationClusterID: quota.FederationClusterID,
		quotaKeyClusterID:           quota.ClusterID,
	})
	oldQuota := &types.ResourceQuota{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldQuota); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": quota})
}

// DeleteQuota delete namespace quota
func (m *ModelResourceQuota) DeleteQuota(ctx context.Context, namespace, federationClusterID, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		quotaKeyNamespace:           namespace,
		quotaKeyFederationClusterID: federationClusterID,
		quotaKeyClusterID:           clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// BatchDeleteQuotaByCluster delete namespace quota by cluster
func (m *ModelResourceQuota) BatchDeleteQuotaByCluster(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		quotaKeyClusterID: clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetQuota get namespace quota
func (m *ModelResourceQuota) GetQuota(ctx context.Context, namespace, federationClusterID, clusterID string) (
	*types.ResourceQuota, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		quotaKeyNamespace:           namespace,
		quotaKeyFederationClusterID: federationClusterID,
		quotaKeyClusterID:           clusterID,
	})
	retNsQuota := &types.ResourceQuota{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNsQuota); err != nil {
		return nil, err
	}
	return retNsQuota, nil
}

// ListQuota list clusters
func (m *ModelResourceQuota) ListQuota(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.ResourceQuota, error) {

	retNsQuotaList := make([]types.ResourceQuota, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultQuotaListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retNsQuotaList); err != nil {
		return nil, err
	}
	return retNsQuotaList, nil
}
