# Feature Specification: 证书过期时间 Prometheus 指标

**Feature Directory**: `specs/stories/1070046748135050873`  
**Created**: 2026-06-09  
**Updated**: 2026-06-11（第 5 轮澄清同步）  
**Status**: Implemented（squash commit 62759c390）  
**Input**: 为 BCS IngressController 新增证书过期时间 Prometheus 指标。实现范围 F-001~F-011（含 NS Scope 凭证路由、CLI/Helm 开关、DescribeCertificate 限流、SSL API 域名可配置、检查周期 60 分钟）；云厂商仅腾讯云；Checker 周期 **60 分钟**（`CheckPer60Min`）；默认关闭，须显式开启。

## Clarifications

### Session 2026-06-09

- Q: CertificateChecker 如何复用现有 Checker 注册与周期执行模式？ → A: 实现 `check.Checker` 接口，通过 `CheckRunner.Register(..., check.CheckPer60Min)` 注册（*第 5 轮由 CheckPer10Min 修订为 CheckPer60Min*）；仅 `opts.Cloud == tencentcloud` 时注册；`Run()` 内 List Ingress、展开 Binding、查询过期时间、更新指标；参照 `ListenerChecker` 维护 `lastBindingSet` 做 `DeleteLabelValues` 清理。
- Q: 腾讯云 DescribeCertificate SDK 应接入哪一层？ → A: 在 `internal/cloud/tencentcloud/` 新增 SSL 客户端封装（`sslclient.go`），提供 `NewSSLClient()` 与 `NewSSLClientWithSecretIDKey()`；上层方法名 `DescribeCertificates(certIDs []string)` 为批量语义封装，底层逐 ID 调用云 API `DescribeCertificate`（单数），单次最多 1000 ID 分页，失败重试 3 次；通过 `metrics.ReportLibRequestMetric(method=DescribeCertificate)` 上报（*第 3 轮修订 method 名*）。
- Q: Namespace Scope 模式下凭证路由如何实现？ → A: 新建 `internal/cloud/namespacedssl/` 镜像 `namespacedlb` 的 `getNsClient` 逻辑；豁免 NS 使用 `defaultClient`（全局凭证），其它 NS 通过 per-NS Secret 或 ControllerConfig 构造 SSLClient；按 Binding 的 `owner_namespace` 获取客户端并按凭证分组批量查询。
- Q: certificate 子系统 Prometheus 指标如何注册与清理？ → A: 新增 `internal/metrics/certificate.go`，`init()` 中 `metrics.Registry.MustRegister`；GaugeVec：`bkbcs_ingressctrl_certificate_days_until_expiry`、`bkbcs_ingressctrl_certificate_query_success`（8 个 label）、`bkbcs_ingressctrl_certificate_bindings_total`（无 label）；`query_success=0` 时删除 `days_until_expiry` series 但保留/设置 `query_success=0`。
- Q: 子需求交付顺序？ → A: 先核心（全局凭证 Binding 展开 + SSL 查询 + 指标 + Checker 注册），再 NS Scope 凭证扩展；核心子需求 #1070046748135054749，NS Scope 子需求 #1070046748135054806。
- Q: Binding 展开应复用哪些现有逻辑？ → A: Checker 层直接 List Ingress CR，对 HTTPS/TCP_SSL/QUIC 协议展开 `IngressListenerCertificate`；不经过 generator 层；`CertEndTime` 解析格式 `2006-01-02 15:04:05`（GMT+8/Asia/Shanghai）；CA 类型 `CertEndTime` 为空时取 `CAEndTimes` 最早值。
- Q: Namespace Scope 模式下 DescribeCertificate 用哪套云凭证？ → A: 按 Ingress 所在 NS 使用对应云凭证；豁免 NS 使用全局凭证（与 Listener 处理模式一致）。
- Q: 指标是否携带 bcs_cluster_id label？ → A: 不携带，由蓝鲸监控平台自动添加。
- Q: 是否支持关闭证书过期检查？ → A: 腾讯云环境下默认开启，无 CLI 开关（*第 2 轮已修订为默认关闭 + CLI/Helm 开关*）。
- Q: 非腾讯云集群部署时的行为？ → A: 仅 `cloud=tencentcloud` 时注册 CertificateChecker。

### Session 2026-06-10（第 2 轮：证书过期检测开关）

- Q: 开关未配置时的默认行为？ → A: **默认关闭**（`certificate_check_enabled=false`）。许多现存腾讯云账号无 `ssl:DescribeCertificates` 权限，默认关闭避免升级后无效 API 调用或启动失败。
- Q: 关闭后的行为？ → A: 不初始化 SSL Client、不注册 CertificateChecker；不主动清理已有 Prometheus series（滚动重启 Pod 后旧 series 随 Pod 销毁自动消失）。
- Q: CLI 参数命名？ → A: CLI `--certificate_check_enabled`（默认 `false`）；Go `CertificateCheckEnabled`；Helm `certificateCheckEnabled`。
- Q: Helm Chart 更新范围？ → A: `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/` 的 `values.yaml` 与 `templates/deployment.yaml`；`true` 时追加 `--certificate_check_enabled=true`，`false`（默认）时不追加该参数。

### Session 2026-06-10（第 3 轮：API 实现与限流）

- Q: 使用 DescribeCertificate 还是 DescribeCertificates 批量？ → A: **继续** `DescribeCertificate`（单数）逐 ID 轮询；`SSLClient.DescribeCertificates` 仅为上层封装方法名；**不**改为 `DescribeCertificates`（复数）批量查询。
- Q: 为何不升级 SDK？ → A: 保持 `tencentcloud-sdk-go v1.0.132` 单体包；子模块 ≥ v1.0.1090 才支持 `CertIds` 批量；升级须移除单体包并联动 CLB/VPC/CVM，存在 omitnil 序列化行为变化等稳定性风险。
- Q: DescribeCertificate 限流如何实现？ → A: **方案 A**——抽取进程级共享 `throttle.RateLimiter`（`sync.Once` 初始化，读取 `TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`），`SdkWrapper` 与 `sslClientImpl` **共用同一令牌桶实例**；每次 `DescribeCertificate` 前 `Accept()`，与 CLB 路径共享 QPS 配额。
- Q: 指标 method 名称？ → A: `DescribeCertificate`（与实际云 API 一致）。

### Session 2026-06-10（第 4 轮：SSL API 域名可配置）

- Q: 需求在 TAPD 中如何组织？ → A: 作为父需求 #1070046748135050873 的增量修订，直接追加 F-010 到原需求文档。
- Q: 未配置 SSL 域名时的默认行为？ → A: 默认 `ssl.tencentcloudapi.com`（与现网硬编码行为一致）。
- Q: 内网 SSL API 域名？ → A: `ssl.internal.tencentcloudapi.com`（对齐 CLB 内网域名命名规则）。
- Q: 配置项命名？ → A: 环境变量 `TENCENTCLOUD_SSL_DOMAIN`；Helm `tencentcloudSslDomain`（对齐 `TENCENTCLOUD_CLB_DOMAIN` / `tencentcloudClbDomain`）。
- Q: SSL 域名配置作用范围？ → A: Controller 全局 env；未来所有 SSL API 调用均复用此配置，禁止硬编码域名。
- Q: NS Scope 下 per-NS Secret 是否支持独立 SSL 域名？ → A: 不需要；SSL 域名为 Controller 全局配置（与 CLB 域名粒度一致）。
- Q: 需求优先级？ → A: Middle。

### Session 2026-06-11（第 5 轮：检查周期调整）

- Q: 调整范围？ → A: **仅** CertificateChecker；PortBindChecker / ListenerChecker 等 1 分钟周期 Checker 不受影响。
- Q: 新周期？ → A: 固定 **60 分钟（1 小时）**；不新增 CLI / Helm 可配置项。
- Q: 单轮 Check 逻辑是否变化？ → A: **不变**；仅调度频率由 10 分钟降为 60 分钟。
- Q: 需求在 TAPD 中如何组织？ → A: 作为父需求 #1070046748135050873 增量修订，追加 F-011。

## User Scenarios & Testing *(mandatory)*

### User Story 1 - 查看 Ingress 证书剩余过期天数 (Priority: P1)

BCS 集群运维人员在 Prometheus / 蓝鲸监控中查看各 Ingress 关联 SSL 证书的剩余过期天数，以便在证书过期前收到告警并定位到具体 Ingress，避免 HTTPS/TCP_SSL/QUIC 业务中断。

**Why this priority**: 这是本功能的核心价值——按 Ingress 维度暴露证书过期天数，弥补 Controller 同步 CLB 时不读取、不回写过期时间的缺口。

**Independent Test**: 在腾讯云集群中设置 `--certificate_check_enabled=true`，创建带有效 certID 的 HTTPS Ingress，等待 CertificateChecker 执行一轮 Check，验证 `bkbcs_ingressctrl_certificate_days_until_expiry` 出现对应 series 且 `query_success=1`，天数与云端 CertEndTime 一致（误差 < 1 天）。

**Acceptance Scenarios**:

1. **Given** `--certificate_check_enabled=true` 且集群中存在 HTTPS Ingress 且 certID 有效、SSL API 权限正常，**When** CertificateChecker 执行一轮 Check，**Then** `bkbcs_ingressctrl_certificate_days_until_expiry` 出现对应 series 且 `query_success=1`，天数与 CertEndTime 一致（误差 < 1 天）。
2. **Given** Ingress 为 MUTUAL 模式且 certID、certCaID 均配置，**When** Check 执行，**Then** 分别产生 `cert_role=server` 与 `cert_role=client_ca` 两条 series。
3. **Given** Ingress 从三处（rule / route / port_mapping）配置证书，**When** Check 执行，**Then** `cert_scope` label 分别正确为 rule / route / port_mapping。
4. **Given** 证书已过期，**When** Check 执行，**Then** `days_until_expiry` 为负数（如过期 3 天 → `-3`）。
5. **Given** 纯 TCP/UDP 四层规则（非 SSL 协议），**When** Check 执行，**Then** 不产生证书 Binding 与指标 series。

---

### User Story 2 - 区分查询成功与查询失败 (Priority: P1)

平台 SRE 需要区分「证书即将过期」与「过期时间查询失败」两种状态，以便告警表达式同时要求 `query_success == 1`，避免 SSL API 故障时产生误报，同时不遗漏真实过期风险。

**Why this priority**: 与核心监控能力同等重要——没有 query_success 区分，API 权限或调用故障会导致误报或漏报。

**Independent Test**: 模拟 SSL API 权限未开通或连续 3 次调用失败，验证受影响 Binding 的 `query_success=0` 且无 `days_until_expiry` series，Controller 其它功能正常。

**Acceptance Scenarios**:

1. **Given** SSL API 权限未开通或连续 3 次调用失败，**When** Check 执行，**Then** 受影响 Binding `query_success=0` 且无 `days_until_expiry` series，Controller 其它功能正常。
2. **Given** 证书 ID 在 API 响应中缺失或失效时间无效，**When** Check 执行，**Then** 对应 Binding `query_success=0`，删除 `days_until_expiry` series，记录 INFO/ERROR 日志。
3. **Given** 运维配置蓝鲸监控告警，**When** 编写 PromQL，**Then** 可基于 `query_success == 1` 过滤，仅对成功查询的 series 评估过期阈值。

---

### User Story 3 - 指标随 Ingress 生命周期自动清理 (Priority: P1)

当 Ingress 被删除或证书配置变更导致 Binding 消失时，上一轮存在的 Prometheus series 应被清理，保证指标与当前集群状态一致。

**Why this priority**: 残留 series 会导致已删除 Ingress 仍触发过期告警，误导运维。

**Independent Test**: 创建 Ingress 并等待指标出现，删除 Ingress 后等待下一轮 Check，验证该 Ingress 相关 series 被 DeleteLabelValues 清理。

**Acceptance Scenarios**:

1. **Given** Ingress 被删除，**When** 下一轮 Check 执行，**Then** 该 Ingress 相关 `days_until_expiry` 与 `query_success` series 均被 DeleteLabelValues 清理。
2. **Given** 上轮存在某 Binding、本轮 Ingress 证书配置已移除，**When** Check 完成，**Then** 该 Binding 全部 label 组合的 series 均被清理。
3. **Given** 每轮 Check 执行，**When** 重建 Binding 集合，**Then** `bkbcs_ingressctrl_certificate_bindings_total` 更新为当前 Binding 总数。

---

### User Story 4 - Namespace Scope 多租户凭证查询 (Priority: P1)

多租户 Namespace Scope 集群管理员希望各 Namespace 使用各自云凭证查询证书过期时间（豁免 NS 使用全局凭证），使监控结果与 Listener 同步行为一致。

**Why this priority**: 联邦/多租户场景下若仍用全局凭证，无法查询各 NS 账号下的证书，监控失效。

**Independent Test**: 启用 `--is_namespace_scope=true`，NS-A 使用独立 Secret 凭证且 Ingress 引用 NS-A 账号下证书，验证使用 NS-A 凭证查询且 `query_success=1`；豁免 NS 验证使用全局凭证。

**Acceptance Scenarios**:

1. **Given** `--is_namespace_scope=true` 且 NS-A 使用独立 Secret 凭证，**When** NS-A 的 Ingress 引用 NS-A 账号下证书，**Then** 使用 NS-A 凭证查询且 `query_success=1`，`days_until_expiry` 与 CertEndTime 一致。
2. **Given** NS 在 `--namespace_scope_exempt_namespaces` 白名单，**When** 该 NS Ingress 引用证书，**Then** 使用 Controller 全局凭证查询且 `query_success=1`。
3. **Given** NS-B 凭证缺失或 SSL API 权限不足，**When** Check 执行，**Then** 仅 NS-B 下 Binding `query_success=0`，NS-A 及其它 NS 指标正常。
4. **Given** `--is_namespace_scope=false`，**When** Check 执行，**Then** 所有 DescribeCertificate 调用使用全局凭证，不读取 per-NS Secret。

---

### User Story 5 - CertificateChecker 集成与 CLI/Helm 开关 (Priority: P1)

平台 SRE 希望通过 CLI 参数或 Helm 配置显式开启证书过期检测：在未开通 SSL API 权限的集群中保持默认关闭，在权限就绪的集群中按需启用；非腾讯云环境不注册 Checker。

**Why this priority**: 默认关闭是生产安全前提——避免无 SSL 权限账号升级后触发无效 API 调用或启动失败（当前无条件注册时 SSL Client 初始化失败会导致 Controller 退出）。

**Independent Test**: 腾讯云环境 `--certificate_check_enabled=true` 时验证 CertificateChecker 注册到 `checkPer60Min`；未指定开关或 `false` 时验证不初始化 SSL Client、不注册 Checker；Helm `certificateCheckEnabled: true` 时 Deployment args 包含对应参数。

**Acceptance Scenarios**:

1. **Given** 腾讯云环境且未指定 `--certificate_check_enabled`，**When** Controller 启动，**Then** 不初始化 SSL Client、不注册 CertificateChecker（默认关闭）。
2. **Given** 腾讯云环境且 `--certificate_check_enabled=true`，**When** Controller 启动，**Then** 初始化 SSL Client、注册 CertificateChecker 且每 **60 分钟**执行。
3. **Given** 腾讯云环境且 `--certificate_check_enabled=false`，**When** Controller 启动，**Then** 不初始化 SSL Client、不注册 CertificateChecker。
4. **Given** Controller 部署云厂商非腾讯云，**When** 启动，**Then** 不注册 CertificateChecker（与开关无关）。
5. **Given** Helm `certificateCheckEnabled: true`，**When** 部署 Controller，**Then** Deployment args 包含 `--certificate_check_enabled=true` 且 CertificateChecker 注册。
6. **Given** Helm `certificateCheckEnabled: false`（默认），**When** 部署 Controller，**Then** Deployment args 不包含 `--certificate_check_enabled` 且 CertificateChecker 不注册。
7. **Given** 曾开启检测后关闭开关，**When** 滚动重启 Pod，**Then** 不主动清理已有 series；旧 series 随旧 Pod 销毁自动消失。

---

### User Story 6 - DescribeCertificate 限流与 SDK 策略 (Priority: P1)

平台 SRE 需要 DescribeCertificate 请求接入 Controller 已有的腾讯云 API 令牌桶限流，避免 Checker 每 60 分钟全量轮询时瞬时打满 SSL API QPS；同时保持现有 SDK 版本不升级，在技术文档中明确轮询代替批量的原因。

**Why this priority**: 大规模 certID 集群在无限流时可能触发云侧 `RequestLimitExceeded`；SDK 升级存在 CLB/ENI 行为回归风险。

**Independent Test**: 设置 `TENCENTCLOUD_RATELIMIT_QPS=10` 且 20 个唯一 certID，验证每轮 Check 中每次 `DescribeCertificate` 前均经过共享令牌桶；`bkbcs_ingressctrl_lib_request_total{method="DescribeCertificate"}` 按实际 API 次数递增。

**Acceptance Scenarios**:

1. **Given** `--certificate_check_enabled=true` 且 200 个唯一 certID，**When** 一轮 Check 执行，**Then** 每次 `DescribeCertificate` 调用前均经过令牌桶限流，且 `bkbcs_ingressctrl_lib_request_total{method="DescribeCertificate"}` 计数为 200（±重试次数）。
2. **Given** `TENCENTCLOUD_RATELIMIT_QPS=10` 且 20 个唯一 certID，**When** 一轮 Check 执行，**Then** DescribeCertificate 调用间隔受共享令牌桶约束（可通过 mock 验证 `Accept()` 调用次数 ≥ 20）。
3. **Given** SSL 客户端实现，**When** 审查底层云 API 调用，**Then** 使用 `DescribeCertificate`（单数）逐 ID 轮询，不调用 `DescribeCertificates`（复数）批量接口；SDK 保持 `v1.0.132` 不升级。
4. **Given** 500 个 Ingress、200 个唯一 certID 的集群，**When** Check 执行，**Then** 单次 Check 在 **60 分钟**周期内完成（含限流等待；200 次 @ 默认 50 QPS 理论下限约 4s，须预留凭证分组与重试余量）。

---

### User Story 7 - SSL API 请求域名可配置 (Priority: P1)

BCS 集群运维人员希望像配置 CLB API 域名一样，通过环境变量或 Helm values 指定腾讯云 SSL 证书 API 的请求域名，以便在内网/私有化部署场景下通过 `ssl.internal.tencentcloudapi.com` 等内网 endpoint 正常查询证书过期时间。

**Why this priority**: 内网集群若仍请求公网 `ssl.tencentcloudapi.com`，证书过期检测将全量 `query_success=0`；与 CLB 内网域名配置模式对齐可降低部署成本。

**Independent Test**: 设置 `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com` 或 Helm `tencentcloudSslDomain`，验证 `NewSSLClientWithSecretIDKey` 创建的 SDK `HttpProfile.Endpoint` 为配置值；未设置时仍为 `ssl.tencentcloudapi.com`。

**Acceptance Scenarios**:

1. **Given** 未设置 `TENCENTCLOUD_SSL_DOMAIN` 且 `--certificate_check_enabled=true`，**When** SSLClient 发起 `DescribeCertificate`，**Then** 请求 endpoint 为 `ssl.tencentcloudapi.com`（与升级前行为一致）（AC-015）。
2. **Given** `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com` 且 `--certificate_check_enabled=true`，**When** SSLClient 发起 `DescribeCertificate`，**Then** 请求 endpoint 为 `ssl.internal.tencentcloudapi.com`（AC-016）。
3. **Given** Helm `tencentcloudSslDomain: ssl.internal.tencentcloudapi.com`，**When** 部署 Controller，**Then** Deployment env 包含 `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com`（AC-017）。
4. **Given** `--is_namespace_scope=true` 且 per-NS Secret 仅含凭证，**When** Check 执行，**Then** 所有 NS 共用 Controller 全局 `TENCENTCLOUD_SSL_DOMAIN` 配置，Secret 不含 SSL 域名。
5. **Given** SSL 域名配置错误导致 API 不可达，**When** Check 执行，**Then** 受影响 Binding `query_success=0`、记录 ERROR 日志，Controller 其它功能正常。

---

### User Story 8 - 证书过期检查周期调整为 1 小时 (Priority: P1)

平台 SRE 希望将 CertificateChecker 执行周期从 10 分钟调整为 60 分钟，以降低 SSL API 调用频率与 Controller 巡检负载，同时证书过期监控仍能在过期前通过 Prometheus 指标触发告警。

**Why this priority**: 证书有效期以月/年计，10 分钟粒度 API 收益有限；60 分钟可将 SSL API 调用频率降低约 6 倍。

**Independent Test**: 腾讯云环境 `--certificate_check_enabled=true` 启动后，验证 CertificateChecker 注册到 `checkPer60Min`；2 小时内 `Run()` 执行 2 次（±启动偏差）；PortBindChecker / ListenerChecker 仍为 1 分钟周期。

**Acceptance Scenarios**:

1. **Given** 腾讯云环境且 `--certificate_check_enabled=true`，**When** Controller 启动，**Then** CertificateChecker 注册成功且每 **60 分钟**触发一次 `Run()`（非 10 分钟）（AC-018）。
2. **Given** CertificateChecker 已注册，**When** 观察 2 小时内 Check 执行次数，**Then** 执行次数为 2 次（±启动时刻偏差）（AC-019）。
3. **Given** `--certificate_check_enabled=true` 且 HTTPS Ingress certID 有效、SSL API 权限正常，**When** CertificateChecker 执行一轮 Check，**Then** 指标行为与 AC-001 一致，仅刷新频率变为 60 分钟（AC-020）。
4. **Given** Controller 运行中，**When** 观察 PortBindChecker / ListenerChecker 等其它 Checker，**Then** 仍保持 **1 分钟**周期，不受本修订影响（AC-021）。

---

### Edge Cases

- List Ingress 失败时，本轮 Check 终止并记录 ERROR 日志，不更新指标。
- UNIDIRECTIONAL 模式不产生 `cert_role=client_ca` Binding。
- SNI 场景 `domain` 取 route 域名，其它 cert_scope 为空字符串 `""`。
- CertificateType=CA 且 CertEndTime 为空时，回退解析 CAEndTimes 取最早到期时间；仍无效则 INFO 日志 + `query_success=0`。
- 同一 CertificateId 被多个 NS 引用且使用不同凭证时，按各自 Binding 所属 NS 的凭证分别查询。
- 单次 DescribeCertificate 超过 1000 个 CertificateId 时分页查询。
- 指标不携带 `bcs_cluster_id` label（由蓝鲸监控平台自动添加）。
- SSL API 调用失败重试 3 次后仍失败，受影响 Binding 置 `query_success=0`。
- 指标写入失败不影响 Controller Reconcile 主流程与其它 Checker。
- 证书过期检测默认关闭；须 `--certificate_check_enabled=true` 或 Helm `certificateCheckEnabled: true` 显式开启。
- 关闭开关后不主动清理已有 Prometheus series；Pod 滚动重启后旧 series 自动消失。
- DescribeCertificate 每次云 API 调用前须经过与 CLB 共用的共享令牌桶限流（方案 A）。
- 实际云 API 为 `DescribeCertificate`（单数）逐 ID 轮询；`SSLClient.DescribeCertificates` 仅为上层封装方法名；本期不升级 Tencent Cloud Go SDK。
- SSL API endpoint 通过 `TENCENTCLOUD_SSL_DOMAIN` 全局配置，默认 `ssl.tencentcloudapi.com`；内网示例 `ssl.internal.tencentcloudapi.com`；per-NS Secret 不支持独立 SSL 域名。
- 证书过期检测开关关闭时，不初始化 SSL Client，域名配置无运行时影响。
- CertificateChecker 执行周期固定 60 分钟，不可通过 CLI / Helm 配置（规则 R-007）；其它 Checker 保持 1 分钟周期（规则 R-008）。
- 指标数据最大 staleness 由 10 分钟变为 60 分钟；Ingress 删除后相关 series 最长 60 分钟内清理。

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: 系统 MUST 从全集群 Ingress CR 展开 SSL 证书 Binding：仅处理 HTTPS、TCP_SSL、QUIC 协议；从 `spec.rules[].certificate`（cert_scope=rule）、`spec.rules[].layer7Routes[].certificate`（cert_scope=route）、`spec.portMappings[].certificate`（cert_scope=port_mapping）三处独立展开；certID 非空产生 `cert_role=server` Binding；mode=MUTUAL 且 certCaID 非空额外产生 `cert_role=client_ca` Binding；对 cert_id 去重供 API 逐 ID 查询。
- **FR-002**: 系统 MUST 对每个 CertificateId **单独调用**腾讯云 SSL `DescribeCertificate`（endpoint 由 `TENCENTCLOUD_SSL_DOMAIN` 指定，默认 `ssl.tencentcloudapi.com`，版本 2019-12-05）；**不调用** `DescribeCertificates`（复数）批量接口；`SSLClient.DescribeCertificates` 为上层批量语义封装，底层逐 ID 轮询；应用层单次最多 1000 ID 分页；解析 CertEndTime（GMT+8，格式 `2006-01-02 15:04:05`）；CA 类型 CertEndTime 为空时取 CAEndTimes 最早值；API 失败重试 3 次；**保持** `tencentcloud-sdk-go v1.0.132` 单体包，**本期不升级 SDK**（子模块 ≥ v1.0.1090 才支持 `CertIds` 批量，升级须联动 CLB/VPC/CVM 并引入 omitnil 行为变化风险）。
- **FR-003**: 系统 MUST 上报 Prometheus 指标：`bkbcs_ingressctrl_certificate_days_until_expiry`（Gauge，已过期为负数）、`bkbcs_ingressctrl_certificate_query_success`（Gauge，0/1）、`bkbcs_ingressctrl_certificate_bindings_total`（Gauge，无 label）；days_until_expiry 与 query_success 共用 8 个 label（owner_namespace、owner_name、cert_id、cert_role、cert_scope、protocol、port、domain）；不携带 bcs_cluster_id label。
- **FR-004**: 系统 MUST 每轮 Check 全量 List Ingress 重建 Binding 集合；对上轮存在、本轮不存在的 Binding 调用 DeleteLabelValues 清理 days_until_expiry 与 query_success series；query_success=0 或失效时间无效时删除 days_until_expiry 对应 series 但保留/设置 query_success=0。
- **FR-005**: 系统 MUST 在 Namespace Scope 模式（`--is_namespace_scope=true`）下按 Ingress 所在 NS 选择云凭证：普通 NS 使用 per-NS Secret / ControllerConfig 凭证；豁免 NS（`--namespace_scope_exempt_namespaces`）使用 Controller 全局凭证；按凭证分组逐 ID 调用 DescribeCertificate；单 NS 凭证失败仅影响该 NS 的 Binding。
- **FR-006**: 系统 MUST 新增独立 CertificateChecker，实现 `check.Checker` 接口，注册到 CheckRunner（周期 **CheckPer60Min，60 分钟**）；注册条件须**同时满足**：云厂商为腾讯云（`opts.Cloud == tencentcloud`）且 `--certificate_check_enabled=true`（**默认 `false`**）；未满足时**不初始化 SSL Client、不注册 CertificateChecker**；非腾讯云环境不注册（与开关无关）；DescribeCertificate 调用计入已有 `bkbcs_ingressctrl_api_*` / `bkbcs_ingressctrl_lib_*` 指标（**method=DescribeCertificate**）。
- **FR-007**: 系统 MUST 在 Helm Chart（`docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/`）新增 `certificateCheckEnabled`（默认 `false`）；`true` 时向 Deployment args 追加 `--certificate_check_enabled=true`；`false` 时不追加该参数（沿用 CLI 默认值）。
- **FR-008**: 系统 MUST NOT 从 CLB API 反查证书过期时间；MUST NOT 向 Ingress / Listener CR 回写过期时间字段；MUST NOT 自动续费或自动更新 Listener 上的 certID。
- **FR-009**: 系统 MUST 在每次 `DescribeCertificate` 云 API 调用**之前**执行限流：**方案 A**——抽取进程级共享 `throttle.RateLimiter`（`sync.Once` 初始化，读取 `TENCENTCLOUD_RATELIMIT_QPS` 默认 50、`TENCENTCLOUD_RATELIMIT_BUCKET_SIZE` 默认 50），`SdkWrapper` 与 `sslClientImpl` 共用同一实例；限流等待不阻塞 Reconcile 主流程（Checker 在独立 goroutine）；触发 `RequestLimitExceeded` 时沿用 CLB 既有重试/退避策略。
- **FR-010**: 系统 MUST NOT 在 AWS / GCP / Azure 环境注册 CertificateChecker 或产出 certificate 子系统指标。
- **FR-011**: SSL API 权限未开通或调用失败时，系统 MUST 置 query_success=0 并记录 ERROR 日志，MUST NOT 影响 Controller 其它功能（Reconcile、其它 Checker）。
- **FR-012**: 指标 series MUST NOT 包含 Secret、AccessKey 等敏感信息；仅暴露证书 ID 与过期天数。
- **FR-013**（对应 req.md F-010）: 系统 MUST 支持 SSL API 请求域名可配置：在 `internal/cloud/tencentcloud/` 定义 `EnvNameTencentCloudSslDomain = "TENCENTCLOUD_SSL_DOMAIN"`；`SSLClient` 创建 SDK Client 时从环境变量读取 endpoint，未配置或为空时使用 `ssl.tencentcloudapi.com`；Helm Chart 新增 `tencentcloudSslDomain`（默认 `ssl.tencentcloudapi.com`）并注入为 `TENCENTCLOUD_SSL_DOMAIN` env（对齐 `tencentcloudClbDomain` 模式）；所有 Controller 内腾讯云 SSL API 调用 MUST 复用此配置，禁止硬编码域名；NS Scope 下 per-NS Secret MUST NOT 携带独立 SSL 域名。
- **FR-014**（对应 req.md F-011）: 系统 MUST 将 CertificateChecker 调度周期由 10 分钟调整为 **60 分钟**：在 `internal/check/checkrunner.go` 新增 `CheckPer60Min` 档位及对应 ticker；`main.go` 注册 CertificateChecker 时使用 `CheckPer60Min`；**不新增** CLI / Helm 周期配置项；单轮 Check 业务逻辑不变；其它 Checker 的 1 分钟周期 MUST NOT 受影响。

### Key Entities

- **Certificate Binding**（逻辑实体，非 CRD）：表示 Ingress 上一处 SSL 证书挂载关系。关键属性：owner_namespace、owner_name、cert_id、cert_role（server/client_ca）、cert_scope（rule/route/port_mapping）、protocol、port、domain。同一 Ingress 可产生多条 Binding。
- **CertEndTime 映射**：CertificateId → 过期时间（Unix 时间戳），由 DescribeCertificate API 响应解析得到。
- **CertificateChecker**：周期性巡检组件，每 **60 分钟** List Ingress → 展开 Binding → 查询过期时间 → 更新/清理 Prometheus 指标；仅显式开启时注册。
- **SharedRateLimiter**（方案 A）：进程级共享腾讯云 API 令牌桶，`SdkWrapper` 与 `sslClientImpl` 共用，统一读取 `TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`。
- **SSL Endpoint 配置**：Controller 全局环境变量 `TENCENTCLOUD_SSL_DOMAIN`；Helm `tencentcloudSslDomain`；与 CLB 域名配置粒度一致。

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 运维人员可在 Prometheus / 蓝鲸监控中按 Ingress 维度（owner_namespace、owner_name）查看 SSL 证书剩余过期天数，成功查询的 series 标记 query_success=1。
- **SC-002**: 证书过期前运维可通过 `min by (owner_namespace, owner_name) (days_until_expiry)` 聚合找到同一 Ingress 下最早到期的证书。
- **SC-003**: SSL API 故障或权限不足时，受影响 Binding 的 query_success=0 且无 days_until_expiry series，运维可编写 `query_success == 1` 过滤条件避免误报。
- **SC-004**: Ingress 删除后，下一轮 Check（最长 **60 分钟**）内相关指标 series 被清理，无残留告警。
- **SC-005**: Namespace Scope 多租户场景下，各 NS 使用正确凭证查询，监控结果与 Listener 同步行为一致。
- **SC-006**: 500 个 Ingress、200 个唯一 certID 的集群，单次 Check 在 **60 分钟周期内**完成（含限流等待；200 次 DescribeCertificate @ 默认 50 QPS 理论下限约 4s，须预留凭证分组与重试余量）。
- **SC-007**: 10 个 NS 各 50 个 Ingress、凭证互不重叠时，按凭证分组完成查询，总耗时不超过 SC-006 基线的 2 倍。
- **SC-008**: Controller 其它功能（Ingress Reconcile、Listener 同步、现有 Checker）在证书过期检查失败时保持正常运行。
- **SC-009**: 未显式开启 `--certificate_check_enabled` 的集群升级后行为与默认关闭一致；显式开启后与既有指标行为相同。
- **SC-010**: DescribeCertificate 每次调用前均经过与 CLB 共用的共享令牌桶限流。
- **SC-011**: 内网部署场景下，运维可通过 `TENCENTCLOUD_SSL_DOMAIN` 或 Helm `tencentcloudSslDomain` 将 SSL API 指向内网 endpoint，证书过期查询与公网部署行为一致。
- **SC-012**: CertificateChecker 每 60 分钟触发一次 `Run()`；2 小时内执行 2 次（±启动偏差）；其它 1 分钟周期 Checker 不受影响。

## Assumptions

- 部署环境为腾讯云（tencentcloud）；开启功能前 Controller 凭证（全局或 per-NS）须开通 `ssl:DescribeCertificates` 只读权限。
- 证书过期检测**默认关闭**；运维在确认 SSL 权限就绪后须显式设置 `--certificate_check_enabled=true` 或 Helm `certificateCheckEnabled: true`。
- Ingress CR 中 SSL 证书配置遵循现有校验规则（certID 必填；MUTUAL 模式 certCaID 必填）。
- 蓝鲸监控平台通过既有 Prometheus scrape 机制采集 Controller metrics 端点，并自动添加 bcs_cluster_id label。
- 告警阈值与 PromQL 表达式由运维自行配置，本需求不提供推荐告警规则。
- 证书过期监控是对腾讯云账号级消息、云监控告警的补充手段，不替代云端兜底告警。
- 子需求交付顺序：先核心 Checker（全局凭证路径）后 NS Scope 凭证扩展，两者合并为本 spec 的完整范围。
- SDK 保持 `github.com/tencentcloud/tencentcloud-sdk-go v1.0.132`，本期不迁移子模块。
- 未配置 `TENCENTCLOUD_SSL_DOMAIN` 时 SSL endpoint 为 `ssl.tencentcloudapi.com`，与升级前硬编码行为一致。
- SSL API 请求域名为 Controller 全局配置；per-NS Secret 仅含凭证，不含 SSL 域名（规则 R-006）。

## Out of Scope

- AWS / GCP / Azure 证书过期监控
- 自动续费或自动更新 Listener certID
- 向 Ingress / Listener CR 回写过期时间字段
- 关闭开关时主动清理已有 Prometheus series
- 蓝鲸监控推荐告警阈值
- 指标携带 bcs_cluster_id label
- 升级 Tencent Cloud Go SDK 以使用 `DescribeCertificates` + `CertIds` 批量接口
