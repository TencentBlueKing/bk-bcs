Table of Contents
=================

* [BCS服务模块安装文档](#bcs服务模块安装文档)
   * [模块信息](#模块信息)
   * [环境要求](#环境要求)
   * [Chart依赖](#chart依赖)
   * [chart使用](#chart使用)
   * [配置说明](#配置说明)
      * [global参数](#global参数)
      * [bcs-init参数](#bcs-init参数)
      * [bcs-api-gateway参数](#bcs-api-gateway参数)
      * [bcs-cluster-manager参数](#bcs-cluster-manager参数)
      * [bcs-storage参数](#bcs-storage参数)
      * [bcs-user-manager参数](#bcs-user-manager参数)

# BCS服务模块安装文档

## 模块信息

* bcs-init
* bcs-api-gateway
* bcs-cluster-manager
* bcs-storage
* bcs-user-manager

## 环境要求

* Kubernetes 1.12+
* Helm 3+

## Chart依赖

* bitnami/common
* bcs-k8s/charts/bcs-gamestatefulset-operator
* bcs-services/charts/bcs-init/cert中需包含以下证书，或已初始化的以下证书secrets
  * ca证书，命名为bcs-ca.crt
  * server证书，命名为bcs-server.crt
  * server私钥，命名为bcs-server.key
  * client证书，命名为bcs-client.crt
  * client私钥，命名为bcs-client.key
  * 未加密的client私钥，命名为bcs-client-unencrypted.key
  * apisix网关 ssl POST DATA，命名为bcs-ssl.json
  * 网关证书，命名可自行定义，并填写在Values中
  * 网关私钥，命名可自行定义，并填写在Values中
* bcs-services/charts/bcs-init/cert中需包含以下证书，或已初始化的以下证书secrets
  * etcd CA证书，命名为etcd-ca.pem
  * etcd 证书，命名为etcd.pem
  * etcd 私钥，命名为etcd-key.pem

## chart使用

```shell
# 安装
# 由于bcs-services依赖bcs-k8s，而bcs-cluster-init与bcs-init的secret信息有重合但非完全一致，安装方式要调整如下
# 1. 安装bcs-init
helm upgrade bcs-init bcs-services/charts/bcs-init -n bcs-system --install
# 2. 屏蔽bcs-cluster-init，安装bcs-k8s
helm upgrade bcs-k8s bcs-k8s -f values.yaml -n bcs-system --install --set bcs-cluster-init.enabled=false
# 3. 屏蔽bcs-init，安装bcs-services
helm upgrade bcs-services bcs-services -f values.yaml -n bcs-system --install --set bcs-init.enabled=false


# 卸载
helm uninstall bcs-services -n bcs-system
```

## 配置说明

helm安装参数详细说明，global参数默认会覆盖各模块中同名参数

### global参数

如果不需要全局设置，请注释

|参数|描述|默认值 |
|---|---|---|
| `global.imageRegistry`    | 全局镜像仓库配置 | 默认为`空`  |
| `global.imagePullSecrets` | 全局拉取镜像secrets设置 | `[]` 已部署的服务请勿中途添加 |
| `global.pullPolicy` | 全局拉取镜像策略   | Always, IfNotPresent，建议设`Always` |
| `global.env.BK_BCS_also_log_to_stderr` | 是否开启标准错误输出日志 | 默认为`true` |
| `global.env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5 |
| `global.env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型   | `render`，默认渲染 |
| `global.env.BK_BCS_bcsEtcdHost` | etcd服务发现IP列表，多个采用逗号分隔   | 默认为`127.0.0.1:2379` |
| `global.env.BK_BCS_bcsZkHost` | zk服务发现/注册IP列表，多个采用逗号分隔   | 默认为`127.0.0.1:2181` |
| `global.env.BK_BCS_queueFlag` | 消息队列开启标志 | 默认为`true` |
| `global.env.BK_BCS_queueKind` | 消息队列类型 | 默认为`rabbitmq`，可选`nats-streaming` |
| `global.env.BK_BCS_queueAddress` | 消息队列地址 | 默认为`空` |
| `global.env.BK_BCS_mongodbAddress` | mongoDB地址 | 默认为`127.0.0.1:27017` |
| `global.env.BK_BCS_mongodbUsername` | mongodb用户名 | 默认为`admin` |
| `global.env.BK_BCS_mongodbPassword` | mongodb密码 | 默认为`空` |
| `global.env.BK_BCS_gatewayToken` | bcs-api-gateway admintoken | 默认为`空` |
| `global.secret.bcsCerts` | 全局bcs server/client证书 secret引用名称   | 默认为`bk-bcs-certs` |
| `global.secret.etcdCerts` | 全局etcd证书secret引用名称   | 默认为`bcs-etcd-certs` |

### bcs-init参数

bcs-init重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `createNamespace`    | 是否创建命名空间 | 默认为`false` |
| `enabled` | 是否安装 | 默认为`false`，不建议调整 |

### bcs-api-gateway参数

bcs-api-gateway重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `apisix.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `apisix.repository` | 模块镜像路径 | 默认为`bcs/apisix`，必选 |
| `apisix.tag` | 模块镜像tag | 默认为`空`，必选 |
| `apisix.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `gateway.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `gateway.repository` | 模块镜像路径 | 默认为`bcs/bcs-gateway-discovery`，必选 |
| `gateway.tag` | 模块镜像tag | 默认为`空`，必选 |
| `gateway.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_adminType` | apisix admin 类型 | 默认为`apisix`，必填 |
| `env.BK_BCS_adminToken` | apisix admin token | 默认为`127.0.0.1:8000`，必填 |
| `env.BK_BCS_adminAPI` | apisix admin API | 默认为`false`，指部署在集群内 |
| `env.BK_BCS_bcsZkHost` | zk服务发现host列表 | 默认为`127.0.0.1:2181`，必填 |
| `env.BK_BCS_zkModules` | 注册在zk的服务名 | 默认为`kubeagent,mesosdriver`，必填 |
| `env.BK_BCS_bcsEtcdHost` | etcd服务发现列表 | 默认为`127.0.0.1:2379`，必填 |
| `env.BK_BCS_etcdGrpcModules` | 注册在etcd的grpc服务名 | 默认为`MeshManager,LogManager`，必填 |
| `env.BK_BCS_etcdHttpModules` | 注册在etcd的http服务名 | 默认为`MeshManager,LogManager,mesosdriver,storage,usermanager`，必填 |
| `env.BK_BCS_apiGatewayCert` | 网关证书文件名 | 默认为`空`，必填 |
| `env.BK_BCS_apiGatewayKey` | 网关私钥文件名 | 默认为`空`，必填 |
| `env.BK_BCS_apiGatewayEtcdHost` | 网关etcdhost | 默认为`http://127.0.0.1:2379`，必填 |
| `env.BK_BCS_gatewayToken` | bcs-api-gateway admintoken | 默认为`空`，必填，如有设置global参数默认被覆盖 |
| `env.BK_BCS_also_log_to_stderr` |  是否开启标准错误输出日志 | 默认为`"true"`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5，如有设置global参数默认被覆盖 |
| `env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型 | `render`，默认渲染，如有设置global参数默认被覆盖 |
| `secret.bcsCerts` | bcs server/client证书 secret引用名称，如有设置global参数默认被覆盖   | 默认为`bk-bcs-certs` |
| `secret.etcdCerts` | etcd证书secret引用名称，如有设置global参数默认被覆盖   | 默认为`bcs-etcd-certs` |

### bcs-cluster-manager参数

bcs-cluster-manager重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-cluster-manager`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_bcsClusterManagerPort` | grpc port | 默认为`8080`，必填 |
| `env.BK_BCS_bcsClusterManagerHTTPPort` | http port | 默认为`8081`，必填 |
| `env.BK_BCS_bcsClusterManagerMetricPort` | metric port | 默认为`8082`，必填 |
| `env.BK_BCS_bcsClusterManagerDebug` | debug模式 | 默认为`false`，可选 |
| `env.BK_BCS_bcsClusterManagerSwaggerDir` | swagger目录 | 默认为`/data/bcs/swagger`，必填 |
| `env.BK_BCS_bcsClusterManagerPeerToken` | tunnel模式token | 默认为`12345678-c714-43d0-8379-d5c2e01e9593`，必填 |
| `env.BK_BCS_bcsEtcdHost` | etcd host | 默认为`空`，必填 |
| `env.BK_BCS_mongodbAddress` | mongoDB地址 | 默认为`127.0.0.1:27017`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_mongodbUsername` | mongoDB用户名 | 默认为`admin`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_mongodbPassword` | mongoDB密码 | 默认为`空`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_bcsClusterManagerMongoConnectTimeout` | mongoDB超时 | 默认为`3`，必填 |
| `env.BK_BCS_bcsClusterManagerMongoDatabase` | clustermanager  数据库名 | 默认为`clustermanager`，必填 |
| `env.BK_BCS_bcsClusterManagerMongoMaxPoolSize` | mongo client最大实例数 | 默认为`0`，0表示不限制，可选 |
| `env.BK_BCS_bcsClusterManagerMongoMinPoolSize` | mongo client最小实例数 | 默认为`0`，0表示不限制，可选 |
| `env.BK_BCS_also_log_to_stderr` | 是否开启标准错误输出日志 | 默认为`true`，必填，如有设置global参数默认被覆盖 |
| `env.BK_BCS_log_level` | 日志级别 | 默认为`3`，最大为`5`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型 | `render`，默认渲染，如有设置global参数默认被覆盖 |
| `secret.bcsCerts` | bcs server/client证书 secret引用名称，如有设置global参数默认被覆盖   | 默认为`bk-bcs-certs` |
| `secret.etcdCerts` | etcd证书secret引用名称，如有设置global参数默认被覆盖   | 默认为`bcs-etcd-certs` |

### bcs-storage参数

bcs-storage重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-storage`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_bcsZkHost` | zk服务发现host列表 | 默认为`127.0.0.1:2181`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_bcsEtcdHost` | etcd服务发现IP列表，多个采用逗号分隔   | 默认为`127.0.0.1:2379`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_mongodbAddress` | mongoDB地址 | 默认为`127.0.0.1:27017`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_mongodbUsername` | mongoDB用户名 | 默认为`admin`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_mongodbPassword` | mongoDB密码 | 默认为`空`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_ConfigDbHost` | configDB地址 | 默认为`127.0.0.1:27017`，如未设置则退化使用mongodbAddress及全局mongodbAddress配置 |
| `env.BK_BCS_ConfigDbUsername` | configDB用户名 |  默认为`admin`，如未设置则退化使用mongodbUsername及全局mongodbUsername配置 |
| `env.BK_BCS_ConfigDbPassword` | mongoDB密码 | 默认为`空`，如未设置则退化使用mongodbPassword及全局mongodbPassword配置 |
| `env.BK_BCS_mongodbOplogCollection` | mongoDB部署形式 | 默认为`oplog.$main`，表示mongoDB单实例部署，`oplog.rs`表示mongoDB集群化部署 |
| `env.BK_BCS_queueFlag` | 消息队列开启标志 | 默认为`true`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_queueKind` | 消息队列类型 | 默认为`rabbitmq`，可选`nats-streaming`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_queueAddress` | 消息队列地址 | 默认为`空`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_resource` | 消息队列推送资源类型 | 默认为`空`，必填，不同资源间使用逗号分隔 |
| `env.BK_BCS_queueClusterId` | clusterID | 默认为`空`，可选，nats需要 |
| `env.BK_BCS_bcsStoragePort` | http port | 默认为`50024`，不建议调整 |
| `env.BK_BCS_bcsStorageMetricPort` | metric port | 默认为`50025`，不建议调整 |
| `env.BK_BCS_eventMaxDay` | event最长存储时间 | 默认为`7`，必填 |
| `env.BK_BCS_eventMaxCap` | event最大存储条数 | 默认为`10000`，必填 |
| `env.BK_BCS_alarmMaxDay` | alarm最长存储时间 | 默认为`7`，必填 |
| `env.BK_BCS_alarmMaxCap` | alarm最大存储条数 | 默认为`10000`，必填 |
| `env.BK_BCS_also_log_to_stderr` |  是否开启标准错误输出日志 | 默认为`"true"`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5，如有设置global参数默认被覆盖 |
| `env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型 | `render`，默认渲染，如有设置global参数默认被覆盖 |
| `secret.bcsCerts` | bcs server/client证书 secret引用名称，如有设置global参数默认被覆盖   | 默认为`bk-bcs-certs` |
| `secret.etcdCerts` | etcd证书secret引用名称，如有设置global参数默认被覆盖   | 默认为`bcs-etcd-certs` |

### bcs-user-manager参数

bcs-user-manager重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-user-manager`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_bcsUserManagerPort` | https port | 默认为`30445`，不建议调整 |
| `env.BK_BCS_bcsUserManagerMetricPort` | metric port   | 默认为`9253`，不建议调整 |
| `env.BK_BCS_bcsUserManagerInsecurePort` | http port | 默认为`8089`，不建议调整 |
| `env.BK_BCS_coreDatabaseDsn` | db dsn | 默认为`空`，必填 |
| `env.BK_BCS_adminUser` | bcs api 用户名 | 默认为`admin`，必填 |
| `env.BK_BCS_adminToken` | bcs api token | 默认为`空`，必填 |
| `env.BK_BCS_bkiamAuthHost` | bkiam认证地址 |  默认为`空`，可选 |
| `env.BK_BCS_tkeSecretId` | tke secret id | 默认为`空`，tke集群必填 |
| `env.BK_BCS_tkeSecretKey` | tke secret key | 默认为`空`，tke集群必填 |
| `env.BK_BCS_tkeCcsHost` | tke ccs host | 默认为`空`，tke集群必填 |
| `env.BK_BCS_tkeCcsPath` | tke ccs 路径 | 默认为`空`，tke集群必填 |
| `env.BK_BCS_bcsZkHost` | zk服务发现host列表 | 默认为`127.0.0.1:2181`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_bcsEtcdHost` | etcd服务发现列表 | 默认为`127.0.0.1:2379`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_also_log_to_stderr` |  是否开启标准错误输出日志 | 默认为`"true"`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5，如有设置global参数默认被覆盖 |
| `env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型 | `render`，默认渲染，如有设置global参数默认被覆盖 |
| `secret.bcsCerts` | bcs server/client证书 secret引用名称，如有设置global参数默认被覆盖   | 默认为`bk-bcs-certs` |
| `secret.etcdCerts` | etcd证书secret引用名称，如有设置global参数默认被覆盖   | 默认为`bcs-etcd-certs` |