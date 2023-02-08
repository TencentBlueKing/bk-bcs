# BCS-Cluster-Resources

> 蓝鲸容器服务（BCS）集群资源层，用于屏蔽底层集群类型，提供统一的 Restful 接口以供 SaaS / OpenAPI 使用。

## 部署

> 注意：该文档主要用于 bcs-cluster-resources 模块单独部署的情况，实际场景中该 Chart 一般作为 bcs-service-stack 的子 Chart 部署，共用相关依赖服务。

### 准备服务依赖

开始部署前，请准备好一套 Kubernetes 集群（版本 1.20 或更高），并安装 Helm 命令行工具（版本 3.5 或更高）。

#### 数据存储

以下为 BCS-Cluster-Resources 必须使用的数据存储服务：

- Etcd：用于同其他 BCS-Service 微服务间进行服务发现，需共用同一实例；
- Redis：用于保存缓存数据等，可与其他 BCS-Service 微服务模块共用实例；

> 注：你可以选择自己搭建，或者直接从云计算厂商处购买这些服务，只要能保证从集群内能正常访问即可。

### 准备 `values.yaml`

BCS-Cluster-Resources 无法直接通过 Chart 所提供的默认 `values.yaml` 完成部署，在执行 `helm install` 安装服务前，你必须按以下步骤准备好匹配当前部署环境的 `values.yaml`。

#### 1. 配置镜像地址

Chart 中已经预设与当前 Chart 版本匹配的容器镜像，您需要将 registry 配置为您所使用的镜像源地址，然后在确认镜像 tag，pullPolicy 是否合适。

```yaml
image:
  registry: "hub.bktencent.com"
  repository: blueking/bcs-cluster-resources
  pullPolicy: IfNotPresent
  tag: v1.28.0-alpha.75
```

> 注：假如服务镜像需凭证才能拉取。请将对应密钥名称写入配置文件中，详细请查看 `values.imagePullSecrets` 配置项说明。

#### 初始化与配置数据存储

准备好服务所依赖的存储后，必须完成以下初始化操作：

**Etcd**

- 使用 etcdctl 命令行工具测试 Etcd 服务可正常使用

**Redis**

- 使用 redis-cli 命令行工具测试 Redis 服务可正常使用

##### 填写数据存储配置

```yaml
svcConf:
  redis:
    address: "127.0.0.1:6379"
    db: 0
    password: "your_redis_password"
  
  ## 注：随 bcs-service-stack 部署则无需调整 values 中的 etcd 配置
  etcd:
    endpoints: "bcs-etcd.bcs-system.svc.cluster.local:2379"
    cert: "/data/bcs/cert/etcd/etcd.pem"
    key: "/data/bcs/cert/etcd/etcd-key.pem"
    ca: "/data/bcs/cert/etcd/etcd-ca.pem"

etcdCertsSecretName: bcs-etcd-certs
```

#### 填写相关服务配置

```yaml
## 启动服务前，检查依赖服务是否可用
## 随 bcs-service-stack 部署时建议开启，会检查并等待依赖的 project-manager, cluster-manager 服务就绪
checkService: true

svcConf:
  ## 全局配置信息
  crGlobal:
    ## 项目基础类配置
    basic:
      ## 蓝鲸应用 ID
      appCode: "bcs_app_code"
      ## 蓝鲸应用 Secret
      appSecret: "bcs_app_secret"
      ## 蓝鲸权限中心访问地址（apigw）
      bkApiGWHost: "http://bkapi.example.com/api/bk-iam/prod"
      ## PaaS 访问主入口地址
      bkPaaSHost: "http://bk.example.com"
      ## 调用 Healthz API 的 Token，一般为长度不小于 32 位的随机字符串
      healthzToken: "your_healthz_token"
      ## 调用清理缓存 API 的 Token，一般为长度不小于 32 位的随机字符串
      cacheToken: "your_cache_token"

    ## 权限中心相关配置
    iam:
      ## 权限中心 API 访问地址
      host: "http://bkiam-api.example.com"
      ## BCS 服务在权限中心注册的系统 ID
      systemID: "bcs_iam_system_id"
```

### 部署 Chart

完成 `values.yaml` 的所有准备工作后，要安装 BCS-Cluster-Resources，你必须先添加一个有效的 Helm repo 仓库。

```shell
## 请将 `<HELM_REPO_URL>` 替换为本 Chart 所在的 Helm 仓库地址
$ helm repo add bkce <HELM_REPO_URL>
```

添加仓库成功后，执行以下命令，在集群内安装名为 `bcs-cluster-resources` 的 Helm release：

> 注：bcs 相关服务一般均部署在 `bcs-system` 命名空间下

```shell
$ helm install bcs-cluster-resources bkce/bcs-cluster-resources -n bcs-system -f values.yaml
```

上述命令将使用指定配置在 Kubernetes 集群中部署 bcs-cluster-resources, 并输出访问指引。

### 卸载 Chart

使用以下命令卸载 `bcs-cluster-resources`:

```shell
$ helm uninstall bcs-cluster-resources -n bcs-system
```

上述命令将移除所有与 bcs-cluster-resources 相关的 Kubernetes 组件，并删除 release。

### 配置说明

以下为可配置的参数列表以及默认值

| 参数                                         | 类型  | 默认值                                        | 描述                                                    |
|---------------------------------------------|------|----------------------------------------------|---------------------------------------------------------|
| enabled                                          | bool | true                                         | 是否启用 bcs-cluster-resources 模块                  |
| image.registry                                   | str  | hub.bktencent.com                            | 镜像源地址                                           |
| image.repository                                 | str  | blueking/bcs-cluster-resources               | 服务镜像地址                                         |
| image.tag                                        | str  | v1.27.0-alpha.9                              | 镜像 tag                                            |
| image.pullPolicy                                 | str  | IfNotPresent                                 | 镜像拉取策略                                         |
| imagePullSecrets                                 | list | []                                           | 镜像拉取密钥                                         |
| fullnameOverride                                 | str  | ""                                           | 覆盖默认名称                                         |
| replicaCount                                     | int  | 1                                            | 默认进程副本数                                       |
| resources.requests.cpu                           | str  | 1                                            | 资源预留（CPU）                                      |
| resources.requests.cpu                           | str  | 512Mi                                        | 资源预留（内存）                                      |
| resources.limits.cpu                             | str  | 2                                            | 资源限制（CPU）                                      |
| resources.limits.memory                          | str  | 4Gi                                          | 资源限制（内存）                                      |
| checkService                                     | bool | false                                        | 是否启用依赖服务检查                                   |
| svcConf.etcd.endpoints                           | str  | bcs-etcd.bcs-system.svc.cluster.local:2379   | Etcd 地址，多个可用英文逗号拼接                         |
| svcConf.etcd.cert                                | str  | /data/bcs/cert/etcd/etcd.pem                 | Etcd 证书路径                                        |
| svcConf.etcd.key                                 | str  | /data/bcs/cert/etcd/etcd-key.pem             | Etcd 密钥路径                                        |
| svcConf.etcd.ca                                  | str  | /data/bcs/cert/etcd/etcd-ca.pem              | Etcd CA 证书路径                                     |
| svcConf.server.cert                              | str  | /data/bcs/cert/bcs/bcs-server.crt            | Server 端证书路径                                    |
| svcConf.server.certPwd                           | str  | ""                                           | Server 端证书密码                                    |
| svcConf.server.key                               | str  | /data/bcs/cert/bcs/bcs-server.key            | Server 端密钥路径                                    |
| svcConf.server.ca                                | str  | /data/bcs/cert/bcs/bcs-ca.crt                | Server 端 CA 证书路径                                |
| svcConf.client.cert                              | str  | /data/bcs/cert/bcs/bcs-client.crt            | Client 端证书路径                                    |
| svcConf.client.certPwd                           | str  | ""                                           | Client 端证书密码                                    |
| svcConf.client.key                               | str  | /data/bcs/cert/bcs/bcs-client.key            | Client 端密钥路径                                    |
| svcConf.client.ca                                | str  | /data/bcs/cert/bcs/bcs-ca.crt                | Client 端 CA 证书路径                                |
| svcConf.swagger.enabled                          | bool | false                                        | 是否启用 swagger 服务（生产环境应禁用）                 |
| svcConf.swagger.dir                              | str  | ""                                           | swagger 挂载目录，测试建议使用 swagger/data            |
| svcConf.log.level                                | str  | info                                         | 日志打印等级，支持 debug/info/warn/error/panic/fatal  |
| svcConf.log.path                                 | str  | /tmp/logs                                    | 日志文件绝对路径                                      |
| svcConf.redis.address                            | str  | 127.0.0.1:6379                               | Redis 访问地址                                       |
| svcConf.redis.db                                 | int  | 0                                            | Redis DB                                            |
| svcConf.redis.password                           | str  | ""                                           | Redis 密码                                           |
| svcConf.crGlobal.auth.disabled                   | bool | false                                        | 关闭用户认证                                          |
| svcConf.crGlobal.auth.jwtPublicKey               | str  | /data/bcs/cert/jwt/public.key                | JWT Public Key 路径                                  |
| svcConf.crGlobal.basic.appCode                   | str  | ""                                           | 蓝鲸应用 ID                                           |
| svcConf.crGlobal.basic.appSecret                 | str  | ""                                           | 蓝鲸应用 Secret                                       |
| svcConf.crGlobal.basic.bkApiGWHost               | str  | ""                                           | 蓝鲸权限中心访问地址（apigw）                           |
| svcConf.crGlobal.basic.bkPaaSHost                | str  | ""                                           | PaaS 访问主入口地址                                    |
| svcConf.crGlobal.basic.healthzToken              | str  | ""                                           | 调用 Healthz API 的 Token                             |
| svcConf.crGlobal.basic.cacheToken                | str  | ""                                           | 调用清理缓存 API 的 Token                              |
| svcConf.crGlobal.bcsApiGW.host                   | str  | https://bcs-api-gateway                 | BCS 网关访问地址                                      |
| svcConf.crGlobal.bcsApiGW.authToken              | str  | ""                                           | BCS 网关 Auth Token                                  |
| svcConf.crGlobal.bcsApiGW.readAuthTokenFromEnv   | bool | true                                         | 开启后将读取 Secret 挂载到容器中 Env 的值                |
| svcConf.crGlobal.iam.host                        | str  | ""                                           | 权限中心 API 访问地址                                   |
| svcConf.crGlobal.iam.systemID                    | str  | ""                                           | BCS 服务在权限中心注册的系统 ID                          |
| svcConf.crGlobal.iam.useBKApiGW                  | bool | true                                         | 是否使用 bkApiGWHost 调用权限中心 API                   |
| svcConf.crGlobal.iam.metric                      | bool | false                                        | 是否启用权限中心 SDK Metric 功能                        |
| svcConf.crGlobal.iam.debug                       | bool | false                                        | 是否启用权限中心 SDK Debug 功能                         |

#### 如何修改配置项

在安装 Chart 时，你可以通过 `--set key=value[,key=value]` 的方式，在命令参数里修改各配置项。例如:

```shell
$ helm install bcs-cluster-resources bkce/bcs-cluster-resources -n bcs-system --set checkService=false
```

此外，你也可以把所有配置项写在 YAML 文件（常被称为 Helm values 文件）里，通过 `-f` 指定该文件来使用特定配置项：

```shell
$ helm install bcs-cluster-resources bkce/bcs-cluster-resources -f values.yaml
```

执行 `helm show values`，你可以查看 Chart 的所有默认配置：

```shell
## 查看默认配置
$ helm get values bcs-cluster-resources -n bcs-system

## 保存默认配置到文件 values.yaml
$ helm get values bcs-cluster-resources > values.yaml
```
