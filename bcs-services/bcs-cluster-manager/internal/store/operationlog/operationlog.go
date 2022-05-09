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

package operationlog

import (
	"context"
	"fmt"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	types "github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/api/clustermanager"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"

	"go.mongodb.org/mongo-driver/bson"
)

const (
	tableName            = "operationlog"
	defaultLogListLength = 3000

	resourceType = "resourcetype"
	resourceID   = "resourceid"
	taskID       = "taskid"
)

// ModelOperationLog database operation for operation_log
type ModelOperationLog struct {
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	isTableEnsured      bool
	isTableEnsuredMutex sync.RWMutex
}

var (
	operationLogIndexes = []drivers.Index{
		{
			Name: tableName + "_idx",
			Key: bson.D{
				bson.E{Key: resourceType, Value: 1},
				bson.E{Key: resourceID, Value: 1},
				bson.E{Key: taskID, Value: 1},
			},
			Unique: true,
		},
	}
)

// New create operationLog model
func New(db drivers.DB) *ModelOperationLog {
	return &ModelOperationLog{
		tableName: util.DataTableNamePrefix + tableName,
		db:        db,
		indexes:   operationLogIndexes,
	}
}

// ensure table
func (m *ModelOperationLog) ensureTable(ctx context.Context) error {
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

// CreateOperationLog create operation log
func (m *ModelOperationLog) CreateOperationLog(ctx context.Context, log *types.OperationLog) error {
	if log == nil {
		return fmt.Errorf("log to be created cannot be empty")
	}
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	if _, err := m.db.Table(m.tableName).Insert(ctx, []interface{}{log}); err != nil {
		return err
	}
	return nil
}

// DeleteOperationLogByResourceID delete operationLog
func (m *ModelOperationLog) DeleteOperationLogByResourceID(ctx context.Context, resourceIndex string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		resourceID: resourceIndex,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		blog.Warnf("no operationLog delete with resourceID %s", resourceIndex)
	}
	return nil
}

// DeleteOperationLogByResourceType delete operationLog
func (m *ModelOperationLog) DeleteOperationLogByResourceType(ctx context.Context, resType string) error {
	if err := m.ensureTable(ctx); err != nil {
		return err
	}
	cond := operator.NewLeafCondition(operator.Eq, operator.M{
		resourceType: resType,
	})
	deleteCounter, err := m.db.Table(m.tableName).Delete(ctx, cond)
	if err != nil {
		return err
	}
	if deleteCounter == 0 {
		blog.Warnf("no operationLog delete with resourceType %s", resType)
	}
	return nil
}

// ListOperationLog list operationLog
func (m *ModelOperationLog) ListOperationLog(ctx context.Context, cond *operator.Condition, opt *options.ListOption) (
	[]types.OperationLog, error) {

	logList := make([]types.OperationLog, 0)
	finder := m.db.Table(m.tableName).Find(cond)
	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(defaultLogListLength)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if opt.All {
		finder = finder.WithLimit(0)
	}

	if err := finder.All(ctx, &logList); err != nil {
		return nil, err
	}

	return logList, nil
}
