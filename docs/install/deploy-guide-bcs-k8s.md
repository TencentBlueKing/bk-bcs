Table of Contents
=================

* [BCS-K8S 安装文档](#bcs-k8s-安装文档)
   * [模块信息](#模块信息)
   * [环境要求](#环境要求)
   * [Chart依赖](#chart依赖)
   * [chart使用](#chart使用)
   * [配置说明](#配置说明)
      * [global参数](#global参数)
      * [bcs-cluster-init参数](#bcs-cluster-init参数)
      * [bcs-gamedeployment-operator参数](#bcs-gamedeployment-operator参数)
      * [bcs-gamestatefulset-operator参数](#bcs-gamestatefulset-operator参数)
      * [bcs-hook-operator参数](#bcs-hook-operator参数)
      * [bcs-k8s-watch参数](#bcs-k8s-watch参数)
      * [bcs-kube-agent参数](#bcs-kube-agent参数)

# BCS-K8S 安装文档

## 模块信息

* bcs-cluster-init
* bcs-gamedeployment-operator
* bcs-gamestatefulset-operator
* bcs-hook-operator
* bcs-k8s-watch
* bcs-kube-agent

## 环境要求

* Kubernetes 1.12+
* Helm 3+

## Chart依赖

* bitnami/common
* bcs-k8s/charts/bcs-cluster-init中需包含以下证书，或已初始化的以下证书secrets
  * ca证书，命名为bcs-ca.crt
  * server证书，命名为bcs-server.crt
  * server私钥，命名为bcs-server.key
  * client证书，命名为bcs-client.crt
  * client私钥，命名为bcs-client.key

## chart使用

```shell

# 非service集群安装
helm upgrade bcs-k8s bcs-k8s -f values.yaml -n bcs-system --install
# service集群安装(需提前安装bcs-services/bcs-init，或已初始化证书secret)
helm upgrade bcs-k8s bcs-k8s -f values.yaml -n bcs-system --install --set bcs-cluster-init.enabled=false


# 卸载
helm uninstall bcs-k8s -n bcs-system
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
| `global.env.BK_BCS_also_log_to_stderr` |  是否开启标准错误输出日志 | 默认为`"true"` |
| `global.env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5 |
| `global.env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型   | `render`，默认渲染 |
| `global.env.BK_BCS_bcsZkHost` | zk服务发现/注册IP列表，多个采用逗号分隔   | 默认为`127.0.0.1:2181` |
| `global.env.BK_BCS_clusterId` | 集群ID | 默认为`BCS-K8S-00000`，必填 |
| `global.secret.bcsCerts` | 全局bcs server/client证书 secret引用名称   | 默认为`bk-bcs-certs` |
| `global.secret.etcdCerts` | 全局etcd证书secret引用名称   | 默认为`bcs-etcd-certs` |

### bcs-cluster-init参数

bcs-cluster-init重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `createNamespace`    | 是否创建命名空间 | 默认为`false` |
| `enabled` | 是否安装 | 默认为`true`，必选，若为service集群则建议置为`false` |

### bcs-gamedeployment-operator参数

bcs-gamedeployment-operator重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-gamedeployment-operator`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |

### bcs-gamestatefulset-operator参数

bcs-gamestatefulset-operator重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-gamestatefulset-operator`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |

### bcs-hook-operator参数

bcs-hook-operator重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-hook-operator`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |

### bcs-k8s-watch参数

bcs-k8s-watch重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-k8s-watch`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_clusterId` | 集群ID | 默认为`空`，必填，如有设置global参数默认被覆盖 |
| `env.BK_BCS_bcsZkHost` | zk host列表 | 默认为`空`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_kubeWatchExternal` | 是否部署在集群外 | 默认为`false`，指部署在集群内 |
| `env.BK_BCS_kubeMaster` | 集群master                           | 默认为`空`，部署在集群外时必填 |
| `env.BK_BCS_customStorage` | 自定义storage | 默认为`空`，管理其他集群时填写 |
| `env.BK_BCS_customNetService` | 自定义netservice | 默认为`空`，管理其他集群时填写 |
| `env.BK_BCS_customNetServiceZK` | 自定义netservice zk | 默认为`空`，管理其他集群时填写 |
| `env.BK_BCS_clientKeyPassword` | client私钥证书密码 | 默认为`空`，如有密码则必填 |
| `env.BK_BCS_also_log_to_stderr` |  是否开启标准错误输出日志 | 默认为`"true"`，如有设置global参数默认被覆盖 |
| `env.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5，如有设置global参数默认被覆盖 |
| `env.BK_BCS_CONFIG_TYPE` | 配置文件生成类型 | `render`，默认渲染，如有设置global参数默认被覆盖 |
| `secret.bcsCerts` | bcs server/client证书 secret引用名称   | 默认为`bk-bcs-certs` |
| `serviceAccount.create` | 是否创建serviceAccount | 默认为`true`，不建议调整 |
| `serviceAccount.name` | serviceAccount名称 | 默认为组件名`bcs-k8s-watch`，不建议调整 |

### bcs-kube-agent参数

bcs-kube-agent重要参数说明

|参数|描述|默认值 |
|---|---|---|
| `replicaCount`    | 容器实例 | 默认为`1`，必选 |
| `image.registry` | 镜像仓库 | 默认为`空`，如有设置global参数默认被覆盖 |
| `image.repository` | 模块镜像路径 | 默认为`bcs/bcs-kube-agent`，必选 |
| `image.tag` | 模块镜像tag | 默认为`空`，必选 |
| `image.pullPolicy` | 拉取镜像策略 | `Always`, `IfNotPresent`，建议设`Always`，如有设置global参数默认被覆盖 |
| `imagePullSecrets` | 镜像仓库secret | 默认为`空`，如有设置将与global参数进行整合 |
| `env.BK_BCS_API` | bcsapi路径 | 默认为`空`，必填 |
| `env.BK_BCS_APIToken` | bcsapi token | 默认为`空`，需要cluster manager 管理权限，必填 |
| `env.BK_BCS_reportPath` | 集群认证信息上报路径 | 默认为`空`，是`env.BK_BCS_API`下的subpath，必填 |
| `env.BK_BCS_clusterId` | 集群ID | 默认为`空`，必填，如有设置global参数默认被覆盖 |
| `env.BK_BCS_kubeAgentWSTunnel` | 是否开启tunnel转发 | 默认为`false`，集群apiserver无法直接被访问时需置为true |
| `env.BK_BCS_websocketPath` | websocket注册路径 | 默认为`/bcsapi/v4/clustermanager/v1/websocket/connect`，不建议调整 |
| `env.BK_BCS_kubeAgentProxy` | ？ | 默认为`空`，可选 |
| `args.BK_BCS_log_level` | 全局日志级别   | `3`，最高为5，如有设置global参数默认被覆盖 |
| `serviceAccount.create` | 是否创建serviceAccount | 默认为`true`，不建议调整 |
| `serviceAccount.name` | serviceAccount名称 | 默认为组件名`bcs-kube-agent`，不建议调整 |