# 工作负载管理指南（Workloads）

在蓝鲸容器管理平台（BCS）的【资源视图】→【工作负载】模块中，用户可以集中管理集群中通过任何方式（Helm、模板文件、资源视图表单/YAML、`kubectl`）部署的所有原生工作负载及 Pod 实例。

---

## 一、支持的工作负载类型与选型

| 工作负载类型 | 适用场景 | 核心特点 |
| :--- | :--- | :--- |
| **Deployment** | 无状态微服务、Web/API 服务 | 最常用控制器；通过 ReplicaSet 管理 Pod 副本，支持声明式滚动升级、一键扩缩容、暂停/恢复发布及历史版本回滚。 |
| **StatefulSet** | 有状态服务（MySQL、Redis、Kafka、ZooKeeper、传统游戏服） | 提供稳定的网络标识（Pod 名称固定为 `<name>-0`, `<name>-1`，配合 Headless Service）、独立的持久化存储（`volumeClaimTemplates`）以及有序的部署/扩缩容/更新。 |
| **DaemonSet** | 节点级守护进程（日志采集 `bkunifylogbeat`、监控 `node-exporter`、网络组件） | 确保集群中每个节点（或匹配 `nodeSelector` 的节点）上精确运行一个 Pod 副本；新节点加入集群时自动部署，节点移除时自动回收。 |
| **Job** | 一次性批处理任务、数据库迁移脚本、离线计算 | 创建一个或多个 Pod 执行任务，直至指定数量的 Pod 成功退出（`Completed`）。 |
| **CronJob** | 周期性定时任务（定时备份、报表生成、巡检） | 基于标准 Crontab 时间表达式（如 `0 2 * * *`）按计划周期性创建 `Job` 运行。 |
| **Pod** | 最小调度与运行单元 | 展示所有控制器创建的 Pod 或独立 Pod，支持查看实时状态、容器详情、标准输出日志、事件及进入 WebConsole。 |

> **游戏与复杂业务扩展**：若您需要**容器原地重启（InplaceUpdate）**、**分步自动化灰度（Canary）**、**删除/更新前 Hook 校验**或**失联节点 Pod 自动漂移**，请参考 [GameDeployment 使用指南](../06-custom-addons-and-game/gamedeployment.md) 与 [GameStatefulSet 使用指南](../06-custom-addons-and-game/gamestatefulset.md)。

---

## 二、创建与查看工作负载

### 1. 创建工作负载（表单模式 / YAML 模式）
1. 进入【资源视图】→【工作负载】→ 选择具体类型（如 `Deployments`）。
2. 在顶部选择目标**集群**和**命名空间**，点击右上角【创建】按钮。
3. 支持两种创建模式自由切换：
   - **表单模式**：通过可视化界面配置基本信息（名称、副本数、更新策略）、容器配置（镜像、端口、启动命令、环境变量、CPU/内存 Request/Limit、存活/就绪探针 `livenessProbe`/`readinessProbe`）、数据卷挂载（PVC、ConfigMap、Secret、HostPath）及节点亲和性调度规则。
   - **YAML 模式**：直接编写或粘贴标准 Kubernetes YAML 清单并一键下发。

### 2. 查询与详情展示
点击任意工作负载名称进入详情页，可查看以下全方位信息：
- **工作负载概览**：当前就绪副本数（`Ready / Desired`）、更新策略、选择器（Selector）、标签与注解、关联的 Service 与历史版本记录。
- **Pod 实例列表**：展示该工作负载关联的所有 Pod 名称、状态（`Running` / `Pending` / `CrashLoopBackOff` 等）、Pod IP、所在节点 IP、重启次数（`Restarts`）与运行时长。
- **容器详情与实时日志**：点击具体 Pod 可查看内部各容器的镜像版本、状态原因、实时标准输出日志（支持下载与行数筛选）、监控曲线（CPU/内存/网络 IO）及 K8S 事件（Events）。

---

## 三、工作负载日常运维操作

在工作负载列表或详情页，支持执行以下常用运维操作：

### 1. 扩缩容（Scale）
- 点击目标工作负载的【扩缩容】按钮，直接调整期望副本数（`replicas`）。
- 调大副本数时立即创建新 Pod；调小副本数时优雅终止多余 Pod；设为 `0` 可暂停服务运行并释放计算资源。

### 2. 滚动升级与编辑更新
- 点击【编辑】（表单或 YAML）修改容器镜像版本或环境变量，提交后控制器将按照配置的 `updateStrategy`（如 `maxUnavailable` 与 `maxSurge`）分批使用新版本 Pod 替换旧版本 Pod，实现业务无中断升级。

### 3. 版本回滚（Rollback）
- 当升级后的新版本出现异常时，可在工作负载详情的【更新记录 / 历史版本】中对比不同 Revision 的 YAML 差异，并一键回滚至上一个稳定版本。

### 4. 重新调度 / 重建 Pod
- **工作负载级重建**：对工作负载触发滚动重启（类似 `kubectl rollout restart`），分批重建该负载下的全部 Pod。
- **单 Pod 重新调度**：在 Pod 列表中对某个异常 Pod 点击【重新调度】（删除该 Pod），工作负载控制器会自动根据副本数要求重新创建并调度一个新的 Pod。

### 5. 删除工作负载
- 点击【删除】将清理该工作负载控制器及其管理的所有 Pod（注意：删除 Deployment 不会自动删除其关联的 Service、ConfigMap 或 PVC）。
