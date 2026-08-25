# 统一报告模板

> 本文件定义 pipeline 中所有合规/评审报告的统一输出格式。
> 适用阶段：plan（plan-report.md）/ tasks-analyze（tasks-report.md）/ validate-arch /
> validate-security / validate-codereview / validate-test。
> speckit.analyze 等命令在执行时通过 prompt 参数引用本文件路径，确保输出格式统一。

## 报告文件命名

`{phase}-report.md`，其中 phase 取值：
- `plan` → `plan-report.md`
- `tasks` → `tasks-report.md`
- `validate-arch` → `validate-arch-report.md`
- `validate-security` → `validate-security-report.md`
- `validate-codereview` → `validate-codereview-report.md`
- `validate-test` → `validate-test-report.md`

## 报告模板

```markdown
# {Phase} Report — Story ${ID}

## Verdict
<pass | needs_fix | spec_insufficient | plan_insufficient | LGTM>

## Checked artifacts
- <被检查的产物文件路径列表>

## Reference baselines
- <本次检查依据的规范文档路径列表>

## Findings

### A1
- **类别**：<Completeness / Architecture / Security / Duplication / CodeStyle / Performance / Testability 等>
- **严重性**：<CRITICAL | HIGH | MEDIUM | LOW>
- **位置**：<file:line 或 file:Lstart-Lend>
- **总结**：<一句话概括问题>
- **根因**：<spec-self | plan-self | tasks-self | code-self | spec-insufficient | plan-insufficient>
- **修改建议**：<具体修复方案>

### A2
...

（无 finding 时写 "无"）
```

## Verdict 判定规则

### plan 阶段

| 条件 | Verdict |
|------|---------|
| 无 finding，或仅 MEDIUM/LOW 级别 | `pass` |
| 存在 HIGH/CRITICAL，全部归因 plan-self | `needs_fix` |
| 存在 HIGH/CRITICAL，任一归因 spec-insufficient | `spec_insufficient` |

### tasks-analyze 阶段

| 条件 | Verdict |
|------|---------|
| 无 finding，或仅 MEDIUM/LOW | `pass` |
| 存在 HIGH/CRITICAL，全部归因 tasks-self | `needs_fix` |
| 存在 HIGH/CRITICAL，任一归因 plan-insufficient | `plan_insufficient` |
| 存在 HIGH/CRITICAL，任一归因 spec-insufficient | `spec_insufficient` |
| 多源归因时取最深处（spec > plan > tasks） | |

### validate 四段（arch / security / codereview / test）

| 条件 | Verdict |
|------|---------|
| 无 [必须] 项（无 CRITICAL/HIGH） | `LGTM` |
| 存在 [必须] 项，且全部归因 `code-self` | `needs_fix` |
| 存在 [必须] 项，任一归因 `plan-insufficient` | `plan_insufficient` |
| 存在 [必须] 项，任一归因 `spec-insufficient` | `spec_insufficient` |

> CodeReview 的 severity 兼容 [必须/建议/Nit] 三级标注：
> CRITICAL/HIGH → [必须]；MEDIUM → [建议]；LOW → [Nit]
>
> 每条 validate finding 必须标注 `code-self` / `spec-insufficient` /
> `plan-insufficient` 根因。多个根因时取最深上游根因作为 verdict；pipeline 对
> `*_insufficient` 必须回退重入而非原地修复。

## Finding 根因枚举

| 阶段 | 可用根因值 |
|------|-----------|
| plan | `plan-self` / `spec-insufficient` |
| tasks-analyze | `tasks-self` / `plan-insufficient` / `spec-insufficient` |
| validate-* | `code-self` / `spec-insufficient` / `plan-insufficient` |
