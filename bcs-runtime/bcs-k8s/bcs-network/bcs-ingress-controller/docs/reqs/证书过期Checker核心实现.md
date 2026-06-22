# 证书过期 Checker 核心实现

## 基本信息

| 字段 | 值 |
|------|-----|
| 需求 ID | 1070046748135054749 |
| 需求名称 | 证书过期 Checker 核心实现 |
| 父需求 | [bcs-ingress-controller] 新增证书过期时间指标 |
| 父需求ID | 短ID: 1070046748135050873 / 长ID: 1070046748135050873 |
| 父需求文档 | docs/reqs/证书过期指标.md |
| 优先级 | Middle |
| 价值规模 | 83（Reach=50, Impact=5, Confidence=100%, Effort=3人天） |
| 预估工时 | 24 人时 |
| 处理人 | adelaidahe |
| 创建时间 | 2026-06-09 |
| 原始需求文档 | docs/reqs/证书过期Checker核心实现.md |

## 依赖关系

无依赖，可独立开发

## 需求背景

### 业务背景

BCS IngressController 同步 CLB 时仅传递证书 ID，不从云端读取、也不向 CR 写入过期时间。现有 `ListenerChecker` 等周期性检查与 Prometheus 指标均未覆盖证书过期检查。证书过期若未被及时发现，可能导致 HTTPS/TCP_SSL/QUIC 业务中断。

本子需求实现证书过期监控的**核心链路**：从 Ingress 展开证书 Binding、对每个 CertificateId 轮询调用腾讯云 `DescribeCertificate`（单数）查询失效时间、上报 Prometheus 指标，并以独立 `CertificateChecker` 每 10 分钟周期性执行。首期在非 Namespace Scope 模式（或等效全局凭证场景）下交付可验证的完整监控能力。

**API 实现说明**：本期**不**使用 `DescribeCertificates`（复数）批量接口；保持 `tencentcloud-sdk-go v1.0.132` 不升级，底层在 `sslclient.go` 中逐 ID 调用 `DescribeCertificate`，详见父需求 F-002 与第 3 轮澄清。

### 用户故事

作为 **BCS 集群运维人员**  
我想要 **在 Prometheus / 蓝鲸监控中查看各 Ingress 关联 SSL 证书的剩余过期天数**  
以便于 **在证书过期前收到告警并定位到具体 Ingress，避免业务 HTTPS 中断**

作为 **平台 SRE**  
我想要 **区分「证书即将过期」与「过期时间查询失败」两种状态**  
以便于 **避免 API 故障时产生误报，同时不遗漏真实过期风险**

## 功能需求

### 核心功能点

| 功能编号 | 功能描述 | 优先级 | 涉及角色 | 备注 |
|---------|---------|--------|---------|------|
| F-001 | 从全集群 Ingress 展开 SSL 证书 Binding 并去重 | P0 | 运维 | 必须 |
| F-002 | 轮询腾讯云 DescribeCertificate 获取过期时间 | P0 | 运维 | 必须；逐 ID；见父需求 F-002 |
| F-009 | DescribeCertificate 接入现有限流 | P0 | SRE | 必须；见父需求 F-009 |
| F-003 | 上报 `days_until_expiry` / `query_success` 等 Prometheus 指标 | P0 | 运维/SRE | 必须 |
| F-004 | 按 Binding 生命周期清理过期 series | P0 | SRE | 必须 |
| F-006 | CertificateChecker 集成及 API/Lib 调用指标扩展 | P0/P1 | SRE | Checker 必须；API 指标应该有 |

### 详细功能描述

#### [F-001] Ingress 证书 Binding 展开

- **输入**：集群内全量 Ingress CR List
- **处理逻辑**：
  1. 仅处理 SSL 协议：HTTPS、TCP_SSL、QUIC（纯 TCP/UDP 四层规则不展开）
  2. 从以下三处独立展开证书配置块：
     - `spec.rules[].certificate` → `cert_scope=rule`
     - `spec.rules[].layer7Routes[].certificate` → `cert_scope=route`
     - `spec.portMappings[].certificate` → `cert_scope=port_mapping`
  3. 每个配置块按规则产生 Binding：
     - `certID` 非空 → 1 条 `cert_role=server`
     - `mode=MUTUAL` 且 `certCaID` 非空 → 额外 1 条 `cert_role=client_ca`
  4. 对 `cert_id` 去重，供 API 批量查询
- **输出**：Binding 列表（含 owner_namespace、owner_name、cert_id、cert_role、cert_scope、protocol、port、domain 等维度）
- **边界条件**：
  - UNIDIRECTIONAL 模式不产生 client_ca Binding
  - SNI 场景 `domain` 取 route 域名，其它为空字符串 `""`
- **异常处理**：
  - List Ingress 失败 → 本轮 Check 终止，记录 ERROR 日志

#### [F-002] 腾讯云 DescribeCertificate 逐 ID 查询

- **输入**：去重后的 CertificateId 列表
- **处理逻辑**：
  1. 对每个 CertificateId 调用 `DescribeCertificate`（单数；`ssl.tencentcloudapi.com`，`2019-12-05`）
  2. **不**调用 `DescribeCertificates` 批量接口；原因与 SDK 策略见父需求 [F-002] 实现说明表
  3. 每次 `DescribeCertificate` 前执行限流（F-009）
  4. 解析响应 `CertEndTime`（GMT+8）；CA 类型回退 `CAEndTimes`
  5. API 失败时重试 3 次（按页/批次）
  6. **本子需求使用 Controller 全局云凭证**（`--is_namespace_scope=false`；NS Scope 由子需求 #2 扩展）
- **SDK 依赖**：`github.com/tencentcloud/tencentcloud-sdk-go v1.0.132`（**不升级**）
- **输出**：CertificateId → 过期时间的映射
- **边界条件**：
  - 接口无需传递 Region 也可使用
  - 仅能查询当前凭证有权限的证书 ID
- **异常处理**：
  - 重试 3 次仍失败 → 受影响 Binding 置 `query_success=0`，删除 `days_until_expiry` 对应 series，记录 ERROR 日志

#### [F-003] Prometheus 指标上报

- **输入**：Binding 列表 + CertificateId 过期时间映射
- **处理逻辑**：
  1. 对每个 Binding 计算 `days_until_expiry = (CertEndTime - now) / 86400`（浮点，已过期为负数）
  2. 成功解析到有效失效时间 → 设置 `days_until_expiry` 且 `query_success=1`
  3. 查询失败 / ID 缺失 / 失效时间无效 → `query_success=0`，DeleteLabelValues 删除 `days_until_expiry` 对应 series
  4. 更新 `bindings_total` 为当前 Binding 总数
  5. 每次 `DescribeCertificate` 计入 `bkbcs_ingressctrl_lib_*`（`method=DescribeCertificate`）
- **输出**：Prometheus scrape 可采集的 Gauge/Counter/Histogram
- **Label 设计**（`days_until_expiry` 与 `query_success` 共用）：

| Label | 含义 | 示例 |
|-------|------|------|
| owner_namespace | 使用该证书的 Ingress 所在 NS | default |
| owner_name | Ingress 名称 | xxx-ingress |
| cert_id | 腾讯云证书 ID | server 用 certID；client_ca 用 certCaID |
| cert_role | server / client_ca | server |
| cert_scope | rule / route / port_mapping | rule |
| protocol | 规则或 mapping 协议 | HTTPS |
| port | 规则端口或 mapping 起始端口（字符串） | 443 |
| domain | SNI 域名，其它为 `""` | example.qq.com |

- **注意**：指标**不携带** `bcs_cluster_id` label，该 label 由蓝鲸监控平台自动添加
- **异常处理**：指标写入失败不影响 Controller 主流程

#### [F-004] 指标刷新与清理

- **输入**：本轮 Check 产生的 Binding 集合
- **处理逻辑**：
  1. 每轮 Check 全量 List Ingress 并重建 Binding 集合
  2. 对上轮存在、本轮不存在的 Binding：DeleteLabelValues（同时清理 `days_until_expiry` 与 `query_success`）
- **输出**：与当前 Ingress 状态一致的指标 series

#### [F-006] CertificateChecker 集成

- **输入**：Controller 启动完成，云厂商为腾讯云
- **处理逻辑**：
  1. 新增独立 `CertificateChecker`，注册到 `CheckRunner` 及 `main.go`
  2. 执行周期：**每 10 分钟**（`CheckPer10Min`）
  3. 腾讯云环境下**默认开启，无 CLI 开关**
  4. 非腾讯云云厂商：**不注册** Checker
- **输出**：周期性更新的证书过期指标

## 非功能需求

### 性能需求

- **检查周期**：默认每 10 分钟执行一次全量 Check
- **API 调用量**：每唯一 certID 一次 `DescribeCertificate`；应用层每页最多 1000 ID 分组
- **限流**：复用 `TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`
- **并发能力**：Checker 在独立 goroutine 执行，不阻塞 Controller Reconcile 主流程

### 安全需求

- **权限控制**：Controller 全局凭证需 `ssl:DescribeCertificates` 只读权限，与 CLB 权限独立，部署前须单独开通
- **数据保护**：指标中不包含 Secret 内容，仅暴露证书 ID 与过期天数

## 验收标准

### 功能验收

- [ ] **AC-001**：Given 集群中存在 HTTPS Ingress 且 certID 有效、SSL API 权限正常、`--is_namespace_scope=false` When CertificateChecker 执行一轮 Check Then `bkbcs_ingressctrl_certificate_days_until_expiry` 出现对应 series 且 `query_success=1`，天数与 CertEndTime 一致（误差 < 1 天）
- [ ] **AC-002**：Given Ingress 为 MUTUAL 模式且 certID、certCaID 均配置 When Check 执行 Then 分别产生 `cert_role=server` 与 `cert_role=client_ca` 两条 series
- [ ] **AC-003**：Given Ingress 从三处（rule / route / port_mapping）配置证书 When Check 执行 Then `cert_scope` label 分别正确为 rule / route / port_mapping
- [ ] **AC-004**：Given 证书已过期 When Check 执行 Then `days_until_expiry` 为负数
- [ ] **AC-005**：Given Ingress 被删除 When 下一轮 Check 执行 Then 该 Ingress 相关 series 被 DeleteLabelValues 清理
- [ ] **AC-008**：Given SSL API 权限未开通或连续 3 次调用失败 When Check 执行 Then 受影响 Binding `query_success=0` 且无 `days_until_expiry` series，Controller 其它功能正常
- [ ] **AC-009**：Given Controller 部署云厂商非腾讯云 When 启动 Then 不注册 CertificateChecker，无 certificate 子系统指标产出
- [ ] **AC-010**：Given 腾讯云环境 When Controller 启动 Then CertificateChecker 默认注册且每 10 分钟执行，无额外 CLI 开关

### 性能验收

- [ ] **AC-P01**：Given 集群 500 个 Ingress、200 个唯一 certID When Check 执行 Then 单次 Check 在 10 分钟内完成（含限流等待）

### 安全验收

- [ ] **AC-S01**：Given Prometheus 指标暴露 When 检查 series label Then 不包含 Secret、AccessKey 等敏感信息

## 边界范围

### 本子需求包含

- Ingress 证书 Binding 展开逻辑（rule / route / port_mapping；server / client_ca）
- 腾讯云 `DescribeCertificate` 逐 ID 轮询（全局凭证；不升级 SDK）
- DescribeCertificate 限流（与 CLB 共用令牌桶配置）
- certificate 子系统 Prometheus 指标（`days_until_expiry`、`query_success`、`bindings_total`）
- 指标 series 刷新与 DeleteLabelValues 清理
- CertificateChecker 注册与 10 分钟周期执行（仅 tencentcloud）
- API/Lib 指标（`method=DescribeCertificate`）

### 本子需求不包含

- Namespace Scope 模式下按 NS 选择云凭证（见子需求「证书过期 NS Scope 凭证」）
- 豁免 NS 白名单全局凭证逻辑
- AWS / GCP / Azure 证书过期监控
- 自动续费或自动更新 Listener certID
- 向 Ingress / Listener CR 回写过期时间字段
- CLI 开关关闭证书过期检查
- 蓝鲸监控推荐告警阈值

## 人力与工时

* 全量工作1位高级工程师完成工时预估：24 人时
* 全量工作1位中级工程师完成工时预估：32 人时

## RICE 评分明细

| 参数 | 值 | 说明 |
|------|-----|------|
| Reach | 50 | 覆盖全部腾讯云 IngressController 部署集群的运维/SRE |
| Impact | 5 | 中优先级，证书过期监控对 HTTPS 业务连续性有重要价值 |
| Confidence | 100% | 需求文档完善，技术方案明确，可复用现有 Checker 模式 |
| Effort | 3 人天 | 预估 24 人时 ÷ 8 |
| **RICE Score** | **83** | 正常排期，按迭代计划推进 |
