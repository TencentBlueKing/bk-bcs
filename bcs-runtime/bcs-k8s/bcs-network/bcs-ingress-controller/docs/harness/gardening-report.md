# 文档园艺报告

> 扫描时间：2026-06-11
> 模式：定向维护（用户指定规范增补）
> 触发来源：用户手动（/harness-engineering 维护现有文档）

## 摘要

| 维度 | 状态 | P0 已修复 | P1 已修复 | Skip |
|------|------|----------|----------|------|
| 1. 路径有效性 | PASS | 0 | 0 | 0 |
| 2. Skill 清单一致性 | PASS | 0 | 0 | 1 |
| 3. 架构描述一致性 | FIXED | 0 | 3 | 0 |
| 4. 技术规范版本 | PASS | 0 | 0 | 2 |
| 5. 词汇表完整性 | FIXED | 2 | 0 | 0 |
| 6. 目录结构一致性 | FIXED | 2 | 0 | 0 |
| 7. 工具依赖一致性 | PASS | 0 | 0 | 0 |
| 8. Dev Map 一致性 | FIXED | 0 | 20 | 0 |

## 已自动修复（本次）

- [P0] 用户指定 / `docs/standards/backend-k8s-operator.md`：§3.4 增补导出函数 GoDoc 注释要求；函数名 ≤ 35 字符扩展至测试函数；§8/§9/§12 同步
- [P0] 用户指定 / `AGENTS.md`：核心约定增补导出 GoDoc 与测试函数命名限制
- [P0] 用户指定 / `docs/harness/architectural-constraints.md`：ARCH-007 扩展至测试函数；新增 ARCH-009 导出函数注释规则
- [P0] 维度 6 / `AGENTS.md`：目录树补充 `docs/reqs/`、`cli-util/`
- [P0] 维度 6 / `AGENTS.md`：已完成特性补充 SSL 证书过期 Prometheus 指标
- [P0] 维度 5 / `docs/glossary.md`：移除重复 ADR 条目
- [P0] 维度 5 / `docs/glossary.md`：新增 CertificateChecker、NamespacedSSL、Cert Binding 术语
- [P1→已执行] 维度 3 / `architectural-constraints.md`：补充 namespacedssl 依赖规则、CertificateChecker 职责、SSL Parse 边界
- [P1→已执行] 维度 8 / `docs/dev-map/source-index.md`：同步 20 个新增/遗漏源文件（证书过期特性 + inspector + cli-util）
- [P1→已执行] 维度 8 / `docs/dev-map/module-index.md`：更新 check/metrics/cloud-adapters 模块，新增 bcs-ingress-inspector、cli-util 模块
- [P1→已执行] 维度 8 / `docs/dev-map/module-dependencies.md`：补充 CertificateChecker → namespacedssl/sslclient/metrics 依赖链

## 待确认方案

无（P1 项归属明确，已随本次扫描一并修复）

## 跳过项

- [Skip] 维度 2：`bcs-cluster-checklist` 为接入仓额外安装 Skill，未在 tool-dependencies.md 登记，不参与一致性检查
- [Skip] 维度 4：`backend-k8s-operator.md`、`api-go-restful.md` 为项目定制规范，无预设库对应文件

## 结论

✅ **Harness 文档已增补两条编码规范（导出 GoDoc、函数名含测试 ≤ 35 字符）。** 与 `.specify/memory/constitution.md` 已有条款保持一致。
