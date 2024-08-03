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

	"github.com/Tencent/bk-bcs/bcs-common/common/task/store/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

type mysqlStore struct {
	dsn   string
	debug bool
	db    *gorm.DB
}

// New init mysql iface.Store
func New(dsn string) (iface.Store, error) {
	store := &mysqlStore{dsn: dsn, debug: false}

	u, err := url.Parse(dsn)
	if err != nil {
		return nil, err
	}
	query := u.Query()

	// 是否开启debug
	debugStr := query.Get("debug")
	if debugStr != "" {
		debug, e := strconv.ParseBool(debugStr)
		if e != nil {
			return nil, e
		}
		store.debug = debug
		query.Del("debug")
		u.RawQuery = query.Encode()
	}

	refinedDsn := u.String()
	db, err := gorm.Open(mysql.Open(refinedDsn))
	if err != nil {
		return nil, err
	}
	store.db = db
	if store.debug {
		db.Debug()
	}

	return store, nil
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
		record := GetTaskRecord(task)
		if err := tx.Create(record).Error; err != nil {
			return err
		}

		steps := GetStepRecords(task)
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
	return nil
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
	return ToTask(&taskRecord, stepRecords), nil
}

func (s *mysqlStore) PatchTask(ctx context.Context, taskID string, patchs map[string]interface{}) error {
	return nil
}
