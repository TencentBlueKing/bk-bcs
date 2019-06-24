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

## 部署BCS K8S组件

在K8S集群中部署如下组件，即能将K8S纳入BCS管理
- bcs-kube-agent:
- bcs-k8s-watch:

### bcs-kube-agent

#### 创建集群相关信息

- 参考[BCS K8S API文档](https://github.com/Tencent/bk-bcs/blob/master/docs/apidoc/k8s.md)
  - 创建用户
  - 创建`user_token`
  - 创建集群
  - 创建`register_token`

#### 制作镜像

```bash
cd $GOPATH/src/bk-bcs/build/bcs.${VERSION}/bin/bcs-kube-agent
docker build .
```

将build后的镜像推送至Docker仓库或者导入至要K8S Master机器本地

#### 修改bcs-kube-agent配置文件

```bash
cd $GOPATH/src/bk-bcs/bcs-k8s/bcs-kube-agent/install
```

- 修改`kube-agent.yaml`配置文件
  - `image`字段：填写镜像的`<REPOSITORY>:<TAG>`
  - `bke-address`字段：填写创建的Cluster_ID

- 修改`kube-agent-secret.yml`文件
  - `token`字段： 将集群`register_token`base64 encoding
  - `bke-cert`字段：将部署BCS Service层组件时生成的CA证书base64 encoding

#### 部署bcs-kube-agent

在K8S集群中执行

```bash
kubectl create -f kube-agent-secret.yml
kubectl apply -f kube-agent.yaml
```

### bcs-k8s-watch

#### 制作镜像

```bash
cd $GOPATH/src/bk-bcs/build/bcs.${VERSION}/bin/bcs-kube-agent
docker build .
```

将build后的镜像推送至Docker仓库或者导入至要K8S Master机器本地

#### 修改bcs-k8s-watch配置文件

参考[bcs-watch部署文档](https://github.com/Tencent/bk-bcs/blob/master/docs/features/k8s-watch/k8s-watch%E9%83%A8%E7%BD%B2%E6%96%87%E6%A1%A3.md)修改配置文件

#### 部署bcs-k8s-watch

将`bcs-datawatch.yaml`放置于每台K8S master机器的manifests目录，默认`/etc/kubernetes/manifests`，以static pod方式运行`bcs-k8s-watch`
