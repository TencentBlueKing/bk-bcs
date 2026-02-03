---
id: security/web-security-guide
name: Web 安全漏洞学习指南
category: security
description: OWASP 十大漏洞原理、影响与修复方案，覆盖 Python/Java 场景
tags: [security, owasp, vulnerability, injection, xss, csrf, ssrf]
updated_at: 2026-01-23
---

# Web 安全漏洞学习指南

## ⚠️ 核心规则

1. **永不信任用户输入** - 所有外部数据必须校验、转义、参数化
2. **最小权限原则** - 仅授予完成任务所需的最小权限
3. **纵深防御** - 多层安全措施，不依赖单一防护

## 十大漏洞速查

| 漏洞 | 危害 | 核心防御 |
|------|------|----------|
| 🔴 注入 | RCE/数据泄露 | 参数化查询 |
| 🔴 XSS | 会话劫持 | 转义输出 |
| 🔴 认证缺陷 | 账户接管 | 强Token+限速 |
| 🔴 敏感数据泄露 | 隐私泄露 | 加密+脱敏 |
| 🔴 访问控制缺失 | 越权操作 | 后端鉴权 |
| 🟡 安全配置错误 | 信息泄露 | 关闭Debug |
| 🟡 CSRF | 伪造操作 | Token验证 |
| 🟡 反序列化 | RCE | 禁用危险接口 |
| 🟡 SSRF | 内网探测 | 白名单URL |
| ⚪ 日志不足 | 无法溯源 | 完整审计 |

## 📦 按需加载资源

| 漏洞类型 | URI |
|----------|-----|
| 注入漏洞 | `skill://web-security-guide/references/injection.md` |
| XSS攻击 | `skill://web-security-guide/references/xss.md` |
| 认证会话 | `skill://web-security-guide/references/auth-session.md` |
| 数据泄露 | `skill://web-security-guide/references/data-exposure.md` |
| 访问控制 | `skill://web-security-guide/references/access-control.md` |
| 配置错误 | `skill://web-security-guide/references/security-config.md` |
| CSRF | `skill://web-security-guide/references/csrf.md` |
| 反序列化 | `skill://web-security-guide/references/deserialization.md` |
| SSRF | `skill://web-security-guide/references/ssrf.md` |
| 日志监控 | `skill://web-security-guide/references/logging-monitoring.md` |

> 💡 先用速查表定位问题，再按需加载详细文档


---
## 📦 可用资源

- `skill://web-security-guide/references/access-control.md`
- `skill://web-security-guide/references/auth-session.md`
- `skill://web-security-guide/references/csrf.md`
- `skill://web-security-guide/references/data-exposure.md`
- `skill://web-security-guide/references/deserialization.md`
- `skill://web-security-guide/references/injection.md`
- `skill://web-security-guide/references/logging-monitoring.md`
- `skill://web-security-guide/references/security-config.md`
- `skill://web-security-guide/references/ssrf.md`
- `skill://web-security-guide/references/xss.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
