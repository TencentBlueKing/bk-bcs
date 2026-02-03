---
id: security/bk-security-redlines
name: 蓝鲸代码安全三大红线
category: security
description: 基于 IEG 安全规范，覆盖输入校验、鉴权、数据加密三大高危领域
tags: [security, input-validation, auth, encryption, redlines]
updated_at: 2026-01-23
---

# 蓝鲸代码安全三大红线

## ⚠️ 核心规则

### 红线 1：外部输入未校验

外部输入进入**高危操作**前，必须完成**服务端强约束校验**（类型/长度/格式/白名单），不可绕过。

**高危操作**：命令执行、模板解释/eval、文件路径、请求目标、SQL/NoSQL 构造、渲染解析、协议输出

### 红线 2：敏感接口未鉴权

高危能力或敏感资源接口必须同时实施**身份认证 + 权限校验**，缺一即违规。

**禁止**：接口无认证、仅登录不校验权限、依赖前端字段/URL/IP 判权、服务间调用无鉴权

### 红线 3：敏感数据未加密

敏感数据（密码/Token/AKSK/私钥/PII）在存储、传输、日志、导出任一环节保护不当即违规。

**禁止**：硬编码凭证、明文存储、异常回显细节、日志记录敏感字段、URL 携带 token

## 常见错误

| 错误做法 | 正确做法 |
|---------|---------|
| ❌ 仅前端/网关校验 | ✅ 服务端强约束校验 |
| ❌ 仅判空/黑名单 | ✅ 白名单 + 类型/长度/格式 |
| ❌ 内网可达即放通 | ✅ 服务间调用也需鉴权 |
| ❌ 日志记录完整请求 | ✅ 脱敏后记录 |

## 📦 按需加载资源

| 资源 | URI | 说明 |
|-----|-----|------|
| 输入校验详解 | `skill://bk-security-redlines/references/input-validation.md` | 7 类高危操作场景 |
| 鉴权检查详解 | `skill://bk-security-redlines/references/auth-check.md` | 8 类鉴权缺失场景 |
| 数据加密详解 | `skill://bk-security-redlines/references/data-encryption.md` | 8 类加密缺失场景 |


---
## 📦 可用资源

- `skill://bk-security-redlines/references/auth-check.md`
- `skill://bk-security-redlines/references/data-encryption.md`
- `skill://bk-security-redlines/references/input-validation.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
