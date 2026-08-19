# 文档园艺报告

> 扫描时间：2026-08-19 13:20
> 模式：全量扫描（full）
> 触发来源：手动（/harness-engineering 维护 harness 文档和 graphify 图谱）
> last-commit: d8b34a165fec4aad4524dc3653a13a7ab6855e77

## 摘要

| 维度 | 状态 | P0 已修复 | P1 已修复 | Skip |
|------|------|----------|----------|------|
| 1. 路径有效性 | PASS | 0 | 0 | 2 |
| 2. Skill 清单一致性 | PASS | 0 | 0 | 2 |
| 3. 架构描述一致性 | PASS | 0 | 0 | 1 |
| 4. 技术规范版本 | FIXED | 2 | 0 | 2 |
| 5. 词汇表完整性 | FIXED | 3 | 1 | 0 |
| 6. 目录结构一致性 | FIXED | 2 | 0 | 0 |
| 7. 工具依赖一致性 | FIXED | 2 | 1 | 1 |
| 8. Dev Map 与 IDE 集成 | FIXED | 4 | 0 | 0 |
| 9. 工作流文档同步 | FIXED | 1 | 1 | 1 |
| 10. 业务规范一致性 | SKIP | 0 | 0 | 1 |
| 11. Standards Rules | FIXED | 2 | 0 | 0 |
| 12. AGENTS 项目记忆 | PASS | 0 | 0 | 1 |

## 已自动修复（本次）

- [P0] 维度 4 / `docs/standards/`：用预设覆写 `security-bk-redlines.md`、`quality-code-review.md`，并补齐 `quality-code-review/` 分册
- [P0] 维度 5 / `glossary.md`：「九维度」→「十二维度」；`speckit-executor-agent` → `speckit-execution-agent`
- [P0] 维度 6 / harness 文档：熵管理度量分母改为 12；目录树去掉对 workflow.md 的强制暗示
- [P0] 维度 7 / `tooling.md`：删除「环境状态」列；补登记 `graph-steward`；Skill 描述改为十二维度
- [P0] 维度 8 / `docs/dev-map/`：写入白名单 `.gitignore`；按模板重写 `README.md`；本地全量生成 `graph.json`（2970 nodes / 6110 edges / 199 communities，结果不入库）
- [P0] 维度 8 / IDE：`harness-ide-setup.sh` 更新 `.cursor/rules/graphify.mdc`（graphify-out → docs/dev-map）；创建 `.codebuddy/rules/graphify.md`；hook-guard 已存在跳过
- [P0] 维度 9 / `AGENTS.md`：删除「开发工作流」强制段与「不允许跳过」；全流程表改为可选 flow-steward
- [P0] 维度 11 / `AGENTS.md`：插入「编码前必读」门闩；`docs/standards/README.md` 补「Agent 加载步骤（强制）」与「加载预算」
- [P0] 维度 11 / IDE Rules：`sync-standards-rules.sh` 写入 standards-gate / api / backend（Cursor `.mdc` + CodeBuddy `.md`）
- [P1→已执行] 维度 9 / `tooling.md`：`workflow-agent` 标为废弃，场景改指 flow-steward / graph-engineering
- [P1→已执行] 维度 7 / `tooling.md`：删除文首权威清单引用块
- [P1→已执行] 维度 5 / `glossary.md`：删除「待补充标记」词条

## 待确认方案

无（用户已「全部确认」）

## 跳过项

- [Skip] 维度 1：`project.json` 为 gitignore 本地配置；`tool-dependencies.md` 为权威清单文件名而非项目内路径
- [Skip] 维度 2：`work-summary`、`bcs-cluster-checklist` 等未在 tool-dependencies 登记的额外 Skill
- [Skip] 维度 3：OCI 仅有 `docs/reqs/` 需求文档，尚无 `internal/cloud/oci` 代码（规划中）
- [Skip] 维度 4：`backend-k8s-operator.md`、`api-go-restful.md` 为项目定制规范，无预设库对应文件
- [Skip] 维度 7：额外 MCP（BCS API / iWiki 等）未在权威清单，保留在「已配置但未列入」表
- [Skip] 维度 9：`docs/workflow.md` 保留为人类可读参考，不覆写、不强制执行
- [Skip] 维度 10：`docs/business-standards/` 目录不存在
- [Skip] 维度 12：本组件无非根 `**/AGENTS.md`；不把 monorepo 其他组件 AGENTS 索引进本文件
