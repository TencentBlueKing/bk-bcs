# bcs-ops

## Usage

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

---

# 集群操作

## 脚本

### 1. 集群控制面故障替换

1. 在正常 master 节点上执行`./bcs-ops --render joincmd`获取加入集群的指令
2. 在新控制面节点上加入集群的指令，加入集群
3. 在新节点上执行命令删除故障的 K8S 节点以及对应的 etcd 节点

```bash
kubectl delete node xxx
etcdctl member remove xxx
```

4.故障节点如果能够登录，执行`./bcs-ops -c master`清理节点

## 标准运维

### 1. 【BCS】K8S master replace

参数

1. master_ip 一个当前存在于集群的 master，且不是本次被替换的 master 的 ip
2. new_master_ip 本次将被替换进集群的 master 的 ip
3. unwanted_master_ip 本次将被替换出集群的 master 的 ip
4. unwanted_master_name 本次将被替换进集群的 master 的节点名
5. workspace 节点上的 bcs-ops 目录

功能描述

1. 扩容 new_master_ip 指定的 master 节点

2. 清理掉 unwanted_master_ip 指定的 master 节点上的 k8s 环境以及 unwanted_master_name 对应的 k8s 节点以及 etcd 节点

---

# etcd 操作

## 脚本

### 1. operate_etcd backup (etcd 备份)

参数

1. endpoint etcd 实例 IP
2. cacert 访问 etcd 的 ca 证书文件路径
3. cert 访问 etcd 的证书文件路径
4. key 访问 etcd 的 key 文件路径
5. backup_file 备份文件路径

功能描述

1. 请求 endpoint 指定的 etcd 实例，获取 snapshot 存储在 backup_file 指定的路径

### 2. operate_etcd restore (etcd 恢复)

> 注意：etcd 集群恢复时所有 etcd 节点都必须使用同一份 snapshot 文件恢复

参数

1. backup_file 备份文件路径
2. data_dir 数据恢复路径
3. member_name 本机的 etcd 节点的名字
4. member_peer 本机的 etcd 节点的 peer url
5. initial_cluster 此次恢复的 etcd 集群所有成员信息

功能描述

1. 根据 member_name，member_peer，initial_cluster 参数将数据从 backup_file 中恢复到 data_dir

### 3. operate_etcd new (etcd 新实例)

参数

1. name etcd 集群名
2. data_dir 数据目录
3. peer_port etcd 节点 peer port
4. service_port etcd 节点 service port
5. metric_port etcd 节点 metric port
6. initial_cluster 此次恢复的 etcd 集群所有成员信息
7. cacert 访问 etcd 的 ca 证书文件路径
8. cert 访问 etcd 的证书文件路径
9. key 访问 etcd 的 key 文件路径

功能描述

1. 根据参数基于原本 kubeadm 创建出来的 etcd.yaml 文件进行替换，并用静态 pod 的方式拉起新集群的本机节点

## 标准运维

### 1. 【BCS】etcd backup

参数

1. host_ip_list 需要进行备份的 etcd 节点 ip，多个使用,隔开
2. cacert 访问 etcd 的 ca 证书文件路径
3. cert 访问 etcd 的证书文件路径
4. key 访问 etcd 的 key 文件路径
5. backup_file 备份文件路径
6. workspace 节点上的 bcs-ops 目录

功能描述

1. 在各个 etcd 节点上，通过本机的 endpoint 获取 snapshot 到 backup_file 指定目录

### 2. 【BCS】etcd restore

参数

1. host_ip_list 需要进行备份的 etcd 节点 ip，多个使用,隔开
2. source_host 备份文件来源机器
3. source_file 备份文件路径
4. data_dir etcd 数据目录
5. clusterinfo_file 集群信息文件路径
6. workspace 节点上的 bcs-ops 目录

功能描述

1. 将 source_file 备份文件从 source_host 传到各台 etcd 节点机器上后，根据 clusterinfo_file 中的信息将数据恢复到 data_dir 指定的目录

### 3. etcd new

参数

1. host_ip_list 新集群的 etcd 节点 ip，多个使用,隔开
2. name etcd 集群名
3. data_dir 数据目录
4. peer_port etcd 节点 peer port
5. service_port etcd 节点 service port
6. metric_port etcd 节点 metric port
7. initial_cluster 此次恢复的 etcd 集群所有成员信息
8. cacert 访问 etcd 的 ca 证书文件路径
9. cert 访问 etcd 的证书文件路径
10. key 访问 etcd 的 key 文件路径
11. workspace 节点上的 bcs-ops 目录

功能描述

1. 根据参数基于原本 kubeadm 创建出来的 etcd.yaml 文件进行替换，并用静态 pod 的方式拉起新集群的所有节点
