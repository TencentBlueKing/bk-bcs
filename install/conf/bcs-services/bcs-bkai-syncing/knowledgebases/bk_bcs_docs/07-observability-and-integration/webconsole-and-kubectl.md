# WebConsole 在线终端与 `kubectl` 常用命令速查

**WebConsole** 是蓝鲸容器管理平台（BCS）内置的浏览器端在线命令行终端服务，无需用户在本地安装 `kubectl` 或配置 Kubeconfig，即可在页面上一键打开集群管理终端或直接进入指定业务容器内部排障。

---

## 一、WebConsole 的两种使用模式

### 1. 集群级 `kubectl` 终端（Cluster WebConsole）
- **入口**：在 BCS 控制台任意页面右下角点击悬浮的 **【WebConsole】** 按钮（或在【集群管理】列表点击对应集群的【WebConsole】），选择目标集群。
- **原理与权限**：系统会在后台动态启动一个预装好 `kubectl` 与当前用户临时凭证的管控 Pod，约 1~2 秒即可进入终端。您在终端内可执行的操作范围严格受限于您在蓝鲸权限中心（IAM）中拥有的集群与命名空间权限。

### 2. 容器级内部终端（Container Exec WebConsole）
- **入口**：进入【资源视图】→【工作负载】→【Pods】→ 点击目标 Pod 名称进入详情页 → 在容器列表中点击右侧的 **【WebConsole】**。
- **作用**：相当于执行 `kubectl exec -it <pod_name> -c <container_name> -n <namespace> -- /bin/sh`（或 `/bin/bash`），直接进入业务容器内部查看进程状态、配置文件或网络连通性。

---

## 二、`kubectl` 常用查询与排障命令速查手册

在集群 WebConsole 或本地终端中，可使用以下高频 `kubectl` 命令进行日常查询与故障诊断：

### 1. 查看集群与节点状态

```bash
# 查看集群 APIServer 及基础服务地址信息
kubectl cluster-info

# 查看集群所有节点状态、角色、版本
kubectl get nodes

# 查看节点的详细IP、所在宿主机内核、容器运行时等信息
kubectl get nodes -o wide

# 查看某个节点的详细资源分配水位（Allocated resources）、污点（Taints）和异常事件（Events）
kubectl describe node <node_name>

# 查看各节点的实时 CPU / 内存物理消耗（需部署 metrics-server）
kubectl top nodes
```

### 2. 查询命名空间与工作负载状态

```bash
# 查看集群所有命名空间
kubectl get ns

# 查看指定命名空间下的所有核心资源概览（Pod、Service、Deployment、StatefulSet、DaemonSet）
kubectl get all -n <namespace>

# 查看指定命名空间下的 Pod 列表（含分配的 Pod IP 和所在节点 Node）
kubectl get pods -n <namespace> -o wide

# 跨所有命名空间查找异常（非 Running / Completed）状态的 Pod
kubectl get pods -A --field-selector=status.phase!=Running,status.phase!=Succeeded

# 按标签（Label）筛选 Pod
kubectl get pods -n <namespace> -l app=nginx

# 查看 BCS 游戏扩展负载（GameDeployment / GameStatefulSet / HookRun）
kubectl get gamedeployment,gamestatefulset,hookrun -n <namespace>

# 查看 Pod 的实时 CPU / 内存资源消耗
kubectl top pods -n <namespace>
```

### 3. 深入排查异常 Pod（Describe 与 Logs）

```bash
# 查看 Pod 详情与底部 Events 事件（排查 Pending、ImagePullBackOff、MountVolume 失败的首选命令）
kubectl describe pod <pod_name> -n <namespace>

# 查看 Pod 标准输出日志（若 Pod 内有多个容器，使用 -c <container_name> 指定容器）
kubectl logs <pod_name> -n <namespace> --tail=200

# 实时跟踪（流式输出）Pod 日志
kubectl logs -f <pod_name> -n <namespace>

# 【关键排障命令】查看刚刚崩溃重启前的上一个容器实例日志（专门排查 CrashLoopBackOff）
kubectl logs <pod_name> -n <namespace> --previous

# 以 YAML 或 JSON 格式导出 Pod 的完整定义与状态
kubectl get pod <pod_name> -n <namespace> -o yaml
```

### 4. 查看网络、配置与存储资源

```bash
# 查看 Service 与关联的实际后端 Endpoints IP 列表
kubectl get svc,ep -n <namespace>

# 查看原生 Ingress 与 BCS 扩展 Ingress / PortPool / PortBinding
kubectl get ingress -n <namespace>
kubectl get ingress.networkextension.bkbcs.tencent.com,portpool,portbinding -n <namespace>

# 查看 ConfigMap、Secret 与 PVC 绑定状态
kubectl get cm,secret,pvc -n <namespace>

# 查看命名空间下按时间排序的最新 K8S 事件
kubectl get events -n <namespace> --sort-by='.lastTimestamp'
```
