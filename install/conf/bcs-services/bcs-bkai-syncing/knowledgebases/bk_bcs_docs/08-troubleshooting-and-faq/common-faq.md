# 蓝鲸容器管理平台（BCS）常见问题汇总（FAQ）

本文汇总了用户在日常使用蓝鲸容器管理平台（BCS）进行项目管理、权限申请、集群纳管、Helm 部署及扩展组件配置时最常遇到的问题与快速解答。

---

## 一、项目与权限类问题

### Q1：为什么进入 BCS 后看不到任何项目，或者点击项目提示无权限？
- **解答**：BCS 采用严格的蓝鲸权限中心（IAM）授权机制。
  - 若需加入已有项目：直接点击项目名称，在弹出的权限申请框中点击【去申请】，申请该项目的 **`项目查看`** 权限。
  - 若需创建新项目：需先在蓝鲸权限中心申请容器管理平台（`bk_bcs_app`）的 **`项目创建`** 权限。

### Q2：启用项目容器服务时，下拉框里为什么找不到我的 CMDB 业务？
- **解答**：出于安全管控，BCS 仅允许绑定当前登录用户在**蓝鲸配置平台（CMDB）**中拥有**“业务运维”**角色的业务。请联系该 CMDB 业务的运维负责人在配置平台将您的账号加入“业务运维”角色后再刷新页面绑定。

### Q3：使用 Kubeconfig 或 API 密钥执行 `kubectl` 时报错 `Forbidden` 怎么办？
- **解答**：API 密钥代表您个人的蓝鲸身份，其操作权限完全由**蓝鲸权限中心（IAM）**控制。请登录蓝鲸权限中心，为您的账号申请对应集群的 **`集群资源查看` / `集群资源管理`** 或对应命名空间的 **`命名空间下资源查看` / `命名空间下资源管理`** 权限。

---

## 二、集群创建与导入类问题

### Q4：通过“云凭证”导入集群和通过“Kubeconfig”导入集群有什么区别？该如何选择？
- **解答**：
  - **通过云凭证导入（推荐公有云集群使用）**：除了能管理集群内的 K8S 资源外，还支持在 BCS 页面直接**添加/删除集群节点**、**管理节点池（NodePool）**、**开启节点自动扩缩容（Cluster Autoscaler）**及联动云负载均衡器。
  - **通过 Kubeconfig 导入**：兼容性最强，支持接入任何符合社区标准的 K8S 集群，但**仅能管理集群内部 K8S 资源**，不支持在 BCS 页面直接扩缩容底层机器节点。

### Q5：如何确认 BCS 上的集群与腾讯云 TKE 控制台上的集群是哪一个对应关系？
- **解答**：切勿仅凭“集群名称”判断（因为两侧名称均支持自定义修改）。请查看 BCS 集群详情页中的 **TKE 集群 ID（底层字段为 `systemID`，格式如 `cls-72b5p8ho`）**，该 ID 与腾讯云 TKE 控制台上的集群 ID **100% 唯一对应且不可修改**。

### Q6：为什么在集群节点列表中，“删除节点”按钮是灰色的无法点击？
- **解答**：为防止直接删除正在运行业务 Pod 的节点造成线上故障，BCS 强制要求**必须先对节点执行【停止调度】（Cordon）**，建议再执行 **【Pod 驱逐】（Drain）** 将业务 Pod 迁移至其他节点后，【删除节点】按钮才会亮起可用。

### Q7：为什么无法关闭或删除节点池（NodePool）？
- **解答**：
  1. 若节点池内仍有存量节点，必须先删除池内节点使节点数归零后才能删除节点池。
  2. 若当前已开启 **Cluster Autoscaler（节点自动扩缩容）** 且该节点池是集群下唯一启用中的节点池，为保证扩缩容组件正常工作，需先停用 Cluster Autoscaler 组件后才能关闭或删除该节点池。

---

## 三、Helm 与应用部署类问题

### Q8：推送 Helm Chart 到 BCS 项目仓库时报 `409 Conflict (already exists)` 或 `411 Length Required` 怎么办？
- **解答**：
  - **`409` 错误**：BCS 仓库禁止覆盖已存在的同版本 Chart。请修改 `Chart.yaml` 中的 `version` 版本号（如从 `0.1.0` 递增为 `0.1.1`）后重新打包推送。
  - **`411` 错误**：通常是因为本地终端配置了不兼容的 `http_proxy` / `https_proxy` 代理。请执行 `unset http_proxy https_proxy` 关闭代理后重试。

### Q9：修改了 ConfigMap 或 Secret 后，为什么业务容器里的配置没有生效？
- **解答**：
  - 如果容器是通过**环境变量（`env` / `envFrom`）**或 **`volumeMounts.subPath`（单文件挂载）**方式引用 ConfigMap/Secret，Kubernetes **不会**自动热更新容器内的值。您需要在修改 ConfigMap/Secret 后，前往【工作负载】对该 Deployment/StatefulSet 执行**滚动重启 / 重建**以使新配置生效。
  - 如果是通过整个目录 Volume 挂载（未使用 `subPath`），Kubelet 会在约 1~2 分钟内自动同步文件变更。

### Q10：`GameDeployment` / `GameStatefulSet` 配置了 `InplaceUpdate`（原地更新），为什么修改配置后没有触发更新或卡住了？
- **解答**：
  - **原地更新（`InplaceUpdate`）仅支持更新容器镜像（`image`）以及 `labels` / `annotations` 字段**。
  - 如果您修改了容器环境变量（`env`）、端口（`ports`）、启动命令（`command`）、资源配额（`resources`）或挂载卷，原地更新无法处理这些结构变更（在 `GameStatefulSet` 中会导致更新卡住）。修改非镜像字段时，请将 `updateStrategy.type` 切换为 **`RollingUpdate`**。

### Q11：为什么删除 `GameDeployment`、`GameStatefulSet` 或 `HookTemplate` 时被拒绝删除？
- **解答**：BCS 游戏负载默认开启了 Admission Webhook **防误删保护**。
  - 如确认需要删除，请先为该资源添加 Label：`io.tencent.bcs.dev/deletion-allow: Always`（或推荐先缩容 `replicas: 0` 并设置 `io.tencent.bcs.dev/deletion-allow: Cascading`），然后再执行删除。
