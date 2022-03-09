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

package namespace

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
	namespaceKeyName                = "name"
	namespaceKeyFederationClusterID = "federationClusterid"
	namespaceTableName              = "namespace"
	defaultNamespaceListLength      = 1000
)

var (
	namespaceClusterIndexes = []drivers.Index{
		{
			Name: namespaceTableName + "_idx",
			Key: bson.D{
				bson.E{Key: namespaceKeyName, Value: 1},
				bson.E{Key: namespaceKeyFederationClusterID, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelNamespace database operation for namespace
type ModelNamespace struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create namespace model
func New(db drivers.DB) *ModelNamespace {
	return &ModelNamespace{
		tableName: util.DataTableNamePrefix + namespaceTableName,
		indexes:   namespaceClusterIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelNamespace) ensureTable(ctx context.Context) error {
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

// CreateNamespace create namespace
func (m *ModelNamespace) CreateNamespace(ctx context.Context, ns *types.Namespace) error {
	if ns == nil {
		return fmt.Errorf("namespace to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{ns}); err != nil {
		return err
	}
	return nil
}

// UpdateNamespace update namespace
func (m *ModelNamespace) UpdateNamespace(ctx context.Context, ns *types.Namespace) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		namespaceKeyName:                ns.Name,
		namespaceKeyFederationClusterID: ns.FederationClusterID,
	})
	oldNs := &types.Namespace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, oldNs); err != nil {
		return err
	}
	return m.db.Table(m.tableName).Update(ctx, cond, operator.M{"$set": ns})
}

// DeleteNamespace delete cluster
func (m *ModelNamespace) DeleteNamespace(ctx context.Context, name, federationClusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		namespaceKeyName:                name,
		namespaceKeyFederationClusterID: federationClusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetNamespace get cluster
func (m *ModelNamespace) GetNamespace(ctx context.Context, name, federationClusterID string) (*types.Namespace, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		namespaceKeyName:                name,
		namespaceKeyFederationClusterID: federationClusterID,
	})
	retNs := &types.Namespace{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, retNs); err != nil {
		return nil, err
	}
	return retNs, nil
}

// ListNamespace list clusters
func (m *ModelNamespace) ListNamespace(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.Namespace, error) {
	retNsList := make([]types.Namespace, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultNamespaceListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &retNsList); err != nil {
		return nil, err
	}
	return retNsList, nil
}
