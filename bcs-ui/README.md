# 蓝鲸容器管理平台SaaS产品层主体功能模块

## 简介

容器管理平台SaaS产品层提供友好的BCS的操作界面，支持对项目集群、节点、命名空间、部署配置、仓库镜像、应用等进行可视化界面操作管理。

## 主体功能

- 集群、节点初始化和管理
- 配置模板集管理
- 网络、资源、应用等可视化操作管理
- WebConsole
- Helm

## 技术栈

- python
- django
- tornado
- nodejs/vue
- mysql
- redis

## 依赖说明
- bk-iam: 蓝鲸权限中心
- bcs-projmgr: 容器管理平台项目管理模块
- bcs-cc: 容器管理平台配置中心模块
- mysql: 容器管理平台数据库