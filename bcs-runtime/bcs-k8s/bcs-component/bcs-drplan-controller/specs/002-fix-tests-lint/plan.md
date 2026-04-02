# Implementation Plan: 完善测试与修复Lint错误

**Branch**: `002-fix-tests-lint` | **Date**: 2026-02-04 | **Spec**: [spec.md](./spec.md)  
**Input**: Feature specification from `specs/002-fix-tests-lint/spec.md`

## Summary

提升 bcs-drplan-controller 的代码质量和测试覆盖率，确保代码通过 lint 检查和 CI/CD 流程。主要包括：
1. **Lint修复**（P1）：修复14个funlen错误、goconst魔法字符串、gosec安全问题、package注释缺失
2. **单元测试**（P2）：提升Controller（60%+）、Executor（70%+）、Webhook（80%+）的测试覆盖率
3. **E2E测试**（P3）：补充完整的Kind+Clusternet集成测试，验证灾备功能的端到端流程

当前状态：测试覆盖率26.1%，14个lint错误阻塞代码合并。目标：覆盖率≥55%，lint无错误，CI/CD全部通过。

## Technical Context

**Language/Version**: Go 1.22+  
**Primary Dependencies**: 
- controller-runtime v0.19 (Kubernetes operator框架)
- Ginkgo v2 + Gomega (BDD测试框架)
- envtest (单元测试中的K8s API模拟)
- golangci-lint v1.56+ (代码质量检查)
- Kind v0.20+ (E2E测试的K8s集群)
- Clusternet v0.18+ (多集群管理，E2E测试依赖)

**Storage**: etcd (Kubernetes API Server提供，envtest/Kind中包含)  
**Testing**: 
- Unit: Ginkgo v2 + Gomega + envtest
- E2E: Ginkgo v2 + Kind + Clusternet
- Coverage: go test -coverprofile

**Target Platform**: Linux (amd64/arm64), Kubernetes 1.28+  
**Project Type**: Kubernetes Operator (Kubebuilder scaffolded)

**Performance Goals**: 
- 单元测试执行时间 <3分钟
- E2E测试执行时间 ≤20分钟（含环境搭建）
- Lint检查时间 <2分钟

**Constraints**: 
- 不改变现有功能行为（仅重构和补充测试）
- 测试必须可重复执行（幂等、隔离）
- E2E测试必须在CI环境可运行（Docker-in-Docker）

**Scale/Scope**: 
- 代码文件：~50个Go文件需要修复lint
- 测试文件：需新增约15个测试文件
- 测试用例：预计新增100+个测试场景
- Lint错误：14个funlen + 若干goconst/gosec/revive

## Constitution Check

*此feature为质量改进，不涉及新的项目或复杂架构引入*

✅ **PASS** - 无需额外复杂度审核

## Project Structure

### Documentation (this feature)

```text
specs/002-fix-tests-lint/
├── spec.md                 # Feature规范（已完成）
├── plan.md                 # 本文件（技术实施计划）
├── checklists/
│   └── requirements.md     # 规范质量检查清单（已完成）
└── [tasks.md]              # 待生成：任务分解（/speckit.tasks）
```

### Source Code (existing Kubebuilder project structure)

```text
bcs-drplan-controller/
├── api/v1alpha1/
│   ├── *_types.go          # CRD定义（需添加package注释）
│   └── constants.go        # 新增：状态常量定义（FR-002）
│
├── internal/
│   ├── controller/
│   │   ├── *_controller.go       # Controller实现（需重构长函数）
│   │   ├── *_controller_test.go  # 单元测试（需补充，目标60%+）
│   │   └── suite_test.go          # 测试套件（已存在）
│   │
│   ├── executor/
│   │   ├── *_executor.go         # Executor实现（需重构长函数）
│   │   └── *_executor_test.go    # 单元测试（需新建，目标70%+）
│   │
│   └── webhook/
│       ├── *_webhook.go          # Webhook实现
│       └── *_webhook_test.go     # 单元测试（需新建，目标80%+）
│
├── test/e2e/
│   ├── e2e_test.go               # E2E测试（需补充业务场景）
│   ├── e2e_suite_test.go         # E2E测试套件（已存在）
│   └── scripts/                  # 新增：E2E自动化脚本（FR-025）
│       ├── setup-e2e-env.sh     # 搭建Kind+Clusternet环境
│       ├── run-e2e-tests.sh     # 执行测试
│       ├── cleanup-e2e-env.sh   # 清理资源
│       └── e2e-full.sh          # Wrapper：完整流程
│
├── .golangci.yml                 # Lint配置（需调整测试文件例外）
├── Makefile                      # 构建工具（可能需添加e2e-*目标）
└── E2E_TESTING_GUIDE.md          # E2E测试指南（已存在）
```

**Structure Decision**: 使用标准Kubebuilder项目结构，新增内容：
1. `api/v1alpha1/constants.go` - 集中定义状态常量
2. `internal/executor/*_test.go` - Executor单元测试（当前缺失）
3. `internal/webhook/*_test.go` - Webhook单元测试（当前缺失）
4. `test/e2e/scripts/` - E2E自动化脚本目录

## Phase 0: Research & Preparation

### 0.1 Lint错误分析

**目标**：全面了解当前14个funlen错误的分布和重构策略

**任务**：
- [ ] 运行 `make lint | tee lint-errors.txt`，记录所有错误详情
- [ ] 对14个funlen错误按文件分组，优先级排序（Controller > Executor > 其他）
- [ ] 对每个过长函数进行职责分析，识别可拆分的职责边界
- [ ] 记录goconst错误涉及的字符串（"Create"、"Patch"等），规划常量命名
- [ ] 检查gosec错误的上下文，确定哪些需要豁免（#nosec）

**输出**：`research.md` - Lint错误详细分析和重构策略

### 0.2 测试覆盖率基线

**目标**：建立当前测试覆盖率基线，识别未覆盖的关键代码路径

**任务**：
- [ ] 运行 `make test`，生成 `cover.out`
- [ ] 使用 `go tool cover -func=cover.out | tee coverage-baseline.txt`，记录各文件覆盖率
- [ ] 识别覆盖率为0的文件（Executor、Webhook）
- [ ] 识别Controller层覆盖率<60%的函数，列出需补充的测试场景
- [ ] 检查 `.cursor/rules/unit-testing-standards.mdc`，确认测试规范

**输出**：`research.md` 附加章节 - 测试覆盖率分析

### 0.3 E2E测试现状

**目标**：了解当前E2E测试框架和缺失的业务场景

**任务**：
- [ ] 阅读 `E2E_TESTING_GUIDE.md`，理解方案A（Kind+Clusternet）的搭建步骤
- [ ] 运行现有E2E测试：`make test-e2e`（如果存在）
- [ ] 检查 `test/e2e/e2e_test.go:267` 的TODO注释，确认需要补充的场景
- [ ] 检查 `example/plan/install/` 中的示例资源，作为E2E测试的测试数据

**输出**：`research.md` 附加章节 - E2E测试现状和TODO

### 0.4 CI/CD流程分析

**目标**：了解当前CI配置，规划lint/test/e2e的集成策略

**任务**：
- [ ] 检查 `.github/workflows/` 或 CI配置文件
- [ ] 确认当前lint、test在CI中的执行时机
- [ ] 规划E2E测试的CI集成（PR验证时执行）
- [ ] 评估CI环境的Docker-in-Docker能力

**输出**：`research.md` 附加章节 - CI/CD集成策略

**预计耗时**：1-2天

---

## Phase 1: Lint修复与常量定义（P1 - 阻塞性问题）

### 1.1 定义常量文件

**目标**：创建 `api/v1alpha1/constants.go`，集中定义所有状态常量

**任务**：
- [ ] 创建 `api/v1alpha1/constants.go`
- [ ] 定义Phase常量（Pending、Running、Succeeded、Failed等）
- [ ] 定义OperationType常量（Execute、Revert）
- [ ] 定义Resource Operation常量（Create、Patch、Delete、Apply）
- [ ] 定义其他常量（default namespace、rollback messages等）
- [ ] 添加package注释和常量分组注释

**验证**：
```bash
# 常量文件可编译
go build ./api/v1alpha1/
```

**预计耗时**：0.5天

### 1.2 修复goconst错误

**目标**：将所有魔法字符串替换为常量引用

**任务**：
- [ ] 批量替换：`"Create"` → `drv1alpha1.OperationCreate`
- [ ] 批量替换：`"Patch"` → `drv1alpha1.OperationPatch`
- [ ] 批量替换：`"Delete"` → `drv1alpha1.OperationDelete`
- [ ] 批量替换：`"Succeeded"` → `drv1alpha1.PhaseSucceeded`
- [ ] 批量替换：`"Failed"` → `drv1alpha1.PhaseFailed`
- [ ] 批量替换其他重复字符串

**验证**：
```bash
make lint | grep goconst
# 预期：无输出
```

**预计耗时**：0.5天

### 1.3 重构funlen错误（按单一职责原则）

**目标**：将14个过长函数重构为职责清晰的小函数

**策略**（基于澄清Q3）：
- 拆分原则：每个函数一个职责（validation、building、execution、error handling）
- 函数命名：使用清晰的名称（如`validateInput`），而非`step1`
- 保持原有逻辑流程，只改变代码组织

**任务**：
- [ ] **Controller层**（约5个funlen）：
  - `DRPlanReconciler.Reconcile` → 拆分为 `validatePlan`、`reconcileStages`、`updateStatus`
  - `DRPlanExecutionReconciler.Reconcile` → 拆分为 `validateExecution`、`executeOrRevert`、`updateStatus`
  - 其他Controller函数按需拆分
  
- [ ] **Executor层**（约7个funlen）：
  - `HTTPActionExecutor.Execute` → 拆分为 `buildRequest`、`sendRequest`、`handleResponse`
  - `JobActionExecutor.Execute` → 拆分为 `createJobSpec`、`createJob`、`waitForCompletion`
  - `LocalizationActionExecutor.Execute` → 拆分为 `buildLocalizationSpec`、`createCR`、`handleResult`
  - 其他Executor函数按需拆分

- [ ] **其他层**（约2个funlen）：
  - `NativePlanExecutor.ExecutePlan/RevertPlan` 等

**验证**：
```bash
make lint | grep funlen
# 预期：无输出

make test
# 预期：所有测试通过（重构不改变行为）
```

**预计耗时**：3-4天

### 1.4 修复gosec和revive错误

**目标**：修复安全问题和代码风格问题

**任务**：
- [ ] **gosec - TLS InsecureSkipVerify**：
  - 检查 `http_executor.go:181` 的上下文
  - 如果确实是测试场景，添加 `// #nosec G402 - Used in test environment only`
  
- [ ] **gosec - Subprocess with variable**：
  - 检查 `test/utils/utils.go` 的kubectl命令
  - 添加 `// #nosec G204 - kubectl path is trusted` 或重构为安全版本

- [ ] **revive - package comments**：
  - 为 `cmd/main.go` 添加：`// Package main is the entry point for bcs-drplan-controller.`
  - 为 `internal/controller` 添加：`// Package controller implements Kubernetes controllers for DR resources.`
  - 为其他包添加规范注释

**验证**：
```bash
make lint
# 预期：退出码0，无任何错误或警告
```

**预计耗时**：0.5天

### 1.5 配置golangci.yml测试文件例外

**目标**：为测试文件配置dot imports例外规则

**任务**：
- [ ] 编辑 `.golangci.yml`
- [ ] 在 `linters-settings.revive.rules` 中添加测试文件例外：
  ```yaml
  - name: dot-imports
    severity: warning
    exclude: ["_test\\.go$"]
  ```
- [ ] 验证Ginkgo/Gomega的dot imports不再报错

**验证**：
```bash
make lint
# 预期：测试文件的dot imports不再报错
```

**预计耗时**：0.5天

**Phase 1 总计**：5-6天

---

## Phase 2: 单元测试补充（P2 - 核心质量保障）

### 2.1 Executor单元测试（目标70%+）

**目标**：为7个executor创建完整的测试文件

**任务**：
- [ ] **HTTP Executor** (`internal/executor/http_executor_test.go`):
  - 测试GET/POST/PUT/DELETE请求构建
  - 测试HTTP响应处理（2xx成功、4xx/5xx失败）
  - 测试Rollback场景（如果有自定义rollback action）
  - Mock HTTP server使用 `httptest.NewServer`
  
- [ ] **Job Executor** (`internal/executor/job_executor_test.go`):
  - 测试Job Spec构建（包含参数渲染）
  - 测试Job创建和状态监控
  - 测试Job成功/失败场景
  - 使用envtest创建真实Job资源

- [ ] **Localization Executor** (`internal/executor/localization_executor_test.go`):
  - 测试Localization CR创建（Create/Patch/Delete操作）
  - 测试namespace正确性（ManagedCluster namespace）
  - 测试参数渲染
  - 测试Rollback（删除CR）

- [ ] **Subscription Executor** (`internal/executor/subscription_executor_test.go`):
  - 测试Subscription CR创建
  - 测试Feed列表构建
  - 测试ClusterAffinity渲染
  - 测试Rollback

- [ ] **K8sResource Executor** (`internal/executor/k8s_resource_executor_test.go`):
  - 测试通用K8s资源创建/更新/删除
  - 测试动态client使用
  - 测试资源状态检查

- [ ] **Stage Executor** (`internal/executor/stage_executor_test.go`):
  - 测试串行执行（Sequential）
  - 测试并行执行（Parallel）
  - 测试FailurePolicy（Stop/Continue）
  - 测试Stage Rollback

- [ ] **Native Executor** (`internal/executor/native_executor_test.go`):
  - 测试完整的Workflow/Plan执行流程
  - 测试错误传播
  - 测试状态更新

**测试规范**（遵循 `.cursor/rules/unit-testing-standards.mdc`）：
- 使用 `Describe("ExecutorName", func() {})` 组织
- 使用 `Context("When xxx", func() {})` 描述场景
- 使用 `It("should xxx", func() {})` 描述期望
- 使用 `By("doing xxx")` 分步骤
- 使用 `Eventually()` 处理异步操作
- 使用常量而非字符串字面量

**验证**：
```bash
go test ./internal/executor/... -coverprofile=executor-cover.out
go tool cover -func=executor-cover.out | grep total
# 预期：total coverage ≥70%
```

**预计耗时**：5-6天（平行开发）

### 2.2 Controller单元测试补充（目标60%+）

**目标**：补充Controller层测试用例，提升覆盖率

**任务**：
- [ ] **DRPlan Controller** (`internal/controller/drplan_controller_test.go`):
  - 补充测试：Plan validation失败场景
  - 补充测试：Stage reconciliation
  - 补充测试：Status更新（Ready/Invalid）
  - 补充测试：Finalizer处理

- [ ] **DRPlanExecution Controller** (`internal/controller/drplanexecution_controller_test.go`):
  - 补充测试：Execute操作完整流程
  - 补充测试：Revert操作完整流程
  - 补充测试：并发Execution冲突检测
  - 补充测试：Plan状态更新

- [ ] **DRWorkflow Controller** (`internal/controller/drworkflow_controller_test.go`):
  - 补充测试：Workflow validation
  - 补充测试：Action依赖检查
  - 补充测试：Rollback定义验证

**验证**：
```bash
make test
go tool cover -func=cover.out | grep 'internal/controller'
# 预期：每个controller文件≥60%
```

**预计耗时**：3-4天

### 2.3 Webhook单元测试（目标80%+）

**目标**：为3个webhook创建完整的测试文件

**任务**：
- [ ] **DRPlan Webhook** (`internal/webhook/drplan_webhook_test.go`):
  - 测试ValidateCreate：拒绝空stages
  - 测试ValidateUpdate：拒绝不兼容的修改
  - 测试ValidateDelete：检查依赖的Execution
  - 测试Default：填充默认值

- [ ] **DRWorkflow Webhook** (`internal/webhook/drworkflow_webhook_test.go`):
  - 测试ValidateCreate：
    - 拒绝无rollback的Patch操作
    - 拒绝循环依赖
    - 拒绝无效的action类型
  - 测试参数校验

- [ ] **DRPlanExecution Webhook** (`internal/webhook/drplanexecution_webhook_test.go`):
  - 测试ValidateCreate：检查Plan存在性
  - 测试Revert类型的revertExecutionRef必填
  - 测试并发Execution检测

**测试方式**：
- 创建测试资源，调用webhook方法
- 验证返回的warnings和errors
- 不需要envtest（webhook是纯函数）

**验证**：
```bash
go test ./internal/webhook/... -coverprofile=webhook-cover.out
go tool cover -func=webhook-cover.out | grep total
# 预期：total coverage ≥80%
```

**预计耗时**：2-3天

**Phase 2 总计**：10-13天

---

## Phase 3: E2E测试与CI集成（P3 - 集成验证）

### 3.1 E2E自动化脚本

**目标**：创建分离但可组合的E2E脚本（基于澄清Q1）

**任务**：
- [ ] **创建 `test/e2e/scripts/setup-e2e-env.sh`**：
  - 检测Kind集群是否存在，不存在则创建
  - 安装CertManager（如需要）
  - 安装Clusternet Hub
  - 安装Clusternet Agent（模拟子集群）
  - 幂等设计：重复执行不报错
  - 输出：环境就绪状态

- [ ] **创建 `test/e2e/scripts/run-e2e-tests.sh`**：
  - 构建controller镜像
  - 加载镜像到Kind
  - 部署controller
  - 运行E2E测试：`go test -tags=e2e ./test/e2e/...`
  - 输出：测试结果

- [ ] **创建 `test/e2e/scripts/cleanup-e2e-env.sh`**：
  - 删除controller部署
  - 删除测试资源
  - 可选：删除Kind集群（通过参数控制）

- [ ] **创建 `test/e2e/scripts/e2e-full.sh`**：
  - Wrapper脚本，依次调用 setup → run → cleanup
  - 错误处理：任何步骤失败时执行cleanup
  - 输出：完整流程状态

**验证**：
```bash
./test/e2e/scripts/e2e-full.sh
# 预期：完整流程成功，总耗时≤20分钟
```

**预计耗时**：2天

### 3.2 E2E业务场景测试

**目标**：补充 `test/e2e/e2e_test.go` 的TODO部分

**任务**：
- [ ] **Scenario 1: DRPlan创建与验证**：
  ```go
  It("should create DRPlan and become Ready", func() {
    By("creating DRWorkflows")
    // Apply subscription + localization workflows
    
    By("creating DRPlan")
    // Apply plan YAML
    
    By("waiting for plan to be Ready")
    Eventually(func() string {
      // Get plan status.phase
    }).Should(Equal(drv1alpha1.PlanPhaseReady))
  })
  ```

- [ ] **Scenario 2: Execute操作**：
  ```go
  It("should execute plan and create Clusternet resources", func() {
    By("creating DRPlanExecution with OperationType=Execute")
    // Apply execution YAML
    
    By("waiting for execution to complete")
    Eventually(func() string {
      // Get execution status.phase
    }, "10m", "10s").Should(Equal(drv1alpha1.PhaseSucceeded))
    
    By("verifying Subscription created")
    // kubectl get subscription
    
    By("verifying Localization created")
    // kubectl get localization -n <namespace>
    
    By("verifying metrics")
    metricsOutput := getMetricsOutput()
    Expect(metricsOutput).To(ContainSubstring(
      `controller_runtime_reconcile_total{controller="drplanexecution",result="success"}`))
  })
  ```

- [ ] **Scenario 3: Revert操作**：
  ```go
  It("should revert plan and cleanup resources", func() {
    By("creating DRPlanExecution with OperationType=Revert")
    // Apply revert execution YAML
    
    By("waiting for revert to complete")
    Eventually(func() string {
      // Get execution status.phase
    }).Should(Equal(drv1alpha1.PhaseSucceeded))
    
    By("verifying resources deleted")
    // kubectl get subscription (should be NotFound)
    // kubectl get localization (should be NotFound)
  })
  ```

- [ ] **Scenario 4: Webhook验证**：
  ```go
  It("should reject invalid DRWorkflow", func() {
    By("creating Workflow with Patch but no rollback")
    // Apply invalid workflow YAML
    // Expect error from webhook
  })
  ```

- [ ] **Scenario 5: FailurePolicy**：
  ```go
  It("should stop on failure when FailurePolicy=Stop", func() {
    By("creating Plan with failing action")
    // Configure an action that will fail
    
    By("creating Execution")
    Eventually(func() string {
      // Get execution status.phase
    }).Should(Equal(drv1alpha1.PhaseFailed))
    
    By("verifying subsequent stages not executed")
    // Check stage statuses
  })
  ```

**测试数据**：使用 `example/plan/install/` 中的YAML作为基础，适当修改

**验证**：
```bash
make test-e2e
# 预期：所有5个场景通过
```

**预计耗时**：4-5天

### 3.3 CI/CD集成

**目标**：配置CI流程执行lint、test、e2e（基于澄清Q2）

**任务**：
- [ ] **更新CI配置**（如`.github/workflows/test.yml`）：
  ```yaml
  on:
    push:
      branches: [ main, master, develop ]
    pull_request:
      branches: [ main, master ]
  
  jobs:
    lint:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
        - run: make lint
    
    test:
      runs-on: ubuntu-latest
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
        - run: make test
        - run: go tool cover -func=cover.out
    
    e2e:
      runs-on: ubuntu-latest
      if: github.event_name == 'pull_request'  # 仅PR时执行
      steps:
        - uses: actions/checkout@v4
        - uses: actions/setup-go@v5
        - run: ./test/e2e/scripts/e2e-full.sh
  ```

- [ ] **添加Makefile目标**（如需要）：
  ```makefile
  .PHONY: e2e-setup
  e2e-setup:
  	@./test/e2e/scripts/setup-e2e-env.sh
  
  .PHONY: e2e-run
  e2e-run:
  	@./test/e2e/scripts/run-e2e-tests.sh
  
  .PHONY: e2e-cleanup
  e2e-cleanup:
  	@./test/e2e/scripts/cleanup-e2e-env.sh
  
  .PHONY: e2e-full
  e2e-full:
  	@./test/e2e/scripts/e2e-full.sh
  ```

**验证**：
- 提交commit，检查lint和test任务执行
- 创建PR，检查e2e任务执行
- 所有检查通过后PR可合并

**预计耗时**：1天

**Phase 3 总计**：7-8天

---

## Testing Strategy

### Unit Testing

**Framework**: Ginkgo v2 + Gomega + envtest

**Approach**:
- Controller层：使用envtest提供的fake K8s API，测试reconciliation逻辑
- Executor层：部分使用envtest（如Job、K8sResource），部分使用mock（如HTTP）
- Webhook层：直接调用validation方法，不需要envtest

**Coverage Targets**:
- Controller: ≥60%
- Executor: ≥70%
- Webhook: ≥80%
- Overall: ≥55%

**Execution**:
```bash
make test
go tool cover -html=cover.out -o coverage.html
```

### E2E Testing

**Framework**: Ginkgo v2 + Kind + Clusternet

**Approach**:
- 使用真实的Kind集群和Clusternet
- 测试完整的灾备流程：Plan创建 → Execute → Revert
- 验证Clusternet资源的实际创建和删除
- 通过metrics验证reconciliation成功

**Scenarios**: 5个核心场景（见Phase 3.2）

**Execution**:
```bash
./test/e2e/scripts/e2e-full.sh
# 或分步执行
make e2e-setup
make e2e-run
make e2e-cleanup
```

### Lint Checking

**Tool**: golangci-lint v1.56+

**Checks**:
- funlen: ≤40语句或≤60行
- goconst: 无重复字符串字面量
- gosec: 无安全问题（或有豁免说明）
- revive: 有package注释，测试文件允许dot imports

**Execution**:
```bash
make lint
```

---

## Risk & Dependencies

### Risks

| Risk | Impact | Mitigation |
|------|--------|-----------|
| 函数重构引入bug | High | 每次重构后立即运行测试，确保行为不变 |
| E2E测试不稳定 | Medium | 使用Eventually而非固定sleep，增加重试机制 |
| CI环境无Docker-in-Docker | Medium | 提前验证CI环境，或配置专用runner |
| 测试覆盖率目标过高 | Low | 优先覆盖核心路径，边界情况可适当放宽 |
| E2E测试耗时超20分钟 | Low | 优化环境搭建脚本，使用缓存 |

### Dependencies

**External**:
- Kind v0.20+ 需预先安装（本地和CI）
- Helm v3 需预先安装（用于Clusternet安装）
- Clusternet Helm charts可访问

**Internal**:
- 项目的 `.cursor/rules/unit-testing-standards.mdc` 已存在
- 项目的 `E2E_TESTING_GUIDE.md` 已存在
- 项目的 `example/plan/install/` 示例YAML已存在

**Blocking**:
- Phase 1（Lint修复）必须在Phase 2（测试补充）之前完成
- Phase 2必须在Phase 3（E2E）之前完成
- 无其他外部依赖阻塞

---

## Success Criteria

**Lint**:
- [ ] `make lint` 退出码0，无任何错误或警告

**Unit Test**:
- [ ] `make test` 所有测试通过
- [ ] Controller层覆盖率≥60%
- [ ] Executor层覆盖率≥70%
- [ ] Webhook层覆盖率≥80%
- [ ] 项目整体覆盖率≥55%

**E2E Test**:
- [ ] 5个核心场景测试通过
- [ ] E2E完整流程（setup+run+cleanup）≤20分钟
- [ ] Clusternet资源正确创建和删除
- [ ] Metrics显示reconcile成功

**CI/CD**:
- [ ] 每次commit自动运行lint和test（约3分钟）
- [ ] PR创建/更新时自动运行e2e（约20分钟）
- [ ] 所有检查通过后允许合并

---

## Estimated Timeline

| Phase | Tasks | Duration | Dependencies |
|-------|-------|----------|--------------|
| **Phase 0** | Research & Preparation | 1-2天 | None |
| **Phase 1** | Lint修复与常量定义 | 5-6天 | Phase 0 |
| **Phase 2** | 单元测试补充 | 10-13天 | Phase 1 |
| **Phase 3** | E2E测试与CI集成 | 7-8天 | Phase 2 |
| **Buffer** | 调试、返工、文档 | 2-3天 | - |
| **Total** | | **25-32天** | |

**并行机会**：
- Phase 2中，Executor/Controller/Webhook测试可并行开发（如有多人）
- Phase 3中，脚本开发和测试场景编写可并行

**里程碑**：
- ✅ M1（Day 7）：Lint全部修复，`make lint` 通过
- ✅ M2（Day 20）：单元测试覆盖率达标，`make test` 通过
- ✅ M3（Day 28）：E2E测试完成，CI集成完成

---

## Next Steps

1. **开始Phase 0**：运行 `make lint` 和 `make test`，生成基线报告
2. **创建tasks.md**：使用 `/speckit.tasks` 将本计划分解为可执行的任务卡片
3. **设置开发环境**：确保本地有Kind、Helm等工具
4. **并行工作**：如有多人，可分配不同的executor/controller/webhook

**下一个命令**：`/speckit.tasks` - 生成任务分解
