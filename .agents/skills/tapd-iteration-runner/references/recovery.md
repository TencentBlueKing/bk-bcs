# Iteration Recovery（迭代级断点恢复）

> 本文档定义 `tapd-iteration-runner` 在检测到未完成的
> `specs/${VERSION}/iteration-state.json` 时的恢复流程。**仅涉及迭代级状态恢复**；
> 单需求层面的卡点（`pending_review` / `questions.md [open]` / `last_failure`）
> 由 loop 在恢复后的下一轮调度时按 `../tapd-story-pipeline/references/state-mutation-guide.md` §7
> runner 视角速查表自然处理，本文档不重复。

## 1. 触发时机

`tapd-iteration-init` 在阶段 2 检测到 `iteration-state.json` 已存在且 `status` 不为 `reported` 时进入恢复。
若 `status=reported`，提示用户该迭代已完成；询问"重新开始"或"退出"。

## 2. 步骤 1：展示进度摘要

读取 `iteration-state.json` 并扫描每个非 `committed` 需求的 `meta.yaml` / `questions.md`，向用户呈现：

```
检测到迭代 ${ITERATION_ID} 的开发进度:
- [OK] 已完成: ${COMMITTED_COUNT} 个需求
- [..] 进行中: ${IN_PROGRESS_COUNT} 个需求
- [  ] 未开始: ${NOT_STARTED_COUNT} 个需求
- [!]  有未答复问题: ${BLOCKED_COUNT} 个需求（meta.yaml.pending_review 非空 或 questions.md 有 [open]）
- [!]  有未解失败:   ${LAST_FAILURE_COUNT} 个需求（meta.yaml.last_failure 非空）
```

如果用户消息已表达继续意图（如"帮我继续"、"接着做"），视为已确认，直接进入步骤 3。
否则询问"是否继续？"。

## 3. 步骤 2：派生缓存回填

`iteration-state.json.stories.*.children.*.phase` 与 `stats` 是 `meta.yaml` 的派生缓存。
迭代过程中可能因外部中断而与各需求 `meta.yaml` 失同步，恢复时需做一次性回填：

对 `sequence` 中**每个非 `committed` 需求**：

1. 读取 `${WORKDIR}/meta.yaml`
2. 把 `meta.yaml.phase` 同步到 `iteration-state.json.stories.*.children.*.phase`
3. 把 `meta.yaml.stats` 同步到 `iteration-state.json.stories.*.children.*.stats`
4. 若 `meta.yaml` 不存在或字段缺失：保持 `iteration-state.json.stories.*.phase = initialized`，由后续 loop 内联执行 pipeline 时自然初始化

> `stories.*.started_at` 与 `stats.duration_sec` 是 runner 自有字段（非 meta.yaml 派生），
> 已持久化于 `iteration-state.json`，恢复时保持不变——committed 时仍以原 `started_at` 计算 `duration_sec`。

## 4. 步骤 3：迭代级一致性检查

| 异常情况 | 处理 |
|---------|------|
| `status = implementing` 但 `sequence` 为空 | 回退到 `tapd-iteration-analysis` 重新生成 `sequence` |
| `sequence` 为空但 `status = analyzed` | 同上 |
| `.tmp/${ID}.md` 缺失但 `sequence` 中包含 `${ID}` | 提示用户检查 `tapd-iteration-analysis` 是否曾完整执行；若否，重新执行 analysis |
| 所有需求 phase 均为 `committed` 但 `status ≠ reported` | 直接调度 `tapd-iteration-report`，跳过 loop |
| `status = bugfix` | 进入 §6 Bugfix 模式 |

需求级一致性问题（meta.yaml 字段不全、history 与 phase 不一致等）**不在恢复阶段处理**——pipeline
启动时会按 `state-mutation-guide.md` §4 校验，失败则写 `last_failure.type=mutation_invalid`，
由 runner loop 按速查表与用户对话修复。

## 5. 步骤 4：进入 loop 继续执行

恢复完毕后，runner 直接进入正常 loop：

```
对 sequence 中每个 ID：
  while phase != committed：
    根据 meta.yaml / questions.md 决策 → 内联执行 pipeline 或与用户对话
```

loop 会自动遇到并处理：
- 仍在 `pending_review` 卡点的需求 → 展示 artifacts 给用户，等待"通过 / 回退 / 放弃"
- 仍有 `[open]` 问题的需求 → 收集用户答复
- `last_failure` 非空的需求 → 按 type 自动重试或与用户对话

恢复流程**不直接进入代问 / confirm 卡点对话**——这些都委托给 loop 的标准决策路径，
确保恢复路径与正常路径行为一致。

## 6. Bugfix 模式恢复

当 `iteration-state.json.status = bugfix` 时（已发布版本出现 bug，回到迭代修复）：

1. `patch` 字段自增 1
2. 确定需要修复的需求——用户在进入 bugfix 模式时指定需要修复的需求 ID 列表（或由 runner 交互式询问"哪些需求需要修复？"），runner 仅对这些需求的 `meta.yaml.phase` 重置到 `confirmed`
   （让 pipeline 从 implement 重新跑起，复用既有的 spec / plan / tasks）
3. 重新进入 §5 loop
4. 全部修复完成后 `status` 自动恢复为 `implementing`（部分需求）或 `completed`（全部完成）

## 7. 不做自动修复的场景

以下情况恢复流程**不**做自动修复，直接呈现给用户人工处理：

| 场景 | 原因 |
|------|------|
| `iteration-state.json` 解析失败（JSON 格式损坏）| 提示用户备份后人工修复；不自动覆盖（详见 `error-handling.md` §1）|
| 某需求目录 `meta.yaml` YAML 格式破损 | 提示用户备份后人工修复；恢复跳过该需求继续处理其他需求 |
| `.tmp/${ID}.md` 与 `sequence` 不一致 | 见 §4 异常表 |

## 8. 参考

- 迭代层错误规则：`error-handling.md`
- 卡点处理速查表（runner 视角）：`../tapd-story-pipeline/references/state-mutation-guide.md` §7
- 需求级状态机字段：`../tapd-story-pipeline/references/context-and-meta-template.md`
