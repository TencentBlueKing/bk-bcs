# StatefulSet

> StatefulSet（STS）是用来管理有状态应用的工作负载 API 对象，其管理某 Pod 集合的部署和扩缩，并为这些 Pod 提供持久存储和持久标识符。

## 什么是 StatefulSet ？

StatefulSet 和 Deployment 类似，StatefulSet 管理基于相同容器规约的一组 Pod。但和 Deployment 不同的是，StatefulSet 为它们的每个 Pod 维护了一个有粘性的 ID。这些 Pod 是基于相同的规约来创建的，但是不能相互替换，即无论怎么调度，每个 Pod 都有一个永久不变的 ID。

如果希望使用存储卷为工作负载提供持久存储，可以使用 StatefulSet 作为解决方案的一部分。尽管 StatefulSet 中的单个 Pod 仍可能出现故障，但持久的 Pod 标识符使得将现有卷与替换已失败 Pod 的新 Pod 相匹配变得更加容易。

## 使用 StatefulSet

常见 StatefulSet 使用场景有：

- 稳定的，唯一的网络标识符：可用于发现集群内的其他成员，一般来说，若 STS 的名字为 `nginx`，副本数为 3，则其 Pod 名称将是 `nginx-0, nginx-1, nginx-2`，在 Pod 被重新调度后，其名称与 HostName 不会发生改变。
- 稳定的，持久的存储：通过 [PV & PVC](https://kubernetes.io/zh/docs/concepts/storage/persistent-volumes/) 来实现，在 Pod 重新调度之后，还可以访问到相同的持久化数据；出于安全考虑，删除 Pod 时候并不会默认清理持久化数据。
- 有序的，优雅的部署和缩放：StatefulSet 的 Pod 是有顺序的，在部署或扩展的时候要依据定义的顺序依次进行（从 0 到 N-1，每个 Pod 运行的条件是前一个（若有）的 Pod 状态为 `Running/Ready`）；反之亦然，从 N-1 到 0 缩减 Pod 的数量。
- 有序的、自动的滚动更新。

## 限制

- 给定 Pod 的存储必须由 [PersistentVolume 驱动](https://github.com/kubernetes/examples/blob/master/staging/persistent-volume-provisioning/README.md) 基于所请求的 `StorageClass` 来提供，或者由管理员预先提供。
- 出于保证数据安全的考虑，删除或者收缩 StatefulSet 并不会删除它关联的存储卷。
- StatefulSet 当前需要 [无头服务](https://kubernetes.io/zh/docs/concepts/services-networking/service/#headless-services) 来负责 Pod 的网络标识。用户需要负责创建此服务。
- 当删除 StatefulSet 时，StatefulSet 不提供任何终止 Pod 的保证。为了实现 StatefulSet 中的 Pod 可以有序地且体面地终止，可以在删除之前将 StatefulSet 的 Pod 副本数量调整为 0。
- 在默认 [Pod 管理策略](https://kubernetes.io/zh/docs/concepts/workloads/controllers/statefulset/#pod-management-policies) （`OrderedReady`）时使用 [滚动更新](https://kubernetes.io/zh/docs/concepts/workloads/controllers/statefulset/#rolling-updates) ，可能进入需要 [人工干预（强制回滚）](https://kubernetes.io/zh/docs/concepts/workloads/controllers/statefulset/#forced-rollback) 才能修复的损坏状态。

## 参考资料

1. [Kubernetes / 工作负载 / StatefulSets](https://kubernetes.io/zh/docs/concepts/workloads/controllers/statefulset/)
2. [kubernetes StatefulSet 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#statefulset-v1-apps)
