# Deployment

> Deployment（Deploy）为 Pods 和 ReplicaSets 提供声明式的更新能力。

## 什么是 Deployment ？

Deployment 是 Kubernetes v1.2 中引入的新概念，目的是为了更好地解决 Pod 的编排问题；Deployment 内部使用 ReplicaSet 来实现，就具体操作而言，可以将 Deployment 看作 Replication Controller（RC）的升级版。

Deployment 可以随时跟踪并知晓其纳管的 Pod 的部署进度，这是相较于 RC 最大的升级，增强了对部署过程的掌握能力。

## 使用 Deployment

常见 Deployment 使用场景有：

- [创建 Deployment 以部署 Pod](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#creating-a-deployment) ，生成的 ReplicaSet 在后台创建 Pods，通过检查 ReplicaSet 的上线状态可以验证 Pod 部署是否成功。
- 通过 [更新 Deployment 的 PodTemplateSpec 以声明 Pod 的新状态](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#updating-a-deployment) 。新的 ReplicaSet 会被创建，Pod 将以受控速率从旧 ReplicaSet 迁移到新 ReplicaSet。
- 如果 Deployment 的当前状态不稳定，可以选择 [回滚到较早的 Deployment 版本](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#rolling-back-a-deployment) 。每次回滚都会更新 Deployment 的修订版本。
- [暂停 Deployment](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#pausing-and-resuming-a-deployment) 以应用对 PodTemplateSpec 所作的多项修改，然后恢复其执行以启动新的上线版本。
- 通过 [查看 Deployment 状态](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#deployment-status) 来判定上线过程是否出现停滞或故障。
- [扩大 Deployment 规模以承担更多负载](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#scaling-a-deployment) 或 [清理较旧的不再需要的 ReplicaSet](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/#clean-up-policy) 。

## 参考资料

1. [Kubernetes / 工作负载 / Deployments](https://kubernetes.io/zh/docs/concepts/workloads/controllers/deployment/)
2. [Kubernetes Deployment 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#deployment-v1-apps)
