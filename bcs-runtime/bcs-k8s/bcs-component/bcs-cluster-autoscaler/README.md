# Cluster Autoscaler

# Introduction

Cluster Autoscaler is a tool that automatically adjusts the size of the Kubernetes cluster when one of the following conditions is true:
* there are pods that failed to run in the cluster due to insufficient
  resources.
* there are nodes in the cluster that have been underutilized for an extended period of time and their pods can be placed on other existing nodes.

# FAQ/Documentation

An FAQ is available [HERE](./FAQ.md).

# Releases

We recommend using Cluster Autoscaler with the Kubernetes master version for which it was meant. The below combinations have been tested on GCP. We don't do cross version testing or compatibility testing in other environments. Some user reports indicate successful use of a newer version of Cluster Autoscaler with older clusters, however, there is always a chance that it won't work as expected.

Starting from Kubernetes 1.12, versioning scheme was changed to match Kubernetes minor releases exactly.

Cluster Autoscaler is designed to run on Kubernetes master node. This is the
default deployment strategy on GCP.
It is possible to run a customized deployment of Cluster Autoscaler on worker nodes, but extra care needs
to be taken to ensure that Cluster Autoscaler remains up and running. Users can put it into kube-system
namespace (Cluster Autoscaler doesn't scale down node with non-mirrored kube-system pods running
on them) and set a `priorityClassName: system-cluster-critical` property on your pod spec
(to prevent your pod from being evicted).

Supported cloud providers:
* GCE https://kubernetes.io/docs/concepts/cluster-administration/cluster-management/
