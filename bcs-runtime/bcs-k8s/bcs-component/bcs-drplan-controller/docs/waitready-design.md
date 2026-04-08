# Subscription waitReady 设计对齐说明

本文档只描述当前已经落地的 `waitReady` 设计与实现，不再保留早期“所有 Action 类型统一等待”的方案草稿。

当前结论只有一条：

- `waitReady` 字段已经存在于 `Action`
- 当前只有 `Subscription` Action 真正实现了 `waitReady`
- `drplan-gen` 当前只会为 hook 相关的 `Subscription` action 自动生成 `waitReady: true`
- 主资源安装 action 默认不自动带 `waitReady`

## 背景

DRPlan 的 action 默认是“请求提交成功即认为当前 action 成功”。这对普通资源创建是可接受的，但对 Helm hook 语义并不够。

以 hook 为例，用户真正需要的是：

- pre-install 资源执行完，再继续主资源下发
- post-install / post-upgrade 在前面的资源真正 ready 后再继续

如果只有顺序、没有等待，那么即使 action 串行执行，也仍然可能出现：

- Subscription CR 已创建，但子集群资源还没真正落地
- Deployment 已分发，但 Pod 还没 ready
- post hook 太早执行，行为和 Helm 不一致

因此当前实现把 `waitReady` 的第一阶段能力落在了 `Subscription` 上。

## 当前范围

`api/v1alpha1/common_types.go` 中 `Action.WaitReady` 的注释已经明确：

- 默认 `false`
- 当前仅对 `Subscription` 生效
- 其他 action 类型暂不在本次范围内

也就是说，本文档不再讨论以下未实现范围：

- `Job` action 的 waitReady
- `KubernetesResource` action 的 waitReady
- `Localization` action 的 waitReady
- `HTTP` action 的 waitReady

这些能力未来可以继续扩展，但不是当前实现的一部分。

## 与 Helm hook 的对标方式

当前对标 Helm hook 的方式不是“多 stage + dependsOn + 多 workflow”这一套旧模型，而是：

1. 一个 `DRPlan` 只生成一个主 stage
2. 一个主 stage 只引用一个统一 `workflow-install.yaml`
3. workflow 内 action 通过 `dependsOn` 形成 hook DAG
4. hook action 使用 `when`
5. hook action 使用 `waitReady: true`

对应关系如下：

- `pre-install` -> `when: mode == "install"`
- `post-install` -> `when: mode == "install"`
- `pre-upgrade` -> `when: mode == "upgrade"`
- `post-upgrade` -> `when: mode == "upgrade"`

运行时通过 `DRPlanExecution.spec.mode` 区分安装还是升级路径：

- `Install`
- `Upgrade`

如果 execution 没有传 `mode`，执行器保持兼容行为，不按 `when` 过滤。

## Subscription waitReady 的真实判定逻辑

`SubscriptionActionExecutor` 的等待逻辑分两段，且当前对 parent / per-cluster child 做了区分。

### 第 1 段：等待父 Subscription 调度完成

父集群侧轮询 `Subscription.status.bindingClusters`。

当满足以下条件时，进入下一段：

- `status.bindingClusters` 存在
- 且列表非空

当前 `bindingClusters` 按如下格式解析：

```yaml
status:
  bindingClusters:
    - cls-a-ns/cls-a
    - cls-b-ns/cls-b
```

也就是：

- 前半段是 `ManagedCluster` 所在 namespace
- 后半段是 `ManagedCluster` 名称

在这一步里，如果任何已绑定集群对应的 `Description.status.phase == Failure`，等待会直接失败，不再继续等超时。

### 第 2 段：针对目标子集群检查 feed 就绪

执行器会针对每个 binding cluster：

1. 读取对应 `ManagedCluster`
2. 从 `ManagedCluster.spec.clusterId` 取出 `clusterID`
3. 以 binding cluster 的 namespace 读取 `child-cluster-deployer` secret
4. 用父集群 apiserver + SocketProxy 构造 child client
5. 对当前 action 中所有 feeds 逐个查询子集群真实资源状态

当前就绪规则如下：

| 资源类型 | ready 判定 |
| --- | --- |
| `Deployment` | `availableReplicas >= spec.replicas` 且 `updatedReplicas >= spec.replicas` |
| `StatefulSet` | `readyReplicas >= spec.replicas` |
| `DaemonSet` | `numberReady >= desiredNumberScheduled` |
| `Job` | `status.conditions` 中出现 `Complete=True` |
| 其他资源 | 资源存在即认为 ready |

说明：

- `Job` 如果出现 `Failed=True`，等待直接失败
- `ConfigMap`、`Secret`、`Service` 这类对象当前按“存在即可”处理
- `feed` 中的模板变量会先渲染，再用于 waitReady 检查，避免用未渲染 namespace/name 去查子集群

## 子集群访问方案

默认实现是 SocketProxy，并且已经预留成 `ChildClusterClientFactory` 接口，后续可扩展 kubeconfig 模式。

当前默认工厂为：

- `NewSocketProxyChildClusterClientFactory`

核心行为：

1. 基于父集群 `rest.Config` 复制出一份 child config
2. 将 `Host` 改写为：

```text
{parentAPIServer}/apis/proxies.clusternet.io/v1alpha1/sockets/{clusterID}/proxy/direct
```

3. 通过请求头注入：

- `Impersonate-User: clusternet`
- `Impersonate-Extra-Clusternet-Token: <token>`

4. token 来源：

- secret 名称固定为 `child-cluster-deployer`
- secret namespace 为当前 binding cluster 的 namespace
- 优先读取 `token`
- 兼容旧 key `child-cluster-token`

## 超时与轮询

当前实现复用 action 的 `timeout` 字段：

- 未设置时默认 `5m`
- 非法或非正数会直接报错

轮询间隔当前为固定常量：

- `5s`

## drplan-gen 的生成约定

当前 `drplan-gen` 的输出约定已经调整为单 workflow 模型：

- 只生成一个 `workflow-install.yaml`
- 主资源 action 为一个普通 `Subscription`
- hook 资源也生成 `Subscription` action，而不是直接生成 `Job` action
- 只有 hook 相关 action 自动带 `waitReady: true`
- 主资源安装 action 默认不自动带 `waitReady`
- hook action 自动写入 `when`

一个典型片段如下：

```yaml
spec:
  actions:
    - name: pre-install-job
      type: Subscription
      when: mode == "install"
      waitReady: true
      clusterExecutionMode: PerCluster
      subscription:
        operation: Create
        name: pre-install-job-sub
        namespace: "$(params.feedNamespace)"
    - name: create-subscription
      type: Subscription
      subscription:
        operation: Apply
        name: demo-subscription
        namespace: "$(params.feedNamespace)"
    - name: post-upgrade-job
      type: Subscription
      when: mode == "upgrade"
      waitReady: true
      clusterExecutionMode: PerCluster
      subscription:
        operation: Create
        name: post-upgrade-job-sub
        namespace: "$(params.feedNamespace)"
```

这种模型依赖两个事实：

- workflow 内通过 `dependsOn` 描述 hook 前后关系
- hook action 的“完成”被 `waitReady` 从“创建成功”提升为“子集群资源 ready”

因此对当前 hook 场景来说，顺序由生成器显式写入 `dependsOn`，而不是依赖“天然串行”。

## 已实现文件

本次能力实际落地集中在以下文件：

- `api/v1alpha1/common_types.go`
- `api/v1alpha1/drplanexecution_types.go`
- `internal/executor/cluster_client.go`
- `internal/executor/readiness.go`
- `internal/executor/subscription_executor.go`
- `internal/executor/subscription_waitready_test.go`
- `internal/generator/planner.go`
- `docs/helm-hook-pattern.md`
- `docs/user-guide.md`
- `docs/drplan-gen-guide.md`

## 非目标

以下内容仍然不是当前版本能力：

- 通用的 `--wait` 等价能力
- 所有 action 类型统一支持 `waitReady`
- 多条件 `when`
- `drplan-gen` 自动为所有 Subscription action 都加 `waitReady`
- 依赖 AggregatedStatuses 作为 ready 判定主来源

## 后续可扩展方向

后续如果继续扩展，可以沿着这几个方向推进：

1. 为 `Job` action 增加原生 waitReady
2. 为 `KubernetesResource` action 增加按 kind 的 ready 判定
3. 新增可配置的 child cluster 访问后端，例如 kubeconfig 模式
4. 在生成器中增加更细粒度的 wait 策略开关
