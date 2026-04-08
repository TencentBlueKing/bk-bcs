# Design: Execution-Level Params with Dynamic valueFrom

## 1. API 层变更

### 1.1 common_types.go — Parameter struct 扩展

在现有 `Parameter` struct（第 24 行）中新增字段：

```go
// ValueFrom dynamically resolves a parameter value from a Kubernetes resource.
// Mutually exclusive with Value.
// +optional
ValueFrom *ParameterValueFrom `json:"valueFrom,omitempty"`
```

新增以下类型定义（追加至文件末尾）：

```go
// ParameterValueFrom specifies the source of a parameter value.
type ParameterValueFrom struct {
    // ManifestRef resolves a value from a Kubernetes resource field.
    // +optional
    ManifestRef *ManifestRef `json:"manifestRef,omitempty"`
}

// ManifestRef describes a Kubernetes resource and a JSONPath to extract a value.
type ManifestRef struct {
    // APIVersion is the resource API version (e.g. "batch/v1")
    // +kubebuilder:validation:Required
    APIVersion string `json:"apiVersion"`

    // Kind is the resource kind (e.g. "Job")
    // +kubebuilder:validation:Required
    Kind string `json:"kind"`

    // Namespace is the resource namespace. Supports $(params.xxx) placeholders.
    // +optional
    Namespace string `json:"namespace,omitempty"`

    // Name is the exact resource name. Mutually exclusive with LabelSelector.
    // Supports $(params.xxx) placeholders.
    // +optional
    Name string `json:"name,omitempty"`

    // LabelSelector is a label selector string (e.g. "app=foo,env=prod").
    // Mutually exclusive with Name. Supports $(params.xxx) placeholders.
    // +optional
    LabelSelector string `json:"labelSelector,omitempty"`

    // JSONPath is the JSONPath expression to extract the value (e.g. "{.metadata.name}")
    // +kubebuilder:validation:Required
    JSONPath string `json:"jsonPath"`

    // Select specifies which resource to pick when multiple matches are found.
    // One of: Last (default), First, Single, Any
    // +kubebuilder:validation:Enum=Last;First;Single;Any
    // +kubebuilder:default=Last
    // +optional
    Select string `json:"select,omitempty"`
}
```

### 1.2 common_types.go — SubscriptionAction.Operation 枚举扩展

将 `SubscriptionAction.Operation` 的 kubebuilder marker 从：

```go
// +kubebuilder:validation:Enum=Create;Patch;Delete
```

改为：

```go
// +kubebuilder:validation:Enum=Create;Apply;Patch;Delete
```

### 1.3 drplanexecution_types.go — DRPlanExecutionSpec 新增 Params

在 `DRPlanExecutionSpec.RevertExecutionRef` 字段之后追加：

```go
// Params are execution-level parameters with highest priority.
// They override DRPlan.spec.globalParams and stage-level params with the same name.
// Supports valueFrom for dynamic resolution.
// +optional
Params []Parameter `json:"params,omitempty"`
```

---

## 2. 执行器层变更

### 2.1 新建 internal/executor/param_resolver.go

职责：将 `[]Parameter`（含 `valueFrom`）解析为 `map[string]interface{}`。

核心函数签名：

```go
// resolveParams resolves a list of Parameters (possibly with valueFrom) into a map.
// alreadyResolved is the current context params used to render manifestRef namespace/name/labelSelector.
func resolveParams(
    ctx context.Context,
    dynamicClient dynamic.Interface,
    mapper meta.RESTMapper,
    params []Parameter,
    alreadyResolved map[string]interface{},
) (map[string]interface{}, error)
```

内部步骤：
1. 遍历 `params`，若 `valueFrom == nil` 直接取 `param.Value`。
2. 若 `valueFrom.manifestRef != nil`：
   a. 用 `utils.RenderTemplate` 渲染 `namespace`、`name`、`labelSelector`（注入 `alreadyResolved`）。
   b. 构造 `schema.GroupVersionResource`（via RESTMapper）。
   c. 按 `name`（单 Get）或 `labelSelector`（List）获取对象。
   d. 多对象时按 `select` 策略排序/选取。
   e. 用 `k8s.io/client-go/util/jsonpath` 提取字段值（去除首尾空白）。
3. 返回 `map[string]interface{}{param.Name: resolvedValue, ...}`。

### 2.2 native_executor.go — globalParams 合并逻辑扩展

在 `ExecutePlan` 的 globalParams 构建段（第 623-633 行）之后，追加 Execution 级参数解析与合并：

```go
// Resolve execution-level params (highest priority)
if len(execution.Spec.Params) > 0 {
    executionParams, err := resolveParams(ctx, e.dynamicClient, e.mapper,
        execution.Spec.Params, globalParams)
    if err != nil {
        return fmt.Errorf("resolving execution params: %w", err)
    }
    for k, v := range executionParams {
        globalParams[k] = v
    }
}
```

`NativePlanExecutor` 需要新增 `dynamicClient dynamic.Interface` 和 `mapper meta.RESTMapper` 字段。

同理，`DRPlan.spec.globalParams` 中的 `valueFrom` 也应在此阶段解析（将第 624-629 行改用 `resolveParams`）。

### 2.3 subscription_executor.go — Apply 操作支持

在 `SubscriptionActionExecutor.Execute` 方法中，将当前硬编码的 `e.client.Create(ctx, sub)` 替换为：

```go
func (e *SubscriptionActionExecutor) applySubscription(
    ctx context.Context, sub *clusternetappsv1alpha1.Subscription, operation string,
) error {
    switch operation {
    case "Apply":
        return e.client.Patch(ctx, sub, client.Apply,
            client.FieldOwner("drplan-controller"),
            client.ForceOwnership)
    default: // Create
        return e.client.Create(ctx, sub)
    }
}
```

---

## 3. Webhook 验证层变更

### 3.1 drplanexecution_webhook.go — 新增 Params 验证

在 `ValidateCreate`/`ValidateUpdate` 中增加：

```go
func validateExecutionParams(params []drv1alpha1.Parameter) field.ErrorList {
    var errs field.ErrorList
    seen := sets.NewString()
    for i, p := range params {
        path := field.NewPath("spec", "params").Index(i)
        if p.Name == "" {
            errs = append(errs, field.Required(path.Child("name"), "must not be empty"))
        }
        if seen.Has(p.Name) {
            errs = append(errs, field.Invalid(path.Child("name"), p.Name, "duplicate parameter name"))
        }
        seen.Insert(p.Name)
        if p.Value != "" && p.ValueFrom != nil {
            errs = append(errs, field.Invalid(path, p.Name, "value and valueFrom are mutually exclusive"))
        }
        if p.ValueFrom != nil && p.ValueFrom.ManifestRef != nil {
            ref := p.ValueFrom.ManifestRef
            if ref.Name != "" && ref.LabelSelector != "" {
                errs = append(errs, field.Invalid(path.Child("valueFrom", "manifestRef"), "",
                    "name and labelSelector are mutually exclusive"))
            }
            if ref.APIVersion == "" {
                errs = append(errs, field.Required(path.Child("valueFrom", "manifestRef", "apiVersion"), ""))
            }
            if ref.Kind == "" {
                errs = append(errs, field.Required(path.Child("valueFrom", "manifestRef", "kind"), ""))
            }
            if ref.JSONPath == "" {
                errs = append(errs, field.Required(path.Child("valueFrom", "manifestRef", "jsonPath"), ""))
            }
        }
    }
    return errs
}
```

---

## 4. DeepCopy 更新

运行 `make generate` 重新生成 `api/v1alpha1/zz_generated.deepcopy.go`，以覆盖新增的：
- `ParameterValueFrom.DeepCopyInto`
- `ManifestRef.DeepCopyInto`
- `DRPlanExecutionSpec.DeepCopyInto`（包含 Params slice）

---

## 5. CRD 更新

运行 `make manifests` 重新生成：
- `config/crd/bases/dr.bkbcs.tencent.com_drplanexecutions.yaml`
- `config/crd/bases/dr.bkbcs.tencent.com_drworkflows.yaml`（SubscriptionAction.operation 枚举扩展）
- `install/helm/.../crds/` 对应文件

---

## 6. when 条件与 execution-params 的配合

`Action.when` 已在 `native_executor.go` 中实现（`shouldExecuteActionByWhen`），
**不是** Phase 2 保留字段，当前已可用。

### 6.1 当前支持的表达式

```
mode == "install"    ← 仅 Install 时执行
mode == "upgrade"    ← 仅 Upgrade 时执行
operation == "upgrade"  ← 向后兼容别名（等同于 mode == "upgrade"）
```

限制：**只支持单个等值比较**，不支持 `||`、`&&`、`in` 等复合表达式。

### 6.2 `mode` 值来源

`mode` 通过 `DRPlanExecution.spec.mode`（`Install` / `Upgrade`）注入 `globalParams`，
在 `ExecutePlan` 的第 631-633 行：

```go
if mode := strings.TrimSpace(execution.Spec.Mode); mode != "" {
    globalParams["mode"] = strings.ToLower(mode)
}
```

execution-params 的合并在此之后执行，**不会覆盖 `mode`**（`mode` 由 `execution.Spec.Mode`
专属控制，不允许通过 `execution.Spec.Params` 覆盖）。

> 实现时需在 `resolveParams` 合并结果写入 `globalParams` 前排除 `mode` key。

### 6.3 对标 Helm Hook 的完整用法示例

```yaml
# DRWorkflow stages
stages:
# 对标 helm pre-install hook
- name: pre-install
  actions:
  - name: init-schema
    type: Subscription
    when: 'mode == "install"'
    subscription: ...

# 对标 helm pre-upgrade hook（含动态 Job 名）
- name: pre-upgrade
  actions:
  - name: db-migrate
    type: Subscription
    when: 'mode == "upgrade"'
    subscription:
      operation: Apply
      name: db-migrate-$(params.dbMigrateJobName)
      ...

# 对标 helm post-install + post-upgrade（不加 when = 总是执行）
- name: post-deploy
  actions:
  - name: validate
    type: Subscription
    subscription: ...
```

**Install 时**（`mode: Install`）：pre-install 执行，pre-upgrade **跳过（PhaseSkipped）**，post-deploy 执行。

**Upgrade 时**（`mode: Upgrade`）：pre-install **跳过**，pre-upgrade 执行（`dbMigrateJobName` 由 `valueFrom` 动态解析），post-deploy 执行。

### 6.4 Helm Hook 对标完整矩阵

| Helm Hook | DRPlan 对应 | 状态 |
|---|---|---|
| `pre-install` | stage + `when: 'mode == "install"'` | ✅ 已支持 |
| `post-install` | stage（install 后）+ `when: 'mode == "install"'` | ✅ 已支持 |
| `pre-upgrade` | stage + `when: 'mode == "upgrade"'` | ✅ 已支持 |
| `post-upgrade` | stage（upgrade 后）+ `when: 'mode == "upgrade"'` | ✅ 已支持 |
| `pre-delete` | `operationType: Revert` 专属 stage | ✅ 已支持 |
| `hook-weight`（排序）| stage 内 action 串行顺序 | ✅ 已支持 |
| `hook-delete-policy: before-hook-creation` | `operation: Apply`（本次新增）| 🔜 新增 |
| Job 名含 Revision | `valueFrom.manifestRef`（本次新增）| 🔜 新增 |
| `hook-delete-policy: hook-succeeded` | 暂无自动清理，`operation: Apply` 规避了对它的需求 | ⚠️ 无需实现 |
| `pre-rollback` / `post-rollback` | Revert 流程（无专属区分）| ⚠️ 部分覆盖 |
| 多条件 `when` | 不支持，仅单个 `mode == X` | ❌ 未来扩展 |

---

## 7. 测试策略

| 文件 | 测试类型 | 覆盖点 |
|---|---|---|
| `internal/executor/param_resolver_test.go` | 单元测试（mock dynamic client）| valueFrom 解析、select 策略、错误分支 |
| `internal/executor/native_executor_test.go`（扩展）| 单元测试 | execution.Spec.Params 优先级合并 |
| `internal/executor/subscription_executor_test.go`（扩展）| 单元测试 | operation=Apply 分支 |
| `internal/webhook/drplanexecution_webhook_test.go`（扩展）| 单元测试 | Params 字段验证 |

---

## 7. 示例用法

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: upgrade-v5
  namespace: bcs-system
spec:
  planRef: bk-cmdb-dr-plan
  operationType: Execute
  mode: Upgrade
  params:
  - name: releaseRevision
    value: "5"                        # 已知时静态填写
  - name: jobName
    valueFrom:
      manifestRef:
        apiVersion: batch/v1
        kind: Job
        namespace: $(params.targetNamespace)
        labelSelector: "app=bk-cmdb,component=auth-register"
        jsonPath: "{.metadata.name}"
        select: Last                   # 取最新创建的 Job 名
```
