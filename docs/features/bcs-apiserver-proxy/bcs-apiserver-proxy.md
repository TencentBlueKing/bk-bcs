# bcs-apiserver-proxy

`bcs-apiserver-proxy`组件主要负责`kubernetes`集群内部组件连接`master`节点的高可用特性，主要是基于动态服务发现`master`节点端点IP和通过`ipvs`建立本地负载均衡守护规则实现高可用访问`master`节点。

主要功能：

* 自动监测集群`master`节点的新增、故障、删除状态，动态服务发现集群`master`节点的端点`IP`
* 建立本地负载`ipvs`规则并动态刷新规则实现高可用访问

## 架构设计

核心原理：通过本地`ipvs`代理节点解决`master`高可用问题，实现负载均衡。每个node节点上都启动一个负载均衡，上游就是master节点，负载方式有很多 ipvs nginx等，最终使用内核ipvs实现后端rs规则动态刷新，实现自动化。

bcs-apiserver-proxy架构工作流程如图示：

![bcs-apiserver-proxy工作流程图](./img/bcs-apiserver-proxy-work-flow.png)

**名词解释**

VS：Virtual Server，指虚拟的apiserver服务地址

RS：Real Server，指真实的apiserver服务地址

## 使用指南

### 使用步骤

1. 下载`bk-bcs`代码，进行代码编译，生成 `bcs-apiserver-proxy`和`apiserver-proxy-tools`

  ```
git clone https://github.com/Tencent/bk-bcs.git
make apiserver-proxy
make apiserver-proxy-tools
  ```

2. 将工具`apiserver-proxy-tools`分发至各个`node`节点的 `/root`目录下, 通过工具`apiserver-proxy-tools`生成本地负载均衡的代理规则

  ```
apiserver-proxy-tools --help 查看帮助
初始化vs本地负载均衡规则(如果是ipv6,需要使用"[]"将ip地址括起来)
apiserver-proxy-tools -cmd init -vs vip:vport -rs master0:port -rs master1:port -rs master2:port -scheduler sh
  ```

3. 将VIP添加到K8S apiserver证书中
4. `kubelet`启动时`kube-config`文件配置连接生成的lvs(https://vip:vport)即可
5. 生成镜像并通过`daemonSet`进行部署，负责维护本地负载均衡规则，动态更新后端 rs ，并将ipvs规则持久化到本地

```
cd  bk-bcs/build/bcs.xxxxxxx-21.06.30/bcs-k8s-master/bcs-apiserver-proxy
docker build -t image名称 .
docker push 上传至镜像仓库
kubectl apply -f bcs-apiserver-proxy.yaml
```

5. 修改kube-proxy配置,添加vip白名单防止kube-proxy删除自定义的ipvs规则，并重启kube-proxy所有pod

```
kubectl edit cm -n kube-system kube-proxy
# excludeCIDRs添加vip
ipvs:
      excludeCIDRs:
        - 10.0.xx.xx/32

# 重启kube-proxy
kubectl rollout restart daemonset -n kube-system kube-proxy
```

### 场景
####  新增node节点
首先要修改master节点kube-public命名空间下名为cluster-info的ConfigMap，将server地址改为vip, 然后使用kubeadm join vip:port添加节点, 添加节点成功后会自动启动`daemonset`的`pod`守护代理规则

####  重启node节点
`apiserver-proxy-tools`第一次初始化同时会创建自动启动任务，重启时从本地持久化文件中恢复负载均衡代理规则。

#### 新增master节点/master节点IP改变/master节点down/master节点恢复
`node`节点上`pod`自动守护规则，当新增master节点、master节点IP改变、master节点down、master节点恢复，均会自动增加或者剔除后端rs节点，实现内部master节点的高可用访问

### 注意

* VIP授权问题，生成K8S证书文件时需要将上述`vip`添加至授权IP列表
* 集群VIP地址不能和集群其他地址段重复
* bcs-apiserver-proxy组件的参数`lvsScheduler`和`ipvsPersistDir`需要与节点上使用apiserver-proxy-tools初始化时一致,建议默认不修改

## 参考
   [lvscare设计](https://github.com/sealyun/lvscare) 
