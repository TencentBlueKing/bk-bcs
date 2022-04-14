![BCS.png](./docs/logo/logo_en.png)

---
[![license](https://img.shields.io/badge/license-mit-brightgreen.svg?style=flat)](https://github.com/Tencent/bk-bcs/blob/master/LICENSE)[![Release Version](https://img.shields.io/badge/release-1.18.8-brightgreen.svg)](https://github.com/Tencent/bk-bcs/releases) ![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/bcs/p-c03c759b697f494ab14e01018eccb052/badge?X-DEVOPS-PROJECT-ID=bcs) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/Tencent/bk-bcs/pulls)

[中文文档](./readme.md)

> **Important Note**: The `master` branch may be in an unstable or unavailable state during development*.
> Please use [releases] (https://github.com/Tencent/bk-bcs/releases) instead of `master` to get stable binaries.

BlueKing Container Service (BCS, same below) is a container management and orchestration platform for the micro-services under the BlueKing ecosystem.

BCS has been running stably for more than three years in Tencent. After more than 30 of different architectures and different scales of service verification, the largest service consists of four independent clusters with more than 500 physical servers, nearly 6000 POD, and more than 30 namespaces for isolation.

BCS is part of the BlueKing ecosystem. Its overall structure is organized according to the BlueKIng PaaS system. This open source part is the BCS back-end, an atomic platform under the BlueKing, which mainly outputs service scheduling and service management capabilities. The web page of the BCS is partially presented by the BlueKing SaaS light weight application. The version of the SaaS can be obtained through the latest BlueKing Community Edition or Enterprise Edition; or through the [BlueKing Official Website] (https://bk.tencent.com/) to get open source information.

## Overview

* [Architecture Design](docs/overview/README.md)
* [code structure](./docs/overview/code_directory.md)
* [Function Description](./docs/overview/function.md)

Please refer to the BlueKing Container Management Platform [white paper](https://docs.bk.tencent.com/bcs/)

## Features

* Support for dual engine orchestration based on k8s and Mesos
  * [Understanding k8s program related information](https://kubernetes.io/zh/)
  * [Understanding information about the mesos program](./docs/features/mesos/基于mesos的容器编排.md)
* Support heterogeneous service access
  * [Understanding Stateful Business Solutions]
  * [Learn about other non-container friendly features](./docs/features/mesos/基于mesos的服务编排.md)
* Cross-cloud management container management
  * [Cross Cloud Container Management Solution](./docs/features/solutions/BCS跨云容器管理方案.md)
  * [Support Windows container]
* Plug-in secondary development capabilities
  * [Understanding Community CNI](https://github.com/containernetworking/cni)
  * [Understanding CSI Standard](https://github.com/container-storage-interface/spec/blob/master/spec.md)
  * [K8S Custom Orchestration Scheduling Policy](./docs/features/solutions/k8s-custom-scheduler.md)
  * [Mesos Custom Scheduling Policy](./docs/features/mesos/自定义编排调度策略.md)

## Experience

* [Resource reuse of R&D environment through BCS](./docs/features/practices/通过BCS解决研发环境的资源问题.md)
* [Rolling upgrade of business through BCS](./docs/features/practices/rolling-update-howto.md)
* [Blue-green release of business completed through BCS](./docs/features/practices/blue-green-deployment.md)
* [istio integration](./docs/features/practices/istio.md) coming soon...

## Getting Started

> BCS is a product launched by BlueKing Smart Cloud Community Edition V5.1. The background service can be deployed and used independently. If you need SaaS support, you need to work with the BlueKing Community Edition software.

> At present, the community version 5.1 is in the beta internal test. If you want to experience it, please fill in the questionnaire and leave the mailbox and other information. The BlueKing will deliver the software by email within 1-2 business days. Thank you for your support and understanding of the BlueKing.
> Questionnaire link: [https://wj.qq.com/s2/3830461/a8bc/](https://wj.qq.com/s2/3830461/a8bc/)

* [Download and Compile](docs/install/source_compile.md)
* [Installation Deployment](docs/install/deploy-guide.md)
* [API Usage Notes](./docs/apidoc/api.md)

## Roadmap

> [history versions](./docs/version/README.md)

* Open all tencent cloud plugins source code
* Kubenretes federation solution intergration
* All documents refactor 

## Contributing

If you are interested in this project and want to contribute and improve the project together, please refer to [contributing](./CONTRIBUTING.md).
[Tencent Open Source Incentive Program](https://opensource.tencent.com/contribution) We encourage developers to participate and contribute, and look forward to your joining us.

## Support

* Refer to `bk-bcs`[installation documentation](docs/install/deploy-guide.md)
* Read [source code](https://github.com/Tencent/bk-bcs)
* Read [wiki](https://github.com/Tencent/bk-bcs/wiki) or ask for help
* Learn about the BlueKing Community: QQ group 495299374
* [Issue](https://github.com/Tencent/bk-bcs/issues) we will check and reply regularly

## FAQ

[BCS FAQ](https://docs.bk.tencent.com/bcs/Container/FAQ/faq.html)
[github wiki FAQ](https://github.com/Tencent/bk-bcs/wiki/FAQ)

## Blueking Community

* [BK-CI](https://github.com/Tencent/bk-ci):BlueKing Continuous Integration Platform is an open source, continuous integration and continuous delivery system that makes it easy to present your development process to you.
* [BK-CMDB](https://github.com/Tencent/bk-cmdb):BlueKing Configuration Platform (Blue Whale CMDB) is an enterprise-level configuration management platform for assets and applications.
* [BK-JOB](https://github.com/Tencent/bk-job):BlueKing JOB is a set of operation and maintenance script management platform with the ability to handle a large number of tasks concurrently.
* [BK-PaaS](https://github.com/Tencent/bk-PaaS):BlueKing PaaS platform is an open development platform that allows developers to create, develop, deploy and manage SaaS applications quickly and easily.
* [BK-SOPS](https://github.com/Tencent/bk-sops):Standard Operations (SOPS) is a system for task scheduling and execution through a visual graphical interface. It is a lightweight in the BlueKing system scheduling SaaS products.

## License

Bk-bcs is based on the MIT protocol. Please refer to [LICENSE](./LICENSE.txt) for details.
