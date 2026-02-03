---
id: sec-nodejs-check
name: Node.js 安全审查
category: security
description: 检查 RCE、SSRF、SQL 注入、路径穿越等安全问题，支持 Express/Koa/NestJS
tags: [security, nodejs, backend, ssrf, sql-injection]
updated_at: 2026-01-20
allowed-tools: [Read, Grep, Glob, Shell]
---

# Node.js 安全审查

## ⚠️ 核心规则

1. **永不信任用户输入** - 所有请求数据须 Schema 验证
2. **安全默认** - 使用安全 API（execFile 而非 exec）
3. **纵深防御** - 输入验证 + 参数化查询 + 输出编码

## 快速开始

```bash
/nodejs-security-check                    # 智能扫描 src 目录
/nodejs-security-check file src/xxx.js    # 扫描指定文件
/nodejs-security-check report             # 生成详细报告
```

## 问题分级

| 前缀 | 含义 | 处理方式 |
|------|------|----------|
| `🔴 严重` | 可被直接利用 | 阻止发布 |
| `🟡 中等` | 需特定条件 | 尽快修复 |
| `⚪ 建议` | 最佳实践 | 可选优化 |

## 检查维度

| 维度 | 检查项 |
|------|--------|
| 输入验证 | Schema 验证、HPP 防护、请求体限制 |
| 命令执行 | exec/spawn 注入、shell: true |
| 文件操作 | 路径穿越、上传安全、ZipSlip |
| 网络请求 | SSRF、私网阻断、DNS 重绑定 |
| 数据库 | SQL 注入、NoSQL 注入、Mass Assignment |
| 认证授权 | JWT 算法、会话安全、权限校验 |

## 📦 按需加载资源

| 资源 | URI |
|-----|-----|
| 完整检查清单 | `skill://nodejs-security-check/references/checklist.md` |
| 修复示例 | `skill://nodejs-security-check/references/fix-examples.md` |
| 评分标准 | `skill://nodejs-security-check/references/scoring-standard.md` |


---
## 📦 可用资源

- `skill://nodejs-security-check/references/checklist.md`
- `skill://nodejs-security-check/references/fix-examples.md`
- `skill://nodejs-security-check/references/report-format.md`
- `skill://nodejs-security-check/references/scoring-standard.md`
- `skill://nodejs-security-check/references/security-toolkit.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
