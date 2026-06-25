# Implementation Plan: 证书过期时间 Prometheus 指标

**Branch**: `stories/1070046748135050873` | **Date**: 2026-06-09 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `specs/stories/1070046748135050873/spec.md`

## Summary

为 BCS Ingress Controller 新增证书过期监控子系统：实现 `CertificateChecker` 周期性（10 分钟）List 全集群 Ingress、展开 SSL 证书 Binding、调用腾讯云 SSL `DescribeCertificates` 批量查询过期时间，并上报/清理 Prometheus 指标。复用现有 `internal/check` Checker 模式与 `internal/metrics` 注册惯例；Namespace Scope 模式下通过 `internal/cloud/namespacedssl/` 镜像 `namespacedlb` 凭证路由。仅 `opts.Cloud == tencentcloud` 时注册 Checker，无 CLI 开关。

## Technical Context

**Language/Version**: Go 1.20+
**Primary Dependencies**: controller-runtime v0.6.3, prometheus/client_golang, 腾讯云 SSL SDK（`github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssl`）
**Storage**: Kubernetes etcd（Ingress CR 只读）；内存 `lastBindingSet`（指标清理状态）
**Testing**: go test（表驱动 + fake client）；`certificatechecker_test.go`、`sslclient_test.go`、`namespacedssl_test.go`
**Target Platform**: Linux（Kubernetes 集群，腾讯云 CLB 环境）
**Project Type**: Kubernetes Controller/Operator 扩展（Checker 子系统）
**Performance Goals**: 500 Ingress / 200 唯一 certID 集群单次 Check < 5 分钟（SC-006）
**Constraints**: 不修改 Ingress/Listener CR；不引入 AWS/GCP/Azure 实现；指标不携带 `bcs_cluster_id`；函数圈复杂度 ≤ 15
**Scale/Scope**: 全集群 Ingress 全量扫描；DescribeCertificates 单次最多 1000 ID 分页；NS Scope 按凭证分组批量查询

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| 原则 | 状态 | 备注 |
|------|------|------|
| I. Go 语言工程规范 | PASS | 显式 error 处理；blog 日志；context 管理 goroutine |
| II. Kubernetes 控制器开发规范 | PASS | Checker 非 Reconcile，走独立周期巡检模式（与 ListenerChecker 一致） |
| III. 代码质量 | PASS | Binding 展开、SSL 查询、指标更新拆分子函数；核心逻辑目标覆盖率 ≥ 80% |
| IV. 英文代码注释 | PASS | 导出类型/函数英文 GoDoc |
| V. 中文沟通语言 | PASS | 计划文档中文 |

**架构决策要点**（基于 spec Clarifications 2026-06-09）：

1. **Checker 而非 Reconcile**：证书过期为只读巡检，不修改 CR，采用 `check.Checker` + `CheckPer10Min`，避免引入新 Reconciler
2. **不经过 generator 层**：Checker 直接 List Ingress 并展开 Binding，与 Listener 同步解耦
3. **SSL 客户端独立封装**：`internal/cloud/tencentcloud/sslclient.go`，复用 `SdkWrapper` 重试与 `ReportLibRequestMetric` 模式
4. **NS Scope 镜像 namespacedlb**：`internal/cloud/namespacedssl/` 独立包，豁免 NS 用 `defaultClient`，避免 cloud 子包互相引用
5. **指标清理策略**：`lastBindingSet` 维护上轮 Binding key 集合，参照 `ListenerChecker.lastMetricMap` 做 `DeleteLabelValues`
6. **分阶段交付**：先全局凭证路径（子需求 #1070046748135054749），再 NS Scope 凭证扩展（子需求 #1070046748135054806）

## Project Structure

### Documentation (this feature)

```text
specs/stories/1070046748135050873/
├── spec.md
├── plan.md              # 本文件
├── research.md          # Phase 0 技术调研
├── data-model.md        # Phase 1 数据模型
├── plan-report.md       # 合规自检报告
└── tasks.md             # Phase 2（/speckit.tasks 产出，本阶段不创建）
```

### Source Code (repository root)

```text
bcs-ingress-controller/
├── internal/
│   ├── check/
│   │   ├── certificatechecker.go       # CertificateChecker 实现
│   │   ├── certificatechecker_test.go  # Binding 展开 + 指标更新单元测试
│   │   └── binding.go                  # CertificateBinding 类型与展开逻辑（可选拆分）
│   ├── metrics/
│   │   ├── certificate.go              # 3 个 GaugeVec + init 注册
│   │   └── certificate_test.go         # 指标 label/helper 测试
│   └── cloud/
│       ├── tencentcloud/
│       │   ├── sslclient.go            # SSLClient + DescribeCertificates
│       │   └── sslclient_test.go       # 分页/重试/时间解析测试
│       └── namespacedssl/
│           ├── namespacedclient.go     # 镜像 namespacedlb 凭证路由
│           └── namespacedclient_test.go
└── main.go                             # tencentcloud 条件下注册 CertificateChecker
```

**Structure Decision**: 全部新增代码落在 `internal/check`、`internal/metrics`、`internal/cloud/tencentcloud`、`internal/cloud/namespacedssl` 四层，符合分层架构（check 层依赖 cloud + metrics，cloud 子包互不引用）。`main.go` 仅追加注册逻辑，不修改 `initClient` 复杂度。

## Implementation Phases（TDD）

### Phase A：指标与数据模型（红→绿）

1. **RED**：编写 `certificate_test.go`，断言 3 个 GaugeVec 名称、label 维度（8 label）、`init()` 注册
2. **GREEN**：实现 `internal/metrics/certificate.go`
3. **RED**：编写 `binding_test.go`（或 `certificatechecker_test.go` 前半），覆盖 FR-001 展开场景（HTTPS/MUTUAL/SNI/三处 cert_scope/非 SSL 跳过）
4. **GREEN**：实现 Binding 展开纯函数 `expandBindings(ingressList)`

### Phase B：SSL 客户端（红→绿）

1. **RED**：`sslclient_test.go` — 分页（>1000 ID）、重试 3 次、CertEndTime 解析（GMT+8）、CA 类型 CAEndTimes 回退
2. **GREEN**：`sslclient.go` — `NewSSLClient()`、`NewSSLClientWithSecretIDKey()`、`DescribeCertificates(certIDs []string)`
3. 集成 `metrics.ReportLibRequestMetric(system="tencentcloud", method="DescribeCertificates")`

### Phase C：CertificateChecker 核心（红→绿）

1. **RED**：`certificatechecker_test.go` — fake client 注入 Ingress fixture，验证指标写入、`query_success` 分支、`DeleteLabelValues` 清理、`bindings_total`
2. **GREEN**：`certificatechecker.go` — `Run()` 流程：List → expand → query（全局凭证）→ setMetric → cleanup
3. **RED**：List Ingress 失败终止、query_success=0 删除 days_until_expiry 保留 query_success

### Phase D：NS Scope 凭证扩展（红→绿）

1. **RED**：`namespacedssl_test.go` — 镜像 `namespacedlb` 测试：豁免 NS 用 defaultClient、per-NS Secret、凭证缺失
2. **GREEN**：`namespacedssl/namespacedclient.go` — `getNsClient(ns)` + 按凭证分组批量查询
3. **RED**：CertificateChecker NS Scope 模式集成测试（按 NS 分组 certID）
4. **GREEN**：CertificateChecker 注入 `NamespacedSSL`，`opts.IsNamespaceScope` 分支

### Phase E：main.go 注册与集成验证

1. `opts.Cloud == tencentcloud` 时 `checkRunner.Register(NewCertificateChecker(...), check.CheckPer10Min)`
2. 非 tencentcloud 不注册（断言无 certificate 指标产出）
3. 全量测试：`go test ./internal/check/... ./internal/metrics/... ./internal/cloud/tencentcloud/... ./internal/cloud/namespacedssl/...`

## Key Design Details

### Binding Key 与指标 Label

8 个 label：`owner_namespace`、`owner_name`、`cert_id`、`cert_role`、`cert_scope`、`protocol`、`port`、`domain`。

Binding 唯一键：`fmt.Sprintf("%s/%s|%s|%s|%s|%s|%s|%s", ns, name, certID, certRole, certScope, protocol, port, domain)`

### 过期天数计算

```go
daysUntilExpiry = float64(certEndTimeUnix - now.Unix()) / 86400.0
```

已过期为负数；使用 `Asia/Shanghai` 解析 `CertEndTime` 格式 `2006-01-02 15:04:05`。

### query_success 与 days_until_expiry 联动

| 场景 | days_until_expiry | query_success |
|------|-------------------|---------------|
| 查询成功 + 有效时间 | Set 值 | 1 |
| API 失败 / ID 缺失 / 时间无效 | DeleteLabelValues | 0 |
| Binding 消失（Ingress 删除/配置变更） | DeleteLabelValues | DeleteLabelValues |

### main.go 注册片段（设计）

```go
if opts.Cloud == constant.CloudProviderTencent {
    certChecker := check.NewCertificateChecker(mgr.GetClient(), sslClient, namespacedSSL, opts)
    checkRunner.Register(certChecker, check.CheckPer10Min)
}
```

## Complexity Tracking

> 无 Constitution 违规需额外论证。`namespacedssl` 与 `namespacedlb` 存在结构相似性，但通过独立包避免 cloud 子包交叉依赖，符合 ADR-0001 模式复用而非合并。

## Test Strategy

| 层级 | 范围 | 关键用例 |
|------|------|---------|
| 单元 | binding 展开 | HTTPS/MUTUAL/UNIDIRECTIONAL/SNI/三处 scope/非 SSL |
| 单元 | sslclient | 分页、重试、GMT+8 解析、CA 回退 |
| 单元 | certificatechecker | 指标写入、清理、List 失败终止 |
| 单元 | namespacedssl | 豁免 NS、per-NS Secret、凭证缺失隔离 |
| 集成 | main 注册 | 仅 tencentcloud 注册（build tag 或 opts 断言） |

验证命令：

```bash
cd .. && make test-ingress-controller
go test -v -run 'TestExpandBindings|TestCertificateChecker|TestDescribeCertificates|TestNamespacedSSL' \
  ./internal/check/... ./internal/cloud/tencentcloud/... ./internal/cloud/namespacedssl/...
```
