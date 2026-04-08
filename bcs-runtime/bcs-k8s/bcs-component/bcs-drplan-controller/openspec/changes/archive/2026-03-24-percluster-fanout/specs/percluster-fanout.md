# Delta Spec: Per-Cluster Fan-out Execution

## 变更范围

本 delta spec 描述 "per-cluster fan-out execution" 特性对现有 API 和内部组件的增量变更。

---

## API 变更

### 1. `Action` struct（`api/v1alpha1/common_types.go`）

**新增字段：**

```go
// ClusterExecutionMode controls how this action is executed across clusters.
// Empty or "Global" (default): executes as a single aggregate action, waitReady waits for ALL clusters.
// "PerCluster": runtime splits into per-cluster child Subscriptions for independent progression.
// Only effective for Subscription type actions with waitReady=true.
// +kubebuilder:validation:Enum="";Global;PerCluster
// +kubebuilder:default=""
// +optional
ClusterExecutionMode string `json:"clusterExecutionMode,omitempty"`
```

**兼容性**：空值 = `Global`，存量 YAML 无需修改。

### 2. `ActionStatus` struct（`api/v1alpha1/common_types.go`）

**新增字段：**

```go
// ClusterStatuses records per-cluster execution state when ClusterExecutionMode=PerCluster.
// Empty for Global actions (backward compatible).
// +optional
ClusterStatuses []ClusterActionStatus `json:"clusterStatuses,omitempty"`
```

### 3. 新增类型 `ClusterActionStatus`（`api/v1alpha1/common_types.go`）

```go
// ClusterActionStatus tracks execution state of a single cluster within a PerCluster action.
type ClusterActionStatus struct {
    // Cluster is the binding cluster identifier (format: "namespace/name")
    Cluster string `json:"cluster"`

    // ClusterID is the ManagedCluster spec.clusterId
    ClusterID string `json:"clusterID"`

    // Phase is the per-cluster action phase: Pending, Running, Succeeded, Failed
    // +kubebuilder:validation:Enum=Pending;Running;Succeeded;Failed
    Phase string `json:"phase"`

    // StartTime is when this cluster's execution started
    // +optional
    StartTime *metav1.Time `json:"startTime,omitempty"`

    // CompletionTime is when this cluster's execution completed
    // +optional
    CompletionTime *metav1.Time `json:"completionTime,omitempty"`

    // Message provides additional information
    // +optional
    Message string `json:"message,omitempty"`
}
```

### 4. 新增常量（`api/v1alpha1/constants.go`）

```go
// ClusterExecutionMode constants
const (
    ClusterExecutionModeGlobal     = "Global"
    ClusterExecutionModePerCluster = "PerCluster"
)
```

---

## 内部组件变更

### 5. `SubscriptionActionExecutor`（`internal/executor/subscription_executor.go`）

**新增方法：**

- `ExecutePerCluster(ctx, action, clusterBinding, params) → (ClusterActionStatus, error)`
  - 为指定集群创建子 Subscription（单集群 subscriber）
  - 对子 Subscription 执行 waitReady
  - 返回该集群的执行状态

- `createChildSubscription(ctx, parentName, parentNS, clusterBinding, feeds, spec) → (*unstructured.Unstructured, error)`
  - 生成子 Subscription，名称格式 `{parentName}--{clusterShortName}`
  - 设置 OwnerReference 指向 parent Subscription
  - subscribers 仅含该单一集群

- `resolveBindingClusters(ctx, namespace, name) → ([]string, error)`
  - 从已创建的 Subscription status 中获取 binding clusters

**修改方法：**

- `Execute()`: 当 `isPerClusterMode(action)` 时，跳过整体 waitReady，改为返回 "PerCluster mode, handled by workflow executor" 状态

**新增辅助函数：**

- `isPerClusterMode(action) bool`: 判断是否满足 per-cluster 拆分条件（mode + type + waitReady 三条件）

### 6. `NativeWorkflowExecutor`（`internal/executor/native_executor.go`）

**重构 `ExecuteWorkflow()`：**

原有逻辑（顺序遍历 action 列表）改为：

1. 将 action 列表按 `clusterExecutionMode` 分组为 batch
2. 对每个 batch：
   - **Global batch**：串行执行（保持原有逻辑）
   - **PerCluster batch**：
     a. 执行第一个 action 的 parent Subscription 获取 binding clusters
     b. 为每个 cluster 启动 goroutine，按序执行 batch 内所有 action 的 per-cluster 分支
     c. 使用 `errgroup` 等待所有 cluster 完成
     d. 汇聚 `ClusterActionStatus` 到各 action 的 `ActionStatus`

**新增方法：**

- `groupActionBatches(actions []Action) → []actionBatch`
- `executePerClusterBatch(ctx, batch, bindingClusters, params) → ([]ActionStatus, error)`
- `aggregateClusterStatuses(clusterStatuses) → Phase`

### 7. `drplan-gen` Generator（`internal/generator/`）

**修改：**

- Hook Subscription action 输出时追加 `clusterExecutionMode: PerCluster`
- 主资源 Subscription action 不设置此字段（默认 Global）

---

## CRD 变更

运行 `make manifests` 后，CRD YAML 将自动反映：
- `Action` 增加 `clusterExecutionMode` 字段
- `ActionStatus` 增加 `clusterStatuses` 数组
- 新增 `ClusterActionStatus` 的 schema

---

## 行为变更矩阵

| clusterExecutionMode | type | waitReady | 运行时行为 |
|---------------------|------|-----------|----------|
| 空/Global | 任意 | 任意 | 现有行为，不变 |
| PerCluster | Subscription | true | 拆分为 per-cluster 子 Subscription，独立推进 |
| PerCluster | Subscription | false | 等同 Global（无 waitReady 则不需要拆分） |
| PerCluster | 非 Subscription | 任意 | 等同 Global（本版本不支持非 Subscription 拆分） |
