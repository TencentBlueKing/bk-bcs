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
	TaskName            string            `json:"taskName" gorm:"type:varchar(255)"`
	CurrentStep         string            `json:"currentStep" gorm:"type:varchar(255)"`
	StepSequence        []string          `json:"stepSequence" gorm:"type:text;serializer:json"`
	CallBackFuncName    string            `json:"callBackFuncName" gorm:"type:varchar(255)"`
	CommonParams        map[string]string `json:"commonParams" gorm:"type:text;serializer:json"`
	ExtraJson           string            `json:"extraJson" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(255)"`
	Message             string            `json:"message" gorm:"type:text"`
	ForceTerminate      bool              `json:"forceTerminate"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	LastUpdate          time.Time         `json:"lastUpdate"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
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
	Input               map[string]string `json:"input" gorm:"type:text;serializer:json"`
	Output              map[string]string `json:"output" gorm:"type:text;serializer:json"`
	Extras              string            `json:"extras" gorm:"type:text"`
	Status              string            `json:"status" gorm:"type:varchar(255)"`
	Message             string            `json:"message" gorm:"type:varchar(255)"`
	SkipOnFailed        bool              `json:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	LastUpdate          time.Time         `json:"lastUpdate"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
}

// TableName ..
func (t *StepRecords) TableName() string {
	return "task_step_records"
}

func GetStepRecords(t *types.Task) []*StepRecords {

	records := make([]*StepRecords, 0, len(t.Steps))
	for _, step := range t.Steps {
		record := &StepRecords{
			TaskID:              t.TaskID,
			Name:                step.Name,
			Alias:               step.Alias,
			Extras:              step.Extras,
			Status:              step.Status,
			Message:             step.Message,
			SkipOnFailed:        step.SkipOnFailed,
			RetryCount:          step.RetryCount,
			Start:               step.Start,
			End:                 step.End,
			ExecutionTime:       step.ExecutionTime,
			MaxExecutionSeconds: step.MaxExecutionSeconds,
			LastUpdate:          step.LastUpdate,
		}
		records = append(records, record)
	}

	return records
}
func GetTaskRecord(t *types.Task) *TaskRecords {
	stepSequence := make([]string, 0, len(t.Steps))
	for i := range t.Steps {
		stepSequence = append(stepSequence, t.Steps[i].Name)
	}

	record := &TaskRecords{
		TaskID:              t.TaskID,
		TaskType:            t.TaskType,
		TaskName:            t.TaskName,
		CurrentStep:         t.CurrentStep,
		StepSequence:        stepSequence,
		CallBackFuncName:    t.CallBackFuncName,
		CommonParams:        t.CommonParams,
		ExtraJson:           t.ExtraJson,
		Status:              t.Status,
		Message:             t.Message,
		ForceTerminate:      t.ForceTerminate,
		Start:               t.Start,
		End:                 t.End,
		ExecutionTime:       t.ExecutionTime,
		MaxExecutionSeconds: t.MaxExecutionSeconds,
		Creator:             t.Creator,
		LastUpdate:          t.LastUpdate,
		Updater:             t.Updater,
	}
	return record
}

func ToTask(task *TaskRecords, steps []*StepRecords) *types.Task {
	t := &types.Task{
		TaskID:              task.TaskID,
		TaskType:            task.TaskType,
		TaskName:            task.TaskName,
		CurrentStep:         task.CurrentStep,
		CallBackFuncName:    task.CallBackFuncName,
		CommonParams:        task.CommonParams,
		ExtraJson:           task.ExtraJson,
		Status:              task.Status,
		Message:             task.Message,
		ForceTerminate:      task.ForceTerminate,
		Start:               task.Start,
		End:                 task.End,
		ExecutionTime:       task.ExecutionTime,
		MaxExecutionSeconds: task.MaxExecutionSeconds,
		Creator:             task.Creator,
		LastUpdate:          task.LastUpdate,
		Updater:             task.Updater,
	}
	t.Steps = make([]*types.Step, 0, len(steps))
	for _, step := range steps {
		t.Steps = append(t.Steps, &types.Step{
			Name:         step.Name,
			Alias:        step.Alias,
			Extras:       step.Extras,
			Status:       step.Status,
			Message:      step.Message,
			SkipOnFailed: step.SkipOnFailed,
		})
	}
	return t
}
