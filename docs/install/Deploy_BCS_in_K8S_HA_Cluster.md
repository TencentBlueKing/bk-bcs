Table of Contents
=================

* [BCS高可用Kubernetes集群部署](#bcs高可用kubernetes集群部署)
   * [目标机器环境准备](#目标机器环境准备)
      * [硬件](#硬件)
      * [操作系统](#操作系统)
      * [网络](#网络)
   * [部署BCS K8S组件](#部署bcs-k8s组件)
      * [Requirements](#requirements)
         * [创建集群相关信息](#创建集群相关信息)
         * [制作并推送镜像](#制作并推送镜像)
      * [部署/更新 bcs-k8s 组件](#部署更新-bcs-k8s-组件)
         * [导入证书](#导入证书)
         * [部署](#部署)

# BCS高可用Kubernetes集群部署

部署BCS管理的高可用Kubernetes集群有2种方式：

- **蓝鲸社区版(BCS增值包)**：使用[容器服务控制台](https://docs.bk.tencent.com/bcs/Container/QuickStart.html)一键创建Kubernetes集群，Master节点推荐3或5台，集群创建成功后，可以进入集群节点列表，为集群增加节点

- **手动部署**：
  - 参考社区的方案搭建
    - [《和我一步步部署 kubernetes 集群》](https://github.com/opsnull/follow-me-install-kubernetes-cluster)
    - 使用[Kubespray](https://kubernetes.io/docs/setup/custom-cloud/kubespray/)
    - 使用[kubeasz](https://github.com/easzlab/kubeasz)(中文)
  - 搭建完Kubernetes集群后，参考本文档指引部署BCS组件

## 目标机器环境准备

### 硬件

| 资源      | 配置    |
| -------- | ------- |
| CPU      | 4核     |
| Mem      | 8GB     |
| Disk     | >= 50GB |

>注：Slave节点可根据业务规模增加机器配置 

### 操作系统

CentOS 7及以上系统，推荐CentOS 7.4

> 注：
>
> 1. 集群网络模式使用Overlay网络模式，需要有NAT模块(iptable_nat)
>
> 2. CRI推荐使用Docker，storage-dirver推荐使用overlay2，CentOS内核版本需要在3.10.0-693及以上支持Overlay2

###  网络

- 放行集群层到服务层的机器所有网络策略
- 放行集群层机器间的所有网络策略

<!-- To-Do: 严格的网络策略 -->

## 部署 BCS K8S 操作

项目提供了打包的 K8S 业务集群组件 helm chart —— [bcs-k8s](../../install/helm/bcs-k8s)，只需部署该 chart 即可将集群接入 BCS Serivce 控制面。该 chart 是以下组件的集合：
- bcs-k8s-watch
- bcs-kube-agent
- bcs-gamestatefulset-operator
- bcs-gamedeployment-operator
- bcs-hook-operator

其中，`bcs-k8s-watch` 与 `bcs-kube-agent` 提供接入 BCS Service 控制面的能力，余下的组件则提供 BCS 定制化的 workload 支持，详细可见对应 workload 的文档：[GameStatefulSet](../features/bcs-gamestatefulset-operator/README.md)，[GameDeployment](../features/bcs-gamedeployment-operator/README.md)

有关该组合 chart 详细配置说明，请参阅 [bcs-k8s doc](./deploy-guide-bcs-k8s.md)

### Requirements

#### 创建集群相关信息

- 参考[BCS K8S API文档](https://github.com/Tencent/bk-bcs/blob/master/docs/apidoc/k8s.md)
  - 创建用户
  - 创建`user_token`
  - 创建集群
  - 创建`register_token`

#### 制作并推送镜像

```bash
# example
# VERSION=v1.20.11
# module=kube-agent
# module_full_name=bcs-kube-agent
# docker_registry=xxx.yyy.zzz/name
VERSION=${VERSION} make ${module}
cd $GOPATH/src/bk-bcs/build/bcs.${VERSION}/bcs-k8s-master/${module_full_name}
docker build . -t ${docker_registry}/${module_full_name}:${VERSION}
docker push ${docker_registry}/${module_full_name}:${VERSION}
```

### 部署/更新 bcs-k8s 组件

#### 导入证书

将证书导入 bcs-k8s/charts/bcs-cluster-init/cert 目录（若不存在请手动创建），包括：
  - ca证书，命名为bcs-ca.crt
  - server证书，命名为bcs-server.crt
  - server私钥，命名为bcs-server.key
  - client证书，命名为bcs-client.crt
  - client私钥，命名为bcs-client.key


#### 部署

按照 bcs-k8s [部署文档](./deploy-guide-bcs-k8s.md)填写必要的部署参数，并执行以下命令

```json
helm upgrade bcs-k8s ./bcs-k8s \ 
-f your_values.yaml \ 
-n bcs-system \ 
--install
```