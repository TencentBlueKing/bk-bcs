# 发布版本信息

## v1.11.11

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


