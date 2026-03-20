# DRPlan Controller 使用指南

本文档介绍如何使用 DRPlan Controller 创建、管理和执行容灾预案。

## 目录

- [概述](#概述)
- [前置条件](#前置条件)
- [快速开始](#快速开始)
- [创建工作流 (DRWorkflow)](#创建工作流-drworkflow)
- [创建预案 (DRPlan)](#创建预案-drplan)
- [触发预案执行](#触发预案执行)
- [观测执行状态](#观测执行状态)
- [执行回滚](#执行回滚)
- [取消执行](#取消执行)
- [完整示例场景](#完整示例场景)

---

## 概述

DRPlan Controller 是一个 Kubernetes Operator，用于定义和执行容灾预案。核心概念：

- **DRWorkflow**: 工作流定义，描述一系列动作的编排逻辑（可复用）
- **DRPlan**: 容灾预案，引用 DRWorkflow 并提供具体参数
- **DRPlanExecution**: 执行记录，记录每次执行的状态和结果

**支持的动作类型**:
- `HTTP`: 调用外部 HTTP 接口
- `Job`: 创建 Kubernetes Job
- `Localization`: 操作 Clusternet Localization 资源（Create/Patch/Delete）
- `Subscription`: 操作 Clusternet Subscription 资源（Create/Patch/Delete）

### 典型容灾切换工作流程

```
┌─────────────────────────────────────────────────────────────────────┐
│                     容灾预案执行流程（Execute）                          │
└─────────────────────────────────────────────────────────────────────┘

  触发方式：kubectl apply -f drplanexecution.yaml

           ↓

  ┌─────────────────────────────────────────────────────────────┐
  │ Step 1: HTTP 通知                                             │
  │  → 发送容灾开始通知到 Webhook                                   │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 2: Localization Create                                 │
  │  → 创建 Localization 定制化 DR 集群配置                         │
  │  → 覆盖副本数、环境变量（如 DR 数据库地址）、标签等                │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 3: Subscription Create                                 │
  │  → 创建 Subscription 将应用注册并分发到 DR 集群                  │
  │  → Clusternet 自动调度资源到匹配的子集群                         │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 4: Job 执行                                              │
  │  → 运行健康检查或数据同步 Job                                   │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 5: HTTP 通知                                             │
  │  → 发送容灾完成通知                                             │
  └─────────────────────────────────────────────────────────────┘

  执行结果：DRPlanExecution.status.phase = Succeeded

┌─────────────────────────────────────────────────────────────────────┐
│                     容灾恢复流程（Revert）                              │
└─────────────────────────────────────────────────────────────────────┘

  触发方式：kubectl apply -f drplanexecution-revert.yaml 
           (must specify revertExecutionRef)

           ↓

  ┌─────────────────────────────────────────────────────────────┐
  │ Step 5: 跳过 HTTP 通知（无法逆向）                               │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 4: 删除 Job（自动逆操作）                                   │
  │  → 清理健康检查 Job 资源                                         │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 3: 删除 Subscription（自动逆操作）                          │
  │  → 停止向 DR 集群分发资源                                        │
  │  → 已部署资源可能需要手动清理                                     │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 2: 删除 Localization（自动逆操作）                          │
  │  → 移除 DR 定制化配置                                            │
  │  → 应用恢复为 Feed 中的原始配置                                   │
  └─────────────────────────────────────────────────────────────┘
           ↓
  ┌─────────────────────────────────────────────────────────────┐
  │ Step 1: 跳过 HTTP 通知（无法逆向）                               │
  └─────────────────────────────────────────────────────────────┘

  恢复结果：DRPlanExecution.status.phase = Succeeded
           应用切回主集群正常服务

关键点：
  • Localization 回滚 = 删除 Localization CR → 自动恢复原始配置
  • 无需记录 primaryDbHost 等原始值，Feed 中的资源就是原始状态
  • Subscription 回滚 = 删除 Subscription CR → 停止资源分发
```

---

## 前置条件

1. Kubernetes 集群（v1.20+）
2. 已安装 DRPlan Controller
3. （可选）如果使用 Localization/Subscription 动作，需要安装 Clusternet

验证 Controller 是否运行：

```bash
kubectl get pods -n drplan-system
```

---

## 快速开始

### 1. 创建一个简单的工作流

创建文件 `simple-workflow.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: simple-notify
  namespace: default
spec:
  description: "简单的通知工作流"
  failurePolicy: Stop
  actions:
    - name: send-notification
      type: HTTP
      http:
        url: "https://notify.example.com/webhook"
        method: POST
        body: |
          {
            "message": "容灾预案已触发",
            "timestamp": "{{ .timestamp }}"
          }
        timeout: 30s
```

应用配置：

```bash
kubectl apply -f simple-workflow.yaml
```

验证工作流状态：

```bash
kubectl get drworkflow simple-notify -o yaml
# 查看 status.phase，应该为 Ready
```

### 2. 创建预案

创建文件 `simple-plan.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: my-first-plan
  namespace: default
spec:
  description: "我的第一个容灾预案"
  workflowRef:
    name: simple-notify
    namespace: default
```

应用配置：

```bash
kubectl apply -f simple-plan.yaml
```

### 3. 触发执行

创建文件 `execute-plan.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: my-first-execution
  namespace: default
spec:
  planRef:
    name: my-first-plan
    namespace: default
  operation: Execute
```

应用配置：

```bash
kubectl apply -f execute-plan.yaml
```

查看执行状态：

```bash
kubectl get drplanexecution my-first-execution -o yaml
```

---

## 创建工作流 (DRWorkflow)

### 基础结构

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: <workflow-name>
  namespace: <namespace>
spec:
  description: "工作流描述"
  
  # 参数定义（可选）
  parameters:
    - name: <param-name>
      type: string  # string | int | bool
      required: true
      default: "默认值"
      description: "参数说明"
  
  # 失败策略: Stop（默认）| Continue
  failurePolicy: Stop
  
  # 动作列表
  actions:
    - name: <action-name>
      type: <action-type>  # HTTP | Job | Localization | Subscription
      <action-specific-config>
      
      # 回滚动作（可选）
      rollback:
        type: <action-type>
        <action-specific-config>
```

### 示例 1: HTTP 动作

```yaml
actions:
  - name: notify-oncall
    type: HTTP
    http:
      url: "{{ .params.notifyURL }}"
      method: POST
      headers:
        Content-Type: "application/json"
        Authorization: "Bearer {{ .params.token }}"
      body: |
        {
          "event": "failover",
          "plan": "{{ .planName }}",
          "cluster": "{{ .params.targetCluster }}"
        }
      timeout: 30s
      expectedStatusCodes: [200, 201, 202]
    # HTTP 动作通常不需要 rollback（无法自动逆向）
```

### 示例 2: Job 动作

```yaml
actions:
  - name: backup-database
    type: Job
    job:
      namespace: "{{ .params.targetNamespace }}"
      generateName: "db-backup-"
      template:
        spec:
          restartPolicy: Never
          containers:
            - name: backup
              image: backup-tool:v1.0
              command: ["backup.sh"]
              args:
                - "--host={{ .params.dbHost }}"
                - "--database={{ .params.dbName }}"
      backoffLimit: 3
      timeout: 600s
    # Job 自动回滚：删除创建的 Job 资源
```

### 示例 3: Localization Create 动作

```yaml
actions:
  - name: deploy-app-to-dr-cluster
    type: Localization
    localization:
      operation: Create
      name: "app-{{ .planName }}"
      namespace: "{{ .params.drClusterNamespace }}"  # ManagedCluster 命名空间
      priority: 500  # 可选，优先级
      feed:  # 源资源引用
        apiVersion: apps/v1
        kind: Deployment
        name: my-app
        namespace: production
      overrides:
        - name: scale-replicas
          type: JSONPatch
          value: |
            - op: replace
              path: /spec/replicas
              value: 3
        - name: add-dr-label
          type: MergePatch
          value: |
            metadata:
              labels:
                dr-mode: "active"
    # Localization Create 自动回滚：删除创建的 Localization 资源
```

### 示例 4: Localization Patch 动作（必须定义 rollback）

```yaml
actions:
  - name: update-replicas
    type: Localization
    localization:
      operation: Patch
      name: existing-app-localization
      namespace: "{{ .params.drClusterNamespace }}"  # ManagedCluster 命名空间
      overrides:
        - name: scale-up
          type: JSONPatch
          value: |
            - op: replace
              path: /spec/replicas
              value: {{ .params.replicas }}
    # Patch 必须显式定义回滚
    rollback:
      type: Localization
      localization:
        operation: Patch
        name: existing-app-localization
        namespace: "{{ .params.drClusterNamespace }}"
        overrides:
          - name: scale-down
            type: JSONPatch
            value: |
              - op: replace
                path: /spec/replicas
                value: 1
```

### 示例 5: Subscription Create 动作

```yaml
actions:
  - name: distribute-to-dr-cluster
    type: Subscription
    subscription:
      operation: Create
      name: "app-dr-subscription-{{ .planName }}"
      namespace: default  # Subscription CR 的命名空间
      schedulingStrategy: Replication  # Replication | Dividing
      feeds:  # 要分发的资源列表
        - apiVersion: apps/v1
          kind: Deployment
          name: my-app
          namespace: production
        - apiVersion: v1
          kind: ConfigMap
          name: my-app-config
          namespace: production
        - apiVersion: v1
          kind: Service
          name: my-app-svc
          namespace: production
      subscribers:  # 目标集群选择
        - clusterAffinity:
            matchLabels:
              region: "{{ .params.targetRegion }}"
              purpose: disaster-recovery
    # Subscription Create 自动回滚：删除创建的 Subscription 资源
```

**说明**：
- `feeds` 定义要分发的资源列表，Clusternet 会将这些资源从 Hub 集群同步到子集群
- `subscribers` 通过标签选择器匹配目标集群
- `schedulingStrategy` 定义调度策略（Replication：全量复制，Dividing：按权重分发）

### 示例 6: KubernetesResource 通用资源操作

```yaml
actions:
  # 场景 1: 创建 ConfigMap
  - name: create-dr-config
    type: KubernetesResource
    resource:
      operation: Create
      manifest: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: "dr-config-{{ .planName }}"
          namespace: "{{ .params.namespace }}"
        data:
          dr-mode: "active"
          dr-region: "{{ .params.targetRegion }}"
          db-host: "{{ .params.drDbHost }}"
    # Create 自动回滚：删除创建的 ConfigMap
  
  # 场景 2: 创建 Secret
  - name: create-dr-secret
    type: KubernetesResource
    resource:
      operation: Create
      manifest: |
        apiVersion: v1
        kind: Secret
        metadata:
          name: "dr-credentials"
          namespace: "{{ .params.namespace }}"
        type: Opaque
        stringData:
          username: "{{ .params.drUsername }}"
          password: "{{ .params.drPassword }}"
    # Create 自动回滚：删除创建的 Secret
  
  # 场景 3: 操作自定义 CRD
  - name: create-custom-resource
    type: KubernetesResource
    resource:
      operation: Apply  # 使用 Apply（幂等）
      manifest: |
        apiVersion: custom.example.com/v1
        kind: BackupPolicy
        metadata:
          name: "dr-backup-policy"
          namespace: "{{ .params.namespace }}"
        spec:
          schedule: "0 */6 * * *"
          retention: 7
          destination: "{{ .params.drBackupLocation }}"
    # Apply 通常用于幂等操作，回滚时需要显式定义
    rollback:
      type: KubernetesResource
      resource:
        operation: Delete
        manifest: |
          apiVersion: custom.example.com/v1
          kind: BackupPolicy
          metadata:
            name: "dr-backup-policy"
            namespace: "{{ .params.namespace }}"
  
  # 场景 4: Patch 操作（必须定义 rollback）
  - name: update-existing-config
    type: KubernetesResource
    resource:
      operation: Patch
      manifest: |
        apiVersion: v1
        kind: ConfigMap
        metadata:
          name: "app-config"
          namespace: "{{ .params.namespace }}"
        data:
          mode: "dr"  # 更新为 DR 模式
    # Patch 必须显式定义回滚
    rollback:
      type: KubernetesResource
      resource:
        operation: Patch
        manifest: |
          apiVersion: v1
          kind: ConfigMap
          metadata:
            name: "app-config"
            namespace: "{{ .params.namespace }}"
          data:
            mode: "normal"  # 恢复为 normal 模式
```

**KubernetesResource 类型说明**：
- **适用场景**：操作专用类型未覆盖的 K8s 资源（ConfigMap、Secret、自定义 CRD 等）
- **operation 类型**：
  - `Create`：创建资源，自动回滚为删除
  - `Apply`：幂等操作（server-side apply），需要显式定义回滚
  - `Patch`：更新资源，**必须**显式定义回滚
  - `Delete`：删除资源，无自动回滚
- **优势**：无需修改 Controller 即可支持任意 K8s 资源
- **劣势**：需要编写完整的 manifest，类型安全性较弱

**使用建议**：
- ✅ 优先使用专用类型（Job、Localization、Subscription）获得更好的用户体验
- ✅ 对于 ConfigMap、Secret、自定义 CRD 等特殊资源，使用 KubernetesResource
- ✅ 复杂场景可以混用专用类型和通用类型

### 参数化工作流

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: parameterized-workflow
spec:
  description: "带参数的工作流示例"
  
  parameters:
    - name: targetCluster
      type: string
      required: true
      description: "目标集群命名空间"
    
    - name: replicas
      type: int
      required: true
      default: 3
      description: "副本数"
    
    - name: notifyURL
      type: string
      required: false
      default: "https://default.notify.com"
  
  actions:
    - name: scale-app
      type: Localization
      localization:
        operation: Create
        name: "scaled-app"
        clusterNamespace: "{{ .params.targetCluster }}"
        overrides:
          - path: /spec/replicas
            value: {{ .params.replicas }}
```

---

## 创建预案 (DRPlan)

DRPlan 通过 Stage 编排多个 Workflow，支持并行执行和依赖管理。

### 基础结构

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: <plan-name>
  namespace: <namespace>
spec:
  description: "预案描述"
  
  # 全局参数（所有 Workflow 共享）
  globalParams:
    - name: <param-name>
      value: "<param-value>"
  
  # 全局失败策略
  failurePolicy: Stop  # Stop | Continue
  
  # Stage 编排
  stages:
    - name: <stage-name>
      description: "Stage 描述"
      dependsOn: [<dependency-stage-name>]  # 可选，依赖的 Stage
      parallel: true  # 可选，是否并行执行本 Stage 内的 Workflow
      workflows:
        - workflowRef:
            name: <workflow-name>
            namespace: <namespace>
          params:  # 可选，覆盖 globalParams
            - name: <param-name>
              value: "<param-value>"
```

**Stage 编排优势**：
- ✅ **并行执行**：Stage 内的多个 Workflow 可并行执行，大幅缩短总耗时
- ✅ **依赖管理**：通过 `dependsOn` 定义 Stage 间的执行顺序
- ✅ **Workflow 复用**：同一个 Workflow 可在不同 Stage 中复用
- ✅ **结构清晰**：按阶段组织，易于理解和维护
- ✅ **灵活编排**：支持复杂的拓扑依赖关系

### 示例 1: 单 Stage 预案

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: failover-to-backup
  namespace: production
spec:
  description: "故障切换到备用集群"
  
  globalParams:
    - name: targetCluster
      value: "backup-cluster-01"
    - name: notifyURL
      value: "https://oncall.example.com/webhook"
  
  stages:
    - name: execute-failover
      workflows:
        - workflowRef:
            name: cluster-failover-workflow
            namespace: production
```

### 示例 2: 多参数预案

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: database-dr-plan
  namespace: production
spec:
  description: "数据库容灾预案"
  
  globalParams:
    - name: drClusterNamespace
      value: "dr-cluster-shanghai"
    - name: drDbHost
      value: "dr-db.example.com"
    - name: replicas
      value: "5"
    - name: backupEnabled
      value: "true"
  
  stages:
    - name: execute-db-failover
      workflows:
        - workflowRef:
            name: db-failover-workflow
            namespace: production
```

### 示例 3: Stage 编排预案

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: multi-service-dr-plan
  namespace: production
spec:
  description: "多服务容灾预案 - Stage 编排"
  
  # 全局参数
  globalParams:
    - name: drClusterNamespace
      value: "clusternet-dr-shanghai"
    - name: targetRegion
      value: "dr-region"
  
  failurePolicy: Stop
  
  # Stage 编排
  stages:
    # Stage 1: 基础设施（并行）
    - name: infrastructure
      description: "数据库和缓存切换"
      parallel: true
      workflows:
        - workflowRef:
            name: mysql-failover
            namespace: production
          params:
            - name: drDbHost
              value: "dr-mysql.example.com"
        
        - workflowRef:
            name: redis-failover
            namespace: production
          params:
            - name: drRedisHost
              value: "dr-redis.example.com"
    
    # Stage 2: 应用服务（并行，依赖 Stage 1）
    - name: applications
      description: "业务服务切换"
      dependsOn: [infrastructure]
      parallel: true
      workflows:
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "order-service"
            - name: replicas
              value: "5"
        
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "payment-service"
            - name: replicas
              value: "3"
    
    # Stage 3: 验证（顺序，依赖 Stage 2）
    - name: validation
      description: "健康检查"
      dependsOn: [applications]
      parallel: false
      workflows:
        - workflowRef:
            name: health-check-workflow
            namespace: production
```

### 验证预案状态

```bash
# 查看预案列表
kubectl get drplan -n production

# 查看预案详情
kubectl get drplan failover-to-backup -n production -o yaml

# 查看预案状态
kubectl get drplan failover-to-backup -n production -o jsonpath='{.status.phase}'
# 输出: Ready / Invalid / Executed
```

---

## 触发预案执行

有两种方式触发预案执行：

### 方式 1: 创建 DRPlanExecution CR（推荐）

创建文件 `execute.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: failover-execution-001
  namespace: production
spec:
  planRef:
    name: failover-to-backup
    namespace: production
  operation: Execute  # Execute | Revert
```

执行：

```bash
kubectl apply -f execute.yaml
```

### 方式 2: 使用 Annotation 触发

直接在 DRPlan 上添加 annotation：

```bash
# 触发执行
kubectl annotate drplan failover-to-backup \
  -n production \
  dr.bkbcs.tencent.com/trigger=execute

# 系统会自动创建 DRPlanExecution CR
# 执行完成后 annotation 会被自动清除
```

### 并发控制

同一个 DRPlan 同时只能有一个执行在进行中：

```bash
# 如果已有执行在运行中，再次触发会被拒绝
kubectl annotate drplan failover-to-backup \
  -n production \
  dr.bkbcs.tencent.com/trigger=execute

# 错误示例输出：
# Error: Plan has an execution in progress: failover-execution-001
```

---

## 观测执行状态

### 查看执行列表

```bash
# 查看所有执行记录
kubectl get drplanexecution -n production

# 输出示例：
# NAME                      PLAN                  OPERATION   PHASE       AGE
# failover-execution-001    failover-to-backup    Execute     Running     30s
# failover-execution-002    failover-to-backup    Revert      Succeeded   5m
```

### 查看执行详情

```bash
kubectl get drplanexecution failover-execution-001 -n production -o yaml
```

**关键字段说明**：

```yaml
status:
  phase: Running  # Pending | Running | Succeeded | Failed | Cancelled
  
  # 参数快照
  resolvedParams:
    targetCluster: "backup-cluster-01"
    notifyURL: "https://oncall.example.com/webhook"
  
  # 执行时间
  startTime: "2026-01-30T10:00:00Z"
  completionTime: "2026-01-30T10:05:30Z"
  
  # 各步骤状态
  actionStatuses:
    - name: notify-oncall
      phase: Succeeded  # Pending | Running | Succeeded | Failed | Skipped
      startTime: "2026-01-30T10:00:00Z"
      completionTime: "2026-01-30T10:00:02Z"
      message: "HTTP request completed with status 200"
    
    - name: scale-app
      phase: Running
      startTime: "2026-01-30T10:00:02Z"
      message: "Creating Localization resource..."
    
    - name: update-db-config
      phase: Pending
      message: "Waiting for previous action to complete"
  
  # 创建的资源引用
  outputs:
    - actionName: scale-app
      localizationRef:
        name: scaled-app
        namespace: default
  
  # 错误信息（如果失败）
  message: "Action 'update-db-config' failed: timeout waiting for Job completion"
```

### 实时监控执行进度

```bash
# 使用 watch 实时查看状态
watch kubectl get drplanexecution failover-execution-001 -n production -o yaml

# 或使用 kubectl wait 等待完成
kubectl wait drplanexecution failover-execution-001 \
  -n production \
  --for=condition=Complete \
  --timeout=600s
```

### 查看 Kubernetes Events

```bash
# 查看执行相关的事件
kubectl describe drplanexecution failover-execution-001 -n production

# 输出示例：
# Events:
#   Type    Reason             Age   Message
#   ----    ------             ----  -------
#   Normal  ExecutionStarted   2m    Execution started for plan failover-to-backup
#   Normal  ActionStarted      2m    Action notify-oncall started
#   Normal  ActionSucceeded    2m    Action notify-oncall completed successfully in 2.1s
#   Normal  ActionStarted      2m    Action scale-app started
#   Normal  ActionSucceeded    1m    Action scale-app completed successfully in 30.5s
#   Normal  ExecutionSucceeded 1m    Execution completed successfully in 45.2s
```

### 查看 DRPlan 当前执行状态

```bash
kubectl get drplan failover-to-backup -n production -o yaml
```

```yaml
status:
  phase: Executed  # Ready | Executed | Invalid
  
  # 当前正在进行的执行（如果有）
  currentExecution:
    name: failover-execution-001
    namespace: production
    operation: Execute
    startTime: "2026-01-30T10:00:00Z"
  
  # 最后一次成功执行
  lastSuccessfulExecution:
    name: failover-execution-002
    namespace: production
    completionTime: "2026-01-30T09:00:00Z"
```

---

## 执行回滚

### 触发回滚的前提条件

1. DRPlan 状态必须为 `Executed`（已执行过）
2. 没有其他执行正在进行中

### 方式 1: 创建 Revert 执行

创建文件 `revert.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: failover-revert-001
  namespace: production
spec:
  planRef:
    name: failover-to-backup
    namespace: production
  operation: Revert  # 回滚操作
```

执行：

```bash
kubectl apply -f revert.yaml
```

### 方式 2: 使用 Annotation 触发回滚

```bash
kubectl annotate drplan failover-to-backup \
  -n production \
  dr.bkbcs.tencent.com/trigger=revert
```

### 回滚执行逻辑

系统会按照**逆序**执行工作流中定义的回滚动作：

| 原动作                             | 回滚行为                         |
| ---------------------------------- | -------------------------------- |
| HTTP 动作（无 rollback）           | 跳过（无法自动逆向）             |
| Job 动作（无 rollback）            | 自动删除创建的 Job 资源          |
| Localization Create（无 rollback） | 自动删除创建的 Localization 资源 |
| Localization Patch（有 rollback）  | 执行定义的回滚动作               |
| Localization Delete（无 rollback） | 跳过（无法恢复）                 |
| Subscription Create（无 rollback） | 自动删除创建的 Subscription 资源 |

**示例**：

原执行顺序：
1. `notify-oncall` (HTTP)
2. `create-localization` (Localization Create)
3. `update-config` (Localization Patch with rollback)

回滚执行顺序：
1. `update-config` → 执行定义的 rollback 动作
2. `create-localization` → 自动删除 Localization 资源
3. `notify-oncall` → 跳过（HTTP 无自动逆向）

---

## 取消执行

### 取消正在运行的执行

**方式 1**: 直接在 DRPlanExecution 上添加 annotation：

```bash
kubectl annotate drplanexecution failover-execution-001 \
  -n production \
  dr.bkbcs.tencent.com/cancel=true
```

**方式 2**: 在 DRPlan 上触发取消（会取消当前执行）：

```bash
kubectl annotate drplan failover-to-backup \
  -n production \
  dr.bkbcs.tencent.com/trigger=cancel
```

### 取消执行的行为

1. **停止后续动作**：尚未执行的动作不会再执行，状态保持 `Pending`
2. **等待当前动作**：正在运行的动作会等待其完成或超时
3. **更新状态**：执行状态变为 `Cancelled`
4. **释放锁定**：清空 `DRPlan.status.currentExecution`，允许新执行

### 限制条件

- 只能取消状态为 `Pending` 或 `Running` 的执行
- 已完成的执行（`Succeeded`/`Failed`/`Cancelled`）无法再次取消

### 验证取消结果

```bash
kubectl get drplanexecution failover-execution-001 -n production -o yaml
```

```yaml
status:
  phase: Cancelled
  message: "Execution cancelled by user"
  actionStatuses:
    - name: notify-oncall
      phase: Succeeded  # 已完成的动作
    - name: scale-app
      phase: Running    # 取消时正在运行，等待完成
    - name: update-db-config
      phase: Pending    # 未执行的动作保持 Pending
```

---

## 完整示例场景

### 场景 1：单应用容灾切换

**目标**：将生产应用从主集群切换到 DR 集群，包括：
1. 发送通知
2. 创建 Localization 定制化 DR 集群的配置（副本数、环境变量等）
3. 创建 Subscription 将应用分发到 DR 集群
4. 等待健康检查

**架构说明**：

在 Clusternet 中，资源分发和定制化是两个独立的概念：

- **Subscription**：负责将资源从 Hub 集群分发到 ManagedCluster（子集群）
  - 定义 `feeds`：要分发的资源列表（Deployment、ConfigMap 等）
  - 定义 `subscribers`：目标集群选择规则（通过标签匹配）
  - Subscription 创建后，Clusternet 会自动将资源调度到匹配的集群

- **Localization**：负责定制化已分发的资源
  - 针对特定 ManagedCluster 命名空间
  - 通过 `overrides` 修改资源配置（副本数、环境变量、标签等）
  - 支持 JSONPatch、MergePatch、Helm 等多种覆盖方式

**容灾流程**：
1. 先创建 **Localization** → 定制化 DR 集群的配置（如使用 DR 数据库地址）
2. 再创建 **Subscription** → 将应用注册并分发到 DR 集群
3. Clusternet 自动完成资源调度和配置应用

**回滚机制**（逆序执行）：
- **Subscription Create 回滚** = 删除 Subscription CR（先执行）
  - 停止向 DR 集群分发资源
  - 已部署的资源不会自动删除，由 Clusternet 的 GC 策略决定
- **Localization Create 回滚** = 删除 Localization CR（后执行）
  - 删除后，Clusternet 会使 Feed 中的原始资源配置生效
  - 例如：Localization 中覆盖 `DB_HOST=dr-mysql`，删除后自动恢复为原始的 `DB_HOST=primary-mysql`
  - **无需额外记录原始值**，Feed 就是原始配置的来源

**回滚顺序说明**：
```
正向执行: Localization Create → Subscription Create
回滚执行: Subscription Delete → Localization Delete (逆序)
```

#### 步骤 1: 创建工作流

`app-failover-workflow.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: app-failover
  namespace: production
spec:
  description: "应用容灾切换工作流"
  
  parameters:
    - name: drClusterNamespace
      type: string
      required: true
      description: "DR 集群命名空间"
    
    - name: drDbHost
      type: string
      required: true
      description: "DR 数据库地址"
    
    - name: replicas
      type: int
      required: true
      default: 3
      description: "应用副本数"
    
    - name: webhookURL
      type: string
      required: false
      default: "https://notify.example.com/webhook"
  
  failurePolicy: Stop
  
  actions:
    # 步骤 1: 发送通知
    - name: notify-start
      type: HTTP
      http:
        url: "{{ .params.webhookURL }}"
        method: POST
        headers:
          Content-Type: "application/json"
        body: |
          {
            "event": "failover_started",
            "plan": "{{ .planName }}",
            "target_cluster": "{{ .params.drClusterNamespace }}",
            "timestamp": "{{ .timestamp }}"
          }
        timeout: 30s
    
    # 步骤 2: 创建 Localization 定制化 DR 集群配置
    - name: customize-dr-deployment
      type: Localization
      localization:
        operation: Create
        name: "app-dr-localization-{{ .planName }}"
        namespace: "{{ .params.drClusterNamespace }}"  # ManagedCluster 命名空间
        priority: 600  # 高优先级确保覆盖生效
        feed:  # 关联到 Deployment 资源
          apiVersion: apps/v1
          kind: Deployment
          name: my-app
          namespace: production
        overrides:
          - name: scale-replicas
            type: JSONPatch
            value: |
              - op: replace
                path: /spec/replicas
                value: {{ .params.replicas }}
          - name: update-db-config
            type: JSONPatch
            value: |
              - op: replace
                path: /spec/template/spec/containers/0/env
                value:
                  - name: DB_HOST
                    value: "{{ .params.drDbHost }}"
                  - name: DR_MODE
                    value: "active"
                  - name: REGION
                    value: "dr-region"
          - name: add-dr-labels
            type: MergePatch
            value: |
              metadata:
                labels:
                  dr-mode: "active"
                  failover-plan: "{{ .planName }}"
      # Localization Create 自动回滚：删除创建的 Localization
    
    # 步骤 3: 创建 Subscription 将应用分发到 DR 集群
    - name: subscribe-app-to-dr
      type: Subscription
      subscription:
        operation: Create
        name: "app-dr-subscription-{{ .planName }}"
        namespace: default  # Subscription CR 的命名空间
        schedulingStrategy: Replication
        feeds:  # 要分发的资源列表
          - apiVersion: apps/v1
            kind: Deployment
            name: my-app
            namespace: production
          - apiVersion: v1
            kind: ConfigMap
            name: my-app-config
            namespace: production
        subscribers:  # 目标集群选择
          - clusterAffinity:
              matchLabels:
                region: dr-region
                purpose: disaster-recovery
      # Subscription Create 自动回滚：删除创建的 Subscription
    
    # 步骤 4: 执行健康检查 Job
    - name: health-check
      type: Job
      job:
        namespace: production
        generateName: "health-check-"
        template:
          spec:
            restartPolicy: Never
            containers:
              - name: checker
                image: curlimages/curl:latest
                command: ["/bin/sh", "-c"]
                args:
                  - |
                    for i in {1..10}; do
                      curl -f http://my-app.production.svc/health && exit 0
                      sleep 10
                    done
                    exit 1
        backoffLimit: 3
        timeout: 300s
    
    # 步骤 5: 发送完成通知
    - name: notify-complete
      type: HTTP
      http:
        url: "{{ .params.webhookURL }}"
        method: POST
        headers:
          Content-Type: "application/json"
        body: |
          {
            "event": "failover_completed",
            "plan": "{{ .planName }}",
            "status": "success",
            "timestamp": "{{ .timestamp }}"
          }
        timeout: 30s
```

应用工作流：

```bash
kubectl apply -f app-failover-workflow.yaml
```

#### 步骤 2: 创建预案

`app-failover-plan.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: app-dr-plan
  namespace: production
spec:
  description: "应用容灾切换预案 - 主机房故障时使用"
  
  # 全局参数
  globalParams:
    - name: drClusterNamespace
      value: "clusternet-dr-shanghai"
    - name: drDbHost
      value: "dr-mysql.shanghai.example.com"
    - name: replicas
      value: "5"
    - name: webhookURL
      value: "https://oncall.example.com/drplan-webhook"
  
  # Stage 编排
  stages:
    - name: execute-failover
      workflows:
        - workflowRef:
            name: app-failover
            namespace: production
```

应用预案：

```bash
kubectl apply -f app-failover-plan.yaml

# 验证预案状态
kubectl get drplan app-dr-plan -n production
# 输出: NAME           STAGES   PHASE   AGE
#       app-dr-plan    1        Ready   5s
```

#### 步骤 3: 触发容灾切换

```bash
# 方式 1: 创建执行 CR
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: app-dr-exec-$(date +%Y%m%d-%H%M%S)
  namespace: production
spec:
  planRef:
    name: app-dr-plan
    namespace: production
  operation: Execute
EOF

# 或方式 2: 使用 annotation
kubectl annotate drplan app-dr-plan \
  -n production \
  dr.bkbcs.tencent.com/trigger=execute
```

#### 步骤 4: 监控执行进度

```bash
# 查看执行状态
EXEC_NAME=$(kubectl get drplanexecution -n production \
  -l dr.bkbcs.tencent.com/plan=app-dr-plan \
  --sort-by=.metadata.creationTimestamp \
  -o jsonpath='{.items[-1].metadata.name}')

echo "监控执行: $EXEC_NAME"

# 实时监控
watch "kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.phase}' && \
  echo && \
  kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.actionStatuses[*].name}' && \
  echo && \
  kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.actionStatuses[*].phase}'"

# 查看详细事件
kubectl describe drplanexecution $EXEC_NAME -n production
```

#### 步骤 5: 验证切换结果

```bash
# 验证 Subscription 是否创建成功
kubectl get subscription -n default | grep app-dr-subscription

# 验证 Localization 是否创建并应用
kubectl get localization -n <dr-cluster-namespace>

# 验证应用是否在 DR 集群部署（通过 Clusternet 查看）
kubectl get base -A | grep my-app

# 查看完整执行结果
kubectl get drplanexecution $EXEC_NAME -n production -o yaml
```

#### 步骤 6: 执行回滚（切回主集群）

当主机房恢复后，执行回滚：

```bash
# 触发回滚
kubectl annotate drplan app-dr-plan \
  -n production \
  dr.bkbcs.tencent.com/trigger=revert

# 监控回滚进度
REVERT_EXEC=$(kubectl get drplanexecution -n production \
  -l dr.bkbcs.tencent.com/plan=app-dr-plan \
  --sort-by=.metadata.creationTimestamp \
  -o jsonpath='{.items[-1].metadata.name}')

kubectl get drplanexecution $REVERT_EXEC -n production -w
```

**回滚执行顺序**（逆序）：
1. `notify-complete` → 跳过（HTTP 无法逆向）
2. `health-check` → 自动删除 Job
3. `subscribe-app-to-dr` → 自动删除 Subscription（停止向 DR 集群分发）
4. `customize-dr-deployment` → 自动删除 Localization（移除 DR 定制化配置）
5. `notify-start` → 跳过（HTTP 无法逆向）

**效果**：回滚后，应用将不再分发到 DR 集群，主集群恢复正常服务

---

### 场景 2：多组件系统容灾切换（Stage 编排模式）

**目标**：将包含多个组件的电商系统从主集群切换到 DR 集群，包括：
- 基础设施层：MySQL、Redis、Kafka（并行切换）
- 应用层：订单服务、支付服务、用户服务、前端（并行切换）
- 验证层：冒烟测试、健康检查（顺序执行）

**架构说明**：

使用 **Stage 编排模式**，将容灾切换分为 5 个阶段：

```
Stage 1: notify-start (顺序)
  └── notification-workflow

Stage 2: infrastructure (并行，依赖 Stage 1)
  ├── mysql-failover
  ├── redis-failover
  └── kafka-failover

Stage 3: applications (并行，依赖 Stage 2)
  ├── order-service-failover
  ├── payment-service-failover
  ├── user-service-failover
  └── frontend-failover

Stage 4: validation (顺序，依赖 Stage 3)
  ├── smoke-test-workflow
  └── health-check-workflow

Stage 5: notify-complete (顺序，依赖 Stage 4)
  └── notification-workflow
```

#### 步骤 1: 创建可复用的 Workflow

**1.1 通知 Workflow**

`notification-workflow.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: notification-workflow
  namespace: production
spec:
  description: "通知工作流"
  parameters:
    - name: webhookURL
      type: string
      required: true
    - name: message
      type: string
      required: true
    - name: event
      type: string
      required: false
  
  actions:
    - name: send-notification
      type: HTTP
      http:
        url: "{{ .params.webhookURL }}"
        method: POST
        headers:
          Content-Type: "application/json"
        body: |
          {
            "event": "{{ .params.event }}",
            "message": "{{ .params.message }}",
            "timestamp": "{{ .timestamp }}"
          }
        timeout: 30s
```

**1.2 MySQL 容灾 Workflow**

`mysql-failover.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: mysql-failover
  namespace: production
spec:
  description: "MySQL 容灾切换"
  parameters:
    - name: drDbHost
      type: string
      required: true
    - name: drDbPort
      type: string
      default: "3306"
  
  actions:
    - name: switch-mysql
      type: HTTP
      http:
        url: "https://db-admin-api.example.com/failover"
        method: POST
        body: |
          {
            "service": "mysql",
            "target_host": "{{ .params.drDbHost }}",
            "target_port": "{{ .params.drDbPort }}"
          }
        timeout: 60s
      rollback:
        type: HTTP
        http:
          url: "https://db-admin-api.example.com/revert"
          method: POST
          body: '{"service": "mysql"}'
```

**1.3 服务容灾 Workflow（通用模板）**

`service-failover.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: service-failover
  namespace: production
spec:
  description: "通用服务容灾切换"
  parameters:
    - name: serviceName
      type: string
      required: true
    - name: drClusterNamespace
      type: string
      required: true
    - name: replicas
      type: int
      default: 3
  
  actions:
    - name: create-localization
      type: Localization
      localization:
        operation: Create
        name: "{{ .params.serviceName }}-localization"
        namespace: "{{ .params.drClusterNamespace }}"
        feed:
          apiVersion: apps/v1
          kind: Deployment
          name: "{{ .params.serviceName }}"
          namespace: production
        overrides:
          - name: scale-replicas
            type: JSONPatch
            value: |
              - op: replace
                path: /spec/replicas
                value: {{ .params.replicas }}
          - name: add-dr-label
            type: MergePatch
            value: |
              metadata:
                labels:
                  dr-mode: "active"
    
    - name: create-subscription
      type: Subscription
      subscription:
        operation: Create
        name: "{{ .params.serviceName }}-subscription"
        namespace: default
        schedulingStrategy: Replication
        feeds:
          - apiVersion: apps/v1
            kind: Deployment
            name: "{{ .params.serviceName }}"
            namespace: production
          - apiVersion: v1
            kind: Service
            name: "{{ .params.serviceName }}"
            namespace: production
        subscribers:
          - clusterAffinity:
              matchLabels:
                purpose: disaster-recovery
    
    - name: health-check
      type: Job
      job:
        namespace: production
        generateName: "health-{{ .params.serviceName }}-"
        template:
          spec:
            restartPolicy: Never
            containers:
              - name: checker
                image: curlimages/curl:latest
                command: ["/bin/sh", "-c"]
                args:
                  - |
                    curl -f http://{{ .params.serviceName }}.production.svc/health && exit 0 || exit 1
        timeout: 300s
```

应用所有 Workflow：

```bash
kubectl apply -f notification-workflow.yaml
kubectl apply -f mysql-failover.yaml
kubectl apply -f service-failover.yaml

# 验证 Workflow 状态
kubectl get drworkflow -n production
```

#### 步骤 2: 创建 Stage 编排预案

`ecommerce-dr-plan-with-stages.yaml`：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlan
metadata:
  name: ecommerce-dr-plan
  namespace: production
spec:
  description: "电商系统完整容灾预案 - 分阶段并行执行"
  
  # 全局参数（所有 Workflow 共享）
  globalParams:
    - name: drClusterNamespace
      value: "clusternet-dr-shanghai"
    - name: targetRegion
      value: "dr-region"
    - name: webhookURL
      value: "https://oncall.example.com/drplan-webhook"
  
  # 全局失败策略
  failurePolicy: Stop
  
  # Stage 编排
  stages:
    # 阶段 1: 发送开始通知
    - name: notify-start
      description: "发送容灾开始通知"
      parallel: false
      workflows:
        - workflowRef:
            name: notification-workflow
            namespace: production
          params:
            - name: message
              value: "电商系统容灾切换开始"
            - name: event
              value: "failover-start"
    
    # 阶段 2: 基础设施切换（并行）
    - name: infrastructure
      description: "数据库、缓存、消息队列切换"
      dependsOn: [notify-start]
      parallel: true  # 并行执行
      workflows:
        - workflowRef:
            name: mysql-failover
            namespace: production
          params:
            - name: drDbHost
              value: "dr-mysql.shanghai.example.com"
            - name: drDbPort
              value: "3306"
        
        - workflowRef:
            name: mysql-failover  # 复用 Workflow，不同参数
            namespace: production
          params:
            - name: drDbHost
              value: "dr-redis.shanghai.example.com"
            - name: drDbPort
              value: "6379"
        
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "kafka"
            - name: replicas
              value: "3"
    
    # 阶段 3: 应用服务切换（并行）
    - name: applications
      description: "业务服务切换"
      dependsOn: [infrastructure]  # 依赖基础设施完成
      parallel: true
      workflows:
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "order-service"
            - name: replicas
              value: "5"
        
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "payment-service"
            - name: replicas
              value: "3"
        
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "user-service"
            - name: replicas
              value: "3"
        
        - workflowRef:
            name: service-failover
            namespace: production
          params:
            - name: serviceName
              value: "frontend"
            - name: replicas
              value: "2"
    
    # 阶段 4: 验证（顺序执行）
    - name: validation
      description: "验证服务可用性"
      dependsOn: [applications]
      parallel: false  # 顺序执行
      workflows:
        - workflowRef:
            name: service-failover  # 复用作为测试
            namespace: production
          params:
            - name: serviceName
              value: "smoke-test"
            - name: replicas
              value: "1"
    
    # 阶段 5: 发送完成通知
    - name: notify-complete
      description: "发送容灾完成通知"
      dependsOn: [validation]
      workflows:
        - workflowRef:
            name: notification-workflow
            namespace: production
          params:
            - name: message
              value: "电商系统容灾切换完成"
            - name: event
              value: "failover-complete"
```

应用预案：

```bash
kubectl apply -f ecommerce-dr-plan-with-stages.yaml

# 验证预案状态
kubectl get drplan ecommerce-dr-plan -n production -o yaml
```

#### 步骤 3: 触发容灾切换

```bash
# 方式 1: 创建执行 CR
kubectl apply -f - <<EOF
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: ecommerce-dr-exec-$(date +%Y%m%d-%H%M%S)
  namespace: production
spec:
  planRef:
    name: ecommerce-dr-plan
    namespace: production
  operation: Execute
EOF

# 方式 2: 使用 annotation
kubectl annotate drplan ecommerce-dr-plan \
  -n production \
  dr.bkbcs.tencent.com/trigger=execute
```

#### 步骤 4: 监控执行进度（Stage 视图）

```bash
# 获取最新执行
EXEC_NAME=$(kubectl get drplanexecution -n production \
  -l dr.bkbcs.tencent.com/plan=ecommerce-dr-plan \
  --sort-by=.metadata.creationTimestamp \
  -o jsonpath='{.items[-1].metadata.name}')

# 查看 Stage 状态
kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.stageStatuses[*].name}' && echo
kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.stageStatuses[*].phase}' && echo

# 查看执行摘要
kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.summary}' | jq

# 查看详细的 Stage 和 Workflow 状态
kubectl get drplanexecution $EXEC_NAME -n production -o yaml | grep -A 50 stageStatuses

# 实时监控
watch "kubectl get drplanexecution $EXEC_NAME -n production -o jsonpath='{.status.summary}' | jq"
```

**执行过程示例输出**：

```json
{
  "totalStages": 5,
  "completedStages": 2,
  "runningStages": 1,
  "pendingStages": 2,
  "failedStages": 0,
  "totalWorkflows": 11,
  "completedWorkflows": 5,
  "runningWorkflows": 4,
  "pendingWorkflows": 2,
  "failedWorkflows": 0
}
```

#### 步骤 5: 查看 Stage 详细状态

```bash
# 查看某个 Stage 的 Workflow 执行情况
kubectl get drplanexecution $EXEC_NAME -n production \
  -o jsonpath='{.status.stageStatuses[?(@.name=="infrastructure")]}' | jq

# 输出示例：
{
  "name": "infrastructure",
  "phase": "Running",
  "parallel": true,
  "dependsOn": ["notify-start"],
  "startTime": "2026-02-02T10:00:05Z",
  "workflowExecutions": [
    {
      "workflowRef": {
        "name": "mysql-failover",
        "namespace": "production"
      },
      "phase": "Succeeded",
      "startTime": "2026-02-02T10:00:05Z",
      "completionTime": "2026-02-02T10:00:45Z",
      "duration": "40s"
    },
    {
      "workflowRef": {
        "name": "mysql-failover",
        "namespace": "production"
      },
      "phase": "Running",
      "startTime": "2026-02-02T10:00:05Z",
      "progress": "2/3 actions completed",
      "currentAction": "health-check"
    },
    {
      "workflowRef": {
        "name": "service-failover",
        "namespace": "production"
      },
      "phase": "Running",
      "startTime": "2026-02-02T10:00:05Z",
      "progress": "1/3 actions completed"
    }
  ]
}
```

#### 步骤 6: 执行回滚

```bash
# 触发回滚
kubectl annotate drplan ecommerce-dr-plan \
  -n production \
  dr.bkbcs.tencent.com/trigger=revert

# 监控回滚进度
REVERT_EXEC=$(kubectl get drplanexecution -n production \
  -l dr.bkbcs.tencent.com/plan=ecommerce-dr-plan \
  --sort-by=.metadata.creationTimestamp \
  -o jsonpath='{.items[-1].metadata.name}')

kubectl get drplanexecution $REVERT_EXEC -n production -o jsonpath='{.status.summary}' | jq
```

**回滚执行顺序**（逆序）：
1. Stage 5 (`notify-complete`) → 跳过
2. Stage 4 (`validation`) → 回滚验证服务
3. Stage 3 (`applications`) → 并行回滚所有应用服务
4. Stage 2 (`infrastructure`) → 并行回滚基础设施
5. Stage 1 (`notify-start`) → 跳过

**效果**：
- 所有服务停止向 DR 集群分发
- Localization 被删除，配置恢复原状
- 系统切回主集群正常服务

---

## 最佳实践

### 1. 参数管理

- 使用有意义的参数名称
- 为参数提供描述信息
- 为可选参数设置合理的默认值
- 在预案中明确记录参数用途

### 2. 回滚策略

- **Patch 操作必须定义 rollback**
- Create 操作可利用自动回滚（删除资源）
- HTTP 通知类动作通常不需要回滚
- 测试回滚流程与正向流程同等重要

### 3. 错误处理

- 使用 `failurePolicy: Stop` 确保故障快速停止
- 为 HTTP/Job 动作设置合理的超时时间
- 关键步骤添加重试机制
- 在通知动作中记录详细的执行上下文

### 4. 监控与审计

- 使用有意义的执行 CR 名称（包含时间戳）
- 定期检查 DRPlanExecution 历史记录
- 关注 Kubernetes Events 中的 Warning 事件
- 保留执行日志用于事后分析

### 5. 测试

- 在非生产环境验证工作流
- 测试各种失败场景
- 验证回滚流程完整性
- 定期演练容灾切换流程

### 6. Clusternet 资源管理

- **Localization 先于 Subscription**：先创建 Localization 定制化配置，再创建 Subscription 分发资源
- **Localization namespace**：必须是 ManagedCluster 的命名空间（如 `clusternet-xxxxx`）
- **优先级管理**：使用 `priority` 字段管理多个 Localization 的覆盖顺序（数值越大优先级越高）
- **集群选择器**：使用有意义的标签（如 `region`、`purpose`）选择目标集群
- **资源清理**：回滚时自动删除 Subscription 会停止资源分发，已部署的资源可能需要手动清理

---

## 故障排查

### 问题 1: 预案状态为 Invalid

**原因**：
- 引用的 DRWorkflow 不存在或被删除
- 必填参数缺失或类型不匹配

**解决**：
```bash
# 检查工作流是否存在
kubectl get drworkflow <workflow-name> -n <namespace>

# 查看预案状态详情
kubectl get drplan <plan-name> -n <namespace> -o yaml | grep -A 10 status
```

### 问题 2: 执行卡在 Pending 状态

**原因**：
- 同一 DRPlan 有其他执行正在进行
- Controller 未运行

**解决**：
```bash
# 检查 DRPlan 当前执行
kubectl get drplan <plan-name> -n <namespace> -o jsonpath='{.status.currentExecution}'

# 检查 Controller Pod
kubectl get pods -n drplan-system

# 查看 Controller 日志
kubectl logs -n drplan-system -l app=drplan-controller --tail=100
```

### 问题 3: 动作执行失败

**原因**：参数替换错误、资源不存在、权限不足等

**解决**：
```bash
# 查看详细错误信息
kubectl get drplanexecution <exec-name> -n <namespace> -o yaml

# 查看失败的动作状态
kubectl get drplanexecution <exec-name> -n <namespace> \
  -o jsonpath='{.status.actionStatuses[?(@.phase=="Failed")]}'

# 查看 Events
kubectl describe drplanexecution <exec-name> -n <namespace>
```

### 问题 4: 无法触发回滚

**原因**：DRPlan 状态不是 `Executed`

**解决**：
```bash
# 检查预案状态
kubectl get drplan <plan-name> -n <namespace> -o jsonpath='{.status.phase}'

# 只有 Executed 状态才能回滚
# 如果是 Ready 状态，说明还未执行过
```

---

## 相关资源

- [架构设计文档](../specs/001-drplan-action-executor/spec.md)
- [数据模型定义](../specs/001-drplan-action-executor/data-model.md)
- [CRD 合约示例](../specs/001-drplan-action-executor/contracts/)
- [快速开始指南](../specs/001-drplan-action-executor/quickstart.md)

---

## 支持与反馈

如有问题或建议，请提交 Issue 到项目仓库。
