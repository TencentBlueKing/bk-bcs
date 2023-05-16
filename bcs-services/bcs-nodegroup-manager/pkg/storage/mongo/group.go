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
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	modelGroupIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: nodeGroupIDKey, Value: 1},
			},
			Name: nodeGroupTableName + "_1",
		},
		{
			Key: bson.D{
				bson.E{Key: nodeGroupIDKey, Value: 1},
			},
			Name: nodeGroupIDKey + "_1",
		},
	}
)

// ModelGroup defines group
type ModelGroup struct {
	Public
}

// NewModelGroup new modelGroup
func NewModelGroup(db drivers.DB) *ModelGroup {
	return &ModelGroup{Public{
		TableName: tableNamePrefix + nodeGroupTableName,
		Indexes:   modelGroupIndexes,
		DB:        db,
	}}
}

// ListNodeGroups list NodeGroups
func (m *ModelGroup) ListNodeGroups(opt *storage.ListOptions) ([]*storage.NodeGroup, error) {
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
			return nil, fmt.Errorf("get group count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	nodeGroupList := make([]*storage.NodeGroup, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nodeGroupIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &nodeGroupList)
	if err != nil {
		return nil, fmt.Errorf("list nodeGroups err:%v", err)
	}
	return nodeGroupList, nil
}

// GetNodeGroup get NodeGroup by nodegroupID
func (m *ModelGroup) GetNodeGroup(nodegroupID string, opt *storage.GetOptions) (*storage.NodeGroup, error) {
	if nodegroupID == "" {
		return nil, fmt.Errorf("nodegroupID is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: nodegroupID,
		isDeletedKey:   opt.GetSoftDeleted,
	})
	retNodeGroup := &storage.NodeGroup{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroup); err != nil {
		// 如果查不到且ErrIfNotExist为false，返回nil，否则返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) && !opt.ErrIfNotExist {
			return nil, nil
		}
		return nil, fmt.Errorf("find nodeGroup error: %v", err)
	}
	return retNodeGroup, nil
}

// CreateNodeGroup create nodeGroup, nodegroupID cannot be empty
func (m *ModelGroup) CreateNodeGroup(nodegroup *storage.NodeGroup, opt *storage.CreateOptions) error {
	if opt == nil || nodegroup.NodeGroupID == "" {
		return fmt.Errorf("CreateOption is nil or nodegroupID is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: nodegroup.NodeGroupID,
		isDeletedKey:   false,
	})
	retNodeGroup := &storage.NodeGroup{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroup); err != nil {
		// 如果查不到，创建
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{nodegroup})
			if err != nil {
				return fmt.Errorf("insert nodeGroup error: %v", err)
			}
			return nil
		}
		return fmt.Errorf("find nodeGroup error: %v", err)
	}

	if !opt.OverWriteIfExist {
		return fmt.Errorf("nodegroup exists")
	}

	// 如果查到，且opt.OverWriteIfExist为true，更新
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": nodegroup}); err != nil {
		return fmt.Errorf("update nodeGroup error: %v", err)
	}
	return nil
}

// UpdateNodeGroup update nodegroup, nodegroupID cannot be empty
func (m *ModelGroup) UpdateNodeGroup(nodegroup *storage.NodeGroup, opt *storage.UpdateOptions) (*storage.NodeGroup,
	error) {
	if opt == nil || nodegroup.NodeGroupID == "" {
		return nil, fmt.Errorf("UpdateOption is nil or nodegroupID is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: nodegroup.NodeGroupID,
		isDeletedKey:   false,
	})
	retNodeGroup := &storage.NodeGroup{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroup); err != nil {
		// 如果找不到且opt.CreateIfNotExist，创建新的
		if errors.Is(err, drivers.ErrTableRecordNotFound) && opt.CreateIfNotExist {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{nodegroup})
			if err != nil {
				return nil, fmt.Errorf("nodegroup does not exist, insert nodeGroup failed: %v", err)
			}
			return nodegroup, nil
		}
		return nil, fmt.Errorf("find nodeGroup failed: %v", err)
	}

	mergeByte, err := MergePatch(retNodeGroup, nodegroup, opt.OverwriteZeroOrEmptyStr)
	if err != nil {
		return nil, fmt.Errorf("mergePatch error:%v", err)
	}
	mergeNodeGroup := &storage.NodeGroup{}
	err = json.Unmarshal(mergeByte, mergeNodeGroup)
	if err != nil {
		return nil, fmt.Errorf("unmarshal mergeNodeGroup error:%v", err)
	}
	// 如果查到,更新
	mergeNodeGroup.UpdatedTime = time.Now()
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": mergeNodeGroup}); err != nil {
		return nil, fmt.Errorf("update nodeGroup error: %v", err)
	}
	return mergeNodeGroup, nil
}

// DeleteNodeGroup delete nodegroup, nodegroupID cannot be empty
func (m *ModelGroup) DeleteNodeGroup(nodegroupID string, opt *storage.DeleteOptions) (*storage.NodeGroup, error) {
	if opt == nil || nodegroupID == "" {
		return nil, fmt.Errorf("DeleteOption is nil or nodegroupID is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: nodegroupID,
		isDeletedKey:   false,
	})
	retNodeGroup := &storage.NodeGroup{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroup); err != nil {
		// 如果找不到且ErrIfNotExist为true，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			if opt.ErrIfNotExist {
				return nil, fmt.Errorf("nodeGroup does not exist")
			}
			// 返回空的strategy
			return nil, nil
		}
		// 返回error
		return nil, fmt.Errorf("find nodeGroup error: %v", err)
	}
	// 如果查到，删除
	if err := m.DB.Table(m.TableName).
		Update(ctx, cond, operator.M{"$set": map[string]interface{}{isDeletedKey: true}}); err != nil {
		return nil, fmt.Errorf("soft delete nodeGroup error: %v", err)
	}
	return retNodeGroup, nil
}
