# 存储管理指南（PV、PVC 与 StorageClass）

由于容器本身的文件系统是临时的，容器重建后内部写入的数据会丢失。对于有状态业务或需要持久化保存数据的场景，蓝鲸容器管理平台（BCS）在【资源视图】→【存储】模块中提供了 **PersistentVolumes（PV）**、**PersistentVolumeClaims（PVC）** 和 **StorageClass（存储类）** 的统一管理。

---

## 一、核心存储概念与关系

```mermaid
flowchart LR
    SC["StorageClass (存储类)<br/>定义底层存储插件与动态制备策略"] -->|动态自动创建| PV["PersistentVolume (PV)<br/>集群级实际存储卷资源"]
    PVC["PersistentVolumeClaim (PVC)<br/>命名空间级存储申请声明"] <-->|绑定 (Bound)| PV
    POD["Pod / 工作负载<br/>(Deployment / StatefulSet)"] -->|volumeMounts 挂载| PVC
```

1. **PersistentVolume（PV，持久卷）**：集群级别的存储资源对象（不归属于特定命名空间），对接底层云硬盘（如腾讯云 CBS）、文件存储（CFS/NFS）或分布式存储（Ceph）。可由管理员预先静态创建，也可由 `StorageClass` 动态自动创建。
2. **PersistentVolumeClaim（PVC，持久卷声明）**：命名空间级别的存储请求。应用开发者无需关心底层存储实现细节，只需声明所需的存储容量（如 `20Gi`）、访问模式（`accessModes`）和存储类名称（`storageClassName`）。
3. **StorageClass（SC，存储类）**：集群级别的动态存储供应模板，定义了 CSI 存储制备器（Provisioner）、磁盘类型（如高性能云盘、SSD 云盘）及卷回收策略（`reclaimPolicy`: `Delete` 或 `Retain`）。

---

## 二、访问模式（`accessModes`）说明

在创建 PV 或 PVC 时，需根据业务读写特征选择对应的访问模式：

| 访问模式 | 缩写 | 含义说明 | 常见底层存储支持 |
| :--- | :--- | :--- | :--- |
| **`ReadWriteOnce`** | `RWO` | 单节点读写：存储卷只能被挂载到**单个节点**上以读写方式使用（同一节点上的多个 Pod 可共享）。 | 云硬盘（CBS / EBS）、Ceph RBD 块存储 |
| **`ReadOnlyMany`** | `ROX` | 多节点只读：存储卷可同时被**多个节点**以只读方式挂载。 | NFS、CFS 文件存储、部分块存储 |
| **`ReadWriteMany`** | `RWX` | 多节点读写：存储卷可同时被**多个节点上的多个 Pod** 以读写方式并发挂载。 | 文件存储（腾讯云 CFS、NFS、CephFS） |

---

## 三、在 BCS 中使用持久化存储的标准流程

### 1. 查看集群可用的 StorageClass
进入【资源视图】→【存储】→【StorageClass】，查看当前集群已配置的存储类名称（如 `cbs`、`cfs`、`standard`）及哪一个是默认存储类（带有 `storageclass.kubernetes.io/is-default-class: "true"` 注解）。

### 2. 创建 PVC（持久卷声明）
进入【资源视图】→【存储】→【PersistentVolumeClaims】，选择集群与命名空间，点击【创建】：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-data-pvc
  namespace: default
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: cbs
  resources:
    requests:
      storage: 20Gi
```

创建后查看 PVC 状态，当状态变为 **`Bound`** 时表示已成功分配并绑定底层 PV（注意：若 StorageClass 的 `volumeBindingMode` 为 `WaitForFirstConsumer`，PVC 会保持 `Pending` 状态，直到挂载该 PVC 的第一个 Pod 被调度时才会完成绑定，属于正常现象）。

### 3. 在工作负载中挂载 PVC

#### 场景 A：Deployment 挂载已有 PVC

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: web-with-storage
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: web-with-storage
  template:
    metadata:
      labels:
        app: web-with-storage
    spec:
      containers:
        - name: nginx
          image: nginx:1.24.0
          volumeMounts:
            - name: data-vol
              mountPath: /var/www/html
      volumes:
        - name: data-vol
          persistentVolumeClaim:
            claimName: app-data-pvc
```

#### 场景 B：StatefulSet / GameStatefulSet 使用 `volumeClaimTemplates` 自动为每个副本创建独立 PVC

```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
  namespace: default
spec:
  serviceName: redis-headless
  replicas: 3
  selector:
    matchLabels:
      app: redis
  template:
    metadata:
      labels:
        app: redis
    spec:
      containers:
        - name: redis
          image: redis:7.0
          volumeMounts:
            - name: redis-data
              mountPath: /data
  volumeClaimTemplates:
    - metadata:
        name: redis-data
      spec:
        accessModes: ["ReadWriteOnce"]
        storageClassName: cbs
        resources:
          requests:
            storage: 10Gi
```
在此模式下，K8S 会自动为 `redis-cluster-0`、`redis-cluster-1`、`redis-cluster-2` 分别创建名为 `redis-data-redis-cluster-0` 等独立的 PVC，即使 Pod 重建或漂移，仍会挂载回原有的持久化数据卷。
