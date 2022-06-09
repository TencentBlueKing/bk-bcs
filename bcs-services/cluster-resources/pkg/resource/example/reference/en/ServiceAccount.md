# ServiceAccount

> ServiceAccount (SA) is the identity of the process running in the Pod to access other resources of the cluster

## User account and service account

Kubernetes distinguishes the concepts of user accounts and service accounts for the following reasons:
- User accounts are for humans. Service accounts are for processes, which run in pods.
- User accounts are intended to be global. Names must be unique across all namespaces of a cluster. Service accounts are namespaced.
- Typically, user accounts involve complex business processes, and their creation also requires special permissions. Service accounts, on the other hand, are more lightweight, allowing users to create service accounts for specific tasks to comply with the principle of minimizing permissions.

## Using ServiceAccount

When each namespace is created, a ServiceAccount named `default` will be created under the namespace by default. If the newly created Pod does not explicitly specify a ServiceAccount, the `default` ServiceAccount will be used.

You can explicitly specify a ServiceAccount for a Pod by editing the `spec.serviceAccountName` field in the Pod Manifest / PodTemplate.

## Manage ServiceAccount

### Create ServiceAccount

You can quickly create a ServiceAccount by delivering the following configuration to the cluster:

````yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot
````

Execute `kubectl get sa build-robot -o yaml`, the output is similar to:

````yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: default
  name: build-robot
  ...
secrets:
- name: build-robot-token-bvbk5
````

It can be observed that when a ServiceAccount is created, a Secret (build-robot-token-bvbk5) is automatically created and bound to the ServiceAccount.

### Add ImagePullSecrets for ServiceAccount

You can execute the following command to create an ImagePullSecrets in the cluster:

```shell
kubectl create secret docker-registry myregistrykey \
  --docker-server=DUMMY_SERVER \
  --docker-username=DUMMY_USERNAME \
  --docker-password=DUMMY_DOCKER_PASSWORD \
  --docker-email=DUMMY_DOCKER_EMAIL
````

Edit the ServiceAccount that needs to be bound:

````yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  ...
secrets:
  ...
imagePullSecrets:
- name: myregistrykey
````

## References

1. [Manage Service Accounts](https://kubernetes.io/docs/reference/access-authn-authz/service-accounts-admin/)
2. [Configure Service Account for Pod](https://kubernetes.io/docs/tasks/configure-pod-container/configure-service-account/)
3. [Kubernetes ServiceAccount field description](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#serviceaccount-v1-core)