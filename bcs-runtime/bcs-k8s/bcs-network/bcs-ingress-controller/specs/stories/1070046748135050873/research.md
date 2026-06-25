# Research: 证书过期时间 Prometheus 指标

**Feature**: `1070046748135050873`
**Date**: 2026-06-09

## 研究问题与结论

### R-001: CertificateChecker 应放在哪个包？

**结论**: `internal/check/certificatechecker.go`，实现 `check.Checker` 接口。

**理由**:
- 现有 `ListenerChecker`、`IngressChecker`、`PortBindChecker` 均在 `internal/check/`，周期由 `CheckRunner` 统一调度
- 证书过期检查为只读巡检，不修改 CR，不符合 Reconcile 模式，无需新建 `{name}controller/`
- `CheckRunner.Register(..., check.CheckPer10Min)` 已有 10 分钟 ticker，直接复用

**替代方案评估**:
- 方案 A：在 `ingresscontroller` Reconcile 中查询 → 拒绝：与同步逻辑耦合，增加 Reconcile 延迟，违反 FR-007
- 方案 B：独立 goroutine 自管 ticker → 拒绝：绕过 CheckRunner 统一管理，不符合现有惯例

### R-002: Binding 展开是否复用 generator 层？

**结论**: 不复用，Checker 层直接 List Ingress 并展开。

**理由**:
- spec 澄清明确：不经过 generator 层；Checker 只需证书挂载关系，不需 Listener 生成模型
- generator 输入为单个 Ingress Reconcile 上下文，Checker 需全集群批量 List，调用模式不同
- 展开逻辑为纯函数（IngressList → []CertificateBinding），可独立测试

**展开规则摘要**:
- 协议过滤：HTTPS、TCP_SSL、QUIC
- 三处来源：`rules[].certificate`、`rules[].layer7Routes[].certificate`、`portMappings[].certificate`
- cert_role：certID → server；MUTUAL + certCaID → client_ca
- domain：route scope 取域名，其它为 `""`

### R-003: 腾讯云 DescribeCertificate SDK 接入层

**结论**: `internal/cloud/tencentcloud/sslclient.go`，独立 `SSLClient` 封装。

**理由**:
- 遵循架构约束：云 SDK 调用封装在 `internal/cloud/tencentcloud/`
- 复用 `SdkWrapper` 的重试模式（参考 `DescribeLoadBalancers` 3 次重试 + `ReportLibRequestMetric`）
- endpoint：`ssl.tencentcloudapi.com`，API 版本 `2019-12-05`，方法 `DescribeCertificates`
- 单次 `Limit` 最大 1000，超过时分页

**API 调用指标**:
```go
metrics.ReportLibRequestMetric("tencentcloud", handler, "DescribeCertificates", status, startTime)
```

### R-004: CertEndTime 解析策略

**结论**: GMT+8（`Asia/Shanghai`）解析格式 `2006-01-02 15:04:05`；CA 类型回退 `CAEndTimes` 取最早值。

**理由**:
- 腾讯云 SSL API 文档约定 CertEndTime 为北京时间字符串
- CA 证书 CertEndTime 可能为空，CAEndTimes 为多域名到期时间列表，取 min 符合"最早过期"监控语义
- 解析失败或空值 → `query_success=0`，不设置 `days_until_expiry`

### R-005: Prometheus 指标注册模式

**结论**: 新增 `internal/metrics/certificate.go`，`init()` 中 `metrics.Registry.MustRegister`。

**理由**:
- 与 `check.go`（ListenerTotal）、`hostnetportpool.go` 模式一致
- 3 个 GaugeVec：
  - `bkbcs_ingressctrl_certificate_days_until_expiry`（8 label）
  - `bkbcs_ingressctrl_certificate_query_success`（8 label，值 0/1）
  - `bkbcs_ingressctrl_certificate_bindings_total`（无 label）
- 不携带 `bcs_cluster_id`（蓝鲸监控平台自动注入）

### R-006: 指标清理策略

**结论**: `CertificateChecker` 维护 `lastBindingSet map[string]struct{}`，每轮重建后 diff 删除。

**理由**:
- 直接参照 `ListenerChecker.lastMetricMap` 模式（setMetric 后遍历 last 删除不存在 key）
- `query_success=0` 时仅删除 `days_until_expiry` 对应 label 组合，保留 `query_success=0` series
- Binding 完全消失时两者均 `DeleteLabelValues`

### R-007: Namespace Scope 凭证路由

**结论**: 新建 `internal/cloud/namespacedssl/`，镜像 `namespacedlb` 的 `getNsClient` 逻辑。

**理由**:
- ADR-0001 已确立豁免 NS 使用 `defaultClient`、其它 NS 走 per-NS Secret/ControllerConfig 模式
- `namespacedlb` 与 SSL 客户端类型不同（`cloud.LoadBalance` vs `SSLClient`），不宜在同一 struct 中混合
- 独立包避免 tencentcloud 子包与 namespacedlb 交叉依赖，符合 ARCH 分层规则
- Secret key 沿用历史拼写 `TENCENTCLOUD_ACESS_KEY`

**凭证分组查询**:
1. 按 Binding 的 `owner_namespace` 获取 SSLClient
2. 将 certID 按 client 实例分组
3. 每组独立调用 `DescribeCertificates`，单 NS 失败不影响其它组

### R-008: main.go 注册条件

**结论**: 仅 `opts.Cloud == tencentcloud`（或 `constant.CloudProviderTencent`）时注册，周期 `CheckPer10Min`。

**理由**:
- FR-008 明确非腾讯云不注册、不产出指标
- 无 CLI 开关，腾讯云环境默认开启
- 注册位置：现有 `checkRunner.Register(...)` 链末尾，与 ListenerChecker 等并列

### R-009: 错误隔离与主流程影响

**结论**: Checker 内所有错误显式处理；指标写入失败记录 ERROR 日志；不 panic、不阻塞 Reconcile。

**理由**:
- FR-009 / SC-008 要求 SSL API 故障不影响 Controller 其它功能
- `Run()` 在独立 goroutine 中执行（CheckRunner 已 `go item.Run()`），天然隔离
- List Ingress 失败 → 本轮终止，不部分更新指标（避免不一致状态）

### R-010: 测试替身策略

**结论**: SSLClient 定义 interface；单元测试用 mock/fake 实现；不依赖真实腾讯云 API。

**理由**:
- `DescribeCertificates` 返回可控的 CertEndTime 映射
- fake client（controller-runtime fake）注入 Ingress fixture
- 表驱动覆盖 spec 中全部 Acceptance Scenarios

## 技术风险与缓解

| 风险 | 影响 | 缓解 |
|------|------|------|
| SSL API 限流 | Check 超时 | 按 1000 分页 + 凭证分组减少调用次数；重试 3 次 |
| 大量 Ingress 内存压力 | lastBindingSet 膨胀 | Binding key 为字符串，500 Ingress 量级可接受 |
| namespacedssl 与 namespacedlb 代码重复 | 维护成本 | 仅镜像 getNsSecret/getNsClient 核心逻辑，不抽取公共包（YAGNI） |
| CA 证书 CAEndTimes 格式多样 | 解析失败 | 取最早有效时间；失败置 query_success=0 + INFO 日志 |

## 依赖确认

| 依赖项 | 状态 | 说明 |
|--------|------|------|
| 腾讯云 SSL SDK | 需引入 | `tencentcloud-sdk-go/tencentcloud/ssl` |
| Ingress CRD 类型 | 已有 | `networkextensionv1.IngressListenerCertificate` |
| CheckRunner | 已有 | `internal/check/checkrunner.go` |
| ReportLibRequestMetric | 已有 | `internal/metrics/metric.go` |
| namespacedlb 模式 | 已有 | ADR-0001 + `namespacedclient.go` 参考 |
