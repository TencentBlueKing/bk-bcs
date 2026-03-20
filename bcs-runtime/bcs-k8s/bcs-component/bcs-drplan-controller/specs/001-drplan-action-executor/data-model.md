# Data Model: 容灾策略 CR 及动作执行器

**Date**: 2026-01-30

## Entity Relationship

```
┌─────────────┐         ┌─────────────┐         ┌──────────────────┐
│ DRWorkflow  │◄────────│   DRPlan    │────────►│ DRPlanExecution  │
│             │ 1:N     │             │ 1:N     │                  │
│ (模板定义)   │         │  (预案实例)  │         │   (执行记录)      │
└─────────────┘         └─────────────┘         └──────────────────┘
       │                                                │
       │ contains                                       │ contains
       ▼                                                ▼
┌─────────────┐                                 ┌──────────────────┐
│   Action    │                                 │  ActionStatus    │
│  (动作定义)  │                                 │   (动作状态)      │
└─────────────┘                                 └──────────────────┘
```

## DRWorkflow

工作流模板，定义可复用的动作序列。

### Spec

| Field         | Type           | Required | Description                          |
| ------------- | -------------- | -------- | ------------------------------------ |
| executor      | ExecutorConfig | No       | 执行引擎配置（预留扩展）             |
| parameters    | []Parameter    | No       | 参数定义列表                         |
| actions       | []Action       | Yes      | 动作列表（按顺序执行）               |
| failurePolicy | string         | No       | 失败策略：FailFast（默认）/ Continue |

### ExecutorConfig（扩展预留）

| Field       | Type        | Required | Description                        |
| ----------- | ----------- | -------- | ---------------------------------- |
| type        | string      | No       | 执行引擎类型：Native（默认）/ Argo |
| argoOptions | ArgoOptions | No       | Argo 引擎配置（type=Argo 时有效）  |

### ArgoOptions（扩展预留）

| Field              | Type        | Required | Description                                      |
| ------------------ | ----------- | -------- | ------------------------------------------------ |
| namespace          | string      | No       | Argo Workflow 命名空间（默认与 DRWorkflow 相同） |
| serviceAccountName | string      | No       | Argo Workflow 使用的 ServiceAccount              |
| ttlStrategy        | TTLStrategy | No       | Workflow 保留策略                                |

### TTLStrategy（扩展预留）

| Field                  | Type | Required | Description    |
| ---------------------- | ---- | -------- | -------------- |
| secondsAfterCompletion | int  | No       | 完成后保留秒数 |
| secondsAfterSuccess    | int  | No       | 成功后保留秒数 |
| secondsAfterFailure    | int  | No       | 失败后保留秒数 |

> **注意**：首版仅实现 Native 引擎，Argo 相关字段为扩展预留，暂不生效。

### Parameter

| Field       | Type   | Required | Description                                |
| ----------- | ------ | -------- | ------------------------------------------ |
| name        | string | Yes      | 参数名称                                   |
| type        | string | No       | 参数类型：string（默认）/ number / boolean |
| required    | bool   | No       | 是否必填（默认 false）                     |
| default     | string | No       | 默认值                                     |
| description | string | No       | 参数描述                                   |

### Action

| Field        | Type                      | Required | Description                                                            |
| ------------ | ------------------------- | -------- | ---------------------------------------------------------------------- |
| name         | string                    | Yes      | 动作名称（唯一）                                                       |
| type         | string                    | Yes      | 动作类型：HTTP / Job / Localization / Subscription / KubernetesResource |
| http         | HTTPAction                | Cond     | HTTP 配置（type=HTTP 时必填）                                          |
| job          | JobAction                 | Cond     | Job 配置（type=Job 时必填）                                            |
| localization | LocalizationAction        | Cond     | Localization 配置（type=Localization 时必填）                          |
| subscription | SubscriptionAction        | Cond     | Subscription 配置（type=Subscription 时必填）                          |
| resource     | KubernetesResourceAction  | Cond     | Kubernetes 资源配置（type=KubernetesResource 时必填）                  |
| timeout      | string                    | No       | 超时时间（默认 5m）                                                    |
| retryPolicy  | RetryPolicy               | No       | 重试策略                                                               |
| rollback     | Action                    | No       | 自定义回滚动作                                                         |
| dependsOn    | []string                  | No       | 依赖的动作名称列表（扩展预留，用于 DAG）                               |
| when         | string                    | No       | 条件表达式（扩展预留，如 `{{ .outputs.step1.phase == 'Succeeded' }}`） |

> **注意**：`dependsOn` 和 `when` 为 Phase 2 扩展预留，首版忽略这些字段，按列表顺序执行。

### HTTPAction

| Field              | Type              | Required | Description                  |
| ------------------ | ----------------- | -------- | ---------------------------- |
| url                | string            | Yes      | 请求 URL（支持参数占位符）   |
| method             | string            | No       | HTTP 方法（默认 GET）        |
| headers            | map[string]string | No       | 请求头                       |
| body               | string            | No       | 请求体（支持参数占位符）     |
| successCodes       | []int             | No       | 成功状态码（默认 [200-299]） |
| insecureSkipVerify | bool              | No       | 跳过 TLS 验证                |

### JobAction

| Field                   | Type            | Required | Description                  |
| ----------------------- | --------------- | -------- | ---------------------------- |
| namespace               | string          | No       | Job 命名空间（默认 default） |
| template                | JobTemplateSpec | Yes      | Job 模板                     |
| ttlSecondsAfterFinished | int             | No       | 完成后保留时间               |

### LocalizationAction

| Field     | Type                   | Required | Description                                                        |
| --------- | ---------------------- | -------- | ------------------------------------------------------------------ |
| operation | string                 | No       | 操作类型：Create（默认）/ Patch / Delete                           |
| name      | string                 | Yes      | Localization CR 名称（支持占位符）                                 |
| namespace | string                 | Yes      | Localization CR 命名空间（即 ManagedCluster 命名空间，支持占位符） |
| priority  | int                    | No       | 优先级，数值越大优先级越高（默认 500）                             |
| feed      | Feed                   | Cond     | 源资源引用（operation=Create 时必填）                              |
| overrides | []LocalizationOverride | No       | 配置覆盖列表（operation=Create/Patch 时有效）                      |

**校验规则**：
- `operation=Create` 时，`feed` 必须提供
- `operation=Patch` 时，父级 Action 必须定义 `rollback` 字段，否则校验失败

### LocalizationOverride

| Field         | Type   | Required | Description                                  |
| ------------- | ------ | -------- | -------------------------------------------- |
| name          | string | Yes      | Override 名称，用于标识                      |
| type          | string | Yes      | Override 类型：JSONPatch / MergePatch / Helm |
| value         | string | Yes      | Override 内容（YAML/JSON 格式）              |
| overrideChart | bool   | No       | 是否覆盖 HelmChart CR（仅 type=Helm 时有效） |

### SubscriptionAction

| Field              | Type         | Required | Description                                 |
| ------------------ | ------------ | -------- | ------------------------------------------- |
| operation          | string       | No       | 操作类型：Create（默认）/ Patch / Delete    |
| name               | string       | Yes      | Subscription 名称（支持占位符）             |
| namespace          | string       | No       | Subscription 命名空间                       |
| feeds              | []Feed       | Cond     | 订阅源列表（operation=Create/Patch 时有效） |
| subscribers        | []Subscriber | Cond     | 订阅者列表（operation=Create/Patch 时有效） |
| schedulingStrategy | string       | No       | 调度策略：Replication / Dividing            |

**校验规则**：
- `operation=Patch` 时，父级 Action 必须定义 `rollback` 字段，否则校验失败

### Feed

| Field      | Type   | Required | Description   |
| ---------- | ------ | -------- | ------------- |
| apiVersion | string | Yes      | 资源 API 版本 |
| kind       | string | Yes      | 资源类型      |
| name       | string | Yes      | 资源名称      |
| namespace  | string | No       | 资源命名空间  |

### Subscriber

| Field           | Type            | Required | Description           |
| --------------- | --------------- | -------- | --------------------- |
| clusterAffinity | ClusterAffinity | No       | 集群亲和性选择        |
| weight          | int             | No       | 权重（Dividing 策略） |

### KubernetesResourceAction

| Field     | Type   | Required | Description                                          |
| --------- | ------ | -------- | ---------------------------------------------------- |
| operation | string | No       | 操作类型：Create（默认）/ Apply / Patch / Delete     |
| manifest  | string | Yes      | Kubernetes 资源 manifest（YAML 格式，支持参数占位符） |

**说明**：
- `manifest` 字段包含完整的 Kubernetes 资源定义（YAML 格式）
- 支持参数占位符（如 `{{ .params.resourceName }}`）
- 支持任意 K8s 资源类型（内置资源和自定义 CRD）

**校验规则**：
- `operation=Patch` 时，父级 Action 必须定义 `rollback` 字段，否则校验失败
- `operation=Create` 时，未定义 `rollback` 则自动回滚为删除该资源

**示例**：
```yaml
resource:
  operation: Create
  manifest: |
    apiVersion: v1
    kind: ConfigMap
    metadata:
      name: "{{ .params.configName }}"
      namespace: "{{ .params.namespace }}"
    data:
      key: "{{ .params.value }}"
```

### RetryPolicy

| Field             | Type   | Required | Description          |
| ----------------- | ------ | -------- | -------------------- |
| limit             | int    | No       | 重试次数（默认 3）   |
| interval          | string | No       | 重试间隔（默认 5s）  |
| backoffMultiplier | float  | No       | 退避倍数（默认 2.0） |

### Status

| Field              | Type        | Description       |
| ------------------ | ----------- | ----------------- |
| phase              | string      | Ready / Invalid   |
| conditions         | []Condition | 状态条件          |
| observedGeneration | int64       | 观测的 generation |

---

## DRPlan

容灾预案，通过 Stage 编排多个 Workflow，支持依赖和并行执行。

### Spec

| Field         | Type        | Required | Description                      |
| ------------- | ----------- | -------- | -------------------------------- |
| description   | string      | No       | 预案描述                         |
| stages        | []Stage     | Yes      | Stage 编排列表                   |
| globalParams  | []Parameter | No       | 全局参数（传递给所有 Workflow）  |
| failurePolicy | string      | No       | 失败策略：Stop（默认）/ Continue |

### Stage

| Field         | Type                | Required | Description                                          |
| ------------- | ------------------- | -------- | ---------------------------------------------------- |
| name          | string              | Yes      | Stage 名称（唯一）                                   |
| description   | string              | No       | Stage 描述                                           |
| dependsOn     | []string            | No       | 依赖的 Stage 名称列表                                |
| parallel      | bool                | No       | 是否并行执行本 Stage 内的所有 Workflow（默认 false） |
| workflows     | []WorkflowReference | Yes      | Workflow 引用列表                                    |
| failurePolicy | FailurePolicy       | No       | Stage 级失败策略（覆盖全局）                         |

### WorkflowReference

| Field       | Type            | Required | Description                            |
| ----------- | --------------- | -------- | -------------------------------------- |
| workflowRef | ObjectReference | Yes      | Workflow 引用（name + namespace）      |
| params      | []Parameter     | No       | 参数（覆盖 globalParams 中的同名参数） |

**参数合并规则**：
1. Workflow 定义的默认值
2. 覆盖为 DRPlan.globalParams 中的值
3. 再覆盖为 WorkflowReference.params 中的值（优先级最高）

### Status

| Field                 | Type            | Description                                                             |
| --------------------- | --------------- | ----------------------------------------------------------------------- |
| phase                 | string          | Ready / Executed / Invalid                                              |
| conditions            | []Condition     | 状态条件                                                                |
| lastExecutionTime     | metav1.Time     | 最后成功操作的完成时间（不论 Execute 还是 Revert）                     |
| lastExecutionRef      | string          | 最后成功操作的 execution 名称（不论 Execute 还是 Revert）              |
| currentExecution      | ObjectReference | 当前进行中的执行引用（用于并发控制，同时只能有一个 execution 运行）    |
| lastProcessedTrigger  | string          | **DEPRECATED** - annotation 触发机制已移除，字段保留仅为向后兼容       |
| executionHistory      | []ExecutionRecord | 最近 10 条执行历史（新到旧），包含 Execute 和 Revert 操作。即使 execution CR 被删除，历史仍保留（通过 finalizer 确保） |
| observedGeneration    | int64           | 观测的 generation                                                       |

### ExecutionRecord

历史执行记录，保存在 `DRPlan.Status.executionHistory` 中，最多保留 10 条。

| Field          | Type        | Description                                                        |
| -------------- | ----------- | ------------------------------------------------------------------ |
| name           | string      | DRPlanExecution 名称                                               |
| namespace      | string      | DRPlanExecution 命名空间                                           |
| operationType  | string      | Execute / Revert - 操作类型                                        |
| phase          | string      | Pending / Running / Succeeded / Failed / Cancelled - 执行状态      |
| startTime      | metav1.Time | 开始时间                                                           |
| completionTime | metav1.Time | 完成时间（可选，如果 execution 被强制删除会自动填充当前时间）     |

**注意**: 
- 历史记录通过 finalizer 机制确保完整性，即使 execution CR 被删除，记录仍会保留
- 如果 execution 在运行时被强制删除，phase 会自动标记为 Cancelled
- 通过 `executionHistory[0]` 可以获取最后一次操作的详情

---

## DRPlanExecution

执行记录，记录单次执行的状态和结果。

### Spec

| Field              | Type   | Required            | Description                                                                                          |
| ------------------ | ------ | ------------------- | ---------------------------------------------------------------------------------------------------- |
| planRef            | string | Yes                 | 关联的 DRPlan 名称                                                                                   |
| operationType      | string | Yes                 | 操作类型：Execute / Revert                                                                           |
| revertExecutionRef | string | **Yes (for Revert)** | **（Revert 必填）** 指定要回滚哪个 execution。必须引用 operationType=Execute 且 phase=Succeeded 的 execution。Webhook 会验证 |

### Status

| Field          | Type             | Description                                        |
| -------------- | ---------------- | -------------------------------------------------- |
| phase          | string           | Pending / Running / Succeeded / Failed / Cancelled |
| startTime      | metav1.Time      | 开始时间                                           |
| completionTime | metav1.Time      | 完成时间                                           |
| stageStatuses  | []StageStatus    | 各 Stage 执行状态                                  |
| summary        | ExecutionSummary | 执行统计摘要                                       |
| message        | string           | 状态消息                                           |
| conditions     | []Condition      | 状态条件                                           |

### StageStatus

| Field              | Type                      | Description                                      |
| ------------------ | ------------------------- | ------------------------------------------------ |
| name               | string                    | Stage 名称                                       |
| phase              | string                    | Pending / Running / Succeeded / Failed / Skipped |
| parallel           | bool                      | 是否并行执行                                     |
| dependsOn          | []string                  | 依赖的 Stage 列表                                |
| startTime          | metav1.Time               | 开始时间                                         |
| completionTime     | metav1.Time               | 完成时间                                         |
| duration           | string                    | 执行时长                                         |
| message            | string                    | 状态消息                                         |
| workflowExecutions | []WorkflowExecutionStatus | 各 Workflow 执行状态                             |

### WorkflowExecutionStatus

| Field          | Type            | Description                                      |
| -------------- | --------------- | ------------------------------------------------ |
| workflowRef    | ObjectReference | Workflow 引用                                    |
| phase          | string          | Pending / Running / Succeeded / Failed / Skipped |
| startTime      | metav1.Time     | 开始时间                                         |
| completionTime | metav1.Time     | 完成时间                                         |
| duration       | string          | 执行时长                                         |
| progress       | string          | 进度信息（如 "2/5 actions completed"）           |
| currentAction  | string          | 当前执行的动作名称                               |
| message        | string          | 状态消息                                         |
| actionStatuses | []ActionStatus  | 动作执行状态（详细信息）                         |

### ExecutionSummary

| Field              | Type | Description        |
| ------------------ | ---- | ------------------ |
| totalStages        | int  | Stage 总数         |
| completedStages    | int  | 已完成 Stage 数    |
| runningStages      | int  | 运行中 Stage 数    |
| pendingStages      | int  | 等待中 Stage 数    |
| failedStages       | int  | 失败 Stage 数      |
| totalWorkflows     | int  | Workflow 总数      |
| completedWorkflows | int  | 已完成 Workflow 数 |
| runningWorkflows   | int  | 运行中 Workflow 数 |
| pendingWorkflows   | int  | 等待中 Workflow 数 |
| failedWorkflows    | int  | 失败 Workflow 数   |

### ActionStatus

| Field          | Type          | Description                                      |
| -------------- | ------------- | ------------------------------------------------ |
| name           | string        | 动作名称                                         |
| phase          | string        | Pending / Running / Succeeded / Failed / Skipped |
| startTime      | metav1.Time   | 开始时间                                         |
| completionTime | metav1.Time   | 完成时间                                         |
| retryCount     | int           | 重试次数                                         |
| message        | string        | 状态消息/错误信息                                |
| outputs        | ActionOutputs | 动作输出（用于回滚）                             |

### ActionOutputs

| Field           | Type            | Description                                      |
| --------------- | --------------- | ------------------------------------------------ |
| jobRef          | ObjectReference | 创建的 Job 引用                                  |
| localizationRef | ObjectReference | 创建的 Localization 引用                         |
| subscriptionRef | ObjectReference | 创建的 Subscription 引用                         |
| resourceRef     | ObjectReference | 创建的通用 K8s 资源引用（KubernetesResource）     |
| httpResponse    | HTTPResponse    | HTTP 响应摘要                                    |

### HTTPResponse

| Field      | Type   | Description    |
| ---------- | ------ | -------------- |
| statusCode | int    | 响应状态码     |
| body       | string | 响应体（截断） |

---

## State Machines

### DRWorkflow Phase

```
           ┌─────────────────────┐
           │      Created        │
           └──────────┬──────────┘
                      │ validate
                      ▼
           ┌─────────────────────┐
      ┌────│       Ready         │◄───────────┐
      │    └──────────┬──────────┘            │
      │               │ spec changed          │
      │               ▼                       │
      │    ┌─────────────────────┐            │
      │    │      Validating     │────────────┘
      │    └──────────┬──────────┘  valid
      │               │ invalid
      │               ▼
      │    ┌─────────────────────┐
      └───►│      Invalid        │
           └─────────────────────┘
```

### DRPlan Phase

```
           ┌─────────────────────┐
           │       Ready         │◄───────────┐
           └──────────┬──────────┘            │
                      │ trigger execute       │ revert succeeded
                      ▼                       │
           ┌─────────────────────┐            │
           │      Executed       │────────────┘
           └──────────┬──────────┘
                      │ workflow deleted
                      ▼
           ┌─────────────────────┐
           │      Invalid        │
           └─────────────────────┘
```

### StageStatus Phase

```
           ┌─────────────────────┐
           │      Pending        │
           └──────────┬──────────┘
                      │ dependencies met
                      ▼
           ┌─────────────────────┐
           │      Running        │
           └──────────┬──────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
          ▼           ▼           ▼
┌───────────────┐ ┌────────┐ ┌───────────┐
│   Succeeded   │ │ Failed │ │ Skipped   │
└───────────────┘ └────────┘ └───────────┘
```

### DRPlanExecution Phase

```
           ┌─────────────────────┐
           │      Pending        │
           └──────────┬──────────┘
                      │ start
                      ▼
           ┌─────────────────────┐
           │      Running        │
           └──────────┬──────────┘
                      │
          ┌───────────┼───────────┐
          │           │           │
          ▼           ▼           ▼
┌───────────────┐ ┌────────┐ ┌───────────┐
│   Succeeded   │ │ Failed │ │ Cancelled │
└───────────────┘ └────────┘ └───────────┘
│    Succeeded    │     │     Failed      │
└─────────────────┘     └─────────────────┘
```

### ActionStatus Phase

```
           ┌─────────────────────┐
           │      Pending        │
           └──────────┬──────────┘
                      │ start
                      ▼
           ┌─────────────────────┐
           │      Running        │◄──────────┐
           └──────────┬──────────┘           │
                      │                      │ retry
          ┌───────────┼───────────┐          │
          │           │           │          │
          ▼           ▼           ▼          │
┌─────────────┐ ┌───────────┐ ┌─────────────┐│
│  Succeeded  │ │  Skipped  │ │   Failed    │┘
└─────────────┘ └───────────┘ └─────────────┘
```

---

## Validation Rules

### DRWorkflow

1. `actions` 不能为空
2. `actions[].name` 在工作流内唯一
3. `actions[].type` 必须是 HTTP/Job/Localization/Subscription/KubernetesResource 之一
4. 参数占位符必须在 `parameters` 中定义
5. `rollback.type` 如果定义，必须有效
6. Localization/Subscription/KubernetesResource `operation=Patch` 时必须定义 `rollback`
7. KubernetesResource 的 `manifest` 必须是有效的 YAML 格式

### DRPlan

1. `stages` 不能为空
2. 所有 Stage 中引用的 Workflow 必须存在
3. Stage 名称必须唯一
4. `dependsOn` 不能形成循环依赖
5. 必填参数（required=true）必须在 `globalParams` 或 WorkflowReference.params 中提供
6. 参数值类型必须匹配定义

### DRPlanExecution

1. `planRef` 引用的 DRPlan 必须存在
2. `operationType=Execute` 时，DRPlan.phase 必须是 Ready
3. `operationType=Revert` 时，DRPlan.phase 必须是 Executed
4. 同一 DRPlan 不能并发执行（currentExecution 为空）

---

## Labels & Annotations

### Labels

| Label                           | Applied To      | Description            |
| ------------------------------- | --------------- | ---------------------- |
| `dr.bkbcs.tencent.com/plan`     | DRPlanExecution | 关联的 DRPlan 名称     |
| `dr.bkbcs.tencent.com/workflow` | DRPlan          | 关联的 DRWorkflow 名称 |

### Annotations

| Annotation                     | Applied To      | Description                         |
| ------------------------------ | --------------- | ----------------------------------- |
| `dr.bkbcs.tencent.com/trigger` | DRPlan          | 触发操作：execute / revert / cancel |
| `dr.bkbcs.tencent.com/cancel`  | DRPlanExecution | 取消执行：true                      |

### Finalizers

| Finalizer                                  | Applied To      | Description      |
| ------------------------------------------ | --------------- | ---------------- |
| `dr.bkbcs.tencent.com/workflow-protection` | DRWorkflow      | 防止被引用时删除 |
| `dr.bkbcs.tencent.com/execution-cleanup`   | DRPlanExecution | 清理创建的资源   |
