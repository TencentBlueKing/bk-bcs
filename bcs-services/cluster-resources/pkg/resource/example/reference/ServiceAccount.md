# ServiceAccount

> ServiceAccount（SA）是 Pod 中运行的进程访问集群其他资源的身份

## 用户账号与服务账号

Kubernetes 区分用户账号和服务账号的概念，主要基于以下原因：
- 用户账号是针对人而言的。 服务账号是针对运行在 Pod 中的进程而言的。
- 用户账号是全局性的，其名称在整个集群中都是唯一的；而服务账号是名字空间作用域的。
- 通常情况下，用户账号会涉及到复杂的业务流程，其创建也需要特殊权限。而服务账号更轻量，允许用户为具体的任务创建服务账号以遵从权限最小化原则。

## 使用 ServiceAccount

每个命名空间创建时，会默认在该命名空间下同时创建名为 `default` 的 ServiceAccount，新建的 Pod 若没有显式指定 ServiceAccount，则会使用 `default` ServiceAccount。

您可以通过编辑 Pod Manifest / PodTemplate 中的 `spec.serviceAccountName` 字段，显式为 Pod 指定 ServiceAccount。

## 管理 ServiceAccount

### 创建 ServiceAccount

您可以通过下发以下配置到集群，快速创建一个 ServiceAccount：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: build-robot
```

执行 `kubectl get sa build-robot -o yaml`，其输出类似于：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  namespace: default
  name: build-robot
  ...
secrets:
- name: build-robot-token-bvbk5
```

可以观察到，创建 ServiceAccount 时，会自动创建一个 Secret（build-robot-token-bvbk5）并绑定到该 ServiceAccount。

### 为 ServiceAccount 添加 ImagePullSecrets

您可以执行以下命令，在集群中创建一个 ImagePullSecrets：

```shell
kubectl create secret docker-registry myregistrykey \
  --docker-server=DUMMY_SERVER \
  --docker-username=DUMMY_USERNAME \
  --docker-password=DUMMY_DOCKER_PASSWORD \
  --docker-email=DUMMY_DOCKER_EMAIL
```

编辑需要绑定的 ServiceAccount 即可：

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  ...
secrets:
  ...
imagePullSecrets:
- name: myregistrykey
```

## 参考资料

1. [管理服务帐号](https://kubernetes.io/zh/docs/reference/access-authn-authz/service-accounts-admin/)
2. [为 Pod 配置服务账户](https://kubernetes.io/zh/docs/tasks/configure-pod-container/configure-service-account/)
3. [Kubernetes ServiceAccount 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#serviceaccount-v1-core)
