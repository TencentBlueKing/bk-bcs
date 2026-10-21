# 新建 K8S 原生集群（私有化自建）

当业务运行在私有化机房或不使用公有云容器服务时，可通过蓝鲸容器管理平台（BCS）在自有 Linux 服务器上自动化搭建高可用 Kubernetes 原生集群并完成纳管。

---

## 一、前提条件与环境准备

### 1. 权限与业务要求
- 已登录蓝鲸平台，拥有目标项目的**项目查看**和**集群创建**权限。
- 项目已关联 CMDB 业务，且当前用户在该 CMDB 业务中具备**业务运维**角色。

### 2. 主机与节点要求
- **GSE Agent 状态**：所有待选用的 Master 与 Worker 节点必须在蓝鲸【节点管理】中安装好 **GSE Agent**，且 Agent 状态为**正常**，并处于所绑定 CMDB 业务的**空闲机**模块下。
- **硬件与系统要求**：
  - **Master 节点**：必须为奇数台（生产高可用要求 **3 台或 5 台**，测试环境支持 1 台）；配置至少 **4 核 CPU / 8GB 内存**。
  - **Worker 节点**：需确保内核已加载 `NAT` 模块。
  - **操作系统**：推荐 CentOS 7.4 及以上版本（Linux 内核版本 `3.10.0-693` 及以上）或平台适配的国产化操作系统。

### 3. 标准运维（SOPS）公共流程检查
BCS 自建 K8S 原生集群底层通过调用蓝鲸**标准运维（bk-sops）**的公共流程完成节点初始化与组件部署。建群前需确保标准运维的【公共流程管理】中存在以下预置流程，且流程 ID 与 `bcs-cluster-manager` 配置严格对应：

| 默认流程 ID | 标准运维公共流程名称 | 对应集群操作类型 |
| :--- | :--- | :--- |
| `10` | `【BCS】bcsops distribute` | 运维脚本分发基础子流程 |
| `11` | `【BCS】Setup Kubernetes on Linux` | 新建集群（`createcluster`） |
| `12` | `【BCS】Add Kubernetes Worker` | 添加集群节点（`addnodestocluster`） |
| `13` | `【BCS】Remove Kubernetes Worker` | 移除集群节点（`deletenodesfromcluster`） |
| `14` | `【BCS】Destroy Cluster` | 销毁/删除集群（`deletecluster`） |
| `15` | `【BCS】安装bcs-kube-agent` | 安装集群代理子流程 |
| `16` ~ `18` | `【BCS】K8S master replace` / `etcd restore` / `etcd backup` | 控制面替换与 Etcd 备份恢复维护流程 |

---

## 二、新建 K8S 原生集群步骤

进入顶部导航栏【集群管理】→ 点击【添加集群】→ 在新建集群分类下选择【K8S 原生集群】，按四步向导完成配置：

### 第一步：填写基本信息
- **集群信息**：
  - **Kubernetes 提供商**：默认 `Kubernetes 社区`。
  - **集群名称**：填写 1~64 字符的集群名称。
  - **集群版本**：选择平台支持的社区 K8S 版本（如 `1.20.15`、`1.23.17`、`1.24.15` 等）。
  - **标签与描述**（可选）：设置集群键值标签与用途说明。
- **集群环境**：选择**所属区域**、**管控区域（云区域）**及**集群环境**（测试环境 / 正式环境）。
- **集群配置**：
  - **操作系统**：选择目标服务器操作系统（如 `centos`）。
  - **运行时组件与版本**：选择容器运行时（推荐使用 `containerd` 如 `1.6.21`，或选择 `docker`）。

### 第二步：填写网络配置
- **基本配置**：
  - **私有网络**：默认 `自定义网络`。
  - **默认网络插件**：支持 `Flannel(Overlay)`。
  - **集群 IP 类型**：支持 `IPv4`；Kubernetes 1.23 及以上版本支持 `IPv4/IPv6` 双栈。
- **容器网络分配**（请确保网段不与机房物理网络冲突）：
  - **Service IP (Service CIDR)**：默认如 `10.96.0.0/12`。
  - **Pod IP (Pod CIDR)**：默认如 `10.244.0.0/16`。
  - **单节点 Pod 数量上限**：可选 `32`、`64`、`128`、`256`。

### 第三步：填写控制面（Master）配置
- **资源选择**：点击【+ 添加服务器】，从关联 CMDB 业务主机拓扑中选择 **3 台或 5 台**（高可用）服务器（控制面节点不足 3 台将无法支持高可用）。
- **Kube-apiserver 高可用方案**：
  - 默认内置 **`bcs-apiserver-proxy`**：以 DaemonSet 形式部署在集群各节点，利用内核 IPVS 通过本地虚拟 IP（VIP）实现集群内组件到多台 Master APIServer 的负载均衡与自动健康探测，确保单台 Master 故障时集群正常运转。

### 第四步：填写节点（Worker）配置与提交
- **添加节点**：从 CMDB 业务主机拓扑中选择 Worker 服务器。
  > **提示**：创建集群时可选择**跳过此步骤**，待 Master 控制面创建并纳管成功后，再通过【节点列表】→【添加节点】分批扩展计算节点。
- **确认创建**：核对配置无误后点击【创建集群】。可在集群列表中点击【查看日志】实时跟踪标准运维任务进度。

---

## 三、自建原生集群常见问题（FAQ）

### Q1：导入的标准运维公共流程 ID 与默认 ID 不一致导致建群报错怎么办？
若手动导入 `bcs_sops_common_bcs.yaml` 后生成的标准运维流程 ID 与默认的 `11`~`14` 不一致，有两种修复方式：

**方法 A（修改 BCS MongoDB 中的 `template_id` 映射）**：
连接 BCS MongoDB，进入 `clustermanager` 库更新 `bcsclustermanagerv2_cloud` 集合：

```javascript
use clustermanager
db.bcsclustermanagerv2_cloud.updateOne(
  {"cloudid": "bluekingCloud"},
  {"$set": {
    "clustermanagement.createcluster.plugins.createTask.params.template_id": "实际Setup流程ID",
    "clustermanagement.deletecluster.plugins.createTask.params.template_id": "实际Destroy流程ID",
    "clustermanagement.addnodestocluster.plugins.createTask.params.template_id": "实际AddWorker流程ID",
    "clustermanagement.deletenodesfromcluster.plugins.createTask.params.template_id": "实际RemoveWorker流程ID"
  }}
)
```

**方法 B（在标准运维 MySQL 数据库中对齐预留 ID）**：
提前备份标准运维数据库后，将流程 ID 更新为预留的 `10`~`15`：

```sql
UPDATE template_commontemplate SET id = 11 WHERE id = <old_id>;
```

### Q2：新建集群初始化失败，Job 报错 `failed to pull image xxx not found`？
- **原因**：初始化 K8S 控制面时需拉取 `kube-apiserver`、`kube-scheduler`、`coredns` 等基础镜像。若默认配置的私有镜像仓库中缺少这些镜像，会导致拉取失败。
- **解决**：可将缺失镜像同步至私有仓库，或修改 `bcs-cluster-manager` 的 MongoDB 配置指向具备完整镜像的仓库地址：

```javascript
use clustermanager
db.bcsclustermanagerv2_cloud.updateOne(
  {"cloudid": "bluekingCloud"},
  {"$set": {
    "platforminfo.image_registry": "hub.bktencent.com"
  }}
)
```

### Q3：安装 `bcs-kube-agent` 步骤报错 `job_fail: [get in ip fail]`？
- **原因**：节点网卡名称不是传统的 `eth0` / `eth1`（例如为 `ens192`、`eno1`、`enp3s0` 等），旧版脚本仅匹配 `eth` 导致获取内网 IP 失败。
- **解决**：
  1. 在标准运维中打开子流程 `【BCS】安装bcs-kube-agent`，将获取内网 IP 的脚本替换为兼容多种网卡名的命令：
     ```bash
     in_ip=$(ip addr show | grep -E '^[0-9]+: (eth|ens|eno|enp)' -A2 | grep 'inet ' | awk '{print $2}' | cut -d/ -f1 | head -1)
     ```
  2. **重要**：保存子流程后，还必须打开父流程 `【BCS】Setup Kubernetes on Linux` 更新子流程引用版本并保存，然后在 BCS 页面点击重试。

### Q4：安装 `bcs-kube-agent` 报错 `no chart version found for bcs-k8s-stack-xxx`？
- **原因**：`bcsclustermanagerv2_cloud` 中配置的 `bcs_agent_chart_version` 版本号在当前 Helm 仓库（`bkrepo_helm_url`）中不存在。
- **解决**：
  1. 查询当前 Helm 仓库中实际存在的 `bcs-k8s-stack` Chart 版本：
     ```bash
     helm search repo blueking/bcs-k8s-stack -l
     ```
  2. 将 MongoDB 中的 `platforminfo.bcs_agent_chart_version` 更新为仓库中存在的版本号（如 `1.29.1`），随后在页面重试：
     ```javascript
     use clustermanager
     db.bcsclustermanagerv2_cloud.updateOne(
       {"cloudid": "bluekingCloud"},
       {"$set": {
         "platforminfo.bcs_agent_chart_version": "1.29.1"
       }}
     )
     ```
