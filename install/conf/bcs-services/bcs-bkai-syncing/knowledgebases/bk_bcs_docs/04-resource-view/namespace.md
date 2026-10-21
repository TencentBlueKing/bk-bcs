# 命名空间与配额管理

在 Kubernetes 集群中，**命名空间（Namespace）**用于在同一个物理集群内划分多个相互隔离的“虚拟逻辑集群”，常用于按团队、微服务模块或环境（如 `dev`、`test`、`stag`、`prod`）隔离资源与权限。

---

## 一、管理命名空间

操作入口：进入顶部导航栏【资源视图】→【命名空间】（或在【集群管理】下管理命名空间）。

### 1. 创建命名空间
1. 选择目标集群，点击【新建】按钮。
2. 填写命名空间配置：
   - **名称**：最长 63 个字符，仅允许使用**小写字母、数字和连字符 `-`**，且必须以字母或数字开头和结尾。
   - **配额设置（ResourceQuota，可选）**：限制该命名空间下所有 Pod 可申请的计算资源上限：
     - `requests.cpu` / `limits.cpu`（CPU 请求与限制总额，如 `16` 核）
     - `requests.memory` / `limits.memory`（内存请求与限制总额，如 `32Gi`）
   - **标签（Labels）与注解（Annotations）**：按需配置（例如启用 `bcs-ingress-controller` 端口池注入时，需添加标签 `bcs-ingress-controller-inject=true`）。
   - **命名空间变量**：为该命名空间设置专属变量值，供 Helm Chart 或模板文件渲染时引用。
3. 点击【确定】完成创建。

> **注意**：命名空间创建后，**名称不支持直接修改**。如需“重命名”，只能新建目标命名空间并迁移资源后再删除旧命名空间。

### 2. 设置命名空间变量值
在【部署管理】→【变量管理】中定义了作用范围为“命名空间变量”的变量后，可在命名空间列表中点击目标命名空间右侧的【设置变量值】，为不同命名空间赋予不同的变量值（如不同环境的数据库地址、域名、副本数）。

### 3. 删除命名空间（高危操作）
- 删除命名空间将**级联删除该命名空间下的所有 Kubernetes 资源**（包括工作负载、Pod、Service、Ingress、ConfigMap、Secret、PVC 等）。
- 执行删除前请务必确认该命名空间下已无在线业务运行。

---

## 二、常见命名空间异常排查

### 命名空间删除后长期卡在 `Terminating` 状态无法清理怎么办？
- **常见原因**：
  1. 命名空间下仍存在带有 `finalizers`（终结器）的自定义资源（如未清理的 `PortPool`、`HookTemplate`、`GameDeployment` 等）。
  2. 集群内某个扩展 APIService（如 `metrics-server` 或自定义 Webhook）异常不可用，导致 K8S Namespace 控制器无法完成资源枚举。
- **排查步骤**：
  1. 查看命名空间详情中的 `status.conditions`，明确是哪个资源或 APIService 阻塞了删除：
     ```bash
     kubectl get ns <namespace_name> -o yaml
     ```
  2. 清理残留的 CRD 资源或修复异常的 APIService。
