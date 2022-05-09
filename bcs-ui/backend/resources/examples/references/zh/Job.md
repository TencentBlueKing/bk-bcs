# Job

> Job 会创建一或多个 Pods，执行一次性任务到目标成功数后结束。

## 什么是 Job ？

Job 会创建一个或者多个 Pods，并将继续重试 Pods 的执行，直到指定数量的 Pods 成功终止。

随着 Pods 成功结束，Job 跟踪记录成功完成的 Pods 个数。当数量达到指定的成功个数阈值时，任务（即 Job）结束。

删除 Job 的操作会清除所创建的全部 Pods。挂起 Job 的操作会删除 Job 的所有活跃 Pod，直到 Job 被再次恢复执行。

Job 支持并行，可以通过指定并行度来同时运行多个 Pod。

## 使用 Job

常见 Job 使用场景有：

- 执行一次性任务，如 DB Migrate，数据采样，单次运算任务等
- 执行并行任务，如 分布式并行计算 等

## 参考资料

1. [Kubernetes / 工作负载 / Jobs](https://kubernetes.io/zh/docs/concepts/workloads/controllers/job/)
2. [kubernetes Job 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#job-v1-batch)
