# StatefulSet

> StatefulSet (STS) is a workload API object used to manage stateful applications. It manages the deployment and scaling of a collection of Pods and provides persistent storage and persistent identifiers for these Pods.

## What is a StatefulSet?

Similar to Deployment, StatefulSet manages a set of Pods based on the same container specification. But unlike Deployments, StatefulSets maintain a sticky ID for each of their Pods. These Pods are created based on the same protocol, but are not interchangeable, that is, no matter how they are scheduled, each Pod has a permanent ID.

If you want to use storage volumes to provide persistent storage for your workloads, you can use a StatefulSet as part of your solution. Although individual Pods in a StatefulSet can still fail, persistent Pod identifiers make it easier to match existing volumes with new Pods that replace failed Pods.

## Using StatefulSet

Common StatefulSet usage scenarios are:

- Stable, unique network identifier: can be used to discover other members in the cluster, in general, if the name of the STS is `nginx` and the number of replicas is 3, its Pod name will be `nginx-0, nginx- 1, nginx-2`, after the Pod is rescheduled, its name and HostName will not change.
- Stable, persistent storage: implemented through [PV & PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/), after Pod rescheduling, you can still access the same persistent data; for security reasons, persistent data is not cleaned up by default when a Pod is deleted.
- Orderly, elegant deployment and scaling: The Pods of a StatefulSet are ordered, and when deploying or expanding, they must be performed in order according to the defined order (from 0 to N-1, each Pod runs on the condition of the previous one ( If any), the Pod status is `Running/Ready`); vice versa, reduce the number of Pods from N-1 to 0.
- Orderly, automatic rolling updates.

## limit

- Storage for a given Pod must be provided by [PersistentVolume driven](https://github.com/kubernetes/examples/blob/master/staging/persistent-volume-provisioning/README.md) based on the requested `StorageClass` , or pre-supplied by the administrator.
- For the sake of data security, deleting or shrinking a StatefulSet will not delete its associated storage volume.
- StatefulSet currently requires [headless-services](https://kubernetes.io/docs/concepts/services-networking/service/#headless-services) to be responsible for Pod's network identity. The user is responsible for creating this service.
- The StatefulSet does not provide any guarantees that the Pod will be terminated when the StatefulSet is deleted. In order to achieve orderly and decent termination of Pods in a StatefulSet, the StatefulSet's number of Pod replicas can be adjusted to 0 before deletion.
- Use [Rolling Updates](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#rolling-updates) (`OrderedReady`) when default [Pod Management Policies](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#pod-management-policies), may enter requiring [human intervention (force rollback)](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/#forced-rollback) to fix the broken state.

## References

1. [Kubernetes / Workloads / StatefulSets](https://kubernetes.io/docs/concepts/workloads/controllers/statefulset/)
2. [kubernetes StatefulSet field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#statefulset-v1-apps)
