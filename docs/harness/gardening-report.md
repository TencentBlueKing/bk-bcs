# 文档园艺报告

> 扫描时间：2026-08-17 15:16 +0800
> 模式：full（全量扫描）
> 触发来源：手动
> last-commit: pending

## 摘要

| 维度 | 状态 | P0 已修复 | P1 待确认 | Skip |
|------|------|----------|----------|------|
| 路径有效性 | PASS | 0 | 0 | 0 |
| Skill 清单 | PASS | 0 | 0 | 0 |
| 架构描述 | PASS | 0 | 0 | 0 |
| 技术规范版本 | PASS | 0 | 0 | 0 |
| 词汇表完整性 | PASS | 0 | 0 | 0 |
| 目录结构 | PASS | 0 | 0 | 0 |
| 工具依赖一致性 | PASS | 0 | 0 | 0 |
| Dev Map 与 IDE 集成 | SKIP | 0 | 0 | 1 |
| 工作流文档同步 | SKIP | 0 | 0 | 1 |
| 业务规范一致性 | SKIP | 0 | 0 | 1 |
| Standards Rules | PASS | 0 | 0 | 0 |
| AGENTS 项目记忆 | PASS | 0 | 0 | 0 |
| 前置 .gitattributes | FIXED | 1 | 0 | 0 |

## 已自动修复（本次）

- [P0] 前置步骤 A / `.gitattributes`：原文件无结尾换行，`harness-setup-git.sh` 把 `docs/dev-map/graph.json merge=graphify` 粘到 `bcs-unified-apiserver` 规则同一行。已拆成独立行，并写入 `docs/harness/gardening-report.md merge=ours`。
- [P1 已确认·不改文件] 维度 12：用户选择方案 A，不把 `.agents/skills/harness-engineering/tests/fixtures/**/AGENTS.md` 收入根索引；根表保持既有 3 条业务工作单元。

## 待确认方案

无（共 0 项）。P1 方案 #1 已确认执行 A，不写入 [gardening-proposals.md](gardening-proposals.md)。

## 跳过项

- [Skip] 维度 8：`$SKILL_ROOT/graphify/SKILL.md` 不存在，根目录无 `docs/dev-map/`，dev-map 与 graphify IDE 集成检测跳过。
- [Skip] 维度 9：无 `workflow-agent` /「不允许跳过工作流」/「## 开发工作流」残留。
- [Skip] 维度 10：`docs/business-standards/README.md` 索引中的 `example-domain.md` 为骨架示例行，正文已写明「当前尚无业务规范文件」。
- [Skip] 维度 6：`apm_modules/`、`build/` 已被 `.gitignore`，不视为目录树缺失。
- [Skip] 项目域：磁盘上 `apm_modules/` 含非隐藏 `SKILL.md`，但该目录 gitignore，不按 mixed 处理，本仓按 code-project。
- [Skip] 维度 12 记忆对比：根 `AGENTS.md` 尚未入库，无 git 基线可对比入口记忆。

## 扫描说明

- `$SKILL_ROOT`：`.agents/skills`（接入仓推荐布局）
- `$PROJECT_DOMAIN`：code-project
- `$REPORT_PATH`：`docs/harness/gardening-report.md`
- `$DIFF_BASE`：空（非 issue 功能分支切出点）
- 工作单元已索引：`bcs-ingress-controller`、`bcs-drplan-controller`、`bcs-terraform-bkprovider`
