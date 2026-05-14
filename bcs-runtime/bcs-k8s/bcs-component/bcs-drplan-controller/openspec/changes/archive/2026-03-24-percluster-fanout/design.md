## Context

DRPlan 通过 Stage/Workflow/Action 编排跨集群下发与运维动作。已实现的 `waitReady` 机制可在 Subscription 创建后轮询子集群资源状态，直到所有 binding clusters 上的 feeds 全部 ready。

然而这一 "全局等待" 语义与 Helm 的预期行为不匹配：Helm 在单个节点/集群上完成 pre-install hook 后，立即继续该集群的后续步骤。在多集群场景下，一个慢集群会阻塞所有集群的执行进度。

本变更引入 **"单 workflow 定义，运行时按 bindingClusters 进行 per-cluster 扇出执行"** 能力，保持定义层简洁的同时，实现运行时集群级独立推进。

## Goals / Non-Goals

**Goals:**

- 为 `Action` 新增 `clusterExecutionMode` 字段，支持 `Global`（默认）和 `PerCluster` 两种模式
- `PerCluster` 模式下，运行时自动将逻辑 Subscription 拆分为 per-cluster 子 Subscription
- 连续的 `PerCluster` action 形成 "per-cluster batch"，同一集群内按序执行，不同集群间独立推进
- `PerCluster → Global` 边界处实现全局屏障（barrier），所有集群完成后才继续
- 扩展 `ActionStatus` 以记录 per-cluster 执行状态
- drplan-gen 自动为 hook Subscription 设置 `clusterExecutionMode: PerCluster`
- **完全向后兼容**：`clusterExecutionMode` 不设置或为空 = `Global` = 现有行为

**Non-Goals:**

- 不实现 `Auto` 模式（自动判断是否拆分）——后续版本再考虑
- 不支持非 Subscription 类型 action 的 `PerCluster` 拆分
- 不支持自定义汇聚策略（如 "N/M 集群成功即可"）
- 不支持跨 workflow 的 per-cluster 状态共享
- 不在本次实现 per-cluster 子 Subscription 的自动 GC Finalizer（先通过 OwnerReference 实现基本清理）

## Decisions

### 1. API 设计

**`Action.ClusterExecutionMode` 字段：**

```go
type Action struct {
    // ...existing fields...

    // ClusterExecutionMode controls how this action is executed across clusters.
    // "Global" (default, also when empty): executes as a single aggregate action.
    // "PerCluster": at runtime, splits into per-cluster child actions for independent progression.
    // Only effective for Subscription actions with waitReady=true.
    // +kubebuilder:validation:Enum="";Global;PerCluster
    // +kubebuilder:default=""
    // +optional
    ClusterExecutionMode string `json:"clusterExecutionMode,omitempty"`
}
```

兼容性：空值 = `Global`，存量 action 无需修改。

**`ClusterActionStatus` 新类型：**

```go
type ClusterActionStatus struct {
    Cluster        string       `json:"cluster"`
    ClusterID      string       `json:"clusterID"`
    Phase          string       `json:"phase"`
    StartTime      *metav1.Time `json:"startTime,omitempty"`
    CompletionTime *metav1.Time `json:"completionTime,omitempty"`
    Message        string       `json:"message,omitempty"`
}
```

**`ActionStatus` 扩展：**

```go
type ActionStatus struct {
    // ...existing fields...

    // ClusterStatuses records per-cluster execution state for PerCluster actions.
    // Empty for Global actions (backward compatible).
    // +optional
    ClusterStatuses []ClusterActionStatus `json:"clusterStatuses,omitempty"`
}
```

### 2. 拆分判定规则

仅当以下条件**全部**满足时，才进行 per-cluster 拆分：

1. `action.ClusterExecutionMode == "PerCluster"`
2. `action.Type == "Subscription"`
3. `action.WaitReady == true`

不满足任一条件 → 按 `Global` 行为执行。这确保只有真正需要 per-cluster 推进的场景才拆分。

### 3. 运行时子 Subscription 生成

**步骤：**

1. 先创建原始（parent）Subscription，获取其 `status.bindingClusters`
2. 对每个 binding cluster，生成子 Subscription：
   - 名称：`{parent-name}--{cluster-name}`（双连字符分隔）
   - subscribers 仅包含该单一集群
   - feeds 与 parent 相同
3. 子 Subscription 设置 OwnerReference 指向 parent Subscription
4. 独立对每个子 Subscription 执行 waitReady 检查

**为什么创建子 Subscription 而不是只做 readiness 查询拆分：**

- 子 Subscription 让 Clusternet 调度器真正只向该集群下发资源
- parent Subscription 可能触发向所有集群下发，但子 Subscription 精确控制范围
- 后续 `PerCluster` action 也可复用相同的集群列表

### 4. Workflow Executor 调度模型

**Batch 分组：**

将 workflow 的 action 列表按 `clusterExecutionMode` 分组为若干 batch：

```
Actions: [A1(PerCluster), A2(PerCluster), A3(Global), A4(PerCluster)]
→ Batch 1: [A1, A2] (PerCluster)
→ Batch 2: [A3]     (Global)
→ Batch 3: [A4]     (PerCluster)
```

**执行规则：**

- **PerCluster batch**：为每个 binding cluster 启动独立 goroutine，按顺序执行 batch 内 action。不同集群并发执行。
- **Global batch**：所有前置 batch 完成后，按现有顺序串行执行。
- **Batch 之间**：严格顺序，上一 batch 全部完成才进入下一 batch。

**PerCluster batch 内的 per-cluster 执行：**

```go
// 伪代码
for clusterIdx, cluster := range bindingClusters {
    go func(cluster) {
        for _, action := range batchActions {
            status := executor.ExecutePerCluster(ctx, action, cluster, params)
            updateClusterActionStatus(action.Name, cluster, status)
            if status.Failed() { break }
        }
    }(cluster)
}
waitAll()
aggregateBatchStatus()
```

**汇聚规则（PerCluster action）：**

| 集群状态分布 | 汇聚后 Phase |
|-------------|-------------|
| 全部 Succeeded | Succeeded |
| 任一 Failed | Failed |
| 任一 Running（无 Failed） | Running |
| 全部 Pending | Pending |

### 5. drplan-gen 适配

```yaml
# Hook Subscription — 需要 per-cluster 独立推进
- name: demo-app-db-migrate
  type: Subscription
  clusterExecutionMode: PerCluster
  when: 'mode == "install"'
  waitReady: true
  subscription:
    name: demo-app-db-migrate-sub
    # ...

# 主资源 Subscription — 整体部署
- name: create-subscription
  type: Subscription
  # clusterExecutionMode 省略或 Global
  waitReady: true
  subscription:
    name: demo-app-sub
    # ...
```

drplan-gen 判定规则：
- 来自 Helm hook 的 Subscription action → `clusterExecutionMode: PerCluster`
- 主资源 Subscription action → 不设置（默认 Global）

### 6. Revert 支持

- `PerCluster` action 的 rollback 需要清理所有子 Subscription
- 通过 OwnerReference 的级联删除，删除 parent Subscription 即可清理子 Subscription
- 回滚时，`ActionStatus.ClusterStatuses` 提供需要回滚的集群信息

## Risks / Trade-offs

- **子 Subscription 命名冲突**：使用双连字符 `--` 分隔 parent-name 和 cluster-name，需确保不超过 K8s 253 字符限制。加入长度检查和截断逻辑。
- **并发 goroutine 管理**：PerCluster batch 中每个集群一个 goroutine，需要 proper context cancellation 和 error propagation（errgroup）。
- **Status 对象体积**：大量集群时 `ClusterStatuses` 可能显著增加 status 大小。后续可考虑独立 ConfigMap 存储详情。第一版限制在合理集群数（< 50）范围内使用。
- **Clusternet 调度行为**：子 Subscription 的单集群 subscriber 需要验证 Clusternet 确实只向该集群调度。如果 Clusternet 行为不符预期，需要 fallback 到直接过滤 readiness check 而不创建子 Subscription。
- **复杂度增加**：workflow executor 从简单串行循环变为 batch + 并发模型，增加了测试和调试的复杂度。通过充分的单元测试和详细日志缓解。
- **Parent Subscription 资源冲突**：parent Subscription 已经创建并触发全集群调度，子 Subscription 可能导致重复 feed 部署。需要评估是否在 PerCluster 模式下跳过 parent Subscription 创建，改为直接创建子 Subscription。
