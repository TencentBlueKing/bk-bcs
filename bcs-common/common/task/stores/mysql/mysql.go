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
	"net/url"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type mysqlStore struct {
	dsn       string
	showDebug bool
	db        *gorm.DB
}

// New init mysql iface.Store
func New(dsn string) (iface.Store, error) {
	store := &mysqlStore{dsn: dsn, showDebug: false}
	store.initDsn(dsn)

	// 是否显示sql语句
	level := logger.Warn
	if store.showDebug {
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

// initDsn 解析debug参数是否开启sql显示, 任意异常都原样不动
func (s *mysqlStore) initDsn(raw string) {
	u, err := url.Parse(raw)
	if err != nil {
		return
	}
	query := u.Query()

	// 是否开启debug
	debugStr := query.Get("debug")
	if debugStr != "" {
		debug, err := strconv.ParseBool(debugStr)
		if err != nil {
			return
		}
		s.showDebug = debug
		query.Del("debug")
		u.RawQuery = query.Encode()
	}

	s.dsn = u.String()
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
		Status:        opt.Status,
		CurrentStep:   opt.CurrentStep,
		Creator:       opt.Creator,
	})

	// mysql store 使用创建时间过滤
	if opt.StartGte != nil {
		tx = tx.Where("created_at >= ?", opt.StartGte)
	}
	if opt.StartLte != nil {
		tx = tx.Where("created_at <= ?", opt.StartLte)
	}

	// 只使用id排序
	tx = tx.Order("id DESC")

	result, count, err := FindByPage[TaskRecord](tx, int(opt.Offset), int(opt.Limit))
	if err != nil {
		return nil, err
	}

	items := make([]*types.Task, 0, len(result))
	for _, record := range result {
		items = append(items, toTask(record, []*StepRecord{}))
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
			if step.Name != task.CurrentStep {
				continue
			}
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

// PatchTask implement istore PatchTask interface
func (s *mysqlStore) PatchTask(ctx context.Context, opt *iface.PatchOption) error {
	return types.ErrNotImplemented
}
