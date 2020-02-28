# 发布版本信息

## 1.15.4

- 发布日期：2020-02-28
- **新增功能**
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
- **功能优化**
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
- **bug修复**
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

- 发布日期：2019-10-30
- **新增功能**
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
- **功能优化**
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
- **bug修复**
  * 修复bcs-api CIDR分配时锁泄露问题[#134]
  * 修复bcs-container-executor部分情况下dockerd异常退出panic的情况[#130]
  * 修复bcs-scheduler启动metrics时panic的问题[#189]
  * 修复bcs-storage读写mongodb cursor泄露问题[#193]
  * 修复bcs-netservice日志输出无法自定义问题[#236]
  * 修复bcs-dns日志输出无法自定义问题[#236]
  * 修复bcs-mesos-adapter节点删除事件丢失问题[#237]

## 1.13.4

- 发布日期：2019-07-26

- **功能优化**
  * bcs-container-executor调整与meos-slave长链接超时时间[#82]

## 1.13.3

- 发布日期：2019-07-12
- 版本信息：1.13.3

- **新增功能**
  * bcs-mesos支持系统常量注入[#19]
  * bcs-mesos调度状态优化，调整LOST状态处理[#26]
  * bcs-mesos支持资源强制限制特性[#27]
  * bcs-mesos调度过程调整，允许更新状态下手动调度容器[#29]
  * bcs-storage扩展自定义额外查询条件[#34]
  * bcs-metricscollector迁移模块[#4]
  * bcs-metricsserver迁移模块[#7]
  * 工具scripts增加go vet支持[#65]
  * bcs-client增加--all-namespace参数支持[#66] 

- **功能优化**
  * 首页产品文档优化[#83]
  * BCS全量代码go vet调整[#70]
  * bcs-mesos容器异常超时调度调整[#24]
  * bcs-api日志调整[#32]

- **bug修复**
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

- 发布日期：2019-04-30
- 版本信息：1.12.6

- **新增功能**
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
   
- **功能优化**
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
  
- **bug修复**
  * 修复bcs-health中因zk acl错误而不断刷日志的问题[#83]
  * 修复bcs-api zookeeper断链后无法发现后端集群的异常[#56]
  * 修复bcs-api针对后端集群事件发生错误时导致的panic[#60]

## 1.11.11

>以下issue索引信息是并非来源github，为保证项目内外一致性，暂不清理

- 发布日期：2019-02-21
- 版本信息：1.11.11

merge截止: !30

- **新增功能**
  - 对容器Label增加namespace/pod_name信息[#18]
  - bcs-api与PAAS/PAAS-Auth解耦[#21]
  - bcs-exporter插件化与标准化[#15]
  - 内部版本与企业版本PAAS-Auth支持[#26]
  - bcs-health的数据流出口规范化[#14]
  - 新增模块bcs-consoleproxy[#28]，并支持https[#32]
  - mesos支持command命令[#6]
  - bcs-api支持websocket反向代理[#33]
  - bcs-api rbac功能增加开关[#34]

- **功能优化**
  - bcs-container-executor支持标准CNI链式调用[#2]
  - 采用go dep裁剪vendor目录[!63]
  - bcs-dns自定义注册插件bcscustom支持多IP地址注册[#9]
  - 代码中敏感信息清理[#20]
  - bcs-api文档补充[#22]
  - 优化与丰富bcs单元测试[#13]

- **bug修复**
  - 修复common.RegisterDiscover Session失效后zookeeper事件无法触发bug[#1]
  - bcs-scheduler修复主机与IP资源精确调度时资源不足的问题[#3]
  - 调整blog中glog的init行为，修复glog的初始化问题[#12]
  - kubernete client-go升级v9.0.0导致配置字段异常问题[#16]
  - kubernete升级1.12.3后，health check出现tls handshake错误问题[#17]
  - bcs-api修复服务发现时可能产生的panic[#23]
  - 修复templates配置文件缺失，将api配置文件命名与其他组件统一[#27]
  - bcs-dns的启动脚本中去除--log，corefile中去除dnslog配置[#38]
  - k8s metric的api路径错误修复[#37]
  - 修复bcs-api进行healthcheck时出现的panic[#48]


