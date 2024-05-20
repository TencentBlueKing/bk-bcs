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

// Package node xxx
package node

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

const (
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	nodeIDKeyName         = "nodeid"
	nodeIPKeyName         = "innerip"
	nodeClusterIDKey      = "clusterid"
	nodeTableName         = "node"
	nodeNameKey           = "nodename"
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

	blog.Infof("ModelNode UpdateNode[%s:%s]", node.NodeID, node.InnerIP)

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

// UpdateClusterNodeByNodeID update node
func (m *ModelNode) UpdateClusterNodeByNodeID(ctx context.Context, node *types.Node) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode UpdateClusterNodeByNodeID[%s:%s]", node.ClusterID, node.NodeID)

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeIDKeyName:    node.NodeID,
		nodeClusterIDKey: node.ClusterID,
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

	blog.Infof("ModelNode DeleteNode[%s]", nodeID)

	if len(nodeID) == 0 {
		return nil
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

// DeleteClusterNode delete node
func (m *ModelNode) DeleteClusterNode(ctx context.Context, clusterID, nodeID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode DeleteClusterNode[%s:%s]", clusterID, nodeID)

	if len(nodeID) == 0 {
		return nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeIDKeyName:    nodeID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// DeleteClusterNodeByName delete node by name
func (m *ModelNode) DeleteClusterNodeByName(ctx context.Context, clusterID, nodeName string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode DeleteClusterNodeByName[%s:%s]", clusterID, nodeName)

	if len(nodeName) == 0 {
		return nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeNameKey:      nodeName,
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

	blog.Infof("ModelNode DeleteNodesByNodeIDs[%s]", strings.Join(nodeIDs, ","))

	if len(nodeIDs) == 0 {
		return nil
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

// DeleteClusterNodesByIPs delete cluster nodes
func (m *ModelNode) DeleteClusterNodesByIPs(ctx context.Context, clusterID string, ips []string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode DeleteClusterNodesByIPs[%s:%s]", clusterID, strings.Join(ips, ","))

	if len(ips) == 0 || clusterID == "" {
		return nil
	}

	condIP := operator.NewLeafCondition(operator.In, operator.M{nodeIPKeyName: ips})
	condCls := operator.NewLeafCondition(operator.Eq, operator.M{nodeClusterIDKey: clusterID})
	cond := operator.NewBranchCondition(operator.And, condIP, condCls)

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

	blog.Infof("ModelNode DeleteNodesByIPs[%s]", strings.Join(ips, ","))

	if len(ips) == 0 {
		return nil
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

// DeleteNodesByClusterID xxx
func (m *ModelNode) DeleteNodesByClusterID(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode DeleteNodesByClusterID[%s]", clusterID)

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

	blog.Infof("ModelNode DeleteNodeByIP[%s]", ip)

	if len(ip) == 0 {
		return nil
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

// DeleteClusterNodeByIP delete cluster node
func (m *ModelNode) DeleteClusterNodeByIP(ctx context.Context, clusterID, ip string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	blog.Infof("ModelNode DeleteClusterNodeByIP[%s:%s]", clusterID, ip)

	if len(ip) == 0 {
		return nil
	}

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeIPKeyName:    ip,
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

// GetNodeByName get node by name
func (m *ModelNode) GetNodeByName(ctx context.Context, clusterID, name string) (*types.Node, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeNameKey:      name,
	})
	retNode := &types.Node{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNode); err != nil {
		return nil, err
	}
	return retNode, nil
}

// GetClusterNode get cluster node
func (m *ModelNode) GetClusterNode(ctx context.Context, clusterID, nodeID string) (*types.Node, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeIDKeyName:    nodeID,
	})
	retNode := &types.Node{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNode); err != nil {
		return nil, err
	}
	return retNode, nil
}

// GetClusterNodeByIP get node
func (m *ModelNode) GetClusterNodeByIP(ctx context.Context, clusterID, ip string) (*types.Node, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeClusterIDKey: clusterID,
		nodeIPKeyName:    ip,
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
