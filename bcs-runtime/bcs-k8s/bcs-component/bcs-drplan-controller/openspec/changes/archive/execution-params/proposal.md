# Proposal: Execution-Level Params with Dynamic valueFrom

## Why

当前 DRPlan 参数只能在 `DRPlan.spec.globalParams` 或 `DRWorkflow.spec.parameters` 中
静态定义。每次 Helm upgrade 产生新 revision 时（如 Job 名含 `{{ .Release.Revision }}`），
用户必须手动修改 DRPlan 资源才能让 Subscription feeds 指向新 Job。

这带来两个痛点：

1. **DRPlanExecution 无法携带执行时参数**：每次执行都要改 DRPlan，导致 Plan 与
   Execution 职责混淆——Plan 描述"怎么做"，Execution 才是"这次具体用什么参数做"。

2. **参数只能静态填写**：如果想让 Job 名从父集群自动发现（按标签取最新），
   目前完全无法实现，必须依赖外部流水线注入。

## What Changes

### 1. DRPlanExecution 支持执行级参数

`DRPlanExecution.spec` 新增可选 `params` 字段，类型为 `[]Parameter`。
执行时与 DRPlan.globalParams 合并（execution.params 优先级最高）。

### 2. Parameter 扩展 valueFrom

`Parameter` struct 新增可选 `valueFrom` 字段，支持从 K8s 资源动态解析参数值：

- `valueFrom.manifestRef`：按 apiVersion/kind + name 精确匹配，或按 labelSelector
  列举后取最新（creationTimestamp Last），再用 jsonPath 提取字段值。
- 支持层级：`DRPlanExecution.spec.params` 和 `DRPlan.spec.globalParams`。

### 3. Subscription executor 支持 Apply 操作

`SubscriptionAction.operation` 新增 `Apply` 语义（server-side apply），
使同名 Subscription 可以在多次执行间更新 feeds，不再报 AlreadyExists。

## Impact

| 变更点 | 文件 |
|---|---|
| API 类型扩展 | `api/v1alpha1/common_types.go`、`api/v1alpha1/drplanexecution_types.go` |
| DeepCopy 重新生成 | `api/v1alpha1/zz_generated.deepcopy.go` |
| CRD 重新生成 | `config/crd/bases/`、`install/helm/.../crds/` |
| 执行器扩展 | `internal/executor/native_executor.go`（globalParams 解析）|
| 新增解析器 | `internal/executor/param_resolver.go`（valueFrom 逻辑）|
| Subscription 执行器 | `internal/executor/subscription_executor.go`（Apply 支持）|
| Webhook 验证 | `internal/webhook/drplanexecution_webhook.go`、`internal/webhook/drplan_validator.go` |
