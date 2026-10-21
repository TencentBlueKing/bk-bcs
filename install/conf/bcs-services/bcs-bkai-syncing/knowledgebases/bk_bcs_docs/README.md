# 蓝鲸容器管理平台（BCS）文档知识库

本知识库提供蓝鲸容器管理平台（BlueKing Container Service，简称 BCS）的产品概述、项目管理与权限配置、集群纳管与运维、Kubernetes 资源视图、Helm 与模板部署、游戏云原生扩展组件、监控日志可观测性及常见故障排查的完整指南。

---

## BCS 资源查询说明

通过平台查询工具查询蓝鲸容器服务（BCS）资源信息时，沿用调用方身份与权限。

### 1. 可查询的对象

- **项目（Project）**：当前用户有权限的项目列表、项目详情（项目 ID `projectID`、项目英文名 `projectCode`、关联 CMDB 业务 `businessID`、是否启用容器服务等）。
- **集群（Cluster）**：指定项目下的集群列表、共享集群列表、集群详情（集群 ID `clusterID`、集群状态、K8S 版本、所属云厂商与地域、网络配置等）。
- **命名空间（Namespace）**：指定集群中的命名空间列表、配额与变量信息。
- **部署管理（Helm & 模板）**：
  - Helm Chart 仓库列表、Chart 列表、Chart 版本及详情。
  - Helm Release 发布列表、发布详情与发布历史记录。
  - 模板空间（文件夹）、模板文件列表、模板版本及部署预览信息、项目变量列表。
- **工作负载与集群资源**：集群中的工作负载（如 Deployment 等）状态、集群组件及其详情。
- **监控与审计**：集群 Prometheus 指标查询（即时查询 `query`、范围查询 `query_range`、序列与标签查询）、项目操作记录（Activity Logs）。

### 2. 查询步骤

1. **确认目标对象**：明确要查询的是项目、集群、命名空间、Helm Release、Chart、模板文件、监控指标还是操作记录。
2. **获取项目标识**：查询集群、命名空间、模板、Helm 发布或操作记录前，需要项目 ID（`project_id`）或项目英文名（`project_code`）。用户未提供时先确认，不要猜测。
3. **获取集群标识**：查询命名空间、工作负载、集群详情、Helm Release 或监控指标（PromQL 标签需包含 `cluster_id`）时，还需要集群 ID（格式通常为 `BCS-K8S-xxxxx`）。
4. **结果说明**：只根据返回的字段说明状态与配置。没有返回结果时，说明尚未查到。

### 3. 限制

- 不要执行删除、安装、升级、回滚、导入集群或创建 WebConsole 会话。
- 不要索取或复述应用密钥（SecretKey / AppSecret）、用户 Token、API 密钥或 Kubeconfig 私钥内容。

---

## 文档目录结构

| 目录 | 模块主题 | 核心内容 |
| :--- | :--- | :--- |
| `01-overview/` | 产品概述与快速入门 | 产品简介与优势、产品架构、核心术语解释、快速入门全流程 |
| `02-project-and-iam/` | 项目管理与权限指南 | 项目创建与绑定业务、IAM 权限申请指南、多云凭证配置、API 密钥与 Kubeconfig 使用、事件查询与操作审计 |
| `03-cluster-management/` | 集群管理与纳管运维 | 集群纳管能力总览、新建 K8S 原生集群、新建 TKE 公有云集群、导入外部集群、跨云区域与 ARM/x86 异构集群纳管、节点管理、节点池弹性伸缩（CA） |
| `04-resource-view/` | K8S 资源视图（Dashboard） | 命名空间管理、工作负载运维（Deployment/StatefulSet/DaemonSet/Job/CronJob/Pod）、网络（Service/Ingress）、配置（ConfigMap/Secret）、存储（PV/PVC/StorageClass）、HPA 与 RBAC |
| `05-deploy-management/` | 部署管理 | Helm Chart 仓库与推送、Helm Release 部署/升级/回滚、Helm 内置变量与技巧、模板文件（Template File）与变量管理 |
| `06-custom-addons-and-game/` | 扩展组件与游戏负载 | GameDeployment、GameStatefulSet、bcs-hook-operator 与分步灰度发布、GeneralPodAutoscaler（GPA）、bcs-ingress-controller 与 PortPool 端口池 |
| `07-observability-and-integration/` | 可观测性与周边集成 | 蓝鲸容器监控与 PromQL 查询、容器日志采集（BcsLogConfig）、WebConsole 在线终端、bcs-k8s-watch（RabbitMQ 订阅）与 CMDB 容器拓扑同步 |
| `08-troubleshooting-and-faq/` | 故障排查与 FAQ | Pod 与工作负载常见异常排查（Pending/CrashLoopBackOff/ImagePullBackOff 等）、集群纳管与 Helm 部署常见 FAQ |
