# Secret

> Secret 是一种包含少量敏感信息例如密码、令牌或密钥的对象，类似于 ConfigMap 但专门用于保存机密数据。
> Secret 信息可能会被放在 Pod 规约中或者镜像中，因而不需要在应用程序代码中包含机密数据。

## 创建 Secret

有几种不同的方式来创建 Secret：

- [使用 kubectl 命令创建 Secret](https://kubernetes.io/zh/docs/tasks/configmap-secret/managing-secret-using-kubectl/)
- [使用配置文件来创建 Secret](https://kubernetes.io/zh/docs/tasks/configmap-secret/managing-secret-using-config-file/)
- [使用 kustomize 来创建 Secret](https://kubernetes.io/zh/docs/tasks/configmap-secret/managing-secret-using-kustomize/)

### Secret 类型

创建 Secret 时，你可以使用 Secret 资源的 `type` 字段， 或者与其等价的 `kubectl` 命令行参数（如果有的话）为其设置类型。 Secret 的 `type` 有助于对不同类型机密数据的编程处理。

Kubernetes 提供若干种内置的类型，用于一些常见的使用场景。 针对这些类型，Kubernetes 所执行的合法性检查操作以及对其所实施的限制各不相同。

| 内置类型                            | 用法                                   |
| ----------------------------------- | -------------------------------------- |
| Opaque                              | 用户定义的任意数据                     |
| kubernetes.io/service-account-token | 服务账号令牌                           |
| kubernetes.io/dockercfg             | ~/.dockercfg 文件的序列化形式          |
| kubernetes.io/dockerconfigjson      | ~/.docker/config.json 文件的序列化形式 |
| kubernetes.io/basic-auth            | 用于基本身份认证的凭据                 |
| kubernetes.io/ssh-auth              | 用于 SSH 身份认证的凭据                |
| kubernetes.io/tls                   | 用于 TLS 客户端或者服务器端的数据      |
| bootstrap.kubernetes.io/token       | 启动引导令牌数据                       |

通过为 Secret 对象的 `type` 字段设置一个非空的字符串值，你也可以定义并使用自己 Secret 类型。如果 `type` 值为空字符串，则被视为 `Opaque` 类型。 Kubernetes 并不对类型的名称作任何限制。不过，如果你要使用内置类型之一， 则你必须满足为该类型所定义的所有要求。

## 使用 Secret

要使用 Secret，Pod 需要引用 Secret。Pod 可以用三种方式之一来使用 Secret：

- 作为挂载到一个或多个容器上的[卷中的文件](https://kubernetes.io/zh/docs/concepts/configuration/secret/#using-secrets-as-files-from-a-pod)
- 作为[容器的环境变量](https://kubernetes.io/zh/docs/concepts/configuration/secret/#using-secrets-as-environment-variables)
- 由 kubelet 在为 [Pod 拉取镜像时使用（imagePullSecret）](https://kubernetes.io/zh/docs/concepts/configuration/secret/#using-imagepullsecrets)

Kubernetes 控制平面也使用 Secret；例如，引导令牌 Secret 是一种帮助自动化节点注册的机制。

Secret 对象的名称必须是合法的 DNS 子域名。在为创建 Secret 编写配置文件时，你可以设置 data 与/或 stringData 字段。data 和 stringData 字段都是可选的。data 字段中所有键值都必须是 base64 编码的字符串。如果不希望执行这种 base64 字符串的转换操作，你可以选择设置 stringData 字段，其中可以使用任何字符串作为其取值。

## 参考资料

1. [Kubernetes / 配置 / Secret](https://kubernetes.io/zh/docs/concepts/configuration/secret/)
2. [Kubernetes Secret 字段说明](https://kubernetes.io/docs/reference/generated/kubernetes-api/v1.21/#secret-v1-core)
