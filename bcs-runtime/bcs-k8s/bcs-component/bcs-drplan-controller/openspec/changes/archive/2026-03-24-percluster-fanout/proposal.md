# Proposal: Per-Cluster Fan-out Execution for Subscription Actions

## Why

当前 `waitReady` 机制等待 **所有** binding clusters 的 feeds 全部 ready 后才继续执行下一个 action。这与 Helm 的预期行为不一致——Helm 在单个集群上完成 pre-install hook 后，该集群即可继续执行后续步骤（主资源部署、post-install hook），而不需要等其他集群。

在大规模多集群场景下，一个慢集群会拖慢所有集群的执行进度，造成不必要的延迟。

## What Changes

### 1. Action API 扩展

在 `Action` struct 上新增 `clusterExecutionMode` 字段：

- **空值/`Global`**（默认）：保持现有行为，Subscription 作为整体执行，waitReady 等待所有集群。存量配置无需修改。
- **`PerCluster`**：运行时按 binding clusters 拆分为独立的子 Subscription，每个集群可独立推进后续 action。

### 2. Status 模型扩展

`ActionStatus` 新增 `clusterStatuses` 字段（`[]ClusterActionStatus`），记录每个集群的执行状态。对于 `Global` action 或未设置 `clusterExecutionMode` 的 action，此字段为空（兼容旧行为）。

### 3. Workflow Executor 调度改造

将连续的 `PerCluster` action 分组为 "per-cluster batch"，每个集群独立地按顺序执行 batch 内的 action。在 `PerCluster → Global` 边界处设置全局屏障（barrier），确保所有集群完成后再执行 Global action。

### 4. Subscription Executor 拆分逻辑

`PerCluster` 模式下，Subscription executor 为每个 binding cluster 生成独立的子 Subscription（单集群 subscriber），并独立追踪 readiness。

### 5. drplan-gen 适配

- Hook Subscription action → `clusterExecutionMode: PerCluster`
- 主资源 Subscription action → `clusterExecutionMode: Global`（或不填）

## Impact

| 组件 | 影响程度 | 说明 |
|------|---------|------|
| `api/v1alpha1/common_types.go` | 中 | 新增字段和类型 |
| `api/v1alpha1/constants.go` | 低 | 新增常量 |
| `internal/executor/native_executor.go` | 高 | 重构 workflow executor 调度逻辑 |
| `internal/executor/subscription_executor.go` | 高 | 新增子 Subscription 生成和清理 |
| `internal/generator/` | 低 | drplan-gen 输出 `clusterExecutionMode` |
| CRD YAML | 自动 | `make manifests` 重新生成 |

### 兼容性

- **存量 DRWorkflow/DRPlan**：`clusterExecutionMode` 为空值，等价 `Global`，行为完全不变。
- **存量 DRPlanExecution status**：`ActionStatus.ClusterStatuses` 为空值，不影响现有消费者。
- **drplan-gen 旧版输出**：不含 `clusterExecutionMode` 字段，运行时按 `Global` 处理。
