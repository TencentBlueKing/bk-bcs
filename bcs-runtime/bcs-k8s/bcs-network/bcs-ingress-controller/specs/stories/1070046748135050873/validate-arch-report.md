# Validate — Architecture（全量复核）

**commit**: `62759c390`  
**基线**: `2f94d1495`  
**范围**: F-001~F-011（五轮澄清合并）  
**verdict**: LGTM

## 功能需求对照

| FR | 描述 | 实现位置 | 结论 |
|----|------|---------|------|
| FR-001 | Ingress Binding 展开 | `internal/check/binding.go` | ✅ HTTPS/TCP_SSL/QUIC；rule/route/port_mapping；server + MUTUAL client_ca |
| FR-002 | DescribeCertificate 逐 ID 查询 | `internal/cloud/tencentcloud/sslclient.go` | ✅ 分页 1000、重试 3 次、CA 回退；不升级 SDK |
| FR-003 | Prometheus 指标 | `internal/metrics/certificate.go` | ✅ 3 个 GaugeVec；8 label；`init()` 注册 |
| FR-004 | series 清理 | `certificatechecker.go` `lastBindingSet` | ✅ DeleteLabelValues 双指标 |
| FR-005 | NS Scope 凭证 | `internal/cloud/namespacedssl/` | ✅ 镜像 namespacedlb；豁免 NS 走 defaultClient |
| FR-006 | Checker 注册 + 开关 | `main.go` + `ShouldRegisterCertificateChecker` | ✅ 腾讯云 + `certificate_check_enabled=true`；`CheckPer60Min` |
| FR-007 | Helm 开关 | `docs/features/.../values.yaml` + `deployment.yaml` | ✅ 默认 false；true 时追加 args |
| FR-008 | API/Lib 指标 | `sslclient.go` `ReportLibRequestMetric` | ✅ method=`DescribeCertificate` |
| FR-009 | 共享限流 | `sharedratelimit.go` + `sdk.go`/`api.go` 改造 | ✅ `GetSharedRateLimiter()` 进程级单例 |
| FR-013 (F-010) | SSL 域名可配置 | `resolveSSLEndpoint()` + Helm env | ✅ 默认 `ssl.tencentcloudapi.com`；全局 env |
| FR-014 (F-011) | 60 分钟周期 | `checkrunner.go` `CheckPer60Min` | ✅ `time.Hour` ticker；其它 Checker 仍 1 分钟 |

## 架构约束

| 约束 | 结论 |
|------|------|
| ARCH-001 分层（check → cloud，cloud 不反向依赖 check） | ✅ |
| Checker 在 `main.go` 注册 | ✅ |
| Annotation/常量不硬编码（SSL 域名走 env） | ✅ |
| 日志使用 `blog` | ✅ |
| 不修改 CRD / 不回写 CR | ✅ |
| NS Scope SSL 域名为全局配置 | ✅ namespacedssl 无独立域名逻辑 |

## 测试执行

```bash
go test ./internal/check/... ./internal/cloud/tencentcloud/... \
        ./internal/cloud/namespacedssl/... ./internal/metrics/... ./internal/option/...
```

**结果**: 全绿（2026-06-11 复核）

## 备注

- `CertificateCheckerRegisterInterval` 常量暴露注册周期契约，便于单测断言（AC-018）。
- `CheckPer10Min` enum 保留，CertificateChecker 已迁移至 `CheckPer60Min`，符合 F-011「不破坏序号」要求。
