## Context

`DRWorkflow.spec.actions` 当前由 `NativeWorkflowExecutor.ExecuteWorkflow` 严格按数组顺序逐个执行（Global 模式）。
`Action.DependsOn []string` 字段已在 `api/v1alpha1/common_types.go` 中定义，但注释为 "reserved for DAG, Phase 2"，运行时被忽略。

已有的并行能力（`ClusterExecutionMode=PerCluster`）仅作用于跨集群维度，workflow 内 action 之间仍无法并行。

目标场景：对标 Helm hook-weight，实现同一 workflow 内多个 pre/post hook 并行执行。

## Goals / Non-Goals

**Goals:**

- 实现 `DependsOn` 字段的 DAG 调度：无依赖关系的 action 并行执行，有依赖的 action 等待所有依赖完成后启动
- 执行前校验：循环依赖、引用不存在的 action 名称，校验失败直接使 workflow 失败
- 向后兼容：所有 action 均无 `DependsOn` 时，退化为现有顺序执行（无行为差异）
- 回滚（RevertWorkflow）按执行记录逆序，DAG 并行组内 action 一起放入同一逆序批次

**Non-Goals:**

- 不支持跨 workflow/stage 的 DAG
- 不修改 Argo 执行引擎路径
- 不在 webhook 校验阶段做 DependsOn 合法性检查（仅在运行时校验）
- 不引入动态 action 插入（action 列表在执行前固定）

## Decisions

### 决策 1：使用 Kahn 算法（BFS 拓扑排序）按层调度

将 action 列表构建为有向图，用 Kahn 算法按拓扑层（level）展开：
- 同层（无互相依赖）action 使用 `sync.WaitGroup` + goroutine 并发执行
- 跨层严格等待上层全部完成（含 `waitReady`）

**备选方案**：DFS 拓扑排序或 channel-based 调度。
Kahn 算法天然产生"层"结构，与 Helm hook-weight 分组语义直接对应，实现简单且易于测试。

### 决策 2：混合模式兼容——无 DependsOn 退化为顺序

检测条件：所有 action 的 `DependsOn` 均为空时，沿用现有 `executeSingleGlobalAction` 串行路径，零行为变化。

**原因**：避免对现有 workflow 引入 goroutine 开销和状态聚合复杂度，最小化回归风险。

### 决策 3：校验在 ExecuteWorkflow 入口处做

在进入调度循环前，调用独立的 `validateDependsOn(actions)` 函数检测：
1. 引用了不存在 action 名的依赖
2. 循环依赖（Kahn 算法完成后剩余节点数 > 0）

校验失败直接将 workflow 状态置为 Failed 并返回，不执行任何 action。

**原因**：运行时校验比 webhook 校验更简单，且不需要额外 RBAC/webhook 配置。

### 决策 4：ActionStatus 聚合顺序按拓扑层内的原始定义顺序

并发执行的同层 action 完成后，按它们在 `workflow.Spec.Actions` 中的原始顺序追加到 `status.ActionStatuses`，保证 status 可读性和回滚逆序的确定性。

## Risks / Trade-offs

- [风险] 并发 action 中某个失败时，`FailFast` 策略需要取消其他正在执行的 action
  → 用 `context.WithCancel` 传递取消信号，各 goroutine 在 action 执行前检查 `ctx.Err()`

- [风险] 同层并发写 `status.ActionStatuses` 存在竞态
  → 各 goroutine 写入独立 slice，层执行完成后主线程按顺序合并，无需锁

- [风险] `DependsOn` 引用当前实现被忽略，存量 workflow 无 DependsOn 字段，升级后行为不变
  → 退化路径（决策 2）保证向后兼容

## Migration Plan

1. 无需数据迁移，API 字段已存在
2. 发布后存量 workflow 行为不变（决策 2）
3. 新增 DependsOn 的 workflow 需用户主动配置，不自动迁移
4. 回滚策略：删除 DAG 调度代码，恢复原 for 循环，API 字段留空不影响功能

## Open Questions

- 无
