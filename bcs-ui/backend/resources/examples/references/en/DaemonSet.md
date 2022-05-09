# DaemonSet

> DaemonSet (DS) stands for `daemon process set`, which is a collection of a series of daemon processes, responsible for managing the daemon processes of the cluster. The DaemonSet ensures that all (or some) nodes run a replica of the Pod. When a node joins the cluster, a Pod is also added for them. These Pods are also recycled when a node is removed from the cluster. Deleting a DaemonSet will delete all Pods it created.

## Using DaemonSet

Common DaemonSet usage scenarios are:

- Run `Cluster Storage/Network Daemon` like ceph, glusterd, nginx etc running on each node
- Run `Log Collection Daemon`, such as logstash, fluentd, etc. running on each node
- Run `Service/Performance/Metrics Monitoring Daemon` such as PrometheusNodeExporter, collectd, Datadog, etc.

A simple usage is to start a DaemonSet on all nodes for each type of daemon. A slightly more complex usage is to deploy multiple DaemonSets for the same daemon; each with different flags and different memory, CPU requirements for different hardware types.

### Pods running DaemonSet only on some Nodes?

DaemonSet creates a new Pod on each cluster node by default, but there are ways to specify nodes:

- nodeSelector: By configuring the `spec.template.spec.nodeSelector` of the DaemonSet, you can make the DaemonSet create Pods only on nodes that can match the NodeSelector.
- affinity: Similarly, `spec.template.spec.affinity` can be used to configure `spec.template.spec.affinity` to use node affinity to control Pods to be deployed only on some nodes.

### DaemonSet Pod scheduling

Normally, the Kubernetes scheduler chooses which machine a Pod runs on. However, DaemonSet Pods are created and scheduled by DaemonController and are therefore different from normal Pods:

- Deterministic Pod scheduling: The `spec.nodeName` is specified when a DaemonSet Pod is created, so its scheduling result is determinable.
- Inconsistency in Pod behavior: A normal Pod is in the Pending state when it is created and waiting for scheduling, and the DaemonSet Pod will not be in the Pending state after it is created.
- Automatically add `node.kubernetes.io/unschedulable:NoSchedule` tolerance to DaemonSet Pods. When scheduling DaemonSet Pods, the `unschedulable` state of the node is not concerned.
- DaemonSet Pods can be created even if the Kubernetes scheduler has not started, which is very helpful for cluster startup.

### Stain and Tolerance

Although Daemon Pods follow the [Taints and Toleration](https://kubernetes.io/docs/concepts/scheduling-eviction/taint-and-toleration/) rules, depending on the relevant characteristics, the controller will automatically set the following tolerances Add to DaemonSet Pod:

| Tolerance Key Name                       | Effect     | Version | Description                                                                                                       |
| ---------------------------------------- | ---------- | ------- | ----------------------------------------------------------------------------------------------------------------- |
| `node.kubernetes.io/not-ready`           | NoExecute  | 1.13+   | DaemonSet Pods will not be evicted when something like a network disconnect causes node issues.                   |
| `node.kubernetes.io/unreachable`         | NoExecute  | 1.13+   | DaemonSet Pods will not be evicted when there is a problem with the node like a network disconnection.            |
| `node.kubernetes.io/disk-pressure`       | NoSchedule | 1.8+    | DaemonSet Pods can tolerate disk pressure properties when scheduled by the default scheduler.                     |
| `node.kubernetes.io/memory-pressure`     | NoSchedule | 1.8+    | DaemonSet Pods can tolerate memory pressure properties when scheduled by the default scheduler.                   |
| `node.kubernetes.io/unschedulable`       | NoSchedule | 1.12+   | DaemonSet Pods can tolerate the unschedulable property set by the default scheduler.                              |
| `node.kubernetes.io/network-unavailable` | NoSchedule | 1.12+   | DaemonSet can tolerate the network-unavailable property set by the default scheduler when using the host network. |

## References

1. [Kubernetes / Workloads / DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/)
2. [kubernetes DaemonSet field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#daemonset-v1-apps)
