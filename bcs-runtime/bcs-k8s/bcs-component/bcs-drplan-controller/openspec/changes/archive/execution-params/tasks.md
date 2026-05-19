# Tasks: Execution-Level Params with Dynamic valueFrom

## 1. API 层 — 数据模型扩展

- [x] 1.1 扩展 `Parameter` struct，新增 `ValueFrom` 字段及相关类型  <!-- TDD 任务 -->
  - [x] 1.1.1 写失败测试：`internal/webhook/drplanexecution_webhook_test.go`
    （测试 `value` 与 `valueFrom` 互斥、`name` 与 `labelSelector` 互斥等 webhook 校验）
  - [x] 1.1.2 验证测试失败（运行：`go test ./internal/webhook/...`，确认因类型不存在而编译失败）
  - [x] 1.1.3 写最小实现：`api/v1alpha1/common_types.go`
    （添加 `ValueFrom *ParameterValueFrom`、`ParameterValueFrom`、`ManifestRef` 类型）
  - [x] 1.1.4 验证测试通过（运行：`go test ./internal/webhook/...`，确认所有 webhook 测试通过）
  - [x] 1.1.5 重构：整理类型注释、kubebuilder markers、字段顺序

- [x] 1.2 为 `DRPlanExecutionSpec` 新增 `Params []Parameter` 字段  <!-- TDD 任务 -->
  - [x] 1.2.1 写失败测试：`internal/webhook/drplanexecution_webhook_test.go`
    （测试 params 重名、格式等 webhook 校验）
  - [x] 1.2.2 验证测试失败（运行：`go test ./internal/webhook/...`，确认失败）
  - [x] 1.2.3 写最小实现：`api/v1alpha1/drplanexecution_types.go`
    （在 `RevertExecutionRef` 后添加 `Params []Parameter`）
  - [x] 1.2.4 验证测试通过（运行：`go test ./internal/webhook/...`，确认通过）
  - [x] 1.2.5 重构：检查 webhook 验证逻辑是否完整

- [x] 1.3 扩展 `SubscriptionAction.Operation` 枚举，新增 `Apply`  <!-- 非 TDD 任务 -->
  - [x] 1.3.1 执行变更：`api/v1alpha1/common_types.go`（修改 kubebuilder:validation:Enum marker）
  - [x] 1.3.2 验证无回归（运行：`go build ./...`，确认编译通过）
  - [x] 1.3.3 检查：确认 marker 已更新，无其他 Enum 引用遗漏

- [x] 1.4 重新生成 DeepCopy 与 CRD  <!-- 非 TDD 任务 -->
  - [x] 1.4.1 执行变更：`make generate && make manifests`
  - [x] 1.4.2 验证无回归（运行：`go test ./api/...`，确认通过）
  - [x] 1.4.3 检查：确认 `zz_generated.deepcopy.go` 包含新类型，CRD yaml 反映新字段

- [x] 1.5 代码审查（全量测试通过，无阻塞问题）

---

## 2. Executor 层 — valueFrom 参数解析

- [x] 2.1 新建 `param_resolver.go`，实现 `resolveParams` 函数  <!-- TDD 任务 -->
  - [x] 2.1.1 写失败测试：`internal/executor/param_resolver_test.go`
  - [x] 2.1.2 验证测试失败
  - [x] 2.1.3 写最小实现：`internal/executor/param_resolver.go`
  - [x] 2.1.4 验证测试通过（8/8 通过）
  - [x] 2.1.5 重构：提取 jsonPath 求值逻辑为独立函数，改善错误信息

- [x] 2.2 在 `native_executor.go` 中集成 execution.Spec.Params 解析与优先级合并  <!-- TDD 任务 -->
  - [x] 2.2.1 写失败测试：`internal/executor/native_executor_test.go`
  - [x] 2.2.2 验证测试失败
  - [x] 2.2.3 写最小实现：`internal/executor/native_executor.go`
  - [x] 2.2.4 验证测试通过（2/2 通过）
  - [x] 2.2.5 重构：为 `NativePlanExecutor` 添加 `dynamicClient`、`mapper` 字段，更新构造函数

- [x] 2.3 代码审查（全量测试通过，无阻塞问题）

---

## 3. Subscription Executor — Apply 操作支持

- [x] 3.1 在 `subscription_executor.go` 中支持 `operation: Apply`  <!-- TDD 任务 -->
  - [x] 3.1.1 写失败测试：`internal/executor/subscription_executor_test.go`
  - [x] 3.1.2 验证测试失败（Apply 报 AlreadyExists）
  - [x] 3.1.3 写最小实现：`internal/executor/subscription_executor.go`（添加 `applySubscription`）
  - [x] 3.1.4 验证测试通过（4/4 通过）
  - [x] 3.1.5 重构：统一错误包装风格

- [x] 3.2 代码审查（全量测试通过，无阻塞问题）

---

## 4. Documentation Sync (Required)

- [x] 4.1 sync design.md: 已补充 when 条件说明、实际实现细节
- [x] 4.2 sync tasks.md: 全量标记所有已完成任务
- [x] 4.3 sync proposal.md: 范围与实现一致，无变化
- [x] 4.4 sync specs/execution-params.md: 已补充 when + mode 说明
- [x] 4.5 Final review: 所有 OpenSpec 文档已同步
