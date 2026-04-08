## Context

DRPlan 通过 Stage/Workflow/Action 编排跨集群下发与运维动作。当前 Subscription Action 仅保证 Subscription CR 创建成功，不保证：

- Subscription 已被调度到目标集群（`status.bindingClusters`）
- 子集群实际资源已创建并达到 Ready

在 Helm 场景中，hook 与 `--wait` 的核心语义依赖“等待完成/就绪”，因此需要引入 `waitReady` 机制。

本变更聚焦 **Subscription** 场景，并使用你已确认可用的 Clusternet SocketProxy 进行子集群直查。
在此基础上补充 install/upgrade 运行时分流能力（`mode + when`），避免为 install/upgrade 维护两套 Plan。

## Goals / Non-Goals

**Goals:**

- 为 Subscription Action 提供 `waitReady` 能力（默认关闭，显式开启才等待）
- 等待策略对标 Helm 的核心诉求：依赖链下游只在上游资源 Ready 后执行
- 子集群状态通过 SocketProxy 直查（降低 AggregatedStatuses 延迟与不确定性）
- `drplan-gen` 基于 Helm hook 语义自动为 hook 对应 Subscription Action 启用 `waitReady: true`
- 通过 `DRPlanExecution.spec.mode` + `Action.when` 支持单 Plan 下 install/upgrade 路径分流
- 预留未来“通过 kubeconfig 访问子集群”的扩展点（可配置，但本次不实现）

**Non-Goals:**

- 不实现 Job / KubernetesResource / Localization 的 `waitReady`（后续变更再覆盖）
- 不新增新的 CRD 字段用于轮询间隔等细粒度参数（先使用内部常量）
- 不引入对 clusternet agent 回报状态的强依赖（仅作为未来可能的 fallback）
- 不实现复杂表达式引擎（仅支持 `mode == "install|upgrade"` 的单条件最小语义）

## Decisions

- **API**: 在 `Action` 中新增 `waitReady: bool`，默认 `false`（兼容旧行为）
- **API**: 在 `DRPlanExecution` 中新增可选 `mode: Install|Upgrade`
- **等待流程**（Subscription）:
  - Phase A: 轮询 Subscription `status.bindingClusters` 非空
  - Phase B: 对每个 binding cluster，通过 SocketProxy 访问子集群 API，逐 feed 校验 readiness
- **集群标识**: 从 `ManagedCluster.spec.clusterId` 获取 `clusterID`
- **token 获取**: 从 binding cluster 的 namespace 中读取 `child-cluster-deployer` Secret（每集群独立 token）
- **可配置访问方式**: 抽象 `ChildClusterClientFactory` 接口；默认 SocketProxy，实现上预留 kubeconfig 分支（仅结构预留）
- **drplan-gen 自动策略**:
  - 仅 hook 对应 Subscription Action 设置 `waitReady: true`，install 主流程默认不设置
  - 默认生成一个 stage + 一个 workflow，hook 与主资源 action 按固定顺序统一落在 `workflow-install.yaml`
- **when 执行策略**:
  - 支持 Action 级 `when`，当前仅支持 `mode == "install"` / `mode == "upgrade"`
  - 不支持多条件表达式（如 `||` / `&&`）
  - `mode` 来源于 `execution.spec.mode` 注入
  - 若未设置 mode，则保持兼容：不按 when 过滤（等价全部执行）
- **多 hook 值策略**: 对 `helm.sh/hook: a,b` 进行拆分，资源可同时进入多个 hook 类型

## Risks / Trade-offs

- **SocketProxy 不可用/网络抖动**: 启用 `waitReady` 的动作可能超时失败；需提供清晰错误信息便于定位（token/网络/权限）
- **轮询成本**: 多集群 * 多 feeds 的轮询会增加请求量；通过固定间隔 + `timeout` 上界控制
- **生成器默认启用 waitReady 的行为变更**: 生成输出更“严格”，执行耗时可能增加，但更符合 Helm 等待语义；必要时可后续追加 CLI flag 做开关（不在本次范围）
- **条件执行语义限制**: 仅支持 mode 单条件等值判断，复杂条件需后续扩展表达式引擎

