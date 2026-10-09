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

// GroupSortableFields 任务组可排序字段
var GroupSortableFields = []string{
	"id",
	"created_at",
	"updated_at",
	"group_id",
	"group_type",
	"group_name",
	"group_index",
	"group_index_type",
	"status",
	"start",
	"end",
	"creator",
	"updater",
}

// TaskGroupRecord 任务组记录
type TaskGroupRecord struct {
	BaseModel
	GroupID             string            `json:"groupID" gorm:"type:varchar(191);uniqueIndex:idx_group_id"`
	GroupType           string            `json:"groupType" gorm:"type:varchar(191);index:idx_group_type"`
	GroupName           string            `json:"groupName" gorm:"type:varchar(255)"`
	GroupIndex          string            `json:"groupIndex" gorm:"type:varchar(191);index:idx_group_index"`
	GroupIndexType      string            `json:"groupIndexType" gorm:"type:varchar(191);index:idx_group_index"`
	Status              string            `json:"status" gorm:"type:varchar(191);index:idx_group_status"`
	Message             string            `json:"message" gorm:"type:text"`
	CurrentBarrier      int               `json:"currentBarrier"`
	TotalCount          int               `json:"totalCount"`
	SuccessCount        int               `json:"successCount"`
	FailureCount        int               `json:"failureCount"`
	SkippedCount        int               `json:"skippedCount"`
	IgnoredCount        int               `json:"ignoredCount"`
	CallbackName        string            `json:"callbackName" gorm:"type:varchar(255)"`
	CallbackResult      string            `json:"callbackResult" gorm:"type:varchar(191)"`
	CallbackMessage     string            `json:"callbackMessage" gorm:"type:text"`
	CommonParams        map[string]string `json:"commonParams" gorm:"type:text;serializer:json"`
	CommonPayload       string            `json:"commonPayload" gorm:"type:text"`
	ExecutionTime       uint32            `json:"executionTime"`
	MaxExecutionSeconds uint32            `json:"maxExecutionSeconds"`
	Start               time.Time         `json:"start"`
	End                 time.Time         `json:"end"`
	Creator             string            `json:"creator" gorm:"type:varchar(255)"`
	Updater             string            `json:"updater" gorm:"type:varchar(255)"`
}

// TableName ..
func (t *TaskGroupRecord) TableName() string {
	return "task_group_records"
}

// BeforeCreate ..
func (t *TaskGroupRecord) BeforeCreate(tx *gorm.DB) error {
	t.normalizeTime()
	return nil
}

// BeforeUpdate ..
func (t *TaskGroupRecord) BeforeUpdate(tx *gorm.DB) error {
	t.normalizeTime()
	return nil
}

func (t *TaskGroupRecord) normalizeTime() {
	if t.Start.IsZero() {
		t.Start = UnixZeroTime
	}
	if t.End.IsZero() {
		t.End = UnixZeroTime
	}
}

// StageRecord 任务组阶段记录
type StageRecord struct {
	gorm.Model
	GroupID           string `json:"groupID" gorm:"type:varchar(191);uniqueIndex:idx_group_id_stage_seq"`
	Seq               int    `json:"seq" gorm:"uniqueIndex:idx_group_id_stage_seq"`
	Name              string `json:"name" gorm:"type:varchar(255)"`
	OnFailure         string `json:"onFailure" gorm:"type:varchar(64)"`
	BlockMessage      string `json:"blockMessage" gorm:"type:text"`
	StartWithPrevious bool   `json:"startWithPrevious"`
	Status            string `json:"status" gorm:"type:varchar(191)"`
	Message           string `json:"message" gorm:"type:text"`
	Dispatched        bool   `json:"dispatched"`
	Total             int    `json:"total"`
	Completed         int    `json:"completed"`
	Succeeded         int    `json:"succeeded"`
	Failed            int    `json:"failed"`
	Skipped           int    `json:"skipped"`
	Ignored           int    `json:"ignored"`
}

// TableName ..
func (s *StageRecord) TableName() string {
	return "task_stage_records"
}

// ToStage 类型转换
func (s *StageRecord) ToStage() *types.Stage {
	return &types.Stage{
		Seq:               s.Seq,
		Name:              s.Name,
		OnFailure:         types.StageFailurePolicy(s.OnFailure),
		BlockMessage:      s.BlockMessage,
		StartWithPrevious: s.StartWithPrevious,
		Status:            s.Status,
		Message:           s.Message,
		Dispatched:        s.Dispatched,
		Total:             s.Total,
		Completed:         s.Completed,
		Succeeded:         s.Succeeded,
		Failed:            s.Failed,
		Skipped:           s.Skipped,
		Ignored:           s.Ignored,
	}
}

func getGroupRecord(g *types.TaskGroup) *TaskGroupRecord {
	return &TaskGroupRecord{
		GroupID:             g.GroupID,
		GroupType:           g.GroupType,
		GroupName:           g.GroupName,
		GroupIndex:          g.GroupIndex,
		GroupIndexType:      g.GroupIndexType,
		Status:              g.Status,
		Message:             g.Message,
		CurrentBarrier:      g.CurrentBarrier,
		TotalCount:          g.TotalCount,
		SuccessCount:        g.SuccessCount,
		FailureCount:        g.FailureCount,
		SkippedCount:        g.SkippedCount,
		IgnoredCount:        g.IgnoredCount,
		CallbackName:        g.CallbackName,
		CallbackResult:      g.CallbackResult,
		CallbackMessage:     g.CallbackMessage,
		CommonParams:        g.CommonParams,
		CommonPayload:       g.CommonPayload,
		ExecutionTime:       g.ExecutionTime,
		MaxExecutionSeconds: g.MaxExecutionSeconds,
		Start:               g.Start,
		End:                 g.End,
		Creator:             g.Creator,
		Updater:             g.Updater,
	}
}

func getStageRecords(g *types.TaskGroup) []*StageRecord {
	records := make([]*StageRecord, 0, len(g.Stages))
	for _, stage := range g.Stages {
		records = append(records, &StageRecord{
			GroupID:           g.GroupID,
			Seq:               stage.Seq,
			Name:              stage.Name,
			OnFailure:         stage.OnFailure.String(),
			BlockMessage:      stage.BlockMessage,
			StartWithPrevious: stage.StartWithPrevious,
			Status:            stage.Status,
			Message:           stage.Message,
			Dispatched:        stage.Dispatched,
			Total:             stage.Total,
			Completed:         stage.Completed,
			Succeeded:         stage.Succeeded,
			Failed:            stage.Failed,
			Skipped:           stage.Skipped,
			Ignored:           stage.Ignored,
		})
	}
	return records
}

func toTaskGroup(record *TaskGroupRecord, stages []*StageRecord) *types.TaskGroup {
	if record == nil {
		return nil
	}
	g := &types.TaskGroup{
		GroupID:             record.GroupID,
		GroupType:           record.GroupType,
		GroupName:           record.GroupName,
		GroupIndex:          record.GroupIndex,
		GroupIndexType:      record.GroupIndexType,
		Status:              record.Status,
		Message:             record.Message,
		CurrentBarrier:      record.CurrentBarrier,
		TotalCount:          record.TotalCount,
		SuccessCount:        record.SuccessCount,
		FailureCount:        record.FailureCount,
		SkippedCount:        record.SkippedCount,
		IgnoredCount:        record.IgnoredCount,
		CallbackName:        record.CallbackName,
		CallbackResult:      record.CallbackResult,
		CallbackMessage:     record.CallbackMessage,
		CommonParams:        record.CommonParams,
		CommonPayload:       record.CommonPayload,
		ExecutionTime:       record.ExecutionTime,
		MaxExecutionSeconds: record.MaxExecutionSeconds,
		Start:               record.Start,
		End:                 record.End,
		CreatedAt:           record.CreatedAt,
		LastUpdate:          record.UpdatedAt,
		Creator:             record.Creator,
		Updater:             record.Updater,
	}

	g.Stages = make([]*types.Stage, 0, len(stages))
	for _, stage := range stages {
		if stage == nil {
			continue
		}
		g.Stages = append(g.Stages, stage.ToStage())
	}
	return g
}

var (
	// updateGroupField 任务组支持更新的字段
	updateGroupField = []string{
		"Status",
		"Message",
		"CurrentBarrier",
		"TotalCount",
		"SuccessCount",
		"FailureCount",
		"SkippedCount",
		"IgnoredCount",
		"CallbackResult",
		"CallbackMessage",
		"CommonParams",
		"CommonPayload",
		"ExecutionTime",
		"Start",
		"End",
		"Updater",
	}

	// updateStageField 阶段支持更新的字段
	updateStageField = []string{
		"Status",
		"Message",
		"Dispatched",
		"Total",
		"Completed",
		"Succeeded",
		"Failed",
		"Skipped",
		"Ignored",
	}
)

func getUpdateGroupRecord(g *types.TaskGroup) *TaskGroupRecord {
	return &TaskGroupRecord{
		Status:          g.Status,
		Message:         g.Message,
		CurrentBarrier:  g.CurrentBarrier,
		TotalCount:      g.TotalCount,
		SuccessCount:    g.SuccessCount,
		FailureCount:    g.FailureCount,
		SkippedCount:    g.SkippedCount,
		IgnoredCount:    g.IgnoredCount,
		CallbackResult:  g.CallbackResult,
		CallbackMessage: g.CallbackMessage,
		CommonParams:    g.CommonParams,
		CommonPayload:   g.CommonPayload,
		ExecutionTime:   g.ExecutionTime,
		Start:           normalizeTime(g.Start),
		End:             normalizeTime(g.End),
		Updater:         g.Updater,
	}
}

func getUpdateStageRecord(s *types.Stage) *StageRecord {
	return &StageRecord{
		Status:     s.Status,
		Message:    s.Message,
		Dispatched: s.Dispatched,
		Total:      s.Total,
		Completed:  s.Completed,
		Succeeded:  s.Succeeded,
		Failed:     s.Failed,
		Skipped:    s.Skipped,
		Ignored:    s.Ignored,
	}
}
