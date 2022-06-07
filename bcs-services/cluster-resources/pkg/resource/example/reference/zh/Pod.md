# Pod

> Pod（Po）是可以在 Kubernetes 中创建和管理的、最小的可部署的计算单元。

## 什么是 Pod ？

Pod 是一组（一个或多个）容器，这些容器共享存储、网络、以及怎样运行这些容器的声明。 Pod 中的内容总是并置（colocated）的并且一同调度，在共享的上下文中运行。

Pod 的共享上下文包括一组 Linux 名字空间、控制组（cgroup）和可能一些其他的隔离 方面，即用来隔离 Docker 容器的技术。 在 Pod 的上下文中，每个独立的应用可能会进一步实施隔离。

Pod 可以认为是特定于应用的 “逻辑主机”，其中包含一个或多个应用容器，这些容器是相对紧密的耦合在一起的。除了应用容器，Pod 还可以包含在 Pod 启动期间运行的 Init 容器。

## 使用 Pod

一般来说，我们很少在 Kubernetes 集群中直接创建一个或多个 Pod，这是因为 Pod 被设计为临时性的，用后即抛的一种实体。Pod 由用户或者控制器创建后，将会被自动调度到合适的集群节点上运行，在 Pod 执行完成，被删除，因资源不足被驱逐或节点失效之前，Pod 都会在该节点上保持运行状态。

常用的 Pod 控制器有以下几类：

- [Deployment](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/): 管理应用副本的 API 对象，定义了 Pod 模版信息，副本数量，运行配置等，适合管理集群中的无状态应用（如 API 等）。

- [StatefulSet](https://kubernetes.io/zh/docs/concepts/workloads/controllers/statefulset/): 管理有状态应用的 API 对象，用来管理 Pod 集合的部署和扩缩，并为这些 Pod 提供持久存储和持久标识符。

- [DaemonSet](https://kubernetes.io/zh/docs/concepts/workloads/controllers/daemonset/): 确保全部或某些节点上运行一个 Pod 的副本，一般用于运行 `集群/日志收集/监控` 的守护进程。

- [Job](https://kubernetes.io/zh/docs/concepts/workloads/controllers/job/): 创建一个或多个 Pod，用于执行一次性任务；如需要 Job 重复执行，可以使用 [CronJob](https://kubernetes.io/zh/docs/concepts/workloads/controllers/cron-jobs/) 。

当然，有部分的场景，更加适合创建独立的 Pod，比如：

- 启动单个 Pod 进行调试，临时验证服务的场景

- 针对集群节点状态，单个 Pod 配置进行调试的场景

## Pod 更新与替换

当某 Pod 控制器的 Pod 模板被改变时，控制器会按最新模板创建新的 Pod ，并回收旧的 Pod 对象，不会现有 Pod 执行更新或者修补操作。

Kubernetes 并不禁止直接管理 Pod，允许对运行中的 Pod 的某些字段执行就地更新操作。不过，类似 patch 和 replace 这类操作有以下的限制：

- Pod 的绝大多数元数据都是不可变的。例如不可改变其 `namespace`、`name`、`uid` 或 `creationTimestamp` 字段；`generation` 字段是比较特别的，如果更新该字段，只能增加字段取值而不能减少。

- 如果 `metadata.deletionTimestamp` 已经被设置，则不可以向 `metadata.finalizers` 列表中添加新的条目。

- Pod 更新不可以改变除 `spec.containers[*].image`、`spec.initContainers[*].image`、`spec.activeDeadlineSeconds` 或 `spec.tolerations` 之外的字段。 对于 `spec.tolerations`，你只被允许添加新的条目到其中。

- 在更新 `spec.activeDeadlineSeconds` 字段时，以下两种更新操作是被允许的：
  - 如果该字段尚未设置，可以将其设置为一个正数；
  - 如果该字段已经设置为一个正数，可以将其设置为一个更小的、非负的整数。

## 参考资料

1. [Kubernetes / 工作负载 / Pods](https://kubernetes.io/zh/docs/concepts/workloads/pods/)
2. [Kubernetes Pod 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#pod-v1-core)
