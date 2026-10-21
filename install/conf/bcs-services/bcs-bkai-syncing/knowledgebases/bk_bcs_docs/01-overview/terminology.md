# 核心术语与概念解释

在使用蓝鲸容器管理平台（BCS）过程中，会涉及以下核心概念与平台专有术语：

---

## 一、平台管理概念

| 术语 | 英文/标识 | 说明 |
| :--- | :--- | :--- |
| **项目** | `Project` (`projectID` / `projectCode`) | BCS 的一级资源组织单元。所有集群、Chart 仓库、模板文件、变量和云凭证均归属于特定项目。启用容器服务时需绑定一个蓝鲸配置平台（CMDB）业务。 |
| **CMDB 业务** | `Business` (`businessID` / `bk_biz_id`) | 蓝鲸配置平台中的业务对象。BCS 项目绑定 CMDB 业务后，可从该业务的主机资源池中获取服务器用于创建集群或添加节点，并将容器拓扑同步至该业务。 |
| **集群 ID** | `Cluster ID` (`BCS-K8S-xxxxx`) | BCS 平台自动生成的集群全局唯一标识符（例如 `BCS-K8S-40001`）。在调用 BCS API、查询监控指标（`cluster_id`）及配置 Kubeconfig 时均使用此 ID。 |
| **云集群 ID** | `systemID` | 云服务商（如腾讯云 TKE）侧生成的原生集群唯一标识（如 `cls-xxxxxxxx`）。用于将 BCS 集群与公有云控制台中的集群进行精确映射。 |
| **云凭证** | `Cloud Token` / `Account` | 用于访问公有云 API 的身份凭据（如腾讯云 SecretID/SecretKey、谷歌云 Service Account JSON），在导入或创建公有云集群时使用。 |
| **共享集群** | `Shared Cluster` | 由平台管理员建设并向多个 BCS 项目开放的公共 K8S 集群。各项目可在共享集群中申请专属命名空间或创建虚拟集群（vCluster）以实现资源复用与隔离。 |

---

## 二、Kubernetes 基础概念

- **容器编排引擎（Kubernetes / K8S）**：用于容器化应用自动化部署、扩缩容和管理的开源编排系统。
- **集群（Cluster）**：容器运行所需的控制面（Master）与工作节点（Node）计算、存储、网络资源的集合。
- **节点（Node）**：已注册到 Kubernetes 集群中的物理机或虚拟机服务器，运行 `kubelet`、`containerd`/`docker` 等组件，为 Pod 提供计算资源。
- **命名空间（Namespace）**：在一个物理 K8S 集群内划分多个“虚拟逻辑集群”的隔离机制，常用于按环境（dev/test/prod）、模块或团队划分资源与配额。
- **Pod**：Kubernetes 中创建和调度的最小执行单元。一个 Pod 包含一个或多个紧密耦合、共享网络命名空间（同一 Pod IP）和存储卷的容器。
- **工作负载（Workload）**：用于管理 Pod 副本生命周期的控制器对象，常见类型包括：
  - **Deployment**：管理无状态应用，支持滚动更新、扩缩容与版本回滚。
  - **StatefulSet**：管理有状态应用，为每个 Pod 提供稳定的网络标识（固定序号名称如 `web-0`、`web-1`）和独立的持久化存储（PVC）。
  - **DaemonSet**：确保集群中全部（或匹配特定标签的）节点上都运行一个 Pod 副本，常用于日志采集、监控 Agent、网络插件。
  - **Job / CronJob**：分别用于一次性批处理任务和按 Crontab 时间表周期性执行的定时任务。
- **配置与密钥（ConfigMap & Secret）**：
  - **ConfigMap**：将非敏感配置数据（配置文件、命令行参数、环境变量）与容器镜像解耦。
  - **Secret**：存储敏感信息（密码、OAuth Token、TLS 证书、镜像仓库拉取凭据）。
- **服务与网络（Service & Ingress）**：
  - **Service**：通过 Label Selector 将一组 Pod 抽象为稳定的访问入口（支持 `ClusterIP`、`NodePort`、`LoadBalancer`、`Headless` 类型）。
  - **Ingress**：管理外部 HTTP/HTTPS 流量访问集群内部 Service 的路由规则（基于域名 Host、路径 Path 及 TLS 证书）。
- **持久卷与声明（PV & PVC & StorageClass）**：
  - **PV（PersistentVolume）**：集群中的一块底层网络或本地存储资源。
  - **PVC（PersistentVolumeClaim）**：用户对存储资源的申请声明，Pod 通过挂载 PVC 使用持久化存储。
  - **StorageClass**：定义动态存储供应策略，使 PVC 能按需自动创建并绑定 PV。

---

## 三、部署与 BCS 扩展概念

- **Helm Chart**：Kubernetes 应用的软件包格式（类似 CentOS 的 RPM/Yum 包），包含渲染 K8S 资源所需的模板文件与默认参数 `values.yaml`。
- **Helm Release**：Helm Chart 在 Kubernetes 集群中安装运行的实例。同一个 Chart 可在同一集群或不同命名空间中安装多次，每次安装都会生成一个独立的 Release。
- **模板文件（Template File）/ 模板集**：BCS 提供的原生 YAML 模板管理机制，支持将多个 K8S 资源配置组织在模板空间中，结合 BCS 变量（`{{ .KEY }}`）进行版本管理与一键下发。
- **GameDeployment / GameStatefulSet**：BCS 针对游戏与复杂应用自研的增强型 CRD 工作负载，扩展了容器原地重启（InplaceUpdate）、镜像热更新（HotPatchUpdate）、分步自动化灰度（Canary）、优雅删除/更新钩子（PreDeleteHook / PreInplaceHook）及防误删等高级能力。
- **HookTemplate / HookRun**：由 `bcs-hook-operator` 提供的发布生命周期钩子模板与运行实例，支持在扩缩容、更新或灰度过程中调用外部 Webhook、查询 Prometheus 指标或 Patch K8S 资源注解。
- **GeneralPodAutoscaler（GPA）**：BCS 自研的通用弹性伸缩器，兼容原生 HPA 并新增 Crontab 定时伸缩、Webhook 自定义伸缩和事件驱动伸缩。
- **bcs-ingress-controller / PortPool**：BCS 自研的多云四层/七层负载均衡控制器，支持通过 `Ingress` CRD 控制云 CLB（NodePort 或直通 Pod），并通过 `PortPool` 端口池为单个 Pod 动态分配公网/内网负载均衡端口。
- **bcs-kube-agent & bcs-k8s-watch**：
  - `bcs-kube-agent`：部署在被纳管集群中的代理组件，通过 WebSocket 隧道将集群 APIServer 安全接入 BCS 控制面（跨云区域或 NAT 环境下必需）。
  - `bcs-k8s-watch`：部署在被纳管集群中的资源采集组件，实时监听集群 K8S 资源变更并上报至 `bcs-storage`，供控制台视图、事件查询和 CMDB 拓扑同步消费。
