# 产出文件格式模板

## gardening-report.md 格式

```markdown
# 文档园艺报告

> 扫描时间：${TIMESTAMP}
> 模式：${MODE}（PR 轻量 / 全量扫描）
> 触发来源：${TRIGGER}（手动 / commit 后 hook / 定时）
> last-commit: ${COMMIT_SHA}

## 摘要

| 维度 | 状态 | P0 已修复 | P1 待确认 | Skip |
|------|------|----------|----------|------|
| 路径有效性 | PASS / FIXED / NEEDS_REVIEW | N | N | N |
| Skill 清单 | ... | ... | ... | ... |
| 架构描述 | ... | ... | ... | ... |
| 技术规范版本 | ... | ... | ... | ... |
| 词汇表完整性 | ... | ... | ... | ... |
| 目录结构 | ... | ... | ... | ... |
| 工具依赖一致性 | ... | ... | ... | ... |
| Standards Rules | PASS / FIXED / SKIP | N | N | N |

## 已自动修复（本次）

- [P0] 维度/文件: 修复描述
...

## 待确认方案

见 [gardening-proposals.md](gardening-proposals.md)（共 N 项）

## 跳过项

- [Skip] 维度/文件: 原因描述
```

> `${COMMIT_SHA}`：本次 gardening 结束时的 HEAD commit SHA（完整 40 位）。有修复时取 amend 后的 SHA；无修复时取扫描时刻的 HEAD SHA。PR 模式下次运行时以此值为 diff 起点。

## gardening-proposals.md 格式

```markdown
# 文档园艺修复方案

> 以下方案需要人工 review 后确认执行。
> 确认方式：逐项审查后告知 Agent "执行方案 #N" 或 "跳过方案 #N"。

## 方案 #1

- **维度**：架构描述一致性
- **文件**：docs/harness/architectural-constraints.md
- **偏差**：第 32 行描述 "Service 层不依赖 Repository 层"，但代码中 service/user.go import 了 repository/user_repo.go
- **可能原因**：A) 文档过时需更新 B) 代码违规需修复
- **建议修复**：更新文档，将 Repository 标记为 Service 的允许依赖

---

## 方案 #2
...
```
