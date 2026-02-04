# Quick Start: 容灾策略 Controller

## 前置条件

- Kubernetes 1.20+ 集群
- kubectl 已配置
- （可选）Clusternet 已部署（如需使用 Localization 动作）

## 安装

### 1. 部署 CRD

```bash
kubectl apply -f config/crd/bases/
```

### 2. 部署 Controller

```bash
# 使用 Helm
helm install bcs-drplan-controller charts/bcs-drplan-controller \
  --namespace bcs-system \
  --create-namespace

# 或使用 kubectl
kubectl apply -f config/manager/
```

### 3. 验证安装

```bash
kubectl get pods -n bcs-system -l app=bcs-drplan-controller
kubectl get crd | grep dr.bkbcs.tencent.com
```

## 快速上手

### Step 1: 创建工作流

```yaml
# workflow.yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: simple-dr-workflow
spec:
  parameters:
    - name: targetCluster
      type: string
      required: true
    - name: notifyURL
      type: string
      default: "https://hooks.example.com/notify"

  failurePolicy: FailFast

  actions:
    - name: notify-start
      type: HTTP
      http:
        url: "{{ .params.notifyURL }}"
        method: POST
        body: '{"event": "dr-started", "cluster": "{{ .params.targetCluster }}"}'
      timeout: "30s"

    - name: switch-traffic
      type: HTTP
      http:
        url: "https://traffic-api.example.com/switch"
        method: POST
        body: '{"target": "{{ .params.targetCluster }}"}'
      rollback:
        type: HTTP
        http:
          url: "https://traffic-api.example.com/revert"
          method: POST

    - name: notify-complete
      type: HTTP
      http:
        url: "{{ .params.notifyURL }}"
        method: POST
        body: '{"event": "dr-completed", "cluster": "{{ .params.targetCluster }}"}'
```

```bash
kubectl apply -f workflow.yaml
kubectl get drworkflow simple-dr-workflow
```

### Step 2: 创建预案

```yaml
# plan.yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: prod-dr-plan
spec:
  description: "生产环境容灾预案"
  
  globalParams:
    - name: targetCluster
      value: "cluster-backup"
    - name: notifyURL
      value: "https://hooks.slack.com/services/xxx"
  
  stages:
    - name: execute-dr
      workflows:
        - workflowRef:
            name: simple-dr-workflow
            namespace: default
```

```bash
kubectl apply -f plan.yaml
kubectl get drplan prod-dr-plan
```

### Step 3: 触发执行

**方式 1: 创建 Execution CR**

```yaml
# execution.yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: prod-dr-exec-001
spec:
  planRef: prod-dr-plan
  operationType: Execute
```

```bash
kubectl apply -f execution.yaml
```

**方式 2: 通过 Annotation 触发**

```bash
kubectl annotate drplan prod-dr-plan dr.bkbcs.tencent.com/trigger=execute
```

### Step 4: 查看执行状态

```bash
# 查看执行记录
kubectl get drplanexecution -l dr.bkbcs.tencent.com/plan=prod-dr-plan

# 查看详细状态
kubectl describe drplanexecution prod-dr-exec-001

# 查看事件
kubectl get events --field-selector involvedObject.name=prod-dr-exec-001
```

### Step 5: 触发恢复

```bash
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: prod-dr-revert-001
spec:
  planRef: prod-dr-plan
  operationType: Revert
  # revertExecutionRef is REQUIRED
  revertExecutionRef: prod-dr-exec-001
EOF
```

### Step 6: 取消执行（可选）

```bash
# 添加 cancel annotation 到正在运行的 execution
kubectl annotate drplanexecution prod-dr-exec-001 dr.bkbcs.tencent.com/cancel=true

# 查看取消后的状态
kubectl get drplanexecution prod-dr-exec-001 -o jsonpath='{.status.phase}'
# 输出: Cancelled
```

## 常用操作

### 查看所有预案

```bash
kubectl get drplan -o wide
```

### 查看预案的执行历史

```bash
kubectl get drplanexecution -l dr.bkbcs.tencent.com/plan=<plan-name>
```

### 查看工作流定义

```bash
kubectl get drworkflow -o yaml
```

### 清理执行记录

```bash
# 删除特定执行记录
kubectl delete drplanexecution <execution-name>

# 删除某预案的所有执行记录
kubectl delete drplanexecution -l dr.bkbcs.tencent.com/plan=<plan-name>
```

## 故障排查

### 预案状态为 Invalid

```bash
kubectl describe drplan <plan-name>
# 查看 Conditions 中的错误信息
```

常见原因：
- 引用的 DRWorkflow 不存在
- 必填参数未提供

### 执行卡在 Running

```bash
kubectl describe drplanexecution <execution-name>
# 查看 actionStatuses 找到卡住的动作
```

检查：
- HTTP 动作的目标接口是否可达
- Job 是否有镜像拉取问题
- Localization 的目标集群是否正常

### 回滚失败

```bash
kubectl get drplanexecution <execution-name> -o yaml
# 查看 actionStatuses 中的 outputs 信息
```

确保：
- 自动回滚需要 outputs 中记录了创建的资源
- 自定义回滚的 HTTP 接口正常

## 进阶：多 Stage 并行编排

### 适用场景

当你需要容灾多个组件（数据库、缓存、多个服务）时，使用多 Stage 编排可以：
- 分阶段执行（如先切基础设施，再切应用）
- 并行执行提升速度（如多个服务同时切换）
- Workflow 复用（同一个 Workflow 用于不同服务）

### 创建多 Stage 编排预案

```yaml
# plan-with-stages.yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: multi-service-dr-plan
  namespace: production
spec:
  description: "多组件容灾预案 - Stage 编排"
  
  # 全局参数
  globalParams:
    - name: drClusterNamespace
      value: "clusternet-dr-cluster"
    - name: webhookURL
      value: "https://hooks.slack.com/services/xxx"
  
  # Stage 编排
  stages:
    # Stage 1: 通知开始
    - name: notify-start
      workflows:
        - workflowRef:
            name: simple-dr-workflow
            namespace: production
          params:
            - name: targetCluster
              value: "backup"
    
    # Stage 2: 基础设施（并行）
    - name: infrastructure
      dependsOn: [notify-start]
      parallel: true  # 并行执行
      workflows:
        - workflowRef:
            name: simple-dr-workflow
            namespace: production
          params:
            - name: targetCluster
              value: "db-backup"
        
        - workflowRef:
            name: simple-dr-workflow
            namespace: production
          params:
            - name: targetCluster
              value: "cache-backup"
    
    # Stage 3: 应用服务（并行）
    - name: applications
      dependsOn: [infrastructure]  # 依赖 infrastructure 完成
      parallel: true
      workflows:
        - workflowRef:
            name: simple-dr-workflow
            namespace: production
          params:
            - name: targetCluster
              value: "app1-backup"
        
        - workflowRef:
            name: simple-dr-workflow
            namespace: production
          params:
            - name: targetCluster
              value: "app2-backup"
```

```bash
kubectl apply -f plan-with-stages.yaml
```

### 触发和监控

```bash
# 触发执行
kubectl annotate drplan multi-service-dr-plan \
  -n production \
  dr.bkbcs.tencent.com/trigger=execute

# 查看 Stage 执行状态
EXEC_NAME=$(kubectl get drplanexecution -n production \
  -l dr.bkbcs.tencent.com/plan=multi-service-dr-plan \
  --sort-by=.metadata.creationTimestamp \
  -o jsonpath='{.items[-1].metadata.name}')

# 查看执行摘要
kubectl get drplanexecution $EXEC_NAME -n production \
  -o jsonpath='{.status.summary}' | jq

# 输出示例：
# {
#   "totalStages": 3,
#   "completedStages": 1,
#   "runningStages": 1,
#   "pendingStages": 1,
#   "totalWorkflows": 5,
#   "completedWorkflows": 1,
#   "runningWorkflows": 2,
#   "pendingWorkflows": 2
# }
```

### 单 Stage vs 多 Stage 对比

| 特性              | 单 Stage                 | 多 Stage 并行编排        |
| ----------------- | ------------------------ | ------------------------ |
| **适用场景**      | 单组件切换               | 多组件系统切换           |
| **并行能力**      | 无（顺序执行）           | Stage 内可并行           |
| **依赖管理**      | 无                       | Stage 间可定义依赖       |
| **Workflow 复用** | 一个 Stage 一个 Workflow | 一个 Workflow 可多次引用 |
| **可维护性**      | 简单直接                 | 结构清晰，易扩展         |

## 进阶：使用通用 K8s 资源操作

### 适用场景

当你需要操作专用类型未覆盖的资源（ConfigMap、Secret、自定义 CRD）时，可以使用 `KubernetesResource` 类型。

### 示例：创建 ConfigMap

```yaml
# workflow-with-configmap.yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: configmap-workflow
spec:
  parameters:
    - name: namespace
      type: string
      default: "production"
    - name: drDbHost
      type: string
      required: true
  
  actions:
    - name: create-dr-config
      type: KubernetesResource
      resource:
        operation: Create
        manifest: |
          apiVersion: v1
          kind: ConfigMap
          metadata:
            name: "dr-config"
            namespace: "{{ .params.namespace }}"
          data:
            dr-mode: "active"
            db-host: "{{ .params.drDbHost }}"
      # Create 自动回滚：删除创建的 ConfigMap
```

### 专用类型 vs 通用类型

| 特性 | 专用类型 (Job/Sub/Loc) | 通用类型 (KubernetesResource) |
|------|----------------------|-------------------------------|
| **类型安全** | ✅ 强类型 | ⚠️ 弱类型（字符串 manifest） |
| **用户体验** | ✅ 简洁直观 | ⚠️ 需要完整 YAML |
| **扩展性** | ⚠️ 需添加新类型 | ✅ 支持任意资源 |
| **自动回滚** | ✅ 智能回滚 | ⚠️ 部分需手动定义 |
| **推荐程度** | ✅✅ 优先使用 | ✅ 特殊场景使用 |

## 下一步

- 查看 [完整 API 文档](./contracts/)
- 了解 [数据模型](./data-model.md)
- 阅读 [完整用户指南](../../docs/user-guide.md)
