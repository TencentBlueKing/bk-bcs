# BSCP部署文档

## 1. 依赖

### 1.1 依赖第三方系统

- 蓝鲸配置平台
- 蓝鲸制品库
- 蓝鲸权限中心
- 蓝鲸 API 网关

### 1.2 依赖第三方组件

- Mysql >= 8.0.17
- Etcd >= 3.2.0
- Redis-Cluster >= 4.0



## 2. BSCP 微服务进程

| 微服务名称           | 描述                             |
| -------------------- | -------------------------------- |
| bk-bscp-apiserver    | 网关微服务，是管理端接口的网关   |
| bk-bscp-authserver   | 鉴权微服务，提供资源操作鉴权功能 |
| bk-bscp-cacheservice | 缓存微服务，提供缓存管理功能     |
| bk-bscp-configserver | 配置微服务，提供各类资源管理功能 |
| bk-bscp-dataservice  | 数据微服务，提供数据管理功能     |
| bk-bscp-feedserver   | 配置拉取微服务，提供拉取配置功能 |



## 3. 前置准备

### 3.1 部署Mysql

[参考官方安装教程] https://dev.mysql.com/doc/mysql-installation-excerpt/8.0/en/linux-installation.html

### 3.2 部署Etcd

[参考官方安装教程] https://etcd.io/docs/v3.2/op-guide/

### 3.3 部署Redis-Cluster

[参考官方安装教程] https://redis.io/docs/manual/scaling/

### 3.4 BSCP应用创建

在蓝鲸开发者中心中创建BSCP应用，应用ID为 bk-bscp。如果使用其他应用id，会导致BSCP在权限中心注册权限模型失败，这是因为权限中心某些版本注册权限模型，会校验 SystemID 和 AppCode是否相同导致。

### 3.5 蓝鲸配置平台

BSCP业务列表来自于蓝鲸配置平台。调用蓝鲸配置平台需要 BSCP appCode、appSecret（appCode、appSecret可以在蓝鲸开发者中心中获取），以及一个有权限拉取蓝鲸配置平台业务列表的用户账号。

### 3.6 蓝鲸制品库

BSCP配置文件内容存放于蓝鲸制品库。需要从制品库获取BSCP平台认证Token，并且通过该Token，在制品库创建一个BSCP项目，以及该项目管理员用户账号。

### 3.7 蓝鲸权限中心

BSCP鉴权操作依赖于蓝鲸权限中心。调用蓝鲸权限中心需要 BSCP appCode、appSecret（appCode、appSecret可以在蓝鲸开发者中心中获取），需要在蓝鲸权限中心添加BSCP应用的白名单。

### 3.8 蓝鲸 API 网关

BSCP接口是通过蓝鲸 API 网关对外提供服务。Release包中的api目录下存放了 apiserver 和 feedserver 网关的资源配置、资源文档，需要将其导入对应的网关，并进行版本发布。此外，还需要获取 apiserver 和 feedserver 网关的API公钥(指纹)。

### 3.9 初始化DB

**登陆数据库**

```shell
mysql -uroot -p
```

**BSCP DB初始化**

```bash
# 使用data-service的migrate子命令进行DB初始化，配置文件路径根据实际情况进行调整
./bk-bscp-dataservice migrate up -c ./etc/data_service.yaml
```


## 4. 修改微服务配置文件

前置准备已经获取到了BSCP配置文件中需要的全部必填配置参数，将 Release包中的 etc 目录下的各微服务配置文件进行修改，部分 mysql 或者 redis 等配置参数可按需配置，如果不配置则使用默认值，配置文件中有详细说明。apiserver_api_gw_public.key 与 feedserver_api_gw_public.key 文件分别替换为 apiserver 和 feedserver 网关的API公钥(指纹)。

**bscp-release/etc下文件说明：**

```shell
├── api_server.yaml												# apiserver 配置文件
├── apiserver_api_gw_public.key						# apiserver 网关的API公钥(指纹)
├── auth_server.yaml											# authserver 配置文件
├── cache_service.yaml										# cacheservice 配置文件
├── config_server.yaml										# configserver 配置文件
├── data_service.yaml											# dataservice 配置文件
├── feed_server.yaml											# feedserver 配置文件
└── feedserver_api_gw_public.key					# feedserver 网关的API公钥(指纹)
```



### 5. 启动服务

**各服务启动命令如下：**

```shell
bk-bscp-apiserver --config-file /data/bkee/etc/bscp/api_server.yaml --public-key /data/bkee/etc/bscp/api_gw_public.key
bk-bscp-authserver --config-file /data/bkee/etc/bscp/auth_server.yaml
bk-bscp-cacheservice --config-file /data/bkee/etc/bscp/cache_service.yaml
bk-bscp-configserver --config-file /data/bkee/etc/bscp/config_server.yaml
bk-bscp-dataservice --config-file /data/bkee/etc/bscp/data_service.yaml
bk-bscp-feedserver --config-file /data/bkee/etc/bscp/feed_server.yaml --public-key /data/bkee/etc/bscp/fs_api_gw_public.key
```
