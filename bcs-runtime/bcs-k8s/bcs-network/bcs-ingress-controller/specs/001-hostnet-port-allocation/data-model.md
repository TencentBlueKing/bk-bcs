# Data Model: HostNetwork 动态端口分配

**Feature**: 001-hostnet-port-allocation  
**Date**: 2026-04-09

---

## 核心实体

### 1. HostNetPortPool (CRD)

**定义位置**: `kubernetes/apis/networkextension/v1/hostnetportpool_types.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| spec.startPort | int32 | 端口范围起始（含） |
| spec.endPort | int32 | 端口范围结束（含） |
| spec.segmentLength | int32 | 最小分配单位（段长度） |
| spec.nodeSelector | map[string]string | 可选：限定可用 Node |
| status.status | string | 当前状态：Ready/NotReady |
| status.nodeAllocations | []NodeHostNetPortPoolStatus | 各 Node 分配概况 |

**状态转换**:
```
Created → Finalizer Added → Cache Synced → Ready
                ↓
         Port Range Shrink → Conflict Event → NotReady → Resolved → Ready
```

---

### 2. 端口段 (Segment)

**概念**: 由 `segmentLength` 个连续端口组成的分配单元

**属性**:
- startPort: 段起始端口
- endPort: 段结束端口（= startPort + segmentLength - 1）
- state: Allocated | Free

**分配规则**:
- portCount = N 时，分配 `ceil(N / segmentLength)` 个连续段
- 段分配必须连续（满足 hostNetwork Pod 端口连续需求）

---

### 3. Pod 分配结果 (BindingResult)

**存储位置**: Pod annotation `hostnetportpool.networkextension.bkbcs.tencent.com/result`

**结构**:
```json
{
  "poolName": "game-server-ports",
  "poolNamespace": "default",
  "nodeName": "node-1",
  "startPort": 30000,
  "endPort": 30009,
  "segmentLength": 10
}
```

**状态注解** (`hostnetportpool.networkextension.bkbcs.tencent.com/status`):
- `NotReady`: 等待分配
- `Ready`: 分配完成
- `Failed`: 分配失败

---

### 4. 内存缓存结构

**Cache 层级**:
```
HostNetPortPoolCache
├── pools: map[poolKey]*PoolAllocator
│   └── nodeAllocators: map[nodeName]*NodeSegmentAllocator
│       ├── totalSegments: []Segment
│       ├── allocatedSegments: map[startPort]*SegmentInfo
│       │   └── podKey, startPort, endPort
│       └── freeSegments: []Segment
```

**关键操作复杂度**:
- AllocateContiguous: O(n) n=段数，需查找连续段
- Release: O(1) 直接释放
- IsPodAllocated: O(m) m=该 Node 已分配段数

---

## 实体关系图

```
┌─────────────────┐       uses        ┌─────────────────┐
│   Pod           │◄─────────────────│  HostNetPortPool│
│  (hostNetwork)  │    annotation   │    (CRD)        │
└────────┬────────┘                 └────────┬────────┘
         │                                 │
         │ allocate                          │ manage
         ▼                                 ▼
┌─────────────────────────────────────────────────┐
│   HostNetPortPoolCache (in-memory per-Node)       │
│   ┌─────────────────────────────────────────────┐│
│   │ NodeSegmentAllocator[node-1]                ││
│   │   ├─ Segment[30000-30009] → Pod-A         ││
│   │   ├─ Segment[30010-30019] → Pod-B         ││
│   │   └─ Segment[30020-30029] (free)          ││
│   └─────────────────────────────────────────────┘│
└─────────────────────────────────────────────────┘
         │
         │ HTTP API
         ▼
┌─────────────────┐
│  initContainer  │
│  (Query Result) │
└─────────────────┘
```

---

## 状态一致性

### 一致性的来源
- **分配结果**: 内存 Cache 是权威来源，Pod Annotation 仅为持久化副本
- **幂等性**: Reconcile 前检查 `cache.IsPodAllocated(podKey)`，避免依赖 Annotation
- **重建**: Controller 重启后从 Pod Annotation 重建 Cache，以 Annotation 为准

### 不一致的恢复
- Patch Annotation 失败时，回滚 Cache 分配
- 定期扫描检查器发现泄漏段时释放

---

## 注解键常量

**位置**: `internal/constant/constant.go`

| 常量名 | 值 |
|--------|-----|
| AnnotationForHostNetPortPool | `hostnetportpool.networkextension.bkbcs.tencent.com` |
| AnnotationForHostNetPortPoolNamespace | `hostnetportpool.networkextension.bkbcs.tencent.com/namespace` |
| AnnotationForHostNetPortPoolPortCount | `hostnetportpool.networkextension.bkbcs.tencent.com/portcount` |
| AnnotationForHostNetPortPoolBindingResult | `hostnetportpool.networkextension.bkbcs.tencent.com/result` |
| AnnotationForHostNetPortPoolBindingStatus | `hostnetportpool.networkextension.bkbcs.tencent.com/status` |
| FinalizerNameHostNetPortPool | `hostnetportpool.bkbcs.tencent.com` |

---

## 指标维度

| 指标名 | 类型 | Labels |
|--------|------|--------|
| hostnet_segment_allocated | Gauge | pool_name, pool_namespace, node_name |
| hostnet_segment_total | Gauge | pool_name, pool_namespace, node_name |
| hostnet_allocate_failed | Gauge | pod_name, pod_namespace |
| hostnet_allocate_failed_total | Counter | pool_name, node_name |
| hostnet_pool_shrink_conflict_total | Counter | pool_name, pool_namespace, node_name |
| hostnet_segment_leaked_total | Counter | pool_name, pool_namespace, node_name |
| hostnet_reconcile_duration | Histogram | - |
| hostnet_cache_rebuild_recovered | Gauge | - |
