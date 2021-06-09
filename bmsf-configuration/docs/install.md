BK-BSCP 部署安装说明
===================

[TOC]

# 安装包结构

```shell
[dev@dev package]$ ll bscp_ee-1.1.1
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

目录说明，

- `bk-bscp-xxxx`: 各个服务模块的安装包, 内含编译后的二进制等;
- `install`: 快速安装工具，适用于社区等非蓝鲸内部运维开发人员和用户进行首次安装部署;
- `support-files`: 企业版规范的安装工具，适用于蓝鲸以及相关服务商进行安装部署和版本升级;

# 快速安装教程

[快速安装参见](fast_install.md)

# 企业版安装方式

略

# healthz

健康检查接口:

```shell
curl -vv http://apiserver-ip:apiserver-port/healthz
```

# 服务升级

升级支持两种模式：

- 全量升级：升级到最新版本
- 指定升级：升级到指定版本

升级指令原语:

```shell
全量升级: curl -vv -X POST http://pacther-ip:pacther-port/api/v2/patch/{operator}

指定版本升级: curl -vv -X POST http://pacther-ip:pacther-port/api/v2/patch/{limit_version}/{operator}
```
