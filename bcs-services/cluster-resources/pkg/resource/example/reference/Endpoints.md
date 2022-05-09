# Endpoints

> Service 与 Pod 之间的桥梁

## 什么是 Endpoints ？

服务（Service）与 Pod 之间并不是直接相连的，有一种资源介于两者之间，这就是 Endpoints 资源。

Endpoints 资源是一个暴露服务的 IP 地址和端口的列表，尽管我们在 Service 的 spec 中定义了 Pod 选择其，但是在重定向传入连接时并不会直接使用它。相反，选择器被用于构建 IP 和端口列表，然后存储在 Endpoints 资源中。当客户端连接到服务是，服务代理根据策略选择这些 IP 和端口对中的一个，并将传入连接重定向到在该位置监听的服务器。

## 使用 Endpoints

**一般来说，我们不需要手动管理 Endpoints 资源**，它应该根据 Service 定义，由集群自动生成，管理及维护。但是，如果 [Service 没有定义选择算符](https://kubernetes.io/zh/docs/concepts/services-networking/service/#services-without-selectors) ，则 Endpoints 不会被自动创建，我们便需要手动对其进行创建与更新。

以下是一个手动管理 Endpoints 资源的例子

1.  创建没有选择器的服务，我们定义一个没有选择器的服务，它会接收端口 80 上的传入连接。

```yaml
apiVersion: v1
kind: Service
metadata:
  name: external-service
spec:
  ports:
    - port: 80
```

2. 为没有选择器的服务创建 Endpoints 资源

```yaml
apiVersion: v1
kind: Endpoints
metadata:
  name: external-service
subsets:
  - addresses:
      - ip: 1.1.1.1 # 服务将连接重定向到 Endpoints 的 IP 地址
      - ip: 2.2.2.2
    ports:
      - port: 80 # Endpoint 的目标端口
```

需要注意的是，Endpoints 对象需与服务保持相同的名称，并包含该服务的目标 IP 地址和端口列表。服务和 Endpoints 资源都发布到服务器后，这时服务就可以像具有 Pod 选择器那样正常可用。在服务创建后创建的容器将包含服务的环境变量，并且与其 IP:port 对的所有连接，都将在服务端点之间进行负载均衡。

> 手动创建的 Endpoints 不会被自动维护，如果发生 Pod 重新调度等事件，导致 Endpoints 中配置的 IP 失效，可能导致服务不可用。

## 参考资料

1. [Kubernetes Endpoints 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#endpoints-v1-core)
