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
2. 集群加入指令只保留，中控机执行 `./bcs-ops --render joincmd` 再次渲染生成加入集群的指令
3. 添加控制平面节点，待添加的节点上按实际渲染执行

```bash
set -a
CLUSTER_ENV=xxx
MASTER_JOIN_CMD=xxx
set +a
./bcs-ops --install master`
```

4. 添加工作平面节点，待添加的节点上按实际渲染执行

```bash
set -a
CLUSTER_ENV=xxx
MASTER_JOIN_CMD=xxx
set +a
./bcs-ops --install node`
```

### 集群 node 节点移除

1. 在中控机上先移除节点

```bash
node_name="node-$(tr ":." "-" <<<"$ip")"
# https://kubernetes.io/zh-cn/docs/tasks/administer-cluster/safely-drain-node/
kubectl drain --ignore-daemonsets $node_name
kubectl delete node $node_name
```

2. 被移除的节点上执行 `./bcs-ops --clean node`

### 中控机安装 helm 工具
`./bcs-ops --install helm`

### 部署 localpv
> 注意：在 添加 node 节点的过程中，并没有挂载 localpv 的目录。localpv 部署依赖 helm。localpv 默认寻找 `/mnt/blueking`下的挂载点。

1. node 节点 执行 `./system/mount_localpv`。该工具会在`/data/bcs/localpv` 目录下创建 20 个子目录，并挂载到对应的`/mnt/blueking/localpv`路径下。

2. 中控机执行 `./bcs-ops --install localpv`

3. 当步骤 2 执行后，新的加入的 node 节点如果要添加 `Persistentvolumes`，先执行步骤 1，后执行步骤 2，即可重启 localpv 的 pod 实现挂载。
