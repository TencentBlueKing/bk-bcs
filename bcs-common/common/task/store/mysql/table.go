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
	TaskType            string            `json:"taskType" gorm:"taskType"`
	TaskName            string            `json:"taskName" gorm:"taskName"`
	CurrentStep         string            `json:"currentStep" gorm:"currentStep"`
	StepSequence        []string          `json:"stepSequence" gorm:"stepSequence"`
	StepIds             map[string]int64  `json:"stepIds" gorm:"stepIds"`
	CallBackFuncName    string            `json:"callBackFuncName" gorm:"callBackFuncName"`
	CommonParams        map[string]string `json:"commonParams" gorm:"commonParams"`
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
	Name                string            `json:"name" gorm:"name"`
	Alias               string            `json:"alias" gorm:"alias"`
	Input               map[string]string `json:"input" gorm:"input"`
	Output              map[string]string `json:"output" gorm:"output"`
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
