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
	"time"

	"gorm.io/gorm"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

/**
字段规范:
1. 字段名使用驼峰命名法，表字段使用 _ 分隔
2. bool/int/float/datetime 等类型使用默认字段类型
3. string 类型必须指定类型和长度
4. index 固定varchar(191), (mysql 5.6索引长度限制767byte, utf8mb4下最长191)
**/

var (
	// UnixZeroTime mysql 8.0 版本以上不能写入, 使用unix 0时作为zero time
	// https://dev.mysql.com/doc/refman/8.0/en/datetime.html
	UnixZeroTime = time.Unix(0, 0)
	// SortableFields 可排序字段
	SortableFields = []string{
		"id",
		"created_at",
		"updated_at",
		"deleted_at",
		"task_id",
		"task_type",
		"task_index",
		"task_index_type",
		"task_name",
		"current_step",
		"callback_name",
		"execution_time",
		"max_execution_seconds",
		"start",
		"end",
		"creator",
		"updater",
	}
)

// BaseModel 添加 CreatedAt 索引
type BaseModel struct {
	gorm.Model
	CreatedAt time.Time `gorm:"index"`
}

// TaskRecord 任务记录
type TaskRecord struct {
	BaseModel
	TaskID              string            `json:"taskID" gorm:"type:varchar(191);uniqueIndex:idx_task_id"` // 唯一索引
	TaskType            string            `json:"taskType" gorm:"type:varchar(191);index:idx_task_type"`
	TaskIndex           string            `json:"TaskIndex" gorm:"type:varchar(191);index:idx_task_index"`
	TaskIndexType       string            `json:"TaskIndexType" gorm:"type:varchar(191);index:idx_task_index"`
	TaskName            string            `json:"taskName" gorm:"type:varchar(255)"`
	CurrentStep         string            `json:"currentStep" gorm:"type:varchar(255)"`
	StepSequence        []string          `json:"stepSequence" gorm:"type:text;serializer:json"`
	CallbackName        string            `json:"callbackName" gorm:"type:varchar(255)"`
	CallbackResult      string            `json:"callbackResult" gorm:"type:varchar(191)"`
	CallbackMessage     string            `json:"callbackMessage" gorm:"type:text"`
	CommonParams        map[string]string `json:"commonParams" gorm:"type:text;serializer:json"`
	CommonPayload       string            `json:"commonPayload" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(191);index:idx_status"`
	Message             string            `json:"message" gorm:"type:text"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	Creator             string            `json:"creator" gorm:"type:varchar(255)"`
	Updater             string            `json:"updater" gorm:"type:varchar(255)"`
}

// TableName ..
func (t *TaskRecord) TableName() string {
	return "task_records"
}

// BeforeCreate ..
func (t *TaskRecord) BeforeCreate(tx *gorm.DB) error {
	if t.Start.IsZero() {
		t.Start = UnixZeroTime
	}
	if t.End.IsZero() {
		t.End = UnixZeroTime
	}
	return nil
}

// BeforeUpdate ..
func (t *TaskRecord) BeforeUpdate(tx *gorm.DB) error {
	if t.Start.IsZero() {
		t.Start = UnixZeroTime
	}
	if t.End.IsZero() {
		t.End = UnixZeroTime
	}
	return nil
}

// StepRecord 步骤记录
type StepRecord struct {
	gorm.Model
	TaskID              string            `json:"taskID" gorm:"type:varchar(191);uniqueIndex:idx_task_id_step_name"`
	Name                string            `json:"name" gorm:"type:varchar(191);uniqueIndex:idx_task_id_step_name"`
	Alias               string            `json:"alias" gorm:"type:varchar(255)"`
	Executor            string            `json:"executor" gorm:"type:varchar(255)"`
	Params              map[string]string `json:"input" gorm:"type:text;serializer:json"`
	Payload             string            `json:"payload" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(255)"`
	Message             string            `json:"message" gorm:"type:text"`
	ETA                 *time.Time        `json:"eta"`
	SkipOnFailed        bool              `json:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount"`
	MaxRetries          uint32            `json:"maxRetries"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
}

// TableName ..
func (t *StepRecord) TableName() string {
	return "task_step_records"
}

// BeforeCreate ..
func (t *StepRecord) BeforeCreate(tx *gorm.DB) error {
	if t.Start.IsZero() {
		t.Start = UnixZeroTime
	}
	if t.End.IsZero() {
		t.End = UnixZeroTime
	}
	return nil
}

// BeforeUpdate ..
func (t *StepRecord) BeforeUpdate(tx *gorm.DB) error {
	if t.Start.IsZero() {
		t.Start = UnixZeroTime
	}
	if t.End.IsZero() {
		t.End = UnixZeroTime
	}
	return nil
}

// ToStep 类型转换
func (t *StepRecord) ToStep() *types.Step {
	return &types.Step{
		Name:                t.Name,
		Alias:               t.Alias,
		Executor:            t.Executor,
		Params:              t.Params,
		Payload:             t.Payload,
		Status:              t.Status,
		Message:             t.Message,
		ETA:                 t.ETA,
		SkipOnFailed:        t.SkipOnFailed,
		RetryCount:          t.RetryCount,
		MaxRetries:          t.MaxRetries,
		ExecutionTime:       t.ExecutionTime,
		MaxExecutionSeconds: t.MaxExecutionSeconds,
		Start:               t.Start,
		End:                 t.End,
		LastUpdate:          t.UpdatedAt,
	}
}
