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

方案说明

* 引入开源apigateway插件，构建api-gateway层
* URL转发规则规范化，转发规则需要明确差异版本支持
* 配置化完成各模块反向代理功能，集成bcs当前服务发现能力实现转发动态化
* 新增熔断和限流保护，开启https/grpcs保护
* 扩展鉴权插件，定向完成权限中心与自定义权限对接
* 扩展临时token授权能力，针对bcs-client、kubectl实现cmd使用临时授权
* bcs-client集成webconsole能力

实现计划

* 构建bcs-api-gateway，部分兼容转发模式，扩展服务注册
* 规范化部分转发模式，如k8s原生转发，mesos console转发
* 从bcs-api剥离用户授权管理单独模块，api-gateway进行反向代理，实现权限模型同步，并兼容当前token管理模式
* 从bcs-api剥离tke集群管理模块，api-gateway进行反向代理，实现tke管理能力插件化
* 基于模块访问构建授权能力
* 支持grpc转发注册

方案结构

![架构图](./gateway-solution.png)

方案选型：kong/apisix

方案上线

* bcs-api-gateway与原bcs-api同时运行，给与外部平台调整时间(3-4个月)
* 老版本bcs-client持续维护，非bug修复不做代码调整
* 扩展bcs-client命令支持用户临时授权，webconsole

相关风险

* 返回码差异

## 方案验证性说明

### kong服务安装

```bash
#!/bin/bash

docker run -tid --name api-gateway-storage \
    --restart always \
    -e "POSTGRES_USER=kong" \
    -e "POSTGRES_PASSWORD=xxx" \
    -e "POSTGRES_DB=kong" \
    -p 5432:5432 \
    -v /data/bcs/postgresql/data:/var/lib/postgresql/data \
    postgres:alpine
```

安装kong

centos采用rpm安装方式，避免自行编译与组装。
[下载地址](https://docs.konghq.com/install/centos/)

```bash
rpm -ivh kong-2.0.4.el7.amd64.rpm
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
plugins = bundled,bkbcs-auth    # 加载bkbcs鉴权插件
```

系统初始化

```shell
kong migrations bootstrap -c /etc/kong/kong.conf
```

### 服务发现demo注册原理说明

以下流程为通过bcs-gateway-discovery自动完成模块注册

#### 注册storage

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

#### 注册mesos-driver

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

#### 注册kube-agent

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

#### bcs-mesh-manager注册

`注意`：在api-gateway转发特性中，bcs-mesh-manager需要同时提供玩http与grpc转发支持。
其中：

* http接口用于向蓝鲸PaaS企业总线注册，为所有SaaS提供相关特性
* grpc接口用于内部模块、bcs-client使用

```shell
#mesh service
curl -XPOST localhost:8001/services \
  -d"name=meshmanager" -d"protocol=grpcs" -d"host=meshmanager.bkbcs.tencent.com"
#mesh header plugin
curl -XPOST localhost:8001/services/meshmanager/plugins \
  -d"name=request-transformer" -d"config.remove.headers=Authorization"
#mesh route
curl -XPOST localhost:8001/services/meshmanager/routes \
  -d"name=bcs-mesh-manager" -d"protocols[]=grpc" -d"protocols[]=grpcs" -d"paths[]=/meshmanager.MeshManager/"

curl -XDELETE localhost:8001/services/meshmanager/routes/meshmanager

#mesh upstream
curl -XPOST localhost:8001/upstreams -d"name=meshmanager.bkbcs.tencent.com" \
  -d"algorithm=round-robin"
#kubeagent upstream target
curl -XPOST localhost:8001/upstreams/meshmanager.bkbcs.tencent.com/targets \
  -d"target=xxxxxxxxxxxx" -d"weight=100"
```

### 服务注册kong细则

#### bcs服务发现扩展

针对服务发现扩展内容

* ipv6
* external_ipv6

#### kong服务命名规范

* 服务信息索引名称：
  * 非集群关联模块以内部定义模块名称为标准，例如storage，cluster等
  * 带有集群信息则加入集群编号，例如mesosdriver和kubedriver等，为mesosdriver-10001，kubedriver-200001
* 服务Host命名规则，使用域bkbcs.tencent.com
  * 非集群模块为服务信息索引 + bkbcs.tencent.com，例如storage.bkbcs.tencent.com
  * 集群模块增加集群ID进行识别，例如01.mesosdriver.bkbcs.tencent.com

#### bcs-gateway-discovery部署

bcs-gateway-discovery的主要用于对接BCS现有的服务发现机制，利用kong admin api完成模块服务数据注册至kong中，
利用kong网关能力实现转发服务转发。bcs-gateway-discovery部署有以下要求：

* 多实例部署，多个实例会通过bcs服务发现会完成master选择，仅有master角色会同步数据至kong中
* 与kong实例必须同机部署，kong admin接口必须仅针对localhost开启
* **注意**：discovery需要请求user-manager提取k8s集群信息，配置中需要有token授权

#### bkbcs-auth鉴权依赖bcs-user-manager

启用kong作为bcs gateway，在部分受限环境中，可以开启bkbcs-auth插件对接bcs-user-manager实现token鉴权。
可以使用bcs-client命令/接口/bk-bcs-saas等完成token申请。在使用kubectl、bcs-client、独立使用接口时
附带对应的token实现gateway受限访问。

#### 基于websocket tunnel的跨云扩展

bcs-user-manager承担了bcs-api所有集群管理能力，包括用于跨云穿透的websocket tunnel。

以下为bcs-mesos-driver的穿透规则

```bash
#mesosdriver service
curl -XPOST localhost:8001/services \
  -d"name=mesosdriver-tunnel" -d"url=https://usermanager.bkbcs.tencent.com/mesosdriver/v4/" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos"
#mesosdriver header plugin
curl -XPOST localhost:8001/services/mesosdriver-tunnel/plugins \
  -d"name=request-transformer" -d"config.remove.headers=Authorization" -d"config.add.headers=Authorization: Bearer xxxxxxxxxxxxxxxxxxxxxx"
#mesosdriver route
curl -XPOST localhost:8001/services/mesosdriver-tunnel/routes \
  -d"name=mesosdriver-tunnel" -d"protocols[]=http" -d"protocols[]=https" \
  -d"paths[]=/bcsapi/v4/scheduler/mesos/" -d"paths[]=/bcsapi/v1/" \
  -d"strip_path=true" \
  -d"tags[]=mesosdriver" -d"tags[]=bcs-mesos"
#route plugin for auth
curl -XPOST localhost:8001/routes/mesosdriver-tunnel/plugins/ \
  -d"name=bkbcs-auth" -d"config.bkbcs_auth_endpoints=https://usermanager.bkbcs.tencent.com" \
  -d"config.keepalive=60000" -d"config.module=mesosdriver" -d"config.retry_count=1" -d"config.timeout=3000" \
  -d"config.token=xxxxxxxxxxxxxxxxxxxxxx"
#no upstream & target, reuse usermanager upstreams
```

以下为bcs-kube-agent的穿透规则

```bash
#kubeagent service
curl -XPOST localhost:8001/services \
  -d"name=kube-agent-tunnel" -d"url=https://usermanager.bkbcs.tencent.com/tunnels/clusters/" \
  -d"tags[]=kubeagent" -d"tags[]=bcs-k8s"
#kubeagent header plugin
curl -XPOST localhost:8001/services/kube-agent-tunnel/plugins \
  -d"name=request-transformer" -d"config.remove.headers=Authorization" -d"config.add.headers=Authorization: Bearer xxxxxxxxxxxxxxxxxxxxxx"
#kubeagent route
curl -XPOST localhost:8001/services/kube-agent-tunnel/routes \
  -d"name=kube-agent-tunnel" -d"protocols[]=http" -d"protocols[]=https" -d"paths[]=/tunnels/clusters/" \
  -d"strip_path=true" -d"tags[]=kubeagent" -d"tags[]=bcs-k8s"
#route plugins
curl -XPOST localhost:8001/routes/kube-agent-tunnel/plugins \
  -d"name=bkbcs-auth" -d"config.bkbcs_auth_endpoints=https://usermanager.bkbcs.tencent.com" \
  -d"config.keepalive=60000" -d"config.module=kubeagent" -d"config.retry_count=1" -d"config.timeout=3000" \
  -d"config.token=xxxxxxxxxxxxxxxxxxxxxx"
```

### openresty支持spdy

kubectl exec命令使用spdy协议，kong仅默认支持http2，需要重编openresty, kong

版本信息：

* kong：2.0.4
* openresty：1．15．8．3

编译依赖：

* 外网访问
* docker
* docker-compose
* nginx-1.15.8-http2-spdy.patch

编译工具：

* github.com/Kong/kong-build-tools.git，`4.8.1`
* github.com/Kong/kong，`2.0.4`

```shell
export http_proxy=http://someproxy.com
export https_proxy=http://someproxy.com

yum install docker-compose

#准备资源
cd ~
git clone github.com/Kong/kong-build-tools
cd ~/kong-build-tools && git checkout 4.8.1
git clone github.com/Kong/kong
cd ~/kong && git checkout 2.0.4

#spdy patch
cp $GOPATH/src/bk-bcs/install/patch/nginx-1.15.8-http2-spdy.patch ~/kong-build-tools/openresty-patches/patches/1．15．8．3
cd ~/kong-build-tools/openresty-build-tools
sed -i '434a\          "--with-http_spdy_module"' kong-ngx-build
```

额外说明：如果需要http_proxy才能访问外网，需要调整kong-build-tools/Makefile中docker build参数，
增加--build-arg http_proxy=http://someproxy.com 等参数。

编译kong时需要依赖一些lua组件，部分组件是使用git协议下载的，不支持http_proxy设置。需要调整为http协议，为build-kong.sh增加

```shell
git config --global url."https://github.com/".insteadOf git@github.com:
git config --global url."https://".insteadOf git://
```

使用centos:7镜像构建kong rpm包

```shell
cd ~/kong-build-tools
make package-kong RESTY_IMAGE_BASE=centos RESTY_IMAGE_TAG=7 PACKAGE_TYPE=rpm
...
...
...
#编译通过后输出在output目录下
ls -lhrt ./output/
total 25M
-rw-r--r-- 1 root root 25M Jul 22 17:43 kong-2.0.4.el7.amd64.rpm
```

## kong方式部署参考流程

以下为手动流程部署参考，实际部署以Job自动化标准任务或者蓝盾流水线为准。

我们需要额外部署kong，bcs-gateway-discovery与bcs-user-manager。二进制发布包目录假设：

```text
bcs-services
|-- bcs-client
|   |-- bcs-client
|   `-- bcs.conf.template
|-- bcs-gateway-discovery
|   |-- Dockerfile
|   |-- kong-2.0.4.el7.amd64.rpm
|   |-- bcs-gateway-discovery
|   |-- bcs-gateway-discovery.json.template
|   |-- bkbcs-auth
|   |   |-- bkbcs.lua
|   |   |-- handler.lua
|   |   `-- schema.lua
|   `-- container-start.sh
|-- bcs-user-manager
|   |-- Dockerfile
|   |-- bcs-user-manager
|   |-- bcs-user-manager.json.template
|   `-- container-start.sh
```

部署目录：/data/bcs/
证书目录：/data/bcs/cert
日志目录：/data/bcs/logs

### 关于证书

模块部署强制使用https协议，需要优先保障证书。归档至/data/bcs/cert目录下，例如：

```text
|-- cert
|   |-- bcs-ca.crt
|   |-- bcs-client.crt
|   |-- bcs-client.key
|   |-- bcs-server.crt
|   `-- bcs-server.key
```

### 部署kong

BCS默认开kong db存储模式，使用postpreSQL，需要提前准备好数据库，具备建库权限。

```shell
cd $release/bcs-services/bcs-gateway-discovery
rpm -ivh kong-2.0.4.el7.amd64.rpm

#auth插件
cp -R bkbcs-auth /usr/local/share/lua/5.1/kong/plugins/

```

kong配置项调整，文件路径/etc/kong/kong.conf

```conf
database = postgres   # 数据库模式
pg_host = 127.0.0.1   # Host of the Postgres server.
pg_port = 5432        # Port of the Postgres server.
pg_timeout = 5000     # Defines the timeout (in ms), for connecting,
pg_user = kong        # Postgres user.
pg_password = kong    #
pg_database = kong    # 数据库名称

client_ssl = on       # 开启客户端SSL
client_ssl_cert = /data/bcs/cert/bcs-server.crt
client_ssl_cert_key = /data/bcs/cert/bcs-server.key
plugins = bundled,bkbcs-auth  # 加载指定插件
proxy_access_log = /data/bcs/logs/bcs/kong-access.log
proxy_error_log = /data/bcs/logs/bcs/kong-error.log
admin_access_log = /data/bcs/logs/bcs/kong-admin_access.log
admin_error_log = /data/bcs/logs/bcs/kong-admin_error.log
```

**注意**考虑安全需求，正式环境需要将kong本身http端口关闭。

kong初始化与启动

```shell
kong migrations bootstrap -c /etc/kong/kong.conf

kong start  -c /etc/kong/kong.conf
```

### 部署bcs-user-manager

bcs-user-manager重构了bcs-api中关于集群和用户管理功能，计划在1.17.x，1.18.x，1.19.x是保持兼容的。
所以，bcs-user-manager与bcs-api共享mysql数据库数据，并且bcs-api中跨云穿透代理的功能默认会合并到该模块，
实际部署时，尽量保障bcs-user-manager与kong在相同可用区/同城同园区，避免网络抖动带来跨云穿透长链中断重连。

配置文件模板请参照[这里](https://github.com/Tencent/bk-bcs/blob/master/install/conf/bcs-services/bcs-user-manager/bcs-user-manager.json.template)。

特别配置参数bootstrap_users说明：

* name：系统默认启动分配的初始化账号名称
* token：32位字母+数字组成的认证admin token，例如：vAyEKvelIqnasMP9sUGWUw1naG8qLues

建议是当系统首次部署初始化时进行配置，配置完成后进行常规平台运维账户管理创建和授权，之后将该参数清空重启。

首次token随机生成可以使用工具生成一个 [http://coolaf.com/tool/rd](http://coolaf.com/tool/rd)

### 部署bcs-gateway-discovery

为了安全考虑，该模块必须与kong同机部署，在容器方案中，bcs-gateway-discovery和kong默认构建在一个镜像中。
重要配置说明：

* admin_api：kong本地默认的管理连接，默认为localhost：8081
* auth_token：cluster-manager使用的amdin token，例如：vAyEKvelIqnasMP9sUGWUw1naG8qLues

完成配置后启动

```shell
cd /data/bcs/bcs-gateway-discovery
./bcs-gateway-discovery --config bcs-gateway-discovery.json &
```

### 系统用户初始化

client配置

```json
{
  "apiserver":"127.0.0.1:8000",
  "bcs_token":"vAyEKvelIqnasMP9sUGWUw1naG8qLues"
}
```

```shell
#client配置后
mkdir -p /var/bcs/
cd /data/bcs/bcs-client
cp bcs.conf.template /var/bcs/bcs.conf

#初始化用户进行授权，主要用于
#* 其他系统调用bcs-api-gateway
#* 用户通过kubectl直接管理特定集群
bcs-client create --type=user --username=someadministrator --usertype=admin
vAyEKvelIqnasMP9sUGWUw1naG8qLues
#为该用户授权
bcs-client grant -t permission -f cluster.grant.json
#其中授权json内容为，其中role划分viewer(信息查阅),manager(管理者)
#{
#  "apiVersion":"v1",
#  "kind":"permission",
#  "metadata": {
#     "name":"someadministrator-pemission"
#  },
#  "spec":{
#     "permissions":[
#       {"user_name":"someadministrator", "resource_type":"cluster", "resource":"BCS-K8S-001", "role":"viewer"}
#     ]
#  }
#}
```

完成授权后，可以使用kubectl针对具体集群可以通过kubeconfig进行访问。相关文件可以通过bk-bcs-saas进行生成，手动编写如下：

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: https://127.0.0.1:8443/tunnels/clusters/BCS-K8S-001/
    insecure-skip-tls-verify: true
  name: cluster
contexts:
- context:
    cluster: cluster
    user: someadministrator
  name: cluster
current-context: cluster
users:
- name: someadministrator
  user:
    token: vAyEKvelIqnasMP9sUGWUw1naG8qLues
```

```shell
kubectl version --kubeconfig ./localkubeconfig
```

### 数据清理

如果存在脏数据，可以主动清理

```shell
curl -s localhost:8001/routes | python -m json.tool | grep name | awk -F'"' '{print $4}' | while read name
do
  curl -XDELETE localhost:8001/routes/$name
done

curl -s localhost:8001/services | python -m json.tool | grep name | awk -F'"' '{print $4}' | while read name
do
  curl -XDELETE localhost:8001/services/$name
done

curl -s localhost:8001/upstreams | python -m json.tool | grep name | awk -F'"' '{print $4}' | while read name
do
  curl -XDELETE localhost:8001/upstreams/$name
done
```

重启bcs-gateway-discovery即可。

## apisix方式部署参考流程

### 安装依赖openresty

```shell
# 安装 epel, `luarocks` 需要它
wget http://dl.fedoraproject.org/pub/epel/epel-release-latest-7.noarch.rpm
sudo rpm -ivh epel-release-latest-7.noarch.rpm

# 添加 OpenResty 源
sudo yum install yum-utils
sudo yum-config-manager --add-repo https://openresty.org/package/centos/openresty.repo
# 如果指定下载版本，修正下/etc/yum.repos.d/openresty.repo下的baseurl中的$releaseserver，例如指定7

yum install -y openresty

# 安装apisix
wget https://github.com/apache/apisix/releases/download/2.1/apisix-2.1-0.el7.noarch.rpm
rpm -ivh apisix-2.1-0.el7.noarch.rpm
```

### etcd启动

```shell
docker run --network host -v /data/bcs/bcs-api-gateway/datas:/var/lib/etcd -tid mirrors.tencent.com/k8s.gcr.io/etcd:3.4.13-0 /usr/local/bin/etcd \
    --advertise-client-urls=http://127.0.0.1:3379 --data-dir=/var/lib/etcd \
    --initial-advertise-peer-urls=http://127.0.0.1:3380 \
    --initial-cluster=apigateway-store=http://127.0.0.1:3380 \
    --listen-client-urls=http://127.0.0.1:3379 \
    --listen-metrics-urls=http://127.0.0.1:3381 --listen-peer-urls=http://127.0.0.1:3380 \
    --name=apigateway-store --snapshot-count=10000

#check
export ETCDCTL_API=3
etcdctl --endpoints=127.0.0.1:3379 get --prefix /apisix
```

### apisix配置

```yaml
apisix:
  node_listen: 8000
  enable_ipv6: false
  admin_key:
    -
      name: "admin"
      key: edd1c9f034335f136f87ad84b625c8f1
      role: admin
  ssl:
    enable: true
    listen_port: 8443
    ssl_cert: /data/bcs/cert/bcs.crt
    ssl_cert_key: /data/bcs/cert/bcs.key
etcd:
  host:
    - "http://127.0.0.1:3379"
nginx_config:
  http_server_configuration_snippet: |
    proxy_ssl_name        $upstream_host;
    proxy_ssl_server_name on;
    proxy_ssl_certificate /data/bcs/cert/bcs.crt;
    proxy_ssl_certificate_key /data/bcs/cert/bcs.key;
  error_log: "/data/bcs/logs/bcs/apisix-error.log"
  http:
    access_log: "/data/bcs/logs/bcs/apisix-access.log"
```

ssl证书注册：

```shell
curl http://127.0.0.1:8000/apisix/admin/ssl/bkbcs \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' \
  -X PUT -d@bkbcs-ssl.json
```

### 服务发现注册

#### 注册bcs-storage

```shell
#upstream
curl http://127.0.0.1:8000/apisix/admin/upstreams/storage \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "type": "roundrobin",
  "nodes": {
    "127.0.0.100:50024": 10
  },
  "retries": 1
}'

#test service
curl http://127.0.0.1:8000/apisix/admin/services/storage \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "upstream_id": "storage",
  "name": "storage",
  "enable_websocket": true,
  "plugins": {
    "limit-req": {
      "rate": 1000,
      "burst": 500,
      "rejected_code": 429,
      "key": "server_addr"
    }
  }
}'

#route, authentication stage
curl http://127.0.0.1:8000/apisix/admin/routes/storage \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "name": "storage",
  "uri": "/bcsapi/v4/storage/*",
  "service_id": "storage",
  "service_protocol": "http",
  "plugins":{
    "proxy-rewrite": {
      "regex_uri": ["/bcsapi/v4/storage/(.*)", "/bcsstorage/v1/$1"],
      "scheme": "https",
      "host": "storage.bkbcs.tencent.com"
    }
  }
}'

## test case
curl -vv http://127.0.0.1:8000/bcsapi/v4/storage/query/mesos/dynamic/clusters/BCS-MESOS-10000/deployment
```

#### 注册bcs-mesos-driver

```shell
#upstream
curl http://127.0.0.1:8000/apisix/admin/upstreams/mesosdriver-10000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "type": "roundrobin",
  "nodes": {
    "127.0.0.99:50020": 10
  },
  "retries": 1
}'

#test service
curl http://127.0.0.1:8000/apisix/admin/services/mesosdriver-10000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "upstream_id": "mesosdriver-10000",
  "name": "mesosdriver-10000",
  "enable_websocket": true,
  "plugins": {
    "limit-req": {
      "rate": 1000,
      "burst": 500,
      "rejected_code": 429,
      "key": "server_addr"
    }
  }
}'

#route, authentication stage, url & header
curl http://127.0.0.1:8000/apisix/admin/routes/mesosdriver-10000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "name": "mesosdriver-10000",
  "uri": "/bcsapi/v4/scheduler/mesos/*",
  "service_id": "mesosdriver-10000",
  "service_protocol": "http",
  "enable_websocket": true,
  "vars": [
    ["http_BCS-ClusterID", "==", "BCS-MESOS-10000"]
  ],
  "plugins":{
    "request-id": {
      "include_in_response": true
    },
    "proxy-rewrite": {
      "regex_uri": ["/bcsapi/v4/scheduler/mesos/(.*)", "/mesosdriver/v4/$1"],
      "scheme": "https",
      "host": "mesosdriver-10000.bkbcs.tencent.com"
    }
  }
}'

## test case
curl --resolve 'bcs-api-gateway.bk.tencent.com:8443:127.0.0.1' -H"BCS-ClusterID: BCS-MESOS-10000" -vv https://bcs-api-gateway.bk.tencent.com:8443//bcsapi/v4/scheduler/mesos/cluster/current/offers

curl -vv -H"BCS-ClusterID: BCS-MESOS-10000" http://127.0.0.1:8000/bcsapi/v4/scheduler/mesos/cluster/current/offers
```

#### bcs-kube-agent注册

```shell
#upstream
curl http://127.0.0.1:8000/apisix/admin/upstreams/kubeagent-15000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -d '
{
  "type": "roundrobin",
  "nodes": {
    "127.0.0.1:6443": 10
  },
  "retries": 1
}'

#test service
curl http://127.0.0.1:8000/apisix/admin/services/kubeagent-15000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "upstream_id": "kubeagent-15000",
  "name": "kubeagent-15000",
  "enable_websocket": true,
  "plugins": {
    "limit-req": {
      "rate": 1,
      "burst": 0,
      "rejected_code": 429,
      "key": "server_addr"
    }
  }
}'

#route, authentication stage, url & header
curl http://127.0.0.1:8000/apisix/admin/routes/kubeagent-15000 \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "name": "kubeagent-15000",
  "uri": "/clusters/BCS-K8S-15000/*",
  "service_id": "kubeagent-15000",
  "service_protocol": "http",
  "enable_websocket": true,
  "plugins": {
    "request-id": {
      "include_in_response": true
    },
    "proxy-rewrite": {
      "regex_uri": ["/clusters/BCS-K8S-15000/(.*)", "/$1"],
      "scheme": "https",
      "host": "kubeagent-15000.bkbcs.tencent.com",
      "headers": {
        "Authorization": "Bearer xxxxxxxxxxxx"
      }
    }
  }
}'

## test case
curl -vv http://127.0.0.1:8000/tunnels/clusters/BCS-K8S-15000/version
```

#### tunnel规则注册

```shell
curl http://127.0.0.1:8000/apisix/admin/routes/kube-agent-tunnel \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "name": "kube-agent-tunnel",
  "uri": "/clusters/*",
  "service_id": "clustermanager-http",
  "service_protocol": "http",
  "enable_websocket": true,
  "plugins": {
    "request-id": {
      "include_in_response": true
    },
    "proxy-rewrite": {
      "regex_uri": ["/clusters/(.*)", "/clustermanager/clusters/$1"],
      "scheme": "https"
    }
  }
}'

curl http://127.0.0.1:8000/apisix/admin/routes/mesosdriver-tunnel \
  -H 'X-API-KEY: edd1c9f034335f136f87ad84b625c8f1' -X PUT -i -d '
{
  "name": "mesosdriver-tunnel",
  "uri": "/bcsapi/v4/scheduler/mesos/*",
  "service_id": "clustermanager-http",
  "service_protocol": "http",
  "enable_websocket": true,
  "plugins": {
    "request-id": {
      "include_in_response": true
    },
    "proxy-rewrite": {
      "regex_uri": ["/bcsapi/v4/scheduler/mesos/(.*)", "/clustermanager/mesosdriver/v4/$1"],
      "scheme": "https"
    }
  }
}'
```
