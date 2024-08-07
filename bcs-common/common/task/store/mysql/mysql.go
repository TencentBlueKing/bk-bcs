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

package mysql

import (
	"context"
	"net/url"
	"strconv"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/store/iface"
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

// EnsureTable 创建db表
func (s *mysqlStore) EnsureTable(ctx context.Context, dst ...any) error {
	// 没有自定义数据, 使用默认表结构
	if len(dst) == 0 {
		dst = []any{&TaskRecords{}, &StepRecords{}}
	}
	return s.db.AutoMigrate(dst...)
}

func (s *mysqlStore) CreateTask(ctx context.Context, task *types.Task) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		record := getTaskRecord(task)
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		steps := getStepRecords(task)
		if err := tx.CreateInBatches(steps, 100).Error; err != nil {
			return err
		}

		return nil
	})
}

func (s *mysqlStore) ListTask(ctx context.Context, opt *iface.ListOption) ([]types.Task, error) {
	return nil, nil
}

func (s *mysqlStore) UpdateTask(ctx context.Context, task *types.Task) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		updateTask := getUpdateTaskRecord(task)
		if err := tx.Model(&TaskRecords{}).Where("task_id = ?", task.TaskID).Updates(updateTask).Error; err != nil {
			return err
		}
		for _, step := range task.Steps {
			if step.Name != task.CurrentStep {
				continue
			}
			updateStep := getUpdateStepRecord(step)
			if err := tx.Where("task_id = ? AND name= ?", task.TaskID, step.Name).Updates(updateStep).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (s *mysqlStore) DeleteTask(ctx context.Context, taskID string) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("task_id = ?", taskID).Delete(&TaskRecords{}).Error; err != nil {
			return err
		}
		if err := tx.Where("task_id = ?", taskID).Delete(&StepRecords{}).Error; err != nil {
			return err
		}
		return nil
	})
}

func (s *mysqlStore) GetTask(ctx context.Context, taskID string) (*types.Task, error) {
	stepRecords := []*StepRecords{}
	if err := s.db.Where("task_id = ?", taskID).Find(&stepRecords).Error; err != nil {
		return nil, err
	}
	taskRecord := TaskRecords{}
	if err := s.db.Where("task_id = ?", taskID).First(&taskRecord).Error; err != nil {
		return nil, err
	}
	return toTask(&taskRecord, stepRecords), nil
}

func (s *mysqlStore) PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	return nil
}
