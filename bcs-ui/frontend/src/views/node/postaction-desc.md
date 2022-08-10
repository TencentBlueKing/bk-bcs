#### 前置初始化

1. ##### 前置初始化在节点添加流程中执行位置

   提交添加节点任务 ---> 节点重装操作系统 ---> **执行前置初始化** ---> 部署kubelet服务 --- >执行后置初始化 ---> 添加节点完成

2. ##### 使用场景

   - DaemonSet进程启动前依赖Woker节点上的特殊目录，需要提前创建
   - DaemonSet进程需要特殊内核参数，进程启动前需要先在Woker节点上配置
   - DaemonSet以hostnetwork模式启动，继承Woker节点的/etc/resolv.conf，且需要配置特殊nameserver
   - 业务其它特殊配置

3. ##### 使用示例

   - Woker节点上创建业务目录

     ```bash
     mkdir /data/logs
     ```

   - 调整Woker节点内核参数

     ```bash
     echo 'net.core.rmem_default = 52428800
     
     net.core.rmem_max = 52428800
     
     net.core.wmem_default = 52428800
     
     net.core.wmem_max = 52428800' >>/etc/sysctl.conf
     
     sysctl -p
     ```

   - 修改节点主机的/etc/resolv.conf

     ```bash
     mkdir -p /data/backup
     cp -a /etc/resolv.conf /data/backup/resolv.conf_$(date +%Y%m%d%H%M%S)
     echo 'nameserver 8.8.8.8
     nameserver 114.114.114.114' >/etc/resolv.conf
     ```
#### 后置初始化

   1. ##### 后置初始化在节点添加流程中执行位置

       提交添加节点任务 ---> 节点重装操作系统 --->执行前置初始化 ---> 部署kubelet服务 --- > **执行后置初始化** ---> 添加节点完成

2. ##### 使用场景

   - 简易脚本执行

     与前置初始化一样，适用于简单初始化场景，逻辑建议不要过于复杂
   
   - 标准运维流程执行
   
     使用蓝鲸标准运维应用作为执行引擎，适用于复杂的初始化场景
