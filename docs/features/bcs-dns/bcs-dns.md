# 域名解析功能说明

该插件是为适配bcs-scheduler service/endpoints的zookeeper数据结构而开发，目标是
为bcs service下的mesos集群提供DNS服务。主要实现借助coredns框架，补充bcs的插件实现
以实现集成。

Service的域名构成分为五个部分：myservice.namespace.svc.clusterid.zone

* myservice：定义的service名字
* namespace：命名空间
* svc：代表该记录为serivce
* clusterid：该Mesos集群的ID
* zone：dns区域信息

在Mesos集群中访问时，如果没有cluster.zone部分，默认解析为本集群的IP地址。

## dns配置说明

```conf
. :53 {
    log . logs/coredns.log "{remote} - {type} {class} {name} {proto} {size} {rcode} {rsize}" {
        class all
    }
    loadbalance round_robin
    cache 5
}
```

* . 说明解析所有域名
* 53 udp服务端口
* log 开启log插件，解析时输出日志
* loadbalance IP解析算法，暂时只是支持轮询
* cache 开启缓存模块，域名TTL时间为5秒

## bcsscheduler插件配置说明

以下为bcs-scheduler插件的配置文件，bcsshceduler块与log/cache/loadbalance同级

```conf
bcsscheduler bcs.com. {
    #cluster ID, whole domain name is $serviceName.$Namespace.svc.$cluster.$zone
    cluster 10002

    #全量同步刷新周期，单位秒，默认为60
    resyncperiod 60

    #Master注册zookeeper
    registery 127.0.0.1:2181 127.0.0.2:2181

    #数据来源，使用空格间隔
    endpoints 127.0.0.1:2181 127.0.0.2:2181
    endpoints-path /blueking

    #endpoints tls配置，配置保留项
    #endpoints-tls <cert-file> <key-file> <ca-file>

    #数据存储，空格间隔
    storage http://127.0.0.1:2379
    #存储路径
    storage-path /bluekingdns
    #存储tls配置
    #storage-tls <cert-file> <key-file> <ca-file>

    #upstream，上级DNS服务，空格间隔
    upstream 127.0.0.2:53

    #没有数据时交由下一个插件处理
    fallthrough
}
```

* bcsscheduler：插件配置起始，bcs.com.为zone，谨记最后需要点，必填
* cluster：集群ID，必填
* resyncperiod：数据全量同步时间，默认60秒
* endpoints：zookeeper IP列表，空格分开
* endpoints-path：scheduler service，endpoint父路径，默认/blueking
* upstream：上级DNS
* fallthrough：错误时是否交由下一个插件处理

## dns功能结构

### dns解析规则

dns域名构成：$name.$namespace.$type.$clusterid.$zone

域名字符不能包含大写，仅包含小写，数字，中横线-。首字母必须为小写。

* name：服务名字
* namespace：mesos中服务的namespace
* type：固定为svc或者pod，未来可能支持更多类型
* clusterid：配置文件中cluster字段
* zone：bcsscheduler配置后的区域信息。

### 集群内解析规则
例如集群clusterid为10002时
  * 全域名支持： 
  
例如：app.bkapp.svc.10002.bcs.com。这种情况下的解析最为准确。

  * 短域名支持：
    - svcname.ns式两段域名
在集群内进行域名解析服务时，bcs-dns支持两段式短域名解析服务，即serviceName.serviceNamespace的格式。
这种两段式的A类请求，bcs-dns默认会为用户加上后边的svc.type.clusterid.zone信息。即app.bkapp，实际上访问的是app.bkapp.svc.1002.bcs.com。
这种短域名的访问方式有一个弊端，即这种a.b式的两段短域名默认都是集群内的服务，如果用户的域名本身也是两段式的，如g.cn，则无法解析。考虑到后期会支持pod的域名解析服务，建议大家在平常使用时下面的短域名。
    - svcname.ns.svc式3段域名    
这种方式可明确指出该域名为集群内域名，解析准确，后续会逐步废弃上面两段式的域名使用方式。大家后续再接业务，推荐使用此种方式。

* 跨集群解析规则
  对跨集群解析的需求，用户使用前文所述的全域名方式即可。
* 非集群域名解析服务
  对于域名不在bcs.com这个zone内的域名，bcs-dns会将解析服务路由到公司的域名服务器进行解析。
  
* 域名递归解析服务
bcs-dns依靠bcs service提供域名递归解析服务，适用于用户将应用的域名解析到指定的IP地址或公司内的域名上，所在在此种场景下，该service的selector并不生效。该服务仅在bcs service 的type为`ClustgerIP`，ClusterIP字段数组不为空的情况下可使用。具体的使用场景如下
  * ClusterIP字段数组中仅包含合法的IP地址
  
  bcs-dns会将该域名解析的请求解析到这些IP地址上。
  * ClusterIP字段数组中仅包含合法的公司内网域名domains
  
  bcs-dns会将该域名解析的请求，拆解为这些domains解析结果的并集
  * ClusterIP字段数组中仅包含合法的bcs.com这个zone中的全域名
  
  这种场景下是跨集群的域名递归解析服务，bcs-dns会对此请求进行跨集群解析并返回
  * ClusterIP字段数组中包含以上三种情况中的两种或多种组合
  
  bcs-dns会将这些请求进行递归解析并进行合并，最终将所有结果返回给用户。
  
  下面给出示例：
  如，service定义如下
```json
{
    "apiVersion": "v1",
    "kind": "service",
    "metadata": {
        "name": "bcs-service",
        "namespace": "bcsns"
    },
    "spec": {
        "type": "ClusterIP",
        "clusterIP": [
            "a.b.com",
            "c.d.com",
            "192.168.1.1",
            "192.168.1.2",
            "othersvc.otherns.svc.1003.bcs.com"
        ],
        "ports": [
            {
                "name": "test-port",
                "domainName": "www.test.com",
                "path": "/test/path",
                "protocol": "http",
                "servicePort": 8889,
                "nodePort": 0
            }
        ]
    }
}
```
请求如下：
```
dig -b 127.0.0.1 -p 54 bcs-service.bcsns
```
返回结果如下：
```shell
> dig -b 127.0.0.1 -p 54 bcs-service.bcsns   

; EDNS: version: 0, flags:; udp: 4096
;; QUESTION SECTION:
;bcs-service.bcsns.		IN	A

;; ANSWER SECTION:
bcs-service.bcsns.	5	IN	A	192.168.1.2
bcs-service.bcsns.	5	IN	A	192.168.1.1
bcs-service.bcsns. 5 IN A	127.0.0.1
bcs-service.bcsns. 5 IN A	127.0.0.2
bcs-service.bcsns. 5 IN A	127.0.0.3
```

在返回的`ANSWER SECTION`我们可以看到5条记录：

- 前两条记录很清楚，就是我们配置的两个IP地址。
- 剩余三条为其他域名的解析结果。

### srv记录支持
bcs-dns仅支持`集群内`**srv**记录的查询功能。

srv的格式为：

`_port._protocol.service.namespace.svc.zone`

测试命令：

dig srv _port._protocol.service.namespace.svc.clusterid.bcs.com


### pod 域名解析
规则： **`podname.`**myservice.namespace.svc.clusterid.zone

使用时支持两种方式：
* 全域名，即:**`podname.`**myservice.namespace.svc.clusterid.zone
* 本集群时，可使用简写方式，即：**`podname.`**myservice.namespace.svc


## 自定义域名规范

自定义域名服务的`zone`在`bcscustom.com`中

整个自定义域名**建议**包含三部分：

`$user`.`$clusterid`.`bcscustom.com`

其中:
* $user：为用户自定义的服务名称，如job.ied
* $clusterid：为该域名所属的cluster。其中`*`为特殊**集群**，代表在所有集群中均适用、解析。

另外变更域名时，必须 `alias` 参数，用于支持一个域名多个IP的情况, alias 为ip的别名，默认值为hash

存储层storage设计:
- 目前使用 etcd v3 存储域名数据

#### 数据存储举例
- domain
```
{
	"domain": "a.b.c.bcscustom.com",
	"messages": [{
		"host": "127.0.0.1"
	}, {
		"host": "127.0.0.2"
	}]
}
```

- etcd v3 data
    + 注意：所有key的尾部都有`/`

```
/bcscustom/com/bcscustom/bcs-k8s-15009/ied/job/71a51120/
{"host":"127.0.0.4"}
/bcscustom/com/bcscustom/bcs-k8s-15009/ied/job/cd8d6f4/
{"host":"127.0.0.5"}
```

#### 使用 etcd v3 接口注意事项

- 因为v3不再提供ls -r的方式查询，只能用 --prefix, 注意这只是前缀匹配
- 比如用--prefix查询key：/bcs/domain, 它会匹配到 /bcs/domain/xxx, /bcs/domain1/xx
- 参考：https://github.com/etcd-io/etcd/blob/master/clientv3/op.go#L370

#### bcscustom插件配置说明
```conf
.:53 {
    log . logs/coredns.log "{remote} - {type} {class} {name} {proto} {size} {rcode} {rsize}" {
        class all
    }
    loadbalance round_robin
    cache 30
    bcscustom bcscustom.com. {
        upstream 127.0.0.1:53 localhost:53
        etcd-endpoints http://localhost:2379 http://127.0.0.1:2379
        etcd-tls cert.pem key.pem ca.pem
        fallthrough
        root-prefix bcscustom
        listen 127.0.0.1:8099
        ca-file ca.pem
        key-file key.pem
        cert-fiile cert.pem
    }
}

```

#### 参数说明
* **upstream** 若配置，将会将本地无法解析的域名路由到配置的dns server去解析。
* **etcd-endpoints** the etcd endpoints. Defaults to "http://localhost:2379".
* `etcd-tls` followed by:

    * no arguments, if the server certificate is signed by a system-installed CA and no client cert is needed
    * a single argument that is the CA PEM file, if the server cert is not signed by a system CA and no client cert is needed
    * two arguments - path to cert PEM file, the path to private key PEM file - if the server certificate is signed by a system-installed CA and a client certificate is needed
    * three arguments - path to cert PEM file, path to client private key PEM file, path to CA PEM
      file - if the server certificate is not signed by a system-installed CA and client certificate
      is needed.
* **fallthrough** 当请求不在本zone时，是否将请求路由到下一个插件, default false.
* **root-prefix** 该bcscustom实例使用的etcd存储的根目录名称。
* **listen** 自定义域名服务所要监听的地址和绑定的端口，无默认值，必须设定。
* **ca-file**, **key-file**, **cert-file** listen所使用的相关证书。

# 自定义域名使用接口说明

## 创建
- 重复创建不会出错
- alias 字段可以留空，留空时默认采用hash值填充
请求：
```shell
curl -H "Content-Type:application/json" -X POST -d '{
    "domain": "demo.ied01.bcscustom.com",
    "messages": [{
        "alias": "ip0",
        "host": "127.0.0.1"
    }]
}' http://127.0.0.1:8099/bcsdns/v1/domains
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "create success",
  "data": null
}
```

## 更新
- 更新一个不存在的key则为创建
- alias 字段可以留空，留空时默认采用hash值填充
请求：
```shell
curl -H "Content-Type:application/json" -X PUT -d '{
    "domain": "demo.ied01.bcscustom.com",
    "messages": [{
        "host": "127.0.0.5"
    }]
}' http://127.0.0.1:8099/bcsdns/v1/domains
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "update success",
  "data": null
}
```

## 查询
- alias 可传可不传，不传时可能会查出多个ip
请求：
```shell
curl http://127.0.0.1:8099/bcsdns/v1/domains?domain=demo.ied01.bcscustom.com
curl http://127.0.0.1:8099/bcsdns/v1/domains?domain=demo.ied01.bcscustom.com&alias=ip0
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": [{
   "host": "127.0.0.5"
  }]
 }
```

## 删除(校验alias, 非空)
- 如果只是删除domain下的一个ip，建议用该接口
- 接口不允许一次删除多个domain
请求：
```shell
curl -X DELETE http://127.0.0.1:8099/bcsdns/v1/domains/alias?domain=lee.ied.bcscustom.com&alias=ip0
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "delete success",
  "data": null
}
```


## 删除(不校验alias, 可以为空)
- 接口不允许一次删除多个domain
请求：
```shell
curl -X DELETE http://127.0.0.1:8099/bcsdns/v1/domains?domain=lee.ied.bcscustom.com
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "delete success",
  "data": null
}
```


## 列出
请求：
```shell
curl http://127.0.0.1:8099/bcsdns/v1/domains/subdomains?zone=bcscustom.com
```
返回：
```json
{
  "result": true,
  "code": 0,
  "message": "list success",
  "data": [
   {
    "domain": "b.bcscustom.com.",
    "messages": [{
     "alias": "ip0",
     "host": "127.0.0.5",
     "priority": 10,
     "weight": 10,
     "ttl": 30
    }]
   },
   {
    "domain": "a.bcscustom.com.",
    "messages": [{
     "alias": "ip1",
     "host": "127.0.0.4",
     "priority": 10,
     "weight": 10,
     "ttl": 30
    }]
   },
   {
    "domain": "job.*.bcscustom.com.",
    "messages": [{
     "alias": "ip0",
     "host": "127.0.0.6",
     "priority": 10,
     "weight": 10,
     "ttl": 30
    }]
   }
  ]
 }
```

```



## kubernetes域名解析

*kubernetes* enables reading zone data from a kubernetes cluster.
It implements the [spec](https://github.com/kubernetes/dns/blob/master/docs/specification.md)
defined for kubernetes DNS-Based service discovery:

Service `A` records are constructed as "myservice.mynamespace.svc.coredns.local" where:

* "myservice" is the name of the k8s service
* "mynamespace" is the k8s namespace for the service, and
* "svc" indicates this is a service
* "coredns.local" is the zone

Pod `A` records are constructed as "1-2-3-4.mynamespace.pod.coredns.local" where:

* "1-2-3-4" is derived from the ip address of the pod
* "mynamespace" is the k8s namespace for the service, and
* "pod" indicates this is a pod
* "coredns.local" is the zone

Endpoint `A` records are constructed as "epname.myservice.mynamespace.svc.coredns.local" where:

* "epname" is the hostname (or name constructed from IP) of the endpoint
* "myservice" is the name of the k8s service that the endpoint serves
* "mynamespace" is the k8s namespace for the service, and
* "svc" indicates this is a service
* "coredns.local" is the zone

Also supported are PTR and SRV records for services/endpoints.

### Syntax

This is an example kubernetes configuration block, with all options described:

```conf
# kubernetes <zone> [<zone>] ...
#
# Use kubernetes middleware for domain "coredns.local"
# Reverse domain zones can be defined here (e.g. 0.0.10.in-addr.arpa),
# or instead with the "cidrs" option.
#
kubernetes coredns.local {

    # resyncperiod <period>
    #
    # Kubernetes data API resync period. Default is 5m
    # Example values: 60s, 5m, 1h
    #
    resyncperiod 5m
    # endpoint <url>
    #
    # Use url for a remote k8s API endpoint.  If omitted, it will connect to
    # k8s in-cluster using the cluster service account.
    #
    endpoint https://k8s-endpoint:8080
    # tls <cert-filename> <key-filename> <cacert-filename>
    #
    # The tls cert, key and the CA cert filenanames for remote k8s connection.
    # This option is ignored if connecting in-cluster (i.e. endpoint is not
    # specified).
    #
    tls cert key cacert

    # namespaces <namespace> [<namespace>] ...
    #
    # Only expose the k8s namespaces listed.  If this option is omitted
    # all namespaces are exposed
    #
    namespaces demo
    # lables <expression> [,<expression>] ...
    #
    # Only expose the records for kubernetes objects
    # that match this label selector. The label
    # selector syntax is described in the kubernetes
    # API documentation: http://kubernetes.io/docs/user-guide/labels/
    # Example selector below only exposes objects tagged as
    # "application=nginx" in the staging or qa environments.
    #
    labels environment in (staging, qa),application=nginx    
    # pods <disabled|insecure|verified>
    #
    # Set the mode of responding to pod A record requests.
    # e.g 1-2-3-4.ns.pod.zone.  This option is provided to allow use of
    # SSL certs when connecting directly to pods.
    # Valid values: disabled, verified, insecure
    #  disabled: Do not process pod requests, always returning NXDOMAIN
    #  insecure: Always return an A record with IP from request (without
    #            checking k8s).  This option is is vulnerable to abuse if
    #            used maliciously in conjuction with wildcard SSL certs.
    #  verified: Return an A record if there exists a pod in same
    #            namespace with matching IP.  This option requires
    #            substantially more memory than in insecure mode, since it
    #            will maintain a watch on all pods.
    # Default value is "disabled".
    #
    pods disabled
    # cidrs <cidr> [<cidr>] ...
    #
    # Expose cidr ranges to reverse lookups.  Include any number of space
    # delimited cidrs, and or multiple cidrs options on separate lines.
    # kubernetes middleware will respond to PTR requests for ip addresses
    # that fall within these ranges.
    #
    cidrs 10.0.0.0/24
    # fallthrough
    #
    # If a query for a record in the cluster zone results in NXDOMAIN,
    # normally that is what the response will be. However, if you specify
    # this option, the query will instead be passed on down the middleware
    # chain, which can include another middleware to handle the query.
    fallthrough

    #cluster id
    cluster 10002
    registery 127.0.0.1:2181 127.0.0.2:2181
    storage https://127.0.0.1:2379 https://127.0.0.2:2379
    storage-path /custompath
    storageTLS <cert-file> <key-file> <ca-file>
}

```

### Wildcards

Some query labels accept a wildcard value to match any value.
If a label is a valid wildcard (\*, or the word "any"), then that label will match
all values.  The labels that accept wildcards are:

* _service_ in an `A` record request: _service_.namespace.svc.zone.
   * e.g. `*.ns.svc.myzone.local`
* _namespace_ in an `A` record request: service._namespace_.svc.zone.
   * e.g. `nginx.*.svc.myzone.local`
* _port and/or protocol_ in an `SRV` request: __port_.__protocol_.service.namespace.svc.zone.
   * e.g. `_http.*.service.ns.svc.`
* multiple wild cards are allowed in a single query.
   * e.g. `A` Request `*.*.svc.zone.` or `SRV` request `*.*.*.*.svc.zone.`

### Deployment in Kubernetes

See the [deployment](https://github.com/coredns/deployment) repository for details on how
to deploy CoreDNS in Kubernetes.

