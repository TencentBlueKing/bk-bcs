# 集群节点弹性伸缩（Cluster Autoscaler）

节点自动扩缩容功能基于 Kubernetes **Cluster Autoscaler（CA）** 组件，配合节点池（NodePool）动态调整集群计算节点数量，从而在业务高峰期自动补充算力，在低谷期回收空闲节点以节约成本。

---

## 一、工作原理与触发时机

- **自动扩容触发条件**：
  - 当集群中存在因计算资源（CPU / 内存等）不足而处于 `Pending` 不可调度状态的 Pod，且 K8S 原生调度算法判定现有节点无法容纳该 Pod 时，触发关联节点池扩容。
  - 若配置了小于 `100%` 的自定义扩容资源阈值，当集群资源分配率（基于 Pod `requests` 计算）达到或超过设定阈值时也会触发扩容。
- **自动缩容触发条件**：
  - 某节点处于空闲状态，或节点装箱率（CPU、内存 `requests` 分配率）低于设定缩容阈值（如 `50%`），且该节点上运行的 Pod 可以安全迁移调度到集群内其他节点。
  - 满足上述条件并持续达到稳定观察窗口后，执行节点缩容下架。
  - **缩容冷却期**：集群触发扩容后的 **20 分钟**内为冷却期，期间不会执行缩容操作，以防止节点频繁抖动。

> **核心说明**：集群弹性伸缩判断节点资源使用率时，是基于 Pod 声明的 **`resources.requests`（请求值）** 进行计算，而非节点的实际物理 CPU/内存负载。通常建议将 **节点自动扩缩容（CA）** 与 **Pod 水平自动扩缩容（HPA / GPA）** 结合使用。

---

## 二、前置条件

1. 集群为通过云凭证创建或导入的集群（支持腾讯云 TKE、谷歌云 GKE、AWS EKS、Azure AKS、华为云 CCE 等）或支持内部节点池的集群。
2. 云凭证必须具备管理容器集群（`tke:*`）、云服务器查询（`cvm:Describe*`）、私有网络查询（`vpc:Describe*`）及弹性伸缩组（`as:*`）的全量权限（详见[云凭证管理](../02-project-and-iam/cloud-credentials.md)）。

---

## 三、配置节点池与开启自动扩缩容

操作入口：进入【集群管理】→ 点击目标集群详情 → 切换到【弹性扩缩容 / 节点池】标签页。

### 1. 新建节点池（NodePool）
- 点击【新建节点池】，配置节点池名称、最大/最小节点数范围、可用区、子网、机型规格（CPU/内存）、系统盘/数据盘、节点标签与污点。
- 创建完成后，等待节点池状态变为**正常**。

### 2. 开启 Cluster Autoscaler 组件
- **强制要求**：集群下**必须至少存在一个状态为“正常”且已启用的节点池**，才允许开启 Cluster Autoscaler 组件。
- 点击【开启自动扩缩容】，按需调整扩缩容策略参数（如缩容阈值、稳定时间、扫描间隔等）并确认启用。

### 3. 编辑、关闭与删除节点池注意事项
- **编辑节点池**：支持修改节点池基本扩缩范围信息；受部分云厂商接口限制，若需更换底层机型配置，建议**新建一个新规格的节点池并关闭旧节点池**（若旧节点池中仍有存量节点，只需关闭自动扩缩容，切勿直接删除）。
- **关闭与删除限制**：
  - 当 Cluster Autoscaler 组件处于开启状态且仅剩 **1 个启用中的节点池**时，为保障 CA 组件正常运行，系统禁止直接关闭或删除该唯一节点池；如需关闭，请先停用 Cluster Autoscaler 组件。
  - 若节点池内仍存在运行中的节点，无法直接删除节点池，需先移除或缩容池内节点至 0。

---

## 四、快速验证节点自动扩缩容

假设当前集群有 1 台 `4核 8G` 的可用节点，可通过部署一个总 CPU Request 为 `5核` 的 Deployment 来模拟触发节点自动扩容：

### 1. 模拟触发扩容

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: ca-scaleup-test
  namespace: default
  labels:
    app: nginx
spec:
  replicas: 5
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
        - name: nginx
          image: nginx:latest
          ports:
            - containerPort: 80
          resources:
            requests:
              cpu: "1"
              memory: "256Mi"
```

下发上述 YAML 后，部分 Pod 将因 `Insufficient cpu` 处于 `Pending` 状态，Cluster Autoscaler 检测到不可调度 Pod 后将自动触发节点池扩容新节点。

### 2. 模拟触发缩容
- 删除上述测试工作负载：`kubectl delete deployment ca-scaleup-test -n default`。
- 等待超过 **20 分钟缩容冷却期** 后，观察空闲节点是否被自动回收下架，也可在节点池详情的【扩缩容记录】中查看每次扩缩容的触发原因与执行结果。
