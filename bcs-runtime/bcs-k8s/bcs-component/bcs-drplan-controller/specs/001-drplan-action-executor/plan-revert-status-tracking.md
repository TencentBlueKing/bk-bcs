# Implementation Plan: Revert 操作状态记录

**Feature**: 为 Revert 操作添加与 Execute 一致的详细状态记录
**Created**: 2026-02-03  
**Status**: Planning  
**Priority**: P2 (Enhancement)

## 问题陈述

### 当前状态

**Execute 操作** 有完整的状态记录：
```yaml
status:
  phase: Succeeded
  startTime: "2026-02-03T10:00:00Z"
  completionTime: "2026-02-03T10:05:00Z"
  stageStatuses:
    - name: stage-1
      phase: Succeeded
      startTime: "2026-02-03T10:00:00Z"
      completionTime: "2026-02-03T10:02:00Z"
      workflowExecutions:
        - workflowRef: {name: workflow-a}
          phase: Succeeded
          progress: "3/3 actions completed"
          actionStatuses:
            - name: action-1
              phase: Succeeded
              startTime: ...
    - name: stage-2
      phase: Succeeded
      ...
  summary:
    totalStages: 2
    completedStages: 2
    succeededStages: 2
```

**Revert 操作** 当前只有简单的状态：
```yaml
status:
  phase: Succeeded
  message: "Plan reverted successfully"
  # ❌ 缺少 stageStatuses - 不知道回滚了哪些 stage
  # ❌ 缺少 workflowExecutions - 不知道每个 workflow 的回滚进度
  # ❌ 缺少 actionStatuses - 不知道具体哪些 action 被回滚
  # ❌ 缺少 summary - 没有统计信息
```

### 问题影响

1. **可观测性差**: 无法通过 `kubectl get drplanexecution` 查看 Revert 的详细进度
2. **故障排查难**: Revert 失败时，不知道在哪个 stage/workflow/action 失败
3. **审计不完整**: 无法追溯具体回滚了哪些资源
4. **用户体验不一致**: Execute 和 Revert 的状态格式不对称

---

## Technical Context

### 相关代码

| 文件 | 当前行为 | 需要修改 |
|------|---------|---------|
| `internal/executor/native_executor.go:300-386` | `RevertPlan()` 不记录 stage 状态 | ✅ 添加状态记录 |
| `internal/executor/stage_executor.go` | `RevertStage()` 返回 error | ✅ 返回 `*StageStatus` |
| `internal/executor/workflow_executor.go` | `RevertWorkflow()` 返回 error | ✅ 返回 `*WorkflowExecutionStatus` |
| `internal/executor/action_executor.go` | `Rollback()` 返回 error | ✅ 返回 `*ActionStatus` |

### 数据结构（已存在，无需修改）

`DRPlanExecutionStatus` 已经包含所有必要字段：
```go
type DRPlanExecutionStatus struct {
    Phase            string                     // ✅ 已有
    StartTime        *metav1.Time              // ✅ 已有
    CompletionTime   *metav1.Time              // ✅ 已有
    StageStatuses    []StageStatus             // ✅ 已有，Revert 应该填充
    Summary          *ExecutionSummary         // ✅ 已有，Revert 应该更新
    Message          string                     // ✅ 已有
}
```

---

## Constitution Check

### Alignment

| 原则 | 是否违反 | 说明 |
|------|---------|------|
| **I. Operator Pattern** | ✅ 符合 | 增强状态观测符合声明式原则 |
| **III. Observability** | ✅ **强化** | 提供完整的 Revert 可观测性 |
| **V. Progressive Execution** | ✅ 符合 | Revert 也应支持渐进式执行和监控 |

### Backward Compatibility

- ✅ **数据结构无变化**: 复用现有 `StageStatus`/`WorkflowExecutionStatus` 结构
- ✅ **API 兼容**: 不改变 CRD schema
- ✅ **升级无影响**: 旧版本创建的 Revert execution 仍可正常查看（只是 status 为空）

---

## Design

### Phase 0: Research

#### R1. 分析 Execute 的状态记录机制

**已调研**（从代码分析得出）：

```go
// ExecutePlan 的状态记录流程
func (e *NativePlanExecutor) ExecutePlan(...) error {
    // 1. 初始化 StageStatuses 和 Summary
    execution.Status.StageStatuses = make([]StageStatus, 0, len(stages))
    execution.Status.Summary = &ExecutionSummary{TotalStages: len(stages)}
    
    // 2. 每个 stage 执行后记录状态
    for _, stage := range readyStages {
        stageStatus, err := e.stageExecutor.ExecuteStage(...)
        e.updateStageStatusInExecution(execution, stageStatus)  // 关键方法
        e.updateExecutionSummary(execution)                      // 更新统计
    }
}
```

**结论**: 
- `ExecuteStage()` 返回 `*StageStatus`（包含 workflow 和 action 状态）
- `updateStageStatusInExecution()` 负责更新/追加 stage 状态
- `updateExecutionSummary()` 计算统计信息

#### R2. 分析 Revert 的当前实现

**当前流程**（native_executor.go L300-386）：

```go
func (e *NativePlanExecutor) RevertPlan(...) error {
    // 1. 获取目标 execution 的 StageStatuses（用于知道要回滚什么）
    targetExecution := &DRPlanExecution{}
    e.client.Get(ctx, targetExecKey, targetExecution)
    
    // 2. 逆序遍历 stage，调用 RevertStage
    for i := len(targetExecution.Status.StageStatuses) - 1; i >= 0; i-- {
        stageStatus := targetExecution.Status.StageStatuses[i]
        if stageStatus.Phase != "Succeeded" {
            continue  // 跳过失败的 stage
        }
        
        // ❌ 问题：RevertStage 返回 error，不返回状态
        err := e.stageExecutor.RevertStage(ctx, plan, stage, &stageStatus)
    }
    
    // 3. 只设置简单的成功状态
    execution.Status.Phase = "Succeeded"
    execution.Status.Message = "Plan reverted successfully"
    // ❌ 没有 StageStatuses、Summary
}
```

**结论**: 
- `RevertStage()` 需要改为返回 `(*StageStatus, error)`
- 需要在 `RevertPlan()` 中初始化和更新 `execution.Status.StageStatuses`

#### R3. Revert 状态语义

**决策**: Revert 的 StageStatus 应该反映**回滚操作本身**的执行情况，而非重复 Execute 的状态

| 字段 | Execute 语义 | Revert 语义 |
|------|-------------|------------|
| `name` | Stage 名称 | Stage 名称（与被回滚的 execution 对应） |
| `phase` | Execute 的结果 | **Rollback 的结果**（Succeeded/Failed/Skipped） |
| `workflowExecutions.actionStatuses` | Action 执行状态 | **Rollback action 执行状态** |
| `message` | Execute 失败原因 | **Rollback 失败原因**（如果失败） |

**示例**：

```yaml
# Execute execution (nginx-exec-001)
stageStatuses:
  - name: deploy-stage
    phase: Succeeded  # 原执行成功
    workflowExecutions:
      - actionStatuses:
          - name: create-localization
            type: Localization
            phase: Succeeded  # 原 action 成功

# Revert execution (nginx-revert-001) - 回滚上面的 execution
stageStatuses:
  - name: deploy-stage
    phase: Succeeded  # 回滚操作成功
    workflowExecutions:
      - actionStatuses:
          - name: create-localization
            type: Localization
            phase: Succeeded  # 回滚 action 成功（删除了 Localization）
            message: "Rolled back: deleted Localization nginx-loc"
```

---

## Phase 1: Data Model & Contracts

### 1.1 修改返回值类型

**文件**: `internal/executor/interfaces.go`

```go
// StageExecutor interface
type StageExecutor interface {
    ExecuteStage(...) (*drv1alpha1.StageStatus, error)  // ✅ 已有
    
    // RevertStage reverts a stage execution (returns rollback status)
    // 修改前: RevertStage(...) error
    // 修改后:
    RevertStage(
        ctx context.Context,
        plan *drv1alpha1.DRPlan,
        stage *drv1alpha1.Stage,
        originalStageStatus *drv1alpha1.StageStatus,  // 原执行的状态
    ) (*drv1alpha1.StageStatus, error)  // 返回回滚操作的状态
}

// WorkflowExecutor interface
type WorkflowExecutor interface {
    ExecuteWorkflow(...) (*drv1alpha1.WorkflowExecutionStatus, error)  // ✅ 已有
    
    // RevertWorkflow reverts a workflow execution (returns rollback status)
    // 修改前: RevertWorkflow(...) error
    // 修改后:
    RevertWorkflow(
        ctx context.Context,
        plan *drv1alpha1.DRPlan,
        workflowRef drv1alpha1.WorkflowReference,
        originalWorkflowStatus *drv1alpha1.WorkflowExecutionStatus,  // 原状态
    ) (*drv1alpha1.WorkflowExecutionStatus, error)  // 返回回滚状态
}

// ActionExecutor interface
type ActionExecutor interface {
    Execute(...) (*drv1alpha1.ActionStatus, error)  // ✅ 已有
    
    // Rollback rolls back an action (returns rollback status)
    // 修改前: Rollback(...) error
    // 修改后:
    Rollback(
        ctx context.Context,
        action *drv1alpha1.Action,
        originalActionStatus *drv1alpha1.ActionStatus,  // 原状态
    ) (*drv1alpha1.ActionStatus, error)  // 返回回滚状态
}
```

### 1.2 状态示例 Contract

**文件**: `specs/001-drplan-action-executor/contracts/drplanexecution-revert-status.yaml`

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: nginx-revert-001
spec:
  planRef: nginx-plan
  operationType: Revert
  revertExecutionRef: nginx-exec-001  # 回滚这个 execution

status:
  phase: Succeeded
  startTime: "2026-02-03T10:10:00Z"
  completionTime: "2026-02-03T10:12:00Z"
  
  # ✅ 新增：详细的回滚状态记录
  stageStatuses:
    - name: deploy-stage
      phase: Succeeded
      startTime: "2026-02-03T10:10:00Z"
      completionTime: "2026-02-03T10:11:30Z"
      duration: "1m30s"
      message: "Stage reverted successfully"
      
      workflowExecutions:
        - workflowRef:
            name: nginx-deploy-workflow
            namespace: default
          phase: Succeeded
          startTime: "2026-02-03T10:10:00Z"
          completionTime: "2026-02-03T10:11:30Z"
          duration: "1m30s"
          progress: "2/2 actions rolled back"
          
          actionStatuses:
            - name: create-localization-a
              type: Localization
              phase: Succeeded
              startTime: "2026-02-03T10:10:00Z"
              completionTime: "2026-02-03T10:10:45Z"
              message: "Rolled back: deleted Localization nginx-loc-a"
              retries: 0
              
            - name: create-localization-b
              type: Localization
              phase: Succeeded
              startTime: "2026-02-03T10:10:45Z"
              completionTime: "2026-02-03T10:11:30Z"
              message: "Rolled back: deleted Localization nginx-loc-b"
              retries: 0
  
  # ✅ 新增：统计信息
  summary:
    totalStages: 1
    completedStages: 1
    succeededStages: 1
    failedStages: 0
    skippedStages: 0
    totalActions: 2
    succeededActions: 2
    failedActions: 0
    skippedActions: 0
  
  message: "Plan reverted successfully: 1 stage, 2 actions rolled back"
```

---

## Phase 2: Implementation Tasks

### Task Breakdown

| ID | Task | Files | Priority | Estimate |
|----|------|-------|----------|----------|
| **T1** | 修改 ActionExecutor.Rollback 返回类型和实现 | `internal/executor/action_executor.go` | P0 | 2h |
| **T2** | 修改 WorkflowExecutor.RevertWorkflow 返回类型和实现 | `internal/executor/workflow_executor.go` | P0 | 2h |
| **T3** | 修改 StageExecutor.RevertStage 返回类型和实现 | `internal/executor/stage_executor.go` | P0 | 2h |
| **T4** | 更新 NativePlanExecutor.RevertPlan 记录状态 | `internal/executor/native_executor.go` | P0 | 3h |
| **T5** | 更新单元测试 | `internal/executor/*_test.go` | P1 | 2h |
| **T6** | 更新文档和示例 | `specs/*/`, `example/*/` | P1 | 1h |
| **T7** | 端到端测试验证 | Manual | P1 | 1h |

**总工作量**: 约 13 小时（1.5 工作日）

---

## Phase 3: Implementation Details

### T1: ActionExecutor.Rollback 返回 ActionStatus

**文件**: `internal/executor/action_executor.go`

**修改前**:
```go
func (e *ActionExecutor) Rollback(ctx context.Context, action *Action) error {
    // 执行回滚
    err := e.deleteResource(...)
    return err  // ❌ 只返回错误
}
```

**修改后**:
```go
func (e *ActionExecutor) Rollback(
    ctx context.Context,
    action *Action,
    originalStatus *ActionStatus,  // 新增：原状态
) (*ActionStatus, error) {
    // 创建回滚状态对象
    rollbackStatus := &ActionStatus{
        Name:      originalStatus.Name,
        Type:      originalStatus.Type,
        Phase:     "Running",
        StartTime: &metav1.Time{Time: time.Now()},
    }
    
    // 执行回滚逻辑
    switch action.Type {
    case "Localization":
        if action.Localization.Operation == "Create" {
            // 自动回滚：删除资源
            err := e.client.Delete(ctx, localization)
            if err != nil {
                rollbackStatus.Phase = "Failed"
                rollbackStatus.Message = fmt.Sprintf("Failed to delete Localization: %v", err)
                return rollbackStatus, err
            }
            rollbackStatus.Message = fmt.Sprintf("Rolled back: deleted Localization %s", loc.Name)
        } else {
            // 自定义回滚
            // ...
        }
    case "Job":
        // 删除 Job
        // ...
    }
    
    // 标记成功
    rollbackStatus.Phase = "Succeeded"
    rollbackStatus.CompletionTime = &metav1.Time{Time: time.Now()}
    
    return rollbackStatus, nil
}
```

### T2: WorkflowExecutor.RevertWorkflow 返回 WorkflowExecutionStatus

**文件**: `internal/executor/workflow_executor.go`

**修改后**:
```go
func (e *NativeWorkflowExecutor) RevertWorkflow(
    ctx context.Context,
    plan *DRPlan,
    workflowRef WorkflowReference,
    originalStatus *WorkflowExecutionStatus,
) (*WorkflowExecutionStatus, error) {
    // 1. 创建回滚状态对象
    rollbackStatus := &WorkflowExecutionStatus{
        WorkflowRef:    workflowRef,
        Phase:          "Running",
        StartTime:      &metav1.Time{Time: time.Now()},
        ActionStatuses: []ActionStatus{},
    }
    
    // 2. 逆序回滚 actions
    for i := len(originalStatus.ActionStatuses) - 1; i >= 0; i-- {
        originalActionStatus := originalStatus.ActionStatuses[i]
        
        if originalActionStatus.Phase != "Succeeded" {
            // 跳过失败的 action
            skippedStatus := ActionStatus{
                Name:    originalActionStatus.Name,
                Type:    originalActionStatus.Type,
                Phase:   "Skipped",
                Message: "Original action did not succeed, skipped rollback",
            }
            rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, skippedStatus)
            continue
        }
        
        // 执行回滚
        rollbackActionStatus, err := e.actionExecutor.Rollback(ctx, action, &originalActionStatus)
        rollbackStatus.ActionStatuses = append(rollbackStatus.ActionStatuses, *rollbackActionStatus)
        
        if err != nil {
            rollbackStatus.Phase = "Failed"
            rollbackStatus.Message = fmt.Sprintf("Failed to rollback action %s", action.Name)
            return e.finalizeWorkflowStatus(rollbackStatus), err
        }
    }
    
    // 3. 标记成功
    rollbackStatus.Phase = "Succeeded"
    rollbackStatus.Progress = fmt.Sprintf("%d/%d actions rolled back",
        len(originalStatus.ActionStatuses), len(originalStatus.ActionStatuses))
    
    return e.finalizeWorkflowStatus(rollbackStatus), nil
}
```

### T3: StageExecutor.RevertStage 返回 StageStatus

**文件**: `internal/executor/stage_executor.go`

**修改后**:
```go
func (e *NativeStageExecutor) RevertStage(
    ctx context.Context,
    plan *DRPlan,
    stage *Stage,
    originalStageStatus *StageStatus,
) (*StageStatus, error) {
    // 1. 创建回滚状态对象
    rollbackStageStatus := &StageStatus{
        Name:               originalStageStatus.Name,
        Phase:              "Running",
        Parallel:           originalStageStatus.Parallel,
        DependsOn:          originalStageStatus.DependsOn,
        StartTime:          &metav1.Time{Time: time.Now()},
        WorkflowExecutions: []WorkflowExecutionStatus{},
    }
    
    // 2. 逆序回滚 workflows
    for i := len(originalStageStatus.WorkflowExecutions) - 1; i >= 0; i-- {
        originalWorkflowStatus := originalStageStatus.WorkflowExecutions[i]
        
        if originalWorkflowStatus.Phase != "Succeeded" {
            // 跳过失败的 workflow
            continue
        }
        
        // 执行回滚
        rollbackWorkflowStatus, err := e.workflowExecutor.RevertWorkflow(
            ctx, plan, originalWorkflowStatus.WorkflowRef, &originalWorkflowStatus)
        
        rollbackStageStatus.WorkflowExecutions = append(
            rollbackStageStatus.WorkflowExecutions, *rollbackWorkflowStatus)
        
        if err != nil {
            rollbackStageStatus.Phase = "Failed"
            rollbackStageStatus.Message = fmt.Sprintf("Failed to rollback workflow %s", 
                originalWorkflowStatus.WorkflowRef.Name)
            return e.finalizeStageStatus(rollbackStageStatus), err
        }
    }
    
    // 3. 标记成功
    rollbackStageStatus.Phase = "Succeeded"
    return e.finalizeStageStatus(rollbackStageStatus), nil
}
```

### T4: NativePlanExecutor.RevertPlan 记录状态

**文件**: `internal/executor/native_executor.go`

**修改前** (L300-386):
```go
func (e *NativePlanExecutor) RevertPlan(...) error {
    // ... 验证逻辑 ...
    
    for i := len(targetExecution.Status.StageStatuses) - 1; i >= 0; i-- {
        stageStatus := targetExecution.Status.StageStatuses[i]
        if stageStatus.Phase != "Succeeded" {
            continue
        }
        
        // ❌ 不返回状态
        err := e.stageExecutor.RevertStage(ctx, plan, stage, &stageStatus)
        if err != nil {
            return fmt.Errorf("failed to revert stage %s: %w", stage.Name, err)
        }
    }
    
    execution.Status.Phase = "Succeeded"
    execution.Status.Message = "Plan reverted successfully"
    return e.updateExecutionStatus(ctx, execution, nil)
}
```

**修改后**:
```go
func (e *NativePlanExecutor) RevertPlan(...) error {
    klog.Infof("Reverting plan: %s/%s", plan.Namespace, plan.Name)
    
    // ✅ 1. 初始化 Revert execution 的状态结构
    if execution.Status.StageStatuses == nil {
        execution.Status.StageStatuses = make([]StageStatus, 0, len(targetExecution.Status.StageStatuses))
    }
    if execution.Status.Summary == nil {
        execution.Status.Summary = &ExecutionSummary{
            TotalStages: len(targetExecution.Status.StageStatuses),
        }
    }
    
    // ... 验证逻辑（revertExecutionRef 等）...
    
    // ✅ 2. 逆序回滚 stages，记录每个 stage 的回滚状态
    for i := len(targetExecution.Status.StageStatuses) - 1; i >= 0; i-- {
        originalStageStatus := targetExecution.Status.StageStatuses[i]
        
        if originalStageStatus.Phase != "Succeeded" {
            klog.V(4).Infof("Skipping revert for stage %s (phase=%s)", 
                originalStageStatus.Name, originalStageStatus.Phase)
            
            // ✅ 记录跳过的 stage
            skippedStageStatus := StageStatus{
                Name:    originalStageStatus.Name,
                Phase:   "Skipped",
                Message: fmt.Sprintf("Original stage phase was %s, skipped rollback", 
                    originalStageStatus.Phase),
            }
            e.updateStageStatusInExecution(execution, &skippedStageStatus)
            continue
        }
        
        // 查找 stage 定义
        var stage *Stage
        for j := range plan.Spec.Stages {
            if plan.Spec.Stages[j].Name == originalStageStatus.Name {
                stage = &plan.Spec.Stages[j]
                break
            }
        }
        
        if stage == nil {
            klog.Warningf("Stage %s not found in plan definition", originalStageStatus.Name)
            continue
        }
        
        klog.Infof("Reverting stage: %s", stage.Name)
        
        // ✅ 执行回滚并获取状态
        rollbackStageStatus, err := e.stageExecutor.RevertStage(
            ctx, plan, stage, &originalStageStatus)
        
        // ✅ 更新 execution 的状态
        e.updateStageStatusInExecution(execution, rollbackStageStatus)
        
        if err != nil {
            klog.Errorf("Failed to revert stage %s: %v", stage.Name, err)
            execution.Status.Phase = "Failed"
            execution.Status.Message = fmt.Sprintf("Failed to revert stage %s: %v", stage.Name, err)
            return e.updateExecutionStatus(ctx, execution, nil)
        }
        
        // Check stage rollback result
        if rollbackStageStatus.Phase == "Failed" {
            execution.Status.Phase = "Failed"
            execution.Status.Message = fmt.Sprintf("Stage %s rollback failed", stage.Name)
            return e.updateExecutionStatus(ctx, execution, nil)
        }
    }
    
    // ✅ 3. 标记整体成功
    execution.Status.Phase = "Succeeded"
    
    // ✅ 4. 生成详细的成功消息
    succeededStages := 0
    skippedStages := 0
    totalActions := 0
    for _, ss := range execution.Status.StageStatuses {
        if ss.Phase == "Succeeded" {
            succeededStages++
            for _, ws := range ss.WorkflowExecutions {
                totalActions += len(ws.ActionStatuses)
            }
        } else if ss.Phase == "Skipped" {
            skippedStages++
        }
    }
    
    execution.Status.Message = fmt.Sprintf(
        "Plan reverted successfully: %d stage(s) rolled back, %d action(s) rolled back, %d stage(s) skipped",
        succeededStages, totalActions, skippedStages)
    
    return e.updateExecutionStatus(ctx, execution, nil)
}
```

---

## Phase 4: Testing

### T5: 单元测试

**新增测试文件**: `internal/executor/native_executor_revert_test.go`

```go
func TestRevertPlan_WithStatusTracking(t *testing.T) {
    // Setup
    executeExecution := &DRPlanExecution{
        ObjectMeta: metav1.ObjectMeta{Name: "exec-001"},
        Spec: DRPlanExecutionSpec{
            OperationType: "Execute",
            PlanRef: "test-plan",
        },
        Status: DRPlanExecutionStatus{
            Phase: "Succeeded",
            StageStatuses: []StageStatus{
                {
                    Name: "stage-1",
                    Phase: "Succeeded",
                    WorkflowExecutions: []WorkflowExecutionStatus{
                        {
                            WorkflowRef: ObjectReference{Name: "workflow-1"},
                            Phase: "Succeeded",
                            ActionStatuses: []ActionStatus{
                                {Name: "action-1", Type: "Localization", Phase: "Succeeded"},
                            },
                        },
                    },
                },
            },
        },
    }
    
    revertExecution := &DRPlanExecution{
        ObjectMeta: metav1.ObjectMeta{Name: "revert-001"},
        Spec: DRPlanExecutionSpec{
            OperationType: "Revert",
            PlanRef: "test-plan",
            RevertExecutionRef: "exec-001",
        },
    }
    
    // Execute
    err := executor.RevertPlan(ctx, plan, revertExecution)
    
    // Verify
    assert.NoError(t, err)
    assert.Equal(t, "Succeeded", revertExecution.Status.Phase)
    
    // ✅ 验证状态记录
    assert.NotEmpty(t, revertExecution.Status.StageStatuses, "Should have stage statuses")
    assert.Equal(t, 1, len(revertExecution.Status.StageStatuses))
    
    stageStatus := revertExecution.Status.StageStatuses[0]
    assert.Equal(t, "stage-1", stageStatus.Name)
    assert.Equal(t, "Succeeded", stageStatus.Phase)
    assert.NotNil(t, stageStatus.StartTime)
    assert.NotNil(t, stageStatus.CompletionTime)
    
    assert.NotEmpty(t, stageStatus.WorkflowExecutions)
    workflowStatus := stageStatus.WorkflowExecutions[0]
    assert.Equal(t, "Succeeded", workflowStatus.Phase)
    
    assert.NotEmpty(t, workflowStatus.ActionStatuses)
    actionStatus := workflowStatus.ActionStatuses[0]
    assert.Equal(t, "Succeeded", actionStatus.Phase)
    assert.Contains(t, actionStatus.Message, "Rolled back")
    
    // ✅ 验证 Summary
    assert.NotNil(t, revertExecution.Status.Summary)
    assert.Equal(t, 1, revertExecution.Status.Summary.TotalStages)
    assert.Equal(t, 1, revertExecution.Status.Summary.SucceededStages)
}
```

### T7: 端到端测试

**测试场景**:

1. **正常回滚**:
   ```bash
   # Execute
   kubectl apply -f drplanexecution-execute.yaml
   kubectl wait --for=condition=Completed drplanexecution/exec-001
   
   # Verify Execute status
   kubectl get drplanexecution exec-001 -o yaml | yq '.status.stageStatuses'
   
   # Revert
   kubectl apply -f drplanexecution-revert.yaml
   kubectl wait --for=condition=Completed drplanexecution/revert-001
   
   # ✅ Verify Revert status
   kubectl get drplanexecution revert-001 -o yaml | yq '.status'
   # 应该包含:
   # - stageStatuses (与 exec-001 数量一致)
   # - summary.succeededStages
   # - 详细的 message
   ```

2. **部分回滚失败**:
   ```bash
   # 模拟：删除 action 创建的资源（导致回滚时找不到）
   kubectl delete localization nginx-loc-a
   
   # Revert
   kubectl apply -f drplanexecution-revert.yaml
   
   # ✅ Verify: 应该记录失败的 action
   kubectl get drplanexecution revert-001 -o yaml | yq '.status.stageStatuses[0].workflowExecutions[0].actionStatuses'
   # 应该看到 phase: Failed 和具体错误信息
   ```

---

## Acceptance Criteria

### 功能验证

- [ ] **AC1**: Revert execution 的 `status.stageStatuses` 包含所有被回滚的 stage
- [ ] **AC2**: 每个 `stageStatus` 包含 `startTime`, `completionTime`, `duration`
- [ ] **AC3**: 每个 `workflowExecution` 包含回滚的 action 状态列表
- [ ] **AC4**: 每个 `actionStatus` 包含 `phase`, `message`（如 "Rolled back: deleted Localization xxx"）
- [ ] **AC5**: `status.summary` 正确统计成功/失败/跳过的 stage 和 action 数量
- [ ] **AC6**: Revert 失败时，`status` 清晰指示失败的 stage/workflow/action

### 性能验证

- [ ] **AC7**: Revert 操作的额外开销 < 5%（主要是状态对象创建和序列化）
- [ ] **AC8**: 大规模 Plan（50 stages, 200 actions）的 Revert status 可正常记录

### 兼容性验证

- [ ] **AC9**: 旧版本创建的 Execute execution 仍可被新版本 Revert
- [ ] **AC10**: 升级后，旧的 Revert execution（无 stageStatuses）仍可查看

---

## Rollout Plan

### Phase 1: Development (Week 1)
- Day 1-2: T1-T3 (修改 executor 接口和实现)
- Day 3: T4 (更新 RevertPlan)
- Day 4: T5 (单元测试)
- Day 5: T6-T7 (文档和 E2E 测试)

### Phase 2: Internal Testing (Week 2)
- 在开发集群部署和测试
- 验证 Acceptance Criteria
- 性能基准测试

### Phase 3: Rollout (Week 3)
- 发布 patch 版本（如 v0.2.1）
- 更新 Helm chart
- 通知用户升级

---

## Open Questions

1. **Q**: Revert 的 `actionStatus.message` 格式是否需要标准化？
   - **A**: 建议格式: `"Rolled back: {operation} {resourceType} {resourceName}"`
   - 例如: `"Rolled back: deleted Localization nginx-loc-a"`

2. **Q**: 如果 Revert 部分失败（3 个 action，2 个成功 1 个失败），整体 phase 应该是 Failed 还是 PartiallyFailed？
   - **A**: 设为 `Failed`，但 `summary` 应显示部分成功信息，让用户知道哪些已回滚

3. **Q**: Revert 的 `stageStatuses` 顺序应该是正序还是逆序？
   - **A**: **逆序**（与回滚执行顺序一致），方便用户从上到下阅读回滚过程

---

## References

- [Execute 状态记录实现](internal/executor/native_executor.go#L219-296)
- [StageStatus 数据结构](api/v1alpha1/drplanexecution_types.go#L74-110)
- [Kubernetes API Conventions - Status](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status)
