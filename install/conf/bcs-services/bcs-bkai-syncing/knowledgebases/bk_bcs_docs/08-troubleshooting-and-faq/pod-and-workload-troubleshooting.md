# Pod 与工作负载常见异常排查手册

当在蓝鲸容器管理平台（BCS）部署应用后，若发现工作负载未就绪（`Ready` 副本数小于期望数）或 Pod 处于非 `Running` 状态，可按照本手册按图索骥快速定位根因。

---

## 一、通用排障三板斧

在 BCS 控制台或 WebConsole 中，排查任何 Pod 异常的标准步骤如下：

1. **看状态（Status）**：进入【资源视图】→【工作负载】→【Pods】（或执行 `kubectl get pods -n <ns>`），明确 Pod 当前处于哪种异常状态（如 `Pending`、`ContainerCreating`、`ImagePullBackOff`、`CrashLoopBackOff`、`OOMKilled`、`Terminating`）。
2. **看事件（Events）**：
   - 在控制台点击异常 Pod 名称 → 切换到【事件（Events）】标签页，或前往【项目管理】→【事件查询】。
   - 或在 WebConsole 执行：`kubectl describe pod <pod_name> -n <ns>`，重点阅读最底部的 `Events` 列表（`Pending`、`ContainerCreating`、`ImagePullBackOff` 95% 的原因都在 Events 中）。
3. **看日志（Logs）**：
   - 若容器已启动但反复崩溃（`CrashLoopBackOff` / `Error`），在控制台查看容器标准输出日志，或执行 `kubectl logs <pod_name> -n <ns> --previous` 查看上一次崩溃退出前的报错堆栈。

---

## 二、各异常状态详细排查与解决指南

### 1. Pod 处于 `Pending` 状态（无法完成调度）

**含义**：Pod 已被创建到集群中，但 Kubernetes 调度器（`kube-scheduler`）找不到任何满足该 Pod 运行要求的节点。

**排查方法**：查看 Pod Events 中的 `FailedScheduling` 告警信息，常见原因及解决方式如下：

| Events 报错信息示例 | 根本原因 | 解决方法 |
| :--- | :--- | :--- |
| `0/5 nodes are available: 5 Insufficient cpu.` 或 `Insufficient memory.` | 集群所有节点的剩余可分配 CPU 或内存不足以满足该 Pod 声明的 `resources.requests` | 1. 调低工作负载过高的 `resources.requests`；<br/>2. 清理集群内闲置工作负载；<br/>3. 在【集群管理】中添加新节点或开启 Cluster Autoscaler 扩容。 |
| `0/5 nodes are available: 5 node(s) didn't match Pod's node affinity/selector.` | Pod 配置的 `nodeSelector` 或 `nodeAffinity`（节点亲和性）在集群中没有任何节点匹配对应标签 | 检查 Pod 的节点标签选择器拼写，或在【集群管理】→【节点列表】中为目标节点打上对应 Label。 |
| `0/5 nodes are available: 3 node(s) had untolerated taint {key: value}.` 或 `node(s) were unschedulable.` | 节点被打上了污点（Taints）或处于“停止调度（Cordon）”状态，而 Pod 未配置对应的容忍度（`tolerations`） | 在【节点列表】中检查节点是否处于停止调度状态并恢复调度，或为 Pod 添加对应的 `tolerations`。 |
| `0/5 nodes are available: 5 node(s) didn't have free ports for the requested pod ports.` | Pod 声明了固定 `hostPort`，而目标节点上的该宿主机端口已被其他 Pod 占用 | 取消固定 `hostPort` 改用 `Service` / `PortPool`，或改用 BCS `randhostport` 随机端口注入插件，或扩容更多节点。 |
| `pod has unbound immediate PersistentVolumeClaims` | Pod 挂载的 `PVC` 处于 `Pending` 未绑定状态 | 前往【资源视图】→【存储】→【PVC】查看该 PVC 未绑定的原因（如 StorageClass 名称错误、云盘配额不足或可用区无库存）。 |

---

### 2. Pod 处于 `ImagePullBackOff` 或 `ErrImagePull` 状态

**含义**：Pod 已调度到节点，但节点上的容器运行时无法从镜像仓库拉取指定的容器镜像。

**常见原因与解决方式**：
1. **镜像地址或 Tag 拼写错误 / 镜像不存在**：
   - 检查工作负载配置的 `image` 仓库域名、路径和 Tag 是否在制品库中真实存在。
2. **私有镜像仓库未配置拉取凭据（`401 Unauthorized` / `authentication required`）**：
   - 在目标命名空间下创建 `kubernetes.io/dockerconfigjson` 类型的 Secret，并在工作负载的 `spec.template.spec.imagePullSecrets`（或关联的 `ServiceAccount`）中引用该 Secret。
3. **跨架构镜像不匹配（`no matching manifest for linux/arm64 in the manifest list entries`）**：
   - Pod 调度到了 `arm64` 节点，但该镜像仅构建了 `amd64` 架构版本。请构建多架构 Manifest 镜像，或通过 `nodeSelector: {"kubernetes.io/arch": "amd64"}` 将 Pod 限制调度到对应架构节点。
4. **节点到镜像仓库网络不通或 HTTP 仓库未加信任**：
   - 检查节点 DNS 解析与防火墙策略；若使用自建 HTTP 仓库，需在节点 `/etc/docker/daemon.json`（或 containerd 配置）中添加 `insecure-registries`。

---

### 3. Pod 卡在 `ContainerCreating` 状态

**含义**：Pod 已调度到节点，正在执行挂载存储卷、分配 CNI 容器网络或执行 Webhook 注入，尚未启动容器进程。

**常见原因与解决方式**：
1. **ConfigMap / Secret / PVC 挂载失败（`MountVolume.SetUp failed`）**：
   - Pod 引用了不存在的 `ConfigMap`、`Secret` 或 `PVC` 名称（检查同命名空间下是否已创建对应配置或存储声明），或者 RWO 块存储卷正被另一个节点挂载占用（`Multi-Attach error`）。
2. **CNI 网络插件分配 IP 失败（`FailedCreatePodSandBox`）**：
   - 在 VPC-CNI 集群中可能是子网 Pod IP 耗尽或弹性网卡配额不足；在 Overlay 集群中检查节点 `flannel` / `calico` Pod 是否正常运行。
3. **端口池（PortPool）未就绪阻塞**：
   - 若使用了 `bcs-ingress-controller` 的端口池注解，检查关联的 `PortPool` 是否存在、状态是否为 `Ready`、可用端口是否已耗尽。

---

### 4. Pod 处于 `CrashLoopBackOff` 或 `Error` 状态

**含义**：容器镜像已成功拉取并启动，但容器内的主进程（PID 1）启动后很快异常退出（Exit Code 非 0），Kubelet 正在按指数退避策略反复重启容器。

**排查步骤**：
1. **查看崩溃日志**：执行 `kubectl logs <pod_name> -n <ns> --previous` 查看容器退出前的报错信息。
2. **查看退出码（Exit Code）**：执行 `kubectl describe pod <pod_name> -n <ns>` 查看 `Last State: Terminated` 中的 `Reason` 和 `Exit Code`：
   - **`Exit Code 137`（`Reason: OOMKilled`）**：容器实际内存使用量超过了 `resources.limits.memory` 限制，被 Linux 内核 OOM Killer 强制杀死。**解决**：调大工作负载的 `limits.memory`，或排查业务代码是否存在内存泄漏（如 JVM `-Xmx` 设置大于容器内存 Limit）。
   - **`Exit Code 1` / `2` / `127`**：业务代码抛错崩溃、缺少配置文件/环境变量，或 `command`/`args` 启动命令路径不存在（`127`）。
   - **`Exit Code 0`（正常退出但导致重启）**：Deployment 的容器主进程必须在前台常驻运行；若启动脚本将服务放到后台（如 `nginx &`）后脚本自身退出，容器会立即结束并在 `Always` 重启策略下进入 `CrashLoopBackOff`。**解决**：确保主进程在前台运行（如 `nginx -g "daemon off;"`）。
   - **存活探针（`livenessProbe`）失败导致被重启**：Events 中出现 `Liveness probe failed: ... Container will be restarted`。说明应用启动过慢或健康检查端口/路径配置有误。**解决**：调大 `initialDelaySeconds` / `failureThreshold`（或配置 `startupProbe`），并核对探针端口与路径。

---

### 5. Pod 处于 `Running` 但 `READY` 为 `0/1`（无法通过 Service 访问）

**含义**：容器进程正在运行，但未通过就绪检查，因此 K8S 不会将该 Pod 加入 Service 的 Endpoints 中接收流量。

**常见原因与解决方式**：
1. **就绪探针（`readinessProbe`）探测失败**：
   - 执行 `kubectl describe pod <pod_name> -n <ns>` 查看 `Readiness probe failed` 的具体原因（检查探针配置的端口号、HTTP Path 是否与容器内应用实际监听的一致）。
2. **ReadinessGate 未就绪**：
   - 若使用的是 `GameDeployment` / `GameStatefulSet`，检查是否正在执行原地更新 `gracePeriodSeconds` 摘流等待。
   - 若开启了 `bcs-ingress-controller` 端口池的 `readinessgate.portpools.networkextension.bkbcs.tencent.com: "true"`，检查 `kubectl get portbinding -n <ns>` 的状态是否已完成 CLB 绑定并变为 `Ready`。
