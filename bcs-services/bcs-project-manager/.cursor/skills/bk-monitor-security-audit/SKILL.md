---
name: bk-monitor-security-audit
description: 对前端代码进行安全审计，检测 XSS、CSRF 等漏洞。当用户请求代码审查或询问代码安全性时使用。
---

# 前端代码安全审计

## 触发场景

1. 代码提交/审查请求
2. 询问代码安全性/漏洞
3. 涉及 DOM 操作、URL 处理、用户输入
4. 提及 XSS、CSRF、注入等安全关键词

## 审计检查清单

1. DOM 操作安全性（innerHTML、v-html）
2. URL/重定向安全
3. 跨域通信安全（postMessage）
4. 动态代码执行（eval）
5. 敏感信息处理
6. 原型污染/ReDoS

## 工作流程

1. 读取代码文件
2. 按检查清单逐项审计
3. 按 report-template.md 输出报告

---

## 📦 可用资源

- `skill://bk-monitor-security-audit/references/audit-rules.md`
- `skill://bk-monitor-security-audit/references/report-template.md`
- `skill://bk-monitor-security-audit/references/security-checklist.md`


---
## 📦 可用资源

- `skill://bk-monitor-security-audit/references/audit-rules.md`
- `skill://bk-monitor-security-audit/references/report-template.md`
- `skill://bk-monitor-security-audit/references/security-checklist.md`

> 根据 SKILL.md 中的 IF-THEN 规则判断是否需要加载
