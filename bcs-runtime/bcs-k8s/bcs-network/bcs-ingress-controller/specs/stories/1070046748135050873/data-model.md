# Data Model: 证书过期时间 Prometheus 指标

**Feature**: `1070046748135050873`
**Date**: 2026-06-09

---

## 核心实体

### 1. CertificateBinding（逻辑实体，非 CRD）

表示 Ingress 上一处 SSL 证书挂载关系，由 Checker 每轮从 Ingress CR 展开生成。

| 字段 | 类型 | 说明 | 指标 label |
|------|------|------|-----------|
| OwnerNamespace | string | Ingress 所在 Namespace | `owner_namespace` |
| OwnerName | string | Ingress 名称 | `owner_name` |
| CertID | string | 腾讯云证书 ID（certID 或 certCaID） | `cert_id` |
| CertRole | string | `server` 或 `client_ca` | `cert_role` |
| CertScope | string | `rule` / `route` / `port_mapping` | `cert_scope` |
| Protocol | string | `HTTPS` / `TCP_SSL` / `QUIC` | `protocol` |
| Port | string | 监听端口（字符串化） | `port` |
| Domain | string | SNI 域名；非 route scope 为 `""` | `domain` |

**生成规则**:

```
FOR each Ingress IN IngressList:
  IF protocol NOT IN (HTTPS, TCP_SSL, QUIC): SKIP

  FOR each cert block IN (rules[].certificate, rules[].layer7Routes[].certificate, portMappings[].certificate):
    IF certID != "":
      EMIT Binding(cert_role=server, cert_scope=<source>)
    IF mode == MUTUAL AND certCaID != "":
      EMIT Binding(cert_id=certCaID, cert_role=client_ca, cert_scope=<source>)
```

**唯一键（BindingKey）**:

```
{owner_namespace}/{owner_name}|{cert_id}|{cert_role}|{cert_scope}|{protocol}|{port}|{domain}
```

**状态**: 每轮 Check 全量重建；不持久化。

---

### 2. CertExpiryMap（运行时映射）

| 字段 | 类型 | 说明 |
|------|------|------|
| key | string | CertificateId |
| value | int64 | 过期时间 Unix 时间戳（秒） |

**来源**: `DescribeCertificates` API 响应解析 `CertEndTime`；CA 类型回退 `CAEndTimes` 最早值。

**生命周期**: 每轮 Check 临时构建，不跨轮缓存（保证过期天数实时计算）。

---

### 3. CertificateChecker（巡检组件）

| 字段 | 类型 | 说明 |
|------|------|------|
| cli | client.Client | K8s client（List Ingress） |
| sslClient | SSLClient | 全局凭证 SSL 客户端（非 NS Scope） |
| namespacedSSL | *NamespacedSSL | NS Scope 凭证路由（可选） |
| opts | *option.ControllerOption | 云厂商、NS Scope 配置 |
| lastBindingSet | map[string]struct{} | 上轮 BindingKey 集合（指标清理） |

**执行流程**:

```
Run()
  ├─ List Ingress → 失败则 ERROR 返回
  ├─ expandBindings() → []CertificateBinding
  ├─ groupByCredential() → map[client][]certID  (NS Scope)
  │    OR collectUniqueCertIDs()               (全局)
  ├─ DescribeCertificates() → CertExpiryMap
  ├─ updateMetrics(bindings, expiryMap)
  │    ├─ Set days_until_expiry + query_success=1 (成功)
  │    ├─ Delete days_until_expiry + Set query_success=0 (失败)
  │    └─ Set bindings_total = len(bindings)
  └─ cleanupStaleMetrics(currentBindingSet)
       └─ DeleteLabelValues for keys in lastBindingSet - currentBindingSet
```

---

### 4. SSLClient（云适配实体）

**定义位置**: `internal/cloud/tencentcloud/sslclient.go`

| 方法 | 说明 |
|------|------|
| `NewSSLClient()` | 环境变量全局凭证 |
| `NewSSLClientWithSecretIDKey(id, key)` | 指定凭证构造 |
| `DescribeCertificates(certIDs []string) (map[string]int64, error)` | 批量查询，分页，重试 3 次 |

**接口抽象**（测试用）:

```go
type SSLClient interface {
    DescribeCertificates(certIDs []string) (map[string]int64, error)
}
```

---

### 5. NamespacedSSL（NS Scope 凭证路由）

**定义位置**: `internal/cloud/namespacedssl/namespacedclient.go`

| 字段 | 类型 | 说明 |
|------|------|------|
| k8sClient | client.Client | 读取 per-NS Secret/Config |
| nsClientSet | map[string]SSLClient | NS → SSL 客户端缓存 |
| defaultClient | SSLClient | 全局凭证客户端 |
| exemptNamespaces | map[string]struct{} | 豁免 NS 白名单 |

**getNsClient(ns) 决策**:

```
IF isExempt(ns):
  RETURN defaultClient
ELSE:
  LOOKUP/CREATE per-NS SSLClient from Secret/ControllerConfig
```

---

## Prometheus 指标模型

### GaugeVec: certificate_days_until_expiry

| 属性 | 值 |
|------|-----|
| 全名 | `bkbcs_ingressctrl_certificate_days_until_expiry` |
| 类型 | Gauge |
| Labels | owner_namespace, owner_name, cert_id, cert_role, cert_scope, protocol, port, domain |
| 值域 | 浮点天数；已过期为负数 |

### GaugeVec: certificate_query_success

| 属性 | 值 |
|------|-----|
| 全名 | `bkbcs_ingressctrl_certificate_query_success` |
| 类型 | Gauge |
| Labels | 同上 8 个 |
| 值域 | 0（失败）/ 1（成功） |

### Gauge: certificate_bindings_total

| 属性 | 值 |
|------|-----|
| 全名 | `bkbcs_ingressctrl_certificate_bindings_total` |
| 类型 | Gauge |
| Labels | 无 |
| 值域 | 当前轮 Binding 总数 |

### 联动规则

| 事件 | days_until_expiry | query_success |
|------|-------------------|---------------|
| 查询成功 | Set(days) | Set(1) |
| 查询失败/无效 | DeleteLabelValues | Set(0) |
| Binding 移除 | DeleteLabelValues | DeleteLabelValues |

---

## 实体关系图

```text
Ingress CR
    │ List (全集群)
    ▼
CertificateBinding[] ──cert_id──► CertExpiryMap
    │                                ▲
    │                                │ DescribeCertificates
    │                                │
    ├─ (全局模式) ──► SSLClient (default)
    │
    └─ (NS Scope) ──► NamespacedSSL.getNsClient(owner_namespace)
                          ├─ exempt NS → defaultClient
                          └─ other NS → per-NS SSLClient

CertificateBinding[] ──► Prometheus GaugeVec (days_until_expiry, query_success)
CertificateBinding[] ──► Prometheus Gauge (bindings_total)

CertificateChecker.lastBindingSet ──diff──► DeleteLabelValues (cleanup)
```

---

## Ingress CR 输入字段映射

| Ingress 字段路径 | Binding 属性 |
|----------------|-------------|
| `metadata.namespace` | owner_namespace |
| `metadata.name` | owner_name |
| `spec.portMappings[].protocol` | protocol |
| `spec.portMappings[].port` | port |
| `*.certificate.certID` | cert_id (cert_role=server) |
| `*.certificate.certCaID` | cert_id (cert_role=client_ca) |
| `*.certificate.mode` | 决定是否产生 client_ca |
| `rules[].layer7Routes[].domain` | domain (cert_scope=route) |

---

## 验证不变量

1. 每个 Binding 最多对应 1 条 `days_until_expiry` + 1 条 `query_success` series（相同 8 label）
2. `query_success=0` 时 `days_until_expiry` 对应 label 组合不存在
3. `bindings_total` 等于当前轮 `len(bindings)`，与成功/失败无关
4. 非 SSL 协议 Ingress 不产生任何 Binding 或指标 series
5. 指标 series 不包含 Secret/AccessKey 等敏感信息
