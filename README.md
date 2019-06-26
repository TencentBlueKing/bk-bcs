![BCS.png](./docs/logo/logo_en.png)
---
[![license](https://img.shields.io/badge/license-mit-brightgreen.svg?style=flat)](https://github.com/Tencent/bk-bcs/blob/master/LICENSE)[![Release Version](https://img.shields.io/badge/release-1.12.x-brightgreen.svg)](https://github.com/Tencent/bk-bcs/releases) ![BK Pipelines Status](https://api.bkdevops.qq.com/process/api/external/pipelines/projects/bcs/p-c03c759b697f494ab14e01018eccb052/badge?X-DEVOPS-PROJECT-ID=bcs) [![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](https://github.com/Tencent/bk-bcs/pulls)   

[中文文档](./README_cn.md)

BlueKing Container Service (BCS) is a container management and orchestration platform for the micro-services under the BlueKing ecosystem.

BlueKing Container Service provides a two-engine-driven container orchestration scheme based on native Kubernetes and mesos `bk-framework`, and users can choose to either one of them to orchestrate their own applications. The Kubernetes method is mainly based on Kubernetes solution. In addition to providing native functional support, it also provides  seamless integration between the native Kubernetes cluster and the BlueKing platform. Users can experience container technology with Kubernetes community edition in the BlueKing platform in a way that is indistinguishable and convenient. The mesos `bk-framwork` solution is a container orchestration program for BlueKing that capable of customization. If you need to create a highly personalized container platform for special application scenarios, the mesos `bk-framework` solution is a great choice.

In addition to the orchestration program, the BlueKing Container Service also provides an abstract service management solution to provide services such as load balancing, service registration, service discovery, DNS, and traffic proxies.

The open source version of the BlueKing Container Service is consistent with the BlueKing Community Edition and is updated synchronously. BlueKing Community Edition will have a built-in SaaS (Software As A Service) to communicate with BCS, this will provide users with interface to view container operations.

## Overview

* [Architecture Design](./docs/overview/architecture.md)
* [code structure](./docs/overview/code_directory.md)
* [Function Description](./docs/overview/function.md)

## Features

* Support for dual engine orchestration based on Kubernetes and Mesos
* Support multi-cluster management
* Support plug-in custom orchestration scheduling strategy
* Support service upgrade, expansion and expansion, rolling upgrade, blue/green deploy strategy, etc.
* Support configmap, secret, disk volume mount, shared disk mount, etc.
* Support basic service management solutions such as service discovery, domain name resolution, and access agents, etc
* Support for scalable resource quota definitions
* Support in-container IPC mechanism
* Support multiple container network solutions (CNI)

For a detailed description of the above features, please refer to the BlueKing Container Management Platform [white paper](https://docs.bk.tencent.com/bcs/)

## Getting Started

* [Download and Compile](docs/install/source_compile.md)
* [Installation Deployment](docs/install/deploy-guide.md)
* [API Usage Notes](./docs/apidoc/api.md)

## Version Plan

* [Version Details](./docs/version/README.md)

## Contributing

If you are interested in this project and want to contribute and improve the project together, please refer to [contributing](./CONTRIBUTING.md).
[Tencent Open Source Incentive Program](https://opensource.tencent.com/contribution) We encourage developers to participate and contribute, and look forward to your joining us.

## Support

* Refer to `bk-bcs`[installation documentation](docs/install/deploy-guide.md)
* Read [source code](https://github.com/Tencent/bk-bcs)
* Read [wiki](https://github.com/Tencent/bk-bcs/wiki) or ask for help
* Learn about the BlueKing Community: QQ group 495299374
* Issues, we will check and reply regularly

## FAQ

[https://github.com/Tencent/bk-bcs/wiki/FAQ](https://github.com/Tencent/bk-bcs/wiki/FAQ)

## License

Bk-bcs is based on the MIT protocol. Please refer to [LICENSE](./LICENSE.TXT) for details.
