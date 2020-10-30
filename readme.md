![蓝鲸容器管理平台.png](./docs/logo/bcs_zh_v2.png)

---
[![license](https://img.shields.io/badge/license-mit-brightgreen.svg?style=flat)](https://github.com/Tencent/bk-bcs/blob/master/LICENSE)[![Release Version](https://img.shields.io/badge/release-1.18.9-brightgreen.svg)](https://github.com/Tencent/bk-bcs/releases) ![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/bcs/p-c03c759b697f494ab14e01018eccb052/badge?X-DEVOPS-PROJECT-ID=bcs) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/Tencent/bk-bcs/pulls)

[EnglishDocs](./readme_en.md)

> **重要提示**: `master` 分支在开发过程中可能处于 *不稳定或者不可用状态* 。
> 请通过[releases](https://github.com/Tencent/bk-bcs/releases) 而非 `master` 去获取稳定的二进制文件。

蓝鲸容器管理平台（Blueking Container Service，简称BCS）是蓝鲸体系下，以容器技术为基础，为各种架构的应用提供编排管理和治理服务的基础平台。BCS支持两种不同
的集群模式，分别为原生K8s模式和基于Mesos自研的模式；k8s模式紧跟社区发展，充分利用社区资源，避免过度修改导致版本碎片；mesos模式针对游戏等复杂应用深度定制，
解决这类应用在微服务过渡阶段容器化的后顾之忧。

BCS在腾讯内部已经稳定运行三年以上，经过几十款不同架构、不同规模的业务验证，其中规模最大的业务包含五个独立的集群，共600+物理机资源（单机48核以上，128G以上内存），近7000 POD，使用30多个命名空间进行隔离。

BCS作为蓝鲸体系的一部分，其整体结构按照蓝鲸PaaS体系组织，本次开源的部分为BCS后台部分，为蓝鲸PaaS体系下的原子平台，主要输出服务编排和服务治理的能力。BCS的操作页面部分通过蓝鲸SaaS轻应用的方式呈现，可以通过最新的蓝鲸社区版或者企业版获取该SaaS的版本；或者直接获取[SaaS开源代码](https://github.com/Tencent/bk-bcs-saas)自行安装部署与集成。

## Overview

* [架构设计](./docs/overview/architecture.md)
* [代码结构](./docs/overview/code_directory.md)
* [功能说明](./docs/overview/function.md)

了解BCS更详细功能，请参考蓝鲸容器管理平台[白皮书](https://docs.bk.tencent.com/bcs/)

## Features

* 支持基于k8s和Mesos双引擎编排
  * [了解k8s方案相关信息](./docs/features/k8s/基于k8s的容器编排.md)
  * [了解mesos方案相关信息](./docs/features/mesos/基于mesos的服务编排.md)
* 支持异构业务接入
  * 有状态业务解决方案
    * [k8s有状态应用部署](./docs/features/solutions/k8s有状态应用方案.md)
    * [mesos有状态应用部署](./docs/features/mesos/基于mesos的服务编排.md#有状态服务的部署方案)
  * [了解其他非容器友好特性的解决方案](./docs/features/mesos/基于mesos的服务编排.md#非容器在bcs部署方案)
* 跨云跨OS管理容器
  * [跨云容器管理方案](./docs/features/solutions/BCS跨云容器管理方案.md)
  * 支持windows容器
* 插件化的二次开发能力
  * 网络插件
    * [了解社区CNI标准](https://github.com/containernetworking/cni)
    * [CNI插件实践](./docs/features/solutions/cni-practise.md)
  * 存储插件
    * [了解社区CSI标准](https://github.com/container-storage-interface/spec/blob/master/spec.md)
    * [CSI插件实战案例](./docs/features/solutions/如何编写一个csi存储插件.md)
  * 编排调度
    * [K8S自定义编排调度策略](./docs/features/solutions/k8s-custom-scheduler.md)
    * [Mesos自定义编排策略](./docs/features/mesos/自定义编排调度策略.md)

## Experience

* [通过BCS解决研发环境的资源复用](./docs/features/practices/通过BCS解决研发环境的资源复用.md)
* [通过BCS完成业务的滚动升级](./docs/features/practices/rolling-update-howto.md)
* [通过BCS完成业务的蓝绿发布](./docs/features/practices/blue-green-deployment.md)
* [BCS集成istio案例](./docs/features/practices/istio.md) coming soon...

## Getting Started

> 容器管理平台是蓝鲸智云社区版V5.1以上推出的产品，后台服务可以独立部署与使用。如果需要SaaS的支持，则需要与蓝鲸社区版软件配合使用。
> 目前社区版5.1在灰度内测中，若想体验，请填写问卷留下邮箱等信息，蓝鲸将在1-2个工作日通过邮箱方式，交付软件。感谢对蓝鲸的支持与理解。
> 问卷链接：[https://wj.qq.com/s2/3830461/a8bc/](https://wj.qq.com/s2/3830461/a8bc/)

* [下载与编译](docs/install/source_compile.md)
* [安装部署](docs/install/deploy-guide.md)
* [API使用说明](./docs/apidoc/api.md)

## Roadmap

> [历史版本详情](./docs/version/README.md)

* 腾讯云K8S相关插件开源
* K8S集群联邦方案集成
* 开源文档重构并归档至蓝鲸文档中心

## Contributing

对于项目感兴趣，想一起贡献并完善项目请参阅[contributing](./CONTRIBUTING.md)。

[腾讯开源激励计划](https://opensource.tencent.com/contribution) 鼓励开发者的参与和贡献，期待你的加入。

## Support

* 参考bk-bcs[安装文档](docs/install/deploy-guide.md)
* 阅读 [源码](https://github.com/Tencent/bk-bcs)
* 阅读 [wiki](https://github.com/Tencent/bk-bcs/wiki) 或者寻求帮助
* 了解蓝鲸社区相关信息：蓝鲸社区版交流QQ群 495299374
* 直接反馈[issue](https://github.com/Tencent/bk-bcs/issues)，我们会定期查看与答复

## FAQ

* [蓝鲸容器FAQ](https://docs.bk.tencent.com/bcs/Container/FAQ/faq.html)
* [蓝鲸文档中心] 建设中...

## Blueking Community

* [BK-BCS-SAAS](https://github.com/Tencent/bk-bcs-saas)：蓝鲸容器SAAS是蓝鲸针对容器管理平台提供的配置设施，为用户提供便利的容器操作。
* [BK-CI](https://github.com/Tencent/bk-ci)：蓝鲸持续集成平台是一个开源的持续集成和持续交付系统，可以轻松将你的研发流程呈现到你面前。
* [BK-CMDB](https://github.com/Tencent/bk-cmdb)：蓝鲸配置平台（蓝鲸CMDB）是一个面向资产及应用的企业级配置管理平台。
* [BK-PaaS](https://github.com/Tencent/bk-PaaS)：蓝鲸PaaS平台是一个开放式的开发平台，让开发者可以方便快捷地创建、开发、部署和管理SaaS应用。
* [BK-SOPS](https://github.com/Tencent/bk-sops)：标准运维（SOPS）是通过可视化的图形界面进行任务流程编排和执行的系统，是蓝鲸体系中一款轻量级的调度编排类SaaS产品

## 认证

**蓝鲸智云容器管理平台**通过中国**云计算开源产业联盟**组织的可信云容器解决方案评估认证。蓝鲸智云容器管理平台在基本能力要求、应用场景技术指标、安全性等解决方案质量方面，以及产品周期、运维服务、权益保障等服务指标的完备性和规范性方面均达到可信云容器解决方案的评估标准。应用场景满足以下四个：

* 开发测试场景
* 持续集成 & 持续交付
* 运维自动化
* 微服务

![certificate](./docs/overview/certificate.jpg)

## License

bk-bcs是基于MIT协议， 详细请参考[LICENSE](./LICENSE.txt)。
