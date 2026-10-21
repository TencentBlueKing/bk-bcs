# API 密钥与 Kubeconfig 使用指南

BCS API 密钥（User Token）是用户通过命令行（`kubectl` / `helm`）或外部程序调用 BCS 网关及 Kubernetes 原生 API 的身份凭证。

---

## 一、适用场景

- **本地或运维机使用 `kubectl`**：配置 `kubeconfig` 文件，在中心运维机或本地终端直接管理 BCS 纳管的任意 K8S 集群（包括跨云区域通过 `bcs-kube-agent` 隧道接入的集群）。
- **调用 Kubernetes 原生 REST API**：在不便安装 `kubectl` 的脚本或自动化平台中，通过 HTTP 请求直接调用集群 K8S API。
- **调用 BCS 平台 OpenAPI**：通过蓝鲸 API 网关或 BCS API 网关查询/操作项目、集群、命名空间与发布资源。

---

## 二、创建 API 密钥与申请权限

### 1. 创建个人 API 密钥

1. 登录 BCS 控制台，点击右上角个人头像或导航菜单进入 **API 密钥** 页面（路径格式：`http(s)://{bcs_domain}/{project_code}/api-key`）。
2. 点击【新建 API 密钥】按钮。
3. 选择**申请期限**：
   - **永久**：长期有效，适合固定运维机配置。
   - **临时（指定天数）**：到期后密钥自动失效，适合短期排障或临时授权。
4. 点击【确定】完成创建。

> **安全提醒**：请妥善保管生成的 API 密钥 Token，切勿将其明文提交至公开代码仓库或发送给未授权人员。

### 2. 在蓝鲸权限中心申请对应权限

API 密钥本身仅代表您的用户身份，实际可操作的集群与命名空间取决于您在**蓝鲸权限中心（IAM）**中拥有的权限：

1. 在 API 密钥页面点击【权限中心】链接，进入蓝鲸权限中心。
2. 点击【申请权限】→【申请自定义权限】，选择系统 **容器管理平台（`bk_bcs_app`）**。
3. 按需选择权限：
   - **只读查看集群资源**：勾选 `集群查看`、`集群资源查看`、`命名空间下资源查看`。
   - **读写管理集群资源**：勾选 `集群管理`、`集群资源管理`、`命名空间下资源管理`。
4. 选择要关联的**项目、集群或命名空间实例**（支持开启【批量编辑】申请整个项目的权限），提交并完成审批。

---

## 三、配置 Kubeconfig 使用 `kubectl`

### 1. 安装 `kubectl` 命令行工具

若运维机尚未安装 `kubectl`，可从 [Kubernetes 官方下载页](https://kubernetes.io/zh-cn/docs/tasks/tools/) 下载对应平台二进制文件并赋予执行权限：

```bash
chmod +x /usr/bin/kubectl
kubectl version --client
```

### 2. 编写 Kubeconfig 配置文件

将以下内容保存为 `~/.kube/config`（或自定义路径如 `/root/.kube/bcs_config`），并将变量替换为实际值：

- `${bcs_api_gateway}`：BCS API 网关地址（可从控制台 API 密钥页面的使用示例中直接复制）。
- `${cluster_id}`：目标集群 ID（如 `BCS-K8S-40001`）。
- `${username}`：您的蓝鲸用户名。
- `${api_key_token}`：刚刚创建的 BCS API 密钥字符串。

#### 独立集群 / 标准集群 Kubeconfig 模板

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: 'https://${bcs_api_gateway}/clusters/${cluster_id}/'
  name: '${cluster_id}'
contexts:
- context:
    cluster: '${cluster_id}'
    user: '${username}'
  name: BCS
current-context: BCS
users:
- name: '${username}'
  user:
    token: '${api_key_token}'
```

#### 共享集群 Kubeconfig 模板

若访问的是共享集群，Server 地址需包含项目英文名 `${project_code}`：

```yaml
apiVersion: v1
kind: Config
clusters:
- cluster:
    server: 'https://${bcs_api_gateway}/projects/${project_code}/clusters/${cluster_id}/'
  name: '${cluster_id}'
contexts:
- context:
    cluster: '${cluster_id}'
    user: '${username}'
  name: BCS
current-context: BCS
users:
- name: '${username}'
  user:
    token: '${api_key_token}'
```

### 3. 执行 `kubectl` 命令验证

```bash
# 若保存在默认路径 ~/.kube/config
kubectl get nodes
kubectl get ns

# 若保存在自定义路径
kubectl --kubeconfig=/root/.kube/bcs_config get pods -n default
```

---

## 四、直接调用 Kubernetes 原生 API

在不具备 `kubectl` 的环境中，可在请求头中携带 `Authorization: Bearer ${API密钥}` 直接通过 HTTPS 调用 BCS 网关代理的 K8S 原生接口：

### 1. 获取集群 Kubernetes 版本

```bash
curl -X GET \
  -H "Authorization: Bearer ${API密钥}" \
  -H "Accept: application/json" \
  "https://${bcs_api_gateway}/clusters/${cluster_id}/version"
```

### 2. 获取集群命名空间列表

```bash
curl -X GET \
  -H "Authorization: Bearer ${API密钥}" \
  -H "Accept: application/json" \
  "https://${bcs_api_gateway}/clusters/${cluster_id}/api/v1/namespaces"
```

### 3. 获取指定命名空间下的 Deployment 列表

```bash
curl -X GET \
  -H "Authorization: Bearer ${API密钥}" \
  -H "Accept: application/json" \
  "https://${bcs_api_gateway}/clusters/${cluster_id}/apis/apps/v1/namespaces/default/deployments"
```

### 4. 获取集群节点资源指标（Metrics API）

```bash
curl -X GET \
  -H "Authorization: Bearer ${API密钥}" \
  -H "Accept: application/json" \
  "https://${bcs_api_gateway}/clusters/${cluster_id}/apis/metrics.k8s.io/v1beta1/nodes"
```

---

## 五、常见问题（FAQ）

### Q1：第三方平台或自动化脚本如何申请 API 密钥？
- **个人实名凭证**：目前在控制台创建的 API 密钥绑定个人实名账号，权限跟随个人在蓝鲸权限中心的授权。
- **蓝鲸应用态调用（推荐系统间集成）**：若通过蓝鲸 API 网关（`bcs-api-gateway`）进行系统间调用，建议使用蓝鲸应用身份（`bk_app_code` + `bk_app_secret`）在蓝鲸 API 网关申请对应接口权限，以实现应用级免用户态或网关代持鉴权。
