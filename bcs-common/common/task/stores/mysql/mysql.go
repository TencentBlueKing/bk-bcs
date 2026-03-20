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

// Package mysql implemented storage interface.
package mysql

import (
	"context"
	"fmt"
	"slices"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type mysqlStore struct {
	dsn   string
	debug bool
	db    *gorm.DB
}

type option func(*mysqlStore)

// WithDebug 是否显示sql语句
func WithDebug(debug bool) option {
	return func(s *mysqlStore) {
		s.debug = debug
	}
}

// New init mysql iface.Store
func New(dsn string, opts ...option) (iface.Store, error) {
	store := &mysqlStore{dsn: dsn, debug: false}
	for _, opt := range opts {
		opt(store)
	}

	// 是否显示sql语句
	level := logger.Warn
	if store.debug {
		level = logger.Info
	}

	db, err := gorm.Open(mysql.Open(store.dsn),
		&gorm.Config{Logger: logger.Default.LogMode(level)},
	)
	if err != nil {
		return nil, err
	}
	store.db = db

	return store, nil
}

// EnsureTable implement istore EnsureTable interface
func (s *mysqlStore) EnsureTable(ctx context.Context, dst ...any) error {
	// 没有自定义数据, 使用默认表结构
	if len(dst) == 0 {
		dst = []any{&TaskRecord{}, &StepRecord{}}
	}
	return s.db.WithContext(ctx).AutoMigrate(dst...)
}

// CreateTask implement istore CreateTask interface
func (s *mysqlStore) CreateTask(ctx context.Context, task *types.Task) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		record := getTaskRecord(task)
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		steps := getStepRecord(task)
		if err := tx.CreateInBatches(steps, 100).Error; err != nil {
			return err
		}

		return nil
	})
}

// ListStepRecordByTaskIDs implement istore ListStepRecordByTaskIDs interface
func (s *mysqlStore) ListStepRecordByTaskIDs(ctx context.Context, taskIDs []string) ([]*StepRecord, error) {
	db := s.db.WithContext(ctx)

	db = db.Where("task_id IN ?", taskIDs)

	result := make([]*StepRecord, 0)
	if err := db.Find(&result).Error; err != nil {
		return nil, err
	}

	return result, nil
}

// ListTask implement istore ListTask interface
func (s *mysqlStore) ListTask(ctx context.Context, opt *iface.ListOption) (*iface.Pagination[types.Task], error) {
	tx := s.db.WithContext(ctx)

	// 条件过滤 0值gorm自动忽略查询
	tx = tx.Where(&TaskRecord{
		TaskID:        opt.TaskID,
		TaskType:      opt.TaskType,
		TaskName:      opt.TaskName,
		TaskIndex:     opt.TaskIndex,
		TaskIndexType: opt.TaskIndexType,
		CurrentStep:   opt.CurrentStep,
		Creator:       opt.Creator,
	})

	// 状态查询：优先使用StatusList，如果为空则使用Status
	if len(opt.StatusList) > 0 {
		tx = tx.Where("status IN ?", opt.StatusList)
	} else if opt.Status != "" {
		tx = tx.Where("status = ?", opt.Status)
	}

	// mysql store 使用创建时间过滤
	if opt.CreatedGte != nil {
		tx = tx.Where("created_at >= ?", opt.CreatedGte)
	}
	if opt.CreatedLte != nil {
		tx = tx.Where("created_at <= ?", opt.CreatedLte)
	}

	// 排序
	if len(opt.Sort) != 0 {
		for field, direction := range opt.Sort {
			if !slices.Contains(SortableFields, field) {
				return nil, fmt.Errorf("invalid sort field: %s", field)
			}
			if direction > 0 {
				tx = tx.Order(field + " ASC")
			} else {
				tx = tx.Order(field + " DESC")
			}
		}
	} else {
		// 只使用id排序
		tx = tx.Order("id DESC")
	}

	result, count, err := FindByPage[TaskRecord](tx, int(opt.Offset), int(opt.Limit))
	if err != nil {
		return nil, err
	}

	taskIDs := make([]string, 0, len(result))
	for _, record := range result {
		taskIDs = append(taskIDs, record.TaskID)
	}
	stepRecord, err := s.ListStepRecordByTaskIDs(ctx, taskIDs)
	if err != nil {
		return nil, err
	}
	taskStepsMap := make(map[string][]*StepRecord, 0)
	for _, record := range stepRecord {
		if _, ok := taskStepsMap[record.TaskID]; !ok {
			taskStepsMap[record.TaskID] = make([]*StepRecord, 0)
		}
		taskStepsMap[record.TaskID] = append(taskStepsMap[record.TaskID], record)
	}

	items := make([]*types.Task, 0, len(result))
	for _, record := range result {
		items = append(items, toTask(record, taskStepsMap[record.TaskID]))
	}

	return &iface.Pagination[types.Task]{
		Count: count,
		Items: items,
	}, nil
}

// UpdateTask implement istore UpdateTask interface
func (s *mysqlStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		updateTask := getUpdateTaskRecord(task)
		if err := tx.Model(&TaskRecord{}).
			Where("task_id = ?", task.TaskID).
			Select(updateTaskField).
			Updates(updateTask).Error; err != nil {
			return err
		}

		for _, step := range task.Steps {
			updateStep := getUpdateStepRecord(step)
			if err := tx.Model(&StepRecord{}).
				Where("task_id = ? AND name= ?", task.TaskID, step.Name).
				Select(updateStepField).
				Updates(updateStep).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DeleteTask implement istore DeleteTask interface
func (s *mysqlStore) DeleteTask(ctx context.Context, taskID string) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", taskID).Delete(&TaskRecord{}).Error; err != nil {
			return err
		}

		if err := tx.Where("task_id = ?", taskID).Delete(&StepRecord{}).Error; err != nil {
			return err
		}
		return nil
	})
}

// GetTask implement istore GetTask interface
func (s *mysqlStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	tx := s.db.WithContext(ctx)
	taskRecord := TaskRecord{}
	if err := tx.Where("task_id = ?", taskID).First(&taskRecord).Error; err != nil {
		return nil, err
	}

	stepRecord := []*StepRecord{}
	if err := tx.Where("task_id = ?", taskID).Find(&stepRecord).Error; err != nil {
		return nil, err
	}
	return toTask(&taskRecord, stepRecord), nil
}
