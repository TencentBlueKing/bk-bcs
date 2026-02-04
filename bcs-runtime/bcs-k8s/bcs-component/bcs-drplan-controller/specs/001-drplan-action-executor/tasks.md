# Tasks: 容灾策略 CR 及动作执行器

**Input**: Design documents from `/specs/001-drplan-action-executor/`
**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Tests**: 测试任务将在集成测试阶段统一进行，首版聚焦核心功能实现。

**Organization**: 任务按 User Story 组织，支持独立实现和测试。

**Logging**: 使用 k8s.io/klog/v2 进行日志输出（Info/Warning/Error/V(level) 分级）

**Recent Updates (2026-02-03)**:
- **Annotation 触发机制已移除**: T033（execute trigger）和 T039（revert trigger）标记为 DEPRECATED
- **revertExecutionRef 改为必填**: T040 更新为强制验证 revertExecutionRef
- **新增字段**: `executionHistory`（最近 10 条历史）和 `lastProcessedTrigger`（DEPRECATED）

## Format: `[ID] [P?] [Story] Description`

- **[P]**: 可并行执行（不同文件，无依赖）
- **[Story]**: 所属用户故事（US1-US13）

## 项目结构

```
bcs-drplan-controller/
├── api/v1alpha1/           # CRD 类型定义
├── internal/
│   ├── controller/         # Reconciler
│   ├── executor/           # 动作执行器
│   ├── webhook/            # Webhook 验证
│   └── utils/              # 工具函数
├── config/                 # K8s 配置
├── cmd/                    # 入口
└── tests/                  # 测试
```

---

## Phase 1: Setup (项目初始化)

**Purpose**: 使用 kubebuilder 初始化项目，创建基础结构

- [X] T001 使用 kubebuilder init 初始化项目 `kubebuilder init --domain bkbcs.tencent.com --repo github.com/Tencent/bk-bcs/bcs-runtime/bcs-k8s/bcs-component/bcs-drplan-controller`
- [X] T002 创建 DRWorkflow CRD 脚手架 `kubebuilder create api --group dr --version v1alpha1 --kind DRWorkflow --resource --controller`
- [X] T003 创建 DRPlan CRD 脚手架 `kubebuilder create api --group dr --version v1alpha1 --kind DRPlan --resource --controller`
- [X] T004 创建 DRPlanExecution CRD 脚手架 `kubebuilder create api --group dr --version v1alpha1 --kind DRPlanExecution --resource --controller`
- [X] T005 [P] 配置 .golangci.yml 遵循 BCS 代码规范
- [X] T006 [P] 配置 Makefile 添加 BCS 特定构建目标
- [X] T007 [P] 创建 Dockerfile 基于 BCS 基础镜像
- [X] T008 [P] 创建 .gitignore 文件（包含 Go 项目通用模式）

**Checkpoint**: ✅ 项目结构就绪，CRD 脚手架已创建

---

## Phase 2: Foundational (核心基础设施)

**Purpose**: 定义 CRD 类型和共享工具，所有 User Story 依赖此阶段

**⚠️ CRITICAL**: 必须完成后才能开始 User Story 实现

### CRD 类型定义

- [X] T009 [P] 定义 DRWorkflow types 在 api/v1alpha1/drworkflow_types.go
- [X] T010 [P] 定义 DRPlan types 在 api/v1alpha1/drplan_types.go（包含 Stage 编排字段）
- [X] T011 [P] 定义 DRPlanExecution types 在 api/v1alpha1/drplanexecution_types.go
- [X] T012 [P] 定义共享类型在 api/v1alpha1/common_types.go（HTTP/Job/Localization/Subscription/KubernetesResource 动作）
- [X] T013 运行 `make generate` 生成 DeepCopy 方法
- [X] T014 运行 `make manifests` 生成 CRD YAML

### 工具函数

- [X] T015 [P] 实现参数模板替换在 internal/utils/template.go（支持 `{{ .params.xxx }}` 和 `{{ .planName }}`）
- [X] T016 [P] 实现重试工具在 internal/utils/retry.go（支持指数退避）
- [X] T017 [P] 实现 klog 初始化在 cmd/main.go（配置日志级别、格式）

### 执行器接口

- [X] T018 定义执行器接口在 internal/executor/interface.go（WorkflowExecutor 和 ActionExecutor 接口）
- [X] T019 定义 Stage 执行器实现在 internal/executor/stage_executor.go（支持并行和依赖管理）

**Checkpoint**: ✅ CRD 类型定义完成，基础工具就绪，可以开始 User Story 实现

---

## Phase 3: MVP - User Story 1-4 基础工作流定义 (Priority: P1) 🎯

**Goal**: 用户可以创建和验证 DRWorkflow 和 DRPlan

**Independent Test**: `kubectl apply` 创建 CR，验证 status.phase 为 Ready

### US1: 定义工作流

- [X] T020 [US1] 实现 DRWorkflow Reconciler 在 internal/controller/drworkflow_controller.go（验证 actions、更新 status）
- [X] T021 [US1] 实现 DRWorkflow Webhook 在 internal/webhook/drworkflow_webhook.go（ValidatingWebhook + MutatingWebhook）

### US2: 参数化工作流

- [X] T022 [US2] 实现参数占位符验证在 internal/controller/drworkflow_controller.go（解析占位符、验证参数定义）

### US3: 步骤回滚定义

- [X] T023 [US3] 实现 rollback 校验逻辑在 internal/webhook/drworkflow_webhook.go（Localization/Subscription/KubernetesResource Patch 必须定义 rollback）

### US4: DRPlan 定义（单 Workflow）

- [X] T024 [US4] 实现 DRPlan Reconciler 在 internal/controller/drplan_controller.go（验证 workflowRef、参数、更新 status）
- [X] T025 [US4] 实现 DRPlan Webhook 在 internal/webhook/drplan_webhook.go（验证参数值、添加 labels）
- [X] T026 [US4] 实现工作流引用保护在 internal/controller/drworkflow_controller.go（Finalizer 阻止被引用时删除）

**Checkpoint**: ✅ 可以创建和验证基础 DRWorkflow 和 DRPlan（已包含 Stage 编排）

---

## Phase 4: User Story 4a - Stage 编排 (Priority: P1)

**Goal**: 支持多 Workflow Stage 编排，实现复杂系统容灾切换

**Independent Test**: 创建包含多个 Stage 的 DRPlan，验证依赖和并行配置有效

### US4a: Stage 编排

- [X] T027 [US4a] 扩展 DRPlan Reconciler 支持 Stage 验证在 internal/controller/drplan_controller.go（验证 stages、dependsOn、循环依赖检测）
- [X] T028 [US4a] 实现 Stage 参数合并逻辑在 internal/utils/params.go（globalParams + Stage params 优先级处理）
- [X] T029 [US4a] 实现 Stage 依赖图构建在 internal/executor/stage_executor.go（拓扑排序、依赖关系图）
- [X] T030 [US4a] 实现 Stage 执行引擎在 internal/executor/stage_executor.go（支持 parallel、dependsOn、FailFast）

**Checkpoint**: ✅ 支持 Stage 编排和并行执行（已在 T019 stage_executor.go 中实现）

---

## Phase 5: User Story 10-11-13 - 执行与恢复 (Priority: P1)

**Goal**: 支持手动触发执行、恢复和取消操作

**Independent Test**: 创建 DRPlanExecution 或添加 annotation 触发执行，验证状态更新

### US10: 手动触发执行

- [X] T031 [US10] 实现 DRPlanExecution Reconciler 在 internal/controller/drplanexecution_controller.go（验证 planRef、解析参数、更新状态）
- [X] T032 [US10] 实现 Native 执行引擎在 internal/executor/native_executor.go（顺序执行 actions、failurePolicy、发送 Events）
- [X] T033 [US10] **[DEPRECATED 2026-02-03]** ~~实现 annotation 触发在 internal/controller/drplan_controller.go（Watch `dr.bkbcs.tencent.com/trigger=execute`）~~ - 已移除，改为仅支持 DRPlanExecution CR 触发
- [X] T034 [US10] 实现并发控制在 internal/controller/drplan_controller.go（检查 currentExecution、拒绝并发）
- [X] T035 [US10] 实现 Kubernetes Events 发送在 internal/executor/events.go（ExecutionStarted、ActionSucceeded 等 11 种事件）
- [X] T036 [US10] 集成 klog 日志在 internal/executor/native_executor.go（Info 级别：关键事件，V(4) 级别：详细信息）

### US11: 手动触发恢复

- [X] T037 [US11] 实现 Revert 执行逻辑在 internal/executor/native_executor.go（逆序遍历、执行 rollback，从 `revertExecutionRef` 指定的目标获取 StageStatuses）
- [X] T038 [US11] 实现回滚决策逻辑在 internal/executor/native_executor.go（GetRollbackAction、自动/自定义 rollback）- 注：rollback.go 已合并到 native_executor.go
- [X] T039 [US11] **[DEPRECATED 2026-02-03]** ~~实现 revert annotation 触发在 internal/controller/drplan_controller.go（Watch `dr.bkbcs.tencent.com/trigger=revert`）~~ - 已移除，改为要求在 DRPlanExecution 中显式指定 `revertExecutionRef`
- [X] T040 [US11] 实现 Revert Webhook 验证在 internal/webhook/drplanexecution_webhook.go（验证 `revertExecutionRef` 必填、目标存在、类型为 Execute、状态为 Succeeded）

### US13: 取消执行

- [X] T041 [US13] 实现 cancel annotation 处理在 internal/controller/drplanexecution_controller.go（Watch `dr.bkbcs.tencent.com/cancel=true`）
- [X] T042 [US13] 实现取消逻辑在 internal/executor/native_executor.go（停止后续 action、标记 Skipped）
- [X] T043 [US13] 实现 DRPlan cancel 触发在 internal/controller/drplan_controller.go（Watch `dr.bkbcs.tencent.com/trigger=cancel`）

**Checkpoint**: ✅ 核心执行、恢复、取消功能就绪，可以运行基本容灾流程

---

## Phase 6: User Story 5 - HTTP 执行器 (Priority: P2) ✅ COMPLETED

**Goal**: 支持执行 HTTP 类型动作

**Independent Test**: 创建包含 HTTP 动作的工作流，触发执行验证 HTTP 请求发送

### US5: HTTP 执行器

- [X] T044 [P] [US5] 实现 HTTP 执行器在 internal/executor/http_executor.go（支持 GET/POST/PUT/DELETE、headers、body、successCodes）
- [X] T045 [US5] 实现 HTTP 超时和重试在 internal/executor/http_executor.go（context.WithTimeout、RetryWithBackoff）
- [X] T046 [US5] 集成 klog 日志在 internal/executor/http_executor.go（Info: 请求/响应状态，V(4): 请求体/响应体）

**Checkpoint**: HTTP 动作可正常执行和回滚

---

## Phase 7: User Story 6 - Job 执行器 (Priority: P2)

**Goal**: 支持执行 Job 类型动作

**Independent Test**: 创建包含 Job 动作的工作流，触发执行验证 Job 创建和完成

### US6: Job 执行器

- [X] T047 [P] [US6] 实现 Job 执行器在 internal/executor/job_executor.go（创建 Job、Watch status、记录 jobRef）
- [X] T048 [US6] 实现 Job 超时处理在 internal/executor/job_executor.go（超时删除 Job）
- [X] T049 [US6] 实现 Job 自动回滚在 internal/executor/rollback.go（DeleteJob 函数）
- [X] T050 [US6] 集成 klog 日志在 internal/executor/job_executor.go（Info: Job 状态变化，V(4): Job manifest）

**Checkpoint**: ✅ Job 动作可正常执行和回滚

---

## Phase 8: User Story 9 - KubernetesResource 通用执行器 (Priority: P2) ✅ COMPLETED

**Goal**: 支持操作任意 Kubernetes 资源（ConfigMap、Secret、CRD 等）

**Independent Test**: 创建包含 KubernetesResource 动作的工作流，验证资源创建/更新/删除

### US9: KubernetesResource 执行器

- [X] T051 [P] [US9] 实现 KubernetesResource 执行器在 internal/executor/k8s_resource_executor.go（支持 Create/Apply/Patch/Delete）
- [X] T052 [US9] 实现 manifest 解析在 internal/executor/k8s_resource_executor.go（YAML 解析、参数替换、dynamic client）
- [X] T053 [US9] 实现 KubernetesResource 自动回滚在 internal/executor/rollback.go（DeleteResource 函数）
- [X] T054 [US9] 集成 klog 日志在 internal/executor/k8s_resource_executor.go（Info: 资源操作状态，V(4): manifest 内容）

**Checkpoint**: ✅ 通用 K8s 资源操作就绪，支持 CRD 扩展

---

## Phase 9: User Story 7-8 - Clusternet 执行器 (Priority: P3) ✅ COMPLETED

**Goal**: 支持执行 Localization 和 Subscription 类型动作

**Independent Test**: 创建包含 Clusternet 动作的工作流，触发执行验证 CR 创建（异步模型）

### US7: Localization 执行器

- [X] T055 [P] [US7] 实现 Localization 执行器在 internal/executor/localization_executor.go（支持 Create/Patch/Delete、参数替换、记录 localizationRef）
- [X] T056 [US7] 实现 Localization 自动回滚在 internal/executor/rollback.go（DeleteLocalization 函数）
- [X] T057 [US7] 集成 klog 日志在 internal/executor/localization_executor.go（Info: 操作状态，V(4): Localization 配置）

### US8: Subscription 执行器

- [X] T058 [P] [US8] 实现 Subscription 执行器在 internal/executor/subscription_executor.go（支持 Create/Patch/Delete、配置 feeds/subscribers）
- [X] T059 [US8] 实现 Subscription 自动回滚在 internal/executor/rollback.go（DeleteSubscription 函数）
- [X] T060 [US8] 集成 klog 日志在 internal/executor/subscription_executor.go（Info: 操作状态，V(4): Subscription 配置）

**Checkpoint**: ✅ Clusternet 集成完成，支持多集群资源下发

---

## Phase 10: User Story 12 - 执行历史 (Priority: P2)

**Goal**: 支持查看执行历史和审计

**Independent Test**: 通过 label selector 查询 DRPlan 的所有执行记录

### US12: 执行历史

- [X] T061 [P] [US12] 实现执行记录 label 管理在 internal/controller/drplanexecution_controller.go（添加 `drplan=<name>` label）
- [X] T062 [US12] 实现执行历史查询在 docs/user-guide.md（补充 kubectl 查询示例）

**Checkpoint**: ✅ 执行历史可追溯

---

## Phase 11: Polish & Cross-Cutting Concerns

**Purpose**: 完善配置、文档、部署和性能优化

- [X] T063 [P] 实现 Reconcile 频率配置在 cmd/main.go（支持环境变量 `RECONCILE_INTERVAL`，默认 30s）
- [X] T064 [P] 实现网络分区检测在 internal/controller/drplanexecution_controller.go（超过 2 分钟标记 Unknown 状态）
- [X] T065 [P] 配置 RBAC 在 config/rbac/（ClusterRole、ServiceAccount、RoleBinding）
- [X] T066 [P] 配置 Webhook 证书在 config/webhook/（cert-manager 集成）
- [X] T067 [P] 更新 README.md（项目简介、快速开始、架构图）
- [X] T068 [P] 创建部署示例在 config/samples/（DRWorkflow、DRPlan、DRPlanExecution 示例）
- [X] T069 [P] 配置 Prometheus metrics（可选，后续版本）

**Checkpoint**: ✅ 项目就绪，可部署和使用

---

## Dependencies & Execution Order

### 依赖关系图

```
Phase 1 (Setup)
    ↓
Phase 2 (Foundational) ← BLOCKING for all User Stories
    ↓
Phase 3 (US1-4) ← MVP Core
    ↓
Phase 4 (US4a) ← Stage Orchestration
    ↓
Phase 5 (US10-11-13) ← Execution Engine ← BLOCKING for Actions
    ↓
    ├── Phase 6 (US5) HTTP Executor
    ├── Phase 7 (US6) Job Executor
    ├── Phase 8 (US9) K8s Resource Executor
    └── Phase 9 (US7-8) Clusternet Executors
    ↓
Phase 10 (US12) Execution History
    ↓
Phase 11 (Polish)
```

### 并行执行机会

**Phase 2 Foundational**（可并行）:
- T009, T010, T011, T012 (CRD types)
- T015, T016, T017 (Utils)

**Phase 3 MVP**（部分并行）:
- T020, T021 可并行（US1）
- T024, T025 可并行（US4）

**Phase 6-9 Executors**（完全并行）:
- T044-046 (HTTP), T047-050 (Job), T051-054 (K8s), T055-060 (Clusternet) 可并行实现

**Phase 11 Polish**（完全并行）:
- T063-069 全部可并行

---

## Implementation Strategy

### MVP Scope (最小可行产品)

**目标**: 支持单 Workflow 的定义、执行和回滚

**包含 Phase**:
- Phase 1: Setup
- Phase 2: Foundational
- Phase 3: US1-4 (工作流和预案定义)
- Phase 5: US10-11-13 (执行引擎)
- Phase 6: US5 (HTTP 执行器) - 最基础的动作类型

**验证标准**:
1. 可以创建 DRWorkflow（包含 HTTP 动作）
2. 可以创建 DRPlan 并传递参数
3. 可以触发执行并查看状态
4. 可以触发回滚
5. 可以取消执行

### Incremental Delivery

1. **MVP** (Phase 1-6): 单 Workflow + HTTP 动作
2. **V0.2** (Phase 7-8): 添加 Job 和 K8s Resource 执行器
3. **V0.3** (Phase 4): 添加 Stage 编排
4. **V0.4** (Phase 9): 添加 Clusternet 集成
5. **V1.0** (Phase 10-11): 完善历史记录、性能优化

---

## Task Summary

- **Total Tasks**: 69
- **Setup Phase**: 8 tasks
- **Foundational Phase**: 11 tasks
- **User Story Tasks**: 45 tasks
  - US1 (P1): 2 tasks
  - US2 (P1): 1 task
  - US3 (P1): 1 task
  - US4 (P1): 3 tasks
  - US4a (P1): 4 tasks
  - US5 (P2): 3 tasks
  - US6 (P2): 4 tasks
  - US7 (P3): 3 tasks
  - US8 (P3): 3 tasks
  - US9 (P2): 4 tasks
  - US10 (P1): 6 tasks
  - US11 (P1): 4 tasks
  - US12 (P2): 2 tasks
  - US13 (P1): 3 tasks
- **Polish Phase**: 7 tasks
- **Parallel Opportunities**: 35+ tasks 可并行执行

**MVP Task Count**: ~30 tasks (Phase 1-6)

---

## Notes

- 所有任务遵循严格的清单格式
- klog 用于所有日志输出（替代 logr）
- Stage 编排采用 FailFast 策略
- Clusternet 动作采用异步模型（CR 创建成功即完成）
- Reconcile 频率可配置，默认 30 秒
- 网络分区超过 2 分钟标记为 Unknown 状态
