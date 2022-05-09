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

package nodegroup

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
	tableName = "nodegroup"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey                   = "nodegroupid"
	clusterIDKey               = "clusterid"
	defaultNodeGroupListLength = 1000
)

var (
	nodeGroupIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: tableKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelNodeGroup database operation for project
type ModelNodeGroup struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create project model
func New(db drivers.DB) *ModelNodeGroup {
	return &ModelNodeGroup{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   nodeGroupIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelNodeGroup) ensureTable(ctx context.Context) error {
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

// CreateNodeGroup create project
func (m *ModelNodeGroup) CreateNodeGroup(ctx context.Context, group *types.NodeGroup) error {
	if group == nil {
		return fmt.Errorf("nodegroup to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{group}); err != nil {
		return err
	}
	return nil
}

// UpdateNodeGroup update nodeGroup with all fields, if some fields are nil
// that field will be overwrite with empty
func (m *ModelNodeGroup) UpdateNodeGroup(ctx context.Context, group *types.NodeGroup) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: group.NodeGroupID,
	})
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": group})
}

// DeleteNodeGroupByClusterID delete nodeGroup by clusterID
func (m *ModelNodeGroup) DeleteNodeGroupByClusterID(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		clusterIDKey: clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodeGroup delete project
func (m *ModelNodeGroup) DeleteNodeGroup(ctx context.Context, nodeGroupID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: nodeGroupID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetNodeGroup get project
func (m *ModelNodeGroup) GetNodeGroup(ctx context.Context, nodeGroupID string) (*types.NodeGroup, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: nodeGroupID,
	})
	group := &types.NodeGroup{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, group); err != nil {
		return nil, err
	}
	return group, nil
}

// ListNodeGroup list clusters
func (m *ModelNodeGroup) ListNodeGroup(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.NodeGroup, error) {
	groups := make([]types.NodeGroup, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultNodeGroupListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &groups); err != nil {
		return nil, err
	}
	return groups, nil
}
