# Validate — Security（全量复核）

**commit**: `62759c390`  
**范围**: F-001~F-011（五轮澄清合并）  
**verdict**: LGTM

## 安全红线检查

| 检查项 | 结论 | 说明 |
|--------|------|------|
| 凭证/密钥不入指标 label | ✅ | 仅暴露 cert_id、过期天数等业务维度 |
| 凭证来源 | ✅ | 全局 env / per-NS Secret / ControllerConfig；复用 ADR-0001 |
| 日志不打印 Secret 内容 | ✅ | sslclient 仅 V(3) 打印请求 JSON，不含 AK/SK |
| 最小权限 | ✅ | 仅 `ssl:DescribeCertificates` 只读；无写操作 |
| 默认关闭降低权限风险 | ✅ | `certificate_check_enabled=false`；无权限账号升级无副作用 |
| 输入校验 | ✅ | certID 来自 CR Spec，经 K8s API 读取；时间解析有格式校验 |
| 限流防滥用 | ✅ | 共享令牌桶约束 DescribeCertificate QPS |
| 故障隔离 | ✅ | 单 NS / 单 cert 失败不影响其它 Binding 与 Controller 主流程 |

## 第 2~5 轮增量安全项

| 轮次 | 变更 | 结论 |
|------|------|------|
| 第 2 轮 | CLI/Helm 开关默认关闭 | ✅ 避免无 SSL 权限账号启动失败或无效调用 |
| 第 3 轮 | 共享限流 | ✅ 降低 SSL API 突发流量风险 |
| 第 4 轮 | SSL 域名 env 注入 | ✅ 无新增凭证面；域名错误仅导致 query_success=0 |
| 第 5 轮 | 周期 60 分钟 | ✅ 降低 API 调用频率；无安全面扩大 |

## 待关注（非阻塞）

| 级别 | 项 | 说明 |
|------|-----|------|
| [建议] | V(3) 日志含 API 响应 JSON | 生产环境默认不输出；若开启高 verbosity 需注意响应体大小 |

## 备注

无 CRITICAL/HIGH 安全问题；符合 `docs/standards/security-bk-redlines.md` 要求。
