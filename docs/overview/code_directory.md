# 代码结构

## bcs-common

bk-bcs公共代码

## bcs-services

服务层服务
* bcs-api: 对外接入服务，提供鉴权和路由分发等功能
* bcs-storage: 存储服务，提供数据存储和查询功能
* bcs-dns: 域名解析
* bcs-health: 监控告警
* bcs-client：客户端

## bcs-mesos

mesos集群层服务
* bcs-mesos-driver： mesos集群接入服务
* bcs-scheduler: mesos集群调度服务
* bcs-mesos-watch: mesos集群动态数据监测和同步服务
* bcs-check： 健康检查
* bcs-container-executor: 容器执行器
* bcs-process-executor: 进程执行器
* bcs-process-daemon: 进程管理和守护

## bcs-k8s

k8s集群层服务
* bcs-kube-agent：集群代理，负责将集群向BCS API注册
* bcs-k8s-watch： 负责将集群数据向BCS storage同步
