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

// BaseModel 基础模型
type BaseModel struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"createdAt" gorm:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt" gorm:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// TaskRecords 任务记录
type TaskRecords struct {
	BaseModel
	TaskID              string            `json:"taskID" gorm:"index:idx_task_id,unique"` // 唯一索引
	TaskType            string            `json:"taskType" gorm:"taskType"`
	TaskName            string            `json:"taskName" gorm:"taskName"`
	CurrentStep         string            `json:"currentStep" gorm:"currentStep"`
	StepSequence        []string          `json:"stepSequence" gorm:"stepSequence;serializer:json"`
	CallBackFuncName    string            `json:"callBackFuncName" gorm:"callBackFuncName"`
	CommonParams        map[string]string `json:"commonParams" gorm:"commonParams;serializer:json"`
	ExtraJson           string            `json:"extraJson" gorm:"extraJson"`
	Status              string            `json:"status" gorm:"status"`
	Message             string            `json:"message" gorm:"message"`
	ForceTerminate      bool              `json:"forceTerminate" gorm:"forceTerminate"`
	Start               string            `json:"start" gorm:"start"`
	End                 string            `json:"end" gorm:"end"`
	ExecutionTime       uint32            `json:"executionTime" gorm:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds" gorm:"maxExecutionSeconds"`
	Creator             string            `json:"creator" gorm:"creator"`
	LastUpdate          string            `json:"lastUpdate" gorm:"lastUpdate"`
	Updater             string            `json:"updater" gorm:"updater"`
}

// TableName ..
func (t *TaskRecords) TableName() string {
	return "task_records"
}

// StepRecords 步骤记录
type StepRecords struct {
	BaseModel
	TaskID              string            `json:"taskID" gorm:"index:idx_task_id"` // 索引
	Name                string            `json:"name"`
	Alias               string            `json:"alias" gorm:"alias"`
	Input               map[string]string `json:"input" gorm:"input;serializer:json"`
	Output              map[string]string `json:"output" gorm:"output;serializer:json"`
	Extras              string            `json:"extras" gorm:"extras"`
	Status              string            `json:"status" gorm:"status"`
	Message             string            `json:"message" gorm:"message"`
	SkipOnFailed        bool              `json:"skipOnFailed" gorm:"skipOnFailed"`
	RetryCount          uint32            `json:"retryCount" gorm:"retryCount"`
	Start               string            `json:"start" gorm:"start"`
	End                 string            `json:"end" gorm:"end"`
	ExecutionTime       uint32            `json:"executionTime" gorm:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds" gorm:"maxExecutionSeconds"`
	LastUpdate          string            `json:"lastUpdate" gorm:"lastUpdate"`
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
