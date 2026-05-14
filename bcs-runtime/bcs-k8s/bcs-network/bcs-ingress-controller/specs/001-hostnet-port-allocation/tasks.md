# Tasks: HostNetwork 动态端口分配

**Input**: 设计文档 `/specs/001-hostnet-port-allocation/`  
**Prerequisites**: plan.md, spec.md, data-model.md, research.md, quickstart.md, contracts/http-api.md  
**技术方案**: 见 `specs/001-hostnet-port-allocation/spec.md`

**组织方式**: 任务按用户故事分组，每个阶段可独立实现和测试。US-1/US-2/US-3 共享控制器文件，合并为一个阶段。

## 格式说明: `[ID] [P?] [Story] 描述`

- **[P]**: 可与同阶段其他 [P] 任务并行执行（不同文件、无依赖）
- **[Story]**: 任务所属的用户故事（US1, US2, ... US8），对应 spec.md 中的 User Story

## 路径约定

- 控制器内部文件：相对于工作区根 `bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/`
- CRD 类型文件：`bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1/`（仓库根相对路径）

---

## Phase 1: Setup（项目初始化）

**目的**: 完成 HostNetPortPool CRD 注册和基础常量补充，使 controller-runtime 能识别新的 CRD 类型。

- [X] T001 在 `bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1/hostnetportpool_types.go` 中添加 `init()` 函数，将 `HostNetPortPool` 和 `HostNetPortPoolList` 注册到 `SchemeBuilder`。参考同目录下 `portpool_types.go` 的 `init()` 写法。
- [X] T002 运行 `controller-gen` 工具重新生成 `bcs-runtime/bcs-k8s/kubernetes/apis/networkextension/v1/zz_generated.deepcopy.go`，确保 HostNetPortPool 相关类型具备 `DeepCopyObject()` 方法。
- [X] T003 [P] 在 `internal/constant/constant.go` 中添加 `FinalizerNameHostNetPortPool` 常量，值为 `"hostnetportpool.bkbcs.tencent.com"`。独立于现有 `FinalizerNameBcsIngressController`，避免两套机制并存时冲突。

---

## Phase 2: Foundational（基础设施 — 缓存 + 类型 + 指标）

**目的**: 实现 HostNetPortPoolCache 内存缓存及其全部方法、类型定义、Prometheus 指标定义。这是所有用户故事的前置依赖。

**⚠️ 关键**: 所有用户故事的实现均依赖此阶段完成。

- [X] T004 [P] 创建类型定义文件 `internal/hostnetportpoolcache/types.go`，定义 `HostNetPortPoolBindingResult` 结构体（JSON 序列化用于 Pod annotation，字段: PoolName, PoolNamespace, NodeName, StartPort, EndPort, SegmentLength）和 `ConflictSegment` 结构体（字段: NodeName, StartPort, EndPort, PodKey，用于 UpdatePool 缩小冲突检测返回值）。参考 data-model.md 第 7、8 节的字段定义。
- [X] T005 [P] 创建 Prometheus 指标定义文件 `internal/metrics/hostnetportpool.go`，参考现有 `internal/metrics/` 代码模式，定义以下指标及上报函数：(1) 缓存状态 Gauge `hostnet_segment_allocated` 和 `hostnet_segment_total`，label: `pool_name` + `pool_namespace` + `node_name`；(2) Reconcile 延迟 Histogram `hostnet_reconcile_duration_seconds`；(3) 泄漏段释放 Counter `hostnet_segment_leaked_total`；(4) 分配失败 Gauge `hostnet_allocate_failed`，label: `pod_name` + `pod_namespace`（失败时置 1/成功后置 0/Pod 删除时清除）；(5) 分配失败 Counter `hostnet_allocate_failed_total`，label: `pool_name` + `node_name`；(6) 缓存重建恢复 Gauge `hostnet_cache_rebuild_pods_recovered`。
- [X] T006 创建缓存核心实现文件 `internal/hostnetportpoolcache/cache.go`，实现以下全部内容：(1) 数据结构定义：`HostNetPortPoolSegment`、`NodeSegmentAllocator`、`HostNetPortPoolEntry`、`HostNetPortPoolCache`（含 sync.Mutex），参考 data-model.md 第 3-6 节；(2) 构造函数 `NewHostNetPortPoolCache()`；(3) 生命周期管理方法：`AddPool`、`RemovePool`、`UpdatePool`（含端口范围缩小冲突检测，返回 `[]ConflictSegment`）；(4) 分配与释放方法：`AllocateContiguous`（first-fit 连续段分配算法，失败时返回诊断信息含 maxContiguousFree）、`Release`、`ReleaseByPodKey`（遍历全量缓存按 podKey 释放）；(5) 查询与重建方法：`GetNodeAllocations`（返回 `[]*NodeHostNetPortPoolStatus`）、`RebuildFromPod`；(6) Node 清理方法：`CleanupNode`；(7) 内部辅助方法：`getOrCreateNodeAllocator`（懒加载策略）。详见 iWiki 技术方案 4.3 节。
- [X] T007 编写缓存单元测试 `internal/hostnetportpoolcache/cache_test.go`，使用 table-driven 测试模式，覆盖率 ≥80%。测试场景包括：(1) AddPool/RemovePool 基本生命周期；(2) AllocateContiguous 单段/多段连续分配、first-fit 算法验证；(3) AllocateContiguous 失败场景（总量不足、碎片化导致连续段不足）及诊断信息验证；(4) Release/ReleaseByPodKey 释放后段状态恢复；(5) UpdatePool 扩大/缩小范围（含冲突检测）；(6) CleanupNode 清理后分配器移除；(7) RebuildFromPod 重建后段标记正确；(8) 懒加载验证：首次 Allocate 时自动创建 NodeSegmentAllocator；(9) 并发安全：多 goroutine 并发 Allocate/Release 不 panic。

**检查点**: 缓存核心就绪 — 后续控制器和 API 的实现可以开始。

---

## Phase 3: US-1 + US-2 + US-3 — 控制器核心：Pod 端口段分配、释放与并发安全 (Priority: P1) 🎯 MVP

**目标**: 实现 HostNetPortAllocatorReconciler 控制器，完成 Pod 调度后端口段自动分配（US-1）、Pod 删除/驱逐时端口段回收（US-2）、互斥锁保证并发分配不冲突（US-3）。

**独立测试**: 创建 HostNetPortPool CRD 和带有 `hostnetportpool.networkextension.bkbcs.tencent.com` annotation 的 hostNetwork Pod，验证：(1) Pod 调度后 Controller 自动分配端口段并注入 annotation；(2) Pod 删除后端口段释放，新 Pod 可复用；(3) 多 Pod 同时调度到同一 Node 时获得不同端口段。

### 控制器框架

- [X] T008 [P] [US-1] 创建 Controller 结构定义 `hostnetportcontroller/controller.go`，包含：(1) `HostNetPortAllocatorReconciler` 结构体（字段: ctx, client, cache *HostNetPortPoolCache, eventer record.EventRecorder, cacheSynced bool）；(2) `SetupWithManager` 方法，使用 `For(&HostNetPortPool{})` 监听 CRD，使用 `Watches(&Pod{})` + 自定义 PodFilter 监听带有 hostnetportpool annotation 的 Pod。参考现有 `portbindingcontroller` 的注册模式。RBAC 声明需包含 pods（get/list/watch/patch）和 hostnetportpools（get/list/watch/update/patch）。
- [X] T009 [P] [US-1] 创建 Pod 事件过滤器 `hostnetportcontroller/podfilter.go`，实现 controller-runtime 的 `handler.EventHandler` 接口。Create/Update/Delete 方法中：(1) 检查 Pod annotation 是否包含 `constant.AnnotationForHostNetPortPool` key，无则忽略；(2) Update 事件中通过 `checkHostNetPortPodNeedReconcile` 检查 `spec.nodeName`、`status.phase`、`metadata.deletionTimestamp`、`hostnetportpool.*` annotation 的变更，参考现有 `portbindingcontroller/podfilter.go` 的 `checkPodNeedReconcile` 模式。PodFilter 是纯过滤器，不持有 cache 引用。

### 核心 Reconcile 逻辑

- [X] T010 [US-1] [US-2] [US-3] 实现 Reconcile 主逻辑 `hostnetportcontroller/reconcile.go`，包含：(1) `Reconcile()` 入口方法：先尝试 Get Pod，若 NotFound 则调用 `handlePodDeletion` 释放端口段（US-2）；若存在则调用 `reconcilePod`；(2) `reconcilePod()` 完整 9 步流程（参考 iWiki 技术方案 4.4.4.2 节）：步骤 1 检查请求 annotation → 步骤 2 检查终态 Phase=Failed/Succeeded 则释放（US-2，驱逐场景）→ 步骤 3 检查 Terminating 则跳过 → 步骤 4 检查 nodeName 是否已调度 → 步骤 5 检查是否已分配（幂等）→ 步骤 6 确定 Pool 坐标 → 步骤 7 验证 Pool 存在 → 步骤 8 计算 segmentsNeeded 并调用 cache.AllocateContiguous（加锁保证并发安全，US-3）→ 步骤 9 Patch annotation 注入分配结果，失败则回滚释放；(3) `handlePodDeletion()` 方法：调用 `cache.ReleaseByPodKey` 释放端口段（US-2）；(4) Kubernetes Event 记录：分配成功记录 Normal/HostNetPortAllocated，分配失败记录 Warning/AllocateHostNetPortFailed，释放记录 Normal/HostNetPortReleased；(5) Prometheus 指标上报：Reconcile 入口记录开始时间并在出口上报 Histogram，分配成功/失败时更新 Gauge 和 Counter。

### 集成注册

- [X] T011 [US-1] 在 `main.go` 中注册 `HostNetPortAllocatorReconciler`：(1) 创建 `HostNetPortPoolCache` 实例；(2) 创建 `HostNetPortAllocatorReconciler` 并调用 `SetupWithManager`。参考现有 PortPoolReconciler 和 PortBindingReconciler 的注册方式。

**检查点**: 此时 US-1、US-2、US-3 应已全部可用。Pod 调度后自动分配端口段，删除/驱逐后释放，并发分配无冲突。

---

## Phase 4: US-4 — HTTP API 查询分配结果 (Priority: P1)

**目标**: 提供 HTTP API 供 Pod 中的 initContainer 查询端口段分配结果，实现即时获取分配结果（比 Downward API Volume 延迟低得多）。

**独立测试**: 部署带 initContainer 的 hostNetwork Pod，initContainer 通过 `wget` 请求 HTTP API 轮询分配状态，验证在 Controller 完成分配后立即获取到 `status=Ready` 及完整分配结果。

- [X] T012 [US-4] 创建 HTTP handler 文件 `internal/httpsvr/hostnetportpool.go`，实现 `getHostNetPortPoolBindingResult` handler 方法。(1) 解析 query 参数 `podName` 和 `podNamespace`，缺少则返回 code=400；(2) 通过 controller-runtime informer cache（`httpServerClient.Mgr.GetClient()`）读取 Pod 对象，Pod 不存在或未携带 HostNetPortPool annotation 则返回 code=404；(3) 读取 Pod 的 `AnnotationForHostNetPortPoolBindingStatus` 和 `AnnotationForHostNetPortPoolBindingResult` annotation；(4) 构建响应数据（复用现有 `CreateResponseData` 模式），status=Ready 时解析 result JSON 并返回，status=NotReady/Failed 时 result 为 null。详见 contracts/http-api.md 的响应格式。复用 `ReportAPIRequestMetric` 记录请求计数与延迟。
- [X] T013 [US-4] 在 `internal/httpsvr/httpserver.go` 的 `InitRouters` 方法中注册新路由：`ws.Route(ws.GET("/api/v1/hostnetportpool/bindingresult").To(httpServerClient.getHostNetPortPoolBindingResult))`。
- [X] T014 [US-4] 在 `main.go` 中确保 `HttpServerClient` 创建时已设置必要引用（Mgr 字段已存在），无需额外字段。验证 HTTP Server 启动后新路由可访问。

**检查点**: initContainer 可通过 HTTP API 轮询并获取端口分配结果，完整的 Pod 端口分配流程端到端可用。

---

## Phase 5: US-5 — Controller 重启后缓存重建 (Priority: P2)

**目标**: Controller 进程重启或 Leader 切换后，从现有 Pod annotation 中重建内存缓存，恢复端口段分配状态，避免端口冲突。

**独立测试**: 在已有 Pod 占用端口段的情况下模拟 Controller 重启，验证 initCache 完成后新 Pod 不会分配到已占用的端口段。

- [X] T015 [US-5] 在 `hostnetportcontroller/controller.go` 中实现 `initCache()` 方法，由 Reconcile 首次调用时触发（通过 `cacheSynced` 标志位控制仅执行一次）：(1) List 集群中所有 `HostNetPortPool` CR，对每个 Pool 调用 `cache.AddPool` 创建缓存条目；(2) List 集群中所有 Pod，过滤出带有 `AnnotationForHostNetPortPool` annotation 且 Phase 非 Failed/Succeeded 的 Pod；(3) 对每个符合条件的 Pod，解析其 `AnnotationForHostNetPortPoolBindingResult` annotation，调用 `cache.RebuildFromPod(poolKey, nodeName, podKey, startPort, endPort)` 恢复分配状态；(4) 重建完成后设置 `hostnet_cache_rebuild_pods_recovered` Gauge 并记录日志含重建耗时。参考现有 PortPoolReconciler 的 `isCacheSync` 模式。

**检查点**: Controller 重启后缓存正确重建，已有 Pod 占用的端口段不会被重复分配。

---

## Phase 6: US-6 — HostNetPortPool CRD 配置变更 (Priority: P2)

**目标**: 支持运维人员修改 HostNetPortPool CRD 的端口范围，Controller 正确同步缓存；缩小范围时检测冲突段并阻止操作。

**独立测试**: 修改 HostNetPortPool CRD 的 `endPort` 字段，验证扩大时新段立即可用，缩小时有冲突的操作被拒绝并在 HostNetPortPool CR 上记录 Warning Event（Reason: `PoolShrinkConflict`）。

- [X] T016 [US-6] 在 `hostnetportcontroller/reconcile.go` 中实现 `reconcilePool()` 方法：(1) 检查 HostNetPortPool CR 是否为新创建（无 Finalizer），若是则添加 `FinalizerNameHostNetPortPool` Finalizer；(2) 若 DeletionTimestamp 已设（CR 被删除），检查是否有 Pod 仍在使用该 Pool 的端口——有则拒绝删除并在 Status 中提示，无则移除 Finalizer 并调用 `cache.RemovePool`；(3) 调用 `cache.UpdatePool()` 同步最新配置，若返回冲突段列表（缩小范围时），拒绝缩小并在 HostNetPortPool CR 上记录 Warning Event（Reason: `PoolShrinkConflict`，Message 含冲突的 Node、端口范围和 Pod），RequeueAfter 30s；冲突解决后记录 Normal Event（Reason: `PoolShrinkResolved`）；(4) 调用 `cache.GetNodeAllocations()` 获取各 Node 分配概况并更新 Status.NodeAllocations。
- [X] T017 [US-6] 修改 `hostnetportcontroller/reconcile.go` 中 `Reconcile()` 入口方法，增加 HostNetPortPool CR 资源类型的识别：尝试 `Get HostNetPortPool`，若成功则调用 `reconcilePool()`。

**检查点**: HostNetPortPool CRD 配置变更正确同步到缓存，Finalizer 保护有 Pod 使用的 Pool 不被误删，缩小范围冲突被检测并报告。

---

## Phase 7: US-7 — 集群扩缩容支持 (Priority: P2)

**目标**: 新 Node 加入集群时通过懒加载机制自动支持端口分配；Node 被移除时清理缓存数据。

**独立测试**: (1) 向集群添加新 Node 并调度 Pod 到该 Node，验证 Controller 懒创建分配器并成功分配端口段；(2) 执行 `kubectl drain` 驱逐 Pod 后删除 Node，验证 Pod 端口段释放、Node 缓存清理。

- [X] T018 [P] [US-7] 创建 Node 事件过滤器 `hostnetportcontroller/nodefilter.go`，实现 controller-runtime 的 `handler.EventHandler` 接口。仅关注 Delete 事件（Node 被移除），将 Node 名称映射为 Reconcile 请求时使用 `__node__` namespace 前缀（如 `{Namespace: "__node__", Name: nodeName}`）以便 Reconcile 入口区分资源类型。Create/Update 事件忽略（新 Node 通过懒加载支持，属性变更与端口分配无关）。参考现有 `portbindingcontroller/nodefilter.go` 模式。
- [X] T019 [US-7] 在 `hostnetportcontroller/reconcile.go` 中实现 Node 事件处理：Reconcile 入口识别 `__node__` namespace 前缀后调用 `cache.CleanupNode(nodeName)`，从所有端口池的 NodeAllocators map 中删除该 Node 的条目。
- [X] T020 [US-7] 修改 `hostnetportcontroller/controller.go` 的 `SetupWithManager`，增加 `Watches(&Node{})` + 自定义 NodeFilter 注册 Node 事件监听。

**检查点**: 集群扩容时新 Node 自动支持端口分配，缩容时 Node 缓存正确清理，无内存泄漏。

---

## Phase 8: US-8 — 端口泄漏定期巡检 (Priority: P3)

**目标**: 配备定期扫描检查器作为事件驱动释放的兜底机制，防止极端情况下的端口泄漏。

**独立测试**: 模拟事件丢失场景（缓存中有已分配段但对应 Pod 已不存在），验证定期扫描检查器最终释放孤立段。

- [X] T021 [US-8] 创建端口泄漏检查器 `internal/check/hostnet_segment_checker.go`，实现 `check.Checker` 接口的 `Run()` 方法：(1) 遍历 HostNetPortPoolCache 中所有已分配段；(2) 对每个已分配段，通过 K8s API 检查对应的 Pod 是否仍然存在且处于活跃状态；(3) 若 Pod 不存在（Get 返回 NotFound）或已处于终态（Phase=Failed/Succeeded），调用 `cache.ReleaseByPodKey` 释放端口段；(4) 释放时递增 `hostnet_segment_leaked_total` Counter 并更新缓存状态 Gauge；(5) 记录日志含泄漏段详情（Node、端口范围、PodKey）。参考现有 `PortLeakChecker` 的实现模式。
- [X] T022 [US-8] 在 `main.go` 中创建 `HostNetPortSegmentChecker` 实例并注册到 `CheckRunner`，参考现有 `PortLeakChecker` 的注册方式。

**检查点**: 定期巡检机制就绪，孤立端口段可被最终释放，配合 Prometheus Counter 可设置告警。

---

## Phase 9: Refactor（澄清驱动的架构重构）

**目的**: 基于 Clarifications Session 2026-04-09 的决策，重构现有代码以提升架构正确性和可维护性。

### 9.1 Controller 拆分（替代 `__node__` 前缀判断）

- [ ] T026 创建 `hostnetportcontroller/pool_controller.go`，实现 `HostNetPortPoolReconciler` 结构体和 `Reconcile()` 方法，处理 HostNetPortPool CR 生命周期（Finalizer 管理、端口范围变更、Status 更新）。从现有 `controller.go` 中提取 `reconcilePool()` 逻辑。
- [ ] T027 创建 `hostnetportcontroller/pod_controller.go`，实现 `PodReconciler` 结构体和 `Reconcile()` 方法，处理 Pod 端口分配与释放。从现有 `controller.go` 中提取 Pod 相关的 reconcile 逻辑。
- [ ] T028 创建 `hostnetportcontroller/node_controller.go`，实现 `NodeReconciler` 结构体和 `Reconcile()` 方法，处理 Node 删除时的缓存清理。从现有 `controller.go` 中提取 Node 事件处理逻辑（移除 `__node__` namespace 前缀判断）。
- [ ] T029 修改 `main.go`，注册三个独立 Reconciler：分别创建 `HostNetPortPoolReconciler`、`PodReconciler`、`NodeReconciler` 实例并调用各自的 `SetupWithManager()`。
- [ ] T030 删除 `hostnetportcontroller/controller.go` 中的 `__node__` namespace 前缀判断逻辑和相关常量。
- [ ] T031 更新 `nodefilter.go`（如存在），移除 Node 事件到 `__node__` namespace 的映射，改为直接触发 `NodeReconciler` 的 Reconcile。

### 9.2 幂等性检查改为 Cache 查询

- [ ] T032 在 `internal/hostnetportpoolcache/cache.go` 中新增 `IsPodAllocated(podKey string) bool` 方法，基于 `NodeSegmentAllocator.allocatedSegments` 映射查询 Pod 是否已分配端口段。
- [ ] T033 修改 `hostnetportcontroller/pod_controller.go`（或现有 controller）的 Reconcile 逻辑：步骤 5（检查是否已分配）改为调用 `cache.IsPodAllocated(podKey)`，而非检查 Pod Annotation `AnnotationForHostNetPortPoolBindingResult`。
- [ ] T034 更新 `hostnetportcontroller/podfilter.go`：移除或简化 Update 事件中关于 `AnnotationForHostNetPortPoolBindingResult` 变更的检查（因为幂等性不再依赖 Annotation）。

### 9.3 非法 portcount 直接报错

- [ ] T035 修改 `hostnetportcontroller/pod_controller.go` 中的端口段计算逻辑：当 `portcount` 解析失败或 <=0 时，(1) 调用 `patchPodBindingStatus(pod, "Failed")`；(2) 记录 Warning Event（Reason: `InvalidPortCount`）；(3) 上报分配失败指标；(4) 返回 `ctrl.Result{}, nil`（不再重试）。移除使用默认值 1 个段的降级逻辑。
- [ ] T036 更新 spec.md 中的 Acceptance Scenario 以匹配新行为（如文档尚未同步）。

### 9.4 端口池缩小冲突 Counter 指标

- [ ] T037 在 `internal/metrics/hostnetportpool.go` 中新增 Counter 指标 `hostnet_pool_shrink_conflict_total`，labels: `pool_name`, `pool_namespace`, `node_name`。
- [ ] T038 修改 `hostnetportcontroller/pool_controller.go` 中的 `reconcilePool()` 逻辑：当 `cache.UpdatePool()` 返回冲突段列表时，对每个冲突递增 `hostnet_pool_shrink_conflict_total` 计数器（labels 包含 Pool 和冲突 Node）。
- [ ] T039 在 `quickstart.md` 中添加 PromQL 告警示例（如尚未添加）。

### 9.5 代码命名和逻辑简化

- [ ] T040 重命名 `hostnetportcontroller/controller.go`（或拆分后的文件）中的 `patchPodStatus` 函数为 `patchPodBindingStatus`，以更准确地反映其功能是 patch Pod annotation 而非 Pod status 子资源。更新所有调用点。
- [ ] T041 简化 `hostnetportcontroller/podfilter.go` 第 119 行（或对应 Update 事件处理）：由于 Reconcile 逻辑中已处理 `DeletionTimestamp != nil` 的跳过逻辑，PodFilter 的 `checkHostNetPortPodNeedReconcile` 可以简化为 `newPod.DeletionTimestamp == nil`。

## Phase 10: Polish（完善与交叉关注点）

**目的**: 跨用户故事的改进和验证。

- [X] T042 [P] Prometheus 指标集成验证：确认分配/释放/失败/泄漏/重建/缩小冲突等场景下各指标数值正确更新
- [X] T043 代码清理：gofmt/goimports 格式化、GoDoc 英文注释完整性检查、lint 错误修复
- [X] T044 运行 `quickstart.md` 端到端验证：按照 quickstart.md 步骤创建 HostNetPortPool → 部署 hostNetwork Pod → 验证分配结果 → 查看端口池使用情况 → 验证非法 portcount 报错行为 → 验证缩小冲突指标

---

## 依赖关系与执行顺序

### 阶段依赖

- **Setup（Phase 1）**: 无依赖 — 可立即开始
- **Foundational（Phase 2）**: 依赖 Setup 完成 — **阻塞所有用户故事**
- **US-1+US-2+US-3（Phase 3）**: 依赖 Phase 2 完成 — **核心 MVP**
- **US-4（Phase 4）**: 依赖 Phase 2 完成 — **可与 Phase 3 并行开发**
- **US-5（Phase 5）**: 依赖 Phase 3 完成（需要 Controller 框架就绪）
- **US-6（Phase 6）**: 依赖 Phase 3 完成（需要 Reconcile 入口就绪）
- **US-7（Phase 7）**: 依赖 Phase 3 完成（需要 Controller 框架就绪）
- **US-8（Phase 8）**: 依赖 Phase 2 完成 — **可与 Phase 3 并行开发**
- **Refactor（Phase 9）**: 依赖 Phase 3 完成（基于现有 Controller 代码重构）
- **Polish（Phase 10）**: 依赖 Phase 9 完成

### 用户故事依赖

- **US-1 (P1)**: 基础 — 无其他故事依赖
- **US-2 (P1)**: 与 US-1 紧密耦合，共享 reconcile.go 文件，同阶段实现
- **US-3 (P1)**: 通过 cache 的 sync.Mutex 保障，在 Phase 2 缓存实现中已覆盖
- **US-4 (P1)**: 独立 — 仅依赖 Phase 2 的类型定义和 Mgr client
- **US-5 (P2)**: 依赖 US-1 的 Controller 框架
- **US-6 (P2)**: 依赖 US-1 的 Reconcile 入口，可与 US-5/US-7 并行
- **US-7 (P2)**: 依赖 US-1 的 Controller 框架，可与 US-5/US-6 并行
- **US-8 (P3)**: 独立 — 仅依赖 Phase 2 的缓存实现

### 依赖图

```
Phase 1 (Setup)
    │
    v
Phase 2 (Foundational: 缓存 + 类型 + 指标)
    │
    ├──────────────────┬──────────────────┐
    v                  v                  v
Phase 3 (US-1/2/3)  Phase 4 (US-4)    Phase 8 (US-8)
    │                  │                  │
    ├─────┬─────┐      │                  │
    v     v     v      │                  │
  Ph.5  Ph.6  Ph.7    │                  │
  (US5) (US6) (US7)   │                  │
    │     │     │      │                  │
    └─────┴─────┴──────┴──────────────────┘
                  │
                  v
          Phase 9 (Refactor)
                  │
                  v
          Phase 10 (Polish)
```

### 并行机会

- Phase 1 中 T001 和 T003 可并行（不同文件）
- Phase 2 中 T004 和 T005 可并行（不同文件，无依赖）
- Phase 3 中 T008 和 T009 可并行（不同文件）
- **Phase 3 和 Phase 4 可并行开发**（控制器和 HTTP API 无直接依赖，共同依赖 Phase 2）
- **Phase 3 和 Phase 8 可并行开发**（检查器仅依赖缓存，不依赖控制器）
- Phase 5、Phase 6、Phase 7 可并行开发（不同文件或文件内不同方法，均独立实现）
- **Phase 9 重构任务可并行**（不同 Reconciler 拆分独立进行，指标添加独立进行）

---

## 并行执行示例

```bash
# Phase 2 中的并行任务:
Task: "创建类型定义 internal/hostnetportpoolcache/types.go"        # T004
Task: "创建 Prometheus 指标定义 internal/metrics/hostnetportpool.go" # T005

# Phase 3 中的并行任务:
Task: "创建 Controller 结构定义 hostnetportcontroller/controller.go"  # T008
Task: "创建 Pod 事件过滤器 hostnetportcontroller/podfilter.go"       # T009

# Phase 3 和 Phase 4 并行:
Task: "实现 Reconcile 主逻辑 hostnetportcontroller/reconcile.go"     # T010 (Phase 3)
Task: "创建 HTTP handler internal/httpsvr/hostnetportpool.go"       # T012 (Phase 4)

# Phase 9 重构中的并行任务:
Task: "创建 Pool/Pod/Node Controller 拆分文件"                         # T026/T027/T028
Task: "添加 IsPodAllocated Cache 方法"                               # T032
Task: "添加缩小冲突 Counter 指标"                                    # T037
Task: "重命名 patchPodStatus 函数"                                   # T040

```bash
# Phase 2 中的并行任务:
Task: "创建类型定义 internal/hostnetportpoolcache/types.go"        # T004
Task: "创建 Prometheus 指标定义 internal/metrics/hostnetportpool.go" # T005

# Phase 3 中的并行任务:
Task: "创建 Controller 结构定义 hostnetportcontroller/controller.go"  # T008
Task: "创建 Pod 事件过滤器 hostnetportcontroller/podfilter.go"       # T009

# Phase 3 和 Phase 4 并行:
Task: "实现 Reconcile 主逻辑 hostnetportcontroller/reconcile.go"     # T010 (Phase 3)
Task: "创建 HTTP handler internal/httpsvr/hostnetportpool.go"       # T012 (Phase 4)
```

---

## 实施策略

### MVP 优先（Phase 1 → 2 → 3 → 4）

1. 完成 Phase 1: Setup（CRD 注册 + 常量）
2. 完成 Phase 2: Foundational（缓存 + 类型 + 指标）
3. 完成 Phase 3: US-1+US-2+US-3（控制器核心 — 分配/释放/并发）
4. 完成 Phase 4: US-4（HTTP API — initContainer 查询）
5. **停下验证**: 运行 quickstart.md 端到端测试，确认核心流程可用
6. 部署/演示 MVP

### 增量交付

1. Setup + Foundational → 基础设施就绪
2. 添加 US-1/2/3 → 端口段分配和释放可用 → MVP
3. 添加 US-4 → initContainer 可获取分配结果 → 完整端到端可用
4. 添加 US-5 → Controller 重启安全
5. 添加 US-6 → CRD 配置可变更
6. 添加 US-7 → 集群弹性伸缩支持
7. 添加 US-8 → 泄漏兜底机制
8. **架构重构 (Phase 9)** → 基于 Clarifications 改进代码质量：
   - Controller 拆分为三个独立 Reconciler
   - 幂等性检查改为 Cache 查询
   - 非法 portcount 直接报错
   - 缩小冲突添加 Counter 指标
   - 代码命名修正（patchPodBindingStatus）
   - 逻辑简化（podfilter DeletionTimestamp 检查）
9. **Polish (Phase 10)** → 最终验证和代码清理
10. 每个故事独立增加价值且不破坏已有功能

---

## 备注

- [P] 任务 = 不同文件、无依赖，可并行
- [Story] 标签将任务映射到 spec.md 中的具体用户故事
- 每个用户故事应可独立完成和测试
- 每个任务或逻辑分组完成后提交 commit
- 在任何检查点处可停下验证已完成的故事
- 避免：模糊任务、同文件冲突、破坏独立性的跨故事依赖
