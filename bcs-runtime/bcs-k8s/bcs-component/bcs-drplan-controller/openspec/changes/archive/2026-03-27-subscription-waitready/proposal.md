## Why

当前 DRPlan 的 `Subscription` Action 在创建 Subscription CR 成功后即返回 `Succeeded`（fire-and-forget），无法保证子集群资源已下发并就绪。
这会导致:

- `dependsOn` 只能保证顺序，无法保证前置资源真正 Ready 后再执行后置步骤
- Helm hook / `--wait` 语义无法对标（例如：post-install 检查可能在应用未 Ready 时就执行）

## What Changes

- 在 API 中为 `Action` 增加 `waitReady` 字段（默认 `false`，保持兼容）
- 在 API 中为 `DRPlanExecution` 增加可选 `mode` 字段（`Install|Upgrade`），用于运行时路径区分
- 为 `Subscription` Executor 实现 `waitReady`：
  - Phase A: 等待 Subscription `status.bindingClusters` 非空
  - Phase B: 通过 Clusternet SocketProxy 直查子集群资源就绪状态（按 feed.Kind 判定）
- 为 Workflow 执行器实现 `Action.when` 最小语义：
  - 支持 `mode == "install"` / `mode == "upgrade"`
  - 不支持多条件表达式
  - 未提供 execution mode 时保持兼容（不按 when 过滤，等价“都执行”）
- 新增 `ChildClusterClientFactory`（默认 SocketProxy 实现），并预留后续切换 kubeconfig 的扩展点（本次不实现）
- `drplan-gen` 基于 Helm hook 语义自动设置：
  - 仅 hook 对应 `Subscription` Action 默认设置 `waitReady: true`（install 主流程不默认设置）
  - hook action 自动写入 `when`（install hooks => `mode == "install"`，upgrade hooks => `mode == "upgrade"`）
  - 支持 `helm.sh/hook` 多值（如 `pre-install,pre-upgrade`）拆分为多个 hook 条目
  - 默认生成一个 stage + 一个 workflow，hook 与主资源 action 统一落在 `workflow-install.yaml`

## Impact

- **API/CRD**: `Action` 增加可选字段 `waitReady`
- **API/CRD**: `DRPlanExecution` 增加可选字段 `mode`
- **RBAC**: controller 需要读取 `Subscription` / `ManagedCluster` / `Secret`
- **生成器**: `drplan-gen` 输出 YAML 变化（默认单 workflow，hook action 带 `when` 与 `waitReady`，execute sample 带 `mode`，golden files 与测试已同步）
- **运行时**:
  - 启用 `waitReady` 的 Subscription Action 将在超时窗口内轮询子集群资源状态
  - `when + mode` 使单 Plan 下 install/upgrade 路径可按执行实例区分

