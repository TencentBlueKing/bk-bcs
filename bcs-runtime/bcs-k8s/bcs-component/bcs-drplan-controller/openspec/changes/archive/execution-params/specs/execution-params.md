# Spec: Execution-Level Params with Dynamic valueFrom

## 1. DRPlanExecution.spec.params（执行级参数）

### 1.1 字段定义（ADDED）

```yaml
# DRPlanExecution
spec:
  params:                # []Parameter, optional
  - name: releaseRevision
    value: "5"
```

- `params` 字段为可选，默认空列表。
- 类型与 `DRPlan.spec.globalParams` 完全一致（`[]Parameter`）。
- 字段位置：位于 `spec.revertExecutionRef` 之后。

### 1.2 参数合并优先级（MODIFIED）

合并顺序（低 → 高，高优先级覆盖同名低优先级）：

```
DRWorkflow.spec.parameters（默认值）
  ↓
DRPlan.spec.globalParams
  ↓
DRPlan.spec.stages[*].params（仅在该 stage 内）
  ↓
DRPlanExecution.spec.params（最高优先级）
```

### 1.3 Webhook 验证

- `params` 内每个 `Parameter` 必须有非空 `name`。
- `name` 不可重复。
- `value` 与 `valueFrom` 不能同时设置（互斥）。
- 参数名格式：`[a-zA-Z][a-zA-Z0-9_-]*`。

---

## 2. Parameter.valueFrom（动态参数解析）

### 2.1 字段定义（MODIFIED：Parameter struct）

```yaml
# Parameter 结构体（delta）
- name: jobName
  valueFrom:
    manifestRef:
      apiVersion: batch/v1
      kind: Job
      namespace: $(params.targetNamespace)  # 支持参数引用（仅 namespace/name/labelSelector）
      labelSelector: "app=bk-cmdb,component=auth"
      jsonPath: "{.metadata.name}"
      select: Last        # 当匹配多个资源时的选择策略
```

#### 字段说明

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `valueFrom.manifestRef.apiVersion` | string | 是 | 资源 apiVersion，如 `batch/v1` |
| `valueFrom.manifestRef.kind` | string | 是 | 资源 Kind，如 `Job` |
| `valueFrom.manifestRef.namespace` | string | 否 | 目标命名空间（可含模板占位符）|
| `valueFrom.manifestRef.name` | string | 否 | 资源精确名称（与 labelSelector 互斥）|
| `valueFrom.manifestRef.labelSelector` | string | 否 | 标签选择器（与 name 互斥）|
| `valueFrom.manifestRef.jsonPath` | string | 是 | JSONPath 表达式，如 `{.metadata.name}` |
| `valueFrom.manifestRef.select` | string | 否 | 多匹配选择策略，默认 `Last`（见下节）|

### 2.2 Select 策略

| 值 | 行为 |
|---|---|
| `Last`（默认）| 按 `metadata.creationTimestamp` 从大到小排序，取第一个 |
| `First` | 按 `metadata.creationTimestamp` 从小到大排序，取第一个 |
| `Single` | 要求精确匹配 1 个，若匹配 0 或 >1 则报错 |
| `Any` | 取列表中任意一个（顺序不定）|

### 2.3 模板渲染范围

`manifestRef` 内字段分两类：

- **可渲染**（在 List/Get 前先做 `$(params.xxx)` 渲染）：`namespace`、`name`、`labelSelector`
- **不渲染**（直接使用原始值）：`apiVersion`、`kind`、`jsonPath`、`select`

### 2.4 适用范围

`valueFrom` 可在以下层级使用：

- `DRPlanExecution.spec.params[*]`
- `DRPlan.spec.globalParams[*]`

### 2.5 Webhook 验证

- `valueFrom` 设置时，`value` 字段必须为空。
- `manifestRef.name` 与 `manifestRef.labelSelector` 不能同时设置。
- `manifestRef.apiVersion`、`manifestRef.kind`、`manifestRef.jsonPath` 必须非空。
- `manifestRef.select` 仅允许 `Last`、`First`、`Single`、`Any`（case-sensitive）。

---

## 3. SubscriptionAction.operation = Apply（ADDED）

### 3.1 Operation 枚举扩展

`SubscriptionAction.operation` 在现有基础上新增枚举值 `Apply`：

| 值 | 行为 |
|---|---|
| `Create`（默认）| 使用 `client.Create`，资源已存在时报错 |
| `Apply` | 使用 Server-Side Apply（`client.Patch(ApplyPatch)`），资源不存在则创建，存在则更新 |

### 3.2 Usage Example

```yaml
# DRWorkflow action
actions:
- name: apply-auth-subscription
  type: Subscription
  subscription:
    operation: Apply        # Server-Side Apply，幂等
    fieldManager: drplan-controller
    name: bk-cmdb-auth-$(params.releaseRevision)
    namespace: bcs-system
    feeds:
    - apiVersion: batch/v1
      kind: Job
      name: bk-cmdb-auth-register-$(params.releaseRevision)
      namespace: $(params.targetNamespace)
    subscribers:
    - clusterAffinity:
        matchLabels:
          env: production
```

---

## 4. 错误处理

| 场景 | 行为 |
|---|---|
| `valueFrom.manifestRef` List 调用失败 | 执行终止，返回错误，Execution Phase 置 Failed |
| List 结果为空 | 返回错误：`no resource found matching <criteria>` |
| JSONPath 表达式无效 | 返回错误：`invalid jsonPath expression` |
| JSONPath 结果为空字符串 | 返回错误：`jsonPath returned empty value` |
| Select=Single 匹配 >1 个 | 返回错误：`expected single match, got N` |

---

## 5. when 条件与 mode 参数（已有能力，补充说明）

`Action.when` 已支持，格式：`mode == "install"` 或 `mode == "upgrade"`。

`mode` 值来自 `DRPlanExecution.spec.mode`，注入 `globalParams["mode"]`，
**不可通过** `execution.spec.params` 中的同名参数覆盖（实现时需保护该 key）。

典型组合：

```yaml
# DRWorkflow action
- name: db-migrate
  type: Subscription
  when: 'mode == "upgrade"'          # 仅 Upgrade 时执行
  subscription:
    operation: Apply                  # 幂等（本次新增）
    name: db-migrate-$(params.dbMigrateJobName)  # 参数来自 execution.spec.params.valueFrom
    ...
```

---

## 6. 向后兼容性

- `DRPlanExecution.spec.params` 为新增可选字段，不设置时行为与现有完全一致。
- `Parameter.valueFrom` 为新增可选字段，不设置时沿用 `value` 语义。
- `SubscriptionAction.operation` 若不设置，默认 `Create`，行为与现有完全一致。
