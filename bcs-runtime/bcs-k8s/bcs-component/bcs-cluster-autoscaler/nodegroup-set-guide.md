# NodeGroupSet 用户文档

## 1. 概述

NodeGroupSet 是 BCS Cluster Autoscaler 的一个高级功能，允许将多个 NodeGroup（节点池）组合成一个 Set（节点池组），并为每个 Set 配置独立的：

- **缓冲资源比例 (Buffer Ratio)**：控制 Set 内资源水位线
- **扩容策略 (Expander)**：控制 Set 内节点池的选择优先级

### 1.1 核心概念

| 概念 | 说明 |
|------|------|
| **NodeGroup** | 节点池，一组具有相同配置的节点 |
| **NodeGroupSet** | 节点池组，多个 NodeGroup 的逻辑分组 |
| **BufferedRatio** | 缓冲资源比例，触发扩/缩容的阈值 |
| **GlobalSet** | 特殊的 Set，包含未分配到任何显式 Set 的 NodeGroup |

### 1.2 适用场景

- **多业务隔离**：不同业务线使用不同的节点池组，各自配置独立的资源水位
- **分级服务**：核心服务和非核心服务使用不同的缓冲比例
- **资源类型分组**：CPU 密集型和内存密集型工作负载分开管理
- **混合云架构**：按可用区、机型等维度分组管理

## 2. 配置说明

### 2.1 启用 NodeGroupSet

通过启动参数指定配置文件路径：

```bash
--nodegroup-set-config=/path/to/nodegroup-set-config.yaml
```

### 2.2 配置文件格式

```yaml
sets:
  # Set 1: 核心服务节点池组
  - setName: "core-services"
    nodeGroupIDs:
      - "ng-core-001"
      - "ng-core-002"
      - "ng-core-003"
    bufferedRatios:
      cpuRatio: 0.3      # CPU 保持 30% 缓冲
      memRatio: 0.3      # 内存保持 30% 缓冲
      resourceRatio: 0.2 # 其他资源保持 20% 缓冲
    expander: "priority"
    expanderPriorities:
      - "ng-core-001"    # 优先级最高
      - "ng-core-002"
      - "ng-core-003"    # 优先级最低

  # Set 2: 普通业务节点池组
  - setName: "general-workloads"
    nodeGroupIDs:
      - "ng-general-001"
      - "ng-general-002"
    bufferedRatios:
      cpuRatio: 0.15     # CPU 保持 15% 缓冲
      memRatio: 0.15     # 内存保持 15% 缓冲
      resourceRatio: 0.1

  # Set 3: GPU 节点池组
  - setName: "gpu-workloads"
    nodeGroupIDs:
      - "ng-gpu-v100"
      - "ng-gpu-a100"
    bufferedRatios:
      cpuRatio: 0.1
      memRatio: 0.1
      resourceRatio: 0.2 # GPU 资源保持更高缓冲
```

### 2.3 配置项详解

| 配置项 | 类型 | 必填 | 说明 |
|--------|------|------|------|
| `setName` | string | 是 | Set 的唯一标识名称 |
| `nodeGroupIDs` | []string | 是 | 该 Set 包含的 NodeGroup ID 列表 |
| `bufferedRatios.cpuRatio` | float | 否 | CPU 缓冲比例 (0.0-1.0)，默认 0 |
| `bufferedRatios.memRatio` | float | 否 | 内存缓冲比例 (0.0-1.0)，默认 0 |
| `bufferedRatios.resourceRatio` | float | 否 | 其他资源缓冲比例 (0.0-1.0)，默认 0 |
| `expander` | string | 否 | Set 级别扩容策略，可选值见下表 |
| `expanderPriorities` | []string | 条件必填 | 当 expander=priority 时必填 |

**支持的 Expander 策略：**

| Expander | 说明 |
|----------|------|
| `random` | 随机选择 NodeGroup |
| `most-pods` | 选择能容纳最多 Pod 的 NodeGroup |
| `least-waste` | 选择资源浪费最少的 NodeGroup |
| `priority` | 按配置的优先级列表选择 NodeGroup |

## 3. 扩缩容逻辑

### 3.1 扩容触发条件

NodeGroupSet 支持两种扩容触发机制：

#### 3.1.1 Buffer 触发扩容（资源水位不足）

当 Set 内的可用资源低于配置的缓冲比例时触发：

```
触发条件: (Set 可用资源 / Set 总资源) < bufferedRatio
```

**示例：**
- Set 配置 `cpuRatio: 0.3`
- Set 当前有 100 核 CPU 总量，已使用 75 核
- 可用率 = (100 - 75) / 100 = 25% < 30%
- **触发扩容**

#### 3.1.2 Pod 触发扩容（Pending Pod）

当有 Pending Pod 且可以被 Set 内的 NodeGroup 调度时触发。

### 3.2 扩容流程

```
┌─────────────────────────────────────────────────────────────────┐
│                      CA 主循环开始                               │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. RefreshNodeGroups: 刷新 NodeGroup → Set 映射                 │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. TakeResourceSnapshot: 创建资源快照                           │
│     - 遍历所有节点，计算资源使用情况                              │
│     - 按 Set 聚合资源信息                                        │
│     - 缓存 NodeToGroup 映射                                      │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. CheckScaleUpNeeded: 检查每个 Set 是否需要扩容                 │
│     - 对比 (1 - 使用率) vs bufferedRatio                         │
│     - 记录触发扩容的 Set 列表                                    │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
              ┌─────────────┴─────────────┐
              │  存在触发扩容的 Set?       │
              └─────────────┬─────────────┘
                   Yes │         │ No
                       ▼         ▼
┌──────────────────────────┐    │
│  4. 执行扩容:             │    │
│  - SetAwareExpander 选择  │    │
│    最佳 NodeGroup        │    │
│  - ClusterResourceEstimator│   │
│    计算扩容节点数量       │    │
│  - 调用 CloudProvider     │    │
│    创建节点               │    │
└──────────────────────────┘    │
                                │
                            ▼   ▼
┌─────────────────────────────────────────────────────────────────┐
│                         继续缩容检查                             │
└─────────────────────────────────────────────────────────────────┘
```

### 3.3 扩容节点数量计算

Estimator 根据 Set 的资源状态计算需要扩容的节点数量：

```go
// 公式: (sum * ratio - left) / (1 - ratio) / nodeCapacity
// 其中:
//   sum = Set 内总资源
//   left = Set 内可用资源
//   ratio = 配置的 bufferedRatio
//   nodeCapacity = 模板节点的资源容量
```

**关键特性：**
- **Set Scope 计算**：只计算 Set 内节点的资源，不包含其他 Set
- **冷启动支持**：Set 为空时（无节点），会扩容至少 1 个节点满足 buffer 需求
- **Upcoming 节点感知**：已申请但未就绪的节点也计入资源计算，避免重复扩容

### 3.4 缩容逻辑

缩容时，CA 会检查删除节点后 Set 的资源是否仍满足 buffer 要求：

```
┌─────────────────────────────────────────────────────────────────┐
│                     filterNodes: 过滤可删除节点                   │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  对每个候选删除节点:                                              │
│  1. 查找节点所属的 Set                                           │
│  2. SimulateNodeRemoval: 模拟删除后的资源状态                     │
│     - 计算删除后的总资源和可用资源                                │
│     - 考虑节点上 Pod 重调度的资源消耗                             │
│  3. 检查删除后是否仍满足 buffer ratio                            │
└───────────────────────────┬─────────────────────────────────────┘
                            │
              ┌─────────────┴─────────────┐
              │  删除后满足 buffer ratio?  │
              └─────────────┬─────────────┘
                   Yes │         │ No
                       ▼         ▼
              ┌────────────┐  ┌────────────────┐
              │ 允许删除    │  │ 标记为不可删除  │
              │            │  │ (BufferNotEnough)│
              └────────────┘  └────────────────┘
```

**累积模拟**：缩容检查采用累积模拟，考虑多个节点同时删除的效果：

```go
// 深拷贝快照用于累积模拟
workingSnapshot = snapshot.DeepCopy()

for _, node := range candidateNodes {
    // 模拟删除当前节点
    setResourceAfterRemoval = SimulateNodeRemoval(set, workingSnapshot, node)
    
    if !meetsBufferRatio(setResourceAfterRemoval) {
        // 不可删除
        continue
    }
    
    // 更新工作快照，累积删除效果
    workingSnapshot.UpdateSetResourceInfo(set, setResourceAfterRemoval)
    workingSnapshot.RemoveNodeFromSnapshot(node)
    
    // 允许删除
    allowedNodes = append(allowedNodes, node)
}
```

### 3.5 GlobalSet 处理

未分配到任何显式 Set 的 NodeGroup 会自动归入 `__global__` Set：

| 场景 | 行为 |
|------|------|
| 扩容触发 | 使用全局 buffer ratio 参数 (`--buffer-cpu-ratio` 等) |
| 扩容策略 | 使用全局 expander 配置 (`--expander`) |
| 缩容检查 | 使用全局 buffer ratio 检查资源水位 |
| 资源计算 | 基于所有 GlobalSet 节点计算，不与其他 Set 混合 |

## 4. Set 级别 Expander

### 4.1 SetAwareExpander 工作流程

```
┌─────────────────────────────────────────────────────────────────┐
│  BestOption: 选择最佳扩容选项                                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  1. 按 Set 分组所有扩容选项                                       │
│     options[SetA] = [ng-a1, ng-a2]                              │
│     options[SetB] = [ng-b1]                                     │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  2. 对每个 Set 应用 Set 级别 expander                            │
│     - 有配置: 使用 Set 的 expander 选择                          │
│     - 无配置: 使用全局 expander                                  │
└───────────────────────────┬─────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────────┐
│  3. 如果多个 Set 都有最佳选项，使用全局 expander 最终选择         │
└─────────────────────────────────────────────────────────────────┘
```

### 4.2 Priority Expander 配置示例

```yaml
- setName: "production"
  nodeGroupIDs:
    - "ng-spot-large"    # 竞价实例（便宜）
    - "ng-ondemand-large" # 按需实例（稳定）
  bufferedRatios:
    cpuRatio: 0.2
  expander: "priority"
  expanderPriorities:
    - "ng-spot-large"      # 优先使用竞价实例
    - "ng-ondemand-large"  # 竞价不可用时使用按需
```

## 5. 监控与告警

### 5.1 关键指标

| 指标 | 说明 |
|------|------|
| `cluster_autoscaler_unremovable_nodes_count` | 不可删除节点数量，label 包含 Set 名称和原因 |
| Set 资源使用率 | 通过日志输出：`Set %s CPU available ratio %.2f` |

### 5.2 日志示例

**扩容触发日志：**
```
I1205 10:00:00.000000   12345 nodegroup_set_manager.go:237] Set core-services CPU available ratio 0.18 < threshold 0.30, scale up needed
```

**缩容阻止日志：**
```
I1205 10:00:00.000000   12345 scale_down.go:xxx] Set core-services: node xxx cannot be removed due to buffer ratio constraint
```

## 6. 最佳实践

### 6.1 Buffer Ratio 设置建议

| 业务类型 | CPU Ratio | Mem Ratio | 说明 |
|----------|-----------|-----------|------|
| 核心服务 | 0.3-0.4 | 0.3-0.4 | 预留更多缓冲，保证扩容时间 |
| 一般服务 | 0.15-0.25 | 0.15-0.25 | 平衡成本与响应速度 |
| 批处理任务 | 0.05-0.1 | 0.05-0.1 | 允许较高利用率 |

### 6.2 NodeGroup 分组策略

1. **按业务线分组**：不同业务使用独立的 Set，避免相互影响
2. **按资源类型分组**：CPU 密集型、内存密集型、GPU 工作负载分开
3. **按可用区分组**：多可用区部署时，按可用区分 Set 可实现区域级容量管理
4. **按优先级分组**：通过 priority expander 实现成本优化（如优先使用竞价实例）

### 6.3 注意事项

1. **NodeGroup 唯一性**：一个 NodeGroup 只能属于一个 Set
2. **配置热更新**：配置文件变更需要重启 CA 生效
3. **资源隔离**：Set 之间资源独立计算，不会相互影响扩缩容决策
4. **兼容性**：未配置 NodeGroupSet 时，CA 使用全局 buffer ratio，行为与原版一致

## 7. 故障排查

### 7.1 常见问题

| 问题 | 可能原因 | 解决方案 |
|------|----------|----------|
| Set 不触发扩容 | buffer ratio 配置为 0 | 检查配置文件中的 cpuRatio/memRatio |
| 节点无法缩容 | 删除后不满足 buffer | 降低 buffer ratio 或等待负载下降 |
| NodeGroup 未分配到 Set | ID 不匹配 | 确认 nodeGroupIDs 与实际 NodeGroup ID 一致 |

### 7.2 调试命令

```bash
# 查看 CA 日志，过滤 Set 相关信息
kubectl logs -n bcs-system deploy/cluster-autoscaler | grep -E "(NodeGroupSet|Set.*scale|Set.*ratio)"

# 查看配置是否加载成功
kubectl logs -n bcs-system deploy/cluster-autoscaler | grep "Successfully loaded NodeGroupSet config"
```

## 8. 附录：配置文件完整示例

```yaml
# nodegroup-set-config.yaml
sets:
  # 核心业务节点池组 - 高可用配置
  - setName: "core-services"
    nodeGroupIDs:
      - "ng-core-zone-a"
      - "ng-core-zone-b"
      - "ng-core-zone-c"
    bufferedRatios:
      cpuRatio: 0.3
      memRatio: 0.3
      resourceRatio: 0.2
    expander: "priority"
    expanderPriorities:
      - "ng-core-zone-a"
      - "ng-core-zone-b"
      - "ng-core-zone-c"

  # 普通业务节点池组 - 成本优化配置
  - setName: "general-workloads"
    nodeGroupIDs:
      - "ng-spot-4c8g"
      - "ng-spot-8c16g"
      - "ng-ondemand-4c8g"
    bufferedRatios:
      cpuRatio: 0.15
      memRatio: 0.15
    expander: "priority"
    expanderPriorities:
      - "ng-spot-4c8g"      # 优先使用小规格竞价
      - "ng-spot-8c16g"     # 其次大规格竞价
      - "ng-ondemand-4c8g"  # 最后按需实例

  # GPU 节点池组
  - setName: "gpu-workloads"
    nodeGroupIDs:
      - "ng-gpu-v100-1"
      - "ng-gpu-a100-1"
    bufferedRatios:
      cpuRatio: 0.1
      memRatio: 0.1
      resourceRatio: 0.25  # GPU 资源预留更多
    expander: "least-waste"

  # 批处理任务节点池组 - 高利用率配置
  - setName: "batch-jobs"
    nodeGroupIDs:
      - "ng-batch-spot"
    bufferedRatios:
      cpuRatio: 0.05
      memRatio: 0.05
```
