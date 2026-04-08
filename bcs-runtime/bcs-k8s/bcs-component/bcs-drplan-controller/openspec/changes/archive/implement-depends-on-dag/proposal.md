## Why

当前 DRWorkflow 中所有 Global action 严格顺序执行，无法并行。
当需要对标 Helm hook-weight 语义（同 weight 的 hook 并行执行、跨 weight 组顺序等待）时，
必须将多个可并行的 action 拆到不同 workflow 或接受顺序执行的性能损失。
`DependsOn` 字段已在 API 中定义（Phase 2 预留），现在实现它以支持 DAG 调度。

## What Changes

- 实现 `Action.DependsOn` 字段的运行时 DAG 调度逻辑
- 没有互相依赖的 action 可并行执行；有依赖的 action 等其所有依赖完成后再启动
- 在 workflow 执行前校验 `DependsOn` 合法性（引用不存在的 action、循环依赖）
- 向后兼容：不设 `DependsOn` 的 workflow 行为与当前完全一致（顺序执行）
- 回滚（RevertWorkflow）保持按执行记录逆序，DAG 并行组内的 action 一起逆序处理

## Capabilities

### New Capabilities

- `workflow-dag-execution`: workflow action 的 DAG 并行调度能力，基于 `DependsOn` 字段构建有向无环图，支持同层并行、跨层顺序等待、循环依赖检测

### Modified Capabilities

（无，API 类型层面无变更，`DependsOn` 字段已存在）

## Impact

- `internal/executor/native_executor.go`：`ExecuteWorkflow` 核心调度逻辑重构为 DAG topo-sort + 并发执行
- `internal/executor/`：新增 DAG 构建与校验工具函数（可单独测试）
- `api/v1alpha1/common_types.go`：无需修改（字段已存在）
- 单元测试：新增 DAG 调度相关测试用例
- 已有测试：不带 `DependsOn` 的 workflow 回归测试需通过
