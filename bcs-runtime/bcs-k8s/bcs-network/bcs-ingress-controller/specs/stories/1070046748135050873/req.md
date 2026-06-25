# [bcs-ingress-controller] 新增证书过期时间指标

## 基本信息

| 字段 | 值 |
|------|-----|
| 需求 ID | 1070046748135050873 |
| 需求名称 | [bcs-ingress-controller] 新增证书过期时间指标 |
| 优先级 | Middle |
| 处理人 | adelaidahe |
| 父需求 | 无 |
| 子需求 | 1070046748135054749（Checker 核心实现）、1070046748135054806（NS Scope 凭证） |
| 创建时间 | 2026-06-09 11:35:21 |
| 原始需求文档 | docs/reqs/证书过期指标.md |
| 本地子需求文档 | docs/reqs/证书过期Checker核心实现.md、docs/reqs/证书过期NS Scope凭证.md |

## 需求背景

### 业务背景

BCS IngressController 将 HTTPS 证书配置写入 Ingress（及生成的 Listener CR），结构体 `IngressListenerCertificate` 保存 TLS 模式、服务端证书 ID（`certID`）、双向认证 CA 证书 ID（`certCaID`）等字段。Controller 同步 CLB 时仅传递证书 ID，**不从云端读取、也不向 CR 写入过期时间**。

现有周期性检查（`ListenerChecker` 等）与 Prometheus 指标均**未覆盖**证书过期检查。证书过期若未被及时发现，可能导致 HTTPS/TCP_SSL/QUIC 业务中断。本需求在 Controller 侧提供证书剩余过期天数指标，供用户在蓝鲸监控平台配置告警策略，作为腾讯云账号级消息、云监控告警的补充手段。

**目标用户**：BCS 集群运维人员、平台 SRE。

**不做本需求的影响**：运维无法通过 IngressController 暴露的 Prometheus 指标按 Ingress 维度发现即将过期的证书，只能依赖腾讯云侧兜底告警，定位到具体 Ingress 的成本更高。

**开关默认关闭的原因**：证书过期检测依赖腾讯云账号具备 `ssl:DescribeCertificates` 只读权限，而许多现存账号未开通该权限。默认关闭可避免未授权账号在升级后触发无效 API 调用或启动失败，由运维在确认权限就绪后显式开启。

### 用户故事

作为 **BCS 集群运维人员**  
我想要 **在 Prometheus / 蓝鲸监控中查看各 Ingress 关联 SSL 证书的剩余过期天数**  
以便于 **在证书过期前收到告警并定位到具体 Ingress，避免业务 HTTPS 中断**

作为 **平台 SRE**  
我想要 **区分「证书即将过期」与「过期时间查询失败」两种状态**  
以便于 **避免 API 故障时产生误报，同时不遗漏真实过期风险**

作为 **多租户 Namespace Scope 集群管理员**  
我想要 **各 Namespace 使用各自云凭证查询证书过期时间（豁免 NS 使用全局凭证）**  
以便于 **联邦/多租户场景下监控结果与 Listener 同步行为一致**

作为 **平台 SRE**  
我想要 **通过 CLI 参数或 Helm 配置显式开启证书过期检测**  
以便于 **在未开通 SSL API 权限的集群中保持默认关闭，在权限就绪的集群中按需启用**

作为 **BCS 集群运维人员**  
我想要 **像配置 CLB API 域名一样，指定腾讯云 SSL 证书 API 的请求域名**  
以便于 **在内网/私有化部署场景下通过 `ssl.internal.tencentcloudapi.com` 等内网 endpoint 正常查询证书过期时间**

作为 **平台 SRE**  
我想要 **证书过期检查以 1 小时为周期执行**  
以便于 **降低 SSL API 调用频率与 Controller 巡检负载，同时仍能在证书过期前通过 Prometheus 指标触发告警**

### 需求来源

- **需求渠道**：技术需求 / 优化需求
- **关联需求**：无
- **参考资料**：
  - [腾讯云 DescribeCertificates API](https://cloud.tencent.com/document/api/400/41671)

## 功能需求

### 核心功能点

| 功能编号 | 功能描述 | 优先级 | 涉及角色 | 备注 |
|---------|---------|--------|---------|------|
| F-001 | 从全集群 Ingress 展开 SSL 证书 Binding 并去重 | P0 | 运维 | 必须 |
| F-002 | 轮询腾讯云 DescribeCertificate 获取过期时间 | P0 | 运维 | 必须；见技术决策 |
| F-009 | DescribeCertificate 调用接入现有限流机制 | P0 | SRE | 必须 |
| F-003 | 上报 `days_until_expiry` / `query_success` 等 Prometheus 指标 | P0 | 运维/SRE | 必须 |
| F-004 | 按 Binding 生命周期清理过期 series | P0 | SRE | 必须 |
| F-005 | Namespace Scope 模式下按 NS 凭证查询（同 Listener 模式） | P0 | 多租户管理员 | 必须 |
| F-006 | CertificateChecker 集成与 CLI 开关控制 | P0 | SRE | 必须 |
| F-007 | Helm Chart 暴露 `certificateCheckEnabled` 配置 | P0 | 运维 | 必须 |
| F-008 | 复用现有 API/Lib 指标记录 DescribeCertificate 调用 | P1 | SRE | 应该有 |
| F-010 | SSL API 请求域名可配置（对齐 CLB 域名配置模式） | P0 | 运维 | 必须；增量修订 |
| F-011 | 将 CertificateChecker 执行周期从 10 分钟调整为 60 分钟 | P0 | SRE | 必须；第 5 轮增量修订 |

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
  4. 对 `cert_id` 去重，供 API 逐 ID 查询
- **输出**：Binding 列表（含 owner_namespace、owner_name、cert_id、cert_role、cert_scope、protocol、port、domain 等维度）
- **边界条件**：
  - UNIDIRECTIONAL 模式不产生 client_ca Binding
  - SNI 场景 `domain` 取 route 域名，其它为空字符串 `""`
- **异常处理**：
  - List Ingress 失败 → 本轮 Check 终止，记录 ERROR 日志

#### [F-002] 腾讯云 DescribeCertificate 逐 ID 查询（轮询）

- **输入**：去重后的 CertificateId 列表
- **处理逻辑**：
  1. 对每个 CertificateId **单独调用** SSL 证书服务 `DescribeCertificate`（单数，endpoint 由 `TENCENTCLOUD_SSL_DOMAIN` 指定，默认 `ssl.tencentcloudapi.com`，版本 `2019-12-05`）
  2. **不调用** `DescribeCertificates`（复数）批量接口；`SSLClient.DescribeCertificates` 方法名保留为对上层暴露的批量语义封装，底层实现为逐 ID 轮询
  3. 每次 `DescribeCertificate` 调用前须执行限流（见 F-009）
  4. 解析响应 `CertEndTime`（GMT+8）为 Unix 时间戳
  5. 若 `CertificateType=CA` 且 `CertEndTime` 为空，参考 `CAEndTimes`（多值取最早到期时间）
  6. 仍无有效失效时间 → 标记为无过期时间，记录 INFO 日志
  7. 单页（一组 certID）API 失败时重试 3 次；分页粒度为每页最多 1000 个 ID（与现实现一致），页内仍逐 ID 调用
- **输出**：CertificateId → 过期时间的映射
- **实现说明（须在技术文档中明确记录）**：

| 项 | 说明 |
|----|------|
| 实际云 API | `DescribeCertificate`（单数），每个证书 ID 一次请求 |
| 为何不用 `DescribeCertificates` 批量 | 项目依赖 `tencentcloud-sdk-go v1.0.132` 单体包，其 `DescribeCertificatesRequest` **无 `CertIds` 字段**；升级至支持 `CertIds` 的子模块版本（≥ v1.0.1090）须移除单体包并联动升级 CLB/VPC/CVM 等子模块，存在 **omitnil 序列化行为变化** 等稳定性风险，与「不破坏现网 CLB/ENI 行为」目标冲突 |
| SDK 策略 | **本期不升级** Tencent Cloud Go SDK，继续使用 `go.mod` 中 `github.com/tencentcloud/tencentcloud-sdk-go v1.0.132` |
| IAM 权限 | 仍须 `ssl:DescribeCertificates` 只读权限（与腾讯云 SSL 证书读接口授权一致） |
| 代码位置 | `internal/cloud/tencentcloud/sslclient.go` — `doDescribeCertificates` 循环调用 `DescribeCertificate` |

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
  5. 每次 `DescribeCertificate` 云 API 调用计入已有 `bkbcs_ingressctrl_api_*` / `bkbcs_ingressctrl_lib_*` 指标（`method=DescribeCertificate`）
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

- **注意**：指标**不携带** `bcs_cluster_id` label，该 label 由蓝鲸监控平台自动添加。
- **异常处理**：指标写入失败不影响 Controller 主流程

#### [F-004] 指标刷新与清理

- **输入**：本轮 Check 产生的 Binding 集合
- **处理逻辑**：
  1. 每轮 Check 全量 List Ingress 并重建 Binding 集合
  2. 对上轮存在、本轮不存在的 Binding：DeleteLabelValues（同时清理 `days_until_expiry` 与 `query_success`）
- **输出**：与当前 Ingress 状态一致的指标 series

#### [F-005] Namespace Scope 多租户凭证

- **输入**：Ingress 所在 Namespace
- **处理逻辑**：
  1. 采用与 Listener 同步相同的凭证选择模式（参见 ADR-0001）
  2. 普通 NS：使用该 NS 的 per-namespace Secret / ControllerConfig 凭证
  3. 豁免 NS（`--namespace_scope_exempt_namespaces`）：使用 Controller 全局凭证
  4. 按凭证分组逐 ID 调用 `DescribeCertificate`（同一凭证下的 CertificateId 合并为一次 Checker 轮次内的查询批次，底层仍逐 ID 轮询）
- **输出**：各 NS Ingress 对应的正确过期天数
- **异常处理**：
  - 某 NS 凭证缺失或权限不足 → 该 NS 下 Binding 置 `query_success=0`，不影响其它 NS

#### [F-006] CertificateChecker 集成与 CLI 开关

- **输入**：Controller 启动完成
- **处理逻辑**：
  1. 新增独立 `CertificateChecker`，注册到 `CheckRunner` 及 `main.go`
  2. 执行周期：**每 60 分钟（1 小时）**
  3. 注册条件须**同时满足**：
     - 云厂商为腾讯云（`--cloud=tencentcloud`）
     - `--certificate_check_enabled=true`（**默认 `false`**）
  4. 未满足注册条件时：**不初始化 SSL Client、不注册 CertificateChecker**
  5. 关闭开关后不主动清理已有 Prometheus series；修改启动参数后滚动重启 Pod，旧 series 随 Pod 销毁自动消失
  6. 非腾讯云云厂商：**不注册** Checker（与开关无关）
- **输出**：周期性更新的证书过期指标（仅开启时）

#### [F-007] Helm Chart 配置

- **输入**：Helm values 中 `certificateCheckEnabled` 字段
- **处理逻辑**：
  1. 在 `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/values.yaml` 新增 `certificateCheckEnabled`，默认 `false`
  2. 在 `templates/deployment.yaml` 中，当 `certificateCheckEnabled: true` 时向容器 args 追加 `--certificate_check_enabled=true`
  3. 当 `certificateCheckEnabled: false`（默认）时不追加该参数（沿用 CLI 默认值 `false`）
- **输出**：运维可通过 Helm values 控制证书过期检测开关

#### [F-008] API/Lib 指标复用

- **输入**：`DescribeCertificate` 云 API 调用
- **处理逻辑**：每次实际 SDK 调用计入已有 `bkbcs_ingressctrl_api_*` / `bkbcs_ingressctrl_lib_*` 指标（`method=DescribeCertificate`）
- **输出**：可观测的 API 调用量与延迟

#### [F-009] DescribeCertificate 限流

- **输入**：`sslclient` 每次发起 `DescribeCertificate` 前
- **处理逻辑**：
  1. **复用** CLB 路径已有的腾讯云 API 令牌桶限流（`SdkWrapper.tryThrottle` / `throttle.NewTokenBucket`）
  2. 限流参数与 CLB 共用环境变量：`TENCENTCLOUD_RATELIMIT_QPS`（默认 50）、`TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`（默认与 CLB 一致）
  3. 在 `doDescribeCertificates` 循环内、每次 `c.api.DescribeCertificate` **之前**调用限流等待
  4. 限流等待不阻塞 Reconcile 主流程（Checker 在独立 goroutine）；可拉长单次 Check 耗时，属预期行为
- **输出**：DescribeCertificate 请求速率受控，避免 Checker 每 60 分钟全量轮询时瞬时打满 SSL API QPS
- **异常处理**：触发腾讯云 `RequestLimitExceeded` 时沿用 CLB 既有重试/退避策略（若 SDK 返回可重试错误码）

#### [F-010] SSL API 请求域名可配置

- **输入**：环境变量 `TENCENTCLOUD_SSL_DOMAIN` 或 Helm values `tencentcloudSslDomain`
- **处理逻辑**：
  1. 在 `internal/cloud/tencentcloud/` 新增常量 `EnvNameTencentCloudSslDomain = "TENCENTCLOUD_SSL_DOMAIN"`，与 CLB 的 `EnvNameTencentCloudClbDomain` 命名风格一致
  2. `SSLClient` 创建 SDK Client 时，从环境变量读取 SSL endpoint；未配置或为空时使用默认值 `ssl.tencentcloudapi.com`（与当前硬编码行为一致）
  3. 内网/私有化部署场景示例值：`ssl.internal.tencentcloudapi.com`（对齐 CLB 内网域名 `clb.internal.tencentcloudapi.com` 的命名规则）
  4. **所有** Controller 内腾讯云 SSL API 调用均复用此 endpoint 配置（当前为 `DescribeCertificate`；未来新增 SSL API 亦须走同一配置，不得硬编码域名）
  5. Helm Chart 在 `values.yaml` 新增 `tencentcloudSslDomain`，默认 `ssl.tencentcloudapi.com`
  6. Helm `templates/deployment.yaml` 向容器 env 注入 `TENCENTCLOUD_SSL_DOMAIN`，取值来自 `.Values.tencentcloudSslDomain`（与 `TENCENTCLOUD_CLB_DOMAIN` / `tencentcloudClbDomain` 注入方式一致）
  7. Namespace Scope 模式下 per-NS Secret **不**支持独立 SSL 域名；SSL 域名为 Controller 全局配置（与 CLB 域名配置粒度一致）
- **输出**：运维可通过环境变量或 Helm values 指定 SSL API 请求域名
- **边界条件**：
  - 未设置 `TENCENTCLOUD_SSL_DOMAIN` → 使用 `ssl.tencentcloudapi.com`
  - 证书过期检测开关关闭时，不初始化 SSL Client，域名配置无运行时影响
- **异常处理**：
  - 域名配置错误导致 API 不可达 → 沿用既有逻辑：`query_success=0`、ERROR 日志，不影响 Controller 其它功能

#### [F-011] 证书过期检查周期调整

- **输入**：Controller 启动且满足 CertificateChecker 注册条件（腾讯云 + `--certificate_check_enabled=true`）
- **处理逻辑**：
  1. CertificateChecker 的定时执行周期由 **每 10 分钟** 改为 **每 60 分钟（1 小时）**
  2. 仅调整 CertificateChecker 的调度周期；其它 Checker（1 分钟周期）**不受影响**
  3. 检查周期为固定值 60 分钟，**不新增** CLI / Helm 可配置项
  4. CertificateChecker 的单轮 Check 逻辑**不变**
- **输出**：Prometheus 证书过期指标每 60 分钟刷新一次
- **边界条件**：指标数据最大 staleness 由 10 分钟变为 60 分钟
- **异常处理**：与父需求一致，Check 失败不影响 Controller 其它功能

## 非功能需求

### 性能需求

- **检查周期**：开启后默认每 60 分钟执行一次全量 Check
- **API 调用量**：每个唯一 CertificateId 每轮 Check 产生 1 次 `DescribeCertificate` 请求；N 个唯一 ID → N 次云 API 调用（受 F-009 限流约束）
- **分页**：应用层按每页最多 1000 个 ID 分组处理，页内逐 ID 轮询
- **并发能力**：单次 Check 内按凭证分组串行查询，组内逐 ID 串行（配合限流）；不阻塞 Controller Reconcile 主流程（Checker 在独立 goroutine 执行）

### 安全需求

- **权限控制**：开启功能前，Controller 凭证须具备 `ssl:DescribeCertificates` 只读权限，与 CLB 权限独立，部署前须单独开通
- **数据保护**：指标中不包含 Secret 内容，仅暴露证书 ID 与过期天数
- **合规要求**：不自动续费、不自动更新 Listener 上的 certID

### 可用性与稳定性

- **容错能力**：默认关闭，避免无 SSL 权限账号升级后产生副作用；开启后若 SSL API 调用失败，置 `query_success=0` 并打 ERROR 日志，**不影响 Controller 其它功能**
- **幂等性**：每轮 Check 全量重建指标，与现有 Checker 模式一致
- **向后兼容**：未显式配置 `--certificate_check_enabled` 的集群升级后行为与默认关闭一致；未配置 `TENCENTCLOUD_SSL_DOMAIN` 时 SSL endpoint 仍为 `ssl.tencentcloudapi.com`，与升级前硬编码行为一致

### 兼容性

- **云厂商**：首期仅腾讯云；AWS / GCP / Azure 不在本期范围
- **接口兼容**：新增 CLI 参数与 Helm values，不修改现有 CRD 字段与 Reconcile 逻辑

## 业务规则

### 业务逻辑规则

- **规则 R-001**：仅监控 Ingress CR 中配置的 SSL 证书，不从 CLB API 反查过期时间（CLB 响应不含失效时间）
- **规则 R-002**：告警表达式须同时要求 `query_success == 1`，避免 API 故障误报
- **规则 R-003**：已过期证书 `days_until_expiry` 为负数（如过期 3 天 → `-3`）
- **规则 R-004**：同一 Ingress 多条 Binding（多端口、多域名、MUTUAL）各自独立 series，可通过 `min by (owner_namespace, owner_name)` 聚合找最早到期证书
- **规则 R-005**：证书过期检测默认关闭，须运维在确认 `ssl:DescribeCertificates` 权限就绪后显式开启
- **规则 R-006**：SSL API 请求域名为 Controller 全局配置，通过环境变量注入；per-NS Secret 仅含凭证，不含 SSL 域名
- **规则 R-007**：CertificateChecker 执行周期固定为 60 分钟，不可通过 CLI / Helm 配置
- **规则 R-008**：其它周期性 Checker（1 分钟周期）不受检查周期调整影响

### 数据校验规则

- **必填字段**：存在证书配置块时 `certID` 必填；`mode=MUTUAL` 时 `certCaID` 必填（与现有 Ingress 校验一致）
- **格式要求**：`CertEndTime` 解析格式 `2006-01-02 15:04:05`（GMT+8）

### 权限规则

- 开启功能后，SSL 只读权限须覆盖集群 Ingress 引用的全部证书 ID，否则无权限的 ID 对应 Binding 为 `query_success=0`

## 外部依赖与集成

### 外部系统集成

| 系统名称 | 交互方式 | 接口说明 | 认证方式 | 文档链接 |
|---------|---------|---------|---------|---------|
| 腾讯云 SSL 证书服务 | HTTPS API | `DescribeCertificate` 逐 ID 查询证书失效时间（轮询，非 `DescribeCertificates` 批量）；endpoint 可配置，默认 `ssl.tencentcloudapi.com`，内网示例 `ssl.internal.tencentcloudapi.com` | Controller 云凭证（per-NS 或全局） | [DescribeCertificate](https://cloud.tencent.com/document/api/400/41674)、[DescribeCertificates](https://cloud.tencent.com/document/api/400/41671)（本期不使用批量） |
| 蓝鲸监控平台 | Prometheus scrape | 采集 Controller metrics 端点 | 平台既有机制 | — |

### 接口契约

`DescribeCertificate` 关键响应字段（单证书详情接口）：

| 参数名 | 类型 | 用途 |
|--------|------|------|
| CertEndTime | string | 证书失效时间，GMT+8，示例 `2025-01-03 07:59:59` |
| CAEndTimes | []string | CA 证书到期时间，CertificateType=CA 时有效 |

### 数据模型

证书 Binding（逻辑实体，非 CRD）：

| 字段 | 说明 |
|------|------|
| owner_namespace / owner_name | Ingress 标识 |
| cert_id | 查询用的腾讯云 CertificateId |
| cert_role | server / client_ca |
| cert_scope | rule / route / port_mapping |
| protocol / port / domain | 证书挂载上下文 |

CLI / Helm 配置：

| 层级 | 名称 | 类型 | 默认值 | 说明 |
|------|------|------|--------|------|
| CLI | `--certificate_check_enabled` | bool | `false` | 为 `true` 时注册 CertificateChecker |
| Go | `CertificateCheckEnabled` | bool | `false` | `ControllerOption` 字段 |
| Helm | `certificateCheckEnabled` | bool | `false` | values.yaml 配置项 |
| 环境变量 | `TENCENTCLOUD_SSL_DOMAIN` | string | `ssl.tencentcloudapi.com` | SSL API 请求域名；未设置时等同默认值 |
| Helm | `tencentcloudSslDomain` | string | `ssl.tencentcloudapi.com` | values.yaml 配置项，注入为 `TENCENTCLOUD_SSL_DOMAIN` |

## 验收标准

### 功能验收

- [ ] **AC-001**：Given `--certificate_check_enabled=true` 且集群中存在 HTTPS Ingress、certID 有效、SSL API 权限正常 When CertificateChecker 执行一轮 Check Then `bkbcs_ingressctrl_certificate_days_until_expiry` 出现对应 series 且 `query_success=1`，天数与 CertEndTime 一致（误差 < 1 天）
- [ ] **AC-002**：Given `--certificate_check_enabled=true` 且 Ingress 为 MUTUAL 模式且 certID、certCaID 均配置 When Check 执行 Then 分别产生 `cert_role=server` 与 `cert_role=client_ca` 两条 series
- [ ] **AC-003**：Given `--certificate_check_enabled=true` 且 Ingress 从三处（rule / route / port_mapping）配置证书 When Check 执行 Then `cert_scope` label 分别正确为 rule / route / port_mapping
- [ ] **AC-004**：Given `--certificate_check_enabled=true` 且证书已过期 When Check 执行 Then `days_until_expiry` 为负数
- [ ] **AC-005**：Given `--certificate_check_enabled=true` 且 Ingress 被删除 When 下一轮 Check 执行 Then 该 Ingress 相关 series 被 DeleteLabelValues 清理
- [ ] **AC-006**：Given `--certificate_check_enabled=true` 且 `--is_namespace_scope=true`、NS-A 使用独立 Secret 凭证 When NS-A 的 Ingress 引用 NS-A 账号下证书 Then 使用 NS-A 凭证查询且 `query_success=1`
- [ ] **AC-007**：Given `--certificate_check_enabled=true` 且 NS 在 `--namespace_scope_exempt_namespaces` 白名单 When 该 NS Ingress 引用证书 Then 使用 Controller 全局凭证查询
- [ ] **AC-008**：Given `--certificate_check_enabled=true` 且 SSL API 权限未开通或连续 3 次调用失败 When Check 执行 Then 受影响 Binding `query_success=0` 且无 `days_until_expiry` series，Controller 其它功能正常
- [ ] **AC-009**：Given Controller 部署云厂商非腾讯云 When 启动 Then 不注册 CertificateChecker，无 certificate 子系统指标产出
- [ ] **AC-010**：Given 腾讯云环境且未指定 `--certificate_check_enabled` When Controller 启动 Then 不初始化 SSL Client、不注册 CertificateChecker（默认关闭）
- [ ] **AC-011**：Given 腾讯云环境且 `--certificate_check_enabled=true` When Controller 启动 Then 初始化 SSL Client、注册 CertificateChecker 且每 60 分钟执行
- [ ] **AC-012**：Given 腾讯云环境且 `--certificate_check_enabled=false` When Controller 启动 Then 不初始化 SSL Client、不注册 CertificateChecker
- [ ] **AC-013**：Given Helm `certificateCheckEnabled: true` When 部署 Controller Then Deployment args 包含 `--certificate_check_enabled=true` 且 CertificateChecker 注册
- [ ] **AC-014**：Given Helm `certificateCheckEnabled: false`（默认）When 部署 Controller Then Deployment args 不包含 `--certificate_check_enabled` 且 CertificateChecker 不注册
- [ ] **AC-015**：Given 未设置 `TENCENTCLOUD_SSL_DOMAIN` 且 `--certificate_check_enabled=true` When SSLClient 发起 `DescribeCertificate` Then 请求 endpoint 为 `ssl.tencentcloudapi.com`（与升级前行为一致）
- [ ] **AC-016**：Given `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com` 且 `--certificate_check_enabled=true` When SSLClient 发起 `DescribeCertificate` Then 请求 endpoint 为 `ssl.internal.tencentcloudapi.com`
- [ ] **AC-017**：Given Helm `tencentcloudSslDomain: ssl.internal.tencentcloudapi.com` When 部署 Controller Then Deployment env 包含 `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com`
- [ ] **AC-018**：Given 腾讯云环境且 `--certificate_check_enabled=true` When Controller 启动 Then CertificateChecker 注册成功且每 **60 分钟** 触发一次 `Run()`（非 10 分钟）
- [ ] **AC-019**：Given CertificateChecker 已注册 When 观察 2 小时内 Check 执行次数 Then 执行次数为 2 次（±启动时刻偏差）
- [ ] **AC-020**：Given `--certificate_check_enabled=true` 且集群中存在 HTTPS Ingress、certID 有效、SSL API 权限正常 When CertificateChecker 执行一轮 Check Then 指标行为与 AC-001 一致，仅刷新频率变为 60 分钟
- [ ] **AC-021**：Given Controller 运行中 When 观察 PortBindChecker / ListenerChecker 等其它 Checker Then 仍保持 **1 分钟** 周期，不受本修订影响

### 性能验收

- [ ] **AC-P01**：Given `--certificate_check_enabled=true` 且集群 500 个 Ingress、200 个唯一 certID When Check 执行 Then 单次 Check 在 60 分钟周期内完成（含限流等待；200 次 DescribeCertificate @ 默认 50 QPS 理论下限约 4s，须预留凭证分组与重试余量）
- [ ] **AC-P02**：Given `--certificate_check_enabled=true` 且 200 个唯一 certID When 一轮 Check 执行 Then 每次 `DescribeCertificate` 调用前均经过令牌桶限流，且 `bkbcs_ingressctrl_lib_request_total{method="DescribeCertificate"}` 计数为 200（±重试次数）
- [ ] **AC-P03**：Given `--certificate_check_enabled=true` 且集群 500 个 Ingress、200 个唯一 certID When Check 执行 Then 单次 Check 在 **60 分钟** 周期内完成（含限流等待）

### 安全验收

- [ ] **AC-S01**：Given Prometheus 指标暴露 When 检查 series label Then 不包含 Secret、AccessKey 等敏感信息

## 边界范围

### 本期包含

- 腾讯云 SSL `DescribeCertificate` 逐 ID 轮询查询（不升级 SDK、不使用 `DescribeCertificates` 批量）
- DescribeCertificate 接入现有限流（`TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`）
- CertificateChecker（60 分钟周期）及 certificate 子系统 Prometheus 指标
- Namespace Scope / 豁免 NS 凭证策略（同 Listener）
- Binding 展开、指标刷新与 series 清理
- API/Lib 调用量与延迟指标扩展
- CLI 参数 `--certificate_check_enabled`（默认 `false`）
- Helm Chart `certificateCheckEnabled` 配置及 Deployment args 传递
- SSL API 请求域名可配置（`TENCENTCLOUD_SSL_DOMAIN` / Helm `tencentcloudSslDomain`）

### 本期不包含

- AWS / GCP / Azure 证书过期监控
- 自动续费或自动更新 Listener certID
- 替代腾讯云账号级消息、云监控告警
- 向 Ingress / Listener CR 回写过期时间字段
- 关闭开关时主动清理已有 Prometheus series
- 需求文档内蓝鲸监控推荐告警阈值（由运维自行配置）
- 指标携带 `bcs_cluster_id` label

## 约束条件

- **技术限制**：仅 controller-runtime Prometheus metrics；日志使用 `blog`
- **部署限制**：开启功能前须预先开通 `ssl:DescribeCertificates` 只读权限
- **架构约束**：Checker 注册须在 `main.go` 同时校验云厂商与 `CertificateCheckEnabled`；Helm Chart 路径为 `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/`；SSL endpoint 配置须在 `SSLClient` 统一读取，禁止在业务代码中硬编码 SSL 域名

## 澄清记录

### 第 1 轮澄清（业务）

**提问时间**：2026-06-09

**Agent 提问**：

1. Namespace Scope 模式下 DescribeCertificates 用哪套云凭证？
2. CertificateChecker 执行频率？
3. 指标是否携带 bcs_cluster_id label？
4. 是否支持关闭证书过期检查？
5. 需求文档是否需给出蓝鲸监控推荐告警规则？
6. 非腾讯云集群部署时的行为？
7. SSL API 权限未开通时的期望行为？
8. 需求优先级？

**用户回复**：

1. **A** — 按 Ingress 所在 NS 使用对应云凭证；豁免 NS 使用全局凭证（与 Listener 处理模式一致）
2. **B** — 独立 CertificateChecker，默认每 10 分钟运行一次
3. **不需要** — 不携带 bcs_cluster_id，由蓝鲸监控自动添加
4. **A** — 腾讯云环境下默认开启，无 CLI 开关（*第 2 轮已修订*）
5. **C** — 不需要推荐告警阈值
6. **A** — 仅 cloud=tencentcloud 时注册 Checker
7. **A** — 权限不足时 query_success=0、ERROR 日志，不影响其它功能
8. **B** — Middle 优先级

### 第 2 轮澄清（增量：证书过期检测开关）

**提问时间**：2026-06-10

**Agent 提问**：

1. 默认行为：开关未配置时是否仍默认开启？
2. 关闭后的行为：不注册 Checker / 仅跳过 Run / 主动清理 series？
3. CLI 参数命名偏好？
4. Helm Chart 更新范围？

**用户回复**：

1. **默认关闭** — 许多现存腾讯云账号无 `ssl:DescribeCertificates` 权限，必须默认关闭以避免升级副作用
2. **A** — 不初始化 SSL Client、不注册 CertificateChecker；不主动清理已有 Prometheus series（滚动重启后旧 series 随 Pod 销毁自动消失）
3. **确认** — CLI：`--certificate_check_enabled`（默认 `false`）；Go：`CertificateCheckEnabled`；Helm：`certificateCheckEnabled`
4. **确认** — 更新 `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/` 的 `values.yaml` 与 `deployment.yaml`

### 第 3 轮澄清（增量：API 实现与限流）

**提问时间**：2026-06-10

**用户诉求（已确认，无需额外追问）**：

1. **继续现有实现** — 使用 `DescribeCertificate`（单数）逐 ID 轮询，**不**改为 `DescribeCertificates`（复数）批量查询
2. **技术文档须明确说明** — 为何采用轮询代替批量、以及出于稳定性考虑**本期不升级** Tencent Cloud Go SDK
3. **限流** — 将 `DescribeCertificate` 请求接入 Controller 已有的腾讯云 API 令牌桶限流，避免每轮 Checker 产生大量瞬时 API 请求

**Agent 确认结论**：

| 项 | 决策 |
|----|------|
| 云 API | `DescribeCertificate` 逐 ID；`SSLClient.DescribeCertificates` 仅为上层封装方法名 |
| SDK | 保持 `tencentcloud-sdk-go v1.0.132` 单体包，不迁移子模块 |
| 不升级原因 | 子模块 ≥ v1.0.1090 才支持 `CertIds`；升级须移除单体包并联动 CLB/VPC/CVM，引入 omitnil 等行为变化风险 |
| 限流 | 复用 `TENCENTCLOUD_RATELIMIT_QPS` / `TENCENTCLOUD_RATELIMIT_BUCKET_SIZE`，每次 `DescribeCertificate` 前 `tryThrottle` |
| 指标 method | `DescribeCertificate`（与实际云 API 一致） |

### 第 4 轮澄清（增量：SSL API 域名可配置）

**提问时间**：2026-06-10

**Agent 提问**：

1. 需求在 TAPD 中如何组织？
2. 未配置 SSL 域名时的默认行为？
3. 内网 SSL API 域名是否为 `ssl.internal.tencentcloudapi.com`？
4. 配置项命名是否对齐 CLB 模式？
5. SSL 域名配置作用范围？
6. Namespace Scope 下 per-NS Secret 是否支持独立 SSL 域名？
7. 需求优先级？

**用户回复**：

1. **B** — 作为父需求 #1070046748135050873 的增量修订，直接追加到原需求文档
2. **A** — 默认 `ssl.tencentcloudapi.com`（与现网硬编码行为一致）
3. **A** — 内网域名为 `ssl.internal.tencentcloudapi.com`
4. **确认** — 环境变量 `TENCENTCLOUD_SSL_DOMAIN`；Helm `tencentcloudSslDomain`
5. **B** — 未来所有 SSL API 调用均复用此配置
6. **A** — 不需要，SSL 域名为 Controller 全局配置
7. **A** — Middle 优先级

### 第 5 轮澄清（增量：检查周期调整）

**提问时间**：2026-06-11

**用户诉求（已确认）**：

1. 将证书过期检查的执行周期从 **10 分钟** 修改为 **1 小时（60 分钟）**
2. 作为父需求 #1070046748135050873 的增量修订

**Agent 确认结论**：

| 项 | 决策 |
|----|------|
| 调整范围 | 仅 CertificateChecker |
| 新周期 | 固定 60 分钟 |
| 是否可配置 | 否（与当前 10 分钟固定行为一致） |
| 其它 Checker | 不受影响（仍为 1 分钟） |
| 指标逻辑 | 不变，仅刷新频率降低 |

## 设计文档摘要

- 证书模型：`IngressListenerCertificate` 含 mode/certID/certCaID；仅 HTTPS/TCP_SSL/QUIC 协议展开
- Binding 展开：rule/route/port_mapping 三处独立展开；server + client_ca（MUTUAL）
- API：腾讯云 SSL `DescribeCertificate`（2019-12-05）逐 ID 轮询；应用层每页最多 1000 ID 分组；**不使用** `DescribeCertificates` 批量
- SDK：保持 `tencentcloud-sdk-go v1.0.132`，不升级
- 限流：DescribeCertificate 接入 `TENCENTCLOUD_RATELIMIT_QPS` 令牌桶
- CertEndTime 无效时：CA 类型参考 CAEndTimes（取最早）；仍无效则 INFO 日志 + 无过期时间
- 指标命名空间 `bkbcs_ingressctrl`，子系统 `certificate`
- 主指标 `bkbcs_ingressctrl_certificate_days_until_expiry`（Gauge，已过期为负）
- 辅助指标 `query_success`（0/1）、`bindings_total`
- 复用已有 `api_*` / `lib_*` 指标，method=DescribeCertificate
- 刷新清理：每轮全量重建，DeleteLabelValues 清理上轮残留 series
- API 失败重试 3 次后 query_success=0 + 删除 days_until_expiry series
- 开关：`--certificate_check_enabled`（默认 `false`）；Helm `certificateCheckEnabled`；关闭时不初始化 SSL Client
- SSL endpoint：`TENCENTCLOUD_SSL_DOMAIN`（默认 `ssl.tencentcloudapi.com`）；Helm `tencentcloudSslDomain`；内网示例 `ssl.internal.tencentcloudapi.com`；所有 SSL API 调用统一复用

## 技术澄清

> 澄清日期：2026-06-11（第 5 轮增量修订）
> 需求复杂度：中等
> 澄清轮次：5

### 技术审查结论

- **技术可行性**：✅ 可行
- **技术风险等级**：低
- **审查说明**：核心实现已完成；本轮增量新增 CLI/Helm 开关，复用 `UptimeCheckDisabled` 同类模式，在 `ShouldRegisterCertificateChecker` 与 `main.go` 增加 `CertificateCheckEnabled` 条件判断，无架构冲突。

### 技术方案概述

- **实现方式**：新增 `CertificateChecker` 周期性（60 分钟）List 全集群 Ingress → 展开 SSL 证书 Binding → 对每个 CertificateId 调用腾讯云 `DescribeCertificate`（单数，限流保护）查询失效时间 → 写入/清理 Prometheus 指标。Namespace Scope 模式下按 NS 凭证分组查询，豁免 NS 使用全局凭证。
- **涉及模块**：
  - `internal/check/certificatechecker.go` — Checker 主流程与 Binding 展开
  - `internal/cloud/tencentcloud/sslclient.go` — SSL SDK 封装、`DescribeCertificate` 逐 ID 轮询与限流
  - `internal/cloud/namespacedssl/` — NS Scope 凭证路由（子需求 #2）
  - `internal/metrics/certificate.go` — certificate 子系统指标定义与注册
  - `internal/option/option.go` — 新增 `CertificateCheckEnabled` CLI 参数
  - `main.go` — 云厂商 + 开关双条件 gating 下注册 Checker
  - `docs/features/bcs-ingress-controller/deploy/helm/bcs-ingress-controller/` — Helm values 与 Deployment args / env
  - `internal/cloud/tencentcloud/sslclient.go`、`sdk.go` — SSL endpoint 环境变量读取（F-010 增量）
- **技术选型**：`github.com/tencentcloud/tencentcloud-sdk-go v1.0.132` 单体包内 `tencentcloud/ssl/v20191205`（**不升级 SDK**）；限流复用 `bcs-common/pkg/throttle` 令牌桶（与 CLB `SdkWrapper` 相同环境变量）；指标通过 `sigs.k8s.io/controller-runtime/pkg/metrics.Registry` 注册；日志使用 `blog`。

### 架构影响

- **新增组件**：
  - `CertificateChecker`（`internal/check/`）— 证书过期巡检
  - `SSLClient`（`internal/cloud/tencentcloud/`）— DescribeCertificate 轮询封装与限流
  - `NamespacedSSLClient`（`internal/cloud/namespacedssl/`，子需求 #2）— per-NS 凭证路由
  - `certificate` 指标子系统（`internal/metrics/certificate.go`）
- **变更组件**：
  - `internal/option/option.go` 新增 `--certificate_check_enabled`（默认 `false`）
  - `ShouldRegisterCertificateChecker` 增加 `CertificateCheckEnabled` 参数
  - `main.go` 在 `cloud=tencentcloud && CertificateCheckEnabled` 时注册 Checker
  - Helm Chart 新增 `certificateCheckEnabled` values 并传递至 Deployment args
  - `sslclient.go` 从 `TENCENTCLOUD_SSL_DOMAIN` 读取 endpoint，替代硬编码；Helm 新增 `tencentcloudSslDomain` 并注入 env（F-010）
- **数据模型变更**：无 CRD 变更；Binding 为内存逻辑实体
- **向后兼容性**：✅ 升级后默认关闭，与未配置开关行为一致；显式开启后与既有指标行为相同

### 外部依赖

| 依赖项 | 类型 | 状态 | 接口文档 | 备注 |
|--------|------|------|---------|------|
| 腾讯云 SSL 证书服务 | SDK (ssl/v20191205 @ v1.0.132) | ✅ 已确认 | [DescribeCertificate](https://cloud.tencent.com/document/api/400/41674) | 实际调用单数接口；endpoint 可配置（`TENCENTCLOUD_SSL_DOMAIN`，默认 `ssl.tencentcloudapi.com`，内网 `ssl.internal.tencentcloudapi.com`）；无需 Region；需 `ssl:DescribeCertificates` 只读权限；**不升级 SDK** |
| controller-runtime metrics | 库 | ✅ 已确认 | internal/metrics/ | 沿用 `bkbcs_ingressctrl` 命名空间 |
| per-NS Secret / ControllerConfig | K8s 资源 | ✅ 已确认 | docs/adr/0001-namespace-scope-exemption.md | NS Scope 凭证来源，与 Listener 同步一致 |

### 技术风险

| 风险 ID | 风险描述 | 影响 | 概率 | 应对措施 |
|---------|---------|------|------|---------|
| TR-001 | SSL API 权限未开通导致全量 query_success=0 | 中 | 中 | 默认关闭开关；部署文档强调开启前须开通 `ssl:DescribeCertificates`；告警表达式要求 `query_success==1` |
| TR-004 | 已部署集群升级后默认关闭导致指标消失 | 低 | 高 | 运维须在权限就绪后显式设置 `certificateCheckEnabled: true` 或 `--certificate_check_enabled=true` |
| TR-002 | 大规模集群单次 Check 耗时过长 | 中 | 中 | 逐 ID 轮询 + 令牌桶限流；默认 50 QPS；可调 `TENCENTCLOUD_RATELIMIT_QPS`；AC-P01 放宽至 60 分钟周期 |
| TR-005 | 逐 ID 轮询瞬时 API 量过大触发云侧限流 | 中 | 中 | F-009：每次 DescribeCertificate 前 tryThrottle；与 CLB 共用 QPS 配额 |
| TR-003 | CA 类型证书 CertEndTime 为空 | 低 | 中 | 回退解析 CAEndTimes 取最早值；仍无效则 INFO 日志 + query_success=0 |
| TR-006 | SSL 内网域名配置错误导致 API 不可达 | 中 | 低 | 部署文档说明内网场景须与 CLB 同步配置内网 endpoint；`query_success=0` 便于运维发现 |

### 技术决策记录

| 决策 | 选择方案 | 备选方案 | 选择理由 |
|------|---------|---------|---------|
| Checker 注册方式 | `CheckRunner` 60 分钟周期档位 | 独立 goroutine | 与现有 Checker 模式一致；仅 CertificateChecker 使用 60 分钟周期 |
| SSL SDK 接入层 | `internal/cloud/tencentcloud/sslclient.go` | check 层直接调 SDK | 符合分层架构 ARCH-001；凭证构造可复用 CLB 模式 |
| NS Scope 凭证路由 | 新建 `namespacedssl` 镜像 namespacedlb | 扩展 NamespacedLB | LoadBalance 与 SSL 查询职责分离，避免接口膨胀 |
| 指标 series 清理 | `lastBindingSet` + DeleteLabelValues | 仅 Set 不清理 | 与 ListenerChecker 一致，防止 Ingress 删除后残留 series |
| 子需求交付顺序 | 先核心（全局凭证）后 NS Scope | 一次性交付 | 核心链路可独立验收；NS Scope 为增量扩展，降低风险 |
| 开关默认值 | `certificate_check_enabled=false` | `true`（第 1 轮） | 许多现存账号无 SSL 权限，默认关闭避免升级副作用 |
| 开关命名 | `certificate_check_enabled`（正向） | `certificate_check_disabled` | 与 `uptime_check_disabled` 语义对称但采用正向命名，默认 false 更直观 |
| SSL 查询 API | `DescribeCertificate` 逐 ID 轮询 | `DescribeCertificates` + `CertIds` 批量 | 保持 SDK v1.0.132 不升级，避免 CLB/VPC/ENI 联动升级与 omitnil 风险；轮询为稳定性取舍 |
| DescribeCertificate 限流 | 复用 CLB 令牌桶（`tryThrottle`） | 无限流 / 独立 QPS 配置 | 与现网 CLB 限流配置统一；避免 Checker 每 60 分钟全量轮询时打满 SSL API |
| SSL endpoint 配置 | 环境变量 `TENCENTCLOUD_SSL_DOMAIN` + Helm `tencentcloudSslDomain` | 硬编码 / 与 CLB 共用同一域名 | 对齐 CLB `TENCENTCLOUD_CLB_DOMAIN` 模式；SSL 与 CLB 为不同云产品须独立 endpoint；默认公网域名保持向后兼容 |
| SSL 域名作用域 | Controller 全局 env | per-NS Secret 独立域名 | 与 CLB 域名配置粒度一致；NS Scope 仅区分凭证不区分 endpoint |

### 测试策略

- **单元测试**：
  - Binding 展开：HTTPS/TCP_SSL/QUIC 协议过滤；rule/route/port_mapping 三处展开；MUTUAL client_ca；SNI domain 取值（表驱动，fake client）
  - CertEndTime 解析：正常时间、已过期（负数天数）、CA 类型 CAEndTimes 回退、无效时间
  - SSL 客户端：逐 ID 轮询、限流调用次数、分页（>1000 ID）、重试 3 次失败路径（mock SDK）
  - 开关 gating：`ShouldRegisterCertificateChecker` 在 `CertificateCheckEnabled=false` 时返回 false
  - CLI 参数：`option.go` 中 `certificate_check_enabled` 默认值与 flag 解析
  - SSL endpoint：`NewSSLClient` 在 env 未设置时使用默认 `ssl.tencentcloudapi.com`；设置后使用自定义域名（mock 验证 `cpf.HttpProfile.Endpoint`）
- **集成测试**：
  - CertificateChecker 端到端：fake Ingress List + mock SSL 响应 → 验证指标 label 与 query_success
  - NS Scope 凭证路由（子需求 #2）：参照 `namespacedclient_test.go` 验证 exempt NS 走 defaultClient、普通 NS 走 per-NS Secret
- **测试数据**：使用 fake controller-runtime client 构造 Ingress CR；SSL API 通过 interface mock

### 补充的验收标准

- [ ] **AC-T01**：Given 腾讯云环境且 `--certificate_check_enabled=true` When Controller 启动 Then `CertificateChecker` 注册到 `CheckRunner` 的 60 分钟周期列表且每 60 分钟触发 `Run()`
- [ ] **AC-T02**：Given `--certificate_check_enabled=false` When Controller 启动 Then 不初始化 SSL Client、`CertificateChecker` 未注册到 60 分钟周期列表
- [ ] **AC-T03**：Given DescribeCertificate 调用成功 When Check 完成 Then `bkbcs_ingressctrl_lib_request_total{system="tencentcloud",method="DescribeCertificate"}` 按实际 API 次数递增
- [ ] **AC-T06**：Given `TENCENTCLOUD_RATELIMIT_QPS=10` 且 20 个唯一 certID When 一轮 Check 执行 Then DescribeCertificate 调用间隔受令牌桶约束（可通过 mock 验证 `tryThrottle` 调用次数 ≥ 20）
- [ ] **AC-T04**：Given 上轮存在某 Binding、本轮 Ingress 已删除 When Check 完成 Then 该 Binding 全部 label 组合的 `days_until_expiry` 与 `query_success` series 均被 DeleteLabelValues 清理
- [ ] **AC-T05**：Given `--is_namespace_scope=false` When Check 执行 Then 所有 DescribeCertificate 调用使用全局 `NewSSLClient()` 凭证，不读取 per-NS Secret
- [ ] **AC-T07**：Given `TENCENTCLOUD_SSL_DOMAIN=ssl.internal.tencentcloudapi.com` When `NewSSLClientWithSecretIDKey` 创建客户端 Then SDK `HttpProfile.Endpoint` 为 `ssl.internal.tencentcloudapi.com`

### 待解决问题

无阻塞性待确认项。

---

## 原需求描述

### 初始需求（2026-06-09）

[bcs-ingress-controller] 新增证书过期时间指标：从 Ingress 展开 SSL 证书 Binding，轮询腾讯云 DescribeCertificate 查询失效时间，上报 Prometheus 指标，支持 Namespace Scope 凭证与 CLI/Helm 开关控制。

### 增量修订（2026-06-10，第 4 轮澄清）

腾讯云的 SSL 证书请求域名需要可以指定，类似 CLB 已有的 `tencentcloudClbDomain: clb.internal.tencentcloudapi.com` 及相关参数，并支持在使用 helm chart 部署时，从 `values.yaml` 中进行指定。

### 增量修订（2026-06-11，第 5 轮澄清）

将证书过期检查（CertificateChecker）的执行周期从 10 分钟调整为 1 小时（60 分钟）；仅影响 CertificateChecker，其它 Checker 保持 1 分钟周期；不新增 CLI / Helm 配置项。
