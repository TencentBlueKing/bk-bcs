BK-BSCP 部署安装
================

[TOC]

# 部署安装流程
## 1.安装MySQL实例

[参见官方安装教程] <https://dev.mysql.com/doc/mysql-installation-excerpt/5.7/en/linux-installation.html>

## 2.安装Etcd集群

[参见官方安装教程] <https://etcd.io/docs/v3.4.0/op-guide/>

## 3.安装NATS消息队列

[参见官方安装教程] <https://docs.nats.io/nats-server/installation>

## 4.打包编译

### 4.1编译

通过git或下载源码切换到指定版本，在项目根目录通过`make`进行编译,

```shell
[dev@dev bk-bscp]$ ll
总用量 92
drwxrwxr-x  2  dev dev  4096 5月  11 12:09 api
drwxrwxr-x 15  dev dev  4096 5月  26 21:23 cmd
drwxrwxr-x  3  dev dev  4096 7月  23 15:49 docs
-rw-rw-r--  1  dev dev  2040 7月  22 14:42 go.mod
-rw-rw-r--  1  dev dev 44964 7月  22 14:42 go.sum
drwxrwxr-x  9  dev dev  4096 5月  11 12:09 internal
-rw-rw-r--  1  dev dev   852 7月  22 14:42 Makefile
drwxrwxr-x  9  dev dev  4096 5月  11 12:09 pkg
-rw-rw-r--  1  dev dev  2865 7月   6 15:10 README.md
drwxrwxr-x  7  dev dev  4096 5月  11 12:09 scripts
drwxrwxr-x  4  dev dev  4096 5月  11 12:09 test
drwxrwxr-x  3  dev dev  4096 5月  11 12:09 third_party
drwxrwxr-x  3  dev dev  4096 5月  11 12:09 tools
[dev@dev bk-bscp]$
[dev@dev bk-bscp]$ make
Building...
```

### 4.2打包

编译完成之后，在根目录下会生成`build`目录, 目录中会按照版本生成最终的打包目录，其中包含每个模块的二进制和配置示例文件，

```shell
[dev@dev bk-bscp]$ ll
总用量 96
drwxrwxr-x  2 dev dev  4096 5月  11 12:09 api
drwxrwxr-x  3 dev dev  4096 7月  23 15:51 build
drwxrwxr-x 15 dev dev  4096 5月  26 21:23 cmd
drwxrwxr-x  3 dev dev  4096 7月  23 15:56 docs
-rw-rw-r--  1 dev dev  2040 7月  22 14:42 go.mod
-rw-rw-r--  1 dev dev 44964 7月  22 14:42 go.sum
drwxrwxr-x  9 dev dev  4096 5月  11 12:09 internal
-rw-rw-r--  1 dev dev   852 7月  22 14:42 Makefile
drwxrwxr-x  9 dev dev  4096 5月  11 12:09 pkg
-rw-rw-r--  1 dev dev  2865 7月   6 15:10 README.md
drwxrwxr-x  7 dev dev  4096 5月  11 12:09 scripts
drwxrwxr-x  4 dev dev  4096 5月  11 12:09 test
drwxrwxr-x  3 dev dev  4096 5月  11 12:09 third_party
drwxrwxr-x  3 dev dev  4096 5月  11 12:09 tools
[dev@dev bk-bscp]$ ll build/
总用量 4
drwxrwxr-x 15 dev dev 4096 7月  23 15:56 bk-bscp.0.5.20-20.07.23
[dev@dev bk-bscp]$
[dev@dev bk-bscp]$ ll build/bk-bscp.0.5.20-20.07.23/
总用量 52
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-accessserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-bcs-controller
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-bcs-sidecar
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-businessserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-client
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-connserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-datamanager
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-gateway
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-integrator
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-templateserver
```

模块列表:

- bk-bscp-gateway: 网关服务
- bk-bscp-accessserver: 接入服务 (必须部署)
- bk-bscp-businessserver: 业务逻辑服务 (必须部署)
- bk-bscp-templateserver: 模板服务
- bk-bscp-integrator: 逻辑集成服务
- bk-bscp-datamanager: 数据服务 (必须部署)
- bk-bscp-bcs-controller: BCS通道控制器 (必须部署)
- bk-bscp-connserver: BCS会话链接服务 (必须部署)
- bk-bscp-client: 客户端管理工具
- bk-bscp-bcs-sidecar: BCS sidecar（一般以Docker镜像的方式编译）

## 5.部署模块服务

上一步编译后得到打包目录, 其中子目录包含二进制、配置实例以及启动控制脚本,

```shell
[dev@dev bk-bscp]$ ll bk-bscp-accessserver/
总用量 28632
-rwxrwxr-x 1 dev dev 29309685 7月  23 15:56 bk-bscp-accessserver
-rw-rw-r-- 1 dev dev     1216 7月  23 15:56 bk-bscp-accessserver.sh
drwxrwxr-x 2 dev dev     4096 7月  23 15:56 etc
[dev@dev bk-bscp]$
[dev@dev bk-bscp]$ ll bk-bscp-accessserver/etc/
总用量 8
-rw-rw-r-- 1 dev dev  751 7月  23 15:56 Dockerfile
-rw-rw-r-- 1 dev dev 1148 7月  23 15:56 server.yaml
```

将安装包放置目标的安装位置，如`/data/bscp/`下,

```shell
[dev@dev bscp]$ ll
总用量 56
drwxr-xr-x  4 bscp bscp 4096 7月  22 18:10 bk-bscp-accessserver
drwxr-xr-x  4 bscp bscp 4096 7月  22 18:10 bk-bscp-bcs-controller
drwxr-xr-x  4 bscp bscp 4096 7月  22 18:10 bk-bscp-businessserver
drwxr-xr-x  4 bscp bscp 4096 7月  22 18:10 bk-bscp-connserver
drwxr-xr-x  4 bscp bscp 4096 7月  22 18:10 bk-bscp-datamanager
drwxr-xr-x  5 bscp bscp 4096 7月  22 18:10 bk-bscp-gateway
drwxr-xr-x  4 bscp bscp 4096 7月   8 11:44 bk-bscp-integrator
drwxr-xr-x  4 bscp bscp 4096 7月   8 11:44 bk-bscp-templateserver
drwxr-xr-x 12 bscp bscp 4096 7月  22 18:10 build
[dev@dev bscp]$
```

按照每个模块`etc/server.yaml`中示例指引完成配置文件的配置，以下以` bk-bscp-accessserver`模块为例，

```shell
# 当前模块相关配置信息
server:
    # 服务名，用于服务发现
    servicename: bk-bscp-accessserver
    # RPC监听配置
    endpoint:
        ip: 127.0.0.1
        port: 9510
    # 模块服务信息描述
    metadata: bk-bscp-accessserver

# 授权检查相关配置信息
auth:
    # 是否启用授权检查
    open: false
    # platform权限, 用于第三方平台集成，赋予一定的平台操作权限
    platform: "platform:pwd"
    # admin权限, 管理员权限，可操作任何资源
    admin: "admin:pwd"

# ETCD集群相关配置信息
etcdCluster:
    # 集群USR接口配置
    endpoints:
        - 127.0.0.1:2379
    # 建立链接超时时间
    dialtimeout: 2s

# 业务逻辑服务相关配置信息
businessserver:
    # 服务名，用于服务发现
    servicename: bk-bscp-businessserver
    # RPC调用超时时间
    calltimeout: 3s

# 配置模版服务相关配置信息
templateserver:
    servicename: bk-bscp-templateserver
    calltimeout: 3s

# 集成器服务相关配置信息
integrator:
    # 服务名，用于服务发现
    servicename: bk-bscp-integrator
    # RPC调用超时时间
    calltimeout: 3s

# 日志相关配置信息
logger:
    level: 3
    maxnum: 5
    maxsize: 200
```

逐一对每个模块进行配置，配置完成后通过启动脚本或systemd拉起并管理进程,

```shell
[bscp@dev /data/bscp/bk-bscp-accessserver]# ll
总用量 28636
-rwxr-xr-x 1 bscp bscp 29309685 7月  22 18:10 bk-bscp-accessserver
-rwxr-xr-x 1 bscp bscp     1216 10月  8 2019 bk-bscp-accessserver.sh
drwxr-xr-x 2 bscp bscp     4096 7月   8 11:44 etc
drwxr-xr-x 2 bscp bscp     4096 7月  22 18:11 log
[bscp@dev /data/bscp/bk-bscp-accessserver]# sh bk-bscp-accessserver.sh status
bscp     15850     1  0 7月22 ?       00:03:41 ./bk-bscp-accessserver run
[bscp@dev /data/bscp/bk-bscp-accessserver]# sh bk-bscp-accessserver.sh restart
Stopping bk-bscp-accessserver:                             [  确定  ]
Starting bk-bscp-accessserver:                             [  确定  ]
```

启动后可通过查看log中是否出现ERROR等错误确认服务模块是否成功启动.

错误说明:

```shell
metrics collector setup/runtime, listen tcp :9100: bind: address already in use
```
该错误为promethus metrics监听地址冲突，发生冲突后不会打断进程启动而是忽略metrics监听操作，配置模块时注意区分metrics.endpoint内容以防冲突，
或将不同模块于不同主机进行部署。

## 6.安装客户端工具

进入安装包中的`bk-bscp-client`目录，安装客户端工具,

```shell
[dev@dev bk-bscp-client]$ ll
总用量 23804
-rwxrwxr-x 1 dev dev 24370589 7月  23 15:56 bk-bscp-client
drwxrwxr-x 2 dev dev     4096 7月  23 15:56 etc
-rw-rw-r-- 1 dev dev        0 7月  23 16:27 install.sh
[dev@dev bk-bscp-client]$
[dev@dev bk-bscp-client]$ sh install.sh my-accessserver-host:9510
```

my-accessserver-host:9510, my-accessserver-host为`bk-bscp-accessserver`模块的域名或IP地址，9510为其默认的监听端口。

安装完成之后，在`/etc/bscp/`路径下会生成`client.yaml`文件，该文件为客户端工具的配置文件，用于设置其所请求访问的目标BSCP系统地址,

```shell
[dev@dev ~]$ cat /etc/bscp/client.yaml
kind: bscp-client
version: 0.1.1
host: 127.0.0.1:9510 # 本地部署
[dev@dev ~]$
```

同时在`/usr/local/bin/`路径下会安装最新的二进制,

```shell
[dev@dev ~]$ which bk-bscp-client
/usr/local/bin/bk-bscp-client
[dev@dev ~]$
```

至此已完成客户端的安装，详细客户端使用参见,
[admin-book](admin_book.md)
[client-book](client_book.md)

## 7.初始化系统

- BSCP系统数据库实例分为两类，一类为系统db即`bscpdb`, 另一类为业务分片DB即`shardingdb`;
- BSCP系统DB`bscpdb`在`bk-bscp-datamanager`模块服务启动后会自动创建;
- 业务分片`shardingdb`需要人为创建，可手动配置`bscpdb`中的`t_sharding`表，也可通过`bk-bscp-client`进行创建;

业务初始化等操作参见[admin-book](admin_book.md)介绍。

至此已完成系统的部署和初始化工作，可通过`bk-bscp-client`进行配置的管理和发布操作, 详细参见[client-book](client_book.md).

## 8.服务监控

BSCP服务模块会将内部关键数据规整为metrics，以供promethus进行监控体系建设。
可部署promethus和grafana服务，实现BSCP模块监控。在项目根目录下`scripts/grafana/`提供已配置好的模块监控面板，导入到grafana即可。

# FAQ

## 模块之间如何进行访问的？
> 内部模块之间通过Etcd自动服务发现。外部可通过`bk-bscp-gatewway`提供的HTTP Restful API或`bk-bscp-accessserver`提供的gRPC API进行访问请求。

## 业务节点如何访问并拉取数据？
> 业务sidecar节点通过域名(可通过环境变量`BSCP_BCSSIDECAR_CONNSERVER_HOSTNAME`设置)访问BSCP的`bk-bscp-connserver`服务，建立信令和数据通道。
