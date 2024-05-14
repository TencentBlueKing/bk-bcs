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

// Package machinery xxx
package machinery

import (
	"context"
	"sync"

	"github.com/Tencent/bk-bcs/bcs-common/common/blog"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/drivers/mongo"
	"github.com/Tencent/bk-bcs/bcs-common/pkg/odm/operator"
	driver "go.mongodb.org/mongo-driver/mongo"

	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/options"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/types"
	"github.com/Tencent/bk-bcs/bcs-services/bcs-cluster-manager/internal/store/util"
)

const (
	// tableName xxx
	tableName = "tasks"
)

// ModelMachineryTask database operation for machinery Task
type ModelMachineryTask struct {
	dbName              string
	tableName           string
	indexes             []drivers.Index
	db                  drivers.DB
	mongoCli            *driver.Client
	isTableEnsured      bool         // nolint
	isTableEnsuredMutex sync.RWMutex // nolint
}

// New create Task model
func New(db drivers.DB, options *mongo.Options) (*ModelMachineryTask, error) {
	cli, err := util.NewMongoCli(options)
	if err != nil {
		return nil, err
	}

	return &ModelMachineryTask{
		dbName:    options.Database,
		tableName: tableName,
		indexes:   nil,
		db:        db,
		mongoCli:  cli,
	}, nil
}

// ensure table
func (m *ModelMachineryTask) ensureTable(ctx context.Context) error { // nolint
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

// ListMachineryTasks list tasks
func (m *ModelMachineryTask) ListMachineryTasks(ctx context.Context, cond *operator.Condition,
	opt *options.ListOption) ([]types.Task, error) {
	taskList := make([]types.Task, 0)

	finder := m.db.Table(m.tableName).Find(cond)

	if len(opt.Sort) != 0 {
		finder = finder.WithSort(util.MapInt2MapIf(opt.Sort))
	}
	if opt.Offset != 0 {
		finder = finder.WithStart(opt.Offset)
	}
	if opt.Limit == 0 {
		finder = finder.WithLimit(util.DefaultLimit)
	} else {
		finder = finder.WithLimit(opt.Limit)
	}

	if opt.All {
		finder = finder.WithLimit(0)
	}

	if err := finder.All(ctx, &taskList); err != nil {
		return nil, err
	}
	return taskList, nil
}

// GetTasksFieldDistinct get tasks distinct field values
func (m *ModelMachineryTask) GetTasksFieldDistinct(ctx context.Context, fieldName string,
	filter interface{}) ([]string, error) {
	table := m.mongoCli.Database(m.dbName).Collection(tableName)

	results, err := table.Distinct(ctx, fieldName, filter)
	if err != nil {
		blog.Errorf("GetTasksFieldDistinct failed: %v", err)
		return nil, err
	}

	return util.SliceInterface2String(results), nil
}
