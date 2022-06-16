# HorizontalPodAutoscaler

> HorizontalPodAutoscaler（HPA）能够自动更新工作负载资源（例如 Deployment 或 StatefulSet），目的是自动扩缩工作负载以满足需求。

## HPA 工作原理

![img](https://d33wubrfki0l68.cloudfront.net/4fe1ef7265a93f5f564bd3fbb0269ebd10b73b4e/1775d/images/docs/horizontal-pod-autoscaler.svg)

HorizontalPodAutoscaler 控制 Deployment 及其 ReplicaSet 的规模，从而达到根据资源使用情况，控制 Pod 数量的效果；Pod 水平自动扩缩实现为一个间歇运行的控制回路，默认间隔为 15 秒。

在每个时间段 内，控制器管理器都会根据每个 HorizontalPodAutoscaler 定义中指定的指标查询资源利用率。 控制器管理器找到由 scaleTargetRef 定义的目标资源，然后根据目标资源的 `.spec.selector` 标签选择 Pod，并从资源指标 API 或自定义指标获取指标 API。

对于按 Pod 统计的资源指标（如 CPU），控制器从资源指标 API 中获取每一个 HorizontalPodAutoscaler 指定的 Pod 的度量值，如果设置了目标使用率，控制器获取每个 Pod 中的容器资源使用情况，并计算资源使用率。如果设置了 target 值，将直接使用原始数据（不再计算百分比）。接下来，控制器根据平均的资源使用率或原始值计算出扩缩的比例，进而计算出目标副本数。

需要注意的是，如果 Pod 某些容器不支持资源采集，那么控制器将不会使用该 Pod 的 CPU / 内存使用率。

## 参考资料

1. [Kubernetes Pod 水平自动扩缩（HPA）](https://kubernetes.io/zh/docs/tasks/run-application/horizontal-pod-autoscale/)
2. [Horizontal Pod Autoscaler 演练](https://kubernetes.io/zh/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)
3. [Kubernetes HPA 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#horizontalpodautoscaler-v2beta2-autoscaling)
