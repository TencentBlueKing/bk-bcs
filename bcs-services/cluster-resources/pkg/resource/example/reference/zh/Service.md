# Service

> Service（SVC）将运行在一组 Pods 上的应用程序公开为网络服务的抽象方法

## 什么是 Service

Kubernetes Service 定义了这样一种抽象：逻辑上的一组 Pod，一种可以访问它们的策略，通常称为微服务。Service 所针对的 Pods 集合通常是通过选择算符来确定的。

举个例子，考虑一个图片处理后端，它运行了 3 个副本。这些副本是可互换的，即前端不需要关心它们调用了哪个后端副本。然而组成这一组后端程序的 Pod 实际上可能会发生变化，前端客户端不应该也没必要知道，而且也不需要跟踪这一组后端的状态。

Service 定义的抽象能够解耦这种关联。

## 定义 Service

Service 在 Kubernetes 中是一类对象，就像所有的 REST 对象一样，我们可以通过 POST 方法，请求 API Server 来创建新的实例。

例如，假定有一组 Pod，它们对外暴露了 9376 端口，同时还被打上 `app=MyApp` 标签：

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: MyApp
  ports:
    - protocol: TCP
      port: 80
      targetPort: 9376
```

上述配置创建一个名称为 `my-service` 的 Service 对象，它会将请求代理到使用 TCP 端口 9376，并且具有标签 `app=MyApp` 的 Pod 上。

Kubernetes 为该服务分配一个 IP 地址（有时称为 `集群 IP`），该 IP 地址由服务代理使用。

服务选择算符的控制器不断扫描与其选择器匹配的 Pod，然后将所有更新发布到称为 `my-service` 的 Endpoint 对象。

需要注意的是，Service 能够将一个接收 port 映射到任意的 targetPort。默认情况下，targetPort 将被设置为与 port 字段相同的值。

Pod 中的端口定义是有名字的，你可以在服务的 targetPort 属性中引用这些名称。这为部署和发展服务提供了很大的灵活性。

由于许多服务需要公开多个端口，因此 Kubernetes 在服务对象上支持多个端口定义。每个端口定义可以具有相同的 protocol，也可以具有不同的协议。

## 服务类型

Kubernetes ServiceTypes 允许指定你所需要的 Service 类型，默认是 ClusterIP。

Type 的取值以及行为如下：

- ClusterIP：通过集群的内部 IP 暴露服务，选择该值时服务只能够在集群内部访问。

- NodePort：通过每个节点上的 IP 和静态端口（NodePort）暴露服务。NodePort 服务会路由到自动创建的 ClusterIP 服务。通过请求 `<节点 IP>:<节点端口>`，可以从集群的外部访问 NodePort 服务。

- LoadBalancer：使用云提供商的负载均衡器向外部暴露服务。外部负载均衡器可以将流量路由到自动创建的 NodePort 服务和 ClusterIP 服务上。

- ExternalName：通过返回 CNAME 和对应值，可以将服务映射到 externalName 字段的内容（例如，foo.bar.example.com）。

> Note: 你需要使用 kube-dns 1.7 及以上版本或者 CoreDNS 0.0.8 及以上版本才能使用 ExternalName 类型。

你也可以使用 Ingress 来暴露自己的服务。Ingress 不是一种服务类型，但它充当集群的入口点。它可以将路由规则整合到一个资源中，因为它可以在同一IP地址下公开多个服务。

## 参考资料

1. [Kubernetes / 网络服务 / Service](https://kubernetes.io/zh/docs/concepts/services-networking/service/)
2. [Kubernetes Service 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#service-v1-core)
