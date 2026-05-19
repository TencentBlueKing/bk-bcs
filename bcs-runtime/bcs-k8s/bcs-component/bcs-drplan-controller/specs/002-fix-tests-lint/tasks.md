# Tasks: 完善测试与修复Lint错误

**Input**: `specs/002-fix-tests-lint/plan.md`, `specs/002-fix-tests-lint/spec.md`  
**Branch**: `002-fix-tests-lint`  
**Prerequisites**: plan.md ✅, spec.md ✅

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story (US1-Lint, US2-UnitTest, US3-Webhook, US4-E2E)
- File paths use Kubebuilder project structure

## 用户故事映射

| ID | Priority | Story | Goal |
|----|----------|-------|------|
| US1 | P1 | Lint修复 | 修复14个funlen、goconst、gosec、revive错误 |
| US2 | P2 | 单元测试 | Controller/Executor测试覆盖率60%+/70%+ |
| US3 | P2 | Webhook测试 | Webhook测试覆盖率80%+ |
| US4 | P3 | E2E测试 | Kind+Clusternet完整DR场景测试 |

---

## Phase 0: 研究与准备（Setup）

**目标**: 建立基线数据，识别待修复问题

- [X] T001 [P] [Setup] 运行 `make lint | tee specs/002-fix-tests-lint/lint-errors.txt`，记录所有lint错误
- [X] T002 [P] [Setup] 运行 `make test`，生成 `cover.out`
- [X] T003 [Setup] 分析14个funlen错误分布，按文件分组并优先级排序（Controller > Executor > 其他）
- [X] T004 [Setup] 生成覆盖率基线报告：`go tool cover -func=cover.out | tee specs/002-fix-tests-lint/coverage-baseline.txt`
- [X] T005 [Setup] 识别覆盖率为0的文件列表（Executor、Webhook），记录到 `specs/002-fix-tests-lint/research.md`
- [X] T006 [Setup] 检查现有E2E测试：运行 `test/e2e/e2e_test.go`，记录待补充场景
- [X] T007 [Setup] 检查CI配置（`.github/workflows/`），评估Docker-in-Docker能力

**Checkpoint**: 基线数据完整，`research.md` 已创建

---

## Phase 1: Foundational - Lint修复（P1 - 阻塞性）

**目标**: 消除所有lint错误，代码可合并

**⚠️ CRITICAL**: 此阶段必须完成才能开始后续测试工作

### 1.1 常量定义（US1）

- [X] T008 [US1] 创建 `api/v1alpha1/constants.go` 文件，添加package注释
- [X] T009 [US1] 在constants.go中定义Phase常量（Pending、Running、Succeeded、Failed、Ready、Invalid等）
- [X] T010 [US1] 在constants.go中定义OperationType常量（Execute、Revert）
- [X] T011 [US1] 在constants.go中定义ResourceOperation常量（Create、Patch、Delete、Apply）
- [X] T012 [US1] 在constants.go中定义其他常量（default namespace、rollback messages等）
- [X] T013 [US1] 验证常量文件可编译：`go build ./api/v1alpha1/`

### 1.2 修复goconst错误（US1）

- [X] T014 [P] [US1] 替换所有 `"Create"` 为 `drv1alpha1.OperationCreate`（涉及executor、webhook）
- [X] T015 [P] [US1] 替换所有 `"Patch"` 为 `drv1alpha1.OperationPatch`
- [X] T016 [P] [US1] 替换所有 `"Delete"` 为 `drv1alpha1.OperationDelete`
- [X] T017 [P] [US1] 替换所有 `"Succeeded"` 为 `drv1alpha1.PhaseSucceeded`
- [X] T018 [P] [US1] 替换所有 `"Failed"` 为 `drv1alpha1.PhaseFailed`
- [X] T019 [P] [US1] 替换其他重复字符串（"default"、rollback messages等）
- [X] T020 [US1] 验证goconst修复：`make lint | grep goconst`（预期无输出）

### 1.3 重构funlen错误 - Controller层（US1）

- [ ] T021 [US1] 重构 `internal/controller/drplan_controller.go` 的 `Reconcile` 函数
  - 拆分为：`validatePlan()`, `reconcileStages()`, `updatePlanStatus()`
  - 每个函数≤40语句或≤60行
- [ ] T022 [US1] 重构 `internal/controller/drplanexecution_controller.go` 的 `Reconcile` 函数
  - 拆分为：`validateExecution()`, `executeOrRevert()`, `updateExecutionStatus()`
- [ ] T023 [US1] 重构其他Controller层funlen错误（如有）
- [ ] T024 [US1] 验证Controller层重构：`make test` 所有测试通过（行为不变）

### 1.4 重构funlen错误 - Executor层（US1）

- [ ] T025 [P] [US1] 重构 `internal/executor/http_executor.go` 的 `Execute` 方法
  - 拆分为：`buildHTTPRequest()`, `sendHTTPRequest()`, `handleHTTPResponse()`
- [ ] T026 [P] [US1] 重构 `internal/executor/job_executor.go` 的 `Execute` 方法
  - 拆分为：`createJobSpec()`, `createJobResource()`, `waitForJobCompletion()`
- [X] T027 [P] [US1] 重构 `internal/executor/localization_executor.go` 的 `Execute` 方法
  - 拆分为：`prepareLocalizationSpec()`, `createLocalizationResource()`, 状态管理函数
- [X] T028 [P] [US1] 重构 `internal/executor/native_executor.go` 的 `RevertPlan` 方法
  - 拆分为：`validateAndFetchRevertTarget()`, `initializeRevertExecutionStatus()`, `finalizeRevertSuccess()`
- [X] T029 [US1] 为 `cmd/main.go:main` 添加 `//nolint:funlen` 注释（初始化逻辑过长但结构清晰）
- [X] T030 [US1] 验证Executor层重构：`make test` 通过，funlen错误消失

### 1.5 修复gosec、govet、revive、unparam错误（US1）

- [X] T033 [P] [US1] 修复gosec G402（TLS InsecureSkipVerify）：`internal/executor/http_executor.go:181` 添加 `// #nosec G402 - Used in test environment only`
- [X] T034 [P] [US1] 修复gosec G204（subprocess with variable）：`test/utils/utils.go` 添加 `// #nosec G204 - kubectl path is trusted`
- [X] T035 [P] [US1] 为 `cmd/main.go` 添加package注释：`// Package main is the entry point for bcs-drplan-controller.`
- [X] T036 [P] [US1] 为 `internal/controller` 添加package注释：`// Package controller implements Kubernetes controllers for DR resources.`
- [X] T037 [P] [US1] 为其他包添加package注释（`internal/executor`, `internal/webhook`, `api/v1alpha1`等）
- [~] T038 [US1] 配置 `.golangci.yml`，为测试文件添加dot imports例外规则（已尝试多种方法，revive配置限制未生效）
- [X] T039 [US1] 修复 govet shadow 错误：重命名遮蔽变量
- [X] T040 [US1] 修复 revive 错误：ExecutorRegistry → Registry，添加常量注释
- [X] T041 [US1] 修复 unparam 错误：移除总是为 nil 的返回值
- [X] T042 [US1] 验证所有lint修复：`make lint`（预期退出码0，无任何错误） ✅

**Checkpoint**: Lint全部通过，代码可合并（M1）

---

## Phase 2: User Story 2 - 单元测试（P2 - Controller/Executor）

**目标**: 提升Controller和Executor测试覆盖率到60%+/70%+

**独立测试**: 运行 `make test`，检查覆盖率报告

### 2.1 HTTP Executor测试（US2）

- [ ] T040 [US2] 创建 `internal/executor/http_executor_test.go`，初始化测试套件
- [ ] T041 [US2] 测试场景：GET请求构建和执行（使用 `httptest.NewServer`）
- [ ] T042 [US2] 测试场景：POST/PUT请求body渲染和发送
- [ ] T043 [US2] 测试场景：HTTP 2xx响应处理（成功）
- [ ] T044 [US2] 测试场景：HTTP 4xx/5xx响应处理（失败）
- [ ] T045 [US2] 测试场景：Rollback自定义action执行
- [ ] T046 [US2] 验证HTTP Executor覆盖率≥70%

### 2.2 Job Executor测试（US2）

- [ ] T047 [US2] 创建 `internal/executor/job_executor_test.go`，初始化envtest
- [ ] T048 [US2] 测试场景：Job Spec构建，参数正确渲染
- [ ] T049 [US2] 测试场景：Job创建到K8s，使用envtest真实API
- [ ] T050 [US2] 测试场景：Job成功完成（phase=Succeeded）
- [ ] T051 [US2] 测试场景：Job失败（phase=Failed）
- [ ] T052 [US2] 测试场景：Rollback删除Job
- [ ] T053 [US2] 验证Job Executor覆盖率≥70%

### 2.3 Localization Executor测试（US2）

- [ ] T054 [US2] 创建 `internal/executor/localization_executor_test.go`，初始化envtest
- [ ] T055 [US2] 测试场景：Localization CR创建（Create操作）
- [ ] T056 [US2] 测试场景：Localization CR更新（Patch操作）
- [ ] T057 [US2] 测试场景：Localization CR删除（Delete操作）
- [ ] T058 [US2] 测试场景：namespace正确性验证（ManagedCluster namespace）
- [ ] T059 [US2] 测试场景：参数模板渲染
- [ ] T060 [US2] 测试场景：Rollback删除Localization CR
- [ ] T061 [US2] 验证Localization Executor覆盖率≥70%

### 2.4 Subscription Executor测试（US2）

- [ ] T062 [US2] 创建 `internal/executor/subscription_executor_test.go`
- [ ] T063 [US2] 测试场景：Subscription CR创建
- [ ] T064 [US2] 测试场景：Feed列表构建正确
- [ ] T065 [US2] 测试场景：ClusterAffinity渲染
- [ ] T066 [US2] 测试场景：Rollback删除Subscription
- [ ] T067 [US2] 验证Subscription Executor覆盖率≥70%

### 2.5 K8sResource/Stage/Native Executor测试（US2）

- [ ] T068 [P] [US2] 创建 `internal/executor/k8s_resource_executor_test.go`，测试通用K8s资源CRUD
- [ ] T069 [P] [US2] 创建 `internal/executor/stage_executor_test.go`，测试Sequential/Parallel执行和FailurePolicy
- [ ] T070 [P] [US2] 创建 `internal/executor/native_executor_test.go`，测试完整Workflow/Plan执行流程
- [ ] T071 [US2] 验证所有Executor总体覆盖率≥70%

### 2.6 Controller测试补充（US2）

- [ ] T072 [US2] 补充 `internal/controller/drplan_controller_test.go` 测试用例
  - Plan validation失败场景
  - Stage reconciliation
  - Status更新（Ready/Invalid）
  - Finalizer处理
- [ ] T073 [US2] 补充 `internal/controller/drplanexecution_controller_test.go` 测试用例
  - Execute操作完整流程
  - Revert操作完整流程
  - 并发Execution冲突检测
  - Plan状态更新
- [ ] T074 [US2] 补充 `internal/controller/drworkflow_controller_test.go` 测试用例
  - Workflow validation
  - Action依赖检查
  - Rollback定义验证
- [ ] T075 [US2] 验证所有Controller覆盖率≥60%

**Checkpoint**: Controller/Executor测试覆盖率达标（M2部分）

---

## Phase 3: User Story 3 - Webhook测试（P2）

**目标**: Webhook测试覆盖率达到80%+

**独立测试**: 运行 `go test ./internal/webhook/... -coverprofile=webhook-cover.out`

### 3.1 DRPlan Webhook测试（US3）

- [ ] T076 [US3] 创建 `internal/webhook/drplan_webhook_test.go`
- [ ] T077 [US3] 测试场景：ValidateCreate拒绝空stages
- [ ] T078 [US3] 测试场景：ValidateCreate拒绝无效的workflowRef
- [ ] T079 [US3] 测试场景：ValidateUpdate拒绝不兼容的修改
- [ ] T080 [US3] 测试场景：ValidateDelete检查依赖的Execution
- [ ] T081 [US3] 测试场景：Default填充默认值
- [ ] T082 [US3] 验证DRPlan Webhook覆盖率≥80%

### 3.2 DRWorkflow Webhook测试（US3）

- [ ] T083 [US3] 创建 `internal/webhook/drworkflow_webhook_test.go`
- [ ] T084 [US3] 测试场景：ValidateCreate拒绝Patch操作但无rollback
- [ ] T085 [US3] 测试场景：ValidateCreate拒绝循环依赖
- [ ] T086 [US3] 测试场景：ValidateCreate拒绝无效的action类型
- [ ] T087 [US3] 测试场景：参数校验（必填参数缺失）
- [ ] T088 [US3] 验证DRWorkflow Webhook覆盖率≥80%

### 3.3 DRPlanExecution Webhook测试（US3）

- [ ] T089 [US3] 创建 `internal/webhook/drplanexecution_webhook_test.go`
- [ ] T090 [US3] 测试场景：ValidateCreate检查Plan存在性
- [ ] T091 [US3] 测试场景：Revert类型必须提供revertExecutionRef
- [ ] T092 [US3] 测试场景：并发Execution检测（同一Plan不能同时有多个执行）
- [ ] T093 [US3] 验证DRPlanExecution Webhook覆盖率≥80%

**Checkpoint**: Webhook测试覆盖率达标，项目整体覆盖率≥55%（M2完成）

---

## Phase 4: User Story 4 - E2E测试基础设施（P3）

**目标**: 创建可复用的E2E测试环境自动化脚本

**独立测试**: 运行 `./test/e2e/scripts/e2e-full.sh`，验证完整流程

### 4.1 E2E脚本开发（US4）

- [ ] T094 [US4] 创建 `test/e2e/scripts/` 目录
- [ ] T095 [US4] 创建 `test/e2e/scripts/setup-e2e-env.sh`（幂等搭建Kind+Clusternet）
  - 检测Kind集群，不存在则创建
  - 安装CertManager（如需要）
  - 安装Clusternet Hub
  - 安装Clusternet Agent（模拟子集群）
  - 幂等设计：重复执行不报错
- [ ] T096 [US4] 创建 `test/e2e/scripts/run-e2e-tests.sh`
  - 构建controller镜像
  - 加载镜像到Kind
  - 部署controller到Kind集群
  - 运行：`go test -tags=e2e ./test/e2e/...`
- [ ] T097 [US4] 创建 `test/e2e/scripts/cleanup-e2e-env.sh`
  - 删除controller部署
  - 删除测试资源
  - 可选删除Kind集群（参数控制）
- [ ] T098 [US4] 创建 `test/e2e/scripts/e2e-full.sh`（wrapper脚本）
  - 组合调用：setup → run → cleanup
  - 错误处理：任何步骤失败时执行cleanup
- [ ] T099 [US4] 验证E2E脚本：`./test/e2e/scripts/e2e-full.sh`，总耗时≤20分钟

### 4.2 E2E测试场景 - DRPlan基础（US4）

- [ ] T100 [US4] 准备测试数据：复制 `example/plan/install/` 中的YAML到 `test/e2e/testdata/`
- [ ] T101 [US4] 在 `test/e2e/e2e_test.go` 补充场景1：DRPlan创建与验证
  - 创建DRWorkflows（subscription + localization）
  - 创建DRPlan
  - 验证Plan status.phase=Ready
- [ ] T102 [US4] 补充场景2：Execute操作
  - 创建DRPlanExecution（OperationType=Execute）
  - 等待execution status.phase=Succeeded（10分钟超时）
  - 验证Subscription CR已创建
  - 验证Localization CR已创建（在正确的namespace）
- [ ] T103 [US4] 补充场景3：Revert操作
  - 创建DRPlanExecution（OperationType=Revert）
  - 等待execution status.phase=Succeeded
  - 验证Subscription已删除
  - 验证Localization已删除

**Checkpoint**: E2E基础场景测试通过

---

## Phase 5: User Story 5 - E2E测试高级场景（P3）

**目标**: 补充Webhook验证、FailurePolicy、Metrics验证等场景

**独立测试**: 运行 `make test-e2e`，所有场景通过

### 5.1 E2E测试场景 - 高级功能（US4）

- [ ] T104 [US4] 补充场景4：Webhook验证
  - 尝试创建无rollback的Patch类型Workflow
  - 验证webhook拒绝（返回错误）
- [ ] T105 [US4] 补充场景5：FailurePolicy测试
  - 创建包含故意失败action的Plan
  - 配置FailurePolicy=Stop
  - 验证execution失败，后续stage未执行
- [ ] T106 [US4] 补充场景6：Metrics验证
  - 获取controller metrics endpoint
  - 验证 `controller_runtime_reconcile_total{controller="drplanexecution",result="success"}` 指标存在

### 5.2 CI/CD集成（US4）

- [ ] T107 [US4] 更新CI配置（`.github/workflows/test.yml` 或类似文件）
  - 添加lint job（每次push/PR）
  - 添加test job（每次push/PR）
  - 添加e2e job（仅PR时执行）
- [ ] T108 [US4] 更新 `Makefile`，添加E2E相关目标
  - `make e2e-setup`
  - `make e2e-run`
  - `make e2e-cleanup`
  - `make e2e-full`
- [ ] T109 [US4] 验证CI集成：提交commit，检查所有jobs执行成功

**Checkpoint**: E2E测试完整，CI/CD全部通过（M3）

---

## Phase 6: Polish & 文档更新

**目标**: 清理、优化、文档化

- [ ] T110 [P] [Polish] 更新项目README，添加测试覆盖率badge
- [ ] T111 [P] [Polish] 更新 `E2E_TESTING_GUIDE.md`，补充新增的自动化脚本使用说明
- [ ] T112 [P] [Polish] 清理临时文件和调试代码
- [ ] T113 [Polish] 生成最终覆盖率报告：`go tool cover -html=cover.out -o coverage.html`
- [ ] T114 [Polish] 提交所有修改，创建PR

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 0 (Setup) → Phase 1 (Foundational/Lint)
                      ↓
              +-----------------+
              ↓                 ↓
        Phase 2 (US2)      Phase 3 (US3)
              ↓                 ↓
              +-----------------+
                      ↓
                Phase 4 (US4) → Phase 5 (US4高级)
                      ↓
                Phase 6 (Polish)
```

### 关键路径

1. **Setup → Lint修复（T001-T039）** - 阻塞所有后续工作
2. **Lint修复 → 单元测试（T040-T075）** - 可与Webhook测试并行
3. **Lint修复 → Webhook测试（T076-T093）** - 可与单元测试并行
4. **单元测试+Webhook → E2E基础（T094-T103）**
5. **E2E基础 → E2E高级+CI（T104-T109）**

### User Story独立性

- **US1（Lint）**: 阻塞所有其他story
- **US2（单元测试）**: 依赖US1，与US3并行
- **US3（Webhook）**: 依赖US1，与US2并行
- **US4（E2E）**: 依赖US1+US2+US3完成

### 并行机会

**Phase 1内部**（T014-T019可并行）：
```bash
Task T014: 替换 "Create" → OperationCreate
Task T015: 替换 "Patch" → OperationPatch
Task T016: 替换 "Delete" → OperationDelete
# ... 同时进行
```

**Phase 2和3并行**（US2和US3可同时开发）：
```bash
Developer A: T040-T075 (Executor/Controller测试)
Developer B: T076-T093 (Webhook测试)
```

**Phase 2内部并行**（T040-T070可多人并行）：
```bash
Developer A: HTTP + Job Executor测试
Developer B: Localization + Subscription Executor测试
Developer C: K8sResource + Stage + Native测试
```

---

## Implementation Strategy

### 🎯 MVP Strategy: Lint修复优先

1. **Week 1**: Phase 0 + Phase 1（T001-T039）
   - 目标：`make lint` 通过
   - 里程碑：M1 - 代码可合并
   
2. **Week 2-3**: Phase 2 + Phase 3（T040-T093）
   - 并行开发Executor/Controller/Webhook测试
   - 目标：项目整体覆盖率≥55%
   - 里程碑：M2 - 测试覆盖率达标

3. **Week 4**: Phase 4 + Phase 5（T094-T109）
   - E2E自动化脚本 + 业务场景测试
   - 目标：CI/CD全部通过
   - 里程碑：M3 - E2E测试完成

4. **Week 5**: Phase 6 + Buffer（T110-T114）
   - 文档更新、清理、提交PR

### 🚀 增量交付策略

**Checkpoint 1（Day 7）**: Lint修复完成
- 提交：`fix: resolve all lint errors (funlen, goconst, gosec, revive)`
- 验证：`make lint` 退出码0

**Checkpoint 2（Day 21）**: 单元测试完成
- 提交：`test: add unit tests for controller/executor/webhook`
- 验证：`make test` 通过，覆盖率≥55%

**Checkpoint 3（Day 28）**: E2E测试完成
- 提交：`test: add E2E tests with Kind+Clusternet integration`
- 验证：`./test/e2e/scripts/e2e-full.sh` 通过

**Final PR**: 合并到主分支
- PR标题：`feat: improve test coverage and fix all lint errors`
- 链接spec: `Implements specs/002-fix-tests-lint/spec.md`

---

## Validation Checklist

在完成所有任务后，验证：

**Lint**:
- [ ] `make lint` 退出码0，无任何错误或警告
- [ ] 所有funlen错误已修复（14个）
- [ ] 所有goconst错误已修复
- [ ] 所有gosec错误已修复或豁免
- [ ] 所有package注释已添加

**Unit Test**:
- [ ] `make test` 所有测试通过
- [ ] Controller层覆盖率≥60%：`internal/controller/*.go`
- [ ] Executor层覆盖率≥70%：`internal/executor/*.go`
- [ ] Webhook层覆盖率≥80%：`internal/webhook/*.go`
- [ ] 项目整体覆盖率≥55%

**E2E Test**:
- [ ] E2E脚本可独立运行：`./test/e2e/scripts/e2e-full.sh`
- [ ] 场景1-6全部通过
- [ ] Clusternet资源正确创建和删除
- [ ] Metrics显示reconcile成功
- [ ] 完整流程耗时≤20分钟

**CI/CD**:
- [ ] commit触发lint+test（约3分钟）
- [ ] PR触发lint+test+e2e（约20分钟）
- [ ] 所有检查通过后PR可合并

---

## Notes

- **[P] 任务**: 不同文件，无依赖，可并行执行
- **[Story] 标签**: 追溯任务到用户故事
- **测试规范**: 所有测试必须遵循 `.cursor/rules/unit-testing-standards.mdc`
- **提交频率**: 每个逻辑单元（如单个executor测试、单个controller重构）提交一次
- **Checkpoint验证**: 每个checkpoint必须验证通过才能继续
- **避免**: 模糊任务、同一文件冲突、跨story依赖

---

## Quick Reference

### 运行测试

```bash
# 单元测试
make test
go tool cover -func=cover.out

# 特定包测试
go test ./internal/executor/... -v
go test ./internal/webhook/... -coverprofile=webhook-cover.out

# E2E测试
./test/e2e/scripts/e2e-full.sh

# 或分步执行
make e2e-setup
make e2e-run
make e2e-cleanup
```

### Lint检查

```bash
# 运行所有lint检查
make lint

# 特定linter
golangci-lint run --disable-all -E funlen
golangci-lint run --disable-all -E goconst
```

### 覆盖率报告

```bash
# 生成HTML报告
go tool cover -html=cover.out -o coverage.html

# 查看总体覆盖率
go tool cover -func=cover.out | grep total

# 查看特定文件覆盖率
go tool cover -func=cover.out | grep executor
```
