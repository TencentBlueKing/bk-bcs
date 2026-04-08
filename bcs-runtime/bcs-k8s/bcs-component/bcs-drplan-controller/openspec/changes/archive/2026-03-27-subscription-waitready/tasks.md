## 1. Subscription waitReady (SocketProxy 直查)

- [x] 1.1 回退当前工作区中非 OpenSpec 文档改动（保持工作区干净）
  - [x] 1.1.1 执行变更：恢复到当前分支 HEAD（不包含本次新生成的 `openspec/changes/` 文档）
  - [x] 1.1.2 验证无回归（运行：`git status`，确认仅剩 OpenSpec 文档改动或工作区干净）
  - [x] 1.1.3 检查：确认不会丢失需要保留的未提交工作

- [x] 1.2 为 Action 增加 `waitReady` 字段（API/CRD）
  - [x] 1.2.1 写失败测试：不适用（CRD 结构变更）
  - [x] 1.2.2 验证测试失败：不适用
  - [x] 1.2.3 写最小实现：`api/v1alpha1/common_types.go`
  - [x] 1.2.4 验证测试通过（运行：`make generate && make manifests && go test ./... -count=1`）
  - [x] 1.2.5 重构：检查字段位置、默认值与 json tag，保持兼容

- [x] 1.3 实现 ChildClusterClientFactory（默认 SocketProxy，预留扩展点）
  - [x] 1.3.1 写失败测试：`internal/executor/cluster_client_test.go`
  - [x] 1.3.2 验证测试失败（运行：`go test ./internal/executor -count=1`）
  - [x] 1.3.3 写最小实现：`internal/executor/cluster_client.go`
  - [x] 1.3.4 验证测试通过（运行：`go test ./internal/executor -count=1`）
  - [x] 1.3.5 重构：抽象接口与默认实现，预留 kubeconfig 分支但不实现

- [x] 1.4 Subscription executor 增加 waitReady 两阶段等待（bindingClusters + 子集群 feeds readiness）
  - [x] 1.4.1 写失败测试：`internal/executor/subscription_waitready_test.go`
  - [x] 1.4.2 验证测试失败（运行：`go test ./internal/executor -count=1`）
  - [x] 1.4.3 写最小实现：`internal/executor/subscription_executor.go`
  - [x] 1.4.4 验证测试通过（运行：`go test ./... -count=1`）
  - [x] 1.4.5 重构：提炼 readiness 判定函数与错误信息，控制圈复杂度与 lint

- [x] 1.5 初始化注入与 RBAC
  - [x] 1.5.1 执行变更：`cmd/main.go`（注入 `*rest.Config` / 预留 factory 选择）、`internal/controller/drplanexecution_controller.go`（RBAC markers）
  - [x] 1.5.2 验证无回归（运行：`make manifests && go test ./... -count=1`）
  - [x] 1.5.3 检查：确认生成的 `config/rbac/role.yaml` 权限最小且覆盖所需资源

- [x] 1.6 代码审查
  - 前置：调用 superpowers:verification-before-completion 运行全量测试
  - 调用 superpowers:requesting-code-review 审查本任务组变更
  - 占位符：
    - `{PLAN_OR_REQUIREMENTS}` → `openspec/changes/2026-03-27-subscription-waitready/specs/subscription-waitready.md` + `openspec/changes/2026-03-27-subscription-waitready/tasks.md`
    - `{WHAT_WAS_IMPLEMENTED}` → `api/v1alpha1/common_types.go`、`internal/executor/*`、`cmd/main.go`、`internal/controller/drplanexecution_controller.go`、生成物
    - `{BASE_SHA}` → 任务组开始前 commit SHA
    - `{HEAD_SHA}` → 当前 HEAD
  - Critical/Important → 停等用户指令
  - Minor/无问题 → 自动继续

## 2. drplan-gen 按 hook 语义自动设置 waitReady（并注入 when）

- [x] 2.1 更新生成器为 hook Subscription Action 写入 `waitReady: true` 和 `when`
  - [x] 2.1.1 写失败测试：`internal/generator/planner_test.go`
  - [x] 2.1.2 验证测试失败（运行：`go test ./internal/generator -count=1`）
  - [x] 2.1.3 写最小实现：`internal/generator/planner.go`
  - [x] 2.1.4 验证测试通过（运行：`go test ./... -count=1`）
  - [x] 2.1.5 重构：统一 Subscription action 字段输出顺序与命名；移除 CLI `--wait` 输入路径；按 hook 类型写入 when

- [x] 2.2 更新 golden files 与集成测试断言
  - [x] 2.2.1 执行变更：更新 `testdata/output/workflow-install.yaml` / `drplan.yaml`、同步 `internal/generator/integration_test.go`
  - [x] 2.2.2 验证无回归（运行：`go test ./... -count=1`）
  - [x] 2.2.3 检查：确认“有 hook 注解/无 hook 注解”两套场景 golden files 一致可复现

- [x] 2.3 代码审查
  - 前置：调用 superpowers:verification-before-completion 运行全量测试
  - 调用 superpowers:requesting-code-review 审查本任务组变更
  - 占位符：
    - `{PLAN_OR_REQUIREMENTS}` → `openspec/changes/2026-03-27-subscription-waitready/specs/subscription-waitready.md` + `openspec/changes/2026-03-27-subscription-waitready/tasks.md`
    - `{WHAT_WAS_IMPLEMENTED}` → `internal/generator/*` + `testdata/output/*`
    - `{BASE_SHA}` → 任务组开始前 commit SHA
    - `{HEAD_SHA}` → 当前 HEAD
  - Critical/Important → 停等用户指令
  - Minor/无问题 → 自动继续

## 4. mode + when 执行分流（install/upgrade）

- [x] 4.1 扩展 execution API
  - [x] 4.1.1 执行变更：`DRPlanExecution.spec` 增加可选 `mode: Install|Upgrade`
  - [x] 4.1.2 验证无回归（运行：`make generate && make manifests && go test ./... -count=1`）

- [x] 4.2 执行器支持 Action.when（最小语义）
  - [x] 4.2.1 执行变更：支持 `mode == "install|upgrade"` 单条件判断（兼容旧 `operation == ...` 写法）
  - [x] 4.2.2 兼容策略：未提供 mode 时不按 when 过滤（全执行）
  - [x] 4.2.3 失败策略：不支持表达式时报错并失败
  - [x] 4.2.4 验证无回归（运行：`go test ./... -count=1 && make lint`）

- [x] 4.3 多 hook 值拆分
  - [x] 4.3.1 执行变更：支持 `helm.sh/hook: a,b` 拆分到多个 hook 类型
  - [x] 4.3.2 测试覆盖：分类与生成断言通过

## 3. Documentation Sync (Required)

- [x] 3.1 sync design.md: 记录技术决策、偏差和实现细节
- [x] 3.2 sync tasks.md: **全量标记**所有层级任务（顶层 + 子任务），将已完成的 `[ ]` 标记为 `[x]`
- [x] 3.3 sync proposal.md: 更新范围/影响（若有变化）
- [x] 3.4 sync specs/*.md: 更新功能需求（若有变化）并确保与实现一致
- [x] 3.5 Final review: 确保所有 OpenSpec 文档反映实际实现

