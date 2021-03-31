BK-BSCP 快速部署安装教程
=======================

[TOC]

# 部署安装流程
## 1.安装MySQL实例

[参见官方安装教程] <https://dev.mysql.com/doc/mysql-installation-excerpt/5.7/en/linux-installation.html>

## 2.安装Etcd集群

[参见官方安装教程] <https://etcd.io/docs/v3.4.0/op-guide/>

## 4.打包编译

### 4.1编译

通过git或下载源码切换到指定版本，在项目根目录通过`make server`进行编译,

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
[dev@dev bk-bscp]$ make server
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
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-apiserver
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-authserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-configserver
drwxrwxr-x 4 dev dev 4096 7月  23 15:56 bk-bscp-templateserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-datamanager
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-gse-controller
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-tunnelserver
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 bk-bscp-patcher
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 install
drwxrwxr-x 3 dev dev 4096 7月  23 15:56 support-files
```

模块列表:

- bk-bscp-apiserver: 接入服务
- bk-bscp-authserver: 权限服务
- bk-bscp-patcher: 升级补丁服务
- bk-bscp-configserver: 配置服务
- bk-bscp-templateserver: 模板服务
- bk-bscp-datamanager: 数据服务
- bk-bscp-gse-controller: GSE通道控制器
- bk-bscp-tunnelserver: GSE通道服务

## 5.部署模块服务

上一步编译后得到打包目录, 其中子目录包含二进制、配置实例以及启动控制脚本,

```shell
[dev@dev bk-bscp]$ ll bk-bscp-apiserver/
总用量 28632
-rwxrwxr-x 1 dev dev 29309685 7月  23 15:56 bk-bscp-apiserver
-rw-rw-r-- 1 dev dev     1216 7月  23 15:56 bk-bscp-apiserver.sh
drwxrwxr-x 2 dev dev     4096 7月  23 15:56 etc
[dev@dev bk-bscp]$
[dev@dev bk-bscp]$ ll bk-bscp-apiserver/etc/
总用量 8
-rw-rw-r-- 1 dev dev 1148 7月  23 15:56 server.yaml.template
```

### 5.1 编辑配置变量

修改`install`目录下`bscp.env`中的服务配置变量(详细参见注释说明)，

### 5.2 生成模块配置

执行`install`目录下`generate.sh`脚本，在对应的服务模块目录下会根据`bscp.env`中的配置渲染模块所需的`server.yaml`

### 5.3 初始化数据库

执行`install`目录下`init_db.sh`脚本，将根据`bscp.env`中的配置和`bscp.sql`中的SQL进行数据库初始化, `bscp.sql`为系统当前
版本最新数据表结构，适用于首次部署安装

### 5.3 安装模块

执行`install`目录下`install.sh`脚本，将在`bscp.env`中`{HOME_DIR}`下安装模块二进制文件和渲染后的配置文件

至此已完成系统的部署和初始化工作。

## 6.拉起并管理进程

执行`install`目录下`restart_all.sh`脚本，启动所有模块。

也可以在安装完成后通过附带的shell启动脚本或配置systemd拉起模块进程,

```shell
[bscp@dev /data/bscp/bk-bscp-apiserver]# ll
总用量 28636
-rwxr-xr-x 1 bscp bscp 29309685 7月  22 18:10 bk-bscp-apiserver
-rwxr-xr-x 1 bscp bscp     1216 7月   8 18:10 bk-bscp-apiserver.sh
drwxr-xr-x 2 bscp bscp     4096 7月   8 11:44 etc

[bscp@dev /data/bscp/bk-bscp-apiserver]# sh bk-bscp-apiserver.sh status
bscp     15850     1  0 7月22 ?       00:03:41 ./bk-bscp-apiserver run
[bscp@dev /data/bscp/bk-bscp-apiserver]# sh bk-bscp-apiserver.sh restart
Stopping bk-bscp-apiserver:                             [  确定  ]
Starting bk-bscp-apiserver:                             [  确定  ]
```

错误说明:

```shell
metrics collector setup/runtime, listen tcp :9100: bind: address already in use
```
该错误为promethus metrics监听地址冲突，发生冲突后不会打断进程启动而是忽略metrics监听操作，配置模块时注意区分metrics.endpoint内容以防冲突，
或将不同模块于不同主机进行部署。

`注意`: 配置文件中默认模块的服务接口监听地址都是0.0.0.0, 这种监听方式可能会暴露到外网（若有外网IP），同时多机器部署时会造成跨机器间模块调用失败，
故此只适合单机器部署全部模块的场景，若多机器部署要修改为非0.0.0.0和127.0.0.1的监听地址，可以通过修改配置文件或命令行参数指定endpoint的方式修改。

## 7.服务升级

执行`install`目录下`upgrade.sh`脚本执行升级，升级支持两种模式：

- 全量升级：升级到最新版本
- 指定升级：升级到指定版本

```shell
Usage:
     upgrade.sh {operator}
```

升级指令原语:

```shell
全量升级: curl -vv -X POST http://pacther-ip:pacther-port/api/v2/patch/{operator}

指定版本升级: curl -vv -X POST http://pacther-ip:pacther-port/api/v2/patch/{limit_version}/{operator}
```

## 8.服务监控

BSCP服务模块会将内部关键数据规整为metrics，以供promethus进行监控体系建设。
可部署promethus和grafana服务，实现BSCP模块监控。在项目根目录下`scripts/grafana/`提供已配置好的模块监控面板，导入到grafana即可。

# FAQ
