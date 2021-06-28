# bcs-networkpolicy

NetworkPolicy 是 Kubernetes 官方定义的一种资源，主要用于对 Pod 进行网络策略设置，比如：允许访问/禁止访问等策略。

官方仅对其进行定义，实际的实现需要依赖 CNI 插件本身来实现，如业界比较有名的 CNI 插件 Calico 等（Flannel 不支持 NetworkPolicy）。

## 解决的问题

BCS 项目既支持 Kubernetes 集群，同时也支持 Mesos 集群。BCS 初衷希望能够通过相同资源定义方式，来拉平 Kubernetes 与 Mesos 在使用上的差异性。

在 NetworkPolicy 的处理上，Mesos 集群使用 Kubernetes 的资源定义来做网络策略实现。即可以通过 Kubernetes 的 NetworkPolicy 资源来定义 Mesos 集群的网络策略。


## 使用文档

资源定义完全沿用 Kubernetes，参考文档：[Network Policies | Kubernetes](https://kubernetes.io/docs/concepts/services-networking/network-policies/)

## 使用示例

如下列举了部分网络策略的定义规则

### 拒绝所有入网流量

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-ingress
  namespace: default
spec:
  podSelector: {}
  policyTypes:
  - Ingress
```

对 `default` 命名空间下的所有 Pod，拒绝所有入网流量，即不允许除自身命名空间 Pod 外的 IP 访问，包括 Node 节点。

### 允许所有入网流量

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-ingress
  namespace: default
spec:
  podSelector: {}
  ingress:
  - {}
  policyTypes:
  - Ingress
```

允许任意 IP 访问 `default` 命名空间下得 Pod。

### 拒绝所有出网流量

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: default-deny-egress
  namespace: default
spec:
  podSelector: {}
  policyTypes:
  - Egress
```

对 `default` 命名空间下的所有 Pod，不允许它们访问除本身命名空间 Pod 外的其它任意 IP。

### 允许所有出网流量

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: allow-all-egress
spec:
  podSelector: {}
  egress:
  - {}
  policyTypes:
  - Egress
```

允许 `default` 命名空间下所有 Pod 访问其它 IP。

### 完整示例

1. 首先 `spec.podSelector` 选中网络策略命中的 Pod;


2. 参数 `ingress` 表示命中的 Pod 的入网策略
    - `from.ipBlock` 表示禁止 IP 段进行访问，`except` 表示除了 IP 段之外的 IP 禁止访问；
    - `from.namespaceSelector` 和 `from.podSelector` 选中允许 Pod 访问，如果 namespaceSelector 为空，则选中 NetworkPolicy 当前命名空间
    - `ports` 表示禁止掉某些端口的访问

3. 参数 `egress` 表示命中的 Pod 的出网策略
    - `to.ipBlock` 表示禁止出网到 IP 段；
    - `to.ports` 表示禁止出网到 IP 段端口

```yaml
---
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: test-network-policy
  namespace: default
spec:
  podSelector:
    matchLabels:
      role: db
  policyTypes:
    - Ingress
    - Egress
  ingress:
    - from:
      - ipBlock:
          cidr: 172.17.0.0/16
          except:
            - 172.17.1.0/24
        - namespaceSelector:
            matchLabels:
              project: myproject
        - podSelector:
            matchLabels:
              role: frontend
      ports:
        - protocol: TCP
          port: 6379
  egress:
    - to:
        - ipBlock:
            cidr: 10.0.0.0/24
      ports:
        - protocol: TCP
          port: 5978
```

