# 发布版本信息

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


