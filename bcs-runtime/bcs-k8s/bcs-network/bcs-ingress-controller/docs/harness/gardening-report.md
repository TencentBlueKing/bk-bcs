# 文档园艺报告

> 扫描时间：2026-08-11 17:06
> 模式：全量扫描（full）
> 触发来源：手动（/harness-engineering 更新 harness 文档）
> last-commit: 6f8c4cbafcebcbe2d894323d8d85aef8e9c4828c

## 摘要

| 维度 | 状态 | P0 已修复 | P1 已修复 | Skip |
|------|------|----------|----------|------|
| 1. 路径有效性 | FIXED | 1 | 0 | 1 |
| 2. Skill 清单一致性 | FIXED | 1 | 0 | 2 |
| 3. 架构描述一致性 | FIXED | 0 | 1 | 1 |
| 4. 技术规范版本 | PASS | 0 | 0 | 2 |
| 5. 词汇表完整性 | FIXED | 2 | 0 | 0 |
| 6. 目录结构一致性 | FIXED | 1 | 0 | 0 |
| 7. 工具依赖一致性 | FIXED | 1 | 1 | 0 |
| 8. Dev Map 与 IDE 集成 | FIXED | 3 | 0 | 0 |
| 9. 工作流文档同步 | FIXED | 2 | 0 | 0 |

## 已自动修复（本次）

- [P0] 维度 1 / `architectural-constraints.md`：`portpool_controller.go` → `portpoolcontroller.go`
- [P0] 维度 2+7 / `tooling.md`：按 tool-dependencies 重建 Skill/Agent/MCP/CLI；补 gongfeng、登记 Skill；嵌套子 skill 合并到 harness-engineering
- [P0] 维度 5 / `glossary.md` + harness 文档：「八维度」→「九维度」；新增「唯一知识来源」
- [P0] 维度 6 / `AGENTS.md`：目录树同步 scripts/、workflow.md、internal 分层说明；≤100 行
- [P0] 维度 8 / `docs/dev-map/graph.json`：graphify AST 全量生成（3975 nodes）
- [P0] 维度 8 / IDE：创建 `.cursor/rules/graphify.mdc`；合并 `.codebuddy/settings.json` hook-guard
- [P0] 维度 9 / 创建 `docs/workflow.md`；`AGENTS.md` 注入「开发工作流」章节
- [P0] `execution-verification.md`：指标表补 certificate / nodeinfoexporter
- [P0] `.gitattributes`：graph.json merge=graphify；gardening-report merge=ours
- [P1→已执行] 维度 3 / `architectural-constraints.md`：补充 10 个 internal 新模块分层
- [P1→已执行] 维度 7+9 / 场景 B/E/F/G + Graph Engineering；AGENTS 全流程覆盖规划扩展

## 待确认方案

无（用户已「全部确认」；方案 #3 OCI 按约定跳过）

## 跳过项

- [Skip] 维度 1：`project.json` 为 gitignore 本地配置，路径缺失记为环境缺口而非失效引用
- [Skip] 维度 2：`work-summary`、`bcs-cluster-checklist` 等未在 tool-dependencies 登记的额外 Skill
- [Skip] 维度 3 / 方案 #3：OCI 仅有 `docs/reqs/` 需求文档，尚无 `internal/cloud/oci` 代码（规划中）
- [Skip] 维度 4：`backend-k8s-operator.md`、`api-go-restful.md` 为项目定制规范，无预设库对应文件；`quality-code-review.md` / `security-bk-redlines.md` 与预设一致
