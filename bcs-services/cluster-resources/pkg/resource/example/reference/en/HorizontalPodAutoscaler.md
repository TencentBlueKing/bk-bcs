# HorizontalPodAutoscaler

> HorizontalPodAutoscaler (HPA) can automatically update workload resources (such as Deployments or StatefulSets) in order to automatically scale workloads to meet demand.

## How HPA works

![img](https://d33wubrfki0l68.cloudfront.net/4fe1ef7265a93f5f564bd3fbb0269ebd10b73b4e/1775d/images/docs/horizontal-pod-autoscaler.svg)

HorizontalPodAutoscaler controls the scale of Deployment and its ReplicaSet, so as to achieve the effect of controlling the number of Pods according to resource usage; Pod horizontal auto-scaling is implemented as an intermittently running control loop, with a default interval of 15 seconds.

During each time period, the controller manager queries resource utilization based on the metrics specified in each HorizontalPodAutoscaler definition. The controller manager finds the target resource defined by scaleTargetRef , then selects Pods based on the `.spec.selector` tag of the target resource, and gets the metrics API from the resource metrics API or custom metrics.

For resource metrics (such as CPU) counted by Pod, the controller obtains the metric value of each Pod specified by the HorizontalPodAutoscaler from the resource metrics API. If the target usage rate is set, the controller obtains the container resource usage in each Pod. And calculate the resource usage. If the target value is set, the raw data will be used directly (no more percentage calculations). Next, the controller calculates the scaling ratio based on the average resource usage or the original value, and then calculates the target number of replicas.

It should be noted that if some containers of a Pod do not support resource harvesting, the controller will not use the CPU/memory usage of that Pod.

## References

1. [Kubernetes Pod Horizontal Autoscale (HPA)](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale/)
2. [Horizontal Pod Autoscaler Walkthrough](https://kubernetes.io/docs/tasks/run-application/horizontal-pod-autoscale-walkthrough/)
3. [Kubernetes HPA field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#horizontalpodautoscaler-v2beta2-autoscaling)