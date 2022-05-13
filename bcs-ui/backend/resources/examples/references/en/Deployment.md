# Deployment

> Deployment (Deploy) provides declarative update capabilities for Pods and ReplicaSets.

## What is a Deployment?

Deployment is a new concept introduced in Kubernetes v1.2, the purpose is to better solve the problem of Pod arrangement; Deployment uses ReplicaSet internally to achieve, in terms of specific operations, Deployment can be regarded as an upgraded version of Replication Controller (RC). .

Deployment can track and know the deployment progress of the Pods it manages at any time, which is the biggest upgrade compared to RC and enhances the ability to master the deployment process.

## Using Deployment

Common deployment usage scenarios are:

- [Create a Deployment to deploy Pods](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#creating-a-deployment), the generated ReplicaSet creates Pods in the background, by checking the online of the ReplicaSet The status verifies that the Pod deployment was successful.
- via [Update the Deployment's PodTemplateSpec to declare the new state of the Pod](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#updating-a-deployment). A new ReplicaSet will be created, and Pods will migrate from the old ReplicaSet to the new ReplicaSet at a controlled rate.
- If the current state of the Deployment is unstable, there is an option to [roll back to an earlier Deployment version](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#rolling-back-a-deployment ) . Each rollback updates the Deployment's revision.
- [Pause Deployment](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#pausing-and-resuming-a-deployment) to apply multiple modifications to the PodTemplateSpec, then resume its Execute to start a new go-live version.
- Use [View Deployment Status](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#deployment-status) to determine if the rollout process is stalled or failed.
- [scaling a Deployment to take on more load](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#scaling-a-deployment) or [cleaning up older ReplicaSets that are no longer needed](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/#clean-up-policy).

## References

1. [Kubernetes / Workloads / Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
2. [kubernetes Deployment field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#deployment-v1-apps)
