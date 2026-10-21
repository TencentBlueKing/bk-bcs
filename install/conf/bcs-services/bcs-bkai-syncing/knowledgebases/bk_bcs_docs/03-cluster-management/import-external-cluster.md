# 导入外部集群指南

蓝鲸容器管理平台（BCS）支持将已存在的公有云集群（腾讯云 TKE、谷歌云 GKE、亚马逊云 EKS、微软云 AKS、华为云 CCE）或任意兼容原生 Kubernetes 的集群导入平台统一纳管。

---

## 一、前提条件

- 已登录蓝鲸平台，拥有目标项目的**项目查看**和**集群创建**权限。
- 当前用户在项目关联的 CMDB 业务中具备**业务运维**角色。

---

## 二、导入方式与操作步骤

操作入口：顶部导航栏【集群管理】→【添加集群】→【导入外部集群】。

### 方式 1：通过云凭证导入（推荐公有云集群使用）

以**腾讯云 TKE** 为例：

1. **配置云凭证**：提前在【项目管理】→【云凭证】中添加腾讯云 `SecretID` 与 `SecretKey`。
2. **填写导入信息**：
   - 在【导入外部集群】中选择【公有云集群】→【腾讯云】。
   - **集群名称与描述**：填写 BCS 侧展示的集群名称（建议使用「地域-用途-序号」规范命名）。
   - **云凭证**：选择已配置的腾讯云凭证。
   - **所属区域**：选择目标集群所在的云地域。
   - **云集群 ID**：下拉选择通过云凭证查询到的目标 TKE 集群（`cls-xxxx`）。
   - **集群访问网络**：选择 `内网访问` 或 `外网访问`（需确保 TKE 侧已开启对应 APIServer 访问入口，且 BCS 后台网络可达）。
   - **管控区域**：选择蓝鲸 GSE Agent 管控区域。
3. **可用性测试与导入**：
   - 点击【可用性测试】，验证 BCS 后台到目标集群 APIServer 的连通性与权限。
   - 测试通过后点击【导入】，等待异步初始化任务（自动部署 `bcs-k8s-watch` 等组件）完成。

### 方式 2：通过 Kubeconfig 文件导入（通用方式）

适用于任意自建 K8S 集群、其他云厂商集群或仅开放 Kubeconfig 的集群：

1. 在【导入外部集群】中选择【KubeConfig】。
2. 填写**集群名称**与**集群描述**。
3. **集群 KubeConfig**：上传 Kubeconfig 文件，或直接粘贴 YAML 格式的 Kubeconfig 内容（注意：Kubeconfig 中的 `server` 地址必须能从 BCS 服务端网络直接访问）。
4. 点击【可用性测试】，验证通过后点击【导入】。

---

## 三、最小权限纳管方案（受限账号 RBAC 配置）

默认情况下，使用具备 `cluster-admin` 权限的账号或 Kubeconfig 导入集群可获得最完整的管理体验。若因安全合规限制无法授予 `cluster-admin` 权限，可按以下**最小权限要求**配置子账号或 ServiceAccount：

### 1. 最小权限构成
1. **全集群所有资源的只读权限**（`get`、`list`、`watch`）：保障 BCS 控制台资源视图能正常展示命名空间、节点、Pod、工作负载等全量资源。
2. **`bcs-system` 命名空间的读写权限**：用于在集群内安装和运行 `bcs-k8s-watch`、`bcs-kube-agent` 组件（若需同步容器数据至 `bcs-storage` 和 CMDB 容器拓扑，此项必开）。

> **避坑提示**：腾讯云 TKE 预置的“只读用户”角色默认缺少集群级 `namespaces` 列表查询权限，直接用于纳管会报错：
> `cluster error, namespaces is forbidden: User "xxx" cannot list resource "namespaces" in API group "" at the cluster scope`。

### 2. 配置步骤（以 TKE 子账号为例）

**第一步：在集群中创建全集群只读 `ClusterRole`（`bcs-readonly`）与 `bcs-system` 命名空间**

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: bcs-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: bcs-readonly
rules:
  - apiGroups: ["*"]
    resources: ["*"]
    verbs: ["get", "list", "watch"]
```

**第二步：为纳管子账号绑定权限**
在腾讯云 TKE 控制台 →【运维中心】→【授权管理】→【新建授权管理】中，为用于纳管的子账号配置两条权限映射：
1. **Namespaces 列表**：选择 `所有 Namespaces` → **权限**：选择 `自定义`（关联上述创建的 `bcs-readonly`）。
2. **Namespaces 列表**：选择 `bcs-system` → **权限**：选择 `开发人员`（具备该命名空间下资源的读写权限）。

---

## 四、导入集群常见可用性测试报错排查

### 1. 报错 `【1405441】cluster endpoint status is not created`
- **原因**：目标 TKE 集群尚未开启 APIServer 的内网或外网访问入口（Endpoint）。
- **解决**：登录腾讯云 TKE 控制台 → 进入目标集群【基本信息】→ 在 **APIServer 信息** 处开启**外网访问**或**内网访问**。若开启外网访问，在 BCS 导入页面选择“外网访问”；若开启内网访问，选择“内网访问”。

### 2. 内网访问报错 `dial tcp 172.16.x.x:443: i/o timeout`
- **原因**：BCS 服务通过内网 IP 连接目标集群 APIServer 超时，说明 BCS 控制面所在网络与目标集群所在的 VPC 内网未打通。
- **解决**：
  - **方案 A（推荐）**：通过云联网（CCN）或 VPC 对等连接打通 BCS 所在网络与目标集群 VPC。
  - **方案 B**：在 TKE 侧开启 APIServer 公网访问，并在 BCS 导入页面切换为“外网访问”。

### 3. 外网访问报错 `dial tcp x.x.x.x:443: i/o timeout`
- **原因**：BCS 服务通过公网域名/IP 访问目标集群 APIServer 时被防火墙或安全组拦截。
- **解决**：检查 TKE 集群 APIServer 公网 CLB 绑定的**安全组入站规则（Inbound）**或访问白名单，放通 BCS 服务端出口 IP 对 **TCP 443 端口**的访问。
