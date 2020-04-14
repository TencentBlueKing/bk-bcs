# bcs api-gateway方案构建

## 原因与背景

* **独立后台服务**

  统一整合容器服务成为独立形态，为client与SaaS提供一致的服务支持。避免功能对SaaS形态的依赖与绑定，
  例如管理内容强制注入，导致一致性被破坏。

* **当前问题**
  * 不具备插件化扩展的基础与框架
  * 缺乏反向代理的转发能可配置能力，每次都需要编写模块代码
  * k8s原生转发使用复杂，两次ID转换不友好
  * cmd无法独立使用webconsole能力

* **插件化需求**

  分离bcs-api中转发特性与管理特性功能（K8S集群管理、用户管理、tke集群管理），增强转发、限流、熔断、https安全防护。
  * 扩展插件化扩展对接权限中心(自定义权限)
  * 扩展用户临时授权能力，可以让用户统一SaaS、kubectl的体验
  * 配置化、插件化实现新功能模块接入api，例如etcd集群管理、api多级级联跨云管理
  * 优化与整合webconsole能力，bcs-clien/kubectl实现webconsole使用
  * 优化kubectl使用方式，简化K8S转发，并方便kubectl使用
  * 实现轻量化部署

## 现有功能分析

* V4版本转发功能(bcsapi/v4)
  * clusterkeeper代理转发，`未来下架`
  * netservice，IP地址池数据查询，`未来下架`
  * networkdetection，集群网络基础连通性探测
  * kubernetes，driver方式下受限的k8s能力转发，基于header转发
  * mesos，driver方式下mesos能力全转发，基于header转发
  * storage，实时缓存数据转发
* mesos webconsole转发(bcsapi/v1/webconsole), websocket特性
* 管理特性功能(/rest)，常规转发
* k8s原生api转发(/tunnels/clusters/{cluster_identifier}/sub), reverse_proxy

## 解决方案 

**方案说明**
* 引入开源apigateway插件，构建api-gateway层
* URL转发规则规范化，转发规则需要明确差异版本支持
* 配置化完成各模块反向代理功能，集成bcs当前服务发现能力实现转发动态化
* 新增熔断和限流保护，开启https保护
* 扩展鉴权插件，定向完成权限中心与自定义权限对接
* 扩展临时token授权能力，针对bcs-client、kubectl实现cmd使用临时授权
* bcs-client集成webconsole能力

**实现计划**
  * 构建bcs-api-gateway，部分兼容转发模式，扩展服务注册
  * 规范化部分转发模式，如k8s原生转发，mesos console转发
  * 从bcs-api剥离用户授权管理单独模块，api-gateway进行反向代理，实现权限模型同步，并兼容当前token管理模式
  * 从bcs-api剥离tke集群管理模块，api-gateway进行反向代理，实现tke管理能力插件化
  * 扩展插件权限计算模块，优先对接openPolicy，实现对外平台与临时授权计算
  * 扩展对接蓝鲸权限中心V3插件，用于实现各平台鉴权

**方案结构**

![](./gateway-solution.png)

选型：
* kong
* caddy2

**方案上线**

* bcs-api-gateway与原bcs-api同时运行，给与外部平台调整时间(3-4个月)
* 老版本bcs-client持续维护，非bug修复不做代码调整
* 扩展bcs-client版本为bkbcsctl支持用户临时授权，webconsole

**相关风险**

* 返回码差异

## kong服务安装


```bash
#!/bin/bash

docker run -tid --name api-gateway-storage \
    --restart always \
    -e "POSTGRES_USER=kong" \
    -e "POSTGRES_DB=kong" \
    -p 5432:5432 \
    -v /data/bcs/postgresql/data:/var/lib/postgresql/data \
    postgres:alpine
```

**安装kong**

centos采用rpm安装方式，避免自行编译与组装。
下载地址: https://docs.konghq.com/install/centos/

```bash
rpm -ivh kong-2.0.2.el7.amd64.rpm
```

相关配置路径：
* openresty, kong源码目录: /usr/local/share/lua/5.1
* nginx,openresty执行工具：/usr/local/openresty
* kong项目目录：/usr/local/kong
* nginx配置文件：/usr/local/kong/nginx.conf
* kong默认配置文件：/etc/kong/kong.conf
* kong的服务转发配置，位置任意，在kong.conf配置中，如启用数据库，该配置忽略

kong.conf配置调整
```conf
database = postgres             # Determines which of PostgreSQL or Cassandra
                                # this node will use as its datastore.
                                # Accepted values are `postgres`,
                                # `cassandra`, and `off`.

pg_host = 127.0.0.1             # Host of the Postgres server.
pg_port = 5432                  # Port of the Postgres server.
pg_timeout = 5000               # Defines the timeout (in ms), for connecting,
                                # reading and writing.

pg_user = kong                  # Postgres user.
pg_password =                   # Postgres user's password.
pg_database = kong 

client_ssl = on                 # Determines if Nginx should send client-side
                                # SSL certificates when proxying requests.
client_ssl_cert = /data/bcs/bcs-api-gateway/cert/bcs.crt              
                                # If `client_ssl` is enabled, the absolute
                                # path to the client SSL certificate for the
                                # `proxy_ssl_certificate` directive. Note that
                                # this value is statically defined on the
                                # node, and currently cannot be configured on
                                # a per-API basis.
client_ssl_cert_key = /data/bcs/bcs-api-gateway/cert/bcs.key           
                                # If `client_ssl` is enabled, the absolute
                                # path to the client SSL key for the
                                # `proxy_ssl_certificate_key` address. Note
                                # this value is statically defined on the
                                # node, and currently cannot be configured on
                                # a per-API basis.
```

系统初始化
```shell
kong migrations bootstrap -c /etc/kong/kong.conf
```

## 服务发现demo注册

**注册storage**

其中service中通过host与upstream关联
```bash
#storage service
curl -XPOST localhost:8001/services \
  -d"name=storage" -d"url=https://storage.bkbcs.tencent.com/bcsstorage/v1/" \
  -d"tags[]=storage" -d"tags[]=bcs-service"
#storage route
curl -XPOST localhost:8001/services/storage/routes \
  -d"name=storage" -d"protocols[]=http" -d"paths[]=/bcsapi/v4/storage/" \
  -d"strip_path=true"
#storage upstream
curl -XPOST localhost:8001/upstreams -d"name=storage.bkbcs.tencent.com" \
  -d"algorithm=round-robin"
#storage upstream target
curl -XPUT localhost:8001/upstreams/storage.bkbcs.tencent.com/targets \
  -d"target=127.0.0.3:8080" -d"weight=100" -d"tags[]=storage" -d"tags[]=bcs-service"
```

**注册mesos-driver**

```bash
#mesosdriver service
curl -XPOST localhost:8001/services \
  -d"name=01.mesosdriver" -d"url=https://01.mesosdriver.bkbcs.tencent.com/mesosdriver/v4/" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos" -d"tags[]=BCS-MESOS-01"
#mesosdriver route
curl -XPOST localhost:8001/services/01.mesosdriver/routes \
  -d"name=01.mesosdriver" -d"protocols[]=http" -d"paths[]=/bcsapi/v4/scheduler/mesos/" \
  -d"strip_path=true" -d"headers.BCS-ClusterID=BCS-MESOS-01" -d"tags[]=BCS-MESOS-01" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos"
#mesosdriver upstream
curl -XPOST localhost:8001/upstreams -d"name=01.mesosdriver.bkbcs.tencent.com" \
  -d"algorithm=round-robin" -d"tags[]=BCS-MESOS-01" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos"
#mesosdriver upstream target
curl -XPOST localhost:8001/upstreams/01.mesosdriver.bkbcs.tencent.com/targets \
  -d"target=127.0.0.2:8080" -d"weight=100" -d"tags[]=BCS-MESOS-01" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos"
```

**注册kube-agent**

```bash
#kubeagent service
curl -XPOST localhost:8001/services \
  -d"name=01.kube-agent" -d"url=https://01.kube-agent.bkbcs.tencent.com/" \
  -d"tags[]=kubeagent" -d"tags[]=bcs-k8s" -d"tags[]=BCS-K8S-01"
#kubeagent header plugin
curl -XPOST localhost:8001/services/01.kube-agent/plugins \
  -d"name=request-transformer" -d"config.remove.headers=Authorization" 
#kubeagent route
curl -XPOST localhost:8001/services/01.kube-agent/routes \
  -d"name=01.kube-agent" -d"protocols[]=http" -d"paths[]=/tunnels/clusters/BCS-K8S-01" \
  -d"strip_path=true" -d"tags[]=BCS-K8S-01" \
  -d"tags[]=kubeagent" -d"tags[]=bcs-k8s"
#kubeagent upstream
curl -XPOST localhost:8001/upstreams -d"name=01.kube-agent.bkbcs.tencent.com" \
  -d"algorithm=round-robin" -d"tags[]=BCS-K8S-01" \
  -d"tags[]=kubeagent" -d"tags[]=bcs-k8s"
#kubeagent upstream target
curl -XPOST localhost:8001/upstreams/01.kube-agent.bkbcs.tencent.com/targets \
  -d"target=127.0.0.1:8080" -d"weight=100" -d"tags[]=BCS-K8S-01" \
  -d"tags[]=kubeagent" -d"tags[]=bcs-k8s"

```

## 服务注册kong细则

### bcs服务发现扩展

针对服务发现扩展内容
* ipv6
* external_ipv6

### kong服务命名规范

* 服务信息索引名称：
  * 非集群关联模块以内部定义模块名称为标准，例如storage，cluster等
  * 带有集群信息则加入集群编号，例如mesosdriver和kubedriver等，为10001-mesosdriver，20027-kubedriver
* 服务Host命名规则，使用域bkbcs.tencent.com
  * 非集群模块为服务信息索引 + bkbcs.tencent.com，例如storage.bkbcs.tencent.com
  * 集群模块增加集群ID进行识别，例如01.mesosdriver.bkbcs.tencent.com