bcs-ops

# Usage

```plaintext
Usage:
  bcs-ops
    [ -h --help -?     show usage ]
    [ -v -V --version  show script version]
    [ -i --install     support: master node helm op]
    [ -r --render      suppport: bcsenv kubeadm joincmd]
    [ -c --clean       support: master node bcsenv op]
```

## 示例

### 集群创建与节点添加

1. 在第一台主机（后称中控机）上启动集群控制平面：`./bcs-ops --instal master`，集群启动成功后会显示加入集群的指令
2. 集群加入指令有效期为 1 小时，中控机执行 `./bcs-ops --render joincmd` 可再次渲染生成加入集群的指令，渲染结果如下所示

```plaintext
======================
# Expand Control Plane, run the following command on new machine
set -a
CLUSTER_ENV=xxxx
MASTER_JOIN_CMD=xxxx
set +a
./bcs-ops -i master
======================
# Expand Worker Plane, run the following command on new machine
set -a
CLUSTER_ENV=xxxx
JOIN_CMD=xxxx
set +a
./bcs-ops -i node
======================
```

3. 添加控制平面节点(master node)，以及添加工作节点(wroker node)，执行第二步渲染生成的的加入集群指令

### 集群 node 节点移除

1. 在中控机上先移除 ip 地址为 `$IP` 节点

```bash
node_name="node-$(tr ":." "-" <<<"$IP")"
# https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/safely-drain-node/
kubectl drain --ignore-daemonsets $node_name
kubectl delete node $node_name
```

2. 被移除的节点上执行 `./bcs-ops --clean node`

### 中控机安装 helm 工具

`./bcs-ops --install helm`

### 部署 localpv

> 注意：localpv 部署依赖 helm。在添加 node 节点的过程中，并没有执行 `mount localpv` 动作。
> $BK_HOME 默认路径为`/data/bcs/`，

1. node 节点执行`./system/mount_localpv /mnt/blueking 20`。该工具会在`/$BK_HOME/localpv`目录下创建 20 个子目录，并通过 mount bind 挂载到对应的`/mnt/blueking/localpv`路径下。若使用节点上已有的挂载点目录，这一步可以跳过。

2. 中控机执行`./k8s/install_localpv /mnt/blueking`，localpv 会寻找节点`/mnt/blueking`下所有的挂载点，创建相应的`Persistentvolumes`资源。

3. 当步骤 2 执行后，新的加入的 node 节点如果要添加`Persistentvolumes`资源，重新执行步骤 1、2，即可重启 localpv 的 pod 实现挂载。


## 环境变量

通过配置环境变量来设置集群相关的参数。在中控机创建集群前，通过 `set -a` 设置环境变量。

### 示例：创建 ipv6 双栈集群
> k8s 1.23 ipv6 特性为稳定版
```bash
set -a
K8S_VER="1.23.17"
K8S_IPv6_STATUS="DualStack"
set +a
./bcs-ops -i master
```


## IP 的获取方式
对于裸金属服务器，ipv4 通过 `10/8` 的默认路由源地址获取，ipv6 则通过 `fd00::/8` 的默认路由源地址获取。如果有多个网卡，可以手动配置该路由的源地址来选择
```bash
# 如果存在则先删除
ip route del 10/8
ip -6 route del fd00::/8
# 添加对应的路由
ip route add 10/8 via <next hop> dev <interface> src <lan_ipv4>
ip -6 route add fd00::/8 via <next hop> dev <interface> src <lan_ipv6>
```
> 注意：`fe80::/10` link-local 地址不能用于 k8s 的 node-ip。
