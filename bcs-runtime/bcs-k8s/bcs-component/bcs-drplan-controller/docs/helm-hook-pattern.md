# DRPlan 实现 Helm Hook 模式

本文档说明当前项目如何用 **单 stage + 单 workflow + `dependsOn` DAG + `when` + `waitReady`** 对标 Helm hook。

## 背景

Helm hook 的核心是三件事：

- 前置 hook 要先执行
- 后置 hook 要在主资源就绪后再执行
- install / upgrade 需要走不同路径

当前项目的默认落地方式已经收敛为：

- 一个 `DRPlan`
- 一个 stage：`install`
- 一个 workflow：`workflow-install.yaml`
- 所有 hook action 与主资源 action 都放在同一个 workflow 中
- 通过 `dependsOn` 表达前后依赖和并行关系

## 完整对标矩阵

| Helm 能力 | DRPlan 实现 | 说明 |
|---|---|---|
| `pre-install` | action 位于主资源前 + `when: mode == "install"` | 仅 Install 时执行 |
| `pre-upgrade` | action 位于主资源前 + `when: mode == "upgrade"` | 仅 Upgrade 时执行 |
| `post-install` | action 位于主资源后 + `when: mode == "install"` | 仅 Install 时执行 |
| `post-upgrade` | action 位于主资源后 + `when: mode == "upgrade"` | 仅 Upgrade 时执行 |
| `pre-install, pre-upgrade` 组合 | 生成两个 action，各自保留 `when` 和 weight | 分别对齐 Install / Upgrade 排序 |
| `post-install, post-upgrade` 组合 | 生成两个 action，各自保留 `when` 和 weight | 分别对齐 Install / Upgrade 排序 |
| 主资源部署 | `operation: Apply`（SSA 幂等） | Install 创建，Upgrade 更新 |
| `hook-weight` 排序 | 按 weight 分组，通过 `dependsOn` 链式依赖 | 同 weight 并行（增强） |
| 同 weight 执行 | 同组 action 共享 `dependsOn`，无互相依赖 | Helm 串行，DRPlan 并行 |
| `hook-delete-policy` | 生成 `hookCleanup.beforeCreate/onSuccess/onFailure` | 分别对齐 Helm 清理时机 |
| hook 默认等待完成 | hook action 上的 `waitReady: true` | Helm 内置行为：上一个 hook 完成后才继续 |
| Helm `--wait` | 主资源 action 可选加 `waitReady: true` | 等待主资源 Ready（当前默认不开启） |
| `test` / `test-success` | 跳过不生成 | 测试 hook 不在部署流程中 |
| `pre-delete` / `post-delete` | 生成 action，`when: mode == "delete"` | 通过 Delete execution 路径触发 |
| `pre-rollback` / `post-rollback` | 生成 action，`when: mode == "rollback"` | 通过 Revert execution 路径触发 |

## 编排方式：dependsOn DAG

`drplan-gen` 使用 `dependsOn` 构建 DAG 依赖图，实现三层编排：

1. **pre-hooks**（按 weight 分组链式依赖）
2. **主资源 action**（`dependsOn` 最后一组 pre-hook）
3. **post-hooks**（`dependsOn` 主资源，按 weight 分组链式依赖）

### DAG 示例

假设 chart 有以下 hook：

- pre-install Job A（weight=-5）
- pre-install Job B（weight=0）
- pre-install Job C（weight=0）
- post-install Job D（weight=-3）
- post-install Job E（weight=-3）
- post-install Job F（weight=10）

生成的 DAG：

```text
A (w=-5) → [B ∥ C] (w=0) → create-subscription → [D ∥ E] (w=-3) → F (w=10)
```

- 同 weight 的 B、C 并行执行（共享 dependsOn: [A]）
- 同 weight 的 D、E 并行执行（共享 dependsOn: [create-subscription]）
- F 等待 D、E 都完成后执行（dependsOn: [D, E]）

### Operation 与 Cleanup

| 资源类型 | operation | 原因 |
|---|---|---|
| 主资源 Subscription | `Apply` | SSA 幂等，Install 创建 / Upgrade 更新 |
| 所有 hook Subscription | `Create` | 创建语义独立，清理由 `hookCleanup` 控制 |

hook cleanup 规则：

- 默认 `beforeCreate: true`
- `hook-succeeded` -> `onSuccess: true`
- `hook-failed` -> `onFailure: true`

## 关键约束

### 1. `when` 支持 mode 等值判断与 `||`

当前支持：

```yaml
when: mode == "install"
when: mode == "upgrade"
when: mode == "install" || mode == "upgrade"
```

install / upgrade 的组合 hook 仍建议拆成独立 action，以表达 pre/post 的前后位置与依赖。

### 2. `waitReady` 当前只对 `Subscription` action 生效

它会：

1. 先等待 `Subscription.status.bindingClusters` 非空
2. 再按绑定集群逐个检查 feeds 是否 Ready

### 3. mode 为空时的兼容行为

当 `DRPlanExecution.spec.mode` 未设置时，所有 action（包括带 `when` 的）都会执行。
这是为了向后兼容旧版本不支持 mode 的场景。

## 示例

### DRWorkflow（带 dependsOn）

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRWorkflow
metadata:
  name: demo-app-install
  namespace: default
spec:
  failurePolicy: FailFast
  parameters:
    - name: feedNamespace
      type: string
      default: default
  actions:
    - name: db-migrate
      type: Subscription
      timeout: 5m
      waitReady: true
      when: mode == "install"
      clusterExecutionMode: PerCluster
      hookCleanup:
        beforeCreate: true
      subscription:
        operation: Create
        name: db-migrate-sub
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: batch/v1
              kind: Job
              name: db-migrate
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}

    - name: create-subscription
      type: Subscription
      timeout: 5m
      dependsOn:
        - db-migrate
      subscription:
        operation: Apply
        name: demo-app-subscription
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: apps/v1
              kind: Deployment
              name: demo-app-server
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}

    - name: health-check
      type: Subscription
      timeout: 5m
      waitReady: true
      when: mode == "install"
      clusterExecutionMode: PerCluster
      dependsOn:
        - create-subscription
      hookCleanup:
        beforeCreate: true
        onSuccess: true
      subscription:
        operation: Create
        name: health-check-sub
        namespace: $(params.feedNamespace)
        spec:
          schedulingStrategy: Replication
          feeds:
            - apiVersion: batch/v1
              kind: Job
              name: health-check
              namespace: $(params.feedNamespace)
          subscribers:
            - clusterAffinity: {}
```

### DRPlanExecution

Install（首次安装）：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: demo-app-install-001
spec:
  mode: Install
  operationType: Execute
  planRef: demo-app
```

Upgrade（后续升级）：

```yaml
apiVersion: dr.bkbcs.tencent.com/v1alpha1
kind: DRPlanExecution
metadata:
  name: demo-app-upgrade-001
spec:
  mode: Upgrade
  operationType: Execute
  planRef: demo-app
```

### 执行行为对比

| action | mode=Install | mode=Upgrade |
|---|---|---|
| db-migrate (`when: install`) | 执行 | 跳过 |
| create-subscription（无 when） | 执行（Apply 创建） | 执行（Apply 更新） |
| health-check (`when: install`) | 执行 | 跳过 |

## 注意事项

1. 当前文档描述的是 `drplan-gen` 的默认生成模型。引擎本身支持更灵活的编排方式。
2. 参数模板推荐使用 `$(params.xxx)`，历史 `{{ .params.xxx }}` 继续兼容。
3. `waitReady` 当前只保证 `Subscription` 路径有效。
4. 回滚按执行记录逆序处理，与 hook 编排无冲突。
