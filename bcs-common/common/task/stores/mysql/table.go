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
3. string 类型必须指定类型和长度，字段是索引的，设置为 varchar(191)
**/

// TaskRecords 任务记录
type TaskRecords struct {
	gorm.Model
	TaskID              string            `json:"taskID" gorm:"type:varchar(255);uniqueIndex:idx_task_id"` // 唯一索引
	TaskType            string            `json:"taskType" gorm:"type:varchar(255)"`
	TaskIndex           string            `json:"TaskIndex" gorm:"type:varchar(255)"`
	TaskIndexType       string            `json:"TaskIndexType" gorm:"type:varchar(255)"`
	TaskName            string            `json:"taskName" gorm:"type:varchar(255)"`
	CurrentStep         string            `json:"currentStep" gorm:"type:varchar(255)"`
	StepSequence        []string          `json:"stepSequence" gorm:"type:text;serializer:json"`
	CallbackName        string            `json:"callbackName" gorm:"type:varchar(255)"`
	CommonParams        map[string]string `json:"commonParams" gorm:"type:text;serializer:json"`
	CommonPayload       []byte            `json:"commonPayload" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(255)"`
	Message             string            `json:"message" gorm:"type:text"`
	ForceTerminate      bool              `json:"forceTerminate"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	Creator             string            `json:"creator" gorm:"type:varchar(255)"`
	Updater             string            `json:"updater" gorm:"type:varchar(255)"`
}

// TableName ..
func (t *TaskRecords) TableName() string {
	return "task_records"
}

// StepRecords 步骤记录
type StepRecords struct {
	gorm.Model
	TaskID              string            `json:"taskID" gorm:"type:varchar(255);index:idx_task_id"` // 索引
	Name                string            `json:"name" gorm:"type:varchar(255)"`
	Alias               string            `json:"alias" gorm:"type:varchar(255)"`
	Executor            string            `json:"executor" gorm:"type:varchar(255)"`
	Params              map[string]string `json:"input" gorm:"type:text;serializer:json"`
	Payload             []byte            `json:"payload" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(255)"`
	Message             string            `json:"message" gorm:"type:varchar(255)"`
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
func (t *StepRecords) TableName() string {
	return "task_step_records"
}

// ToStep 类型转换
func (t *StepRecords) ToStep() *types.Step {
	return &types.Step{
		Name:                t.Name,
		Alias:               t.Alias,
		Executor:            t.Executor,
		Params:              t.Params,
		Payload:             t.Payload,
		Status:              t.Status,
		Message:             t.Message,
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
