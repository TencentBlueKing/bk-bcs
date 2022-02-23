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

package cluster

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
	clusterKeyName           = "clusterid"
	clusterTableName         = "cluster"
	defaultClusterListLength = 2000
)

var (
	clusterIndexes = []drivers.Index{
		{
			Name: clusterTableName + "_idx",
			Key: bson.D{
				bson.E{Key: clusterKeyName, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelCluster database operation for cluster
type ModelCluster struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create cluster model
func New(db drivers.DB) *ModelCluster {
	return &ModelCluster{
		tableName: util.DataTableNamePrefix + clusterTableName,
		indexes:   clusterIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelCluster) ensureTable(ctx context.Context) error {
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

// CreateCluster create cluster
func (m *ModelCluster) CreateCluster(ctx context.Context, cluster *types.Cluster) error {
	if cluster == nil {
		return fmt.Errorf("cluster to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{cluster}); err != nil {
		return err
	}
	return nil
}

// UpdateCluster update cluster
func (m *ModelCluster) UpdateCluster(ctx context.Context, cluster *types.Cluster) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		clusterKeyName: cluster.ClusterID,
	})
	oldCluster := &types.Cluster{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldCluster); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": cluster})
}

// DeleteCluster delete cluster
func (m *ModelCluster) DeleteCluster(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		clusterKeyName: clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetCluster get cluster
func (m *ModelCluster) GetCluster(ctx context.Context, clusterID string) (*types.Cluster, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		clusterKeyName: clusterID,
	})
	retCluster := &types.Cluster{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retCluster); err != nil {
		return nil, err
	}
	return retCluster, nil
}

// ListCluster list clusters
func (m *ModelCluster) ListCluster(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.Cluster, error) {
	retClusterList := make([]types.Cluster, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultClusterListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if opt.All {
		finder = finder.WithLimit(0)
	}

	if err := finder.All(ctx, &retClusterList); err != nil {
		return nil, err
	}
	return retClusterList, nil
}
