# Clarification Questions — Story 1070046748135050873

## Q1 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：CertificateChecker 如何复用现有 Checker 注册与周期执行模式？
**影响**：决定 Checker 骨架与 main.go 注册方式；阻塞。
**建议候选**：
- A. 实现 `check.Checker` 接口，通过 `CheckRunner.Register(..., check.CheckPer10Min)` 注册（推荐：与 ListenerChecker 一致，且 `CheckPer10Min` 已定义但尚未被使用）
- B. 独立 goroutine + time.Ticker，绕过 CheckRunner
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A。`CertificateChecker` 实现 `Checker.Run()`，在 `main.go` 中仅当 `opts.Cloud == constant.CloudTencentCloud` 时注册到 `CheckRunner`，周期为 `CheckPer10Min`（10 分钟）。`Run()` 内 List 全集群 Ingress、展开 Binding、查询过期时间、更新指标；参考 `ListenerChecker` 维护 `lastBindingSet` 做 `DeleteLabelValues` 清理。
**第 5 轮修订（F-011）**：周期已由 `CheckPer10Min`（10 分钟）调整为 `CheckPer60Min`（60 分钟）；`main.go` 使用 `CertificateCheckerRegisterInterval` 常量注册。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：internal/check/checkrunner.go、internal/check/interface.go、internal/check/listenerchecker.go、main.go

## Q2 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：腾讯云 DescribeCertificate SDK 应接入哪一层、如何构造客户端？
**影响**：决定云适配层代码位置与凭证注入方式；阻塞。
**建议候选**：
- A. 在 `internal/cloud/tencentcloud/` 新增 SSL 客户端封装，参照 `NewClb()` / `NewClbWithSecretIDKey()` 模式（推荐）
- B. 在 `internal/check/` 直接引用 SSL SDK
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A，符合 ARCH 分层约束（Checker 依赖 cloud 适配层，cloud 子包不反向依赖 check）。新增 `internal/cloud/tencentcloud/sslclient.go`（或等价文件），引入 `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl/v20191205`，endpoint 为 `ssl.tencentcloudapi.com`，API 版本 `2019-12-05`。提供 `NewSSLClient()`（全局凭证，读 `TENCENTCLOUD_ACCESS_KEY_ID` / `TENCENTCLOUD_ACESS_KEY` 环境变量）和 `NewSSLClientWithSecretIDKey(id, key)`（per-NS 凭证）。批量查询封装 `DescribeCertificates(certIDs []string)`，单次最多 1000 ID 并分页；失败重试 3 次；调用结束通过 `metrics.ReportLibRequestMetric(system=tencentcloud, handler=sdk, method=DescribeCertificates, ...)` 上报。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：internal/cloud/tencentcloud/clb.go、internal/cloud/tencentcloud/sdk.go、docs/harness/architectural-constraints.md、docs/reqs/证书过期Checker核心实现.md

## Q3 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：Namespace Scope 模式下 DescribeCertificate 凭证路由如何复用 namespacedlb 模式？
**影响**：决定多租户场景实现路径与子需求 #2 边界；阻塞。
**建议候选**：
- A. 新建 `internal/cloud/namespacedssl/` 镜像 `namespacedlb` 的 `getNsClient` 逻辑（推荐）
- B. 在 `NamespacedLB` 上扩展 SSL 查询方法
- C. CertificateChecker 内直接读 per-NS Secret
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A。`NamespacedLB` 封装的是 `cloud.LoadBalance`（CLB 操作），与 SSL 证书查询职责不同，不宜在其上扩展。新建 `NamespacedSSL`（或 `NamespacedSSLClient`），复用 ADR-0001 数据流：`exemptNamespaces` 中的 NS 使用 `defaultClient`（controller 全局凭证），其它 NS 通过 per-NS Secret（`ingress-secret.networkextension.bkbcs.tencent.com`）或 ControllerConfig（`ingress-config.networkextension.bkbcs.tencent.com`）构造 `SSLClient`。`main.go` 的 `newNamespacedLBWithExempt` 模式可平移：`IsNamespaceScope=false` 时注入全局 `SSLClient`；`true` 时注入 `NamespacedSSL`。CertificateChecker 按 Binding 的 `owner_namespace` 获取客户端，再按凭证实例分组批量查询（同一凭证下的 certID 合并，遵守 1000 上限分页）。单 NS 凭证失败仅影响该 NS 的 Binding。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：docs/adr/0001-namespace-scope-exemption.md、internal/cloud/namespacedlb/namespacedclient.go、main.go、docs/reqs/证书过期NS Scope凭证.md

## Q4 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：certificate 子系统 Prometheus 指标如何注册与清理？
**影响**：决定指标文件结构与 series 生命周期管理；阻塞。
**建议候选**：
- A. 新增 `internal/metrics/certificate.go`，`init()` 中 `metrics.Registry.MustRegister`（推荐，与 check.go 一致）
- B. 在 main.go 手动注册
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A。新增 GaugeVec：`bkbcs_ingressctrl_certificate_days_until_expiry`（8 个 label：owner_namespace、owner_name、cert_id、cert_role、cert_scope、protocol、port、domain）、`bkbcs_ingressctrl_certificate_query_success`（同上 label）、`bkbcs_ingressctrl_certificate_bindings_total`（无 label）。series 清理参照 `ListenerChecker.lastMetricMap`：CertificateChecker 维护上轮 Binding label 集合，每轮全量重建后对消失的 Binding 调用 `DeleteLabelValues`；`query_success=0` 或失效时间无效时删除 `days_until_expiry` 对应 series 但保留/设置 `query_success=0`。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：internal/metrics/check.go、internal/metrics/metric.go、internal/check/listenerchecker.go、docs/reqs/证书过期指标.md

## Q5 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：子需求拆分后的实现顺序与代码归属？
**影响**：决定迭代交付节奏与文件边界；非阻塞。
**建议候选**：
- A. 先交付核心子需求（Binding 展开 + 全局凭证 SSL 查询 + 指标 + Checker 注册），再交付 NS Scope 凭证扩展（推荐）
- B. 一次性实现全部逻辑
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A，与已拆分文档一致。核心子需求（#1070046748135054749）交付：`internal/check/certificatechecker.go`（Binding 展开 + Checker 主流程）、`internal/cloud/tencentcloud/sslclient.go`、`internal/metrics/certificate.go`、`main.go` 注册。NS Scope 子需求（#1070046748135054806）在此基础上新增 `internal/cloud/namespacedssl/` 并扩展 CertificateChecker 的凭证分组逻辑，不改动指标定义与 Checker 注册。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：docs/reqs/证书过期Checker核心实现.md、docs/reqs/证书过期NS Scope凭证.md

## Q6 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：Binding 展开应复用哪些现有类型与协议判断逻辑？
**影响**：决定 Ingress 遍历实现细节；非阻塞。
**建议候选**：
- A. 直接遍历 `networkextensionv1.Ingress` Spec 的 rules / layer7Routes / portMappings，协议判断复用 `constant.Layer4Protocol` 与 `validate.go` 中 `isSSLProtocol` 等价逻辑（推荐）
- B. 通过 generator 层间接展开
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-06-09T12:00:00+08:00
**答复**：采用方案 A。Checker 层直接 List Ingress CR，对 HTTPS/TCP_SSL/QUIC 协议展开 `IngressListenerCertificate`（certID → cert_role=server；MUTUAL + certCaID → cert_role=client_ca）。不经过 generator 层，避免引入额外依赖。`CertEndTime` 解析使用 `time.ParseInLocation("2006-01-02 15:04:05", ..., Asia/Shanghai)`；CA 类型证书 `CertEndTime` 为空时取 `CAEndTimes` 最早值。
**答复方**：subagent(自答) / ts=2026-06-09T12:00:00+08:00
**文档来源**：internal/constant/constant.go、internal/cloud/tencentcloud/validate.go、docs/reqs/证书过期指标.md

## Q7 [resolved_by_user] — 来源：tapd-story-prepare（第 4 轮澄清）
**问题**：腾讯云 SSL API 请求域名是否可配置？如何对齐 CLB 内网域名配置模式？
**影响**：决定内网/私有化部署场景下证书过期检测是否可用；阻塞 F-010 实现。
**建议候选**：
- A. 环境变量 `TENCENTCLOUD_SSL_DOMAIN` + Helm `tencentcloudSslDomain`，对齐 `TENCENTCLOUD_CLB_DOMAIN` / `tencentcloudClbDomain`（推荐）
- B. 与 CLB 共用同一域名环境变量
- C. per-NS Secret 支持独立 SSL 域名
**提出方**：tapd-story-prepare / attempt=3 / round=3 / ts=2026-06-10T19:52:00+08:00
**答复**：采用方案 A。默认 `ssl.tencentcloudapi.com`（向后兼容）；内网示例 `ssl.internal.tencentcloudapi.com`；Controller 全局 env 注入；未来所有 SSL API 调用复用此配置；per-NS Secret 不含 SSL 域名（与 CLB 域名粒度一致）。作为父需求增量修订 F-010，Middle 优先级。
**答复方**：user / ts=2026-06-10T19:52:00+08:00
**文档来源**：docs/reqs/证书过期指标.md §F-010、docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/values.yaml（tencentcloudClbDomain 参考）、internal/cloud/tencentcloud/sdk.go（EnvNameTencentCloudClbDomain）
