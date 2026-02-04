# Feature Specification: 容灾策略 CR 及动作执行器

**Feature Branch**: `001-drplan-action-executor`  
**Created**: 2026-01-30  
**Updated**: 2026-02-02  
**Status**: Draft  
**Input**: 定义容灾策略 CR，支持关联多种动作流程（Clusternet Localization/Subscription 下发、HTTP 接口执行、Job 资源创建、通用 K8s 资源操作）

## Clarifications

### Session 2026-01-30

- Q: 执行历史清理策略？ → A: 不自动清理，用户手动管理
- Q: HTTP 动作中敏感数据如何处理？ → A: 首版不处理，后续迭代支持 Secret 引用
- Q: 系统预期规模？ → A: 单集群最多 500 DRPlan
- Q: Localization 操作类型？ → A: `operation: Create|Patch|Delete`。Create 需要 `feed`（源资源）和 `overrides`（覆盖规则），Patch 必须显式定义 rollback（无法自动推断逆操作）
- Q: Localization namespace 含义？ → A: `namespace` 字段指 Localization CR 的命名空间，即 ManagedCluster 的命名空间（如 clusternet-xxxxx）
- Q: Localization overrides 结构？ → A: 每个 override 包含 `name`（标识）、`type`（JSONPatch/MergePatch/Helm）、`value`（patch 内容）
- Q: Subscription 支持？ → A: 支持 `type: Subscription`，与 Localization 规则相同（Create 自动回滚删除，Patch 强制定义 rollback）
- Q: 执行引擎扩展性？ → A: 首版内置引擎（顺序执行），预留 `executor` 字段支持 Argo Workflows 等外部引擎扩展
- Q: 参数替换错误处理？ → A: 未定义参数 → 返回错误阻止执行；空值参数 → 替换为空字符串；模板语法错误 → 返回错误阻止执行

### Session 2026-02-02

- Q: DRPlanExecution Controller 的 reconcile 循环以什么频率检查执行状态？ → A: 可配置，默认 30 秒
- Q: 当 DRPlanExecution 正在执行时，如果 Controller 与 API Server 之间出现网络分区超过 2 分钟，系统应该如何处理？ → A: 标记执行为 Unknown 状态，需要人工介入检查
- Q: 当一个 Stage 配置为 `parallel: true` 并包含多个 Workflow 时，如果其中一个 Workflow 执行失败，系统应该如何处理其他正在并行执行的 Workflow？ → A: 立即停止其他 Workflow（不再启动新的），等待当前运行的完成，Stage 标记为 Failed（FailFast 策略）
- Q: Controller 应该如何记录执行日志以便排查问题？ → A: Info 级别记录关键事件（执行开始/成功/失败、Stage 切换），Debug 级别记录详细参数、动作内容和 HTTP 请求体，使用结构化 JSON 日志格式
- Q: 当 DRWorkflow 创建 Localization CR 后，系统应该如何确认 Clusternet 已完成下发？ → A: 创建 CR 后立即标记动作成功，不等待同步（异步模型，如需确认同步完成可在后续步骤添加健康检查动作）

### Session 2026-02-03

- Q: 为什么移除 annotation 触发机制（`dr.bkbcs.tencent.com/trigger`）？ → A: (1) Revert 操作无法通过 annotation 精确指定要回滚哪个 execution，只能回滚 `LastExecutionRef`，在多次执行后易混淆；(2) Annotation 是可变的，不符合声明式 GitOps 原则，难以追踪和审计；(3) 需要幂等性跟踪（`lastProcessedTrigger`）增加复杂度
- Q: `revertExecutionRef` 为什么从可选改为必填？ → A: 强制要求显式指定回滚目标，防止自动选择错误的 execution（如 `LastExecutionRef` 被新执行覆盖），增强操作可预测性和审计性。Webhook 会验证目标 execution 存在、类型为 Execute、状态为 Succeeded
- Q: `executionHistory` 字段的用途是什么？ → A: 在 `DRPlan.Status` 中保留最近 10 条执行历史（新到旧），包含 name、namespace、operationType、phase、时间等信息，用于快速查看执行趋势和审计，无需查询所有 DRPlanExecution CR。即使 execution CR 被删除，历史记录也会保留（通过 finalizer 机制确保）
- Q: `lastProcessedTrigger` 字段为什么标记为 DEPRECATED？ → A: 此字段用于 annotation 触发的幂等性跟踪，随着 annotation 触发机制移除已无实际用途，保留仅为向后兼容，未来版本将删除
- Q: `lastExecutionRef` 的语义是什么？ → A: **始终指向最后一个成功的操作**（不论是 Execute 还是 Revert）。通过 `executionHistory[0].operationType` 可以区分操作类型。这确保了完整的操作时间线，不会因为 Revert 操作而丢失最后操作的引用

## 架构概述

系统采用三层 CRD 设计，实现关注点分离：

| CRD | 职责 | 说明 |
|-----|------|------|
| **DRPlan** | 预案定义和编排 | 定义"是什么"：预案名称、Stage 编排（多 Workflow）、全局参数、触发配置 |
| **DRWorkflow** | 工作流定义 | 定义"怎么做"：单个组件的动作列表、执行顺序、失败策略（可复用） |
| **DRPlanExecution** | 执行记录 | 记录"执行情况"：实时状态、Stage 进度、Workflow 状态、动作进度、结果 |

### Stage 编排

**设计理念**：
- **DRPlan 层**：负责顶层编排，将多个 Workflow 组织为 Stage，支持 Stage 间依赖和并行
- **DRWorkflow 层**：保持单一职责，每个 Workflow 只负责一个组件的切换（如 MySQL、Redis、某个服务）
- **复用性**：一个 Workflow 可被多个 Plan 和多个 Stage 引用

**典型场景示例**：

```
电商系统容灾切换:
├── Stage 1: notify-start (顺序) 
│   └── notification-workflow
├── Stage 2: infrastructure (并行，依赖 Stage 1)
│   ├── mysql-failover
│   ├── redis-failover  
│   └── kafka-failover
├── Stage 3: applications (并行，依赖 Stage 2)
│   ├── order-service-failover
│   ├── payment-service-failover
│   ├── user-service-failover
│   └── frontend-failover
├── Stage 4: validation (顺序，依赖 Stage 3)
│   ├── smoke-test-workflow
│   └── health-check-workflow
└── Stage 5: notify-complete (顺序，依赖 Stage 4)
    └── notification-workflow
```

### 触发机制

| 触发方式 | 说明 | 状态 |
|---------|------|------|
| **Manual** | 手动创建 DRPlanExecution CR 触发执行 | ✅ 支持 |
| ~~Annotation~~ | ~~通过 annotation 标记 DRPlan 触发执行~~ | ❌ 已移除（2026-02-03） |
| Event | 监听其他 CR 状态变化自动触发 | 🔜 规划中 |
| Schedule | 定时调度触发（Cron） | 🔜 规划中 |

**触发方式**：

```bash
# Execute: 创建 DRPlanExecution CR 触发执行
kubectl apply -f drplanexecution-execute.yaml

# Revert: 创建回滚 execution（revertExecutionRef 必填）
kubectl apply -f drplanexecution-revert.yaml
```

**注意**：
- ⚠️ **Annotation 触发已移除**（2026-02-03）：原 `dr.bkbcs.tencent.com/trigger=execute|revert|cancel` 机制已完全移除
  - **移除原因**: 无法精确指定 revert 目标、不符合 GitOps 声明式原则、需要额外的幂等性跟踪
  - **替代方案**: 直接创建 DRPlanExecution CR（更明确、可追溯、易审计）
- ✅ **Webhook 验证**: 创建时自动验证所有必填字段和引用的有效性（见 4.1.1.1 节）
- ✅ **GitOps 友好**: 声明式 CR 支持版本控制、代码审查、回滚历史

### 执行引擎（Executor）

系统采用可插拔的执行引擎设计，支持后续扩展：

| 引擎 | 说明 | 状态 |
|------|------|------|
| **Native** | 内置引擎，顺序执行 | ✅ 首版实现 |
| Argo | 转换为 Argo Workflow 执行 | 🔜 扩展预留 |

**首版（Native 引擎）**：
- 顺序执行 actions 列表
- 内置重试、超时、回滚逻辑
- 轻量无外部依赖

**扩展版（Argo 引擎）**：
- DRWorkflow → Argo Workflow 自动转换
- Execute/Revert 分别生成独立的 Argo Workflow
- 利用 Argo 的 DAG、并行、UI 能力
- 回滚逻辑由 Controller 生成为 Argo Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│                      DRWorkflow (统一抽象)                       │
└───────────────────────────┬─────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
┌─────────────────────┐           ┌─────────────────────┐
│   Native Executor   │           │    Argo Executor    │
│   (首版实现)         │           │    (扩展预留)        │
│                     │           │                     │
│   Controller 直接   │           │   生成 Argo WF      │
│   执行动作          │           │   Watch 状态同步     │
└─────────────────────┘           └─────────────────────┘
```

**渐进式演进路线**：

| Phase | 功能 | 引擎 |
|-------|------|------|
| **Phase 1** | 顺序执行、回滚、参数化 | Native |
| **Phase 2** | DAG（dependsOn）、条件执行（when） | Native 扩展 |
| **Phase 3** | Argo 集成、可视化 UI | Argo 适配器 |

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 定义工作流 (Priority: P1)

作为平台运维人员，我需要通过 DRWorkflow CR 定义一个可复用的动作流程，描述一系列按顺序执行的动作。

**Why this priority**: 工作流是容灾执行的基础单元，必须先能定义工作流才能组装预案。

**Independent Test**: 可以通过 `kubectl apply` 创建 DRWorkflow CR，并通过 `kubectl get drworkflow` 验证工作流已正确存储。

**Acceptance Scenarios**:

1. **Given** 集群中未部署任何工作流，**When** 运维人员通过 kubectl 创建一个 DRWorkflow CR（包含动作列表、失败策略），**Then** 系统接受该 CR 并将状态设置为 Ready
2. **Given** 已存在一个 DRWorkflow CR，**When** 运维人员修改动作列表，**Then** 系统更新工作流并重新验证配置
3. **Given** 创建的 DRWorkflow 包含无效配置，**When** 提交该 CR，**Then** Webhook 拒绝创建并返回明确的错误信息

---

### User Story 2 - 定义参数化工作流 (Priority: P1)

作为平台运维人员，我需要在 DRWorkflow 中使用参数占位符，使工作流更加通用和可复用。

**Why this priority**: 参数化是工作流复用的基础，同一个工作流可以通过不同参数应用于不同场景。

**Independent Test**: 可以创建包含参数占位符的 DRWorkflow，验证参数定义被正确存储。

**Acceptance Scenarios**:

1. **Given** 运维人员创建 DRWorkflow 时，**When** 在动作配置中使用 `{{ .params.targetCluster }}` 等占位符，并在 `parameters` 字段定义参数名和默认值，**Then** 系统接受配置并将工作流状态设置为 Ready
2. **Given** DRWorkflow 定义了必填参数（无默认值），**When** 提交该 CR，**Then** 系统标记该参数为必填，在 Plan 引用时必须提供
3. **Given** DRWorkflow 使用了未定义的参数占位符，**When** 提交该 CR，**Then** Webhook 拒绝并提示未定义的参数

---

### User Story 3 - 定义步骤回滚动作 (Priority: P1)

作为平台运维人员，我需要能够为工作流中的每个步骤定义回滚动作，以便在执行 revert 时能够正确恢复状态。

**Why this priority**: 步骤级别的回滚定义是自动恢复的基础，确保每个操作都有对应的逆操作。

**Independent Test**: 可以创建包含 rollback 定义的步骤，触发 revert 时验证回滚动作被正确执行。

**Acceptance Scenarios**:

1. **Given** 步骤定义了自定义 rollback 动作，**When** 触发 revert，**Then** 系统执行自定义的 rollback 动作
2. **Given** 步骤未定义 rollback 动作（Localization 类型，operation=Create），**When** 触发 revert，**Then** 系统自动执行逆操作（删除 Localization）
3. **Given** 步骤未定义 rollback 动作（Job 类型），**When** 触发 revert，**Then** 系统自动执行逆操作（删除 Job）
4. **Given** 步骤未定义 rollback 动作（HTTP 类型），**When** 触发 revert，**Then** 系统跳过该步骤的回滚（HTTP 无默认逆操作）
5. **Given** 执行成功的步骤列表为 [A, B, C]，**When** 触发 revert，**Then** 系统按逆序执行回滚：C → B → A

---

### User Story 4 - 定义预案并传递参数 (Priority: P1)

作为平台运维人员，我需要通过 DRPlan CR 定义一个预案，关联一个工作流并传递具体的参数值。

**Why this priority**: 预案是运维人员操作的入口，参数传递让同一工作流可以用于不同场景（首次部署、容灾切换等）。

**Independent Test**: 可以创建 DRPlan 并传递参数，验证参数值覆盖工作流的默认值。

**Acceptance Scenarios**:

1. **Given** 已存在一个定义了参数的 DRWorkflow，**When** 运维人员创建 DRPlan 并在 `params` 中提供参数值，**Then** 系统验证参数有效并将预案状态设置为 Ready
2. **Given** DRWorkflow 定义了必填参数，**When** DRPlan 未提供该参数值，**Then** Webhook 拒绝并提示缺少必填参数
3. **Given** DRPlan 提供了工作流未定义的参数，**When** 提交该 CR，**Then** 系统忽略多余参数（或可配置为警告）
4. **Given** DRPlan 关联的 DRWorkflow 被删除，**When** 系统检测到引用失效，**Then** DRPlan 状态变为 Invalid 并记录原因

---

### User Story 4a - 定义多 Workflow Stage 编排预案 (Priority: P1)

作为平台运维人员，我需要通过 DRPlan 的 Stage 编排多个 Workflow，实现复杂系统的容灾切换。

**Why this priority**: 实际生产环境中，一个系统通常包含多个组件（数据库、缓存、应用服务等），需要分阶段、分组并行执行。

**Independent Test**: 可以创建包含多个 Stage 的 DRPlan，每个 Stage 引用不同的 Workflow，验证执行时按 Stage 顺序和依赖关系执行。

**Acceptance Scenarios**:

1. **Given** 已存在多个 DRWorkflow（mysql-failover、redis-failover等），**When** 运维人员创建 DRPlan 并在 `stages` 中引用这些 Workflow，**Then** 系统验证所有引用的 Workflow 存在并将预案状态设置为 Ready
2. **Given** DRPlan 定义了多个 Stage，**When** Stage 之间定义了 `dependsOn` 依赖关系，**Then** 系统验证依赖关系不形成循环依赖
3. **Given** Stage 定义了 `parallel: true`，**When** 执行该 Stage，**Then** 系统并行执行该 Stage 内的所有 Workflow
4. **Given** Stage 定义了 `parallel: false`，**When** 执行该 Stage，**Then** 系统顺序执行该 Stage 内的所有 Workflow
5. **Given** Stage 定义了 `dependsOn: [stage1]`，**When** 执行该 Stage，**Then** 系统等待 stage1 完成后才开始执行
6. **Given** DRPlan 定义了 `globalParams`，**When** Stage 中的 Workflow 执行时，**Then** globalParams 自动传递给所有 Workflow（Stage 级 params 优先级更高）
7. **Given** Stage 配置为 `parallel: true` 并包含多个 Workflow，**When** 其中一个 Workflow 执行失败，**Then** 系统立即停止启动新的 Workflow，等待当前运行的 Workflow 完成，Stage 标记为 Failed

---

### User Story 5 - 执行 HTTP 动作 (Priority: P2)

作为平台运维人员，我需要在工作流中定义 HTTP 动作，当工作流被执行时能够调用外部 API 接口。

**Why this priority**: HTTP 动作是最通用的集成方式，可以对接任意外部系统。

**Independent Test**: 可以创建包含 HTTP 动作的 DRWorkflow，通过预案触发执行，验证目标 HTTP 接口被正确调用。

**Acceptance Scenarios**:

1. **Given** 一个包含 HTTP 动作的工作流（指定 URL、Method、Headers、Body），**When** 工作流被执行，**Then** 系统按配置发送 HTTP 请求，并在执行记录中记录响应状态码和耗时
2. **Given** HTTP 动作配置了重试策略（3 次重试，间隔 5 秒），**When** 目标接口返回 5xx 错误，**Then** 系统自动重试直到成功或达到重试上限
3. **Given** HTTP 动作配置了超时时间（30 秒），**When** 目标接口响应超时，**Then** 系统终止该请求并标记动作失败

---

### User Story 6 - 执行 Job 动作 (Priority: P2)

作为平台运维人员，我需要在工作流中定义 Job 动作，当工作流被执行时能够在集群中创建 Kubernetes Job。

**Why this priority**: Job 动作提供了在集群内执行复杂恢复逻辑的能力。

**Independent Test**: 可以创建包含 Job 动作的 DRWorkflow，通过预案触发执行，验证 Job 被正确创建并运行完成。

**Acceptance Scenarios**:

1. **Given** 一个包含 Job 动作的工作流（指定 Job 模板、命名空间），**When** 工作流被执行，**Then** 系统在指定命名空间创建 Job
2. **Given** Job 动作正在执行，**When** Job 运行成功，**Then** 执行记录中该动作状态标记为成功
3. **Given** Job 动作配置了超时时间（10 分钟），**When** Job 运行超时，**Then** 系统尝试清理 Job 资源，标记动作失败

---

### User Story 7 - 执行 Clusternet Localization 动作 (Priority: P3)

作为平台运维人员，我需要在工作流中定义 Clusternet Localization 动作，当工作流被执行时能够通过 Clusternet 将配置下发到目标集群。

**Why this priority**: Clusternet 集成是 BCS 生态的重要能力，但依赖外部组件。

**Independent Test**: 可以创建包含 Localization 动作的 DRWorkflow，通过预案触发执行，验证 Localization CR 被正确创建。

**Acceptance Scenarios**:

1. **Given** 一个包含 Localization 动作的工作流（operation=Create），**When** 工作流被执行，**Then** 系统在 Hub 集群创建 Localization CR，动作立即标记为成功（异步模型，不等待 Clusternet 同步完成）
2. **Given** 一个包含 Localization 动作的工作流（operation=Patch），**When** 工作流被执行，**Then** 系统在 Hub 集群更新指定的 Localization CR，动作立即标记为成功
3. **Given** 一个包含 Localization 动作的工作流（operation=Delete），**When** 工作流被执行，**Then** 系统在 Hub 集群删除指定的 Localization CR，动作立即标记为成功
4. **Given** Localization CR 创建失败（API Server 拒绝），**When** 检测到创建失败，**Then** 系统根据重试策略决定是否重试
5. **Given** 需要确保 Clusternet 同步完成，**When** 定义工作流时，**Then** 可在 Localization 动作后添加 HTTP 健康检查动作或 Job 动作验证目标集群状态
6. **Given** Localization 动作 operation=Patch 且未定义 rollback，**When** 保存 DRWorkflow，**Then** 校验失败并提示必须定义 rollback

---

### User Story 8 - 执行 Clusternet Subscription 动作 (Priority: P3)

作为平台运维人员，我需要在工作流中定义 Clusternet Subscription 动作，当工作流被执行时能够通过 Clusternet 将资源分发到多个目标集群。

**Why this priority**: Subscription 是 Clusternet 多集群分发的核心能力，与 Localization 互补。

**Independent Test**: 可以创建包含 Subscription 动作的 DRWorkflow，通过预案触发执行，验证 Subscription CR 被正确创建。

**Acceptance Scenarios**:

1. **Given** 一个包含 Subscription 动作的工作流（operation=Create），**When** 工作流被执行，**Then** 系统在 Hub 集群创建 Subscription CR，动作立即标记为成功（异步模型，不等待 Clusternet 分发完成）
2. **Given** 一个包含 Subscription 动作的工作流（operation=Patch），**When** 工作流被执行，**Then** 系统在 Hub 集群更新指定的 Subscription CR，动作立即标记为成功
3. **Given** 一个包含 Subscription 动作的工作流（operation=Delete），**When** 工作流被执行，**Then** 系统在 Hub 集群删除指定的 Subscription CR，动作立即标记为成功
4. **Given** Subscription CR 创建失败（API Server 拒绝），**When** 检测到创建失败，**Then** 系统根据重试策略决定是否重试
5. **Given** 需要确保 Clusternet 分发完成，**When** 定义工作流时，**Then** 可在 Subscription 动作后添加 HTTP 健康检查动作或 Job 动作验证目标集群资源状态
6. **Given** Subscription 动作 operation=Patch 且未定义 rollback，**When** 保存 DRWorkflow，**Then** 校验失败并提示必须定义 rollback

---

### User Story 9 - 执行通用 Kubernetes 资源动作 (Priority: P2)

作为平台运维人员，我需要在工作流中定义通用的 Kubernetes 资源操作，以便操作任意类型的 K8s 资源（包括自定义 CRD）。

**Why this priority**: 提供灵活的扩展能力，支持操作 ConfigMap、Secret、自定义 CRD 等专用类型未覆盖的资源。

**Independent Test**: 可以创建包含 KubernetesResource 动作的 DRWorkflow，通过预案触发执行，验证任意 K8s 资源被正确创建/更新/删除。

**Acceptance Scenarios**:

1. **Given** 一个包含 KubernetesResource 动作的工作流（operation=Create），**When** 工作流被执行，**Then** 系统根据 manifest 创建指定的 K8s 资源
2. **Given** 一个包含 KubernetesResource 动作的工作流（operation=Apply），**When** 工作流被执行，**Then** 系统执行 server-side apply 操作
3. **Given** 一个包含 KubernetesResource 动作的工作流（operation=Delete），**When** 工作流被执行，**Then** 系统删除指定的 K8s 资源
4. **Given** KubernetesResource 动作的 manifest 中包含参数占位符，**When** 工作流被执行，**Then** 系统正确替换占位符后再创建资源
5. **Given** KubernetesResource 动作 operation=Create 且未定义 rollback，**When** 触发回滚，**Then** 系统自动删除创建的资源
6. **Given** KubernetesResource 动作 operation=Patch 且未定义 rollback，**When** 保存 DRWorkflow，**Then** 校验失败并提示必须定义 rollback
7. **Given** KubernetesResource 支持任意 CRD，**When** 工作流被执行，**Then** 系统可以操作自定义资源而无需修改 Controller 代码

---

### User Story 10 - 手动触发执行 (Priority: P1)

作为平台运维人员，我需要能够手动触发预案的执行（首次部署、容灾切换等），并通过执行记录实时查看执行状态。

**Why this priority**: 手动触发和状态查看是运维人员的核心操作。

**Independent Test**: 可以通过创建 DRPlanExecution CR 或 annotation 触发预案执行，验证执行记录实时更新状态。

**Acceptance Scenarios**:

1. **Given** 一个状态为 Ready 的 DRPlan，**When** 运维人员创建 DRPlanExecution CR（指定操作类型为 Execute），**Then** 系统开始执行关联的 execute 工作流，执行记录状态变为 Running
2. **Given** 一个状态为 Ready 的 DRPlan，**When** 运维人员添加 annotation `dr.bkbcs.tencent.com/trigger=execute`，**Then** 系统自动创建 DRPlanExecution CR 并开始执行
3. **Given** 执行正在进行中，**When** 运维人员查询 DRPlanExecution，**Then** 可以看到当前执行进度、每个动作的状态（Pending/Running/Success/Failed）
4. **Given** 工作流执行完成，**When** 所有动作成功，**Then** DRPlanExecution 状态变为 Succeeded，DRPlan 状态变为 Executed
5. **Given** 执行过程中某个动作失败，**When** 工作流配置为 FailFast，**Then** 后续动作不再执行，DRPlanExecution 状态变为 Failed
6. **Given** DRPlan 已有执行正在进行中，**When** 尝试再次触发执行，**Then** 系统拒绝并提示"已有执行正在进行中"

---

### User Story 11 - 手动触发恢复 (Priority: P1)

作为平台运维人员，我需要能够手动触发预案的恢复操作，将系统切回到之前的状态。

**Why this priority**: 恢复操作必须由运维人员确认后手动触发，确保操作安全可控。

**Independent Test**: 可以通过创建 DRPlanExecution CR 或 annotation 触发恢复操作，验证恢复动作按顺序执行。

**Acceptance Scenarios**:

1. **Given** 一个已成功执行的 DRPlan（状态为 Executed），**When** 运维人员创建 DRPlanExecution CR（指定操作类型为 Revert），**Then** 系统开始执行关联的 revert 工作流
2. **Given** 一个已成功执行的 DRPlan（状态为 Executed），**When** 运维人员添加 annotation `dr.bkbcs.tencent.com/trigger=revert`，**Then** 系统自动创建 DRPlanExecution CR 并开始恢复
3. **Given** revert 执行完成，**When** 所有动作成功，**Then** DRPlanExecution 状态变为 Succeeded，DRPlan 状态变为 Ready
4. **Given** DRPlan 尚未执行（状态为 Ready），**When** 尝试创建 Revert 类型的执行，**Then** Webhook 拒绝并提示"尚未执行预案"

---

### User Story 12 - 查看执行历史 (Priority: P2)

作为平台运维人员，我需要能够查看一个容灾预案的所有执行记录，了解历史执行情况。

**Why this priority**: 执行历史对于审计和问题排查至关重要。

**Independent Test**: 可以通过 label selector 查询某个 DRPlan 关联的所有 DRPlanExecution。

**Acceptance Scenarios**:

1. **Given** 一个 DRPlan 已执行过多次 execute 和 revert，**When** 运维人员通过 `kubectl get drplanexecution -l drplan=<name>` 查询，**Then** 可以看到所有执行记录及其状态
2. **Given** 某次执行失败，**When** 查看对应的 DRPlanExecution，**Then** 可以看到失败的动作、错误信息、执行耗时

---

### User Story 13 - 取消执行 (Priority: P1)

作为平台运维人员，我需要能够取消正在进行的执行，以便在发现问题时紧急停止操作。

**Why this priority**: 紧急停止能力是生产环境的必要安全机制。

**Independent Test**: 可以通过 annotation 取消正在执行的任务，验证后续动作不再执行。

**Acceptance Scenarios**:

1. **Given** 一个 DRPlanExecution 正在执行（状态为 Running），**When** 运维人员添加 annotation `dr.bkbcs.tencent.com/cancel=true`，**Then** 系统停止后续动作执行，状态变为 Cancelled
2. **Given** 执行被取消，**When** 当前正在运行的动作完成或超时，**Then** 该动作状态正常记录，后续动作标记为 Skipped
3. **Given** 执行被取消，**When** 取消完成，**Then** DRPlan.status.currentExecution 清空，允许新的执行
4. **Given** 执行已完成（Succeeded/Failed/Cancelled），**When** 尝试取消，**Then** 系统忽略取消请求
5. **Given** 通过 DRPlan annotation 触发取消 `dr.bkbcs.tencent.com/trigger=cancel`，**When** 存在进行中的执行，**Then** 系统取消该执行

---

### Edge Cases

- 执行过程中 Controller 重启：系统应能从断点恢复继续执行
- Controller 与 API Server 网络分区：如果执行中出现网络分区超过 2 分钟，标记执行为 Unknown 状态，需要人工介入检查实际状态
- 同一预案被重复触发：系统应防止并发执行相同预案
- 工作流被修改时有执行在进行：执行应使用创建时的工作流快照和参数快照
- DRWorkflow 被引用时不能删除：需要检查引用关系
- DRPlanExecution 数量过多：系统不自动清理，由用户手动管理（kubectl delete）
- 参数值包含特殊字符：需要正确处理 JSON/YAML 中的转义
- 参数类型不匹配：需要验证参数值符合定义的类型
- 循环参数引用：参数值不能引用其他参数（避免复杂性）

## Requirements *(mandatory)*

### Functional Requirements

**DRWorkflow 相关**：
- **FR-001**: 系统 MUST 提供 DRWorkflow CRD，允许用户定义工作流名称、动作列表、失败策略
- **FR-002**: DRWorkflow MUST 支持 HTTP 类型动作（URL、Method、Headers、Body、Timeout、Retry）
- **FR-003**: DRWorkflow MUST 支持 Job 类型动作（Job 模板、命名空间、超时时间）
- **FR-004**: DRWorkflow MUST 支持 Localization 类型动作（operation: Create/Patch/Delete，目标集群、Localization 模板）
- **FR-004a**: Localization 动作 operation=Patch 时 MUST 显式定义 rollback 动作（校验规则）
- **FR-004b**: DRWorkflow MUST 支持 Subscription 类型动作（operation: Create/Patch/Delete，feeds、subscribers）
- **FR-004c**: Subscription 动作 operation=Patch 时 MUST 显式定义 rollback 动作（校验规则）
- **FR-004d**: DRWorkflow MUST 支持 KubernetesResource 类型动作（通用 K8s 资源操作）
- **FR-004e**: KubernetesResource 动作 MUST 支持 operation: Create/Apply/Patch/Delete
- **FR-004f**: KubernetesResource 动作 MUST 支持通过 manifest 字段定义资源（YAML 格式，支持参数占位符）
- **FR-004g**: KubernetesResource 动作 operation=Create 时，未定义 rollback 则自动回滚为删除资源
- **FR-004h**: KubernetesResource 动作 operation=Patch 时 MUST 显式定义 rollback 动作（校验规则）
- **FR-005**: DRWorkflow MUST 支持动作的顺序执行
- **FR-006**: DRWorkflow MUST 支持动作级别的重试策略和超时配置
- **FR-007**: DRWorkflow MUST 支持工作流级别的失败策略（FailFast 或 Continue）
- **FR-008**: DRWorkflow MUST 支持定义参数列表（参数名、类型、默认值、是否必填）
- **FR-009**: DRWorkflow MUST 支持在动作配置中使用参数占位符 `{{ .params.<name> }}`
- **FR-010**: 系统 MUST 在执行时将参数占位符替换为实际值

**步骤回滚**：
- **FR-011**: DRWorkflow 的每个步骤 MUST 支持定义可选的 `rollback` 动作
- **FR-012**: 系统 MUST 支持自动逆向回滚（当步骤未定义 rollback 时）
- **FR-013**: Localization 动作（operation=Create）的自动逆操作 MUST 为删除该 Localization
- **FR-013a**: Localization 动作（operation=Patch）MUST 使用显式定义的 rollback（无自动逆操作）
- **FR-013b**: Subscription 动作（operation=Create）的自动逆操作 MUST 为删除该 Subscription
- **FR-013c**: Subscription 动作（operation=Patch）MUST 使用显式定义的 rollback（无自动逆操作）
- **FR-013d**: KubernetesResource 动作（operation=Create）的自动逆操作 MUST 为删除该资源
- **FR-013e**: KubernetesResource 动作（operation=Patch）MUST 使用显式定义的 rollback（无自动逆操作）
- **FR-014**: Job 动作的自动逆操作 MUST 为删除该 Job
- **FR-015**: HTTP 动作无默认逆操作，回滚时 MUST 跳过（除非定义了自定义 rollback）
- **FR-016**: Revert 执行时 MUST 按已成功步骤的逆序执行回滚
- **FR-017**: 系统 MUST 在 Execution 中记录每个步骤创建的资源（用于自动回滚）

**DRPlan 相关**：
- **FR-018**: 系统 MUST 提供 DRPlan CRD，允许用户定义预案名称、Stage 编排、全局参数
- **FR-019**: DRPlan MUST 支持定义多个 Stage，每个 Stage 可包含多个 Workflow 引用
- **FR-020**: DRPlan MUST 支持 globalParams 自动传递给所有 Stage 中的 Workflow
- **FR-020a**: Stage 级 WorkflowReference params MUST 覆盖 globalParams 中的同名参数
- **FR-021**: DRPlan MUST 验证所有必填参数都已提供值
- **FR-021a**: Stage 中引用的 Workflow MUST 验证其必填参数在 globalParams 或 WorkflowReference.params 中已提供
- **FR-022**: DRPlan MUST 在引用的工作流不存在或被删除时更新状态为 Invalid
- **FR-023**: DRPlan MUST 维护预案状态：Ready / Executed / Invalid
- **FR-023a**: DRPlan Stage MUST 支持 `parallel` 字段控制 Stage 内 Workflow 是否并行执行
- **FR-023b**: DRPlan Stage MUST 支持 `dependsOn` 字段定义 Stage 间依赖关系
- **FR-023c**: 系统 MUST 验证 Stage 依赖关系不形成循环依赖
- **FR-023d**: 系统 MUST 在执行时按拓扑排序执行 Stage（满足依赖关系）
- **FR-023e**: DRPlan stages 字段 MUST 不为空（至少包含一个 Stage）
- **FR-023f**: 当 Stage 配置为 `parallel: true` 时，如果任一 Workflow 执行失败，系统 MUST 立即停止启动新的 Workflow，等待当前运行的 Workflow 完成，并将 Stage 标记为 Failed（FailFast 策略）

**DRPlanExecution 相关**：
- **FR-024**: 系统 MUST 提供 DRPlanExecution CRD，记录每次执行的状态和结果
- **FR-025**: DRPlanExecution MUST 包含：关联的 DRPlan、操作类型（Execute/Revert）、执行状态
- **FR-026**: DRPlanExecution MUST 在 Status 中记录实际使用的参数值（参数快照）
- **FR-027**: DRPlanExecution MUST 在 Status 中记录每个动作的执行状态、开始时间、结束时间、错误信息
- **FR-027a**: DRPlanExecution MUST 在 Status 中记录每个 Stage 的执行状态（Stage 模式）
- **FR-027b**: DRPlanExecution MUST 在 Status 中记录每个 Stage 内 Workflow 的执行状态和进度
- **FR-028**: DRPlanExecution MUST 记录每个步骤创建的资源引用（outputs），用于自动回滚
- **FR-029**: DRPlanExecution MUST 在动作执行时发送 Kubernetes Events
- **FR-029a**: DRPlanExecution MUST 在 Stage 开始/完成时发送 Kubernetes Events
- **FR-030**: 系统 MUST 防止在不合法状态下创建执行（如 Ready 状态不能创建 Revert 执行）

**触发机制**：
- **FR-031**: 系统 MUST 支持通过创建 DRPlanExecution CR 手动触发执行
- **FR-032**: 系统 MUST 支持通过 annotation `dr.bkbcs.tencent.com/trigger=execute` 触发执行
- **FR-033**: 系统 MUST 支持通过 annotation `dr.bkbcs.tencent.com/trigger=revert` 触发恢复
- **FR-034**: 系统 MUST 在 annotation 触发后自动创建对应的 DRPlanExecution CR
- **FR-035**: 系统 MUST 在 annotation 触发完成后自动清除该 annotation
- **FR-036**: 系统 MUST 防止同一 DRPlan 并发执行（已有执行进行中时拒绝新触发）

**取消执行**：
- **FR-037**: 系统 MUST 支持通过 annotation `dr.bkbcs.tencent.com/cancel=true` 取消 DRPlanExecution
- **FR-038**: 系统 MUST 支持通过 DRPlan annotation `dr.bkbcs.tencent.com/trigger=cancel` 取消当前执行
- **FR-039**: 取消执行时 MUST 停止后续 Action 执行，当前运行的 Action 等待完成或超时
- **FR-040**: 取消的执行状态 MUST 变为 Cancelled
- **FR-041**: 取消后 DRPlan.status.currentExecution MUST 清空，允许新执行
- **FR-042**: 已完成的执行（Succeeded/Failed/Cancelled）不响应取消请求

**通用**：
- **FR-043**: 系统 MUST 提供 Webhook 验证，在 CR 创建/更新时校验配置的合法性
- **FR-044**: 被引用的 DRWorkflow 不能被删除（需要先解除引用）

### Kubernetes Events 定义

系统在执行过程中发送以下 Kubernetes Events（关联到 DRPlanExecution）：

| Event Type | Reason | Message 模板 | 触发时机 |
|------------|--------|--------------|---------|
| Normal | ExecutionStarted | `Execution started for plan {planRef}` | 执行开始 |
| Normal | ActionStarted | `Action {actionName} started` | 动作开始 |
| Normal | ActionSucceeded | `Action {actionName} completed successfully in {duration}` | 动作成功 |
| Warning | ActionFailed | `Action {actionName} failed: {error}` | 动作失败 |
| Warning | ActionRetrying | `Action {actionName} retrying ({count}/{limit}): {error}` | 动作重试 |
| Normal | ExecutionSucceeded | `Execution completed successfully in {duration}` | 执行成功 |
| Warning | ExecutionFailed | `Execution failed at action {actionName}: {error}` | 执行失败 |
| Normal | ExecutionCancelled | `Execution cancelled by user` | 执行取消 |
| Normal | RevertStarted | `Revert started for plan {planRef}` | 回滚开始 |
| Normal | RevertSucceeded | `Revert completed successfully` | 回滚成功 |
| Warning | RevertFailed | `Revert failed at action {actionName}: {error}` | 回滚失败 |

### Key Entities

```
┌─────────────────────────────────────────────────────────────────┐
│                        DRWorkflow                               │
│  (工作流编排 - 定义参数、动作和回滚)                                 │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ metadata:                                                │   │
│  │   name: generic-deploy-workflow                         │   │
│  │ spec:                                                    │   │
│  │   parameters:                                            │   │
│  │     - name: targetCluster                                │   │
│  │       type: string                                       │   │
│  │       required: true                                     │   │
│  │     - name: notifyURL                                    │   │
│  │       type: string                                       │   │
│  │       default: "https://default.notify.com"             │   │
│  │                                                          │   │
│  │   failurePolicy: FailFast                                │   │
│  │   actions:                                               │   │
│  │     # Step 1: HTTP 通知（无默认逆操作，回滚时跳过）          │   │
│  │     - name: notify-oncall                                │   │
│  │       type: HTTP                                         │   │
│  │       http:                                              │   │
│  │         url: "{{ .params.notifyURL }}"                  │   │
│  │         method: POST                                     │   │
│  │       # 未定义 rollback，HTTP 类型回滚时跳过               │   │
│  │                                                          │   │
│  │     # Step 2: Localization Create（有自动逆操作：删除）     │   │
│  │     - name: scale-up-target                              │   │
│  │       type: Localization                                 │   │
│  │       localization:                                      │   │
│  │         operation: Create       # Create | Patch | Delete│   │
│  │         name: "dr-scale-{{ .planName }}"                │   │
│  │         namespace: "{{ .params.targetCluster }}"        │   │
│  │         feed: {...}             # 源资源引用              │   │
│  │         overrides: [...]        # 配置覆盖               │   │
│  │       # 未定义 rollback，自动逆操作：删除此 Localization    │   │
│  │                                                          │   │
│  │     # Step 3: Localization Patch（必须显式定义 rollback）   │   │
│  │     - name: update-replicas                              │   │
│  │       type: Localization                                 │   │
│  │       localization:                                      │   │
│  │         operation: Patch                                 │   │
│  │         name: "existing-localization"                    │   │
│  │         overrides: [...]  # 修改副本数为 3                 │   │
│  │       rollback:           # Patch 必须定义 rollback        │   │
│  │         type: Localization                               │   │
│  │         localization:                                    │   │
│  │           operation: Patch                               │   │
│  │           name: "existing-localization"                  │   │
│  │           overrides: [...]  # 恢复副本数为 1               │   │
│  │                                                          │   │
│  │     # Step 3: Job（有自动逆操作：删除）                     │   │
│  │     - name: run-sync-job                                 │   │
│  │       type: Job                                          │   │
│  │       job:                                               │   │
│  │         namespace: default                               │   │
│  │         template: {...}                                  │   │
│  │       # 未定义 rollback，自动逆操作：删除此 Job            │   │
│  │                                                          │   │
│  │     # Step 4: HTTP 带自定义回滚                           │   │
│  │     - name: switch-dns                                   │   │
│  │       type: HTTP                                         │   │
│  │       http:                                              │   │
│  │         url: "https://dns-api/switch"                   │   │
│  │         method: POST                                     │   │
│  │         body: '{"target": "{{ .params.targetCluster }}"}'│   │
│  │       rollback:                     # 自定义回滚动作       │   │
│  │         type: HTTP                                       │   │
│  │         http:                                            │   │
│  │           url: "https://dns-api/revert"                 │   │
│  │           method: POST                                   │   │
│  │                                                          │   │
│  │ status:                                                  │   │
│  │   phase: Ready                                           │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              ▲
                              │ 引用 + 传参
┌─────────────────────────────────────────────────────────────────┐
│                          DRPlan                                 │
│  (预案 - 关联工作流并传递参数)                                     │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ metadata:                                                │   │
│  │   name: prod-cluster-dr-plan                            │   │
│  │ spec:                                                    │   │
│  │   description: "生产环境双活切换"                              │   │
│  │                                                          │   │
│  │   # 全局参数                                              │   │
│  │   globalParams:                                          │   │
│  │     - name: targetCluster                                │   │
│  │       value: "cluster-backup-prod"                       │   │
│  │     - name: notifyURL                                    │   │
│  │       value: "https://prod.notify.com/alert"            │   │
│  │     - name: timeout                                      │   │
│  │       value: "15m"                                       │   │
│  │                                                          │   │
│  │   # Stage 编排                                            │   │
│  │   stages:                                                │   │
│  │     - name: execute-failover                             │   │
│  │       workflows:                                         │   │
│  │         - workflowRef:                                   │   │
│  │             name: generic-deploy-workflow                │   │
│  │             namespace: production                        │   │
│  │                                                          │   │
│  │ status:                                                  │   │
│  │   phase: Ready | Executed | Invalid                     │   │
│  │   lastExecutionTime: ...                                 │   │
│  │   lastExecutionRef: prod-dr-exec-20260130-001           │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ 触发执行
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      DRPlanExecution                            │
│  (执行记录 - 记录参数快照和执行状态)                                │
│  ┌─────────────────────────────────────────────────────────┐   │
│  │ metadata:                                                │   │
│  │   name: prod-dr-exec-20260130-001                       │   │
│  │   labels:                                                │   │
│  │     drplan: prod-cluster-dr-plan                        │   │
│  │ spec:                                                    │   │
│  │   planRef: prod-cluster-dr-plan                         │   │
│  │   operationType: Execute | Revert                       │   │
│  │ status:                                                  │   │
│  │   phase: Pending | Running | Succeeded | Failed         │   │
│  │   startTime: ...                                         │   │
│  │   completionTime: ...                                    │   │
│  │   # 参数快照（执行时的实际参数值）                          │   │
│  │   resolvedParams:                                        │   │
│  │     targetCluster: "cluster-backup-prod"                │   │
│  │     notifyURL: "https://prod.notify.com/alert"          │   │
│  │     timeout: "15m"                                       │   │
│  │   actionStatuses:                                        │   │
│  │     - name: notify-oncall                                │   │
│  │       phase: Succeeded                                   │   │
│  │       startTime: ...                                     │   │
│  │       completionTime: ...                                │   │
│  │     - name: switch-traffic                               │   │
│  │       phase: Running                                     │   │
│  │       startTime: ...                                     │   │
│  │     - name: verify-health                                │   │
│  │       phase: Pending                                     │   │
│  └─────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────┘
```

**实体关系**：
- **DRWorkflow** 定义参数模板、动作列表和步骤回滚，可被多个 DRPlan 和 Stage 复用
- **DRPlan** 通过 Stage 编排多个 Workflow，传递 globalParams 和 Stage 级参数
- **DRPlanExecution** 记录执行时的 Stage 状态、Workflow 执行状态、动作状态和 outputs

**参数替换流程**：
1. DRWorkflow 定义参数（parameters）和占位符（`{{ .params.xxx }}`）
2. DRPlan 提供 globalParams（全局参数）和 Stage 级 params（优先级更高）
3. 执行时系统将占位符替换为实际值
4. 参数合并优先级：Workflow 默认值 < globalParams < Stage params

**回滚流程**：

```
Execute 执行顺序: Step1 → Step2 → Step3 → Step4
                    │       │       │       │
                    ▼       ▼       ▼       ▼
               记录 outputs (创建的资源引用)

Revert 回滚顺序:  Step4 → Step3 → Step2 → Step1  (逆序)
                    │       │       │       │
                    ▼       ▼       ▼       ▼
              自定义回滚  自动删除  自动删除  跳过(HTTP)
```

| 动作类型 | 有自定义 rollback | 无自定义 rollback |
|---------|------------------|------------------|
| **Localization** (operation=Create) | 执行自定义 rollback | 自动删除创建的 Localization |
| **Localization** (operation=Patch) | 执行自定义 rollback | **不允许**（校验时强制要求 rollback） |
| **Localization** (operation=Delete) | 执行自定义 rollback | 无默认逆操作（跳过） |
| **Subscription** (operation=Create) | 执行自定义 rollback | 自动删除创建的 Subscription |
| **Subscription** (operation=Patch) | 执行自定义 rollback | **不允许**（校验时强制要求 rollback） |
| **Subscription** (operation=Delete) | 执行自定义 rollback | 无默认逆操作（跳过） |
| **KubernetesResource** (operation=Create) | 执行自定义 rollback | 自动删除创建的资源 |
| **KubernetesResource** (operation=Patch) | 执行自定义 rollback | **不允许**（校验时强制要求 rollback） |
| **KubernetesResource** (operation=Delete) | 执行自定义 rollback | 无默认逆操作（跳过） |
| **Job** | 执行自定义 rollback | 自动删除创建的 Job |
| **HTTP** | 执行自定义 rollback | 跳过（无默认逆操作） |

**回滚决策逻辑**：
1. 获取上次 Execute 成功的步骤列表（按执行顺序）
2. 逆序遍历每个步骤
3. 如果步骤定义了 `rollback` → 执行自定义回滚动作
4. 如果步骤未定义 `rollback` → 检查动作类型和 operation：
   - Localization(Create)/Subscription(Create)/KubernetesResource(Create)/Job → 从 outputs 获取资源引用，执行删除
   - Localization(Delete)/Subscription(Delete)/KubernetesResource(Delete)/HTTP → 跳过
   - Localization(Patch)/Subscription(Patch)/KubernetesResource(Patch) → 不会出现（校验已拦截）
5. 记录每个步骤的回滚状态

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 运维人员可以在 5 分钟内定义工作流并创建容灾预案
- **SC-002**: 触发执行后，第一个动作在 10 秒内开始执行
- **SC-003**: 单个工作流支持至少 20 个动作
- **SC-004**: 动作执行状态变更在 5 秒内反映到 DRPlanExecution Status 字段
- **SC-005**: Controller 重启后，正在执行的流程可以在 30 秒内恢复执行
- **SC-006**: 所有动作执行都产生可追溯的 Kubernetes Events 和结构化日志
- **SC-007**: 可以通过 kubectl 快速查询某个预案的所有执行历史

## Scale Targets

- 单集群最多 500 个 DRPlan
- 单集群最多 5000 个 DRPlanExecution（历史记录，用户手动清理）
- 单个 DRWorkflow 最多 20 个 Action

## Performance Requirements

- **Reconcile 频率**：DRPlanExecution Controller 的 reconcile 间隔可配置，默认 30 秒
- **状态更新延迟**：动作执行状态变更在 5 秒内反映到 Status 字段（SC-004）
- **执行启动延迟**：触发执行后，第一个动作在 10 秒内开始执行（SC-002）
- **恢复时间**：Controller 重启后，正在执行的流程可以在 30 秒内恢复执行（SC-005）

## Observability Requirements

### 日志策略

Controller 使用结构化 JSON 日志格式，遵循以下日志级别策略：

**Info 级别**（生产环境默认）：
- 执行开始/完成/失败（DRPlanExecution 级别）
- Stage 开始/完成/失败
- Workflow 开始/完成/失败
- 动作开始/成功/失败（简要信息，不含详细内容）
- 关键状态变更（Phase 转换）

**Debug 级别**（排查问题时启用）：
- 参数替换详情（占位符 → 实际值）
- 动作执行详情：HTTP 请求/响应体、Job manifest、Localization/Subscription 完整配置
- Reconcile 循环触发和处理详情
- 资源引用解析过程

**Warning 级别**：
- 重试触发（动作失败后的自动重试）
- 配置问题（如引用的 Workflow 不存在）

**Error 级别**：
- 执行失败且无法重试
- 系统级错误（API Server 通信失败、资源创建失败）

### Kubernetes Events

所有关键操作都通过 Kubernetes Events 记录（关联到 DRPlanExecution），便于 `kubectl describe` 查看。

### 指标（Metrics）

- 暂不强制要求 Prometheus 指标（v1 范围外）
- 后续版本可添加：执行次数、成功率、平均执行时间等

## Assumptions

- Clusternet 已部署在 Hub 集群，并正确配置了目标集群的访问权限
- Localization/Subscription 动作采用异步模型：CR 创建成功后动作即标记为成功，不等待 Clusternet 同步完成。如需确保同步完成，用户应在后续步骤添加健康检查动作
- HTTP 动作的目标接口可以从 Controller 所在网络访问
- Job 动作的容器镜像可以从集群内拉取
- 用户具备足够的 RBAC 权限来创建和管理相关资源

## Out of Scope (v1)

- **敏感数据管理**：HTTP 动作中的敏感数据（API Token、密码）首版不支持 Secret 引用，需明文配置或由用户在目标接口侧处理认证。后续版本迭代支持 `secretKeyRef`
