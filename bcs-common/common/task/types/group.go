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

// Package types for task
package types

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

const (
	// StageStatusNotStarted 阶段尚未下发
	StageStatusNotStarted = "NOTSTARTED"
	// StageStatusRunning 阶段已下发, 组内任务执行中
	StageStatusRunning = "RUNNING"
	// StageStatusSuccess 阶段内任务全部成功
	StageStatusSuccess = "SUCCESS"
	// StageStatusFailure 阶段内存在失败任务
	StageStatusFailure = "FAILURE"
	// StageStatusSkipped 因前序阶段阻断而未执行
	StageStatusSkipped = "SKIPPED"
)

// DefaultGroupMessage 任务组初始描述
const DefaultGroupMessage = "task group initializing"

// StageFailurePolicy 阶段内出现失败任务后, 对后续阶段的处理策略
type StageFailurePolicy string

const (
	// StageFailureBlock 阻断后续阶段, 尚未下发的任务置为失败
	StageFailureBlock StageFailurePolicy = "block"
	// StageFailureContinue 忽略本阶段失败, 继续下发后续阶段
	StageFailureContinue StageFailurePolicy = "continue"
	// StageFailureSkip 阻断后续阶段, 尚未下发的任务置为撤销而非失败
	StageFailureSkip StageFailurePolicy = "skip"
)

// Valid 是否为合法策略
func (p StageFailurePolicy) Valid() bool {
	switch p {
	case StageFailureBlock, StageFailureContinue, StageFailureSkip:
		return true
	default:
		return false
	}
}

// String toString
func (p StageFailurePolicy) String() string {
	return string(p)
}

// Stage 任务组内的一个执行阶段。阶段内任务并行, 阶段之间默认串行。
//
// 框架不理解阶段的业务含义, 只按 Seq 顺序推进。
// 调用方负责把业务语义映射成阶段顺序与阻断文案。
type Stage struct {
	// Seq 阶段序号, 组内唯一且严格递增
	Seq int `json:"seq"`
	// Name 阶段的业务语义名, 仅用于展示, 例如 "priority-100"
	Name string `json:"name"`

	// OnFailure 本阶段出现失败任务时对后续阶段的处理策略
	OnFailure StageFailurePolicy `json:"onFailure"`

	// BlockMessage 阻断后续任务时写入这些任务的 message。
	// 由调用方提供, 框架不构造任何业务文案。
	BlockMessage string `json:"blockMessage"`

	// StartWithPrevious 为 true 时与前一阶段同时下发, 不等待前一阶段完成。
	// 用于表达「一组不参与串行编排的任务与首批任务并发起跑」这类场景。
	// 首个阶段该字段无意义。
	StartWithPrevious bool `json:"startWithPrevious"`

	Status     string `json:"status"`
	Message    string `json:"message"`
	Dispatched bool   `json:"dispatched"`

	Total     int `json:"total"`
	Completed int `json:"completed"`
	Succeeded int `json:"succeeded"`
	Failed    int `json:"failed"`
	Skipped   int `json:"skipped"`
	// Ignored 阶段内幂等忽略的任务数, 在阻断判定上等同成功, 仅计数上区分
	Ignored int `json:"ignored"`
}

// IsTerminal 阶段是否已达终态
func (s *Stage) IsTerminal() bool {
	switch s.Status {
	case StageStatusSuccess, StageStatusFailure, StageStatusSkipped:
		return true
	default:
		return false
	}
}

// GroupInfo 任务组基本信息
type GroupInfo struct {
	GroupType      string
	GroupName      string
	GroupIndex     string // GroupIndex 业务资源索引, 便于按资源反查任务组
	GroupIndexType string
	Creator        string
}

// GroupOptions 任务组可选项
type GroupOptions struct {
	CallbackName        string
	MaxExecutionSeconds uint32
}

// GroupOption 任务组可选项设置函数
type GroupOption func(opt *GroupOptions)

// WithGroupCallback 设置任务组级回调名称
func WithGroupCallback(callbackName string) GroupOption {
	return func(opt *GroupOptions) {
		opt.CallbackName = callbackName
	}
}

// WithGroupMaxExecutionSeconds 设置任务组最大执行时长
func WithGroupMaxExecutionSeconds(timeout uint32) GroupOption {
	return func(opt *GroupOptions) {
		opt.MaxExecutionSeconds = timeout
	}
}

// TaskGroup 一组需要协同编排的任务。
//
// 组内任务按 Stage 分阶段推进: 同一阶段的任务并行执行, 后一阶段等待前一阶段
// 全部到达终态后才下发。任务与阶段的归属关系记录在 Task.GroupID/Task.StageSeq 上。
type TaskGroup struct {
	GroupID        string `json:"groupID"`
	GroupType      string `json:"groupType"`
	GroupName      string `json:"groupName"`
	GroupIndex     string `json:"groupIndex"`
	GroupIndexType string `json:"groupIndexType"`

	Status  string `json:"status"`
	Message string `json:"message"`

	// CurrentBarrier 当前已下发的栅栏序号。一个栅栏由一个阶段, 以及紧随其后
	// 所有 StartWithPrevious 为 true 的阶段组成, 它们同时下发。
	CurrentBarrier int `json:"currentBarrier"`

	TotalCount   int `json:"totalCount"`
	SuccessCount int `json:"successCount"`
	FailureCount int `json:"failureCount"`
	SkippedCount int `json:"skippedCount"`
	// IgnoredCount 组内幂等忽略的任务数, 不计入 SuccessCount, 也不会让任务组收敛为失败
	IgnoredCount int `json:"ignoredCount"`

	CallbackName    string            `json:"callbackName"`
	CallbackResult  string            `json:"callbackResult"`
	CallbackMessage string            `json:"callbackMessage"`
	CommonParams    map[string]string `json:"commonParams"`
	CommonPayload   string            `json:"commonPayload"`

	MaxExecutionSeconds uint32 `json:"maxExecutionSeconds"`
	ExecutionTime       uint32 `json:"executionTime"`

	Creator    string    `json:"creator"`
	Updater    string    `json:"updater"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
	CreatedAt  time.Time `json:"createdAt"`
	LastUpdate time.Time `json:"lastUpdate"`

	Stages []*Stage `json:"stages"`
}

// NewTaskGroup 创建任务组
func NewTaskGroup(info GroupInfo, opts ...GroupOption) *TaskGroup {
	defaultOptions := &GroupOptions{}
	for _, opt := range opts {
		opt(defaultOptions)
	}

	return &TaskGroup{
		GroupID:             uuid.NewString(),
		GroupType:           info.GroupType,
		GroupName:           info.GroupName,
		GroupIndex:          info.GroupIndex,
		GroupIndexType:      info.GroupIndexType,
		Status:              TaskStatusInit,
		Message:             DefaultGroupMessage,
		Creator:             info.Creator,
		Updater:             info.Creator,
		CommonParams:        make(map[string]string, 0),
		CommonPayload:       DefaultPayloadContent,
		CallbackName:        defaultOptions.CallbackName,
		MaxExecutionSeconds: defaultOptions.MaxExecutionSeconds,
		LastUpdate:          time.Now(),
		Stages:              make([]*Stage, 0),
	}
}

// AddStage 追加一个阶段
func (g *TaskGroup) AddStage(stage *Stage) *TaskGroup {
	if stage == nil {
		return g
	}
	if stage.Status == "" {
		stage.Status = StageStatusNotStarted
	}
	if stage.OnFailure == "" {
		stage.OnFailure = StageFailureBlock
	}
	g.Stages = append(g.Stages, stage)
	return g
}

// GetGroupID get group id
func (g *TaskGroup) GetGroupID() string {
	return g.GroupID
}

// GetStatus get group status
func (g *TaskGroup) GetStatus() string {
	return g.Status
}

// SetStatus set group status
func (g *TaskGroup) SetStatus(status string) *TaskGroup {
	g.Status = status
	return g
}

// GetMessage get group message
func (g *TaskGroup) GetMessage() string {
	return g.Message
}

// SetMessage set group message
func (g *TaskGroup) SetMessage(msg string) *TaskGroup {
	g.Message = msg
	return g
}

// GetCallback get group callback name
func (g *TaskGroup) GetCallback() string {
	return g.CallbackName
}

// SetCallback set group callback name
func (g *TaskGroup) SetCallback(name string) *TaskGroup {
	g.CallbackName = name
	return g
}

// GetCommonPayload unmarshal group common payload to struct obj
func (g *TaskGroup) GetCommonPayload(obj any) error {
	if len(g.CommonPayload) == 0 {
		g.CommonPayload = DefaultPayloadContent
	}
	return json.Unmarshal([]byte(g.CommonPayload), obj)
}

// SetCommonPayload marshal struct obj to group common payload
func (g *TaskGroup) SetCommonPayload(obj any) error {
	result, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	g.CommonPayload = string(result)
	return nil
}

// GetStage 按序号获取阶段
func (g *TaskGroup) GetStage(seq int) (*Stage, bool) {
	for _, stage := range g.Stages {
		if stage.Seq == seq {
			return stage, true
		}
	}
	return nil, false
}

// IsTerminal 任务组是否已达终态
func (g *TaskGroup) IsTerminal() bool {
	return IsTaskTerminal(g.Status)
}

// Validate 校验任务组
func (g *TaskGroup) Validate() error {
	if g.GroupName == "" {
		return fmt.Errorf("group name is required")
	}
	if len(g.Stages) == 0 {
		return fmt.Errorf("group stages empty")
	}

	prevSeq := 0
	for i, stage := range g.Stages {
		// 阶段序号必须严格递增且唯一
		if i > 0 && stage.Seq <= prevSeq {
			return fmt.Errorf("stage seq must be ascending and unique, got %d after %d", stage.Seq, prevSeq)
		}
		prevSeq = stage.Seq

		// 阶段失败策略必须合法
		if !stage.OnFailure.Valid() {
			return fmt.Errorf("stage %d has invalid failure policy %q", stage.Seq, stage.OnFailure)
		}
	}
	return nil
}

// barriers 计算每个阶段所属的栅栏序号。
// StartWithPrevious 为 true 的阶段与前一阶段同属一个栅栏, 一并下发；
// 栅栏之间串行, 后一栅栏等待前一栅栏内所有阶段到达终态。
func (g *TaskGroup) barriers() []int {
	result := make([]int, len(g.Stages))
	current := 0
	for i, stage := range g.Stages {
		if i > 0 && !stage.StartWithPrevious {
			current++
		}
		result[i] = current
	}
	return result
}

// InitialStages 返回首个栅栏内的阶段, 即创建任务组后应立即下发的阶段
func (g *TaskGroup) InitialStages() []*Stage {
	return g.stagesInBarrier(0)
}

func (g *TaskGroup) stagesInBarrier(barrier int) []*Stage {
	bars := g.barriers()
	stages := make([]*Stage, 0, 1)
	for i, stage := range g.Stages {
		if bars[i] == barrier {
			stages = append(stages, stage)
		}
	}
	return stages
}

// MarkStageDispatched 标记阶段已下发
func (g *TaskGroup) MarkStageDispatched(stages []*Stage) {
	for _, stage := range stages {
		stage.Dispatched = true
		if stage.Status == StageStatusNotStarted {
			stage.Status = StageStatusRunning
		}
	}
}

// AdvanceResult 一次推进的结果, 由调用方负责落库与下发
type AdvanceResult struct {
	// StageCompleted 本次推进使该阶段到达终态, 未完成则为 nil
	StageCompleted *Stage
	// NextStages 需要下发的下一批阶段
	NextStages []*Stage
	// BlockedStages 因阻断被直接终结的阶段
	BlockedStages []*Stage
	// BlockedSeqs 被阻断阶段的序号, 供存储层按 stage_seq 批量终结任务
	BlockedSeqs []int
	// BlockedMessage 阻断原因, 写入被终结任务的 message
	BlockedMessage string
	// BlockedTaskStatus 被阻断任务应写入的状态
	BlockedTaskStatus string
	// GroupCompleted 任务组是否在本次推进后达到终态
	GroupCompleted bool
}

// Advance 依据组内某个任务的终态推进任务组状态。
//
// 这是一个纯函数式的状态机: 只修改内存中的 TaskGroup 与 Stage, 不做任何 IO。
// 调用方 (存储层) 负责在同一个事务内保证「任务终态写入」与「本次推进结果落库」的原子性,
// 并保证同一个任务只会被推进一次。
func (g *TaskGroup) Advance(stageSeq int, taskStatus string) *AdvanceResult {
	result := &AdvanceResult{}

	stage, ok := g.GetStage(stageSeq)
	if !ok || stage.IsTerminal() {
		return result
	}

	stage.Completed++
	switch taskStatus {
	case TaskStatusIgnored:
		stage.Ignored++
		g.IgnoredCount++
	case TaskStatusSuccess:
		stage.Succeeded++
		g.SuccessCount++
	default:
		stage.Failed++
		g.FailureCount++
	}

	// 阶段内仍有任务在执行
	if stage.Completed < stage.Total {
		return result
	}

	if stage.Failed > 0 {
		stage.Status = StageStatusFailure
	} else {
		stage.Status = StageStatusSuccess
	}
	result.StageCompleted = stage

	g.advanceBarrier(stage, result)
	g.settleIfCompleted(result)
	return result
}

// advanceBarrier 在某个阶段终结后, 判断其所属栅栏是否整体完成, 并决定阻断还是下发下一栅栏
func (g *TaskGroup) advanceBarrier(completed *Stage, result *AdvanceResult) {
	bars := g.barriers()
	current := 0
	for i, stage := range g.Stages {
		if stage.Seq == completed.Seq {
			current = bars[i]
			break
		}
	}

	// 同栅栏内还有阶段未终结, 等待
	for i, stage := range g.Stages {
		if bars[i] == current && !stage.IsTerminal() {
			return
		}
	}

	// 栅栏内任一阶段失败且策略要求阻断, 则终结所有尚未下发的后续阶段
	for i, stage := range g.Stages {
		if bars[i] != current || stage.Failed == 0 || stage.OnFailure == StageFailureContinue {
			continue
		}
		g.blockRemaining(bars, current, stage, result)
		return
	}

	for i, stage := range g.Stages {
		if bars[i] != current+1 {
			continue
		}
		stage.Dispatched = true
		stage.Status = StageStatusRunning
		result.NextStages = append(result.NextStages, stage)
	}
	if len(result.NextStages) > 0 {
		g.CurrentBarrier = current + 1
	}
}

// blockRemaining 终结 current 栅栏之后所有尚未下发的阶段
func (g *TaskGroup) blockRemaining(bars []int, current int, failed *Stage, result *AdvanceResult) {
	blockedStatus := StageStatusFailure
	taskStatus := TaskStatusFailure
	if failed.OnFailure == StageFailureSkip {
		blockedStatus = StageStatusSkipped
		taskStatus = TaskStatusRevoked
	}

	for i, stage := range g.Stages {
		if bars[i] <= current || stage.Dispatched || stage.IsTerminal() {
			continue
		}
		remaining := stage.Total - stage.Completed
		stage.Status = blockedStatus
		stage.Message = failed.BlockMessage
		stage.Dispatched = true
		stage.Completed = stage.Total
		if blockedStatus == StageStatusSkipped {
			stage.Skipped += remaining
			g.SkippedCount += remaining
		} else {
			stage.Failed += remaining
			g.FailureCount += remaining
		}
		result.BlockedStages = append(result.BlockedStages, stage)
		result.BlockedSeqs = append(result.BlockedSeqs, stage.Seq)
	}

	if len(result.BlockedStages) > 0 {
		result.BlockedMessage = failed.BlockMessage
		result.BlockedTaskStatus = taskStatus
	}
}

// settleIfCompleted 所有阶段终结后收敛任务组终态
func (g *TaskGroup) settleIfCompleted(result *AdvanceResult) {
	for _, stage := range g.Stages {
		if !stage.IsTerminal() {
			return
		}
	}

	now := time.Now()
	g.End = now
	g.SetExecutionTime(g.Start, now)
	// 任务组终态只有 SUCCESS/FAILURE: 幂等忽略视同成功, 数量由 IgnoredCount 单独暴露给调用方
	if g.FailureCount > 0 {
		g.SetStatus(TaskStatusFailure).SetMessage("task group finished with failures")
	} else {
		g.SetStatus(TaskStatusSuccess).SetMessage("task group finished successfully")
	}
	result.GroupCompleted = true
}

// SetExecutionTime set execution time
func (g *TaskGroup) SetExecutionTime(start time.Time, end time.Time) *TaskGroup {
	g.ExecutionTime = uint32(end.Sub(start).Milliseconds())
	return g
}

// StageCounter 某个阶段内按任务实际状态统计的结果, 用于崩溃后对账
type StageCounter struct {
	Total     int
	Succeeded int
	Failed    int
	Skipped   int
	Ignored   int
	Pending   int
}

// Reconcile 以任务存储中的实际终态重建阶段与任务组计数。
//
// 服务在推进过程中崩溃时, 任务终态可能已落库但组计数尚未更新。恢复流程读取
// 每个阶段的真实任务状态分布后调用本方法, 重新收敛内存状态, 并返回需要补下发
// 或补阻断的结果。
func (g *TaskGroup) Reconcile(counters map[int]*StageCounter) *AdvanceResult {
	result := &AdvanceResult{}

	g.SuccessCount, g.FailureCount, g.SkippedCount, g.IgnoredCount = 0, 0, 0, 0
	for _, stage := range g.Stages {
		counter, ok := counters[stage.Seq]
		if !ok {
			counter = &StageCounter{Total: stage.Total, Pending: stage.Total}
		}
		if counter.Total > 0 {
			stage.Total = counter.Total
		}
		stage.Succeeded, stage.Failed = counter.Succeeded, counter.Failed
		stage.Skipped, stage.Ignored = counter.Skipped, counter.Ignored
		stage.Completed = counter.Succeeded + counter.Failed + counter.Skipped + counter.Ignored

		g.SuccessCount += stage.Succeeded
		g.FailureCount += stage.Failed
		g.SkippedCount += stage.Skipped
		g.IgnoredCount += stage.Ignored

		switch {
		case !stage.Dispatched:
			stage.Status = StageStatusNotStarted
		case stage.Completed < stage.Total:
			stage.Status = StageStatusRunning
		case stage.Failed > 0:
			stage.Status = StageStatusFailure
		case stage.Skipped > 0 && stage.Succeeded == 0 && stage.Ignored == 0:
			stage.Status = StageStatusSkipped
		default:
			stage.Status = StageStatusSuccess
		}
	}

	g.resumeFromReconciled(result)
	g.settleIfCompleted(result)
	return result
}

// resumeFromReconciled 对账后从最靠前的未完成栅栏继续推进
func (g *TaskGroup) resumeFromReconciled(result *AdvanceResult) {
	bars := g.barriers()
	maxBarrier := 0
	if len(bars) > 0 {
		maxBarrier = bars[len(bars)-1]
	}

	for barrier := 0; barrier <= maxBarrier; barrier++ {
		stages := g.stagesInBarrier(barrier)

		// 本栅栏尚未下发, 补发
		if undispatched := filterUndispatched(stages); len(undispatched) > 0 {
			for _, stage := range undispatched {
				stage.Dispatched = true
				stage.Status = StageStatusRunning
			}
			result.NextStages = append(result.NextStages, undispatched...)
			g.CurrentBarrier = barrier
			return
		}

		// 本栅栏仍在执行, 等待其自然推进
		if !allTerminal(stages) {
			g.CurrentBarrier = barrier
			return
		}

		for _, stage := range stages {
			if stage.Failed == 0 || stage.OnFailure == StageFailureContinue {
				continue
			}
			g.blockRemaining(bars, barrier, stage, result)
			return
		}
	}
}

func filterUndispatched(stages []*Stage) []*Stage {
	result := make([]*Stage, 0, len(stages))
	for _, stage := range stages {
		if !stage.Dispatched {
			result = append(result, stage)
		}
	}
	return result
}

func allTerminal(stages []*Stage) bool {
	for _, stage := range stages {
		if !stage.IsTerminal() {
			return false
		}
	}
	return true
}
