# HTTP API Contract: HostNetPortPool 端口分配结果查询

**Feature Branch**: `001-hostnet-port-allocation`  
**Date**: 2026-03-16  
**Base Path**: `/ingresscontroller`  
**Service**: `bcsingresscontroller.bcs-system.svc.cluster.local:18088`

## 接口列表

### GET /api/v1/hostnetportpool/bindingresult

查询指定 Pod 的 HostNetPortPool 端口段分配结果。供 Pod 中的 initContainer 轮询使用。

#### 请求

| 参数 | 位置 | 类型 | 必填 | 说明 |
|------|------|------|------|------|
| `podName` | query | string | 是 | Pod 名称 |
| `podNamespace` | query | string | 是 | Pod 所在 namespace |

**示例请求**:
```
GET /ingresscontroller/api/v1/hostnetportpool/bindingresult?podName=game-server-1&podNamespace=game-ns
```

#### 响应

复用现有 `APIResponse` 结构（`internal/httpsvr/response.go` 中的 `CreateResponseData`）。

##### 成功 — 分配完成（status=Ready）

```json
{
  "code": 0,
  "message": "",
  "data": {
    "status": "Ready",
    "result": {
      "poolName": "game-server-ports",
      "poolNamespace": "game-ns",
      "nodeName": "node-1",
      "startPort": 30000,
      "endPort": 30029,
      "segmentLength": 10
    }
  }
}
```

##### 成功 — 尚未分配（status=NotReady）

```json
{
  "code": 0,
  "message": "",
  "data": {
    "status": "NotReady",
    "result": null
  }
}
```

##### 成功 — 分配失败（status=Failed）

```json
{
  "code": 0,
  "message": "",
  "data": {
    "status": "Failed",
    "result": null
  }
}
```

##### 错误 — Pod 不存在或未使用 HostNetPortPool

HTTP 200，但 `code` 为非零值：

```json
{
  "code": 404,
  "message": "pod game-ns/game-server-1 not found or not using HostNetPortPool",
  "data": null
}
```

##### 错误 — 缺少参数

```json
{
  "code": 400,
  "message": "empty parameter: both podName and podNamespace are required",
  "data": null
}
```

#### 响应字段说明

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 0 表示成功，非零表示错误 |
| `message` | string | 错误信息（成功时为空字符串） |
| `data.status` | string | `Ready` / `NotReady` / `Failed` |
| `data.result.poolName` | string | HostNetPortPool CRD 名称 |
| `data.result.poolNamespace` | string | HostNetPortPool CRD 命名空间 |
| `data.result.nodeName` | string | Pod 所在 Node |
| `data.result.startPort` | int | 分配的起始端口（闭区间） |
| `data.result.endPort` | int | 分配的结束端口（闭区间） |
| `data.result.segmentLength` | int | 段长度 |

#### initContainer 使用示例

```yaml
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
```

## Kubernetes Events 契约

### Pod Events

Controller 在以下场景记录 Event 到 Pod：

| Event Type | Reason | Message 格式 | 触发条件 |
|------------|--------|-------------|----------|
| Normal | HostNetPortAllocated | `Allocated port segment {startPort}-{endPort} on node {nodeName}` | 分配成功 |
| Warning | AllocateHostNetPortFailed | `No available segment on node {nodeName} from pool {poolKey}` | 端口段不足 |
| Warning | AllocateHostNetPortFailed | `HostNetPortPool {poolKey} not found` | Pool CR 不存在 |
| Normal | HostNetPortReleased | `Released port segment {startPort}-{endPort} on node {nodeName}` | 端口段释放 |

### HostNetPortPool CR Events

Controller 在以下场景记录 Event 到 HostNetPortPool CR 上（替代 `Status.Conditions`，因 SDK 版本 `k8s.io/apimachinery v0.18.6` 不支持 `metav1.Condition`）：

| Event Type | Reason | Message 格式 | 触发条件 |
|------------|--------|-------------|----------|
| Warning | PoolShrinkConflict | `Cannot shrink port range: {count} segment(s) in use: node={nodeName} ports={startPort}-{endPort} pod={podKey}` | 缩小端口范围时存在已分配的冲突段 |
| Normal | PoolShrinkResolved | `Port range shrink completed successfully` | 冲突段全部释放后缩小成功 |
