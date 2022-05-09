# Ingress

> Ingress 是对集群中服务的外部访问进行管理的 API 对象，可以提供负载均衡、SSL 终结和基于名称的虚拟托管，其典型的访问方式是 HTTP。

## 什么是 Ingress ？

Ingress 公开了从集群外部到集群内服务的 HTTP 和 HTTPS 路由。流量路由由 Ingress 资源上定义的规则控制。

下面是一个将所有流量都发送到同一 Service 的简单 Ingress 示例：

![img](https://i.stack.imgur.com/qF2u2.png)

可以将 Ingress 配置为服务提供外部可访问的 URL、负载均衡流量、终止 SSL/TLS，以及提供基于名称的虚拟主机等能力。[Ingress 控制器](https://kubernetes.io/zh/docs/concepts/services-networking/ingress-controllers/) 通常负责通过负载均衡器来实现 Ingress。

Ingress 不会公开任意端口或协议。将 HTTP 和 HTTPS 以外的服务公开到 Internet 时，通常使用 [Service.Type=NodePort](https://kubernetes.io/zh/docs/concepts/services-networking/service/#type-nodeport) 或 [Service.Type=LoadBalancer](https://kubernetes.io/zh/docs/concepts/services-networking/service/#loadbalancer) 类型的服务。

## 使用 Ingress

> 注意！你需要具有 [Ingress 控制器](https://kubernetes.io/zh/docs/concepts/services-networking/ingress-controllers/) 才能满足 Ingress 的需求，仅创建 Ingress 资源本身没有任何效果。
>
> 你可能需要部署 Ingress 控制器，例如 [ingress-nginx](https://kubernetes.github.io/ingress-nginx/deploy/) 。也可以选择其他类型的控制器。

一个简单的 Ingress 资源示例如下：

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: minimal-ingress
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
    - host: "*.foo.com"
      http:
        paths:
          - pathType: Prefix
            path: /foo
            backend:
              service:
                name: svc-alpha
                port:
                  number: 80
```

> 注：Ingress 经常使用注解（annotations）来配置一些选项，具体取决于 Ingress 控制器，如 [重写目标注解](https://github.com/kubernetes/ingress-nginx/blob/main/docs/examples/rewrite/README.md) 。不同的 Ingress 控制器支持不同的注解。用户需要查看对应的说明文档以了解支持哪些注解。

Ingress [规约](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/api-conventions.md#spec-and-status) 提供了配置负载均衡器或者代理服务器所需的所有信息，其中包含与所有传入请求匹配的规则列表。Ingress 资源仅支持用于转发 HTTP 流量的规则。

### Ingress 规则

每个 HTTP 规则（spec.rules）都包含以下信息：

- 可选的 host。此示例中指定 host（\*.foo.com），因此适用该 rule 的 Host 流量将会转发到对应的 Service；若未指定 host，因此该规则适用于通过指定 IP 地址的所有入站 HTTP 通信。
- 路径列表 paths（如 /foo），每个路径都有一个由 `Service Name & Port` 定义的关联后端。在负载均衡器将流量定向到引用的服务之前，主机和路径都必须匹配传入请求的内容。
- backend 是 Service 服务和端口名称的组合。与规则的 host 和 path 匹配的对 Ingress 的 HTTP（和 HTTPS）请求将发送到列出的 backend。（注：不同 apiVersion 的 Ingress Backend 配置结构略有差异）

通常在 Ingress 控制器中会配置 [defaultBackend（默认后端）](https://kubernetes.io/zh/docs/concepts/services-networking/ingress/#default-backend) ，以服务于任何不符合规则中 path 的请求。

### 路径类型

Ingress 中的每个路径都需要有对应的路径类型。未明确设置 pathType 的路径无法通过合法性检查。当前支持的路径类型有三种：

- `ImplementationSpecific`：对于这种路径类型，匹配方法取决于 [IngressClass](https://kubernetes.io/zh/docs/concepts/services-networking/ingress/#ingress-class) 。具体实现可以将其作为单独的 pathType 处理或者与 Prefix 或 Exact 类型作相同处理。
- `Exact`：精确匹配 URL 路径，且区分大小写。
- `Prefix`：基于以 / 分隔的 URL 路径前缀匹配。匹配区分大小写，并且对路径中的元素逐个完成。路径元素指的是由 / 分隔符分隔的路径中的标签列表。如果每个 `p` 都是请求路径 `p` 的元素前缀，则请求与路径 `p` 匹配。

> 说明： 如果路径的最后一个元素是请求路径中最后一个元素的子字符串，则不会匹配 （例如：/foo/bar 匹配 /foo/bar/baz, 但不匹配 /foo/barbaz）。

#### 示例

| 类型   | 路径                        | 请求路径     | 匹配与否？             |
| ------ | --------------------------- | ------------ | ---------------------- |
| Prefix | /                           | （所有路径） | 是                     |
| Exact  | /foo                        | /foo         | 是                     |
| Exact  | /foo                        | /bar         | 否                     |
| Exact  | /foo                        | /foo/        | 否                     |
| Exact  | /foo/                       | /foo         | 否                     |
| Prefix | /foo                        | /foo, /foo/  | 是                     |
| Prefix | /foo/                       | /foo, /foo/  | 是                     |
| Prefix | /aaa/bb                     | /aaa/bbb     | 否                     |
| Prefix | /aaa/bbb                    | /aaa/bbb     | 是                     |
| Prefix | /aaa/bbb/                   | /aaa/bbb     | 是，忽略尾部斜线       |
| Prefix | /aaa/bbb                    | /aaa/bbb/    | 是，匹配尾部斜线       |
| Prefix | /aaa/bbb                    | /aaa/bbb/ccc | 是，匹配子路径         |
| Prefix | /aaa/bbb                    | /aaa/bbbxyz  | 否，字符串前缀不匹配   |
| Prefix | /, /aaa                     | /aaa/ccc     | 是，匹配 /aaa 前缀     |
| Prefix | /, /aaa, /aaa/bbb           | /aaa/bbb     | 是，匹配 /aaa/bbb 前缀 |
| Prefix | /, /aaa, /aaa/bbb           | /ccc         | 是，匹配 / 前缀        |
| Prefix | /aaa                        | /ccc         | 否，使用默认后端       |
| 混合   | /foo (Prefix), /foo (Exact) | /foo         | 是，优选 Exact 类型    |

#### 多重匹配

在某些情况下，Ingress 中的多条路径会匹配同一个请求。 这种情况下最长的匹配路径优先。 如果仍然有两条同等的匹配路径，则精确路径类型优先于前缀路径类型。

### 主机名通配符

主机名可以是精确匹配（例如“foo.bar.com”）或者使用通配符来匹配 （例如“\*.foo.com”）。 精确匹配要求 HTTP host 头部字段与 host 字段值完全匹配。 通配符匹配则要求 HTTP host 头部字段与通配符规则中的后缀部分相同。

| 主机       | host 头部       | 匹配与否？                          |
| ---------- | --------------- | ----------------------------------- |
| \*.foo.com | bar.foo.com     | 基于相同的后缀匹配                  |
| \*.foo.com | baz.bar.foo.com | 不匹配，通配符仅覆盖了一个 DNS 标签 |
| \*.foo.com | foo.com         | 不匹配，通配符仅覆盖了一个 DNS 标签 |

## 参考资料

1. [Kubernetes / 网络服务 / Ingress](https://kubernetes.io/zh/docs/concepts/services-networking/ingress/)
2. [Kubernetes Ingress 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#ingress-v1-networking-k8s-io)
