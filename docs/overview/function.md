# 功能说明

## 资源定义

mesos方案对于微服务的抽象主要定义了application，configmap，deployment等资源抽象。业务可以使用这些资源来达到描述服务的效果，资源文件的详细描述如下：

* [deployment](../templates/mesos-artifact/deployment.md)
* [application](../templates/mesos-artifact/application.md)
* [service](../templates/mesos-artifact/service.md)
* [configmap](../templates/mesos-artifact/configmap.md)
* [secret](../templates/mesos-artifact/secret.md)
* [process](../templates/mesos-artifact/process.md)

k8s方案使用了原生的k8s资源抽象，包括：ReplicaSet, Deployments, StatefulSets等。业务使用上述资源来实现抽象服务的效果，资源文件的详情请参考k8s官方文档。

## 功能详情
### POD
- k8s方案：POD
- mesos方案：Taskgroup

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- Pod是所有业务类型的基础，它是一个或多个容器的组合，例如：业务容器，sidecar容器
- 这些容器共享存储、网络和命名空间，以及如何运行的规范
- 在Pod中，所有容器都被统一编排和调度

### 无状态应用
- k8s方案：ReplicaSet/Deployment
- mesos方案：Application/Deployment

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 支持设定编排策略（Restart、Kill等）；
- 支持设定容器启动参数；
- 支持设定镜像版本及加载方式；
- 支持设定端口映射；
- 支持设定环境变量；
- 支持设定label、备注；
- 支持设定卷、configmap、secret的关联关系；
- 支持设定网络；
- 支持设定资源配额设定；
- 支持设定健康检查机制；

**差异点**

- K8S支持设定生命周期、就绪检查；
- K8S支持更多的存储driver，并且从1.9版本开始支持CSI；
- 调度策略配置不同：
  - K8S：基于label的调度策略，通过selector进行筛选
  - mesos：基于变量运算符的调度策略，支持获取CC的属性作为调度依据
- 自定义调度支持方式不同：
  - K8S：支持自定义controller
  - mesos：支持自定义调度插件
- mesos方案支持POD固定IP调度
- mesos方案支持bridge模式下端口随机

### 有状态应用
- k8s方案：StatefulSets
- mesos方案：Application

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- POD拥有独立唯一不变的ID；
- POD拥有独立唯一的域名；
- 挂载独立的数据卷；

### 任务
- k8s方案：Job/CronJob
- mesos方案：Application

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 短任务单次执行，例如数据库初始化任务；

**差异点**

- mesos方案无专门workload，通过设置Application的参数实现
- K8S支持部署定时任务

### DaemonSet
- k8s方案：DaemonSet
- mesos方案：Application

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 按主机1：1动态部署容器；
- 能动态跟随物理机资源缩扩容；

**差异点**

- mesos方案无专门Workload，需要通过设置Application的调度算法实现相同功能；

### 部署方案
- k8s方案：Deployment
- mesos方案：Deployment

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- recreate操作方式；

**差异点**

- rollingUpdate：
  - K8S方案支持设置滚动方式，最大不可用数和最大更新数
  - mesos方案支持设置滚动方式，周期，频率，和手动模式

### 服务
- k8s方案：Service
- mesos方案：Service

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 通过自定义域名描述外部服务；
- 对RS或者Application的服务做抽象；
- 支持容器故障或者扩缩容时Backends的动态刷新；
- 支持简单的负载均衡策略；
- 支持http、tcp、udp等主流通讯协议；

**差异点**

- mesos方案支持通过location对不同域名进行分流
- mesos方案支持对两个Application之间设置流量权重
- K8S支持更多对外暴露服务折射方式，除ClusterIP，LB外，还是支持NodePort

### Ingress
- k8s方案：Ingress
- mesos方案：Ingress

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 实现集群外部对集群内服务的访问；
- 支持http、tcp；
- 支持轮询的负载均衡方式；

**差异点**

- mesos方案支持流量权重设置；
- mesos方案支持Balance Source；

### 配置文件
- k8s方案：Configmap
- mesos方案：Configmap

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- Configmap内容落地为文件或者环境变量；

**差异点**

- K8S方案支持在POD command中以变量方式引用；
- mesos方案支持远端配置，包括文本和二进制配置文件

### 加密配置
- k8s方案：Secrets
- mesos方案：Secrets

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 对敏感信息加密；

### 挂载卷
- k8s方案：Volumes
- mesos方案：Volumes

##### 主要功能及差异说明
**k8s方案，mesos方案均支持**

- 支持Empty DIR；
- 支持Host Path；
- 支持挂载Ceph、nfs等远端目录；
- 支持subPath；

## 操作说明

mesos方案提供了bcs-client客户端，支持对于application，configmap，deployment等资源文件的创建，更新，删除等操作，具体的操作手册请参考：
[mesos资源操作手册](../features/bcs-client/bcs-client_HANDBOOK.md)

k8s方案支持使用默认的kubectl客户端，实现对资源文件的操作，具体的操作请参考k8s官方文档。