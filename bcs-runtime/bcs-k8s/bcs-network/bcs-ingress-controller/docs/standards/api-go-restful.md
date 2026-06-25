# HTTP 管理 API 规范（go-restful）

<!--
  项目定制规范，基于 internal/httpsvr/ 实际实现编写。
  可贡献回 harness-engineering/assets/standards/ 预设库。
-->

> BCS Ingress Controller 内部管理 API 规范。路由注册于 `internal/httpsvr/httpserver.go` 的 `InitRouters()`。

---

## 一、架构概述

### 1.1 请求链路

```
HTTP Client
  → go-restful WebService（main.go 注册）
  → HttpServerClient 处理方法（internal/httpsvr/*.go）
  → Manager.GetClient() / 内存 Cache / 云适配器
  → CreateResponseData() 统一 JSON 响应
  → metrics.ReportAPIRequestMetric() 上报延迟
```

### 1.2 文件组织

| 文件 | 职责 |
|------|------|
| `httpserver.go` | `HttpServerClient` 结构体 + `InitRouters()` 路由注册 |
| `response.go` | `APIRespone` 结构与 `CreateResponseData*` 工厂 |
| `portpool.go` | PortPool / NodePortBinding 查询 |
| `ingress.go` | Ingress 列表 |
| `listener.go` | Listener 条件查询 |
| `node.go` | Node 信息查询 |
| `hostnetportpool.go` | HostNetPortPool 绑定结果 |
| `check_bind_status.go` | 绑定状态检查 |
| `aga_support.go` | AWS AGA 入口查询 |
| `readiness_probe.go` | 健康检查 |

**新增 API 规则**：处理器放独立 `{resource}.go`，路由在 `InitRouters()` 注册，响应走 `response.go`。

---

## 二、路由规范

### 2.1 现有路由清单

| 方法 | 路径 | 处理器 | 说明 |
|------|------|--------|------|
| GET | `/api/v1/ingresss` | `listIngress` | 列出所有 Ingress CRD（注意历史拼写 ingresss） |
| GET | `/api/v1/portpools` | `listPortPool` | 列出所有 PortPool |
| GET | `/api/v1/nodeportbindings` | `getNodePortBindings` | 查询节点端口绑定；Query: `nodes`（逗号分隔，可选） |
| GET | `/api/v1/listeners/{condition}/{namespace}/{name}` | `listListener` | 按条件查询 Listener |
| GET | `/api/v1/node` | `listNode` | 查询 Node；Query: `node_name` 或 `node_ip`（至少一个） |
| GET | `/api/v1/aga_entrance` | `getPodRelatedAgaEntrance` | AWS AGA 入口；Query: `pod_name`, `pod_namespace` |
| GET | `/api/v1/check_bind_status` | `CheckBindStatus` | 检查 Ingress 绑定状态 |
| GET | `/api/v1/hostnetportpool/bindingresult` | `getHostNetPortPoolBindingResult` | HostNet 绑定结果；Query: `podName`, `podNamespace` |
| GET | `/readiness_probe` | `readinessProbe` | 就绪探针（非 /api/v1 前缀） |

### 2.2 URL 设计规则

| 规则 | 说明 |
|------|------|
| 管理 API 前缀 | `/api/v1/`（readiness 除外） |
| 资源名词 | 小写复数或 snake_case（如 `nodeportbindings`、`aga_entrance`） |
| 路径参数 | 使用 go-restful `{param}` 语法 |
| 查询参数 | 使用 `request.QueryParameter("name")` |
| 历史兼容 | 已有拼写错误（`ingresss`）**禁止修改**，新 API 须正确拼写 |

### 2.3 新增路由模板

```go
// httpserver.go InitRouters()
ws.Route(ws.GET("/api/v1/{resources}").To(httpServerClient.listXxx))

// {resource}.go
func (h *HttpServerClient) listXxx(request *restful.Request, response *restful.Response) {
    startTime := time.Now()
    mf := func(status string) {
        metrics.ReportAPIRequestMetric("list_xxx", "GET", status, startTime)
    }
    // 业务逻辑 ...
    mf(strconv.Itoa(http.StatusOK))
    _, _ = response.Write(CreateResponseData(nil, "success", data))
}
```

---

## 三、响应格式规范

### 3.1 标准结构

所有 API 使用 `APIRespone`（注意历史拼写 Respone）：

```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `code` | int | 成功时为 `200`（`http.StatusOK`）；错误时为 HTTP 状态码 |
| `message` | string | 成功时通常为 `"success"` 或业务描述；错误时为 `err.Error()` |
| `data` | object/array/null | 业务数据；错误时为 `null` |

### 3.2 响应工厂函数

| 函数 | 场景 |
|------|------|
| `CreateResponseData(err, msg, data)` | 通用：err!=nil 时 code=500，否则 code=200 |
| `CreateResponseDataWithCode(code, err)` | 需指定 HTTP 状态码（如 404、400） |

**必须**通过 `response.go` 工厂函数构造响应，禁止在 handler 中手写 JSON。

### 3.3 错误响应约定

| HTTP code | 使用场景 | 构造方式 |
|-----------|---------|---------|
| 200 | 成功 | `CreateResponseData(nil, "success", data)` |
| 400 | 参数缺失/非法 | `CreateResponseData(err, "", nil)` 或 `CreateResponseDataWithCode(400, err)` |
| 404 | 资源不存在 | `CreateResponseDataWithCode(http.StatusNotFound, err)` |
| 500 | 内部错误 | `CreateResponseData(err, "", nil)`（默认 500） |

错误信息写入 `message` 字段，**禁止**在响应中暴露云凭证或 Secret 内容。

---

## 四、Handler 编写规范

### 4.1 必须步骤

1. 记录 `startTime`，定义 `mf` 闭包调用 `metrics.ReportAPIRequestMetric(handler, method, status, startTime)`
2. 校验 Query / Path 参数，缺失时返回 400 并调用 `mf`
3. 执行业务逻辑，错误时记录 `blog.Errorf` 并返回错误响应
4. 成功时 `mf(strconv.Itoa(http.StatusOK))` 后写入响应

### 4.2 参数校验示例

```go
podName := request.QueryParameter("podName")
podNamespace := request.QueryParameter("podNamespace")
if podName == "" || podNamespace == "" {
    mf(strconv.Itoa(http.StatusBadRequest))
    writeResponse(response, CreateResponseData(
        fmt.Errorf("empty parameter: both podName and podNamespace are required"), "", nil))
    return
}
```

### 4.3 K8s API 错误处理

```go
if k8serrors.IsNotFound(err) {
    mf(strconv.Itoa(http.StatusNotFound))
    writeResponse(response, CreateResponseDataWithCode(http.StatusNotFound, err))
    return
}
```

### 4.4 写入响应

优先 `response.Write(CreateResponseData(...))`；HostNetPortPool 等场景可用 `writeResponse()` 包装以统一日志。

---

## 五、Metrics 规范

每个 handler 须上报 API 指标：

```go
metrics.ReportAPIRequestMetric("<handler_name>", "GET", "<status_code>", startTime)
```

指标注册于 `internal/metrics/metric.go`：

- `bkbcs_ingressctrl_api_request_total{handler, method, status}`
- `bkbcs_ingressctrl_api_request_latency_seconds{handler, method, status}`

`handler_name` 使用 snake_case，与 Prometheus 标签保持一致（如 `list_port_pool`、`get_hostnet_binding_result`）。

---

## 六、数据类型

| 类型 | 序列化规则 |
|------|-----------|
| CRD List/Item | 直接序列化 K8s API 类型（json tag 由 CRD 定义决定） |
| 时间 | K8s 默认 RFC 3339 |
| 空列表 | 序列化为 `[]`，不使用 `null` |
| 缓存 Map | 按 handler 定义的结构返回 |

---

## 七、安全规范

| 规则 | 说明 |
|------|------|
| 输入校验 | 所有 Query/Path 参数必须校验非空和格式 |
| 无写操作 | 当前 API 均为 GET 只读查询，新增写操作须经过安全评审 |
| 日志脱敏 | `blog` 中不打印 Secret、云 AK/SK |
| 错误信息 | 对用户友好，不泄露内部堆栈 |

---

## 八、新增 API 检查清单

- [ ] 路由在 `InitRouters()` 注册
- [ ] 处理器方法挂载在 `HttpServerClient`
- [ ] 使用 `CreateResponseData` / `CreateResponseDataWithCode`
- [ ] 调用 `ReportAPIRequestMetric`
- [ ] 参数校验 + K8s NotFound 处理
- [ ] 错误路径均调用 `mf` 上报状态码
- [ ] 本规范路由清单表已同步更新

---

## 九、接口变更管理

| 规则 | 说明 |
|------|------|
| 禁止删除已发布路由 | 标记废弃并在文档注明替代方案 |
| 禁止修改路径拼写 | 已有 `ingresss` 等历史路径保持兼容 |
| 新增字段 | `data` 内结构向后兼容，新字段可选 |
| 文档同步 | 变更时更新本规范 §2.1 路由清单 |
