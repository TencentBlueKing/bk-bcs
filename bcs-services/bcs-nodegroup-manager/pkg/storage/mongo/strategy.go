/*
 * Tencent is pleased to support the open source community by making Blueking Container Service available.
 *  Copyright (C) 2019 THL A29 Limited, a Tencent company. All rights reserved.
 *  Licensed under the MIT License (the "License"); you may not use this file except
 *  in compliance with the License. You may obtain a copy of the License at
 *  http://opensource.org/licenses/MIT
 *  Unless required by applicable law or agreed to in writing, software distributed under
 *  the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 *  either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	modelStrategyIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: nameKey, Value: 1},
			},
			Unique: true,
			Name:   strategyTableName + "_1",
		},
	}
)

// ModelStrategy defines strategy
type ModelStrategy struct {
	Public
}

// NewModelStrategy returns a new ModelStrategy
func NewModelStrategy(db drivers.DB) *ModelStrategy {
	return &ModelStrategy{Public{
		TableName: tableNamePrefix + strategyTableName,
		Indexes:   modelStrategyIndexes,
		DB:        db,
	}}
}

// ListNodeGroupStrategies 查询NodeGroupMgrStrategy列表，可设置page和limit，page默认值为0，limit默认值为10
func (m *ModelStrategy) ListNodeGroupStrategies(opt *storage.ListOptions) ([]*storage.NodeGroupMgrStrategy, error) {
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	page := opt.Page
	limit := opt.Limit

	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	if !opt.DoPagination && opt.Limit == 0 {
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get strategy count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	strategyList := make([]*storage.NodeGroupMgrStrategy, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nameKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &strategyList)
	if err != nil {
		return nil, fmt.Errorf("list nodeGroupMgrStrategy err:%v", err)
	}
	return strategyList, nil
}

// ListNodeGroupStrategiesByType 通过类型查询NodeGroupMgrStrategy列表，可设置page和limit，page默认值为0，limit默认值为10
func (m *ModelStrategy) ListNodeGroupStrategiesByType(strategyType string,
	opt *storage.ListOptions) ([]*storage.NodeGroupMgrStrategy, error) {
	if opt == nil {
		return nil, fmt.Errorf("ListOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	page := opt.Page
	limit := opt.Limit

	cond := make([]*operator.Condition, 0)
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		strategyTypeKey: strategyType,
		isDeletedKey:    opt.ReturnSoftDeletedItems,
	}))
	if !opt.DoPagination && opt.Limit == 0 {
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get strategy count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	strategyList := make([]*storage.NodeGroupMgrStrategy, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nameKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &strategyList)
	if err != nil {
		return nil, fmt.Errorf("list nodeGroupMgrStrategy err:%v", err)
	}
	return strategyList, nil
}

// GetNodeGroupStrategy 通过name查询单个NodeGroupMgrStrategy
func (m *ModelStrategy) GetNodeGroupStrategy(name string, opt *storage.GetOptions) (*storage.NodeGroupMgrStrategy,
	error) {
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nameKey:      name,
		isDeletedKey: opt.GetSoftDeleted,
	})
	retStrategy := &storage.NodeGroupMgrStrategy{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retStrategy); err != nil {
		blog.Infof("nodeGroupMgrStrategy[%s] not exist", name)
		// 如果查不到，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) && !opt.ErrIfNotExist {
			return nil, nil
		}
		return nil, fmt.Errorf("find nodeGroupMgrStrategy error: %v", err)
	}
	return retStrategy, nil
}

// CreateNodeGroupStrategy 创建NodeGroupStrategy，如果CreateOptions.OverWriteIfExist为true，更新已存在的内容
func (m *ModelStrategy) CreateNodeGroupStrategy(strategy *storage.NodeGroupMgrStrategy,
	opt *storage.CreateOptions) error {
	if opt == nil {
		return fmt.Errorf("CreateOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nameKey:      strategy.Name,
		isDeletedKey: false,
	})
	retStrategy := &storage.NodeGroupMgrStrategy{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retStrategy); err != nil {
		// 如果查不到，创建
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{strategy})
			if err != nil {
				return fmt.Errorf("nodeGroupMgrStrategy does not exist, insert error: %v", err)
			}
			return nil
		}
		return fmt.Errorf("find nodeGroupMgrStrategy error: %v", err)
	}
	// 如果查到，且opt.OverWriteIfExist为true，更新
	if !opt.OverWriteIfExist {
		return fmt.Errorf("nodeGroupMgrStrategy exists")
	}
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": strategy}); err != nil {
		return fmt.Errorf("update nodeGroupMgrStrategy error: %v", err)
	}
	return nil
}

// UpdateNodeGroupStrategy 更新NodeGroupMgrStrategy，如果opt.CreateIfNotExist为true且不存在，创建新的，
func (m *ModelStrategy) UpdateNodeGroupStrategy(strategy *storage.NodeGroupMgrStrategy,
	opt *storage.UpdateOptions) (*storage.NodeGroupMgrStrategy, error) {
	if opt == nil {
		return nil, fmt.Errorf("UpdateOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nameKey:      strategy.Name,
		isDeletedKey: false,
	})
	retStrategy := &storage.NodeGroupMgrStrategy{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retStrategy); err != nil {
		// 如果找不到且opt.CreateIfNotExist，创建新的
		if errors.Is(err, drivers.ErrTableRecordNotFound) && opt.CreateIfNotExist {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{strategy})
			if err != nil {
				return nil, fmt.Errorf("nodeGroupMgrStrategy does not exist, insert error: %v", err)
			}
			return strategy, nil
		}
		return nil, fmt.Errorf("find nodeGroupMgrStrategy error: %v", err)
	}
	mergeByte, err := MergePatch(retStrategy, strategy, opt.OverwriteZeroOrEmptyStr)
	if err != nil {
		return nil, fmt.Errorf("merge nodeGroupMgrStrategy error:%v", err)
	}
	mergeStrategy := &storage.NodeGroupMgrStrategy{}
	err = json.Unmarshal(mergeByte, mergeStrategy)
	if err != nil {
		return nil, fmt.Errorf("unmarshal mergeStrategy error:%v", err)
	}
	// 如果查到,更新
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": mergeStrategy}); err != nil {
		return nil, fmt.Errorf("update nodeGroupMgrStrategy error: %v", err)
	}
	return mergeStrategy, nil
}

// DeleteNodeGroupStrategy 删除NodeGroupMgrStrategy，如果找不到且opt.ErrIfNotExist为true，返回error
func (m *ModelStrategy) DeleteNodeGroupStrategy(name string, opt *storage.DeleteOptions) (*storage.NodeGroupMgrStrategy,
	error) {
	if opt == nil {
		return nil, fmt.Errorf("DeleteOption is nil")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nameKey:      name,
		isDeletedKey: false,
	})
	retStrategy := &storage.NodeGroupMgrStrategy{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retStrategy); err != nil {
		// 如果找不到且ErrIfNotExist为true，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			if opt.ErrIfNotExist {
				return nil, fmt.Errorf("nodeGroupMgrStrategy does not exist")
			}
			// 返回nil
			return nil, nil
		}
		// 返回error
		return nil, fmt.Errorf("find nodeGroupMgrStrategy error: %v", err)
	}
	// 如果查到，删除
	if err := m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": map[string]interface{}{isDeletedKey: true}}); err != nil {
		return nil, fmt.Errorf("soft delete nodeGroupStrategy error: %v", err)
	}
	return retStrategy, nil
}
