# BCS-Platform-Manager

> 蓝鲸容器服务（BCS）集群平台管理，用于运维，提供统一的 Restful 接口以供平台管理使用。

## 部署

> 注意：该文档主要用于 bcs-platform-manager 模块单独部署的情况，实际场景中该 Chart 一般作为 bcs-service-stack 的子 Chart 部署，共用相关依赖服务。

### 准备服务依赖

开始部署前，请准备好一套 Kubernetes 集群（版本 1.20 或更高），并安装 Helm 命令行工具（版本 3.5 或更高）。

#### 数据存储

以下为 BCS-Platform-Manager 必须使用的数据存储服务：

- Etcd：用于同其他 BCS-Service 微服务间进行服务发现，需共用同一实例；
- Redis：用于保存缓存数据等，可与其他 BCS-Service 微服务模块共用实例；

> 注：你可以选择自己搭建，或者直接从云计算厂商处购买这些服务，只要能保证从集群内能正常访问即可。

### 准备 `values.yaml`

BCS-Platform-Manager 无法直接通过 Chart 所提供的默认 `values.yaml` 完成部署，在执行 `helm install` 安装服务前，你必须按以下步骤准备好匹配当前部署环境的 `values.yaml`。

#### 1. 配置镜像地址

Chart 中已经预设与当前 Chart 版本匹配的容器镜像，您需要将 registry 配置为您所使用的镜像源地址，然后在确认镜像 tag，pullPolicy 是否合适。

```yaml
image:
  registry: "hub.bktencent.com"
  repository: blueking/bcs-platform-manager
  pullPolicy: IfNotPresent
  tag: v1.29.0
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
    ## 为空则从环境变量获取
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

## 服务配置信息
svcConf:
  web:
    ## 服务web开放前缀
    route_prefix: "/bcsapi/v4/platformmanager/v1"
  
  ## 日志配置
  logging:
    log_dir: "./tmp/logs"
    log_max_size: 500
    log_max_num: 3
    logtostderr: true

  ## 基础相关配置
  base_conf:
    http_port: 8099
    bind_address: ""
    ## 为空则从环境变量获取
    app_code: ""
    ## 为空则从环境变量获取
    app_secret: ""
    ## 为空则从环境变量获取
    system_id: ""
    visitors_white_list: ["admin"]

  ## 权限中心相关配置
  iam_conf:
    gateway_server: ""    
    ## 项目基础类配置
  
  ## BCS 相关配置
  bcs_conf:
    host: "http://bcs-api-gateway.bcs-system.svc.cluster.local"
    token: ""
    ## 为空则从环境变量获取
    jwt_public_key: ""
    readAuthTokenFromEnv: true

  ## Etcd 相关配置
  etcd:
    endpoints: "bcs-etcd:2379"
    cert: "/data/bcs/cert/etcd/etcd.pem"
    key: "/data/bcs/cert/etcd/etcd-key.pem"
    ca: "/data/bcs/cert/etcd/etcd-ca.pem"

  ## Redis 配置信息
  redis:
    address: "bcs-redis-master:6379"
    db: 2
    password: ""
    ## 以下项非必须可不启用
    # dialTimeout: 2
    # readTimeout: 1
    # writeTimeout: 1
    # poolSize: 64
    # minIdleConns: 64

  ## Mongo 配置信息
  mongo:
    ## 为空则从环境变量获取
    address: ""
    authdatabase: admin
    database: bcsplatformmanager
    username: "root"
    password: ""
    connecttimeout: 5
    maxpoolsize: 0
    minpoolsize: 0

  ## 链路追踪配置
  tracing:
    enabled: false
    endpoint: ""
    token: ""
```

### 部署 Chart

完成 `values.yaml` 的所有准备工作后，要安装 BCS-Platform-Manager，你必须先添加一个有效的 Helm repo 仓库。

```shell
## 请将 `<HELM_REPO_URL>` 替换为本 Chart 所在的 Helm 仓库地址
$ helm repo add bkce <HELM_REPO_URL>
```

添加仓库成功后，执行以下命令，在集群内安装名为 ` bcs-platform-manager` 的 Helm release：

> 注：bcs 相关服务一般均部署在 `bcs-system` 命名空间下

```shell
$ helm install bcs-platform-manager bkcebcs-platform-manager -n bcs-system -f values.yaml
```

上述命令将使用指定配置在 Kubernetes 集群中部署 bcs-platform-manager, 并输出访问指引。

### 卸载 Chart

使用以下命令卸载 `bcs-platform-manager`:

```shell
$ helm uninstall bcs-platform-manager -n bcs-system
```

上述命令将移除所有与 bcs-platform-manager 相关的 Kubernetes 组件，并删除 release。

### 配置说明

以下为可配置的参数列表以及默认值

| 参数                                         | 类型  | 默认值                                        | 描述                                                    |
|---------------------------------------------|------|----------------------------------------------|---------------------------------------------------------|
| enabled                                          | bool | true                                         | 是否启用 bcs-platform-manager 模块                  |
| image.registry                                   | str  | hub.bktencent.com                            | 镜像源地址                                           |
| image.repository                                 | str  | blueking/bcs-platform-manager               | 服务镜像地址                                         |
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
| svcConf.redis.password                           | str  | ""                                           | Redis 密码                                           |                                          |
| svcConf.bcs_conf.jwtPublicKey               | str  | /data/bcs/cert/jwt/public.key                | JWT Public Key 路径                                  |
| svcConf.global.bkAPP.appCode                   | str  | ""                                           | 蓝鲸应用 ID                                           |
| svcConf.global.bkAPP.appSecret                 | str  | ""                                           | 蓝鲸应用 Secret                                       |
| svcConf.global.bkIAM.gateWayHost               | str  | ""                                           | 蓝鲸权限中心访问地址（apigw）                           |
| svcConf.bcs_conf.host                   | str  | https://bcs-api-gateway                 | BCS 网关访问地址                                      |
| svcConf.bcs_conf.token              | str  | ""                                           | BCS 网关 Auth Token                                  |
| svcConf.bcs_conf.readAuthTokenFromEnv   | bool | true                                         | 开启后将读取 Secret 挂载到容器中 Env 的值                |
| svcConf.base_conf.http_port                        | str  | ""                                           | 服务 HTTP 端口                                       |
| svcConf.base_conf.bind_address                    | str  | ""                                           | 服务绑定的 IP 地址                                |

#### 如何修改配置项

在安装 Chart 时，你可以通过 `--set key=value[,key=value]` 的方式，在命令参数里修改各配置项。例如:

```shell
$ helm install bcs-platform-manager bkce/bcs-platform-manager -n bcs-system --set checkService=false
```

此外，你也可以把所有配置项写在 YAML 文件（常被称为 Helm values 文件）里，通过 `-f` 指定该文件来使用特定配置项：

```shell
$ helm install bcs-platform-manager bkce/bcs-platform-manager -f values.yaml
```

执行 `helm show values`，你可以查看 Chart 的所有默认配置：

```shell
## 查看默认配置
$ helm get values bcs-platform-manager -n bcs-system

## 保存默认配置到文件 values.yaml
$ helm get values bcs-platform-manager > values.yaml
```
