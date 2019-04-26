# 蓝鲸容器管理平台，双引擎驱动的容器编排方案

![head](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/opensource.png)

## 导读

蓝鲸容器管理平台(BCS，Blueking Container Service)是高度可扩展、灵活易用的容器管理服务。蓝鲸容器服务支持两种
不同的容器编排方案，分别为原生Kubernetes模式和基于mesos自研的模式。使用该服务，用户无需关注基础设施的安装、运维
和管理，只需要调用简单的API或者client，便可对容器进行启动、停止等操作，查看集群、容器及服务的状态，以及使用各种组
件服务。用户可以依据自身的需要选择集群模式和容器编排的方式，以满足业务的特定要求。

![整体架构](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/bcs-modules.jpg)

## 功能特性

![重要功能列表](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/functions-list.png)

* 支持基于k8s和Mesos双引擎编排
* 支持多集群管理
* 支持插件化自定义编排调度策略
* 支持服务升级，扩缩容，滚动升级，蓝绿发布等
* 支持configmap，secret，磁盘卷挂载，共享盘挂载等
* 支持服务发现，域名解析，访问代理等基础服务治理方案
* 支持可扩展的资源配额定义
* 支持容器内IPC机制
* 支持多种容器网络方案
* 支持跨云管理

## 价值

蓝鲸容器管理平台可以独立部署，通过API或者命令行工具的方式对多个容器集群进行管理，实现容器服务的各项操作。并内置了
跨集群的名字服务，应用级的负载均衡，远程metrics采集，多集群容器数据聚合管理等，满足一般容器管理的需求。针对当下
比较便利的Service Mesh方案istio，平台也进行了适配，不管还是kubernetes还是自研Mesos调度框架，都可以无缝集成。

考虑对于陈旧业务微服务化，容器化的难度，我们在实践过程中，在自研的Mesos framework中还加入进程调度，可以最大限度
帮助业务从进程部署，平滑切换到容器-进程混合部署，最后全容器微服务化部署。

同时，蓝鲸容器管理平台也可以很容易与蓝鲸其他开源的平台进行集成，例如蓝鲸PAAS平台，蓝鲸数据平台，集成蓝鲸PAAS平台，
可以直接使用蓝鲸原生的容器管理UI，易于完成对大量的Deployment，application，service文件管理，版本管理等。集成
蓝鲸数据平台，更好完成日志采集、性能采集、metrics采集、容器监控，实现数据过滤清洗与数据呈现。相关细节可以参阅
[这里](https://bk.tencent.com/product/)。

## 未来

BCS团队对容器管理平台进行开源，希望将我们的技术和沉淀反馈给社区，期望能帮助更多的人解决问题；同时也邀请容器技术
爱好者一起参与建设，让产品变更更加强大和易用，构建生态活跃的技术社区。

![体验指引](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/guids.png)

![如何体验](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/BCS-exp.png)

![版本参与](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/workflows.png)

## 开源协议

蓝鲸容器管理平台采用的是MIT开源协议。MIT是和BSD一样宽范的许可协议，BCS团队只想保留版权，而无任何其他的限制。
也就是说，你必须在你的发行版里包含原许可协议的声明，无论你是以二进制发布的还是以源代码发布的。


## 欢迎交流

![主页](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/homelink.png)

![交流渠道](https://github.com/Tencent/bk-bcs/blob/master/docs/wiki/QR-Code.png)
