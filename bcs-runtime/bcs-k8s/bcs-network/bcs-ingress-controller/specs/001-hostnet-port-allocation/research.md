# Research: HostNetwork 动态端口分配

**Feature Branch**: `001-hostnet-port-allocation`  
**Date**: 2026-03-16

## 研究问题与结论

### R-001: 新缓存（HostNetPortPoolCache）是否应复用现有 portpoolcache 包？

**结论**: 不复用，独立实现。

**理由**:
- 现有 `portpoolcache.Cache` 绑定了 `CachePoolItem`（对应 LB 上的 PoolItem）和协议维度（TCP/UDP/TCP_UDP），数据结构包含 `PortListMap: map[protocol]*CachePortList`，强耦合了 LB 端口管理语义。
- 新缓存的核心差异：per-Node 隔离（同一端口段在不同 Node 上独立分配）、无协议维度、连续多段分配（first-fit 算法）。
- 复用会引入大量适配代码和不必要的抽象层，违反 Constitution 中"复杂度 MUST 有合理理由，遵循 YAGNI 原则"。

**替代方案评估**:
- 方案 A：在 `portpoolcache` 中加入 Node 维度 → 改动过大，影响现有 PortPool 功能稳定性。
- 方案 B：抽取公共接口再各自实现 → 两者语义差异太大（单端口分配 vs 连续段分配，全局 vs per-Node），公共接口会过于抽象。

### R-002: HostNetPortPoolCache 的新包放置位置

**结论**: `internal/hostnetportpoolcache/` 目录，独立于 `internal/portpoolcache/`。

**理由**:
- 与现有 `portpoolcache` 包同级，保持包结构的清晰性。
- 包内文件组织参考 `portpoolcache` 的模式：`cache.go`（主入口）、`types.go`（数据结构）、`cache_test.go`（单元测试）。
- 遵循 Constitution 中"包命名 MUST 使用小写单词"。

### R-003: Controller 是否应独立为新目录还是放在现有 portbindingcontroller 下？

**结论**: 独立为新目录 `hostnetportcontroller/`，与 `portbindingcontroller/` 同级。

**理由**:
- 技术方案明确要求"独立的 Controller"，与 PortBinding 的职责完全不同（一个管 LB Listener，一个管 Node 端口段）。
- 现有 `portbindingcontroller` 已有 7 个文件 + PodFilter + NodeFilter，加入不相关的逻辑会增加认知负担。
- 独立目录便于后续维护和代码审查。

**文件规划**:
- `hostnetportcontroller/controller.go` — Reconciler 定义与 SetupWithManager
- `hostnetportcontroller/podfilter.go` — Pod 事件过滤器
- `hostnetportcontroller/nodefilter.go` — Node 事件过滤器
- `hostnetportcontroller/reconcile.go` — 核心 Reconcile 逻辑（reconcilePod、reconcilePool、reconcileNode）

### R-004: Pod 事件过滤器应检查哪些字段变更才入队？

**结论**: 参考现有 `portbindingcontroller/podfilter.go` 的 `checkPodNeedReconcile` 模式，对 HostNet Pod 检查以下字段变更：

1. `spec.nodeName` 变更（空 → 有值，表示调度完成）
2. `status.phase` 变更（Running → Failed，表示驱逐/失败）
3. `metadata.deletionTimestamp` 变更（nil → 非 nil，表示开始删除）
4. `metadata.annotations` 中 `hostnetportpool.*` 相关 key 变更

**理由**:
- 与现有 PodFilter 的检查逻辑对齐，避免遗漏关键状态变更。
- `nodeName` 检查是新增需求——现有 PodFilter 不需要，因为 PortPool 在 Webhook 阶段分配，不依赖 nodeName。
- `status.phase` 检查是驱逐场景释放端口段的关键入口。

### R-005: HTTP API 查询端口分配结果的实现方式

**结论**: 在现有 `internal/httpsvr/` 中新增 `hostnetportpool.go` 文件，在 `InitRouters` 中注册新路由。

**理由**:
- 现有 HTTP Server 已通过 Service 暴露在集群内（端口 18088），API 路径统一使用 `/ingresscontroller/api/v1/` 前缀。
- `HttpServerClient` 结构体需新增一个字段持有 `HostNetPortPoolCache` 引用，用于查询分配结果。
- 查询逻辑：根据 podName/podNamespace 从 K8s API 获取 Pod，读取其 result 和 status annotation（即 `AnnotationForHostNetPortPoolBindingResult` 和 `AnnotationForHostNetPortPoolBindingStatus`），或直接从缓存中查找 Pod 的分配状态。

**API 设计决策**:
- 选择从 Pod annotation 读取而非从缓存读取，因为 annotation 是分配结果的事实来源（source of truth），且 Controller 重启后缓存可能尚未完成重建。
- 但为了降低 API Server 压力，优先从 controller-runtime 的 informer cache（`client.Reader`）读取 Pod，而非直接访问 API Server。

### R-006: 缓存重建的时机与策略

**结论**: 在 Reconcile 首次被调用时执行 `initCache()`，通过 `cacheSynced` 标志位控制仅执行一次。

**理由**:
- 参考现有 `PortPoolReconciler` 的 `isCacheSync` 模式——在首次 Reconcile 时同步缓存。
- 不在 `SetupWithManager` 中同步，因为此时 informer cache 可能尚未就绪。
- 重建数据源：List 所有 `HostNetPortPool` CR + List 所有 Pod（过滤带 hostnetportpool annotation 的）。

### R-007: HostNetPortPool CRD 是否需要在 scheme 中注册？

**结论**: 需要。当前 `hostnetportpool_types.go` 中缺少 `init()` 函数和 `SchemeBuilder.Register`。

**理由**:
- 查看现有 `portpool_types.go`，每个 CRD 类型都在 `init()` 中注册到 SchemeBuilder。
- 查看 `groupversion_info.go`，使用统一的 `SchemeBuilder` 注册所有类型。
- 缺少注册会导致 controller-runtime 无法识别 HostNetPortPool 类型，`Get`/`List`/`Watch` 操作会失败。

**行动项**: 在 `hostnetportpool_types.go` 中添加 `init()` 函数注册到 SchemeBuilder。

### R-008: 定期扫描检查器（HostNetPortSegmentChecker）的集成方式

**结论**: 实现 `check.Checker` 接口（`Run()` 方法），注册到现有的 `CheckRunner` 中。

**理由**:
- 现有 `internal/check/checkrunner.go` 提供了统一的定期检查运行器，管理多个 Checker 的生命周期。
- `PortLeakChecker` 已经按此模式集成，新增的 `HostNetPortSegmentChecker` 遵循相同模式。
- 在 `main.go` 中创建 Checker 实例并注册到 CheckRunner。

### R-009: Node 事件的 namespace 编码策略

**结论**: 使用 `__node__` 前缀区分 Node 事件，与技术方案保持一致。

**理由**:
- 现有 `portbindingcontroller/nodefilter.go` 使用 `nodePortBindingNs` 字段（通常为 `bcs-system`），将 Node 事件映射到特定 namespace 下的 PortBinding 资源。
- HostNet 场景不涉及 PortBinding，但需要在 Reconcile 入口区分 Pod、HostNetPortPool、Node 三种资源。
- 使用 `__node__` 前缀是一种简洁的 convention，Reconcile 入口通过字符串前缀判断资源类型。

### R-010: HostNetPortPool CRD 的 deepcopy 生成

**结论**: 需要为 `hostnetportpool_types.go` 中的类型生成 DeepCopy 方法。

**理由**:
- controller-runtime 的 `client.Object` 接口要求实现 `DeepCopyObject()` 方法。
- 现有类型的 DeepCopy 方法在 `zz_generated.deepcopy.go` 中自动生成（通过 `controller-gen`）。
- 新增 CRD 类型后需重新运行代码生成工具。

**行动项**: 确认 HostNetPortPool 和 HostNetPortPoolList 类型已有 `+k8s:deepcopy-gen` 标记（已确认有），运行 `controller-gen` 生成 DeepCopy 代码。
