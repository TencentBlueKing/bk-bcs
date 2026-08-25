# Reentry Protocol — 可重入协议

## 目录

- [§1. 两类重入的正式定义](#1-两类重入的正式定义)
  - [1.1 代问重入](#11-代问重入blocked-on-questions-reentry)
  - [1.2 回退重入](#12-回退重入rollback-reentry)
  - [1.3 两类重入对照](#13-两类重入对照)
- [§2. 决策矩阵](#2-决策矩阵subagent-返回--pipeline-主编排动作)
- [§3. pipeline 主编排动作序列](#3-pipeline-主编排动作序列四种典型流程)
  - [3.1 代问重入](#31-代问重入b)
  - [3.2 同 attempt 内 round 重试](#32-同-attempt-内-round-重试f--m--r--w)
  - [3.3 回退重入](#33-回退重入cdghiknopsvx)
  - [3.4 Validate 四段并行 + 修复决策](#34-validate-四段并行--修复决策t--u)
- [§4. 失败触发源的映射](#4-失败触发源的映射)
- [§5. code_preserved 信号的端到端传播](#5-code_preserved-信号的端到端传播)
- [§6. 幂等性与审计约束](#6-幂等性与审计约束)
- [§7. 与恢复机制的衔接](#7-与恢复机制的衔接)

---

本文件定义 pipeline 主编排在子 skill 返回**非 `ok`**状态时如何决策与重入。前三份 references
（`context-and-meta-template.md` / `questions-md-template.md` / `subagent-prompt-template.md`）
定义了**各自文件**的格式与契约；本文件则把它们串起来，形成**端到端的可重入流程**。

适用范围：`tapd-story-specify` / `tapd-story-plan` / `tapd-story-tasks` /
`tapd-story-implement` / `tapd-story-validate` 五个走 subagent 的子 skill。

---

## 1. 两类重入的正式定义

### 1.1 代问重入（Blocked-on-Questions Reentry）

**触发**：subagent 返回 `status=blocked_on_questions`（**仅 specify 阶段**）。
**本质**：同一 attempt 内，subagent 因无法自答的新问题暂停，由主会话代问 → 用户答复
  → 重新拉起 subagent 继续。

**标志**：
- `meta.yaml.attempts` **不变**
- `meta.yaml.specify_attempts[-1].rounds` **追加**一条 round +1 记录
- `context.md` 通常不重写（白名单未变）；`context_revision` 不变
- 代码产物（不涉及）/ 文档产物：由下一轮 subagent 覆盖生成

### 1.2 回退重入（Rollback Reentry）

**触发**（三选一）：
1. subagent 返回 `status=fail`
2. subagent 返回 `status=ok` 但 `compliance.verdict ∈ {spec_insufficient,
   plan_insufficient}`
3. 同 attempt 内 round 超限（specify 代问 >5 / 其他阶段 round >3）

**本质**：跨 attempt 回退——当前阶段或上游阶段的产出需要重做，带着结构化改进方案
（`iteration-patches/attempt-${N+1}.md`）从 `target_phase_to_re_enter` 重新进入。

**标志**：
- 仅当 `attempts < max_attempts` 时，`meta.yaml.attempts` **+1**；否则以 `attempts exhausted` fail 卡点退出
- 生成新的 `iteration-patches/attempt-${N+1}.md`
- `context.md` 按 patch 重写；`context_revision` +1
- 需求 phase 回退到 `target_phase_to_re_enter` 对应链节点
- 产物保留（不手工删除），由下一轮 subagent 覆盖；代码产物按
  `code_preserved` 信号决定保留或回滚

### 1.3 两类重入对照

| 维度 | 代问重入 | 回退重入 |
|------|---------|---------|
| 触发状态 | `blocked_on_questions` | `fail` / `*_insufficient` / round 超限 |
| 适用阶段 | **仅** specify | 所有五个 subagent 阶段 |
| meta.yaml.attempts | 不变 | 仅当 `attempts < max_attempts` 时 +1 |
| round | +1 | 重置为 1 |
| context.md 重写 | 否 | 是（context_revision +1）|
| 跨阶段 | 不跨（就在本阶段） | 跨（回到 `target_phase_to_re_enter`）|
| iteration-patches | 不生成 | **必须**生成 `attempt-${N+1}.md` |
| 代码处理 | 不涉及 | 按 `code_preserved` 信号决定 |

---

## 2. 决策矩阵（subagent 返回 → pipeline 主编排动作）

pipeline 主编排收到 subagent 返回 JSON 后，按下表定位"动作编号"执行：

| # | 当前 phase | subagent status | compliance.verdict | round 上限 | pipeline 主编排动作 |
|---|----------|-----------------|-------------------|----------|-----------|
| A | specify | ok | — | — | **推进**：phase → tech-clarified，清空 last_failure |
| B | specify | blocked_on_questions | — | round <5 | **代问重入**（见 §3.1）|
| C | specify | blocked_on_questions | — | round =5 | **回退重入**（超限升级，target=specify，patch 含 patch_to_req 把剩余 open 问题汇总给用户补 req.md）|
| D | specify | fail | — | — | **回退重入**（target 由 pipeline 主编排在退出报告中给出建议，由调用方按 `state-mutation-guide.md` §2 卡点 3 确认）|
| E | plan | ok | pass | — | **推进**：phase → researched |
| F | plan | ok | needs_fix | round <3 | **同 attempt 内 round +1 重跑 plan**（见 §3.2）|
| G | plan | ok | needs_fix | round =3 | **回退重入**（超限升级，target=plan，patch 来自 compliance.violations）|
| H | plan | ok | spec_insufficient | — | **回退重入**（target=specify）|
| I | plan | fail | — | — | **回退重入**（target 由 issue 分析；通常 specify 或 plan）|
| J | tasks-generate | ok | — | — | **进入下一段** tasks-analyze |
| K | tasks-generate | fail | — | — | **回退重入**（target 由 issue 分析；通常 plan）|
| L | tasks-analyze | ok | pass | — | **推进**：phase → tasks-generated |
| M | tasks-analyze | ok | needs_fix | round <3 | **同 attempt 内 round +1 重跑 tasks-generate + tasks-analyze**（整组重跑）|
| N | tasks-analyze | ok | needs_fix | round =3 | **回退重入**（超限升级，target=tasks）|
| O | tasks-analyze | ok | plan_insufficient | — | **回退重入**（target=plan）|
| P | tasks-analyze | ok | spec_insufficient | — | **回退重入**（target=specify）|
| Q | implement | ok | — | — | **推进**：phase → implemented（迭代级 `iteration-state.json.status=implementing` 由 runner 在 loop 中维护，pipeline 不直接写）|
| R | implement | fail | — | round <3 | **主会话与用户讨论 issue**，三选一：原地修复重试（§3.2）/ 回退重入 / 终止 |
| S | implement | fail | — | round =3 | **回退重入**（超限升级，target 按 issue 分析确定）|
| T | validate-{arch,security,codereview,test} | ok | LGTM | — | 等待四段全部返回；全部 LGTM 后**进入修复决策**（见 §3.4）|
| U | validate-{arch,security,codereview,test} | ok | needs_fix | — | 等待四段全部返回；按汇总决策（见 §3.4）|
| V | validate-{arch,security,codereview,test} | fail | — | — | **回退重入**（target 由 issue 分析）|
| W | validate-fix | ok | — | round <3 | **回到 validate-{arch,security,codereview,test} 四段并行校验**（round +1）|
| X | validate-fix | fail | — | round =3 | **回退重入**（超限升级，target 按评审报告归因最深处）|

> 表中"round <3"与"round =3"合并计算的是"同 attempt 内 round 已经是第几次"。
> 特殊地：specify 阶段 round 上限是 **5**；其他阶段上限是 **3**。
> 任一回退重入在 attempts 递增前还必须满足 `attempts < max_attempts`；否则写
> `attempts exhausted` fail 卡点并停止，见 `state-mutation-guide.md` 的“Attempt 上限”。

---

## 3. pipeline 主编排动作序列（四种典型流程）

### 3.1 代问重入（#B）

```
pipeline 主编排
  1. 解析 subagent JSON：
     - status=blocked_on_questions
     - questions_delta.added_open = ["Q3", "Q7"]
  2. 展开代问：
     - 从 ${WORK_DIR}/questions.md 按编号提取 Q3 / Q7 条目
     - 向用户展示"问题 + 影响 + 建议候选"，请求答复
  3. 收集用户答复：
     - 将答复按 §questions-md-template §5.2 同步更新 questions.md
     - 对应条目 status: open → answered
     - 将答复融入 ${WORK_DIR}/req.md 的"技术澄清"章节（覆盖性写入）
  4. 更新 meta.yaml：
     - specify_attempts[-1].rounds 追加一条：
         { round: ${新 round}, status: answered, at: ${TS} }
  5. 重新渲染 SUBAGENT_PROMPT（§subagent-prompt-template §3.3）：
     - 追加"本轮为第 ${ROUND} 次代问重入 — 用户已答复"增量段
  6. 重新 Task() 拉起 subagent（round +1，attempts 不变）
  7. 回到决策矩阵头部
```

### 3.2 同 attempt 内 round 重试（#F / #M / #R / #W）

```
pipeline 主编排
  1. 解析 subagent JSON：
     - status=ok / compliance.verdict=needs_fix
     - 或 status=fail + 用户决定"重试"
  2. 更新 meta.yaml：
     - ${STAGE}_attempts[-1].rounds 追加一条：
         { round: ${新 round}, status: fail/needs_fix,
           issue: ${issue 摘要},
           failed_tests: [...],       # 仅 implement
           key_errors: [...] }        # 仅 implement
  3. 代码类阶段（implement / validate-fix）：
     - **不**回滚代码，不动 implement_baseline_commit
     - 维持当前工作目录状态作为下轮修复的起点
     文档类阶段（plan / tasks-analyze）：
     - 产物文件保留快照 .prev-attempt${N}-round${R-1}（subagent 命令块中已处理）
     - 下轮 subagent 会覆盖生成
  4. context.md 不重写（context_revision 不变）
  5. 重新渲染 SUBAGENT_PROMPT（§subagent-prompt-template §3.2）：
     - 追加"已知失败原因"增量段（含 issue / failed_tests / key_errors）
  6. 重新 Task() 拉起 subagent（round +1，attempts 不变）
  7. 回到决策矩阵头部；若 round 达到上限，按 §3.3 升级
```

### 3.3 回退重入（#C/D/G/H/I/K/N/O/P/S/V/X）

```
pipeline 主编排
  1. 解析失败根因：
     - 优先用 subagent 返回的 compliance.violations[].root_cause_attribution
     - 次用 issue 字段 + 上一阶段产物
     - 必要时与用户讨论确认

  2. 确定 target_phase_to_re_enter：
     - 按归因"最深处"原则：spec > plan > tasks > code
     - 多源归因取最深

  3. 决定 code_preserved：
     - validate 阶段触发的回退：默认 code_preserved = true
       （例外：evidence 显示实现路径整体走偏 → 用户确认后置 false）
     - implement 阶段触发的回退：默认 code_preserved = false
       （代码刚开始实现，回滚代价低）
     - plan / tasks 阶段触发的回退：不涉及代码（无 code_preserved 字段）

  4. 生成 iteration-patches/attempt-${N+1}.md（§context-and-meta-template §3.2）：
     - failed_phase
     - target_phase_to_re_enter
     - root_cause（复用 plan-report.md / tasks-report.md 的归因段）
     - code_preserved（仅 validate/implement 回退时有意义）
     - unresolved_findings（仅 validate 回退时必填）
     - patch_to_context（需要新增/移除的 background 或 code scope 条目）
     - patch_to_req（需要补充到 req.md 的需求细节）
     - expected_improvements（下游优先覆盖的要点清单）

  5. 代码处理（仅 code_preserved=false 时）：
     - 由 target 对应的子 skill 在重新进入前做路径级 git checkout
       （实际执行见 tapd-story-implement §4.2 / tapd-story-validate §5.2）
     - 回滚锚点始终用 meta.yaml.implement_baseline_commit，**不是**已删除的
       validate_baseline_commit
     - 严禁任何全局 git 操作（reset --hard / stash / switch 等）

  6. 更新 meta.yaml：
     - 校验 attempts < max_attempts 后 attempts: +1
     - last_failure: { phase, at, summary, attempt_patch, unresolved_findings_count }
     - history: 追加一条 { kind: rollback 或 code_preserved_rollback,
                           from_phase, to_phase, attempt_delta, note }
     - 清空当前阶段的 ${STAGE}_attempts[-1].rounds
       （新 attempt 从 rounds=[] 起记）

  7. 重写 context.md（§context-and-meta-template §1.4）：
     - 按 target_phase_to_re_enter 选择对应白名单规则
     - 合并 patch_to_context
     - 合并 Improvement notes（摘自 attempt-${N+1}.md）
     - 若 code_preserved=true，Improvement notes 追加 [code_preserved] 标记行
     - context_revision: +1

  8. 更新 req.md（若 patch_to_req 有内容）：
     - 按 patch_to_req 覆盖对应章节
     - 不动 questions.md（保留历史）

  9. 更新 iteration-state.json：
     - stories.${ID}.phase 回退到 target_phase_to_re_enter 对应链节点

  10. 跳转调度 target_phase_to_re_enter 对应的子 skill
      - 重新从 §1 决策矩阵头部开始循环
```

### 3.4 Validate 四段并行 + 修复决策（#T / #U）

```
pipeline 主编排
  1. 并行 Task() 四段 subagent：
     validate-arch / validate-security / validate-codereview / validate-test
     - 使用同一 attempt/round
     - 各段写入独立 process-validate-*.log

  2. 等待四段全部返回；汇总 compliance.verdict 与 violations

  3. 汇总决策（参见 tapd-story-validate §2）：
     | 条件 | 动作 |
     | 四份全 LGTM，且无 CRITICAL/HIGH finding | 推进 phase → validated（跳过修复）|
     | 任一 needs_fix，全部归因 code-self     | 启动 validate-fix（见下）|
     | 任一 spec_insufficient / plan_insufficient | 回退重入（§3.3，target 按最深归因）|
     | 任一 status=fail                        | 回退重入（§3.3）|

  4. 若启动 validate-fix：
     - Task() 拉起修复 subagent（§subagent-prompt-template §2.9）
     - 修复完成后不推进 phase，**回到步骤 1** 重新并行四段校验
     - round 计数规则：修复 + 四段回归校验 **整组完成后** round +1
     - 同 attempt 内最多 3 个 round（整组算 1 个 round）
     - 3 个 round 后仍 needs_fix → 回退重入（§3.3）
```

---

## 4. 失败触发源的映射

subagent 可能以多种方式表达"出了问题"，pipeline 主编排需要按下表统一映射到两类重入：

| 触发源 | 出处 | 映射到 |
|--------|------|-------|
| `status=fail + issue` | subagent 执行异常 / 越界 / 测试失败 | 回退重入 或 round 重试（视 round 与阶段）|
| `compliance.verdict=needs_fix` | plan / tasks-analyze / validate-* 的自检结论（仅 code-self） | round 重试（达上限后升级为回退）|
| `compliance.verdict=spec_insufficient` | plan / tasks-analyze / validate-* | **直接**回退重入（target=specify）|
| `compliance.verdict=plan_insufficient` | tasks-analyze / validate-* | **直接**回退重入（target=plan）|
| `status=blocked_on_questions` | specify 阶段 | 代问重入（达上限后升级为回退）|
| round 达上限（specify >5，其他 >3）| pipeline 主编排计数 | **强制升级**为回退重入 |

原则：
- `needs_fix` 仅表示 `code-self` 可原地修复；上游缺漏归因一律通过 `*_insufficient` 直接回退
- round 上限是**硬约束**，防止无穷重试消耗 token

---

## 5. code_preserved 信号的端到端传播

`code_preserved` 是 validate 阶段回退时引入的关键信号，用来让下游子 skill 识别
"代码保留 vs 回滚"。完整传播链：

```
validate 阶段四段评审发现 spec/plan-insufficient
  │
  ▼
§3.3 步骤 3：默认 code_preserved=true
  │
  ▼
§3.3 步骤 4：写入 iteration-patches/attempt-${N+1}.md 的 code_preserved 字段
  │
  ▼
§3.3 步骤 5：不做 git checkout（保留当前代码）
  │
  ▼
§3.3 步骤 7：context.md 的 Improvement notes 追加 [code_preserved] 标记行 +
            unresolved_findings 清单
  │
  ▼
target_phase_to_re_enter（specify 或 plan）子 skill 重新进入
  │ ── 子 skill 不关心 code_preserved（只修改规范文档）
  ▼
推进回到 tapd-story-implement 子 skill 重新进入
  │
  ▼
implement 识别 attempt-${N+1}.md.code_preserved=true：
  - 保留原 meta.yaml.implement_baseline_commit（不重新取）
  - SUBAGENT_PROMPT 注入 unresolved_findings + patch 摘要
  - speckit.implement 做差量修复（§subagent-prompt-template §2.5 关键约束段）
  │
  ▼
implement → validate 正常推进
```

### code_preserved=false 的例外分支

仅当 validate 回退时，用户或 pipeline 主编排基于 subagent 返回的 evidence 判定"实现路径整体走偏"（如分层架构根本性
错误、需要重新选型），才置 `code_preserved=false`：

- §3.3 步骤 3 明确 false
- §3.3 步骤 5 执行路径级 `git checkout <implement_baseline_commit> -- <Code scope paths>`
- 回滚显式排除 `${WORKDIR}/` 目录（保留所有可重入元数据）
- implement 重新进入时取新的 baseline（因为代码已经回滚）

---

## 6. 幂等性与审计约束

可重入机制必须满足以下硬约束，否则可能造成状态混乱或审计轨迹缺失：

### 6.1 幂等性

| 操作 | 幂等保证 |
|------|---------|
| subagent 重复拉起（网络重试等） | 允许，产物覆盖生成；覆盖前做 `.prev-attemptN-roundR` 快照 |
| context.md 重写 | 允许，context_revision 单调递增从不回退 |
| iteration-patches/attempt-${N}.md 写入 | 一次性写入，**不允许修改**；若 patch 需要补充信息，追加说明章节 |
| meta.yaml 更新 | 全文覆盖写入（非追加字段），确保原子性 |

### 6.2 审计完整性

| 规则 | 说明 |
|------|------|
| `history[]` 仅追加不删除 | 完整记录每次 phase 推进/回退/代问事件 |
| `iteration-patches/attempt-*.md` 归档永久保留 | 即使该 patch 已被覆盖，文件也保留 |
| `process.log` 仅追加不截断 | banner 区分不同 attempt/round/stage，便于事后回溯 |
| `*.prev-attempt${N}-round${R}` 快照文件保留到迭代结束 | 由 `tapd-iteration-report` 或 `tapd-story-commit` 决定是否清理 |
| `questions.md` 条目状态不回退 | 已 answered 的条目不可改回 open；见 §questions-md-template §2.2 |

### 6.3 并发防护

当前架构下**不存在并发**——pipeline 主编排是单线程决策，subagent 运行时主会话不会并行调度。
唯一的"并行"是 validate 阶段四段同时 Task()，但它们写入的是四份**不同**文件
（validate-arch/security/codereview/test-report.md），且各自写入独立日志文件（process-validate-*.log），天然无冲突。

---

## 7. 与恢复机制的衔接

pipeline 自身**不读** `iteration-state.json`，恢复时只依赖 `${WORKDIR}/meta.yaml` 与
`${WORKDIR}/questions.md`。pipeline 每次启动都执行同一套幂等流程：

1. 加载 `${WORKDIR}/meta.yaml`（不存在则按入参初始化为 `phase=initialized`、`attempts=1`、`round=1`）
2. 按 `state-mutation-guide.md` §4 做外部修改合法性校验：
   - phase 与产物不匹配（如 `phase=confirmed` 但 `tasks.md` 缺失）→ 写 `last_failure.type=mutation_invalid` 退出
3. 按 `state-mutation-guide.md` §3 卡点判定优先级检查：
   - `phase == committed` → 完工退出
   - `last_failure` 非空：
     - `type=system` 且 `attempts==1` → 自动清空并 `attempts++`，继续推进
     - 其他类型 → 退出（卡点信息已落盘）
   - `pending_review` 非空 → 退出（confirm 卡点）
   - `questions.md` 有 `[open]` 条目 → 退出（blocked 卡点）
4. 否则根据 `meta.yaml.phase` 选择对应子 skill 继续推进
5. 若读取的 `iteration-patches/attempt-${attempts}.md` 存在（外部回退重入时由调用方写入）：
   作为下一轮 subagent 的改进方案输入（按 `subagent-prompt-template.md §3.1` 重入增量段渲染）

**pipeline 视角下"恢复"和"正常启动"是同一段代码**——pipeline 是幂等的，每次调起都读
meta.yaml 决定下一步。没有专门的"恢复模式"分支。

**迭代级恢复**（多个需求的批量调度恢复、`iteration-state.json` 与各需求 meta.yaml 的派生
缓存同步）由 runner 负责，详见 `../../tapd-iteration-runner/references/recovery.md`。
pipeline 不参与迭代级恢复，仅响应 runner 的下一次 Task 调度。
