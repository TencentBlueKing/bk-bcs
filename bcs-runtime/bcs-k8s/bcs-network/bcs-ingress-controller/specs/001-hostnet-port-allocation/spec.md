# Feature Specification: HostNetwork 动态端口分配

**Feature Branch**: `001-hostnet-port-allocation`  
**Created**: 2026-03-16  
**Status**: Draft  
**Input**: 基于 iWiki 技术方案和已完成的 HostNetPortPool CRD 类型定义，实现 Pod 使用 hostNetwork 网络模式时，IngressController 的动态端口分配逻辑。

## Clarifications

### Session 2026-03-16

- Q: 缓存状态 Gauge 指标（已分配段数/总段数）应使用什么 label 粒度？ → A: `pool_name` + `pool_namespace` + `node_name`（per-pool-per-node）
- Q: 端口段分配操作是否需要独立的延迟 Histogram 指标？ → A: 记录 Reconcile 整体延迟（含 K8s API 读写）的 Histogram，不单独记录纯内存分配操作延迟
- Q: 定期检查器发现泄漏段时是否需要独立的 Prometheus 指标？ → A: 需要，新增 Counter 记录泄漏段释放总数（便于告警）
- Q: 端口段分配失败时是否需要类似 portAllocateFailedGauge 的指标？ → A: 两者都要——Gauge（label: `pod_name` + `pod_namespace`）标记失败 Pod + Counter（label: `pool_name` + `node_name`）记录失败总次数
- Q: 缓存重建完成时是否需要指标记录重建耗时和恢复的 Pod 数量？ → A: 重建耗时用日志记录（仅发生一次），恢复的 Pod 数量作为 Gauge 暴露（验证重建正确性）
- Q: Spec 中 "bindingresult" / "bindingstatus" 术语是否与实际 annotation key 一致？ → A: 不一致。实际 annotation key 的区分词为 `result`（`hostnetportpool.result.xxx`）和 `status`（`hostnetportpool.status.xxx`），spec 中应统一使用 `result` / `status` annotation 而非 `bindingresult` / `bindingstatus`
- Q: HostNetPortPool 是否应复用现有 `FinalizerNameBcsIngressController`？ → A: 不应复用。复用会导致同时使用 HostNetPortPool 和 PortPool 时 Finalizer 冲突，需新增独立常量 `FinalizerNameHostNetPortPool`（值：`hostnetportpool.bkbcs.tencent.com`），遵循 FR-017 独立子系统原则

### Session 2026-03-17

- Q: HostNetPortPool CRD Status 中是否应使用 `metav1.Condition` 记录端口缩小冲突等错误信息？ → A: 不应使用。`bcs-network` 的 `go.mod` 将 `k8s.io/apimachinery` replace 为 `v0.18.6`，该版本不支持 `metav1.Condition`（v0.19.0 引入）。所有冲突/错误信息改为通过 Kubernetes Warning Event 记录在 HostNetPortPool CR 上。CRD Status 中移除 `Conditions` 字段，仅保留 `Status` 和 `NodeAllocations`。

### Session 2026-04-09

- Q: Controller 是否应使用 `__node__` 特殊 Namespace 判断 Node 删除事件，还是创建多个 Reconcile() 方法？ → A: 拆分三个独立 Reconciler：`HostNetPortPoolReconciler`、`PodReconciler`、`NodeReconciler`，各自注册到 Manager，避免使用特殊 Namespace 前缀判断事件类型
- Q: 判断 Pod 是否已分配端口时，应使用 Pod Annotation 还是 Cache？ → A: 使用 Cache 查询（`cache.IsPodAllocated(podKey)`），而非 Pod Annotation，避免 APIServer 压力大时 Annotation 更新延迟导致读取旧值
- Q: 非法 portCount（解析失败或 <=0）时应使用默认值还是直接报错？ → A: 直接报错：标记分配失败，设置 status=Failed，记录 Warning Event，不使用默认值
- Q: 端口池缩小时冲突的 Metric 应使用 Counter 还是 Gauge？ → A: Counter：`hostnet_pool_shrink_conflict_total`，每次冲突 +1，labels: pool_name, pool_namespace, node_name

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Pod 自动获取 Node 端口段 (Priority: P1)

运维人员在集群中配置好 HostNetPortPool CRD（声明端口范围和段长度），然后部署使用 `hostNetwork` 模式的 GameWorkload Pod 并通过 annotation 指定端口池名称。Pod 被调度到某个 Node 后，IngressController 自动为该 Pod 在其所在 Node 上分配一段可用的连续端口，并将分配结果注入 Pod 的 annotation。Pod 中的 initContainer 通过 HTTP API 获取分配结果后，业务容器启动并使用分配到的端口。

**Why this priority**: 这是整个功能的核心价值——让 hostNetwork Pod 无需预先声明 `containerPort`，由 Controller 在调度后动态分配端口段，解决 DSHosting 海外非自研云场景下 Pod 直通 Node 端口的需求。

**Independent Test**: 创建 HostNetPortPool CRD 和带有对应 annotation 的 hostNetwork Pod，验证 Pod 调度后 Controller 自动分配端口段并注入 annotation，initContainer 通过 HTTP API 成功获取分配结果。

**Acceptance Scenarios**:

1. **Given** 集群中存在 HostNetPortPool CRD（startPort=30000, endPort=30100, segmentLength=10），**When** 用户创建带有 `hostnetportpool.networkextension.bkbcs.tencent.com: "game-server-ports"` annotation 的 hostNetwork Pod 且 Pod 被调度到 node-1，**Then** Controller 为该 Pod 分配一个端口段（如 30000-30009），并将 result annotation（分配结果 JSON）和 status annotation（=Ready）注入 Pod annotation。
2. **Given** Pod 已调度但未指定 `portcount` annotation，**When** Controller 处理该 Pod，**Then** 默认分配 1 个段（segmentLength 个端口）。
3. **Given** Pod 指定 `portcount=25` 且 segmentLength=10，**When** Controller 处理该 Pod，**Then** 分配 ceil(25/10)=3 个连续段（共 30 个端口）。
4. **Given** Pod 指定非法 `portcount`（如非数字或 <=0），**When** Controller 处理该 Pod，**Then** 标记分配失败（status=Failed），记录 Warning Event，不分配端口段。
4. **Given** Pod 尚未被调度（nodeName 为空），**When** Controller 收到该 Pod 的事件，**Then** 跳过处理，等待调度完成后再处理。

---

### User Story 2 - Pod 删除与驱逐时端口段回收 (Priority: P1)

当 Pod 被删除、驱逐或进入终态（Failed/Succeeded）时，Controller 自动释放该 Pod 占用的端口段，使其可被后续新 Pod 复用。

**Why this priority**: 端口段回收与分配同等重要——如果端口段不能正确回收，会导致可用端口逐渐耗尽，新 Pod 无法获取端口。

**Independent Test**: 在已分配端口段的 Pod 上执行删除/驱逐操作，验证 Controller 释放端口段，新 Pod 可以复用该端口段。

**Acceptance Scenarios**:

1. **Given** Pod-A 在 node-1 上占用端口段 30000-30009，**When** Pod-A 被主动删除且从 API Server 中完全移除，**Then** Controller 释放端口段 30000-30009，新 Pod 调度到 node-1 时可分配到该段。
2. **Given** Pod-A 在 node-1 上占用端口段 30000-30009，**When** Pod-A 被驱逐（Phase 变为 Failed），**Then** Controller 立即释放端口段，不等待 Pod 对象被 GC 清理。
3. **Given** Pod-A 处于 Terminating 状态（DeletionTimestamp 已设但容器仍在运行），**When** Controller 收到该 Pod 的事件，**Then** 不释放端口段（避免新 Pod 复用该端口导致冲突），等到 Pod 完全移除后释放。
4. **Given** Pod-A 占用多个连续段（如 3 个段，30000-30029），**When** Pod-A 被删除，**Then** Controller 释放该 Pod 名下的所有段。

---

### User Story 3 - 同 Node 并发分配不冲突 (Priority: P1)

多个 Pod 同时调度到同一个 Node 时，Controller 通过互斥锁保证端口段分配不会冲突——每个端口段在同一 Node 上只会被分配给一个 Pod。

**Why this priority**: 端口冲突会导致 Pod 启动失败、业务中断，是功能正确性的基本保障。

**Independent Test**: 同时创建多个 hostNetwork Pod 并调度到同一 Node，验证各 Pod 获得不同的端口段，无重叠。

**Acceptance Scenarios**:

1. **Given** node-1 上有 10 个可用端口段，**When** Pod-A 和 Pod-B 同时调度到 node-1 并触发 Reconcile，**Then** 两个 Pod 获得不同的端口段（如 Pod-A 获得 30000-30009，Pod-B 获得 30010-30019）。
2. **Given** 不同 Node 使用相同的 HostNetPortPool，**When** Pod-A 调度到 node-1 和 Pod-B 调度到 node-2，**Then** 两个 Pod 可以获得相同编号的端口段（如都获得 30000-30009），因为 Node 间端口独立。

---

### User Story 4 - HTTP API 查询分配结果 (Priority: P1)

Pod 中的 initContainer 通过 HTTP API 请求 IngressController Service 查询端口分配结果，轮询等待分配完成后再启动业务容器。

**Why this priority**: 由于分配发生在 Pod 调度之后，annotation 在 Pod 创建时可能尚不存在，HTTP API 是业务容器获取端口信息的唯一可靠方式，且比 Downward API Volume 延迟低得多。

**Independent Test**: 部署带 initContainer 的 Pod，initContainer 通过 HTTP API 轮询分配状态，验证在 Controller 完成分配后立即获取到结果。

**Acceptance Scenarios**:

1. **Given** Pod 已调度且 Controller 完成端口分配，**When** initContainer 请求 `GET /ingresscontroller/api/v1/hostnetportpool/bindingresult?podName=<name>&podNamespace=<ns>`，**Then** 返回 `status=Ready` 及包含 poolName、nodeName、startPort、endPort、segmentLength 的分配结果。
2. **Given** Pod 已调度但 Controller 尚未完成分配，**When** initContainer 请求该 API，**Then** 返回 `status=NotReady`，initContainer 继续轮询。
3. **Given** Pod 未携带 HostNetPortPool annotation 或 Pod 不存在，**When** 请求该 API，**Then** 返回 HTTP 404。

---

### User Story 5 - Controller 重启后缓存重建 (Priority: P2)

Controller 进程重启或 Leader 切换后，自动从现有 Pod 的 annotation 中重建内存缓存，恢复端口段分配状态，确保不会将已占用的端口段重复分配给新 Pod。

**Why this priority**: 缓存重建保障 Controller 高可用——重启后不丢失状态，避免端口冲突。

**Independent Test**: 在已有 Pod 占用端口段的情况下重启 Controller，验证重建后新 Pod 不会分配到已占用的端口段。

**Acceptance Scenarios**:

1. **Given** Pod-A 占用 node-1 上的段 30000-30009，**When** Controller 重启并完成缓存重建，**Then** 新 Pod 调度到 node-1 时分配的是下一个可用段（如 30010-30019），不会与 Pod-A 冲突。
2. **Given** Pod-A 在 Controller 重启前已被删除，**When** Controller 重启并重建缓存，**Then** Pod-A 占用的端口段自然不在缓存中（因为 Pod 已不存在），处于可用状态。

---

### User Story 6 - HostNetPortPool CRD 配置变更 (Priority: P2)

运维人员修改 HostNetPortPool CRD 的端口范围时，Controller 正确同步缓存。扩大范围时追加新段；缩小范围时，如果被移除范围内存在已分配的段，拒绝缩小并报告冲突信息。

**Why this priority**: 支持运维人员灵活调整端口池配置，同时保护正在使用的端口段不被误删。

**Independent Test**: 修改 HostNetPortPool CRD 的 endPort 字段，验证 Controller 正确扩大/拒绝缩小端口范围并更新 Status。

**Acceptance Scenarios**:

1. **Given** HostNetPortPool endPort=30100，**When** 运维人员将 endPort 改为 30200，**Then** Controller 为各 Node 追加新段，新 Pod 可使用 30100-30199 范围的端口。
2. **Given** HostNetPortPool endPort=30100 且 node-1 上段 30050-30059 正在被 Pod-X 使用，**When** 运维人员将 endPort 改为 30050，**Then** Controller 拒绝缩小操作，在 HostNetPortPool CR 上记录 Warning Event（Reason: `PoolShrinkConflict`），列出冲突的 Node、端口范围和 Pod。
3. **Given** 缩小操作被拒绝后冲突 Pod 被删除，**When** Controller 重新检查，**Then** 执行缩小操作并记录 Normal Event（Reason: `PoolShrinkResolved`）。

---

### User Story 7 - 集群扩缩容支持 (Priority: P2)

新 Node 加入集群时，Controller 通过懒加载机制自动支持；Node 被移除时，Controller 清理该 Node 对应的缓存数据。

**Why this priority**: 保障集群弹性伸缩场景下端口分配功能的正常运作。

**Independent Test**: 向集群添加新 Node 并调度 Pod 到该 Node，验证分配正常；移除 Node 后验证缓存被清理。

**Acceptance Scenarios**:

1. **Given** 集群新增 node-3，**When** Pod 首次调度到 node-3，**Then** Controller 懒创建 node-3 的分配器并成功分配端口段。
2. **Given** node-2 上有 Pod 占用端口段，**When** 执行 `kubectl drain node-2` 驱逐 Pod 后删除 Node，**Then** Pod 端口段在驱逐/删除时释放，Node 从缓存中移除。

---

### User Story 8 - 端口泄漏定期巡检 (Priority: P3)

Controller 配备定期扫描检查器（HostNetPortSegmentChecker），周期性检查缓存中所有已分配段对应的 Pod 是否仍存在，释放孤立段，作为事件驱动释放的兜底机制。

**Why this priority**: 作为安全网防止极端情况下的端口泄漏（如事件丢失、Controller 重启期间 Pod 被删除等），提升系统稳健性。

**Independent Test**: 模拟事件丢失场景（如删除 Pod 但跳过事件处理），验证定期扫描检查器最终释放孤立的端口段。

**Acceptance Scenarios**:

1. **Given** 缓存中段 30000-30009 标记为 Pod-A 已分配，但 Pod-A 已不存在于集群中，**When** 定期扫描检查器执行，**Then** 释放该段。
2. **Given** 缓存中段 30000-30009 标记为 Pod-A 已分配且 Pod-A Phase=Failed，**When** 定期扫描检查器执行，**Then** 释放该段。

---

### Edge Cases

- Pod 在调度完成到 Controller 分配端口之间被删除时如何处理？（Controller 应在 Reconcile 中检测 Pod 不存在后跳过分配）
- 同一 Pod 重复触发 Reconcile 时如何避免重复分配？（通过查询 Cache 检查 Pod 是否已分配，即 `cache.IsPodAllocated(podKey)`，而非依赖 Pod Annotation）
- Patch annotation 失败时如何保证缓存一致性？（必须回滚：释放刚分配的端口段）
- HostNetPortPool CRD 不存在时 Pod 请求端口段如何处理？（设置 status annotation=Failed 并记录 Warning Event）
- 端口段碎片化导致连续段不足但总空闲段充足时如何提示用户？（分配失败信息包含最大连续空闲段数诊断信息）
- Node 短暂离线后恢复时缓存是否受影响？（不影响——只有 Node 对象被删除才触发缓存清理）
- Controller 不可用期间 Pod 被删除或驱逐，端口段如何释放？（定期扫描检查器和启动时缓存重建机制兜底）

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 提供 HostNetPortPool CRD，支持配置端口范围（startPort、endPort）和最小分配单位（segmentLength），作为端口段分配的资源池。
- **FR-002**: 系统 MUST 在 Pod 调度到 Node 后（spec.nodeName 非空）才执行端口段分配，不在 Pod 创建阶段分配。
- **FR-003**: 系统 MUST 将分配结果（poolName、poolNamespace、nodeName、startPort、endPort、segmentLength）和分配状态（Ready/NotReady/Failed）通过 Patch 操作注入 Pod 的 annotation。
- **FR-004**: 系统 MUST 保证同一 Node 上的同一端口段不会被重复分配给不同 Pod（通过互斥锁保护）；幂等性检查 MUST 通过查询内存 Cache（而非 Pod Annotation）判断 Pod 是否已分配，避免 APIServer 延迟导致竞态问题。
- **FR-005**: 系统 MUST 在 Pod 完全删除（从 API Server 移除）或进入终态（Phase=Failed/Succeeded）后释放其占用的端口段。
- **FR-006**: 系统 MUST NOT 在 Pod 处于 Terminating 状态（DeletionTimestamp 已设但容器仍在运行）时释放端口段。
- **FR-007**: 系统 MUST 支持 Pod 请求多个连续段（通过 portcount annotation），按 ceil(portCount / segmentLength) 个连续段分配；当 portcount 解析失败或小于等于 0 时，系统 MUST 标记分配失败（status=Failed），记录 Warning Event，不使用默认值。
- **FR-008**: 系统 MUST 提供 HTTP API（`GET /ingresscontroller/api/v1/hostnetportpool/bindingresult`），供 Pod 内 initContainer 查询端口分配结果。
- **FR-009**: 系统 MUST 在 Controller 重启后从现有 Pod 的 annotation 中重建内存缓存，恢复端口段分配状态。
- **FR-010**: 系统 MUST 为 HostNetPortPool CRD 添加 Finalizer 删除保护，确保有 Pod 使用时不会被误删。
- **FR-011**: 系统 MUST 在 HostNetPortPool 端口范围缩小且被移除范围内存在已分配段时，拒绝缩小操作并通过 Warning Event（Reason: `PoolShrinkConflict`）在 HostNetPortPool CR 上报告冲突信息（含冲突的 Node、端口范围和 Pod）；同时 MUST 递增 Counter 指标 `hostnet_pool_shrink_conflict_total`（labels: pool_name, pool_namespace, node_name）记录冲突次数。
- **FR-012**: 系统 MUST 在 Node 从集群中移除时清理该 Node 对应的缓存数据。
- **FR-013**: 系统 MUST 配备定期扫描检查器，周期性检查已分配段对应的 Pod 是否仍存在，释放孤立段作为兜底机制。
- **FR-014**: 系统 MUST 在端口段分配成功/失败时记录 Kubernetes Event 到 Pod 上。
- **FR-015**: 系统 MUST 在 HostNetPortPool 的 Status 中维护各 Node 的分配概况（nodeName、allocatedCount、totalSegments）。
- **FR-016**: 系统 MUST 在 Patch annotation 失败时回滚已分配的端口段，保证缓存与实际状态一致。
- **FR-017**: 新功能 MUST 作为独立子系统集成，使用独立的 CRD、缓存、Controller 和 Annotation 前缀，与现有 PortPool 机制互不干扰；Controller 设计 MUST 采用三个独立 Reconciler（`HostNetPortPoolReconciler`、`PodReconciler`、`NodeReconciler`）分别处理不同资源类型，避免使用特殊 Namespace 前缀判断事件类型。
- **FR-018**: 系统 MUST 为所有关键操作暴露 Prometheus 指标，参考现有 `internal/metrics/` 代码模式，包括但不限于：事件过滤器计数（复用 `IncreaseEventCounter`）、HTTP API 请求计数与延迟（复用 `ReportAPIRequestMetric`）、端口段分配/释放操作指标、缓存状态 Gauge（label 维度：`pool_name` + `pool_namespace` + `node_name`，暴露已分配段数和总段数）、Reconcile 整体延迟 Histogram（含 K8s API 读写，不单独记录纯内存分配操作延迟）、Reconcile 失败计数（复用 `IncreaseFailMetric`）、泄漏段释放 Counter（`hostnet_segment_leaked_total`，由定期检查器在发现并释放泄漏段时递增，便于告警）、分配失败 Gauge（label: `pod_name` + `pod_namespace`，失败时置 1 / 成功后置 0 / Pod 删除时清除，与 `portAllocateFailedGauge` 一致）+ 分配失败 Counter（label: `pool_name` + `node_name`，记录分配失败总次数）、端口池缩小冲突 Counter（`hostnet_pool_shrink_conflict_total`，labels: pool_name, pool_namespace, node_name，每次冲突 +1）、缓存重建恢复 Pod 数量 Gauge（重建耗时通过日志记录）。

### Key Entities

- **HostNetPortPool**: 端口池资源，定义可分配的端口范围和段长度。集群级别，Namespace 隔离。关键属性：startPort、endPort、segmentLength。
- **端口段（Segment）**: 分配的最小单位，由 segmentLength 个连续端口组成。Per-Node 隔离——同一端口段在不同 Node 上独立分配。
- **Node 分配器（NodeSegmentAllocator）**: 管理某个 Node 上所有端口段的分配状态，按需懒创建。
- **分配结果（BindingResult）**: 记录 Pod 获得的端口段信息，包含 poolName、nodeName、startPort、endPort 等。通过 annotation 存储在 Pod 上。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Pod 调度完成后，端口段分配在单次 Reconcile 周期内完成（通常秒级），initContainer 通过 HTTP API 可立即获取分配结果。
- **SC-002**: Pod 删除或驱逐后，端口段在下一次 Reconcile 周期内被释放，可被新 Pod 复用。
- **SC-003**: 多个 Pod 并发调度到同一 Node 时，端口段分配无冲突，每个 Pod 获得独立的端口范围。
- **SC-004**: Controller 重启后，缓存重建完成且不产生端口冲突——已有 Pod 占用的端口段不会被重复分配。
- **SC-005**: HostNetPortPool 端口范围扩大后，新段立即可用于分配；缩小时有冲突的操作被拒绝并给出明确诊断信息。
- **SC-006**: 集群扩缩容场景下，新 Node 上的端口分配自动可用，移除 Node 后缓存被正确清理。
- **SC-007**: 新功能与现有 PortPool 机制完全独立，两者可在同一集群中并行使用互不干扰。
- **SC-008**: 端口泄漏定期巡检机制确保孤立的端口段最终被释放，不会出现永久性端口泄漏。

## Assumptions

- Pod 使用 `hostNetwork: true` 模式，不经过云负载均衡器（CLB/ALB/NLB 等），端口直接在 Node 上使用。
- Pod 的 `dnsPolicy` 需设置为 `ClusterFirstWithHostNet`，以便 initContainer 能解析集群内 Service 域名。
- IngressController 已通过 Service 对集群内暴露 RESTful API（端口 18088），新增的 HTTP API 在此 Service 上扩展。
- 端口段是 per-Node 隔离的，同一端口段可以同时分配给不同 Node 上的不同 Pod。
- 端口池数量和 Node 数量在合理范围内（通常几十到几百），缓存遍历开销可接受。
- 业务侧需在 Pod Spec 中配置 initContainer 来等待端口分配结果，Controller 不负责注入 initContainer。
