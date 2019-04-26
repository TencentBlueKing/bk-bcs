# Mesos service定义

## json文件

service主要用**服务发现**，**DNS基础数据**，**loadbalance服务导出**。

```json
{
    "apiVersion":"v4",
    "kind":"service",
    "metadata":{
        "name":"template-service",
        "namespace":"defaultGroup",
        "labels":{
            "BCSGROUP": "external",
            "BCSBALANCE": "source|roundrobin|leastconn",
            "BCS-WEIGHT-summer": "3",
            "BCS-WEIGHT-wingame": "7"
        }
    },
    "spec": {
        "selector": {
            "label-one": "summer",
            "label-two": "wingame"
        },
        "type": "ClusterIP|NodePort|None|Integration",
        "clusterIP": ["127.0.0.1", "127.0.0.2"],
        "ports": [
            {
                "name": "http_8080",
                "domainName": "a.business.qq.com",
                "path": "/local/path",
                "protocol": "http",
                "servicePort": 80,
                "targetPort": 8080,
                "nodePort": 31000
            },
            {
                "name": "tcp-28800",
                "domainName": "tcp.a.business.qq.com",
                "path": "/local/path",
                "protocol": "tcp",
                "servicePort": 28800,
                "targetPort": 8080,
                "nodePort": 31001
            }
        ]
    }
}
```

## endpoints

如果存在endpoints，DNS直接watch并直接关联DNS解析记录。
如果没有，自行关联service和taskgroup信息，解析status信息提取IP。

```json
{
    "apiVersion":"v1",
    "kind":"endpoint",
    "metadata":{
        "name":"template-endpoint",
        "namespace":"defaultGroup",
        "label":{
            "io.tencent.bcs.cluster": "SET-SH-setname"
        }
    },
    "eps": [
        {
            "nodeIP": "127.0.0.1",
            "containerIP": "127.0.1.2"
        },{
            "nodeIP": "127.0.0.2",
            "containerIP": "127.0.1.1"
        }
    ]
}
```

## 特别字段说明

**clusterIP**：考虑未来需要做外部域名访问和服务发现，clusterIP暂留用于指向proxyIP信息

**ports**[x]说明：

* name：该name需要和taskgroup中ports字段中的name一致
* domainName：http协议的状态下，需要填写域名信息，做转发用途
* servicePort：主要用于负载均衡和服务导出端口
* targetPort：用于指向taskgroup中ports字段中目标端口，默认目标端口为containerPort，如果有hostPort，则指向hostPort

**BCSGROUP**: 用于service导出标识
**BCSBALANCE**：用于服务导出负载均衡算法，默认值为roundrobin
**BCS-WEIGHT-**: 当使用selector匹配多个application时，用于表达多个application之间的权重，该值为大于等于0的整数，类型为string。如果等于0，则该application没有流量导入。

## **bcs-loadbalance功能使用**

如果要启动loadbalance的功能，需要：

* label中增加特殊字段**"BCSGROUP":"external"**：external代表bcs-loadbalance模块的集群ID，默认值external
* label中增加特殊字段**"BCSBALANCE":"roundrobin"**：负载均衡算法，默认值为roundrobin，其他值为source（ip_hash），leastconn
* application有定义ports信息，并和service对应

ports信息在loadbalance中的含义说明：

* protocol：http或者tcp，当前不支持udp
* name：名字为application中定义port的名字，必须要对应
* domainName：协议为http时有效
* servicePort：如果协议是http，该值被忽略，loadbalance默认使用80端口；如果是tcp，loadbalance则监听该指定端口，各service之间该端口不能冲突

bcs-loadbalance可以工作两种环境下：

* overlay方式，默认使用servicePort作为服务端口，流量转发至containerPort
* underlay方式，使用servicePort作为服务端口，流量转发至hostPort（限于host/bridge模式）

## taskgroup 服务端口机制

服务端口数据来源于镜像字段ports，主要包含：

* containerPort：容器中使用端口
* servicePort：多实例状态使用的服务端口，用于服务导出和负载均衡使用
* hostPort：从容器映射到主机的端口
* hostIP：预留字段，当物理主机包含多个网卡时，默认是所有网卡都映射。如果0.0.0.0监听存在风险，可以监听内网
* protocol：协议，tcp/udp，http（默认转换为tcp处理，仅对datawatch/loadbalance生效）
* name：域名信息，http时必须为域名，必须唯一
* path: 路由转发，用于转发到后端对应的endpoint

填写端口信息，至少需要填写name、protocol、containerPort、servicePort信息。如果不需要实现端口映射，hostPort无需填写。

docker存在网络模式：None，Host，Bridge，User（CNM/CNI），端口映射对None（当前Executor暂未校验）

**hostPort**使用定义：-1为不使用，0为随机端口，正数为指定端口

针对docker容器端口映射实现，端口绑定有两个情况：

* **指定端口**绑定

下发参数中没有parameter参数-P，并直接指定hostPort，executor默认直接将该指定参数提交给docker进行绑定。端口冲突是否需要用户执行决断。
如果没有特别指定调度策略将实例分开，有极大几率造成端口冲突。

* Host模式下hostPort保持和containerPort一致，方便datawatch处理

* **随机端口**绑定

下发参数中存在-P，并且有ports字段，hostPort字段默认为0，scheduler使用offer上报的端口信息确认端口，更新taskgroup字段，并下发给Executor。
随机端口绑定可以**消除端口冲突**的问题。**有随机端口的情况下，为了方便应用获取到端口信息，需要将port端口根据需要导入环境变量PORT0 ~ PORTn**

**更新**（2017-05-22，根据LOL需求更新）：

* Host模式下:
  * containerPort设置为0，默认使用offer提供端口资源，hostPort和containerPort保持一致。port端口根据序号导入环境变量PORT0 ~ PORTn
  * containerPort指定端口，默认传递该端口，port端口根据序号导入环境变量PORT0 ~ PORTn
* Bridge模式
  * hostPort不填写默认补齐为-1，仅有containerPort
  * hostPort为0，默认随机端口，使用offer端口资源，port端口根据序号导入环境变量PORT0 ~ PORTn

### 服务端口代码调整说明

* executor

executor上报容器状态时，增加ports字段，结构包括name，hostPort，containerPort，protocol

* api/scheduler

使用随机端口时，默认hostPort直接填写0，使用offer里面端口资源更新hostPort信息，不使用HostPort，建议默认值为负数

* datawatch-mesos/datawatch-kube/loadbalance支持

上报ExportService时，结构调整，当前为

```go
//ExportService info to hold export service
type ExportService struct {
    Cluster     string       `json:"cluster"`     //cluster info
    Namespace   string       `json:"namespace"`   //namespace info, for business
    ServiceName string       `json:"serviceName"` //service name
    ServicePort []ExportPort `json:"ports"`       //export ports info
    BCSGroup    []string     `json:"BCSGroup"`    //service export group
    SSLCert     bool         `json:"sslcert"`     //SSL certificate for ser
    Balance     string       `json:"balance"`     //loadbalance algorithm, default source
    MaxConn     int          `json:"maxconn"`     //max connection setting
}
```

字段信息：

* SSLCert：是否使用https，当前默认不开启，忽略
* Balance：负载均衡的算法，默认是source，忽略；
* MaxConn：最大连接数，默认是20000

调整ServicePort和Backends字段，对其进行合并

```go
type ExportPort struct {
    BCSVHost    string    `json:"BCSVHost"`
    Protocol    string    `json:"protocol"`
    Path        string    `json:"path"`
    ServicePort int       `json:"servicePort"`
    Backends    []Backend `json:"backends"`
}

type Backend struct {
    TargetIP string `json:"targetIP"`
    TargetPort int   `json:"targetPort"`
}
```

datawatch构建ExportService时需要按照新结构进行。构建ExportService信息需要结合Service和TaskGroup信息。

* service中，多个ports对应多个ExportPort
* Service.ports[i].name == TaskGroup.container[i].ports[i].name
* 提取TaskGroup多个实例的containerIP，NodeIP，ContainerPort，HostPort
* backend组装说明，多个实例有多个backend
  * 如果TaskGroup网络模式host，取NodeIP和ContainerPort，组成backend
  * 网络模式为bridge，如果有HostPort：NodeIP + HostPort；如果没有，则ContainerIP + ContainerPort
  * 其他模式：ContainerIP + ContainerPort

