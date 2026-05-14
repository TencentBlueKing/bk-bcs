# Quickstart: HostNetwork 动态端口分配

**Feature Branch**: `001-hostnet-port-allocation`

## 前置条件

- Kubernetes 集群运行正常
- BCS IngressController 已部署（Service: `bcsingresscontroller.bcs-system.svc.cluster.local:18088`）
- HostNetPortPool CRD 已注册到集群

## 步骤 1：创建 HostNetPortPool

```yaml
apiVersion: networkextension.bkbcs.tencent.com/v1
kind: HostNetPortPool
metadata:
  name: game-server-ports
  namespace: game-ns
spec:
  startPort: 30000
  endPort: 30100
  segmentLength: 10
```

```bash
kubectl apply -f hostnetportpool.yaml
```

验证：

```bash
kubectl get hostnetportpool game-server-ports -n game-ns -o yaml
# 确认 status.status 为 Ready
```

## 步骤 2：部署 hostNetwork Pod

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: game-server-1
  namespace: game-ns
  annotations:
    hostnetportpool.networkextension.bkbcs.tencent.com: "game-server-ports"
spec:
  hostNetwork: true
  dnsPolicy: ClusterFirstWithHostNet
  volumes:
    - name: shared-data
      emptyDir: {}
  initContainers:
    - name: wait-for-ports
      image: busybox:1.36
      env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: POD_NAMESPACE
          valueFrom:
            fieldRef:
              fieldPath: metadata.namespace
      command:
        - sh
        - -c
        - |
          echo "Waiting for port allocation..."
          while true; do
            RESULT=$(wget -qO- "http://bcsingresscontroller.bcs-system.svc.cluster.local:18088/ingresscontroller/api/v1/hostnetportpool/bindingresult?podName=$POD_NAME&podNamespace=$POD_NAMESPACE" 2>/dev/null)
            STATUS=$(echo "$RESULT" | grep -o '"status":"[^"]*"' | head -1 | cut -d'"' -f4)
            if [ "$STATUS" = "Ready" ]; then
              echo "Port allocation ready: $RESULT"
              START_PORT=$(echo "$RESULT" | grep -o '"startPort":[0-9]*' | cut -d: -f2)
              END_PORT=$(echo "$RESULT" | grep -o '"endPort":[0-9]*' | cut -d: -f2)
              echo "$START_PORT" > /shared/start_port
              echo "$END_PORT" > /shared/end_port
              break
            fi
            echo "Status: $STATUS, retrying in 2s..."
            sleep 2
          done
      volumeMounts:
        - name: shared-data
          mountPath: /shared
  containers:
    - name: game-server
      image: game-server:latest
      volumeMounts:
        - name: shared-data
          mountPath: /shared
```

```bash
kubectl apply -f pod.yaml
```

## 步骤 3：验证分配结果

```bash
# 查看 Pod annotation
kubectl get pod game-server-1 -n game-ns -o jsonpath='{.metadata.annotations}' | jq .

# 预期输出包含：
# "hostnetportpool.result.networkextension.bkbcs.tencent.com": "{\"poolName\":\"game-server-ports\",...,\"startPort\":30000,\"endPort\":30009,...}"
# "hostnetportpool.status.networkextension.bkbcs.tencent.com": "Ready"
```

```bash
# 通过 HTTP API 查询
kubectl exec -n game-ns game-server-1 -c game-server -- \
  wget -qO- "http://bcsingresscontroller.bcs-system.svc.cluster.local:18088/ingresscontroller/api/v1/hostnetportpool/bindingresult?podName=game-server-1&podNamespace=game-ns"
```

## 步骤 4：请求多段分配（可选）

如果 Pod 需要超过 segmentLength 个端口，添加 `portcount` annotation：

```yaml
metadata:
  annotations:
    hostnetportpool.networkextension.bkbcs.tencent.com: "game-server-ports"
    hostnetportpool.portcount.networkextension.bkbcs.tencent.com: "25"
```

Controller 会分配 ceil(25/10) = 3 个连续段（30 个端口）。

## 步骤 5：查看端口池使用情况

```bash
kubectl get hostnetportpool game-server-ports -n game-ns -o yaml
# 查看 status.nodeAllocations 了解各 Node 的使用情况
```

## 常见问题

**Q: Pod 一直处于 Init 状态？**
- 检查 `dnsPolicy` 是否设置为 `ClusterFirstWithHostNet`
- 检查 IngressController Service 是否可达
- 如果 hostNetwork Pod 无法解析 Service DNS 名称，改用 Service ClusterIP 直连：
  ```bash
  # 获取 ClusterIP
  kubectl get svc <ingress-controller-svc> -n bcs-system -o jsonpath='{.spec.clusterIP}'
  # initContainer 中将 DNS 名称替换为 ClusterIP
  ```
- 检查 Pod annotation 中 HostNetPortPool 名称是否正确
- 查看 `kubectl describe pod` 中的 Events

**Q: 分配状态为 Failed？**
- 检查 HostNetPortPool CRD 是否存在于正确的 namespace
- 检查端口池是否有足够的可用段

**Q: 分配状态为 NotReady？**
- 检查 Pod 是否已调度到 Node（`spec.nodeName` 非空）
- 检查目标 Node 上是否有足够的连续空闲段
- 查看 `kubectl describe pod` 中的 Warning Events（包含碎片化诊断信息）

## 注意事项

### 1. Controller 架构

Controller 由三个独立的 Reconciler 组成，分别处理不同类型的资源：
- **HostNetPortPoolReconciler**: 处理 HostNetPortPool CR 的生命周期（Finalizer、端口范围变更）
- **PodReconciler**: 处理 Pod 的端口分配和回收
- **NodeReconciler**: 处理 Node 删除时的缓存清理

这种设计避免了使用特殊 Namespace 前缀（如 `__node__`）来判断事件类型，符合 controller-runtime 最佳实践。

### 2. 幂等性检查机制

Controller 使用**内存 Cache**而非 Pod Annotation 来判断 Pod 是否已分配端口。原因如下：
- APIServer 在负载较高时，Annotation 更新可能有延迟
- 依赖 Annotation 判断可能导致竞态条件和重复分配
- Cache 查询是实时的，不受 APIServer 延迟影响

**注意**: 如果通过 `kubectl edit` 等方式手动修改 Pod Annotation，Controller 不会感知。应始终通过删除并重建 Pod 来触发重新分配。

### 3. 非法 portcount 处理

如果 `portcount` annotation 存在以下情况，Controller 会**直接报错**（status=Failed），不会使用默认值：
- 无法解析为整数（如非数字字符）
- 小于或等于 0 的值

错误时 Pod 会记录 Warning Event，例如：
```
Warning  InvalidPortCount  portcount annotation "abc" is invalid: invalid syntax
```

使用前请确保 `portcount` 是正整数，或省略该 annotation（此时默认分配 1 个段）。

### 4. 端口池缩小冲突监控

当 HostNetPortPool 的端口范围缩小时，如果被移除范围内存在已分配的段，Controller 会拒绝缩小并记录：
- **Warning Event** (Reason: `PoolShrinkConflict`): 列出冲突的 Node、端口范围和 Pod
- **Counter 指标** `hostnet_pool_shrink_conflict_total`: 可用于 Prometheus 告警

查询冲突指标：
```bash
# 查看 Controller 的 metrics 端点
curl http://<controller-pod>:8080/metrics | grep hostnet_pool_shrink_conflict_total

# PromQL 告警规则示例
increase(hostnet_pool_shrink_conflict_total[5m]) > 0
```

**解决冲突**: 删除占用冲突段的 Pod，Controller 会自动检测到冲突解除并执行缩小操作。
