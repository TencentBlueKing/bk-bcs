![BK-BSCP logo](./docs/img/logo.png)

BK-BSCP 蓝鲸服务配置平台
========================

[TOC]

## What's It ?

蓝鲸服务配置平台(BlueKing Service Configuration Platform), 是蓝鲸体系内的针对微服务的配置原子平台。
结合蓝鲸容器服务(BCS)、蓝鲸管控平台(GSE)、蓝盾流水线等服务, 提供微服务配置的版本管理、版本发布和内容渲染等功能。

## Overview

BK-BSCP 蓝鲸服务配置平台

* 配置热更新
* 配置版本管理
* 定制化灰度发布策略
* 配置内容模板渲染
* 支持文本、二进制文件
* 支持业务集群管理
* 支持容器与非容器环境
* 蓝盾流水线插件集成

## Getting Started

BK-BSCP提供命令行客户端和蓝盾流水线等多种使用方式，可根据需求选择。

### 命令行方式

*环境变量配置*

```shell
export BSCP_BUSINESS=X-Game
export BSCP_OPERATOR=MrMGXXXX
```

*4条命令即可完成配置发布*

示例，对GameServer应用模块下的server.yaml配置文件进行版本发布操作:

step1 创建提交(Create Commit)：
```shell
bk-bscp-client create commit --app gameserver --cfgset server.yaml --config-file ./new-release-server.yaml
Create Commit successfully: M-2ef13220-d142-11e9-a11f-5254000ea971
```

step2 确认提交(Confirm Commit):
```shell
bk-bscp-client confirm commit --Id M-2ef13220-d142-11e9-a11f-5254000ea971
Confirm Commit successfully: M-2ef13220-d142-11e9-a11f-5254000ea971
```

step3 创建版本(Create Release):
```shell
bk-bscp-client create release --app gameserver --commitId M-2ef13220-d142-11e9-a11f-5254000ea971 --name new-release --strategy blue-strategy
Create Release successfully: R-85a74cc6-d14b-11e9-a11f-5254000ea971
```

step4 发布(Publish Release):
```shell
bk-bscp-client confirm release --Id R-85a74cc6-d14b-11e9-a11f-5254000ea971
release R-85a74cc6-d14b-11e9-a11f-5254000ea971 confirm to publish successfully
```

## Documents

[客户端手册](docs/client_book.md)
[系统对接指南](api/api.md)
[逻辑集成功能介绍](docs/integrator.md)
[模板渲染介绍](docs/template.md)
[配置发布策略](docs/strategy.md)
[设计文档](docs/arch.md)
[项目规范](docs/standard.md)
[部署安装教程](docs/install.md)
