# Tasks: 证书过期时间 Prometheus 指标

**Input**: 设计文档 `specs/stories/1070046748135050873/`  
**Prerequisites**: plan.md, spec.md, data-model.md, research.md  
**技术方案**: 见 `specs/stories/1070046748135050873/plan.md`  
**TDD 模式**: 每个实现阶段先写失败测试（RED），再实现代码（GREEN）

**组织方式**: 任务按用户故事分组，每个阶段可独立实现和测试。US1~US4 为 P1，US5~US8 为 P1（第 2、3、4、5 轮澄清增量）。

**增量说明（2026-06-11）**: 五轮澄清（F-001~F-011）已全部实现并 squash 至 commit `62759c390`；三维度校验复核 LGTM（2026-06-11）。

## 格式说明: `[ID] [P?] [Story] 描述`

- **[P]**: 可与同阶段其他 [P] 任务并行执行（不同文件、无依赖）
- **[Story]**: 任务所属的用户故事（US1~US7），对应 spec.md 中的 User Story

## 路径约定

- 所有路径相对于工作区根 `bcs-runtime/bcs-k8s/bcs-network/bcs-ingress-controller/`

---

## Phase 1: Setup（项目初始化）

**目的**: 确认依赖与参考实现，为后续 TDD 开发做好准备。

- [x] T001 确认 `go.mod` 已包含腾讯云 SSL SDK 依赖 `github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl`，若缺失则添加并 `go mod tidy`
- [x] T002 [P] 阅读参考实现：`internal/check/listenerchecker.go`（`lastMetricMap` 清理模式）、`internal/check/checkrunner.go`（`CheckPer10Min` 注册）、`internal/metrics/check.go`（GaugeVec 注册惯例）
- [x] T003 [P] 阅读参考实现：`internal/cloud/namespacedlb/namespacedclient.go`（NS Scope 凭证路由模式）与 `docs/adr/0001-namespace-scope-exemption.md`（豁免 NS 规则）

---

## Phase 2: Foundational（指标与 Binding 数据模型 — 阻塞前置）

**目的**: 实现 Prometheus 指标定义与 Ingress Binding 展开纯函数，是所有用户故事的前置依赖。

**⚠️ 关键**: 所有用户故事的实现均依赖此阶段完成。

### Tests for Foundational（RED）

- [x] T004 [P] 编写 `internal/metrics/certificate_test.go`：断言 3 个 GaugeVec/Gauge 名称（`bkbcs_ingressctrl_certificate_days_until_expiry`、`bkbcs_ingressctrl_certificate_query_success`、`bkbcs_ingressctrl_certificate_bindings_total`）、8 个 label 维度、`init()` 注册到 `metrics.Registry`
- [x] T005 [P] 编写 `internal/check/certificatechecker_test.go` 前半部分（或 `internal/check/binding_test.go`）：覆盖 FR-001 Binding 展开场景——HTTPS/MUTUAL/UNIDIRECTIONAL/SNI/三处 cert_scope（rule/route/port_mapping）/非 SSL 协议跳过

### Implementation for Foundational（GREEN）

- [x] T006 实现 `internal/metrics/certificate.go`：定义 3 个指标变量、`init()` 中 `metrics.Registry.MustRegister`、导出 `SetCertificateDaysUntilExpiry`、`SetCertificateQuerySuccess`、`SetCertificateBindingsTotal`、`DeleteCertificateMetrics` helper 函数
- [x] T007 实现 `internal/check/binding.go`：定义 `CertificateBinding` 结构体、`BindingKey()` 方法（8 label 唯一键）、`expandBindings(ingressList []networkextension.Ingress) []CertificateBinding` 纯函数，覆盖 FR-001 全部展开规则

**检查点**: 指标可注册、Binding 展开逻辑可通过单元测试——后续 SSL 查询与 Checker 实现可以开始。

---

## Phase 3: User Story 1 — 查看 Ingress 证书剩余过期天数 (Priority: P1) 🎯 MVP

**目标**: 运维人员可在 Prometheus 中按 Ingress 维度查看 SSL 证书剩余过期天数（`days_until_expiry` + `query_success=1`）。

**独立测试**: 在腾讯云集群创建带有效 certID 的 HTTPS Ingress，等待 CertificateChecker 执行一轮 Check，验证 `bkbcs_ingressctrl_certificate_days_until_expiry` 出现对应 series 且 `query_success=1`，天数与 CertEndTime 一致（误差 < 1 天）。

### Tests for User Story 1（RED）

- [x] T008 [P] [US1] 编写 `internal/cloud/tencentcloud/sslclient_test.go`：分页（>1000 ID）、重试 3 次、`CertEndTime` GMT+8 解析（`2006-01-02 15:04:05`）、CA 类型 `CAEndTimes` 回退、mock `ReportLibRequestMetric`
- [x] T009 [P] [US1] 编写 `internal/check/certificatechecker_test.go` 核心场景：fake client 注入 Ingress fixture，验证 `days_until_expiry` 写入、`query_success=1`、`bindings_total` 更新、MUTUAL 模式产生 server/client_ca 双 series、已过期证书为负数

### Implementation for User Story 1（GREEN）

- [x] T010 [US1] 实现 `internal/cloud/tencentcloud/sslclient.go`：`SSLClient` 接口、`NewSSLClient()`、`NewSSLClientWithSecretIDKey()`、`DescribeCertificates(certIDs []string) (map[string]int64, error)`，底层逐 ID 调用 `DescribeCertificate`；单次最多 1000 ID 分页、失败重试 3 次（*指标 method 与限流见 Phase 9 T036*）
- [x] T011 [US1] 实现 `internal/check/certificatechecker.go` 骨架：`CertificateChecker` 结构体（cli、sslClient、opts、lastBindingSet）、`NewCertificateChecker()` 构造函数、实现 `check.Checker` 接口的 `Run()` 方法——List Ingress → expandBindings → collectUniqueCertIDs → DescribeCertificates（全局凭证）→ updateMetrics → 设置 bindings_total

**检查点**: 全局凭证路径下，HTTPS Ingress 的证书过期天数可正确上报到 Prometheus。

---

## Phase 4: User Story 2 — 区分查询成功与查询失败 (Priority: P1)

**目标**: 通过 `query_success` label 区分「证书即将过期」与「过期时间查询失败」，避免 SSL API 故障误报。

**独立测试**: 模拟 SSL API 权限未开通或连续 3 次调用失败，验证受影响 Binding 的 `query_success=0` 且无 `days_until_expiry` series，Controller 其它功能正常。

### Tests for User Story 2（RED）

- [x] T012 [P] [US2] 扩展 `internal/check/certificatechecker_test.go`：API 连续失败 → `query_success=0` 且 `days_until_expiry` 被 DeleteLabelValues；证书 ID 在响应中缺失 → `query_success=0`；CertEndTime 无效 → `query_success=0` + INFO/ERROR 日志断言
- [x] T013 [P] [US2] 扩展 `internal/cloud/tencentcloud/sslclient_test.go`：重试 3 次后仍失败返回 error、CA 类型 CAEndTimes 仍无效时返回空映射

### Implementation for User Story 2（GREEN）

- [x] T014 [US2] 完善 `internal/check/certificatechecker.go` 的 `updateMetrics()`：`query_success=1` 时 Set days + success；`query_success=0` 时 Delete days_until_expiry 但 Set query_success=0；API 批量失败时按 certID 逐个标记失败；记录 blog INFO/ERROR 日志（FR-011）
- [x] T015 [US2] 完善 `internal/cloud/tencentcloud/sslclient.go` 错误处理：重试耗尽后返回 error、无效时间解析跳过并记录日志

**检查点**: PromQL 可基于 `query_success == 1` 过滤，仅对成功查询的 series 评估过期阈值。

---

## Phase 5: User Story 3 — 指标随 Ingress 生命周期自动清理 (Priority: P1)

**目标**: Ingress 删除或证书配置变更时，上一轮残留的 Prometheus series 被自动清理。

**独立测试**: 创建 Ingress 并等待指标出现，删除 Ingress 后等待下一轮 Check，验证该 Ingress 相关 series 被 DeleteLabelValues 清理。

### Tests for User Story 3（RED）

- [x] T016 [P] [US3] 扩展 `internal/check/certificatechecker_test.go`：上轮存在 Binding、本轮消失 → `days_until_expiry` 与 `query_success` 均被 DeleteLabelValues；证书配置移除场景；`bindings_total` 更新为当前轮总数；List Ingress 失败 → 本轮终止、不更新指标、记录 ERROR 日志

### Implementation for User Story 3（GREEN）

- [x] T017 [US3] 实现 `internal/check/certificatechecker.go` 的 `cleanupStaleMetrics()`：维护 `lastBindingSet map[string]struct{}`，对 `lastBindingSet - currentBindingSet` 调用 `DeleteCertificateMetrics`；每轮结束更新 `lastBindingSet`；List Ingress 失败时提前返回不修改指标

**检查点**: Ingress 删除后最长 **60 分钟**内相关指标 series 被清理，无残留告警。

---

## Phase 6: User Story 4 — Namespace Scope 多租户凭证查询 (Priority: P1)

**目标**: Namespace Scope 模式下各 NS 使用各自云凭证查询证书过期时间，豁免 NS 使用全局凭证。

**独立测试**: 启用 `--is_namespace_scope=true`，NS-A 使用独立 Secret 凭证且 Ingress 引用 NS-A 账号下证书，验证使用 NS-A 凭证查询且 `query_success=1`；豁免 NS 验证使用全局凭证。

### Tests for User Story 4（RED）

- [x] T018 [P] [US4] 编写 `internal/cloud/namespacedssl/namespacedclient_test.go`：豁免 NS 使用 defaultClient、普通 NS 使用 per-NS Secret 构造 SSLClient、凭证缺失仅影响该 NS、按凭证分组批量查询
- [x] T019 [P] [US4] 扩展 `internal/check/certificatechecker_test.go` NS Scope 场景：按 owner_namespace 分组 certID、不同 NS 凭证隔离、同一 certID 多 NS 不同凭证分别查询

### Implementation for User Story 4（GREEN）

- [x] T020 [US4] 实现 `internal/cloud/namespacedssl/namespacedclient.go`：`NamespacedSSL` 结构体（k8sClient、nsClientSet、defaultClient、exemptNamespaces）、`NewNamespacedSSL()`、`getNsClient(ns string) SSLClient`，镜像 `namespacedlb` 的凭证路由逻辑
- [x] T021 [US4] 扩展 `internal/check/certificatechecker.go`：注入 `namespacedSSL *namespacedssl.NamespacedSSL`；`opts.IsNamespaceScope` 分支调用 `groupByCredential()` 按 NS 分组批量查询；非 NS Scope 保持全局凭证路径（FR-005）

**检查点**: 多租户场景下各 NS 使用正确凭证查询，监控结果与 Listener 同步行为一致。

---

## Phase 7: User Story 5 — CertificateChecker 集成与 CLI/Helm 开关 (Priority: P1)

**目标**: 腾讯云环境下通过 `--certificate_check_enabled`（默认 `false`）与 Helm `certificateCheckEnabled` 显式控制 CertificateChecker 注册；非腾讯云环境不注册。

**独立测试**: `--certificate_check_enabled=true` 时验证 Checker 注册到 `checkPer10Min`；未指定或 `false` 时不 init SSL Client、不注册；Helm values 正确传递 args。

> **⚠️ 第 1 轮实现偏差**: T022/T023 按「默认开启、无 CLI 开关」完成；Phase 9（T027~T029）须修正。

### Tests for User Story 5（RED）— 第 1 轮

- [x] T022 [P] [US5] 编写 `internal/check/certificatechecker_test.go` 或 `main_test.go`：断言 `opts.Cloud == tencentcloud` 时 Checker 注册到 `CheckPer10Min`；非 tencentcloud 不注册；DescribeCertificates 成功时 `bkbcs_ingressctrl_lib_request_total{method="DescribeCertificates"}` 递增

### Implementation for User Story 5（GREEN）— 第 1 轮

- [x] T023 [US5] 修改 `main.go`：当 `opts.Cloud == constant.CloudProviderTencent` 时构造 `SSLClient`、`NamespacedSSL`（若 NS Scope）、`NewCertificateChecker(...)` 并 `checkRunner.Register(certChecker, check.CheckPer10Min)`；非腾讯云跳过注册（FR-006、FR-010）

**检查点（第 1 轮）**: 腾讯云环境每 10 分钟自动巡检——**待 Phase 9 修正为默认关闭 + 显式开启**。

---

## Phase 8: Polish & Cross-Cutting Concerns

**目的**: 全量测试、覆盖率验证与代码质量收尾。

- [x] T024 [P] 运行全量测试：`cd .. && make test-ingress-controller` 及 `go test -v -run 'TestExpandBindings|TestCertificateChecker|TestDescribeCertificates|TestNamespacedSSL' ./internal/check/... ./internal/metrics/... ./internal/cloud/tencentcloud/... ./internal/cloud/namespacedssl/...`
- [x] T025 [P] 确认核心逻辑（binding 展开、SSL 查询、指标更新、NS Scope 路由）单元测试覆盖率 ≥ 80%，函数圈复杂度 ≤ 15
- [x] T026 代码审查自检：blog 日志（禁止 log/klog）、英文 GoDoc 注释、错误显式处理、指标不携带 `bcs_cluster_id` label、series 不含 Secret/AccessKey 等敏感信息（FR-012）

---

## Phase 9: 第 2、3 轮澄清增量（CLI 开关 + 限流方案 A） 🆕

**目的**: 同步 `req.md` 第 2 轮（默认关闭 + CLI/Helm 开关）与第 3 轮（DescribeCertificate 限流方案 A、SDK 不升级、method 更名）至代码与测试。

**前置**: Phase 1~8 已完成；本阶段修正第 1 轮与澄清结论的偏差。

### User Story 5 增量 — CLI/Helm 开关（第 2 轮）

#### Tests（RED）

- [x] T027 [P] [US5] 扩展 `internal/check/certificatechecker_test.go`：`ShouldRegisterCertificateChecker(cloud, enabled)` 在 `enabled=false` 时返回 false；`cloud!=tencentcloud` 时无论 enabled 均返回 false；`cloud=tencentcloud && enabled=true` 返回 true
- [x] T028 [P] [US5] 扩展 `internal/option/option_test.go`（或新建）：`--certificate_check_enabled` 默认值 `false`、flag 解析为 `CertificateCheckEnabled`

#### Implementation（GREEN）

- [x] T029 [US5] 修改 `internal/option/option.go`：新增 `CertificateCheckEnabled bool` 字段与 `--certificate_check_enabled` flag（默认 `false`）
- [x] T030 [US5] 修改 `internal/check/certificatechecker.go`：`ShouldRegisterCertificateChecker(cloud string, enabled bool) bool` 双条件判断
- [x] T031 [US5] 修改 `main.go`：仅 `ShouldRegisterCertificateChecker(opts.Cloud, opts.CertificateCheckEnabled)` 为真时 init SSL Client 并注册 Checker；关闭时不 init、不因 SSL 凭证缺失 `os.Exit(1)`
- [x] T032 [P] [US5] 修改 Helm Chart（`docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/`）：`values.yaml` 新增 `certificateCheckEnabled: false`；`templates/deployment.yaml` 条件追加 `--certificate_check_enabled=true`（FR-007）

**检查点**: AC-010~AC-014、AC-T01/T02 通过；无 SSL 权限集群升级后 Controller 正常启动（默认关闭）。

### User Story 6 增量 — DescribeCertificate 限流方案 A（第 3 轮）

**限流方案 A 设计**:
- 新增 `internal/cloud/tencentcloud/sharedratelimit.go`（或等价文件）：`sync.Once` 初始化进程级共享 `throttle.RateLimiter`，从 `TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE` 读取（默认 50/50，与 `SdkWrapper.loadEnv` 逻辑一致）
- 重构 `SdkWrapper` 与 `APIWrapper`：构造时改用 `GetSharedRateLimiter()`，移除各自独立 `NewTokenBucket` 实例（或保留构造但指向共享实例）
- `sslClientImpl.doDescribeCertificates`：每次 `c.api.DescribeCertificate` **之前**调用 `GetSharedRateLimiter().Accept()`
- `ReportLibRequestMetric` method 改为 `"DescribeCertificate"`（与实际云 API 一致）

#### Tests（RED）

- [x] T033 [P] [US6] 扩展 `internal/cloud/tencentcloud/sslclient_test.go`：mock 共享限流器，验证 N 个 certID 产生 ≥ N 次 `Accept()`；指标 method 为 `DescribeCertificate`
- [x] T034 [P] [US6] 扩展 `internal/cloud/tencentcloud/sharedratelimit_test.go`（或 sslclient_test）：`GetSharedRateLimiter()` 多次调用返回同一实例；env 变量正确解析

#### Implementation（GREEN）

- [x] T035 [US6] 实现 `sharedratelimit.go`：`GetSharedRateLimiter() throttle.RateLimiter` + `trySharedThrottle()` helper；重构 `SdkWrapper`/`APIWrapper` 接入共享实例（FR-009 方案 A）
- [x] T036 [US6] 修改 `sslclient.go`：`doDescribeCertificates` 循环内每次 API 前 `trySharedThrottle()`；`ReportLibRequestMetric` method 改为 `DescribeCertificate`；文件顶部 GoDoc 注释说明 SDK v1.0.132 不升级、逐 ID 轮询原因（FR-002）
- [x] T037 [P] [US6] 更新 T022 相关断言：`method="DescribeCertificate"` 替代 `DescribeCertificates`

**检查点**: AC-P02、AC-T03、AC-T06 通过；CLB 与 SSL 路径共用同一 QPS 配额。

### Phase 9 收尾

- [x] T038 [P] 运行回归测试：`go test -v -run 'TestShouldRegisterCertificateChecker|TestDescribeCertificates|TestSharedRateLimiter' ./internal/check/... ./internal/cloud/tencentcloud/... ./internal/option/...`
- [x] T039 重新执行 validate（架构 / 安全 / CodeReview），确认增量符合 FR-006/FR-007/FR-009

---

## Phase 10: 第 4 轮澄清增量（SSL API 域名可配置） 🆕

**目的**: 同步 `req.md` 第 4 轮 F-010 至代码与 Helm：SSL endpoint 通过 `TENCENTCLOUD_SSL_DOMAIN` 可配置，对齐 CLB `tencentcloudClbDomain` 模式。

**前置**: Phase 9 已完成；`sslclient.go` 当前硬编码 `sslEndpoint = "ssl.tencentcloudapi.com"`，须改为环境变量读取。

**参考实现**:
- CLB 常量：`internal/cloud/tencentcloud/sdk.go` — `EnvNameTencentCloudClbDomain`
- CLB Helm 注入：`values.yaml` `tencentcloudClbDomain` → `deployment.yaml` `TENCENTCLOUD_CLB_DOMAIN` env
- SSL 设计：`resolveSSLEndpoint()` 读取 `TENCENTCLOUD_SSL_DOMAIN`，空值回退 `ssl.tencentcloudapi.com`

### User Story 7 增量 — SSL API 域名可配置（第 4 轮）

#### Tests（RED）

- [X] T040 [P] [US7] 扩展 `internal/cloud/tencentcloud/sslclient_test.go`：env 未设置时 `NewSSLClientWithSecretIDKey` 的 SDK `HttpProfile.Endpoint` 为 `ssl.tencentcloudapi.com`（AC-015）
- [X] T041 [P] [US7] 扩展 `sslclient_test.go`：设置 `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com` 时 endpoint 为内网域名（AC-016、AC-T07）
- [X] T042 [P] [US7] 新增 `resolveSSLEndpoint` 单元测试（可 colocate 于 `sslclient_test.go`）：env 为空/未设置 → 默认；env 有值 → 自定义；禁止在业务代码硬编码域名

#### Implementation（GREEN）

- [X] T043 [US7] 修改 `internal/cloud/tencentcloud/sdk.go`：新增常量 `EnvNameTencentCloudSslDomain = "TENCENTCLOUD_SSL_DOMAIN"`（与 CLB 命名风格一致）
- [X] T044 [US7] 修改 `internal/cloud/tencentcloud/sslclient.go`：抽取 `resolveSSLEndpoint()` 读取 env；`NewSSLClientWithSecretIDKey` 使用 `cpf.HttpProfile.Endpoint = resolveSSLEndpoint()` 替代硬编码 `sslEndpoint` 常量；保留 `sslEndpoint` 仅作默认值（FR-013）
- [X] T045 [P] [US7] 修改 Helm Chart（`docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/`）：`values.yaml` 新增 `tencentcloudSslDomain: ssl.tencentcloudapi.com`；`templates/deployment.yaml` 在 env 段注入 `TENCENTCLOUD_SSL_DOMAIN`（对齐 `TENCENTCLOUD_CLB_DOMAIN` 注入方式）（AC-017）

**检查点**: AC-015~AC-017、AC-T07 通过；未配置 env 时行为与升级前一致；内网场景可指向 `ssl.internal.tencentcloudapi.com`。

### Phase 10 收尾

- [X] T046 [P] 运行回归测试：`go test -v -run 'TestResolveSSLEndpoint|TestSSLClientEndpoint|TestDescribeCertificates' ./internal/cloud/tencentcloud/...`
- [X] T047 重新执行 validate（架构 / 安全 / CodeReview），确认增量符合 FR-013；检查 `namespacedssl` per-NS 路径未引入独立 SSL 域名逻辑

---

## Phase 11: 第 5 轮澄清增量（检查周期 10min→60min） 🆕

**目的**: 同步 `req.md` 第 5 轮 F-011 至代码：`CertificateChecker` 调度周期由 10 分钟调整为 60 分钟。

**前置**: Phase 10 已完成；`main.go` 当前 `checkRunner.Register(certChecker, check.CheckPer10Min)`，须迁移至 `CheckPer60Min`。

**设计要点**:
- `internal/check/checkrunner.go` 新增 `CheckPer60Min` 常量及 `checkPer60Min []Checker` + `ticker60Min`（`time.Hour`）
- `CheckPer10Min` 保留但 CertificateChecker 不再使用（避免破坏 enum 序号）
- 不新增 CLI / Helm 配置项；单轮 Check 逻辑不变

### User Story 8 增量 — 检查周期调整（第 5 轮）

#### Tests（RED）

- [X] T048 [P] [US8] 扩展 `internal/check/checkrunner_test.go`（或新建）：`Register(..., CheckPer60Min)` 的 Checker 仅由 60 分钟 ticker 触发；`CheckPerMin` Checker 仍由 1 分钟 ticker 触发（AC-021）
- [X] T049 [P] [US8] 扩展 `internal/check/certificatechecker_test.go`：断言 `ShouldRegisterCertificateChecker` 为真时注册 interval 为 `CheckPer60Min`（非 `CheckPer10Min`）（AC-018）

#### Implementation（GREEN）

- [X] T050 [US8] 修改 `internal/check/checkrunner.go`：新增 `CheckPer60Min`；`Register` switch 分支；`Start()` 增加 `ticker60Min := time.NewTicker(time.Hour)` 及对应 select case（FR-014）
- [X] T051 [US8] 修改 `main.go`：`checkRunner.Register(certChecker, check.CheckPer60Min)` 替代 `CheckPer10Min`（FR-014）

**检查点**: AC-018~AC-021、AC-P03、AC-T01（修订为 60 分钟）通过；其它 Checker 1 分钟周期不变。

### Phase 11 收尾

- [X] T052 [P] 运行回归测试：`go test -v -run 'TestCheckRunner|TestShouldRegisterCertificateChecker' ./internal/check/...`
- [X] T053 重新执行 validate（架构 / 安全 / CodeReview），确认增量符合 FR-014；文档中「10 分钟」描述已同步为「60 分钟」

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: 无依赖，可立即开始
- **Foundational (Phase 2)**: 依赖 Setup — **阻塞所有用户故事**
- **US1 (Phase 3)**: 依赖 Foundational — MVP 核心
- **US2 (Phase 4)**: 依赖 US1（共用 CertificateChecker 与 sslclient）
- **US3 (Phase 5)**: 依赖 US1（共用 CertificateChecker）
- **US4 (Phase 6)**: 依赖 US1（在 Checker 上扩展 NS Scope 分支）
- **US5 (Phase 7 + 9)**: 依赖 US1~US4 核心逻辑；Phase 9 修正开关 gating
- **US6 (Phase 9)**: 依赖 US1 sslclient 骨架；可与 US5 Phase 9 并行（不同文件）
- **Polish (Phase 8)**: 依赖 Phase 1~7 完成
- **Phase 9 增量**: 依赖 Phase 1~8；修正第 1 轮偏差
- **Phase 10 增量**: 依赖 Phase 9；F-010 SSL 域名可配置
- **Phase 11 增量**: 依赖 Phase 10；F-011 检查周期 60 分钟

### User Story Dependencies

```text
Foundational ──► US1 (MVP) ──┬──► US2
                              ├──► US3
                              └──► US4 ──► US5 (Phase 7)
                                              │
                                              ▼
                                    Phase 9 ──┬── US5 开关增量 (T027~T032)
                                              └── US6 限流增量 (T033~T037)
                                              │
                                              ▼
                                    Phase 10 ─── US7 SSL 域名 (T040~T047)
                                              │
                                              ▼
                                    Phase 11 ─── US8 检查周期 (T048~T053)
```

- **US1**: Foundational 完成后即可开始，无其他故事依赖
- **US2/US3**: 可与 US4 在 US1 完成后并行（不同文件/函数）
- **US4**: 依赖 US1 的 Checker 骨架，可与 US2/US3 并行
- **US5**: 需 US1~US4 逻辑就绪；Phase 9 修正 gating
- **US6**: 依赖 US1 的 sslclient；Phase 9 与 US5 增量可并行
- **US7**: 依赖 Phase 9 完成的 sslclient；仅修改 endpoint 解析与 Helm env 注入
- **US8**: 依赖 Phase 10；仅修改 CheckRunner 调度档位与 main.go 注册 interval

### Within Each User Story（TDD）

- 测试任务（RED）必须先于实现任务（GREEN）
- 同故事内标记 [P] 的测试可并行编写
- 实现任务按依赖顺序：sslclient → checker 核心 → 分支扩展

### Parallel Opportunities

```bash
# Phase 2 并行编写测试：
T004: internal/metrics/certificate_test.go
T005: internal/check/binding_test.go

# Phase 3 并行编写测试：
T008: internal/cloud/tencentcloud/sslclient_test.go
T009: internal/check/certificatechecker_test.go (US1 场景)

# US1 完成后 US2/US3/US4 可并行：
Developer A: T012-T015 (US2 query_success)
Developer B: T016-T017 (US3 cleanup)
Developer C: T018-T021 (US4 NS Scope)
```

---

## Parallel Example: User Story 1

```bash
# Step 1: 并行编写 RED 测试
Task T008: "sslclient_test.go — 分页/重试/时间解析"
Task T009: "certificatechecker_test.go — 指标写入/bindings_total"

# Step 2: 顺序实现 GREEN
Task T010: "sslclient.go — DescribeCertificates"
Task T011: "certificatechecker.go — Run() 全局凭证路径"
```

---

## Implementation Strategy

### MVP First（仅 User Story 1）

1. 完成 Phase 1: Setup
2. 完成 Phase 2: Foundational（指标 + Binding 展开）
3. 完成 Phase 3: User Story 1（SSL 查询 + 指标上报）
4. **停止并验证**: 全局凭证路径下 HTTPS Ingress 证书天数可查询
5. 可先交付子需求 #1070046748135054749（核心路径）

### Incremental Delivery

1. Setup + Foundational → 基础就绪
2. US1 → 证书天数可查（MVP）
3. US2 → 查询失败可区分
4. US3 → 指标自动清理
5. US4 → NS Scope 多租户（子需求 #1070046748135054806）
6. US5 → Checker 注册（Phase 7，第 1 轮）
7. Polish → 全量测试通过（Phase 8）
8. **Phase 9** → 第 2 轮开关 + 第 3 轮限流方案 A（修正第 1 轮偏差）
9. **Phase 10** → 第 4 轮 SSL API 域名可配置（F-010）
10. **Phase 11** → 第 5 轮检查周期 10min→60min（F-011）

### 子需求交付顺序

| 子需求 | 覆盖 Phase | 说明 |
|--------|-----------|------|
| #1070046748135054749 | Phase 1~3 + US2/US3 + US5 + **Phase 9** + **Phase 10** + **Phase 11** | 全局凭证核心路径 + 开关/限流/SSL 域名/60 分钟周期 |
| #1070046748135054806 | Phase 6 (US4) + **Phase 9** + **Phase 10** + **Phase 11** | NS Scope 凭证扩展（SSL 域名为全局配置；周期 60 分钟） |

---

## Notes

- [P] 任务 = 不同文件、无未完成依赖
- [Story] 标签映射 spec.md 用户故事，便于追溯
- 每个用户故事应可独立测试验证
- TDD：先确认测试失败（RED），再实现（GREEN）
- 参照 `ListenerChecker.lastMetricMap` 模式实现 `lastBindingSet`
- `namespacedssl` 独立包，避免 cloud 子包互相引用
- 指标写入失败 MUST NOT 影响 Reconcile 与其它 Checker（FR-011）
- Phase 9 限流 MUST 使用方案 A（共享 `GetSharedRateLimiter()`），禁止 sslClient 独立 QPS 桶
- Helm Chart 路径在 monorepo `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/`
- Phase 10 SSL 域名 MUST 为 Controller 全局 env；`namespacedssl` MUST NOT 读取 per-NS SSL 域名
- Phase 11 周期 MUST 为固定 60 分钟；MUST NOT 新增 CLI/Helm 周期配置项；MUST NOT 影响其它 Checker 的 1 分钟周期

---

## Task Summary

| 指标 | 数值 |
|------|------|
| 总任务数 | 53（T001~T047 已完成，T048~T053 待办） |
| Setup | 3 ✅ |
| Foundational | 4 ✅ |
| US1 (P1) | 4 ✅ |
| US2 (P1) | 4 ✅ |
| US3 (P1) | 2 ✅ |
| US4 (P1) | 4 ✅ |
| US5 (P1) | 8 ✅ |
| US6 (P1) | 5 ✅ |
| US7 (P1) | 8 ✅ |
| US8 (P1) | 6（T048~T053，待办） |
| Polish | 3 ✅ |
| Phase 9 收尾 | 2 ✅ |
| Phase 10 收尾 | 2 ✅ |
| Phase 11 收尾 | 2 ✅ |
| 可并行任务 [P] | 14（Phase 1~8）+ 6（Phase 9）+ 4（Phase 10）+ 3（Phase 11） |
| MVP 范围 | Phase 1~3（T001~T011） |
| **当前待办** | **无（五轮澄清已 squash 至 commit 62759c390）** |
