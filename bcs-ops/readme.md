# bcs-ops

## Usage

```plaintext
Usage:
  bcs-ops
    [ -h --help -?     show usage ]
    [ -v -V --version  show script version]
    [ -i --install     support: master node helm op]
    [ -r --render      support: bcsenv kubeadm joincmd]
    [ -c --clean       support: master node bcsenv op]
	[ -e --check       support: all]
    [ -e --check ]
```

## 预置检查

机器执行`./bcs-ops --check all`，脚本将对这些 `check_kernel check_swap check_selinux check_firewalld check_yum_proxy check_http_proxy check_openssl check_hostname check_tools` 项目进行统一检查。您应当注意检查结果为 `[FATAL]` 项目，并在准备环境的过程中进行调整。

## 准备环境（可选）

### linux 内核版本要求
对于 ipv4 集群，linux 内核至少为 `3.10.0`
对于 ipv6 集群，linux 内核版本至少为 `4.19.1`

### iptables 策略

> 在安装的过程中`bcs-ops`会关闭机器的防火墙，`systemctl stop firewalld;systemctl disable firewalld`。

集群的机器之间应该放通如下的端口，可以使用`system/config_iptables.sh <src_cidr4> <src_cidr6>` 对集群 cidr 网段的相关协议/端口进行放通。使用`system/config_iptables.sh -h` 查看使用方法。

如果机器并没有拦截这些协议/端口，可以忽略这一步。

#### k8s

| **组件**   | **协议/端口**      | **说明**                  |
| ---------- | ------------------ | ------------------------- |
| apiserver  | tcp/6443           | secure-port               |
| controller | tcp/10257          | secure-port               |
| scheduler  | tcp/10259          | secure-port               |
| etcd       | tcp/2379, tcp/2380 | advertise_port, peer-port |
| kubelet    | tcp/10250          | metric-server need        |

#### flannel

| 模式    | 平台    | 协议端口             | 说明                              |
| ------- | ------- | -------------------- | --------------------------------- |
| vxlan   | linux   | udp/8472             |                                   |
| vxlan   | windows | udp/4789             |                                   |
| host-gw | linux   | udp/51820, udp/51821 | 前者为 ipv4，后者为 ipv6          |
| udp     |         | 8285                 | 仅当内核/网络不支持 vxlan/host-gw |

### bcs-ops 获取 IP / IP6 的方式

对于裸金属服务器，ipv4 通过 `10/8` 的默认路由源地址获取，ipv6 则通过 `fd00::/8` 的默认路由源地址获取。如果有多个网卡，可以手动配置该路由的源地址。

```bash
# 如果存在则先删除
ip route del 10/8
ip -6 route del fd00::/8
# 添加对应的路由
ip route add 10/8 via <next hop> dev <interface> src <lan_ipv4>
ip -6 route add fd00::/8 via <next hop> dev <interface> src <lan_ipv6>
```

> 注意：`fe80::/10` link-local 地址不能用于 k8s 的 node-ip。

也可以在执行脚本安装前直接手动设定

```bash
set -x
LAN_IP=<YOUR LAN IP>
LAN_IPv6=<YOUR LAN ipv6> #if enable K8S_IPv6_STATUS=dualstack
set +x
```

## 安装示例

目前仅支持 k8s `1.20.15` （默认）, `1.23.17` , `1.24.15` , `1.27.15`, `1.28.11`版本。

### 集群创建与节点添加

1. 通过`set -a 命令`配置环境变量，环境变量配置见[`环境变量`](#环境变量)
2. `./bcs-ops -r bcsenv` 在第一台主机（后称中控机）上渲染配置文件 `env/bcs.env`
3. 在中控机上启动集群控制平面：`./bcs-ops --install master`，集群启动成功后会显示加入集群的指令
4. 集群加入指令有效期为 1 小时，中控机执行 `./bcs-ops --render joincmd` 可再次渲染生成加入集群的指令，渲染结果如下所示

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

5. 添加控制平面节点(master node)，以及添加工作节点(wroker node)，执行第二步渲染生成的的加入集群指令

> `./bcs-ops -h`查看支持的命令

### 集群 node 节点移除

1. 在中控机上先移除 ip 地址为 `$IP` 节点

   ```bash
   node_name="node-$(tr ":." "-" <<<"$IP")"
   # https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/safely-drain-node/
   kubectl drain --ignore-daemonsets $node_name
   kubectl delete node $node_name
   ```

2. 被移除的节点上执行 `./bcs-ops --clean node`

## 环境变量

通过配置环境变量来设置集群相关的参数。在中控机创建集群前，通过 `set -a` 设置环境变量。 你可以执行 `system/config_envfile.sh -init` 查看默认的环境变量。
注意，当你要使用多个特性时，相关的环境变量都得申明

### host 环境变量

| 环境变量          | 默认值      | 说明                                                                       |
| ----------------- | ----------- | -------------------------------------------------------------------------- |
| `BK_HOME`         | `/data/bcs` | 软件根路径                                                                 |
| `K8S_IPv6_STATUS` | `Disable`   | 集群 ipv6 状态<br />支持`SingleStack`（ ipv6 单栈），`DualStack`双栈<br /> |
| `LAN_IP`          | `""`        | 主机 IP 地址，通过默认路由获取                                             |
| `LAN_IPv6`        | `""`        | 主机 IPv6 地址，通过默认路由获取                                           |
| `BCS_OFFLINE`     | `""`        | 离线安装，`1`为开启                                                        |
| `INSTALL_METHOD`  | `yum`       | 软件安装方式，默认`yum`。离线安装开启时此配置失效                          |
| `BCS_SYSCTL`      | `1`         | 系统调优，默认开启                                                         |

### 仓库地址 环境变量

| 环境变量         | 默认值                                                  | 说明                               |
| ---------------- | ------------------------------------------------------- | ---------------------------------- |
| `MIRROR_URL`     | `https://mirrors.tencent.com`                           | yum 仓库镜像地址，默认腾讯镜像仓库 |
| `REPO_URL`       | `https://bkopen-1252002024.file.myqcloud.com/ce7/tools` | tool (yq/jq\) 工具下载地址         |
| `MIRROR_IP`      |                                                         | mirrors.tencentyun.com 解析地址    |
| `REPO_MIRRORS`   | `https://mirror.ccs.tencentyun.com`                     | 容器镜像源，默认腾讯云             |
| `BK_PUBLIC_REPO` | `hub.bktencent.com`                                     | 容器镜像默认 registry              |
| `BKREPO_URL`     | `https://hub.bktencent.com/chartrepo/blueking<br` />    | 蓝鲸 helm chart 仓库               |

### K8S 环境变量

#### k8s 基础环境变量

| 环境变量         | 默认值                                            | 说明                                                                                                                |
| ---------------- | ------------------------------------------------- |-------------------------------------------------------------------------------------------------------------------|
| `K8S_VER`        | `1.20.15`                                         | k8s 版本，现支持 `1.20.151.23.171.24.15`                                                                                |
| `ETCD_LIB`       | `${BK_HOME}/lib/etcd`                             | 控制平面 etcd 根目录                                                                                                     |
| `KUBELET_LIB`    | `${BK_HOME}/lib/kubelet`                          | kubelet 根目录                                                                                                       |
| `K8S_EXTRA_ARGS` | `allowed-unsafe-sysctls: 'net.ipv4.tcp_tw_reuse'` | 配置kubelet自定义参数，格式为A: B, C: D[cluster sysctl](https://kubernetes.io/docs/tasks/administer-cluster/sysctl-cluster/) |
| `BCS_CP_WORKER`  | `0`                                               | 是否单节点集群，`0`关闭，`1`开启，开启后控制平面污点取消                                                                                   |

#### k8s 网络配置

| 环境变量           | 默认值                                                                                      | 说明                                                                                          |
| ------------------ | ------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------- |
| `K8S_CTRL_IP`      | 控制平面默认是自身。<br />工作平面默认是中控机 ip，如果开启了 apiserver hpa，则是 VIP<br /> | 访问控制平面 ip。<br />                                                                       |
| `K8S_SVC_CIDR`     | `10.96.0.0/12`                                                                              | 集群 svc ipv4 网段                                                                            |
| `K8S_POD_CIDR`     | `10.244.0.0/16`                                                                             | 集群 pod ipv4 网段                                                                            |
| `K8S_SVC_CIDRv6`   | `fd00::1234:5678:1:0/112`                                                                   | 集群 svc ipv6 网段（`K8S_IPv6_STATUS`不为`Disable`时有效）                                    |
| `K8S_POD_CIDRv6`   | `fd00::1234:5678:0/104`                                                                     | 集群 pod ipv6 网段（`K8S_IPv6_STATUS`不为`Disable`时有效）                                    |
| `K8S_MASK`         | `24`                                                                                        | 集群 pod ipv4 掩码长度                                                                        |
| `K8S_IPv6_MASK`    | `120`                                                                                       | 集群 pod ipv6 掩码长度                                                                        |
| `K8S_CNI`          | `flannel`                                                                                   | CNI 插件，现仅支持 `flannel`                                                                  |
| `ENABLE_MULTUS_HA` | `true`                                                                                      | [MULTUS_CNI](https://k8snetworkplumbingwg.github.io/multus-cni/docs/quickstart.html) 默认启用 |

#### CRI 环境变量

| 环境变量              | 默认值                      | 说明                                                       |
| --------------------- | --------------------------- | ---------------------------------------------------------- |
| `CRI_TYPE`            | `docker`                    | 容器运行时，支持`docker` 和 `containerd`                   |
| `INSECURE_REGISTRY`   | `""`                        | 信任的 registry                                            |
| `DOCKER_VER`          | `19.03.9`                   | 默认安装的 docker 版本                                     |
| `DOCKER_LIB`          | `${BK_HOME}/lib/docker`     | docker root path                                           |
| `DOCKER_LIVE_RESTORE` | `false`                     | docker config `live-restore`，重启后容器状态恢复，默认关闭 |
| `DOCKER_BRIDGE`       | `""`                        | docker 网桥设备                                            |
| `CONTAINERD_VER`      | `1.6.21`                    | 默认安装的 containerd 版本                                 |
| `CONTAINERD_LIB`      | `${BK_HOME}/lib/containerd` | containerd 根路径                                          |

#### CSI 环境变量

| 环境变量                | 默认值               | 说明                                           |
| ----------------------- | -------------------- | ---------------------------------------------- |
| `k8s_csi`               | `""`                 | CSI 选择，现仅支持 `localpv`                   |
| `localpv_dir`           | `${BK_HOME}/localpv` | localpv mount source，被挂载路径               |
| `localpv_dst_dir`       | `/mnt/blueking`      | localpv mount point，挂载路径                  |
| `localpv_count`         | `20`                 | localpv 创建的目录数（影响节点挂载的 PV 数量） |
| `localpv_reclaimpolicy` | `Delete`             | localpv 的 PV 回收策略，默认删除               |

#### apiserver ha 环境变量

| 环境变量                  | 默认值                      | 说明                                                                                                                                                                                                                                      |
| ------------------------- | --------------------------- |-----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `ENABLE_APISERVER_HA`     | `false`                     | apiserver ha 模式，默认关闭。                                                                                                                                                                                                                   |
| `APISERVER_HA_MODE`       | `bcs-apiserver-proxy`       | 模式选择，支持 [bcs-apiserver-proxy](https://github.com/TencentBlueKing/bk-bcs/blob/625be3183d99ee3500123016a6dea99d78165565/docs/features/bcs-apiserver-proxy/bcs-apiserver-proxy.md#L1), [kube-vip](https://kube-vip.io/), third-party(直接使用) |
| `VIP`                     |                             | VIP 地址，可配置与集群内不冲突的 ip 地址                                                                                                                                                                                                                |
| `VS_PORT`                 | `6443`                      | bap 代理端口                                                                                                                                                                                                                                |
| `APISERVER_PROXY_VERSION` | `v1.29.0-alpha.130-tencent` | bap 镜像版本                                                                                                                                                                                                                                |
| `PROXY_TOOL_PATH`         | `/usr/bin`                  | bap 工具安装目录                                                                                                                                                                                                                              |
| `PERSIST_DIR`             | `/root/.bcs`                | bap 持久化目录                                                                                                                                                                                                                               |
| `LVS_SCHEDULER`           | `rr`                        | bap 负载均衡策略                                                                                                                                                                                                                              |
| `MANAGER_INTERVAL`        | `10`                        | bap 监听时间                                                                                                                                                                                                                                |
| `DEBUG_MODE`              | `true`                      | bap DEBUG 模式 默认开启                                                                                                                                                                                                                       |
| `LOG_LEVEL`               | `3`                         | bap 日志等级                                                                                                                                                                                                                                |
| `KUBE_VIP_VERSION`        | `v0.5.12`                   | kube-vip 镜像版本                                                                                                                                                                                                                           |
| `BIND_INTERFACE`          | `""`                        | kube-vip 绑定网卡名                                                                                                                                                                                                                          |
| `VIP_CIDR`                | `32`                        | VIP CIDR 掩码长度                                                                                                                                                                                                                           |
| `EXTERNAL_VIP`            | `""`                        | 外部 LB ip                                                                                                                                                                                                                                |
| `EXTERNAL_HOST`           | `""`                        | 外部域名解析域名                                                                                                                                                                                                                                |

### 示例

#### a. 选择 `1.24.15` 的 k8s 版本 ，并使用 `containerd` 作为容器运行时

```bash
set -a
K8S_VER="1.24.15"
CRI_TYPE="containerd"
set +a
```

#### b. 创建 ipv6 双栈集群

> k8s 1.23 ipv6 特性为稳定版，仅支持 >=1.23.x 版本开启 ipv6 特性

```bash
set -a
K8S_VER="1.23.17"
K8S_IPv6_STATUS="DualStack"
set +a
./bcs-ops -i master
```

#### c. 修改镜像 registry，并信任

相关环境变量。镜像仓库默认为蓝鲸官方镜像仓库`hub.bktencent.com`，如果采用自己的镜像仓库，并且没有证书信任，需要添加下面两项环境变量

```bash
# 默认镜像地址
set -a
BK_PUBLIC_REPO=hub.bktencent.com
# 信任不安全的registry
INSECURE_REGISTRY=""
set +a
```

#### d. 离线安装

离线安装资源清单见 `env/offline-manifest.yaml`。

你需要把对应的离线包解压到 bcs-ops 的工作根目录下 `tar xfvz bcs-ops-offline-${version}.tgz`，并且安装对应的版本 `${VERSION}`。

```bash
set -a
BCS_OFFLINE="1"
K8S_VER="${VERSION}"
set +a
```

#### e. 开启 apiserver 高可用

APISERVER_HA_MODE 支持 [bcs-apiserver-proxy](https://github.com/TencentBlueKing/bk-bcs/blob/master/docs/features/bcs-apiserver-proxy/bcs-apiserver-proxy.md)（默认） 和 kube-vip。

```bash
set -a
VIP=192.168.1.1 # 按照实际的需求填写，避免冲突
ENABLE_APISERVER_HA=true
APISERVER_HA_MODE=bcs-apiserver-proxy
set +a
```

## k8s 插件

bcs-ops 脚本工具集也支持安装 k8s 相关插件。多数的插件需要通过 `helm` 的方式安装。因此，你需要在中控机上执行 `./bcs-ops --install helm`。

### csi

安装的 k8s 组件由 `K8S_CSI` 环境变量决定，默认为空，只支持 `localpv`

#### localpv

相关配置项，中控机启动前需要运行

```bash
# 申明 CSI 组件 为 `localpv`
K8S_CSI=localpv
# localpv 挂载点，默认为${BK_HOME}/localpv
LOCALPV_DIR=${LOCALPV_DIR:-${BK_HOME}/localpv}
# 创建的 localpv 数量，默认为20个
LOCALPV_COUNT=${LOCALPV_COUNT:-20}
# localpv 回收策略，默认为pvc删除后清理
LOCALPV_reclaimPolicy=${LOCALPV_reclaimPolicy:-"Delete"}
```

当 `K8S_CSI` 为 `localpv` 时。在部署的时候，将以挂载点进行自身绑定挂载，并把规则写入到 `/etc/fstab` 中，如下所示

```plaintext
${BK_HOME}/localpv/volxx ${BK_HOME}/localpv/volxx none defaults,bind 0 0
```

如果你需要安装 `localpv`，中控机执行：`./k8s/install_localpv`

### ingress-controller

#### nginx-ingress-controller

中控机执行 `bcs-ops/k8s/install_nginx_ingress.sh`
note: 默认 nodePort 为 32080 和 32443。不启用 hostNetwork 模式。

```yaml
service:
  type: NodePort
  nodePorts:
    http: 32080
    https: 32443
hostNetwork: false
```

---

# 编译打包

## 离线包制作

离线包的制作依赖命令工具 [yq](https://github.com/mikefarah/yq) 和 [skopeo](https://github.com/containers/skopeo)，请提前安装对应的工具。
制作 bcs-ops 所支持的离线包版本。

```bash
make build_offline_pkg
```

如果你只想制作对应版本的离线包（该版本应该在`env/offline-manifest.yaml`中出现）。

```bash
./offline_package.sh env/offline-manifest.yaml <verion>
```

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

# 标准运维操作

见 ["bcs-ops X bk-sops 使用文档.md"](sops/bcs-ops%20X%20bk-sops%20使用文档.md)
