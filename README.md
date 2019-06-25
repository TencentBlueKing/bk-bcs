![蓝鲸容器管理平台.png](./docs/logo/bcs_zh.png)
---
[![license](https://img.shields.io/badge/license-mit-brightgreen.svg?style=flat)](https://github.com/Tencent/bk-bcs/blob/master/LICENSE)[![Release Version](https://img.shields.io/badge/release-1.12.x-brightgreen.svg)](https://github.com/Tencent/bk-bcs/releases) ![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/bcs/p-95397dbecda4442795dd0125a33069cb/badge?X-DEVOPS-PROJECT-ID=bcs) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/Tencent/bk-bcs/pulls)                                                                                                                                                     
蓝鲸容器管理平台（Blueking Container Service，简称BCS）是蓝鲸体系下，以容器技术为基础，为微服务业务提供编排管理的基础服务平台。

蓝鲸容器管理平台提供了基于原生k8s和mesos+bk-framework的双引擎驱动的容器编排方案，用户可以选择使用其中一种来编排自己的应用。其中k8s方式以社区方案为主，除提供原生功能支持外，还实现了原生k8s与蓝鲸体系的无缝结合，用户可以在蓝鲸体系下以与传统无差异的方式体验容器技术和k8s社区带来的便利。mesos+bk-framwork方案是蓝鲸为需要深度定制的用户准备的可进行二次开发的容器编排方案，如果你需要打造极具个性化，需要面向特殊应用场景的容器平台，mesos+bk-framework方案是不二选择。

除编排方案外，蓝鲸容器管理平台还提供了无差异的服务治理方案，为业务提供服务注册与服务发现、负载均衡、DNS、流量代理等服务。

本次开源的版本，与蓝鲸社区版中的蓝鲸容器管理平台版本保持一致并且同步更新。并且蓝鲸社区版内会有内置的SaaS对接蓝鲸容器管理平台，为用户提供容器编排的界面化操作。

## Overview

* [架构设计](./docs/overview/architecture.md)
* [代码结构](./docs/overview/code_directory.md)
* [功能说明](./docs/overview/function.md)

## Features

* 支持基于k8s和Mesos双引擎编排
* 支持多集群管理
* 支持插件化自定义编排调度策略
* 支持服务升级，扩缩容，滚动升级，蓝绿发布等
* 支持configmap，secret，磁盘卷挂载，共享盘挂载等
* 支持服务发现，域名解析，访问代理等基础服务治理方案
* 支持可扩展的资源配额定义
* 支持容器内IPC机制
* 支持多种容器网络方案

如果想了解以上功能的详细说明，请参考蓝鲸容器管理平台[白皮书](https://docs.bk.tencent.com/bcs/)

## Getting Started

> 容器管理平台是蓝鲸智云社区版V5.1以上推出的产品，后台服务可以独立部署与使用。如果需要SaaS的支持，则需要与蓝鲸社区版软件配合使用。

> 目前社区版5.1在灰度内测中，若想体验，请填写问卷留下邮箱等信息，蓝鲸将在1-2个工作日通过邮箱方式，交付软件。感谢对蓝鲸的支持与理解。
> 问卷链接：[https://wj.qq.com/s2/3830461/a8bc/](https://wj.qq.com/s2/3830461/a8bc/)

> 蓝鲸社区版5.1完全开放下载时间为2019-07-05

* [下载与编译](docs/install/source_compile.md)
* [安装部署](docs/install/deploy-guide.md)
* [API使用说明](./docs/apidoc/api.md)

## Version Plan

* [版本详情](./docs/version/README.md)

## Contributing

对于项目感兴趣，想一起贡献并完善项目请参阅[contributing](./CONTRIBUTING.md)。

[腾讯开源激励计划](https://opensource.tencent.com/contribution) 鼓励开发者的参与和贡献，期待你的加入。

## Support

* 参考bk-bcs[安装文档](docs/install/deploy-guide.md)
* 阅读 [源码](https://github.com/Tencent/bk-bcs)
* 阅读 [wiki](https://github.com/Tencent/bk-bcs/wiki) 或者寻求帮助
* 了解蓝鲸社区相关信息：蓝鲸社区版交流QQ群 495299374
* 直接反馈issue，我们会定期查看与答复

## FAQ

[https://github.com/Tencent/bk-bcs/wiki/FAQ](https://github.com/Tencent/bk-bcs/wiki/FAQ)

## License

bk-bcs是基于MIT协议， 详细请参考[LICENSE](./LICENSE.TXT)。
