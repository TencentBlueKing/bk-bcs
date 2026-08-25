# Code Review — 路由表（Route First）

> **强制**：收集 diff 并开始按清单挑问题前，先完成本表匹配。  
> 仅服务 `code-review` skill / `code-reviewer` agent。不依赖其他角色 skill。

## 匹配规则

1. 自上而下；命中主路由后停止。
2. 优先级：**范围场景** → **清单深度** → **报告形态**。
3. 方法论正文仍以本 skill 的 5 步流程为准；本表决定**加读哪些 reference**。

## 路由矩阵

| 信号 | 主路由 ID | 必读 reference | 禁止事项 |
|------|-----------|----------------|----------|
| 默认 / 智能评审 / 未指定范围 | `cr-auto` | `git-scenarios.md`（范围检测）+ `checklist.md` | 未看 diff 凭空评论 |
| staged / 提交前 | `cr-staged` | `git-scenarios.md` + `checklist.md` | 评审未暂存的无关脏文件当本次范围 |
| last-commit / 指定 commit | `cr-commit` | `git-scenarios.md` + `checklist.md` | 扩到整个分支历史 |
| AI 生成代码重点审 | `cr-ai` | `checklist.md`（AI 节）+ `confidence-filtering.md` | 只挑风格 nit |
| 需要严格置信度/降噪 | `cr-confidence` | `confidence-filtering.md` + `checklist.md` | 低置信度当 CRITICAL |
| 用户指定报告格式/评分 | `cr-report` | `report-format.md` / `scoring-standard.md`（按需）+ `checklist.md` | 自创与 skill 冲突的分级 |
| 只要示例对照 | `cr-example` | `report-examples.md` | 用示例代替真实评审 |

## 始终可附带

| 条件 | 附加加载 |
|------|----------|
| 输出正式报告 | `report-format.md` |
| 涉及批准/合并建议 | `scoring-standard.md` |
| 后端 Go/Node 专项 | `checklist.md` 对应专项节 |

## 路由输出示例

```text
[code-review routing] id=cr-staged → git-scenarios + checklist
```
