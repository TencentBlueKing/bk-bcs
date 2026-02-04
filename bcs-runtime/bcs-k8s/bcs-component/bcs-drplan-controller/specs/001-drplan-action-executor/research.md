# Research: 容灾策略 CR 及动作执行器

**Date**: 2026-01-30  
**Status**: Complete

## 1. Kubernetes Operator 框架选择

**Decision**: controller-runtime + kubebuilder

**Rationale**:
- kubebuilder 是 Kubernetes 官方推荐的 Operator 脚手架工具
- controller-runtime 提供成熟的 Reconcile 循环、Leader Election、Metrics 等能力
- BCS 项目已有 kubebuilder 使用经验，团队熟悉

**Alternatives Considered**:
- Operator SDK: 封装了 kubebuilder，但增加了额外抽象层，不必要
- client-go 裸写: 需要手动处理太多细节，开发效率低
- kopf (Python): 不符合 BCS Go 技术栈要求

## 2. CRD API 版本策略

**Decision**: 使用 `v1alpha1` 作为初始版本

**Rationale**:
- alpha 版本表明 API 可能变更，用户有预期
- 待功能稳定后升级到 `v1beta1`，最终 `v1`
- 遵循 Kubernetes API 版本演进惯例

**Migration Path**:
1. `v1alpha1`: 初始开发，API 可能频繁变更
2. `v1beta1`: 功能稳定，API 冻结，开始收集生产反馈
3. `v1`: 生产就绪，提供长期支持

## 3. 参数模板引擎

**Decision**: 使用 Go text/template

**Rationale**:
- Go 标准库，无额外依赖
- 语法简单（`{{ .params.xxx }}`），用户易学
- 支持条件、循环等高级特性（预留扩展）

**Alternatives Considered**:
- Jsonnet: 功能强大但学习曲线陡峭
- CEL (Common Expression Language): Kubernetes 原生支持，但主要用于验证而非模板
- envsubst: 功能太简单，不支持默认值

**Implementation Notes**:
```go
// 参数替换示例
tmpl, _ := template.New("action").Parse(actionConfig)
var buf bytes.Buffer
tmpl.Execute(&buf, map[string]interface{}{
    "params": resolvedParams,
    "planName": plan.Name,
})
```

**错误处理策略**:

| 错误场景                           | 处理方式           | 说明                                  |
| ---------------------------------- | ------------------ | ------------------------------------- |
| 未定义参数 `{{ .params.unknown }}` | 返回错误，阻止执行 | 模板使用 `Option("missingkey=error")` |
| 空值参数 `{{ .params.empty }}`     | 替换为空字符串     | 允许空值，由动作自行处理              |
| 模板语法错误 `{{ .params.`         | 返回错误，阻止执行 | `template.Parse()` 返回错误           |
| 类型不匹配（期望 int 得到 string） | 替换后由动作校验   | 模板引擎不做类型检查                  |

```go
// 推荐实现
func RenderTemplate(templateStr string, data map[string]interface{}) (string, error) {
    tmpl, err := template.New("").Option("missingkey=error").Parse(templateStr)
    if err != nil {
        return "", fmt.Errorf("template parse error: %w", err)
    }
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", fmt.Errorf("template execute error: %w", err)
    }
    return buf.String(), nil
}
```

## 4. HTTP 客户端设计

**Decision**: 使用标准 net/http + 自定义 Transport

**Rationale**:
- 标准库足够满足需求
- 通过自定义 Transport 实现超时、重试、日志
- 避免引入第三方 HTTP 客户端库

**Features**:
- 连接超时、读写超时独立配置
- 指数退避重试
- 请求/响应日志（可配置级别）
- 支持 TLS 证书配置

## 5. Job 执行监控

**Decision**: Watch Job 状态 + 超时定时器

**Rationale**:
- Informer Watch 实时感知 Job 状态变化
- 超时定时器兜底，避免 Job 卡住

**State Mapping**:
| Job 状态       | Action 状态      |
| -------------- | ---------------- |
| Active > 0     | Running          |
| Succeeded >= 1 | Succeeded        |
| Failed >= 1    | Failed           |
| Timeout        | Failed (timeout) |

## 6. Clusternet Localization 集成

**Decision**: 直接创建/更新 Localization CR

**Rationale**:
- Localization 是 Clusternet 的标准 CRD
- 通过 client-go 直接操作，无需额外抽象
- 监控 Localization Status 判断下发结果

**Dependencies**:
- `github.com/clusternet/clusternet/pkg/apis/apps/v1alpha1`

**Auto-Rollback**:
- 记录创建的 Localization 名称到 outputs
- Revert 时删除该 Localization

## 7. 回滚策略设计

**Decision**: 步骤级回滚 + 自动逆操作

**Rationale**:
- 每个步骤可选定义 `rollback` 动作
- 未定义时使用类型特定的自动逆操作
- 按执行成功步骤的逆序回滚

**Auto-Rollback Mapping**:
| Action Type  | Auto-Rollback       |
| ------------ | ------------------- |
| Localization | Delete Localization |
| Job          | Delete Job          |
| HTTP         | Skip (no default)   |

**Implementation**:
```go
// 回滚决策
for i := len(succeededSteps) - 1; i >= 0; i-- {
    step := succeededSteps[i]
    if step.Rollback != nil {
        // 执行自定义回滚
        executeAction(step.Rollback)
    } else {
        // 自动逆操作
        switch step.Type {
        case "Localization":
            deleteLocalization(step.outputs.localizationRef)
        case "Job":
            deleteJob(step.outputs.jobRef)
        case "HTTP":
            // skip
        }
    }
}
```

## 8. 并发控制

**Decision**: 单 DRPlan 互斥执行 + 乐观锁

**Rationale**:
- 同一 DRPlan 不能并发执行（Execute/Revert 互斥）
- 使用 DRPlan.Status.currentExecution 记录当前执行
- 创建 Execution 时检查并发，冲突则拒绝

**Implementation**:
```go
// Webhook 验证
if plan.Status.CurrentExecution != "" {
    return admission.Denied("execution already in progress")
}
```

## 9. 断点恢复

**Decision**: 基于 DRPlanExecution.Status.ActionStatuses 恢复

**Rationale**:
- 每个 Action 执行状态实时更新到 Status
- Controller 重启后读取 Status，跳过已成功的 Action
- 从 Running/Pending 的 Action 继续执行

**Recovery Logic**:
```go
for _, action := range workflow.Actions {
    status := findActionStatus(execution, action.Name)
    if status != nil && status.Phase == "Succeeded" {
        continue // skip completed
    }
    executeAction(action)
}
```

## 10. 测试策略

**Decision**: 分层测试

| 层级        | 工具          | 覆盖范围                  |
| ----------- | ------------- | ------------------------- |
| Unit        | go test       | executor、utils、template |
| Integration | envtest       | controller reconcile      |
| E2E         | kind + ginkgo | 完整流程                  |

**Mock Strategy**:
- HTTP: httptest.Server
- Job: fake client-go
- Localization: fake client-go

## 11. 扩展性预留

**Decision**: 可插拔执行引擎设计，首版 Native 引擎，预留 Argo 扩展

### 11.1 执行引擎架构

```
┌─────────────────────────────────────────────────────────────────┐
│                   WorkflowExecutor Interface                    │
│                                                                 │
│  Execute(workflow, params) → error                              │
│  Revert(workflow, outputs) → error                              │
│  GetStatus() → ExecutionStatus                                  │
│  Cancel() → error                                               │
└───────────────────────────┬─────────────────────────────────────┘
                            │
            ┌───────────────┴───────────────┐
            ▼                               ▼
┌─────────────────────┐           ┌─────────────────────┐
│   NativeExecutor    │           │    ArgoExecutor     │
│   (Phase 1)         │           │    (Phase 3)        │
└─────────────────────┘           └─────────────────────┘
```

### 11.2 Native 引擎（首版）

```go
type NativeExecutor struct {
    client     client.Client
    httpClient *http.Client
}

func (e *NativeExecutor) Execute(workflow *DRWorkflow, params map[string]string) error {
    for _, action := range workflow.Spec.Actions {
        if err := e.executeAction(action, params); err != nil {
            return err
        }
    }
    return nil
}
```

### 11.3 Argo 引擎（扩展预留）

**转换策略**：DRWorkflow → Argo Workflow

| DRWorkflow           | Argo Workflow                                     |
| -------------------- | ------------------------------------------------- |
| `actions[]`          | `spec.templates[]`                                |
| `type: HTTP`         | `http` template                                   |
| `type: Job`          | `container` template                              |
| `type: Localization` | `resource` template (action: create/patch/delete) |
| `type: Subscription` | `resource` template                               |
| `parameters`         | `spec.arguments.parameters`                       |
| `rollback`           | 生成独立的 Revert Workflow                        |

**回滚处理**：
```
Execute 触发 → 生成 Argo Workflow (正向)
Revert 触发  → 生成 Argo Workflow (逆向，基于 outputs)
```

**状态同步**：
```go
type ArgoExecutor struct {
    argoClient argoclient.Interface
}

func (e *ArgoExecutor) syncStatus(argoWf *argoworkflow.Workflow, execution *DRPlanExecution) {
    for _, node := range argoWf.Status.Nodes {
        // 同步每个步骤的状态到 DRPlanExecution.actionStatuses
    }
}
```

### 11.4 DAG 扩展（Phase 2）

```yaml
spec:
  actions:
    - name: notify-start
      type: HTTP
      # 无依赖，最先执行
      
    - name: create-localization
      type: Localization
      dependsOn: [notify-start]  # 依赖 notify-start
      
    - name: create-subscription
      type: Subscription
      dependsOn: [notify-start]  # 与 create-localization 并行
      
    - name: run-sync-job
      type: Job
      dependsOn: [create-localization, create-subscription]  # 等待两者
      when: "{{ .outputs.create-localization.phase == 'Succeeded' }}"
```

**DAG 执行逻辑**：
1. 构建依赖图（拓扑排序）
2. 并行执行无依赖的动作
3. 动作完成后检查依赖它的动作是否可执行
4. 支持 `when` 条件判断

### 11.5 演进路线

| Phase | 功能                   | 引擎        | 复杂度   |
| ----- | ---------------------- | ----------- | -------- |
| **1** | 顺序执行、回滚、参数化 | Native      | ~1500 行 |
| **2** | DAG、条件执行、并行    | Native 扩展 | +~800 行 |
| **3** | Argo 集成、可视化      | Argo 适配器 | +~600 行 |

**Design Principle**: 
- 首版保持简单，验证核心流程
- 通过接口抽象支持多引擎
- 字段预留确保无 breaking change
