# BCS高可用Mesos集群部署

部署BCS管理的Mesos高可用集群有2种方式：

- **蓝鲸社区版(BCS增值包)**：使用[容器服务控制台](https://docs.bk.tencent.com/bcs/Container/QuickStart.html)一键创建Mesos集群，Master节点推荐3或5台，集群创建成功后，可以进入集群节点列表，为集群增加节点

- **手动部署**： 参考本文档指引部署BCS高可用Mesos集群

## 目标机器环境准备

### 硬件

| 资源      | 配置    |
| -------- | ------- |
| CPU      | 4核     |
| Mem      | 8GB     |
| Disk     | >= 50GB |

>注：Slave节点可根据业务规模增加机器配置 

### 操作系统

CentOS 7及以上系统（内核版本3.10.0-693及以上），推荐CentOS 7.4

> 注：集群网络模式使用Overlay网络模式，需要有NAT模块(iptable_nat)

### 网络

- 放行集群层到服务层的机器所有网络策略
- 放行集群层机器间的所有网络策略

<!-- To-Do: 严格的网络策略 -->

## 组件说明

- 开源组件依赖
  - Master层
    - zookeeper(>=3.4.8): 集群信息存储
    - mesos(>=1.1.0)
    - etcd(推荐>=3.3.10): Overlay网络模式使用，为flannel维护网络的分配情况
  - Node层
    - docker(推荐使用18.09.X)
    - flannel: 通过给每台宿主机分配一个子网的方式为容器提供虚拟网络，它基于Linux TUN/TAP，使用UDP封装IP包来创建Overlay网络
- BCS组件
  - Master层
    - bcs-mesos-driver
    - bcs-mesos-watch
    - bcs-scheduler
    - bcs-check
    - bcs-dns
    - bcs-health-slave
  - Node层
    - bcs-container-executor

## 部署Mesos集群

高可用部署思路：确保存储组件zookeeper、etcd高可用，Mesos Master高可用，Master层的BCS组件部署2台以上

- zookeeper高可用[部署指引](http://zookeeper.apache.org/doc/r3.4.10/zookeeperStarted.html)
- etcd高可用[部署指引](https://etcd.io/docs/v3.3.12/op-guide/clustering/)
- mesos master高可用[部署指引](http://mesos.apache.org/documentation/latest/high-availability/)
- 部署BCS组件参考文档[bcs项目单机部署流程(mesos)](./mesos-deploy-in-single-guide.md): Master层的BCS组件部署2台以上
