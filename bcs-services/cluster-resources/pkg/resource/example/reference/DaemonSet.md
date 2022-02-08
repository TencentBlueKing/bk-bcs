# DaemonSet

> DaemonSet（DS）即 `守护进程集`，是一系列守护进程的集合，负责纳管集群的守护进程。DaemonSet 确保全部（或者某些）节点上运行一个 Pod 的副本。当有节点加入集群时，也会为他们新增一个 Pod。当有节点从集群移除时，这些 Pod 也会被回收。删除 DaemonSet 将会删除它创建的所有 Pod。

## 使用 DaemonSet

常见 DaemonSet 使用场景有：

- 运行 `集群存储/网络 Daemon`，如每个节点上运行的 ceph，glusterd，nginx 等
- 运行 `日志采集 Daemon`，如每个节点上运行的 logstash，fluentd 等
- 运行 `服务/性能/指标监控 Daemon`，如 PrometheusNodeExporter，collectd，Datadog 等

一种简单的用法是为每种类型的守护进程在所有的节点上都启动一个 DaemonSet。一个稍微复杂的用法是为同一种守护进程部署多个 DaemonSet；每个具有不同的标志，并且对不同硬件类型具有不同的内存、CPU 要求。

### 仅在部分 Node 上运行 DaemonSet 的 Pod ？

DaemonSet 默认在所有集群节点上各新建一个 Pod，但是也有方法可以特殊指定节点：

- nodeSelector：通过配置 DaemonSet 的 `spec.template.spec.nodeSelector`，可以让 DaemonSet 仅在能够匹配上 NodeSelector 的节点上创建 Pod。
- affinity：相似的，可以用过配置 `spec.template.spec.affinity`，利用节点亲和性来控制 Pod 仅部署在部分节点上。

### DaemonSet Pod 调度

正常情况下，Pod 运行在哪个机器上是由 Kubernetes 调度器自行选择的。然而，DaemonSet Pod 由 DaemonController 创建并调度，因此与普通的 Pod 有所不同：

- Pod 调度的确定性：DaemonSet Pod 创建时即制定了 `spec.nodeName`，因此其调度结果是可确定的。
- Pod 行为的不一致性：正常 Pod 在被创建后等待调度时处于 Pending 状态，DaemonSet Pod 创建后不会处于 Pending 状态下。
- 自动添加 `node.kubernetes.io/unschedulable：NoSchedule` 容忍度到 DaemonSet Pod。在调度 DaemonSet Pod 时，不会关心节点的 `unschedulable` 状态。
- 即使 Kubernetes 调度器还没有启动，DaemonSet Pod 也可以创建，这对集群启动是非常有帮助的。

### 污点与容忍度

尽管 Daemon Pods 遵循 [污点和容忍度](https://kubernetes.io/zh/docs/concepts/scheduling-eviction/taint-and-toleration/) 规则，根据相关特性，控制器会自动将以下容忍度添加到 DaemonSet Pod：

| 容忍度键名                               | 效果       | 版本  | 描述                                                                              |
| ---------------------------------------- | ---------- | ----- | --------------------------------------------------------------------------------- |
| `node.kubernetes.io/not-ready`           | NoExecute  | 1.13+ | 当出现类似网络断开的情况导致节点问题时，DaemonSet Pod 不会被逐出。                |
| `node.kubernetes.io/unreachable`         | NoExecute  | 1.13+ | 当出现类似于网络断开的情况导致节点问题时，DaemonSet Pod 不会被逐出。              |
| `node.kubernetes.io/disk-pressure`       | NoSchedule | 1.8+  | DaemonSet Pod 被默认调度器调度时能够容忍磁盘压力属性。                            |
| `node.kubernetes.io/memory-pressure`     | NoSchedule | 1.8+  | DaemonSet Pod 被默认调度器调度时能够容忍内存压力属性。                            |
| `node.kubernetes.io/unschedulable`       | NoSchedule | 1.12+ | DaemonSet Pod 能够容忍默认调度器所设置的 unschedulable 属性.                      |
| `node.kubernetes.io/network-unavailable` | NoSchedule | 1.12+ | DaemonSet 在使用宿主网络时，能够容忍默认调度器所设置的 network-unavailable 属性。 |

## 参考资料

1. [Kubernetes / 工作负载 / DaemonSet](https://kubernetes.io/zh/docs/concepts/workloads/controllers/daemonset/)
2. [Kubernetes DaemonSet 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#daemonset-v1-apps)
