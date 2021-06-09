# 发布版本信息

## v1.20.11

* 发布日期：2021-04-28
* **功能优化**
  * 优化部分模块日志级别、日志标准错误输出可配置[#872]
  * bcs-storage支持自定义数据与k8s资源 labelselector[#852]
  * bcs-storage支持软删除[#852]
* **BUG修复**
  * bcs-storage修复数据mongo put接口的问题[#840]

## v1.20.9

* 发布日期：2021-04-08
* **新增功能**
  * bcs-webhook-server支持多DB同时授权[#654]
  * bcs-gamedeployment-operator在canaryUpdate过程中支持hook回调[#656]
  * bcs-gamedeployment-operator支持PreDeleteHook[#656]
  * bcs-clb-controller支持Mesos deployment 1对1端口映射规则[#669]
  * 新增模块bcs-hook-operator，用于支持自定义workload Pod生命周期间hook调用[#678]
  * bcs-gamestatefulset-operator支持PreDeleteHook和canaryUpdate[#680]
  * bcs-webhook-server新增imageloader插件，在自定义workload InplaceUpdate模式下实现镜像预热，缩短容器重启时间[#684]
  * bcs-logbeat-siedear支持单容器多套日志采集配置[#688]
  * bcs-logbeat-siedear支持自动采集容器挂载目录日志，支持宿主机日志采集[#689]
  * bcs-logbeat-siedear支持windows下的容器标准输出与挂载日志采集[#690]
  * bcs-ingress-controller支持namespace隔离特性[#702]
  * GameStatefulSet，GameDeployment增强参数校验[#709]
  * bcs-api-gateway支持apisix扩展[#713]
  * bcs-logbeat-sidecar支持采集器整合打包上报配置[#725]
  * 新增bcs-cluster-manager模块，管理集群信息，跨集群命名空间与集群tunnel链接[#738]
  * bcs-storage清理zookeeper服务发现，支持etcd服务发现[#739]
  * bcs-storage支持数据事件发送至消息对列[#742]
  * bcs-logbeat-sidecar支持bk-bcs-saas下发Pod selector大小写不敏感[#763]
  * bcs-hook-operator增加hookrun快速成功选项[#766]
  * bcs-kube-agent支持腾讯云TKE容器集群上报
  * bcs-cluster-manager兼容user-manager CIDR管理接口[#795]
  * bcs-storage mongodb接口支持聚合查询[#792]
  * bcs-storage 支持labelSelector查询[#852]
  * bcs-storage 支持自定义资源CRUD[#851]
* **功能优化**
  * bcs-storage重构：mongodb升级至v4，数据存储模型归并至相同的collection[#566]
  * bcs-webhook-server重构：定义hook插件接口支持hook自定义特性扩展[#674]
  * bcs-gamedeployment-operator hook功能重构，支持bcs-hook-operator回调状态互动[#679]
  * bcs-ingress-controller在status字段中显示clb vip信息[#699]
  * bcs-ingress-controller增加listener创建和失败事件和listener健康检查事件[#700]
  * bcs-webhook-server插件BcsDBPrivConfig CRD 添加operator字段[#715]
  * bk-bcs项目go mod依赖梳理[#722]
  * bcs-k8s-watch容器化版本配置调整[#728]
  * 优化common代码中消息对列日志过多问题[#756]
  * bcs-user-manager清理tunnel server特性[#770]
  * bcs-mesos-watch裁剪zookeeper服务发现bcs-storage特性
  * bcs-api-gateway apisix扩展下线/bcsapi/v1/接口支持
  * bcs-api-gateway kubernetes集群管理接口调整为/cluster/$clusterID
  * 腾讯云集群CIDR管理功能迁移至bcs-cluster-manager
  * bcs-storage优化消息队列日志输出日志[#787]
  * bcs-cluster-manager优化集群重复创建错误信息[#738]
  * bcs-api-gateway增加非标准模块metricservice请求转发[#818]
  * bcs-user-manager清理zookeeper服务发现接口[#713]
  * bcs-storage删除数据接口调整为可重入[#797]
  * bcs-cluster-manager保留kube-agent tunnel退出后credential信息[#789]
  * bcs-storage优化事件接口，删除事件可以正常返回数据[#840]
  * bcs-storage优化非订阅资源日志输出[#787]
* **BUG修复**
  * bcs-api-gateway apisix扩展插件修复lua模块判定错误[#713]
  * bcs-api-gateway修复网络中断导致leader elect逻辑异常问题[#817]
  * bcs-storage修复数据更新接口索引重复的问题[#840]
  * [Mesos] bcs-scheduler修复taskgroup脏数据问题[#664]
  * [Mesos] bcs-service-prometheus修复selector包含特殊字符导致退出问题[#671]
  * [Mesos] bcs-container-executor修复非私有地址导致容器无法调度的问题[#675]
  * bcs-dns，bcs-netservice修复非私有地址获取本地IP失败的问题[#675]
  * bcs-client修复update操作导致panic问题[#682]
  * bcs-ingress-controller修复listenerID缺失导致clb listener更新失败问题[#686]
  * bcs-logbeat-sidecar修复采集路径中存在软连接导致无法监听路径事件问题[#692]
  * bcs-clb-controller修复更新clb listener规则时缺失规则ID的问题[#694]
  * bcs-client修复cancel，pause，resume命令无法设置clusterid的问题[#696]
  * bcs-client修复deployment滚动更新时显示Application错误的问题[#697]
  * bcs-cloud-network-agent修复创建nodenetwork失败的问题[#708]
  * bcs-webhook-server修复不兼容kubernetes 1.12.6版本的问题[#712]
  * bcs-storage修复动态数据查询时空数据返回格式错误问题[#746]
  * bcs-netservice创建大量地址池超时问题[#759]
  * bcs-cloud-network-agent兼容nodenetwork定义[#774]
  * bcs-cluster-manager修复多实例kube-agent情况下单agent链接中断引起转发异常[#790]

## 1.18.12

* 发布日期：2020-12-30
* **BUG修复**
  * [Mesos] bcs-container-executor修复ip查询失败导致的panic[#675]

## 1.18.11

* 发布日期：2020-11-25
* **BUG修复**
  * [Mesos] bcs-scheduler修复taskgroup脏数据问题[#664]
  * [Mesos] bcs-service-prometheus修复selector异常导致退出问题[#671]
  * [Mesos] bcs-service-prometheus生成Node配置时产生死锁问题[#564]
  * [Mesos] bcs-service-prometheus清理label冲突问题[#644]
  * [Mesos] bcs-service-prometheus修复配置文件重复问题[#647]
  * bcs-qcloud-eip & bcs-cloudnetwork修复弹性网卡个数错误问题

## 1.18.10

* 发布日期：2020-11-05
* **功能优化**
  * bcs-api鉴权中心版本兼容导致flag redefine问题

## 1.18.9

* 发布日期：2020-10-30
* **功能优化**
  * bcs-user-manager不同步旧模式cluster credential数据[#641]

## 1.18.8

* 发布日期：2020-10-27
* **功能优化**
  * [Mesos] bcs-dns移入coredns 1.3.0项目中编译
  * [Mesos] bcs-scheduler修复taskgroup实例数超过200+卡顿问题[#559]
  * bcs-api-gateway, bcs-kube-watch容器化配置模板问题修复[#479]
  * bcs-api支持蓝鲸权限中心v3[#567]
  * bcs-k8s-driver支持1.16+版本API请求[#567]
  * bcs-api, bcs-user-manager修复tke cidr接口缺失问题[#535]
  * bcs-gamestatefulset-operator修复bitbucket代码引用问题
  * bcs-api兼容蓝鲸鉴权中心v2，v3版本
* **BUG修复**
  * bcs-api, bcs-user-manager修复websocket tunnel隧道长链接泄露问题[#584]
  * [Mesos] bcs-scheduler修复command health check失败问题[#586]
  * [Mesos] bcs-service-prometheus生成Node配置时产生死锁问题[#564]
  * [Mesos/K8S] 修复qcloud-eip多网卡bug问题[#556]
  * [Mesos] bcs-mesos-watch修复deployment消息堵塞消费过慢的问题[#552]
  * [Mesos] bcs-scheduler修复保存version latest报错问题[#525]
  * [Mesos] bcs-scheduler修复DaemonSet死锁问题[#546]
  * [Mesos] bcs-scheduler修复image secret缺失导致容器创建失败问题[#615]
  * [Mesos] bcs-scheduler修复metrics上报统计问题[#618]
  * [Mesos] 修复CPU因为ticker没有关闭导致过高问题[#478]

## 1.18.3

* 发布日期：2020-08-04
* **新增功能**
  * [Mesos]bcs-scheduler支持Daemonset特性[#207]
  * [Mesos]bcs-service-prometheus支持etcd存储模式[#473]
  * 新增模块bcs-bkcmdb-synchronizer，支持容器数据收录至蓝鲸cmdb[#476]
  * [K8S]新增模块bcs-cc-agent，为容器同步主机节点属性信息[#496]
  * [K8S/Mesos]bcs-cloudnetwork-agent支持腾讯云underlay方案初始化[#499]
  * [K8S/Mesos]开源bcs-egress-controller模块[#501]
  * [K8S]开源bk-cmdb-operator模块[#503]
  * [Mesos] Mesos方案支持prometheus ServiceMonitor[#514]
  * [K8S/Mesos] qcloud-eip插件支持在线扩容并兼容bridge模式[#515]
  * [K8S/Mesos] bcs-user-manager支持websocket tunnel模式实现集群托管[#521]
  * [Mesos] 支持机器异构场景下均匀调度容器[#25]
  * [Mesos] bcs-scheduler支持环境变量数据引用[#533]
* **功能优化**
  * [K8S/Mesos]容器日志采集方案支持采集路径模糊匹配，上报Label开关[#472]
  * 清理所有模块中对蓝鲸license服务依赖[#474]
  * [Mesos] bcs-scheduler支持mesos方案下容器corefile目录权限[#481]
  * [K8S/Mesos] bcs-loadbalance增强proxy转发规则冲突检测能力[#482]
  * [K8S/Mesos] bcs-datawatch优化同步netservice underlay资源超时问题[#483]
  * [Mesos] bcs-scheduler优化deepcopy导致CPU消耗过高问题[#485]
  * [K8S/Mesos]针对AWS网络方案中使用到的secret进行加密[#490]
  * [K8S]StatefulSetPlus更名为GameStatefulSet[#498]
  * 全模块容器化配置统一调整[#508]
  * [Mesos] 优化bcs-scheduler metrics数据[#532]
* **BUG修复**
  * 修复所有模块中ticker未关闭问题[#478]
  * [K8S]修复bcs-k8s-watch同步数据至bcs-storage数据不一致问题[#488]
  * [K8S/Mesos]修复AWS弹性网卡方案无法联通问题[#489]
  * [Mesos]修复bcs-mesos-adatper因为zookeeper异常时导致服务发现异常问题[#491]
  * [K8S/Mesos]修复bcs-loadbalance更新转发规则异常问题[#493]
  * [K8S/Mesos]修复bcs-webhook-server覆盖用户init-container的问题[#495]
  * [K8S]修复bcs-api因为权限问题无法使用kubectl exec与webconsole问题[#504]
  * [K8S]修复bcs-api websocket tunnel异常问题[#510]
  * [K8S/Mesos]修复腾讯云网络插件qcloud-eip与全局路由方案冲突问题[#515]
  * [Mesos] bcs-scheduler修复taskgroup因为自定义数据过大导致数据重复问题[#523]
  * [Mesos] bcs-scheduler修复etcd存储模式下taskgroup数据保存超时问题[#525]
  * [Mesos] bcs-scheduler优化scale up超时问题[#527]

## 1.17.5

* 发布日期：2020-07-06
* **新增功能**
  * [Mesos]bcs-scheduler支持污点与容忍性调度能力[#398]
  * [Mesos]bcs-mesos支持容器CPUSet绑定特性[#407]
  * [K8S/Mesos]bk-bcs开源分布式配置中心服务(bscp) [#443]
  * [K8S/Mesos]bcs-api以websocket支持跨云反向注册特性，支持跨云环境中以websocket实现反向通道注册[#412]
  * [K8S]bcs-k8s-driver支持websocket实现服务注册[#413]
  * [K8S]bcs-kube-agent支持websocket实现服务注册[#414]
  * [Mesos]bcs-mesos-driver支持websocket实现服务注册[#415]
  * [K8S/Mesos]新增bcs-networkpolicy模块支持K8S、Mesos网络策略[#417]
  * [K8S/Mesos]新增bcs-cpuset-device插件支持K8S、Mesos环境下CPU资源绑定调度[#424]
  * [K8S/Mesos]新增bcs-cloud-network支持腾讯云、AWS环境下CNI插件自动化安装与环境初始化[#426]
  * [K8S/Mesos]新增bcs-eni网络插件，支持腾讯云、AWS underlay方案[#426]
  * [K8S/Mesos]新增bcs-gateway-discovery模块支持bcs-api-gateway实现服务注册[#427]
  * [K8S/Mesos]使用kong重构bcs-api实现bcs服务网关[#427]
  * [K8S/Mesos]新增bcs-user-manager模块实现bcs集群与用户token鉴权[#433]
  * [Mesos]bcs-client依赖bcs-user-manager支持用户授权命令[#434]
  * [Mesos]bcs-client在Mesos环境下支持exec命令实现远程容器访问[#452]
  * [K8S/Mesos] bmsf-configuration配置服务支持自定义模板渲染[#454]
  * 分布式配置中心支持reload命令下发[#469]
* **功能优化**
  * [Mesos]bcs-scheduler优化对mesos version对象命名长度限制[#383]
  * [Mesos]bcs-container-executor针对Pod异常退出时保留镜像便于问题排查[#396]
  * [Mesos]bcs-container-executor针对Pod状态增加OOM状态[#397]
  * [Mesos]mesos-webconsole重构，通过bcs-mesos-driver实现console代理[#430]
  * [K8S/Mesos]bk-bcs日志采集方案重构，支持非webhook方案实现采集信息注入[#432]
  * [K8S]bcs-kube-agent支持bcs-api-gateway方式注册[#435]
  * [K8S/Mesos]bcs-user-manager支持token有效期限定刷新[#457]
  * [Mesos]bcs-scheduler etcd存储模式下优化对kube-apiserver限流问题[#462]
  * [Mesos]优化bcs-scheduler访问etcd ratelimiter[#462]
  * [K8S/Mesos]修复因为ticker没有关闭导致CPU过高问题[#478]
  * [Mesos]优化bcs-scheduler因为DeepCopy导致CPU过高问题[#485]
* **BUG修复**
  * [Mesos]bcs-scheduler修复容器退出时间过长时导致的事务性超时问题[#381]
  * [K8S/Mesos]bcs-webhook-server修复蓝鲸日志采集hook中环境变量错误覆盖问题[#400]
  * [Mesos]bcs-container-executor修复Pod中多容器情况下容器异常退出无法上报状态的问题[#406]
  * [K8S/Mesos]修复bcs-ipam插件回收IP资源时netns可能为空的问题[#437]
  * [K8S/Mesos]修复bcs-loadbalance针对后端转发状态判定异常问题[#441]
  * [K8S]bcs-api修复因为client-go缓存导致切换kube-apiserver引发异常问题[#445]
  * [Mesos]修复bcs-messos-watch同步bcs-netservice资源超时问题[#483]
  * [Mesos]修复bcs-messos-adapter服务发现异常问题[#491]
  * [K8S]修复bcs-api/bcs-kube-agent websocket tunnel模式下无法执行exec的问题[#504]
  * [K8S]修复bcs-api tunnel模式下服务发现问题[#510]
  * [K8S/Mesos]修复腾讯云网络插件qcloud-eip与全局路由方案冲突问题[#515]

## 1.16.4

* 发布日期：2020-05-08
* **BUG修复**
  * [Mesos] bcs-scheduler修复etcd模式下taskgroup存储失败问题[#436]

## 1.16.3

* 发布日期：2020-04-20
* **新增功能**
  * [Mesos] bcs-scheduler从1.17.x紧急合入支持Taints，Tolerations调度能力[#398]
  * [K8S] 新增statefulplus自定义workload[#346]
  * [K8S] bcs-k8s-watch支持CRD数据同步至storage[#363]
  * [K8S] bcs-kube-agent支持跨云网络代理功能[#376]
  * [K8S] bcs-kube-driver支持跨云网络代理功能[#378]
  * [K8S] bcs-kube-watch支持跨云向storage同步数据[#377]
  * [K8S] bcs-api支持通过外网访问bcs-kube-driver[#378]
  * [Mesos] 新增1.15.x版本mesos数据迁移工具[#359]
  * [Mesos] bcs-logbeat-sidecar支持自定义日志tag[#358]
  * [Mesos] bcs-client支持批量json/yaml形式资源批量处理命令apply/clean[#362]
  * [Mesos] bcs-api支持yaml格式Mesos资源创建[#362]
  * [Mesos/K8S] bcs-webhook-server支持bscp-sidecar注入[#366]
  * [Mesos] 新增基础网络连通性检测模块bcs-network-detection[#361, #391]
* **功能优化**
  * [Mesos] bcs-scheduler在etcd存储模式下过滤掉不符合规范label[#351]
  * [Mesos/K8S] bcs-webhook-server CRD version group调整[#374]
  * [Mesos/K8S] bcs-clb-controller基于腾讯云SDK限制优化CLB后端实例创建[#373]
  * [Mesos/K8S] bcs-webhook-server支持非标准日志标识注入[#385]
  * [Mesos/K8S] bcs-logbeat-sidecar支持单容器多种日志采集规则[#372]
  * [Mesos/K8S] 优化BCS服务发现公共组件[#384]
* **BUG修复**
  * [Mesos] bcs-webhook-server修复注入配置sidecar异常的问题[#366]
  * [Mesos] bcs-scheduler修复etcd存储模式下namespace,name长度异常问题[#383]

## 1.15.4

* 发布日期：2020-02-28
* **新增功能**
  * [Mesos] bcs-scheduler支持通过image名给与指定容器下发指令[#290]
  * [K8S/Mesos] datawatch支持同步集群underlay IP资源[#315]
  * [K8S/Mesos] BCS容器访问mysql自动授权方案[#308]
  * [K8S/Mesos] bcs-api支持storage数据存储事件监听接口[#315]
  * [Mesos] bcs-storage支持mesos process数据查阅[#195]
  * [Mesos] bcs-scheduler增加获取deployment、application定义接口[#332]
  * [K8S] bcs-k8s-driver支持1.14+ kubernetes版本[#336]
  * bcs-scheduler数据存储扩展etcd存储方式[#213]
  * bcs CNI网络插件qcloud-eip支持腾讯云VPC方案[#209]
  * bcs-dns特性扩展，支持mesos etcd数据存储方式[#251]
  * 增加k8s开源模块：bcs-k8s-csi-tencentcloud，使用CSI标准支持腾讯云CBS[#260,#261]
  * 开源bcs-clb-controller模块，容器自动集成腾讯云clb实现负载均衡[#247]
  * 重构并开源容器日志采集方案bcs-logbeat-sidecar[#259]
  * bcs mesos admissionwebhook支持webhook服务自定义URL[#279]
  * bcs-client增加CRD操作命令[#269]
  * bcs-scheduler支持mesos annotations中指定IP资源调度[#286]
  * bcs-scheduler支持Mesos类型自定义命令下发至指定容器[#290]
  * bcs-consoleproxy支持非UTF8格式交互[#282]
  * 开源bcs-log-webhook-server[#280]
  * bcs-hpacontroller支持mesos etcd数据存储方式[#253]
  * bcs-scheduler支持脚本方式检查容器健康[#248]
* **功能优化**
  * etcd存储优化label格式过滤[#351]
  * bcs-webhook-server开源与重构，支持多webhook能力[#295]
  * 优化蓝鲸日志采集方案，并开源配置刷新配置插件[#295]
  * bcs-loadbalance优化haproxy配置刷新方式[#310]
  * CNI社区插件整理[#325]
  * 优化bcs-dns日志输出方式[#236]
  * 优化bcs-netservice日志输出目录[#236]
  * 优化bcs-loadbalance无损刷新haproxy backend方式[#258]
  * 优化bcs-scheduler，admissionwebhook支持namespace selector特性[#264]
  * 优化bmsf-mesos-adapter，兼容application名字中包含下划线的情况[#270]
  * 优化bcs-mesos-driver，增加deployment/application名字合法性校验[#268]
  * 优化bcs-mesos-driver webhook机制，利用bcs-scheduler CRD格式提升性能[#267]
  * 优化bcs-service-prometheus，调整BCS各模块服务发现信息为独立配置文件[#277]
  * 利用etcd存储重构bcs mesos CRD机制[#269]
  * 调整metric-collector模板文件，调整资源使用限制[#287]
  * bcs-scheduler调整mesos service selector特性[#285]
  * bcs-container-executor优化无法识别网络容器退出问题[#245]
  * 优化流水线自动构建流程
* **bug修复**
  * 修复module-discovery服务发现lb节点的问题[#311]
  * 修复bcs-loadbalance haproxy状态显示与panic问题[#313, #320]
  * 修复bcs-scheduler在etcd存储下taskgroup数据不一致的问题[#327]
  * 修复bcs-scheduler在容器较长graceexit条件下状态变为Lost问题[#334]
  * 修复bcs-k8s-watch节点事件阻塞数据汇聚的问题[#220]
  * 修复基础库zkClient级联节点创建的问题[#227]
  * 修复bmsf-mesos-adapter因为Mesos集群未初始化导致退出的问题[#229]
  * 修复bcs-scheduler针对Application缺失的数据类型
  * 修复bcs-storage读取mongodb数据时context泄露问题
  * 修复bcs-netservice专用client存在网络链接泄露的问题[#273]

## 1.14.5

* 发布日期：2019-10-30
* **新增功能**
  * bcs-process-executor模块开源[#9]
  * bcs-process-daemon模块开源[#10]
  * bcs-dns增加prometheus metrics支持[#156]
  * bcs-loadbalance支持prometheus metrics[#161]
  * bcs-storage支持prometheus metrics，代码风格调整[#159]
  * bcs-api支持腾讯云TEK容器集群管理功能[#96]
  * bcs-scheduler支持prometheus metrics[#168]
  * bcs-sd-promethues支持bcs-loadbalance服务发现[#169]
  * bcs服务发现SDK支持bcs-loadbalance服务发现[#170]
  * bcs-api支持prometheus metrics[#172]
  * bcs mesos部分增加容器数据操作SDK[#115]
  * bcs-api支持管理腾讯云TKE容器集群[#96]
  * bcs-api增加网段分配存储用于统一云化资源管理[#134]
  * bcs-container-executor容器上报状态通道调整为自定义消息上报[#129]
  * bcs-mesos-datawatch、bcs-mesos-driver调整服务发现注册至集群层zookeeper[#136]
  * 新增bcs-services层、bcs集群层服务发现sdk[#137]
  * bcs-sd-prometheus模块开源：对接prometheus服务发现采集BCS信息[#138]
  * bcs-consoleproxy支持独立会话保持特性[#141]
  * bcs-netservice模块开源，并支持prometheus采集方案[#86]
  * bcs-mesos-datawatch下线自定义healthcheck机制，支持prometheus采集方案[#145]
  * bcs-mesos-datawatch支持跨云部署[#175]
  * bcs-api针对Mesos集群支持跨云请求转发[#175]
  * bcs-storage支持跨云服务发现[#175]
  * bcs-mesos-driver支持跨云服务注册[#175]
  * bmsf-mesh-adaptor模块开源[#177]
  * bcs-mesos-executor支持prometheus text collector[#178]
  * bcs-k8s-ip-scheduler模块开源[#184]
  * bcs-loadbalance支持prometheus metrics[#161]
  * bcs-scheduler增加全量CRD数据读取接口[#198]
  * bcs-api增加对TKE容器网段管理功能[#202]
  * bcs-hpacontroller模块开源[#181]
* **功能优化**
  * bcs-loadbalance haproxy metrics重构prometheus metrics采集方式[#162]
  * bcs-loadbalance镜像调整，优化启动脚本[#162]
  * bcs-loadbalance服务注册同时支持集群层与服务层zookeeper[#164]
  * 更新bcs-mesos prometheus方案文档
  * bcs-mesos-datawatch代码复杂度优化[#71]、[#72]
  * bcs-api代码复杂度、注释优化[#144]
  * metrics采集方案文档更新
  * bcs-mesos-datawatch优化zookeeper服务发现策略[#183]
  * bcs-k8s-datawatch优化容器监控告警细则[#192]
  * bcs-scheduler优化taskgroup因资源不足无法调度提示语[#103]
  * bcs-storage优化metrics数据量[#185]
* **bug修复**
  * 修复bcs-api CIDR分配时锁泄露问题[#134]
  * 修复bcs-container-executor部分情况下dockerd异常退出panic的情况[#130]
  * 修复bcs-scheduler启动metrics时panic的问题[#189]
  * 修复bcs-storage读写mongodb cursor泄露问题[#193]
  * 修复bcs-netservice日志输出无法自定义问题[#236]
  * 修复bcs-dns日志输出无法自定义问题[#236]
  * 修复bcs-mesos-adapter节点删除事件丢失问题[#237]

## 1.13.4

* 发布日期：2019-07-26

* **功能优化**
  * bcs-container-executor调整与meos-slave长链接超时时间[#82]

## 1.13.3

* 发布日期：2019-07-12
* 版本信息：1.13.3

* **新增功能**
  * bcs-mesos支持系统常量注入[#19]
  * bcs-mesos调度状态优化，调整LOST状态处理[#26]
  * bcs-mesos支持资源强制限制特性[#27]
  * bcs-mesos调度过程调整，允许更新状态下手动调度容器[#29]
  * bcs-storage扩展自定义额外查询条件[#34]
  * bcs-metricscollector迁移模块[#4]
  * bcs-metricsserver迁移模块[#7]
  * 工具scripts增加go vet支持[#65]
  * bcs-client增加--all-namespace参数支持[#66]
* **功能优化**
  * 首页产品文档优化[#83]
  * BCS全量代码go vet调整[#70]
  * bcs-mesos容器异常超时调度调整[#24]
  * bcs-api日志调整[#32]
* **bug修复**
  * bcs-container-executor修复CNI异常调用错误显示问题[#88]
  * Makefile修复非Linux环境编译错误问题[#57]
  * bcs-container-executor修复启动阶段panic问题[#23]
  * Makefile修复sirupsen依赖问题
  * bcs-mesos修复容器LOST状态异常问题[#23]
  * bcs-mesos修复并发状态容器自定义命令执行结果丢失问题[#30]
  * bcs-mesos修复application调度异常问题与日志[#17] [#14]
  * bcs-mesos修复取消滚动升级超时问题[#42]

## 1.12.6

>以下issue索引信息是并非来源github，为保证项目内外一致性，暂不清理

* 发布日期：2019-04-30
* 版本信息：1.12.6

* **新增功能**
  * bcs-container-executor支持CNI路径、网络镜像配置化
  * bcs-health支持告警信息转存bcs-storage
  * bcs-mesos支持AutoScaling特性[#10]
  * bcs-scheduler针对IP插件支持独立tls证书
  * bcs-scheduler支持healthcheck多次连续失败后进行重新调度[#31]
  * bcs-scheduler对调度插件支持自定义目录[#50]
  * bcs-scheduler新增Node节点资源排序功能，均衡节点容器分布[#80]
  * bcs-loadbalance新增开源版本dockerfile[#65]
  * bcs-client支持Get命令，获取资源定义文件[#73]
  * bcs-client支持https方式链接bcs-api[#78]
  * bcs-mesos-driver支持web-hook特性[#68]
* **功能优化**
  * 进程启动参数增加--config_file，兼容--file参数[#52]
  * LICENSE文件更新，修正复制glog代码中的copyright[#72]
  * bcs-kube-agent链接bcs-api时支持insecureSkipVerify[#75]
  * bcs-data-watch优化exportservice数据同步，提升数据同步效率[#79]
  * bcs-api配置项json化[#52]
  * bcs-scheduler、bcs-mesos-watch清理appsvc无用代码
  * bcs-scheduler容器调度日志优化
  * bcs-mesos-watch清理已注释代码
  * bcs-scheduler代码清理
  * bcs-loadbalance调整tls证书目录，并支持tls命令行参数
  * bcs-loadbalance镜像中nginx用户调整为bcs[#61]
  * bcs-mesos-driver清理v1http无用代码
  * bcs-consoleproxy以及与bcs-webconsole代码重构
  * k8s文档优化[#46]
  * bcs-executor优化healthcheck上报数据[#30]
  * bcs-scheduler优化滚动更新时healthcheck机制[#55]
  * 文档完善，增加k8s和Mesos资源分类和功能[#63]
  * bcs-client重构，并移除ippool命令[#66]
  * 清理bcs-scheduler ingress数据定义文档[#86]
  * bcs-api增加用户token类型校验，用于开源使用[#53]
  * bcs-kube-agent目录调整[#2],[#4]
  * 全项目代码复杂度优化
  * 全项目重复代码优化
* **bug修复**
  * 修复bcs-health中因zk acl错误而不断刷日志的问题[#83]
  * 修复bcs-api zookeeper断链后无法发现后端集群的异常[#56]
  * 修复bcs-api针对后端集群事件发生错误时导致的panic[#60]

## 1.11.11

>以下issue索引信息是并非来源github，为保证项目内外一致性，暂不清理

* 发布日期：2019-02-21
* 版本信息：1.11.11

merge截止: !30

* **新增功能**
  * 对容器Label增加namespace/pod_name信息[#18]
  * bcs-api与PAAS/PAAS-Auth解耦[#21]
  * bcs-exporter插件化与标准化[#15]
  * 内部版本与企业版本PAAS-Auth支持[#26]
  * bcs-health的数据流出口规范化[#14]
  * 新增模块bcs-consoleproxy[#28]，并支持https[#32]
  * mesos支持command命令[#6]
  * bcs-api支持websocket反向代理[#33]
  * bcs-api rbac功能增加开关[#34]
* **功能优化**
  * bcs-container-executor支持标准CNI链式调用[#2]
  * 采用go dep裁剪vendor目录[!63]
  * bcs-dns自定义注册插件bcscustom支持多IP地址注册[#9]
  * 代码中敏感信息清理[#20]
  * bcs-api文档补充[#22]
  * 优化与丰富bcs单元测试[#13]
* **bug修复**
  * 修复common.RegisterDiscover Session失效后zookeeper事件无法触发bug[#1]
  * bcs-scheduler修复主机与IP资源精确调度时资源不足的问题[#3]
  * 调整blog中glog的init行为，修复glog的初始化问题[#12]
  * kubernete client-go升级v9.0.0导致配置字段异常问题[#16]
  * kubernete升级1.12.3后，health check出现tls handshake错误问题[#17]
  * bcs-api修复服务发现时可能产生的panic[#23]
  * 修复templates配置文件缺失，将api配置文件命名与其他组件统一[#27]
  * bcs-dns的启动脚本中去除--log，corefile中去除dnslog配置[#38]
  * k8s metric的api路径错误修复[#37]
  * 修复bcs-api进行healthcheck时出现的panic[#48]
