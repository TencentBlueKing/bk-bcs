Table of Contents
=================

* [BCS服务高可用安装](#bcs服务高可用安装)
   * [依赖](#依赖)
   * [概要说明](#概要说明)
   * [安装部署](#安装部署)
      * [Requirements:](#requirements)
         * [准备证书文件](#准备证书文件)
      * [【推荐】容器化部署](#推荐容器化部署)
         * [创建 Secret](#创建-secret)
         * [安装业务集群（bcs-k8s）模块](#安装业务集群bcs-k8s模块)
         * [安装Service集群（bcs-services）模块](#安装service集群bcs-services模块)
      * [【不推荐】本地部署](#不推荐本地部署)
         * [准备配置文件](#准备配置文件)
         * [启动服务](#启动服务)
      * [服务分布表{#layers}](#服务分布表layers)
         * [后台层](#后台层)
         * [Mesos 集群层 - Master 节点](#mesos-集群层---master-节点)
      * [Mesos 集群层 - Node节点](#mesos-集群层---node节点)
         * [K8S 集群](#k8s-集群)
      * [部署验证](#部署验证)
   * [接入蓝鲸社区版5.1+](#接入蓝鲸社区版51)
      * [1. 构建社区版规范化的目录结构](#1-构建社区版规范化的目录结构)
      * [2. 替换社区版中的bcs 后台部分](#2-替换社区版中的bcs-后台部分)


# BCS服务高可用安装

>文档持续完善中~

该文档说明如何针对BCS Service层进行部署，其他请参照：

* [单机Mesos集群部署](./mesos-deploy-in-single-guide.md)
* [Kubernetes高可用部署](./Deploy_BCS_in_K8S_HA_Cluster.md)
* [Mesos集群高可用部署](./Deploy_BCS_in_Mesos_HA_Cluster.md)

## 依赖

- etcd 3.12
- docker
- zookeeper
- MongoDB 4
- rabbitmq
- Harbor

## 概要说明

容器管理平台(BCS) 后台包含以下组件

- bcs-api (进程化部署)
- bcs-gateway-discovery (容器化部署)
- bcs-dns
- bcs-user-manager
- bcs-cluster-manager
- bcs-storage
- bcs-ops
- bcs-mesos-driver
- bcs-mesos-watch
- bcs-container-executor
- bcs-check
- bcs-scheduler
- bcs-client

BCS Service 是业务集群的控制面，对业务集群提供信息聚合、管控和访问能力。BCS Service 层的服务通过以下组件提供：

- bcs-api (进程化部署)
- bcs-gateway-discovery (容器化部署)
- bcs-cluster-manager
- bcs-user-manager
- bcs-storage

## 安装部署

> Note:
> 
> BCS Service 层全部服务已支持容器化部署，建议您使用容器化部署方式以降低部署和管理难度。

### Requirements:

- 操作系统： CentOS 7+

- 部署依赖服务：

  - zookeeper，MongoDB，Docker,  Harbor，etcd 等略

  > Note:
  >
  > 将已安装好的上述服务的相关信息准备好，以备用。如 IP,  域名，端口，账号密码等信息

#### 准备证书文件

- 为方便，这里使用cfssl，cfssljson两个小工具来生成证书。需要实现准备证书生成配置文件

  - ca-csr.json， 文件样例

    ```json
    {
        "CN": "bcs",
        "key": {
            "algo": "rsa",
            "size": 2048
        },
        "names": [
            {
                "C": "CN",
                "L": "SZ",
                "O": "TX",
                "ST": "GD",
                "OU": "CA"
            }
        ]
    }
    ```

    

  - ca-config.json 文件样例：

    ```json
    {
        "signing": {
            "default": {
                "expiry": "43800h"
            },
            "profiles": {
                "server": {
                    "expiry": "43800h",
                    "usages": [
                        "signing",
                        "key encipherment",
                        "server auth"
                    ]
                },
                "client": {
                    "expiry": "43800h",
                    "usages": [
                        "signing",
                        "key encipherment",
                        "client auth"
                    ]
                },
                "peers": {
                    "expiry": "43800h",
                    "usages": [
                        "signing",
                        "key encipherment",
                        "server auth",
                        "client auth"
                    ]
                }
            }
        }
    }
    ```

- 生成ca证书

  - 创建证书目录
  - 生成ca根证书

`cfssl gencert --initca ca-csr.json | cfssljson -bare bcs-ca`

生成 bcs-ca.pem，bcs-ca.key 两个文件。etcd-ca.key文件可以保管到秘密位置，注意不要泄露，etcd-ca.pem 文件后续步骤备用

- 生成client及server证书

```bash
# bcs-client 证书
cfssl gencert -ca=bcs-ca.pem \
  -ca-key=bcs-ca-key.pem \
  -config=ca-config.json \
  -profile=client \
  ca-csr.json | cfssljson -bare bcs-client
  
# bcs-server 证书
cfssl gencert -ca=bcs-ca.pem \
  -ca-key=bcs-ca-key.pem \
  -config=ca-config.json \
  -profile=server \
  ca-csr.json | cfssljson -bare bcs-server
```

> Note：
>
> 按照上述方式，生成 etcd 服务所需要的证书

- 构造 apisix ssl POST DATA

需要提前申请域名证书，将域名证书填入以下格式的json串，命名为bcs-ssl.json，格式如下

```json
{
"cert": "${域名证书}",
"key": "${域名证书私钥}",
"sni": "${域名格式，例如*.bk.tencent.com}"
}
```


### 【推荐】容器化部署

> 容器化部署特殊环境要求:
> - Kubernetes 1.12+
> - Helm 3+
> - Service 层服务镜像
>   - bcs-gateway-discovery
>   - bcs-storage
>   - bcs-user-manager
>   - bcs-cluster-manager
> - 业务集群服务镜像
>   - bcs-kube-agent
>   - bcs-k8s-watch
>   - bcs-gamestatefulset-operator
>   - bcs-gamedeployment-operator
>   - bcs-hook-operator

> Note: 
> BCS Service 集群是特殊的业务集群，其安装包括业务集群模块安装及 Service 模块安装。Service 集群的安装将分为以下几个步骤

> Node: 以下 [bcs-services](../../install/helm/bcs-services) 及 [bcs-k8s](../../install/helm/bcs-k8s) 均指 helm chart


#### 创建 Secret

> 由于 service 集群所需的证书与业务集群不完全一致，需要首先创建独立 secret。

将证书导入 bcs-services/charts/bcs-init/cert 目录（若不存在请手动创建），包括：
  - etcd CA证书，命名为 `etcd-ca.pem`
  - etcd 证书，命名为 `etcd.pem`
  - etcd 私钥，命名为 `etcd-key.pem`
  - ca证书，命名为 `bcs-ca.crt`
  - server证书，命名为 `bcs-server.crt`
  - server私钥，命名为 `bcs-server.key`
  - client证书，命名为 `bcs-client.crt`
  - client私钥，命名为 `bcs-client.key`
  - 未加密的client私钥，命名为 `bcs-client-unencrypted.key`
  - apisix网关 ssl POST DATA，命名为 `bcs-ssl.json`
  - api 网关证书，命名可自行定义，并填写在 bcs-services 的 bcs-api-gateway.env.BK_BCS_apiGatewayCert 中
  - api 网关私钥，命名可自行定义，并填写在 bcs-services 的 bcs-api-gateway.env.BK_BCS_apiGatewayKey 中

执行以下命令，安装 secret
> 命名空间建议选择bcs-system，如果有特殊需要，可以安装在其他命名空间
> 
> 由于 bcs-k8s/bcs-webhook-server 通过 k8s service 域名访问，需要生成以下域名的证书及私钥，并以base64编码填写在 bcs-k8s 的 `bcs-webhook-server.serverCert`, `bcs-webhook-server.serverKey`, `bcs-webhook-server.caBundle` 中
>
> ${service}
>
> ${service}.${namespace}
>
> ${service}.${namespace}.svc

```shell
helm update bcs-init ./bcs-services/charts/bcs-init -n bcs-system --install
```

#### 安装业务集群（bcs-k8s）模块

按照 bcs-k8s [部署文档](./deploy-guide-bcs-k8s.md#配置说明) 填写必要的部署参数，并执行以下命令
```json
helm upgrade bcs-k8s ./bcs-k8s \ 
-f your_values.yaml \ 
-n bcs-system \ 
--set bcs-cluster-init.enabled=false \ 
--install
```

#### 安装Service集群（bcs-services）模块

按照 bcs-services [部署文档](./deploy-guide-bcs-services.md#配置说明) 填写必要的部署参数，并执行以下命令
```json
helm upgrade bcs-services ./bcs-services \ 
-f your_values.yaml \ 
-n bcs-system \ 
--set bcs-init.enabled=false \ 
--install
```

### 【不推荐】本地部署

#### 准备配置文件

> Note: 
> 文档默认来源请参照编译文档。
> Build 出来的产物中的配置文件模板与下述文件名不相同，默认为：config_file.template，为便于识别，以下配置文件名做了相应的调整。

- bcs-api.json

  ```JSON
  {
      "edition": "ee",  # 标注对外,不可更改
      "address": "__LAN_IP__",   # 填写部署bcs-api的主机IP
      "port": __BCS_API_HTTPS_PORT__, # 定义一个bcs-api使用的https端口，一般为443
      "log_dir": "__INSTALL_PATH__/logs/bcs/", # 指定日志存放路径
      "pid_dir": "/var/run/bcs/",
      "insecure_address": "__LAN_IP__",  # 填写部署bcs-api的主机IP
      "insecure_port": __BCS_API_HTTP_PORT__,  # http 端口，一般为80
      "metric_port": __BCS_API_METRIC_PORT__,  # 指标数据端口
      "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__", # 逗号分隔的zk地址，如127.0.0.1:2181,127.0.0.1:2181
      "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem", # 证书 ca 文件路径，按需修改
      "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem", # server证书
      "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",# server 密钥
      "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem", # client 证书
      "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem", # client 密钥
      "local_ip": "__LAN_IP__", # 本机IP，通常用内网IP
      "bkiam_auth": {
        "auth": false, # 是否启用权限中心(蓝鲸一级SaaS)，当前版本不需要修改，设置为false
        "bkiam_auth_host": "http://__IAM_HOST__", # 蓝鲸权限中心域名
        "bkiam_auth_app_code": "__APP_CODE__", # 权限中心app_code
        "bkiam_auth_app_secret": "__APP_TOKEN__" # 权限中心app_token
      },
      "bke": {
        "mysql_dsn":  "__MYSQL_BCS_USER__:__MYSQL_BCS_PASS__@tcp(__MYSQL_BCS_IP0__:__MYSQL_BCS_PORT__)/bke_core?charset=utf8mb4&parseTime=True&loc=Local", # MYSQL连接信息。
        "bootstrap_users": [  # 调用k8s相关资源时需要使用的账号及凭证信息， 开源版本未使用到
          {
            "name": "__BKE_ADMIN_USER__", 
            "is_super_user": true,
            "tokens": [
              "__BKE_ADMIN_ENCRYPT_TOKEN__"
            ]
          }
        ],
        "turn_on_rbac": false,  # 对接权限中心的开关，默认关闭
        "turn_on_auth": false,	# 同上
        "turn_on_conf": false		# 同上
      }
  }
  ```

- bcs-check.json 

  ```json
  {
    "address": "__LAN_IP__",
    "port": __BCS_CHECK_PORT__,
    "metric_port": __BCS_CHECK_METRIC_PORT__,
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem", # 证书 ca 文件路径，按需修改
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem", # server证书
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",# server 密钥
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem", # client 证书
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem", # client 密钥
    "mesos_zookeeper": "__COMMA_SEP_LIST_ZK_MESOS_SERVER__",
    "cluster": "__MESOS_CLUSTER_ID__" # 创建的mesos业务集群ID, 采用三段式,如：BCS-MESOS-10001，最后一截为数字
  }
  ```

- bcs-storage.json

  ```json
  {
    "address": "__LAN_IP__",
    "port": __BCS_STORAGE_PORT__,
    "log_dir": "__INSTALL_PATH__/logs/bcs/",
    "pid_dir": "/var/run/bcs/",
    "metric_port": __BCS_STORAGE_METRIC_PORT__,
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "database_config_file": "__INSTALL_PATH__/etc/bcs/storage-database.conf",
    "event_max_day": __BCS_EVENT_MAX_DAY__, # 事件数据保留天数
    "event_max_cap": __BCS_EVENT_MAX_CAP__, # 事件数据保留天数(每个集群)
    "alarm_max_day": __BCS_ALARM_MAX_DAY__, # 告警数据保留天数
    "alarm_max_cap": __BCS_ALARM_MAX_CAP__, # 告警数据保留条数（每个集群）
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem",
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem",
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem"
  }
  ```

- bcs-health-master.json

  ```json
  {
    "address": "__LAN_IP__",
    "port": __BCS_HEALTH_MASTER_PORT__,
    "log_dir": "__INSTALL_PATH__/logs/bcs/",
    "pid_dir": "/var/run/bcs/",
    "metric_port": __BCS_HEALTH_MASTER_METRIC_PORT__,
    "local_ip": "__LAN_IP__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem",
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem",
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem",
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem",
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "enable_storage_alarm": true,
    "etcd": {
      "etcd_endpoints": "https://__ETCD_IP__:__ETCD_CLIENT_PORT__", # etcd 客户端信息
      "etcd_root_path": "/bcshealtch",
      "etcd_ca_file": "__INSTALL_PATH__/cert/etcd/etcd-ca.pem",
      "etcd_cert_file": "__INSTALL_PATH__/cert/etcd/etcd.pem",
      "etcd_key_file": "__INSTALL_PATH__/cert/etcd/etcd-key.pem",
      "etcd_key_password": "__ETCD_KEY_PASS__"
    }
  }
  ```

- bcs-health-slave.json

  ```json
  {
    "cluster_name": "health-slave-default",
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "log_dir": "__INSTALL_PATH__/logs/bcs/",
    "pid_dir": "/var/run/bcs",
    "local_ip": "__LAN_IP__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem",
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem",
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem",
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem",
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",
    "metric_port": __BCS_HEALTH_SLAVE_METRIC_PORT__,
    "ls_address": "",
    "ls_ca_file": "",
    "ls_client_cert_file": "",
    "ls_client_key_file": "",
    "zones": []
  }
  ```

- bcs-scheduler.json

  ```json
  {
    "address": "__LAN_IP__",
    "port": __BCS_SCHEDULER_PORT__,
    "metric_port": __BCS_SCHEDULER_METRIC_PORT__,
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem", # 证书 ca 文件路径，按需修改
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem", # server证书
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",# server 密钥
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem", # client 证书
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem", # client 密钥
    "use_cache": false, # 默认false
    "regdiscv": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",  # 用于服务发现的ZK， 格式：ip1:port，多个用逗号分隔
    "mesos_regdiscv": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",  # 用于mesos服务发现的ZK，格式同上
    "zkhost": "__COMMA_SEP_LIST_ZK_BCS_SERVER__", # 用于存储配置数据的ZK
    "plugins": "", # 一般不需要制定，使用underley ip管理时。制定为ip-resource, 会在调度时把ip资源纳入考虑范围
    "cluster": "__MESOS_CLUSTER_ID__"    # 创建的mesos业务集群ID, 采用三段式,如：BCS-MESOS-10001，最后一截为数字
  }
  ```

- bcs-dns.conf

  ```ini
  .:53 {
      log . "{remote} - {type} {class} {name} {proto} {size} {rcode} {rsize}" {
          class all
      }
      loadbalance round_robin
      cache 5
      bcsscheduler bcs.com. {
          cluster __MESOS_CLUSTER_ID_SUFFIX__   # MESOS_CLUSTER_ID 取值最后一截的数字
          resyncperiod 30
          endpoints __SPACE_SEP_LIST_ZK_BCS_SERVER__  # 空格分隔的ZK信息。(ip:port)
          endpoints-path /blueking
          fallthrough
  
          upstream __SERVICE_DNS_UPSTREAM__
          registery __SPACE_SEP_LIST_ZK_BCS_SERVER__  # 空格分隔的ZK信息。(ip:port)
          storage __SPACE_SEP_LIST_ETCD_SERVER__
          storage-tls cert/etcd/etcd.pem cert/etcd/etcd-key.pem cert/etcd/ca.pem
          storage-path /bluekingdns
      }
      proxy bcscustom.com. __SERVICE_DNS_UPSTREAM__ {
          policy round_robin
          fail_timeout 5s
          max_fails 0
          spray
      }
      proxy . __DNS_UPSTREAM__ {
          policy round_robin
          fail_timeout 5s
          max_fails 0
          spray
      }
  }
  ```

- bcs-mesos-driver.json

  ```json
  {
    "address": "__LAN_IP__",
    "port": __BCS_MESOS_DRIVER_PORT__,
    "metric_port": __BCS_MESOS_DRIVER_METRIC_PORT__,
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem", # 证书 ca 文件路径，按需修改
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem", # server证书
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",# server 密钥
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem", # client 证书
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem", # client 密钥
    "sched_regdiscv": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "cluster": "__BCS_MESOS_CLUSTER_ID__"
  }
  ```

- bcs-mesos-watch.json

  ```json
  {
    "address": "${localIp}",
    "port": ${bcsMesosWatchPort},
    "metric_port": ${bcsMesosWatchMetricPort},
    "bcs_zookeeper": "__COMMA_SEP_LIST_ZK_BCS_SERVER__",
    "ca_file": "__INSTALL_PATH__/cert/bcs/bcs-ca.pem", # 证书 ca 文件路径，按需修改
    "server_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-server.pem", # server证书
    "server_key_file": "__INSTALL_PATH__/cert/bcs/bcs-server-key.pem",# server 密钥
    "client_cert_file": "__INSTALL_PATH__/cert/bcs/bcs-client.pem", # client 证书
    "client_key_file": "__INSTALL_PATH__/cert/bcs/bcs-client-key.pem", # client 密钥
    "clusterinfo": "__COMMA_SEP_LIST_ZK_BCS_SERVER__/blueking",
    "cluster": "__BCS_MESOS_CLUSTER_ID__"
  }
  ```


#### 启动服务

bcs所有服务启动使用统一的方式：  `<程序> -f <配置文件>`

如： `./bcs-api -f bcs-api.json`

将上述配置文件中的变量替换成对应的服务后，启动进程即可。



注意事项：

1. 参考以下[服务分布表](#layers)表来进行服务的启动，后台层与集群层机器`建议`分别部署
2. master节点主机数为奇数

### 服务分布表{#layers}

#### 后台层

| 工程              | 社区版     | 开源版         |
| ----------------- | ---------- | -------------- |
| MongDB            | √    3.4.9 | √    建议3.4+  |
| etcd              | √    3.1.8 | √    建议 3.1+ |
| zookeeper         | √    3.4.6 | √    建议3.4+  |
| bcs-dns           | √          | √              |
| bcs-api           | √          | √              |
| bcs-ops           | √          | √              |
| bcs-storage       | √          | √              |
| bcs-health-master | √          | √              |
| bcs-health-slave  | √          | √              |

#### Mesos 集群层 - Master 节点

| 工程             | 社区版 | 开源版 |
| ---------------- | ------ | ------ |
| zookeeper        | √      | √      |
| etcd             | √      | √      |
| mesos-master     | √      | √      |
| bcs-dns          | √      | √      |
| bcs-scheduler    | √      | √      |
| bcs-mesos-driver | √      | √      |
| bcs-mesos-watch  | √      | √      |
| bcs-health-slave | √      | √      |
| bcs-check        | √      | √      |

### Mesos 集群层 - Node节点

| 工程                   | 社区版  | 开源版       |
| ---------------------- | ------- | ------------ |
| flannel                | 0.10.0  | 建议 0.10.0+ |
| docker                 | ce 18.0 | 建议 18.0+   |
| mesos-slave            | √       | √            |
| bcs-container-executor | √       | √            |

Mesos集群部署请参照[Mesos集群高可用部署](./Deploy_BCS_in_Mesos_HA_Cluster.md)

#### K8S 集群

K8S集群请参照[Kubernetes高可用部署](./Deploy_BCS_in_K8S_HA_Cluster.md)

### 部署验证

参考 docs/features/bcs-client/bcs-client_HANDBOOK.md，使用bcs-client进行集群创建操作

## 接入蓝鲸社区版5.1+

首先确认要接入的目标社区版为5.1以上，带有bcs内容的版本: `ls -l /data/src/bcs`

### 1. 构建社区版规范化的目录结构

​	打包二进制

> 1. 接入社区版中，并不是所有的二进制都需要，这里仅列出社区版需要的工程文件
> 2. 因为社区版已经带有各工程的配置文件，因此，不需要再自行准备配置文件。

```text
 bcs/server
  ├── bin
  │   ├── bcs-api
  │   ├── bcs-dns
  │   ├── bcs-health-master
  │   ├── bcs-health-slave
  │   ├── bcs-ops
  │   ├── bcs-storage
  └── VERSION
```

将版本号(可以自定义)写入VERSION 文件， 如： github-1.1.0

可以用以下命令快速处理(在build目录下执行)

```bash
mkdir bin
cd build/bcs.295eb49-19.06.05/
tar zcf bcs-server-github-1.1.0.tgz bin/bcs-{api,dns,ops,storage,health-slave,health-master}
```

 

### 2. 替换社区版中的bcs 后台部分

​	登陆社区版中控机，执行备份，替换，安装操作。命令序列如下：

```bash
cd /data/src
rsync -a /data/src/bcs/server /data/backup/bcs/  # 备份

tar xvf bcs-server-github-1.1.0.tgz -C /data/src/bcs/server/  # 替换掉社区版中对应的二进制

## 执行安装
./bkcec sync bcs
echo api ops dns storage health-master health-slave | xargs -n1 ./bkcec stop bcs
echo api ops dns storage health-master health-slave | xargs -n1 ./bkcec install bcs
echo api ops dns storage health-master health-slave | xargs -n1 ./bkcec start bcs
```

