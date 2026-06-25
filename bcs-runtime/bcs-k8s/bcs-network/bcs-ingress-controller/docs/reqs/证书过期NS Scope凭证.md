# 证书过期 NS Scope 凭证

## 基本信息

| 字段 | 值 |
|------|-----|
| 需求 ID | 1070046748135054806 |
| 需求名称 | 证书过期 NS Scope 凭证 |
| 父需求 | [bcs-ingress-controller] 新增证书过期时间指标 |
| 父需求ID | 短ID: 1070046748135050873 / 长ID: 1070046748135050873 |
| 父需求文档 | docs/reqs/证书过期指标.md |
| 优先级 | Middle |
| 价值规模 | 67（Reach=20, Impact=5, Confidence=100%, Effort=1.5人天） |
| 预估工时 | 12 人时 |
| 处理人 | adelaidahe |
| 创建时间 | 2026-06-09 |
| 原始需求文档 | docs/reqs/证书过期NS Scope凭证.md |

## 依赖关系

| 依赖类型 | 依赖需求 | 需求ID | 说明 |
|---------|---------|------|------|
| 强依赖 | 证书过期 Checker 核心实现 | 长ID：1070046748135054749 | 依赖 Binding 展开、DescribeCertificate 轮询客户端、限流与 Prometheus 指标写入及 CertificateChecker 框架；本子需求仅扩展凭证选择逻辑 |

## 需求背景

### 业务背景

在联邦/多租户 BCS 集群中，启用 `--is_namespace_scope=true` 后，各 Namespace 使用独立的腾讯云凭证同步 Listener。若证书过期查询仍使用 Controller 全局凭证，则无法查询到各 NS 账号下的证书失效时间，监控结果与 Listener 同步行为不一致。

本子需求在已交付的 `CertificateChecker` 核心链路上，扩展 `DescribeCertificate` 轮询的**凭证路由**：按 Ingress 所在 Namespace 选择 per-NS Secret / ControllerConfig 凭证；豁免 NS 使用全局凭证。底层仍为逐 ID 调用 `DescribeCertificate`（非 `DescribeCertificates` 批量），限流策略与父需求 F-009 一致。设计依据与 Listener 同步模式一致，参见 [ADR-0001 Namespace Scope Exemption](docs/adr/0001-namespace-scope-exemption.md)。

### 用户故事

作为 **多租户 Namespace Scope 集群管理员**  
我想要 **各 Namespace 使用各自云凭证查询证书过期时间（豁免 NS 使用全局凭证）**  
以便于 **联邦/多租户场景下监控结果与 Listener 同步行为一致**

## 功能需求

### 核心功能点

| 功能编号 | 功能描述 | 优先级 | 涉及角色 | 备注 |
|---------|---------|--------|---------|------|
| F-005 | Namespace Scope 模式下按 NS 凭证分组轮询 DescribeCertificate | P0 | 多租户管理员 | 必须 |

### 详细功能描述

#### [F-005] Namespace Scope 多租户凭证

- **输入**：Ingress 所在 Namespace、去重后的 CertificateId 列表（由核心 Checker 产出）
- **处理逻辑**：
  1. 当 `--is_namespace_scope=true` 时，采用与 Listener 同步相同的凭证选择模式
  2. 普通 NS：使用该 NS 的 per-namespace Secret / ControllerConfig 凭证
  3. 豁免 NS（`--namespace_scope_exempt_namespaces`）：使用 Controller 全局凭证
  4. 按凭证分组逐 ID 调用 `DescribeCertificate`（同一凭证下的 ID 在同一批次内轮询，遵守限流与每页最多 1000 ID 的应用层分组）
  5. 将各凭证组的查询结果合并为 CertificateId → 过期时间映射，供指标写入使用
- **输出**：各 NS Ingress 对应的正确过期天数与 `query_success` 状态
- **边界条件**：
  - `--is_namespace_scope=false` 时行为与子需求「证书过期 Checker 核心实现」一致（全局凭证），无额外逻辑
  - 同一 CertificateId 被多个 NS 引用且使用不同凭证时，按各自 Binding 所属 NS 的凭证分别查询
- **异常处理**：
  - 某 NS 凭证缺失或权限不足 → 该 NS 下 Binding 置 `query_success=0`，删除对应 `days_until_expiry` series，记录 ERROR 日志，**不影响其它 NS**

## 非功能需求

### 性能需求

- **凭证分组**：单次 Check 内按凭证分组串行或适度并发调用 API，不阻塞 Reconcile 主流程
- **调用量**：每个凭证组内逐 ID 轮询，受父需求 F-009 限流约束；应用层每页最多 1000 ID

### 安全需求

- **权限控制**：各 NS 凭证须独立具备 `ssl:DescribeCertificates` 只读权限，覆盖该 NS Ingress 引用的证书 ID
- **数据保护**：指标 label 仍不包含 Secret 或 AccessKey

## 验收标准

### 功能验收

- [ ] **AC-006**：Given `--is_namespace_scope=true` 且 NS-A 使用独立 Secret 凭证 When NS-A 的 Ingress 引用 NS-A 账号下证书 Then 使用 NS-A 凭证查询且 `query_success=1`，`days_until_expiry` 与 CertEndTime 一致
- [ ] **AC-007**：Given NS 在 `--namespace_scope_exempt_namespaces` 白名单 When 该 NS Ingress 引用证书 Then 使用 Controller 全局凭证查询且 `query_success=1`
- [ ] **AC-008-NS**：Given NS-B 凭证缺失或 SSL API 权限不足 When Check 执行 Then 仅 NS-B 下 Binding `query_success=0`，NS-A 及其它 NS 指标正常

### 性能验收

- [ ] **AC-P02**：Given 10 个 NS 各 50 个 Ingress、凭证互不重叠 When Check 执行 Then 按 10 个凭证组完成查询，总耗时不超过核心 Checker 性能基线（AC-P01）的 2 倍

## 边界范围

### 本子需求包含

- `--is_namespace_scope=true` 时按 Ingress NS 选择云凭证
- 豁免 NS 白名单（`--namespace_scope_exempt_namespaces`）使用全局凭证
- 按凭证分组 DescribeCertificate 轮询及结果合并
- 单 NS 凭证失败时的隔离容错（`query_success=0`）

### 本子需求不包含

- Binding 展开逻辑（已在核心子需求实现）
- Prometheus 指标定义与 CertificateChecker 注册（已在核心子需求实现）
- 非 Namespace Scope 模式下的全局凭证路径（已在核心子需求实现）
- 新增 CLI 开关或修改 NS Scope 豁免白名单机制本身

## 人力与工时

* 全量工作1位高级工程师完成工时预估：12 人时
* 全量工作1位中级工程师完成工时预估：16 人时

## RICE 评分明细

| 参数 | 值 | 说明 |
|------|-----|------|
| Reach | 20 | 仅影响启用 Namespace Scope 的多租户/联邦集群管理员 |
| Impact | 5 | 中优先级，保障多租户场景监控与 Listener 同步行为一致 |
| Confidence | 100% | 需求明确，可复用 Listener 凭证选择模式 |
| Effort | 1.5 人天 | 预估 12 人时 ÷ 8 |
| **RICE Score** | **67** | 正常排期，依赖核心子需求完成后实施 |
