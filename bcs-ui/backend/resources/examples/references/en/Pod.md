# Pod

> A Pod (Po) is the smallest deployable computing unit that can be created and managed in Kubernetes.

## What is a Pod?

A Pod is a group (one or more) of containers that share storage, networking, and a declaration of how to run those containers. Content in a Pod is always colocated and scheduled together, running in a shared context.

A Pod's shared context includes a set of Linux namespaces, control groups (cgroups), and possibly some other aspects of isolation, the techniques used to isolate Docker containers. In the context of Pods, each individual application may further enforce isolation.

A Pod can be thought of as an application-specific "logical host" that contains one or more application containers that are relatively tightly coupled together. In addition to application containers, Pods can also contain Init containers that run during Pod startup.

## Using Pods

Generally speaking, we rarely create one or more Pods directly in a Kubernetes cluster because Pods are designed to be temporary, disposable entities. After the Pod is created by the user or the controller, it will be automatically scheduled to run on the appropriate cluster node. After the Pod is executed, deleted, expelled due to insufficient resources, or the node fails, the Pod will remain running on the node.

Commonly used pod controllers have the following categories:

- [Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/): API object for managing application replicas, defines Pod template information, number of replicas, running configuration, etc., suitable for managing clusters Stateless applications (such as APIs, etc.) in .

- [StatefulSet](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/): API object for managing stateful applications, used to manage the deployment and scaling of Pod collections, and for these Pods Provides persistent storage and persistent identifiers.

- [DaemonSet](https://kubernetes.io/docs/concepts/workloads/controllers/daemonset/): Make sure all or some nodes run a copy of Pod, generally used to run `cluster/log collection/ Monitor the daemon of`.

- [Job](https://kubernetes.io/docs/concepts/workloads/controllers/job/): Create one or more Pods to perform one-time tasks; if you need to repeat the Job, you can use [CronJob](https://kubernetes.io/docs/concepts/workloads/controllers/cron-jobs/).

Of course, there are some scenarios that are more suitable for creating independent Pods, such as:

- Start a single Pod for debugging and temporarily verify the service's scenario

- Scenarios for debugging a single Pod configuration against cluster node status

## Pod update and replacement

When the Pod template of a Pod controller is changed, the controller will create a new Pod according to the latest template, and recycle the old Pod objects, without updating or patching existing Pods.

Kubernetes does not prohibit direct management of pods, allowing in-place update operations to be performed on certain fields of running pods. However, operations like patch and replace have the following limitations:

- The vast majority of a Pod's metadata is immutable. For example, its `namespace`, `name`, `uid` or `creationTimestamp` fields cannot be changed; the `generation` field is special. If you update this field, you can only increase the field value and not decrease it.

- If `metadata.deletionTimestamp` is already set, no new entries can be added to the `metadata.finalizers` list.

- Pod updates cannot change fields other than `spec.containers[*].image`, `spec.initContainers[*].image`, `spec.activeDeadlineSeconds` or `spec.tolerations`. For `spec.tolerations`, you are only allowed to add new entries to it.

- When updating the `spec.activeDeadlineSeconds` field, the following two update operations are allowed:
  - if the field is not already set, it can be set to a positive number;
  - If the field is already set to a positive number, it can be set to a smaller, non-negative integer.

## References

1. [Kubernetes / Workloads / Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
2. [kubernetes Pod field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#pod-v1-core)
