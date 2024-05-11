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

package mongo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-nodegroup-manager/pkg/storage"
)

var (
	modelActionIndexes = []drivers.Index{
		{
			Key: bson.D{
				bson.E{Key: "_id", Value: 1},
				bson.E{Key: nodeGroupIDKey, Value: 1},
				bson.E{Key: clusterIDKey, Value: 1},
				bson.E{Key: eventKey, Value: 1},
			},
			Name:   actionTableName + "_1",
			Unique: true,
		},
	}
)

// ModelAction defines action
type ModelAction struct {
	Public
}

// NewModelAction new ModelAction
func NewModelAction(db drivers.DB) *ModelAction {
	return &ModelAction{Public{
		TableName: tableNamePrefix + actionTableName,
		Indexes:   modelActionIndexes,
		DB:        db,
	}}
}

// ListNodeGroupAction list NodeGroupAction by nodeGroupID, if nodeGroupID is empty, return all
func (m *ModelAction) ListNodeGroupAction(nodeGroupID string,
	opt *storage.ListOptions) ([]*storage.NodeGroupAction, error) {
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
	if nodeGroupID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			nodeGroupIDKey: nodeGroupID,
		}))
	}
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	if !opt.DoPagination && opt.Limit == 0 {
		// nolint
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get action count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	nodeGroupActionList := make([]*storage.NodeGroupAction, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nodeGroupIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &nodeGroupActionList)
	if err != nil {
		return nil, err
	}
	return nodeGroupActionList, nil
}

// GetNodeGroupAction get NodeGroupAction by nodegroupID and event
func (m *ModelAction) GetNodeGroupAction(nodeGroupID, event string, opt *storage.GetOptions) (*storage.NodeGroupAction,
	error) {
	if nodeGroupID == "" || event == "" {
		return nil, fmt.Errorf("request param illegal, nodegroupID and event cannot be empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: nodeGroupID,
		eventKey:       event,
		isDeletedKey:   opt.GetSoftDeleted,
	})
	retNodeGroupAction := &storage.NodeGroupAction{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroupAction); err != nil {
		// 如果查不到，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) && !opt.ErrIfNotExist {
			return nil, nil
		}
		return nil, fmt.Errorf("find nodeGroup error : %v", err)
	}
	return retNodeGroupAction, nil
}

// CreateNodeGroupAction create NodeGroupAction, nodegroupID, clusterID and event cannot be empty
func (m *ModelAction) CreateNodeGroupAction(action *storage.NodeGroupAction, opt *storage.CreateOptions) error {
	if opt == nil || action.NodeGroupID == "" || action.Event == "" {
		return fmt.Errorf("CreateOption is nil or nodegroupID/event is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: action.NodeGroupID,
		clusterIDKey:   action.ClusterID,
		eventKey:       action.Event,
		isDeletedKey:   false,
	})
	retAction := &storage.NodeGroupAction{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retAction); err != nil {
		// 如果查不到，创建
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{action})
			if err != nil {
				return fmt.Errorf("insert nodeGroupAction failed: %v", err)
			}
			return nil
		}
		return fmt.Errorf("find nodeGroupAction failed: %v", err)
	}
	if !opt.OverWriteIfExist {
		return fmt.Errorf("nodegroupAction exists")
	}
	// 如果查到，且opt.OverWriteIfExist为true，更新
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": action}); err != nil {
		return fmt.Errorf("overwrite nodeGroupActioin err: %v", err)
	}
	return nil
}

// UpdateNodeGroupAction update NodeGroupAction, nodegroupID and event can not be empty
func (m *ModelAction) UpdateNodeGroupAction(action *storage.NodeGroupAction,
	opt *storage.UpdateOptions) (*storage.NodeGroupAction, error) {
	if opt == nil || action.NodeGroupID == "" || action.Event == "" {
		return nil, fmt.Errorf("UpdateOption is nil or nodegroupID/event is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		nodeGroupIDKey: action.NodeGroupID,
		isDeletedKey:   false,
	})
	retAction := &storage.NodeGroupAction{}
	if err = m.DB.Table(m.TableName).Find(cond).One(ctx, retAction); err != nil {
		// 如果找不到且opt.CreateIfNotExist，创建新的
		if errors.Is(err, drivers.ErrTableRecordNotFound) && opt.CreateIfNotExist {
			_, err = m.DB.Table(m.TableName).Insert(ctx, []interface{}{action})
			if err != nil {
				return nil, fmt.Errorf("nodeGroupAction does not exist, insert error: %v", err)
			}
			return action, nil
		}
		return nil, fmt.Errorf("find nodeGroupAction error: %v", err)
	}

	mergeByte, err := MergePatch(retAction, action, opt.OverwriteZeroOrEmptyStr)
	if err != nil {
		return nil, fmt.Errorf("mergePatch error:%v", err)
	}
	mergeNodeGroupAction := &storage.NodeGroupAction{}
	err = json.Unmarshal(mergeByte, mergeNodeGroupAction)
	if err != nil {
		return nil, fmt.Errorf("unmarshal mergeNodeGroupAction error:%v", err)
	}
	// 如果查到,更新
	mergeNodeGroupAction.UpdatedTime = time.Now()
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": mergeNodeGroupAction}); err != nil {
		return nil, fmt.Errorf("update nodeGroup error: %v", err)
	}
	return mergeNodeGroupAction, nil
}

// DeleteNodeGroupAction soft delete NodeGroupAction, nodegroupID and event cannot be empty
func (m *ModelAction) DeleteNodeGroupAction(action *storage.NodeGroupAction,
	opt *storage.DeleteOptions) (*storage.NodeGroupAction, error) {
	if opt == nil || action.NodeGroupID == "" || action.Event == "" {
		return nil, fmt.Errorf("DeleteOption is nil or nodegroupID/event is empty")
	}
	ctx := context.Background()
	err := ensureTable(ctx, &m.Public)
	if err != nil {
		return nil, err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		taskIDKey:      action.TaskID,
		nodeGroupIDKey: action.NodeGroupID,
		eventKey:       action.Event,
		isDeletedKey:   false,
	})
	retNodeGroupAction := &storage.NodeGroupAction{}
	if err := m.DB.Table(m.TableName).Find(cond).One(ctx, retNodeGroupAction); err != nil {
		// 如果找不到且ErrIfNotExist为true，返回error
		if errors.Is(err, drivers.ErrTableRecordNotFound) {
			if opt.ErrIfNotExist {
				return nil, fmt.Errorf("find nodeGroupAction error: %v", err)
			}
			// 返回空的strategy
			return nil, nil
		}
		// 返回error
		return nil, fmt.Errorf("find nodeGroupAction error: %v", err)
	}
	// 如果查到，删除
	retNodeGroupAction.IsDeleted = true
	retNodeGroupAction.UpdatedTime = time.Now()
	if err := m.DB.Table(m.TableName).Update(ctx, cond, operator.M{"$set": retNodeGroupAction}); err != nil {
		return nil, fmt.Errorf("soft delete nodeGroupAction error: %v", err)
	}
	return retNodeGroupAction, nil
}

// ListNodeGroupActionByTaskID list NodeGroupAction by taskID, if nodeGroupID is empty, return all
func (m *ModelAction) ListNodeGroupActionByTaskID(taskID string,
	opt *storage.ListOptions) ([]*storage.NodeGroupAction, error) {
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
	if taskID != "" {
		cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
			taskIDKey: taskID,
		}))
	}
	cond = append(cond, operator.NewLeafCondition(operator.Eq, operator.M{
		isDeletedKey: opt.ReturnSoftDeletedItems,
	}))
	if !opt.DoPagination && opt.Limit == 0 {
		// nolint
		count, err := m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).Count(ctx)
		if err != nil {
			return nil, fmt.Errorf("get action count err:%v", err)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	nodeGroupActionList := make([]*storage.NodeGroupAction, 0)
	err = m.DB.Table(m.TableName).Find(operator.NewBranchCondition(operator.And, cond...)).
		WithSort(map[string]interface{}{nodeGroupIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &nodeGroupActionList)
	if err != nil {
		return nil, err
	}
	return nodeGroupActionList, nil
}

// ListNodeGroupActionByEvent list NodeGroupAction by nodeGroupID, if nodeGroupID is empty, return all
func (m *ModelAction) ListNodeGroupActionByEvent(event string,
	opt *storage.ListOptions) ([]*storage.NodeGroupAction, error) {
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

	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		eventKey:     event,
		isDeletedKey: opt.ReturnSoftDeletedItems,
	})
	if !opt.DoPagination && opt.Limit == 0 {
		count, countErr := m.DB.Table(m.TableName).Find(cond).Count(ctx)
		if countErr != nil {
			return nil, fmt.Errorf("get action count err:%v", countErr)
		}
		limit = int(count)
	} else if limit == 0 {
		limit = defaultSize
	}
	nodeGroupActionList := make([]*storage.NodeGroupAction, 0)
	err = m.DB.Table(m.TableName).Find(cond).
		WithSort(map[string]interface{}{nodeGroupIDKey: 1}).
		WithStart(int64(page*limit)).WithLimit(int64(limit)).All(ctx, &nodeGroupActionList)
	if err != nil {
		return nil, err
	}
	return nodeGroupActionList, nil
}
