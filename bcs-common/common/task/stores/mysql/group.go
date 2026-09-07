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
	"fmt"
	"slices"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/Tencent/bk-bcs/bcs-common/common/task/stores/iface"
	"github.com/Tencent/bk-bcs/bcs-common/common/task/types"
)

// EnsureGroupTable implement istore EnsureGroupTable interface
func (s *mysqlStore) EnsureGroupTable(ctx context.Context, dst ...any) error {
	if len(dst) == 0 {
		dst = []any{&TaskGroupRecord{}, &StageRecord{}}
	}
	return s.db.WithContext(ctx).AutoMigrate(dst...)
}

// CreateGroup implement istore CreateGroup interface
func (s *mysqlStore) CreateGroup(ctx context.Context, group *types.TaskGroup, tasks []*types.Task) error {
	if group == nil {
		return fmt.Errorf("task group to be created cannot be empty")
	}

	groupRecord := getGroupRecord(group)
	stageRecords := getStageRecords(group)

	taskRecords := make([]*TaskRecord, 0, len(tasks))
	stepRecords := make([]*StepRecord, 0, len(tasks))
	for _, task := range tasks {
		taskRecords = append(taskRecords, getTaskRecord(task))
		stepRecords = append(stepRecords, getStepRecord(task)...)
	}

	// 任务组、阶段与组内全部任务同事务落库, 避免部分任务落库后编排计划缺失
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(groupRecord).Error; err != nil {
			return err
		}
		if err := tx.CreateInBatches(stageRecords, batchSize).Error; err != nil {
			return err
		}
		if len(taskRecords) == 0 {
			return nil
		}
		if err := tx.CreateInBatches(taskRecords, batchSize).Error; err != nil {
			return err
		}
		return tx.CreateInBatches(stepRecords, batchSize).Error
	})
}

// GetGroup implement istore GetGroup interface
func (s *mysqlStore) GetGroup(ctx context.Context, groupID string) (*types.TaskGroup, error) {
	return getGroupTx(s.db.WithContext(ctx), groupID, false)
}

// UpdateGroup implement istore UpdateGroup interface
func (s *mysqlStore) UpdateGroup(ctx context.Context, group *types.TaskGroup) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return saveGroupTx(tx, group)
	})
}

// ListGroup implement istore ListGroup interface
func (s *mysqlStore) ListGroup(ctx context.Context, opt *iface.ListGroupOption) (
	*iface.Pagination[types.TaskGroup], error) {
	tx := s.db.WithContext(ctx)

	if opt == nil {
		return nil, fmt.Errorf("list group option cannot be empty")
	}
	// 条件过滤 0值gorm自动忽略查询
	tx = tx.Where(&TaskGroupRecord{
		GroupID:        opt.GroupID,
		GroupType:      opt.GroupType,
		GroupName:      opt.GroupName,
		GroupIndex:     opt.GroupIndex,
		GroupIndexType: opt.GroupIndexType,
		Creator:        opt.Creator,
	})

	if len(opt.StatusList) > 0 {
		tx = tx.Where("status IN ?", opt.StatusList)
	} else if opt.Status != "" {
		tx = tx.Where("status = ?", opt.Status)
	}

	if opt.CreatedGte != nil {
		tx = tx.Where("created_at >= ?", opt.CreatedGte)
	}
	if opt.CreatedLte != nil {
		tx = tx.Where("created_at <= ?", opt.CreatedLte)
	}

	if len(opt.Sort) != 0 {
		for _, sort := range opt.Sort {
			if !slices.Contains(GroupSortableFields, sort.Field) {
				return nil, fmt.Errorf("invalid sort field: %s", sort.Field)
			}
			if sort.Desc {
				tx = tx.Order(sort.Field + " DESC")
			} else {
				tx = tx.Order(sort.Field + " ASC")
			}
		}
	} else {
		tx = tx.Order("id DESC")
	}

	records, count, err := FindByPage[TaskGroupRecord](tx, int(opt.Offset), int(opt.Limit))
	if err != nil {
		return nil, err
	}

	groupIDs := make([]string, 0, len(records))
	for _, record := range records {
		groupIDs = append(groupIDs, record.GroupID)
	}

	stageMap, err := s.listStageRecordByGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}

	items := make([]*types.TaskGroup, 0, len(records))
	for _, record := range records {
		items = append(items, toTaskGroup(record, stageMap[record.GroupID]))
	}

	return &iface.Pagination[types.TaskGroup]{Count: count, Items: items}, nil
}

func (s *mysqlStore) listStageRecordByGroupIDs(ctx context.Context, groupIDs []string) (
	map[string][]*StageRecord, error) {
	result := make(map[string][]*StageRecord, len(groupIDs))
	if len(groupIDs) == 0 {
		return result, nil
	}

	records := make([]*StageRecord, 0)
	if err := s.db.WithContext(ctx).Where("group_id IN ?", groupIDs).
		Order("seq ASC").Find(&records).Error; err != nil {
		return nil, err
	}
	for _, record := range records {
		result[record.GroupID] = append(result[record.GroupID], record)
	}
	return result, nil
}

// ListStageTaskIDs implement istore ListStageTaskIDs interface
func (s *mysqlStore) ListStageTaskIDs(ctx context.Context, groupID string, stageSeqs []int,
	statusList []string) ([]string, error) {
	return listStageTaskIDsTx(s.db.WithContext(ctx), groupID, stageSeqs, statusList)
}

// AdvanceGroup implement istore AdvanceGroup interface
func (s *mysqlStore) AdvanceGroup(ctx context.Context, groupID, taskID, taskStatus string) (
	*iface.GroupAdvanceResult, error) {
	result := &iface.GroupAdvanceResult{Advance: &types.AdvanceResult{}}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 幂等闸门: 抢占任务的计数资格, 抢不到说明该任务已被推进过
		claimed, stageSeq, err := claimTaskTx(tx, groupID, taskID)
		if err != nil {
			return err
		}

		group, err := getGroupTx(tx, groupID, true)
		if err != nil {
			return err
		}
		result.Group = group
		if !claimed {
			return nil
		}

		result.Advance = group.Advance(stageSeq, taskStatus)
		return settleAdvanceTx(tx, group, result)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ReconcileGroup implement istore ReconcileGroup interface
func (s *mysqlStore) ReconcileGroup(ctx context.Context, groupID string) (*iface.GroupAdvanceResult, error) {
	result := &iface.GroupAdvanceResult{Advance: &types.AdvanceResult{}}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		group, err := getGroupTx(tx, groupID, true)
		if err != nil {
			return err
		}
		result.Group = group

		counters, err := countStageTaskStatusTx(tx, groupID)
		if err != nil {
			return err
		}

		// 已达终态但未计数的任务在本次对账中被一并收敛, 之后的回调不会重复推进
		if err = tx.Model(&TaskRecord{}).
			Where("group_id = ? AND group_counted = ? AND status IN ?", groupID, false, types.TerminalTaskStatus).
			Update("group_counted", true).Error; err != nil {
			return err
		}

		result.Advance = group.Reconcile(counters)
		return settleAdvanceTx(tx, group, result)
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ResetGroupStages implement istore ResetGroupStages interface
func (s *mysqlStore) ResetGroupStages(ctx context.Context, groupID string, fromSeq int) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		group, err := getGroupTx(tx, groupID, true)
		if err != nil {
			return err
		}

		retryStatus := []string{types.TaskStatusFailure, types.TaskStatusTimeout, types.TaskStatusRevoked}
		taskIDs := make([]string, 0)
		if err = tx.Model(&TaskRecord{}).
			Where("group_id = ? AND stage_seq >= ? AND status IN ?", groupID, fromSeq, retryStatus).
			Pluck("task_id", &taskIDs).Error; err != nil {
			return err
		}

		for _, chunk := range chunkSlice(taskIDs, batchSize) {
			if len(chunk) == 0 {
				continue
			}
			if err = tx.Model(&TaskRecord{}).Where("task_id IN ?", chunk).Updates(map[string]any{
				"status":           types.TaskStatusInit,
				"message":          types.DefaultTaskMessage,
				"group_counted":    false,
				"current_step":     "",
				"callback_result":  "",
				"callback_message": "",
				"execution_time":   0,
				"start":            UnixZeroTime,
				"end":              UnixZeroTime,
			}).Error; err != nil {
				return err
			}
			if err = tx.Model(&StepRecord{}).Where("task_id IN ?", chunk).Updates(map[string]any{
				"status":         types.TaskStatusNotStarted,
				"message":        "",
				"retry_count":    0,
				"execution_time": 0,
				"start":          UnixZeroTime,
				"end":            UnixZeroTime,
			}).Error; err != nil {
				return err
			}
		}

		// 阶段恢复为未下发, 由后续对账按实际任务状态重建计数并下发。
		// 这里只改内存对象, 统一由 saveGroupTx 落库, 避免与其写回的阶段数据互相覆盖。
		group.FailureCount, group.SkippedCount = 0, 0
		for _, stage := range group.Stages {
			if stage.Seq < fromSeq {
				continue
			}
			stage.Status = types.StageStatusNotStarted
			stage.Message = ""
			stage.Dispatched = false
			stage.Failed, stage.Skipped = 0, 0
			stage.Completed = stage.Succeeded
		}

		group.SetStatus(types.TaskStatusRunning).SetMessage("task group retrying")
		group.End = UnixZeroTime
		return saveGroupTx(tx, group)
	})
}

// settleAdvanceTx 把一次推进结果落库: 终结被阻断任务、收集待下发任务、写回任务组与阶段
func settleAdvanceTx(tx *gorm.DB, group *types.TaskGroup, result *iface.GroupAdvanceResult) error {
	advance := result.Advance

	if len(advance.BlockedSeqs) > 0 {
		blockedIDs, err := listStageTaskIDsTx(tx, group.GroupID, advance.BlockedSeqs, types.PendingTaskStatus)
		if err != nil {
			return err
		}
		if err = blockTasksTx(tx, blockedIDs, advance.BlockedTaskStatus, advance.BlockedMessage); err != nil {
			return err
		}
		result.BlockedTaskIDs = blockedIDs
	}

	if len(advance.NextStages) > 0 {
		seqs := make([]int, 0, len(advance.NextStages))
		for _, stage := range advance.NextStages {
			seqs = append(seqs, stage.Seq)
		}
		// 只投递待执行的任务: 重试场景下阶段会被重新下发, 其中已成功的任务不应再次入队
		nextIDs, err := listStageTaskIDsTx(tx, group.GroupID, seqs, types.PendingTaskStatus)
		if err != nil {
			return err
		}
		result.NextTaskIDs = nextIDs
	}

	return saveGroupTx(tx, group)
}

// claimTaskTx 抢占任务在所属任务组内的计数资格, 返回是否抢占成功以及任务所属阶段
func claimTaskTx(tx *gorm.DB, groupID, taskID string) (bool, int, error) {
	record := &TaskRecord{}
	if err := tx.Select("stage_seq").Where("task_id = ? AND group_id = ?", taskID, groupID).
		First(record).Error; err != nil {
		return false, 0, err
	}

	claim := tx.Model(&TaskRecord{}).
		Where("task_id = ? AND group_id = ? AND group_counted = ?", taskID, groupID, false).
		Update("group_counted", true)
	if claim.Error != nil {
		return false, 0, claim.Error
	}
	return claim.RowsAffected > 0, record.StageSeq, nil
}

// blockTasksTx 批量终结被阻断的任务, 并同时标记已计数避免重复推进
func blockTasksTx(tx *gorm.DB, taskIDs []string, status, message string) error {
	if len(taskIDs) == 0 {
		return nil
	}
	now := time.Now()
	for _, chunk := range chunkSlice(taskIDs, batchSize) {
		if err := tx.Model(&TaskRecord{}).Where("task_id IN ?", chunk).Updates(map[string]any{
			"status":        status,
			"message":       message,
			"group_counted": true,
			"end":           now,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

func listStageTaskIDsTx(tx *gorm.DB, groupID string, stageSeqs []int, statusList []string) ([]string, error) {
	if len(stageSeqs) == 0 {
		return nil, nil
	}

	db := tx.Model(&TaskRecord{}).Where("group_id = ? AND stage_seq IN ?", groupID, stageSeqs)
	if len(statusList) > 0 {
		db = db.Where("status IN ?", statusList)
	}

	taskIDs := make([]string, 0)
	if err := db.Order("id ASC").Pluck("task_id", &taskIDs).Error; err != nil {
		return nil, err
	}
	return taskIDs, nil
}

// stageStatusCount 阶段内任务按状态的分布
type stageStatusCount struct {
	StageSeq int
	Status   string
	Total    int
}

// countStageTaskStatusTx 按阶段统计任务的实际状态分布
func countStageTaskStatusTx(tx *gorm.DB, groupID string) (map[int]*types.StageCounter, error) {
	rows := make([]*stageStatusCount, 0)
	if err := tx.Model(&TaskRecord{}).
		Select("stage_seq, status, count(*) as total").
		Where("group_id = ?", groupID).
		Group("stage_seq, status").
		Scan(&rows).Error; err != nil {
		return nil, err
	}

	counters := make(map[int]*types.StageCounter)
	for _, row := range rows {
		counter, ok := counters[row.StageSeq]
		if !ok {
			counter = &types.StageCounter{}
			counters[row.StageSeq] = counter
		}
		counter.Total += row.Total

		switch row.Status {
		case types.TaskStatusSuccess:
			counter.Succeeded += row.Total
		case types.TaskStatusFailure, types.TaskStatusTimeout:
			counter.Failed += row.Total
		case types.TaskStatusRevoked:
			counter.Skipped += row.Total
		default:
			counter.Pending += row.Total
		}
	}
	return counters, nil
}

// getGroupTx 读取任务组, forUpdate 为 true 时对任务组行加锁以串行化并发推进
func getGroupTx(tx *gorm.DB, groupID string, forUpdate bool) (*types.TaskGroup, error) {
	db := tx
	if forUpdate {
		db = db.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	record := &TaskGroupRecord{}
	if err := db.Where("group_id = ?", groupID).First(record).Error; err != nil {
		return nil, err
	}

	// 阶段行必须与任务组行用同样的锁读: REPEATABLE READ 下普通 SELECT 走事务的一致性视图,
	// 而该视图在等待任务组行锁之前就已建立, 读回的阶段计数会是前一个推进事务提交前的旧值,
	// 并发推进时后写回的绝对计数会覆盖掉前一次的增量, 阶段永远凑不满 Total。
	stageDB := tx
	if forUpdate {
		stageDB = tx.Clauses(clause.Locking{Strength: "UPDATE"})
	}

	stages := make([]*StageRecord, 0)
	if err := stageDB.Where("group_id = ?", groupID).Order("seq ASC").Find(&stages).Error; err != nil {
		return nil, err
	}
	return toTaskGroup(record, stages), nil
}

func saveGroupTx(tx *gorm.DB, group *types.TaskGroup) error {
	group.LastUpdate = time.Now()

	if err := tx.Model(&TaskGroupRecord{}).Where("group_id = ?", group.GroupID).
		Select(updateGroupField).Updates(getUpdateGroupRecord(group)).Error; err != nil {
		return err
	}

	for _, stage := range group.Stages {
		if err := tx.Model(&StageRecord{}).
			Where("group_id = ? AND seq = ?", group.GroupID, stage.Seq).
			Select(updateStageField).Updates(getUpdateStageRecord(stage)).Error; err != nil {
			return err
		}
	}
	return nil
}
