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

package scalingoption

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
	tableName = "clusterautoscalingoption"
	//! we don't setting bson tag in proto file,
	//! all struct key in mongo is lowcase in default
	tableKey                = "clusterid"
	defaultOptionListLength = 1000
)

var (
	scalingOptionIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: tableKey, Value: 1},
			},
			Unique: true,
		},
	}
)

// ModelAutoScalingOption database operation for option
type ModelAutoScalingOption struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

// New create ClusterAutoScalingOption model
func New(db drivers.DB) *ModelAutoScalingOption {
	return &ModelAutoScalingOption{
		tableName: util.DataTableNamePrefix + tableName,
		indexes:   scalingOptionIndexes,
		db:        db,
	}
}

// ensure table
func (m *ModelAutoScalingOption) ensureTable(ctx context.Context) error {
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

// CreateAutoScalingOption create cluster autoscaling option
func (m *ModelAutoScalingOption) CreateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error {
	if option == nil {
		return fmt.Errorf("ClusterAutoScalingOption to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}

	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{option}); err != nil {
		return err
	}
	return nil
}

// UpdateAutoScalingOption update option with all fileds, if some fields are nil
// that field will be overwrite with empty
func (m *ModelAutoScalingOption) UpdateAutoScalingOption(ctx context.Context, option *types.ClusterAutoScalingOption) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: option.ClusterID,
	})
	return m.db.Table(m.tableName).Upsert(ctx, cond, operator.M{"$set": option})
}

// DeleteAutoScalingOption delete Cluster AutoScaling option
func (m *ModelAutoScalingOption) DeleteAutoScalingOption(ctx context.Context, clusterID string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: clusterID,
	})
	_, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	return nil
}

// GetAutoScalingOption get option
func (m *ModelAutoScalingOption) GetAutoScalingOption(ctx context.Context, clusterID string) (*types.ClusterAutoScalingOption, error) {
	if err := m.ensureTable(ctx); err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		tableKey: clusterID,
	})
	option := &types.ClusterAutoScalingOption{}
	if err := m.db.Table(m.tableName).Find(cond).One(ctx, option); err != nil {
		return nil, err
	}
	return option, nil
}

// ListAutoScalingOption list cluster autoscaling option according search condition
func (m *ModelAutoScalingOption) ListAutoScalingOption(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.ClusterAutoScalingOption, error) {
	optionList := make([]types.ClusterAutoScalingOption, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultOptionListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}
	if err := finder.All(ctx, &optionList); err != nil {
		return nil, err
	}
	return optionList, nil
}
