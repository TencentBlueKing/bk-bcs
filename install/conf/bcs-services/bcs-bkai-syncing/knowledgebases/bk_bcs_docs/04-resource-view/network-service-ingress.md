# 网络管理指南（Service、Ingress 与 Endpoints）

在 Kubernetes 中，Pod 是生命周期短暂的动态实例，其 IP 地址会随重建而变化。BCS 在【资源视图】→【网络】模块中提供了 **Service**、**Ingress** 和 **Endpoints** 的可视化管理，实现集群内外流量的服务发现与负载均衡。

---

## 一、Kubernetes Service 管理

**Service** 是将一组运行同一应用的 Pod 抽象为稳定访问入口的 API 对象，通常通过 **Label Selector（标签选择器）** 自动关联后端 Pod。

### 1. Service 类型（`spec.type`）说明

| Service 类型 | 作用范围与访问方式 | 典型使用场景 |
| :--- | :--- | :--- |
| **`ClusterIP`**（默认） | 分配一个集群内部虚拟 IP（ClusterIP），**仅限集群内部**通过 `<ClusterIP>:<port>` 或集群 DNS 域名 `<svc>.<ns>.svc.cluster.local` 访问。 | 集群内部微服务互访、数据库/缓存（如 Redis）内部暴露。 |
| **`Headless Service`**（`clusterIP: None`） | 不分配虚拟 IP，DNS 查询直接返回后端所有就绪 Pod 的真实 IP 列表；配合 StatefulSet 时为每个 Pod 生成固定域名 `<pod-name>.<svc>.<ns>.svc.cluster.local`。 | StatefulSet 有状态应用、客户端自行实现负载均衡或点对点通信。 |
| **`NodePort`** | 在 `ClusterIP` 基础上，在集群每个节点上开放一个静态端口（默认范围 `30000-32767`）。集群外部可通过 `<任意NodeIP>:<NodePort>` 访问服务。 | 开发测试环境外部快速访问、对接外部四层负载均衡器。 |
| **`LoadBalancer`** | 在公有云环境下自动创建并绑定云厂商负载均衡器（如腾讯云 CLB），分配外部 VIP 并将流量转发至后端服务。 | 公有云生产环境直接对外暴露 TCP/UDP/HTTP 服务。 |
| **`ExternalName`** | 通过返回 CNAME 记录将服务映射到集群外部域名（如 `db.example.com`），不创建代理或 Endpoints。 | 集群内应用统一通过 K8S 服务名访问集群外部托管数据库或第三方服务。 |

### 2. Service 配置项说明与 YAML 示例

- **`spec.selector`**：匹配后端 Pod 的标签（如 `app: servergame`）。当匹配上的 Pod 处于 `Ready` 状态时，其 IP 和端口会自动注册到同名 **Endpoints** 对象中。
- **`spec.ports[]`**：
  - `port`：Service 自身暴露的服务端口。
  - `targetPort`：转发到后端 Pod 容器上的目标端口（端口号或容器端口名称 `name`）。
  - `nodePort`：当类型为 `NodePort` 或 `LoadBalancer` 时在节点上监听的端口（不填则自动随机分配）。
  - `protocol`：支持 `TCP`（默认）、`UDP`、`SCTP`。

#### 示例 1：NodePort 类型 Service

```yaml
apiVersion: v1
kind: Service
metadata:
  name: servergame-svc
  namespace: default
spec:
  type: NodePort
  selector:
    app: servergame
  ports:
    - name: http
      protocol: TCP
      port: 8080
      targetPort: 8080
      nodePort: 30080
```

#### 示例 2：Headless Service（配合 StatefulSet）

```yaml
apiVersion: v1
kind: Service
metadata:
  name: game-stateful-headless
  namespace: default
spec:
  clusterIP: None
  selector:
    app: game-stateful
  ports:
    - name: tcp-port
      protocol: TCP
      port: 80
      targetPort: 80
```

---

## 二、Kubernetes 原生 Ingress 管理

**Ingress** 用于管理集群外部访问集群内服务的七层（HTTP / HTTPS）路由规则，提供基于域名（Host）和 URL 路径（Path）的虚拟主机路由、负载均衡及 SSL/TLS 卸载能力。

> **注意**：
> 1. 集群中必须部署对应的 **Ingress Controller**（如社区 `ingress-nginx` 或云厂商原生 Ingress 控制器），Ingress 资源规则才会生效。
> 2. 若需在多云环境下使用增强版四层/七层 CLB 控制、StatefulSet 端口段映射或动态端口池（PortPool），请使用 BCS 自研的 [bcs-ingress-controller](../06-custom-addons-and-game/bcs-ingress-controller.md)（API Group 为 `networkextension.bkbcs.tencent.com/v1`）。

### 1. 原生 Ingress（`networking.k8s.io/v1`）配置示例

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: demo-web-ingress
  namespace: default
  annotations:
    kubernetes.io/ingress.class: "nginx"
spec:
  tls:
    - hosts:
        - demo.bcs.example.com
      secretName: demo-tls-secret
  rules:
    - host: demo.bcs.example.com
      http:
        paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: servergame-svc
                port:
                  number: 8080
```

### 2. 核心配置项解析

- **基于域名的虚拟主机（`spec.rules[].host`）**：指定外部访问的 HTTP/HTTPS 域名（如 `demo.bcs.example.com`），匹配该域名及路径 `path` 的请求将被转发至对应的后端 `service.name` 和 `service.port`。
- **路径匹配类型（`pathType`）**：
  - `Prefix`：基于以 `/` 分隔的 URL 路径前缀匹配。
  - `Exact`：精确匹配 URL 路径，区分大小写。
  - `ImplementationSpecific`：匹配逻辑取决于具体的 IngressClass 实现。
- **HTTPS / TLS 证书配置（`spec.tls`）**：
  - 引用同一命名空间下包含 TLS 证书（`tls.crt`）和私钥（`tls.key`）的 Secret（类型为 `kubernetes.io/tls` 或 `Opaque`）：

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: demo-tls-secret
  namespace: default
type: kubernetes.io/tls
data:
  tls.crt: <Base64编码的证书内容>
  tls.key: <Base64编码的私钥内容>
```

---

## 三、Endpoints 查看与网络连通性排查

在【资源视图】→【网络】→【Endpoints】中，可以查看每个 Service 实际关联的后端 Pod IP 与端口列表。

**常见排障技巧**：
- 若通过 Service 无法访问应用，首先检查同名 **Endpoints** 列表是否为空：
  - 若 **Endpoints 为空**：说明 Service 的 `spec.selector` 标签与 Pod 的 `metadata.labels` 不匹配，或者后端 Pod 尚未通过 `readinessProbe` 就绪探针检查（处于 `NotReady` 状态）。
  - 若 **Endpoints 正常有 Pod IP**：说明 Service 标签关联正常，请进一步检查 Service 的 `targetPort` 是否与容器内部实际监听的端口一致。
