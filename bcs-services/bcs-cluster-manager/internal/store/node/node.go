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

package node

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
	nodeIDKeyName         = "nodeid"
	nodeIPKeyName         = "innerip"
	nodeClusterIDKey      = "clusterid"
	nodeTableName         = "node"
	defaultNodeListLength = 5000
)

var (
	nodeIndexes = []drivers.Index{
		{
			Name: nodeTableName + "_idx",
			Key: bson.D{
				bson.E{Key: nodeIDKeyName, Value: 1},
				bson.E{Key: nodeIPKeyName, Value: 1},
				bson.E{Key: nodeClusterIDKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelNode database operation for node
type ModelNode struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create node model
func New(db drivers.DB) *ModelNode {
	return &ModelNode{
		tableName: util.DataTableNamePrefix + nodeTableName,
		indexes:   nodeIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelNode) ensureTable(ctx context.Context) error {
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

// CreateNode create node
func (m *ModelNode) CreateNode(ctx context.Context, node *types.Node) error {
	if node == nil {
		return fmt.Errorf("node to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{node}); err != nil {
		return err
	}
	return nil
}

// UpdateNode update node
func (m *ModelNode) UpdateNode(ctx context.Context, node *types.Node) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIDKeyName: node.NodeID,
		nodeIPKeyName: node.InnerIP,
	})
	oldNode := &types.Node{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldNode); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": node})
}

// DeleteNode delete node
func (m *ModelNode) DeleteNode(ctx context.Context, nodeID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIDKeyName: nodeID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodesByNodeIDs delete node
func (m *ModelNode) DeleteNodesByNodeIDs(ctx context.Context, nodeIDs []string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.In, operator.M{
		nodeIDKeyName: nodeIDs,
	})

	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodesByIPs delete node
func (m *ModelNode) DeleteNodesByIPs(ctx context.Context, ips []string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	cond := operator.NewLeafCondition(operator.In, operator.M{
		nodeIPKeyName: ips,
	})

	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodesByClusterID deleteNodes by clusterID
func (m *ModelNode) DeleteNodesByClusterID(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteNodeByIP delete node
func (m *ModelNode) DeleteNodeByIP(ctx context.Context, ip string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIPKeyName: ip,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetNode get node
func (m *ModelNode) GetNode(ctx context.Context, nodeID string) (*types.Node, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIDKeyName: nodeID,
	})
	retNode := &types.Node{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNode); err != nil {
		return nil, err
	}
	return retNode, nil
}

// GetNodeByIP get node
func (m *ModelNode) GetNodeByIP(ctx context.Context, ip string) (*types.Node, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIPKeyName: ip,
	})
	retNode := &types.Node{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNode); err != nil {
		return nil, err
	}
	return retNode, nil
}

// ListNode list nodes
func (m *ModelNode) ListNode(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]*types.Node, error) {
	retNodeList := make([]*types.Node, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultNodeListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retNodeList); err != nil {
		return nil, err
	}
	return retNodeList, nil
}
