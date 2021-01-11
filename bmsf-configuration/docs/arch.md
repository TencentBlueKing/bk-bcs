BK-BSCP 平台架构设计
==========================

[TOC]

# 平台架构

![avatar](./img/platform.png)

# 集成服务设计

## 模块

![avatar](./img/middle-arch.png)

* APIServer: 接入服务，负责协议转换、URL路由、权限过滤、制品库服务代理；
* AuthServer: 权限服务，负责完成内部权限审核;
* Other Auxiliary Modules: 其他辅助服务，如对账、补单、定时任务等辅助模块;

# 原子服务设计

## 基础概念

![avatar](./img/objects.png)

* Business: 业务划分, 系统内部不创建业务资源，只做外部业务的关联管理;
* App: 业务之下的具体应用模块, 为系统中的最小管理单元，外部系统也是以此进行关联对接;
* Config: 配置, 可理解为以往的单一配置文件;

## 模块

![avatar](./img/atomic-arch.png)

* ConfigServer: 配置服务, 提供gRPC协议服务, 负责原子接口逻辑和较复杂的逻辑集成；
* TemplateServer: 模板服务，负责配置模板的管理和内容渲染；
* DataManager: 数据代理服务, 提供统一的缓存、DB分片存储能力;
* BCS-Controller: BCS容器环境的控制器，负责策略控制和版本下发；
* GSE-Controller: GSE侧的控制器，负责策略控制和版本下发；
* ConnServer: BCS容器环境会话链接服务，与BCS-Sidecar配合维护下发通道;
* TunnelServer: GSE侧通道服务，负责GSE通道的数据下行；
* BCS-Sidecar: BCS容器环境sidecar，以sidecar模式运行，完成配置版本的拉取、生效和反馈上报；
* GSE-plugin: GSE插件，与GSE Agent配合构建进程、容器混合环境下的配置通道;

## 数据流
> 配置发布的主要逻辑, 其他复杂逻辑不做展示说明

![avatar](./img/logic.png)

# Q&A
