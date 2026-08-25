# Context.md 与 Meta.yaml 模板

## 目录

- [Context.md 与 Meta.yaml 模板](#contextmd-与-metayaml-模板)
  - [目录](#目录)
  - [1. context.md](#1-contextmd)
    - [1.1 文件职责](#11-文件职责)
    - [1.2 生命周期](#12-生命周期)
    - [1.3 文件结构](#13-文件结构)
    - [1.4 各阶段白名单生成规则](#14-各阶段白名单生成规则)
    - [1.5 subagent 遵守的约定（重复强调）](#15-subagent-遵守的约定重复强调)
  - [2. meta.yaml](#2-metayaml)
    - [2.1 文件职责](#21-文件职责)
    - [2.2 生命周期](#22-生命周期)
    - [2.3 完整字段模板](#23-完整字段模板)
    - [2.4 字段所有权](#24-字段所有权)
    - [2.5 与 iteration-state.json 的边界](#25-与-iteration-statejson-的边界)
  - [3. iteration-patches/attempt-${N}.md](#3-iteration-patchesattempt-nmd)
    - [3.1 文件职责](#31-文件职责)
    - [3.2 文件结构](#32-文件结构)
    - [3.3 命名与持久化](#33-命名与持久化)

---

本文件定义需求目录下两份核心元数据文件的结构与字段契约：

- **`context.md`**：本需求当前阶段的"上下文白名单"。由 pipeline 主编排按阶段重写，subagent 只读。
  通过白名单明确指定每个 subagent 可读的背景知识路径（而非无差别加载 AGENTS.md），
  让每个 subagent 获得**精确、聚焦**的背景知识。
- **`meta.yaml`**：本需求的执行元数据（非内容产物）。承载 attempts/rounds/baseline 等
  与可重入执行强相关的状态，与迭代级 `iteration-state.json`（由 `tapd-iteration-runner` 拥有）职责分离。

两份文件都存放在需求目录 `${WORKDIR}/`，与 `req.md` / `spec.md` 等文档并列。
独立模式下 `${WORKDIR}` 默认为 `specs/stories/${ID}/`；被 runner 调度时为 `specs/${VERSION}/${ID}/`。

---

## 1. context.md

### 1.1 文件职责

`context.md` 是 pipeline 主编排给 subagent 的"工作须知"：告诉 subagent 本阶段**必读什么、
可读什么、不可读什么、允许触达什么代码**。subagent 必须严格按白名单工作，
不主动扩展范围。

### 1.2 生命周期

| 触发时机 | 操作 | `meta.yaml.context_revision` |
|---------|------|-----------------------------|
| 首次进入某阶段 | pipeline 主编排按该阶段规则生成 context.md | +1（从 0 起记） |
| 回退重入携带 `patch_to_context` | pipeline 主编排合并 patch 后重写 context.md | +1 |
| 同 attempt 内 round 重试 | 不重写（白名单未变，仅 prompt 附加"已知失败原因"） | 不变 |
| 代问重入（specify 阶段） | 不重写（仅 questions.md 与 round 变化） | 不变 |

> `context_revision` 递增是单调的，**从不回退**——即使用户撤销某次 patch，也应当以
> "再加一次 patch 把白名单改回来"的方式实现，以保持审计轨迹。

### 1.3 文件结构

context.md 是标准 Markdown，由五个一级标题段组成。段落顺序固定，缺失段用"无"明确标注。

```markdown
# Context for Story ${ID}

## Stage
<specify | plan | tasks | implement | validate>

## Source artifacts
<本需求产物中本阶段必读的文件列表。按顺序递进：>
- ${WORKDIR}/req.md                       # 所有阶段必读
- ${WORKDIR}/spec.md                      # plan 阶段起必读
- ${WORKDIR}/plan.md                      # tasks 阶段起必读
- ${WORKDIR}/research.md                  # tasks 阶段起必读
- ${WORKDIR}/data-model.md                # 若存在则必读
- ${WORKDIR}/tasks.md                     # implement 阶段起必读
- ${WORKDIR}/validate-arch-report.md        # validate fix 阶段必读
- ${WORKDIR}/validate-security-report.md    # 同上
- ${WORKDIR}/validate-codereview-report.md  # 同上
- ${WORKDIR}/validate-test-report.md        # 同上

## Project background
<由 pipeline 主编排从 AGENTS.md 抽取的"二级文档路径"白名单，每条必带"用途"注释：>
- docs/architecture/layering.md                       # 用途：分层依赖约束
- docs/security/redlines.md                           # 用途：安全红线（输入校验/鉴权/加密）
- proto/order/v1/order.proto                          # 用途：订单域接口契约
- internal/order/README.md                            # 用途：订单模块边界与职责
- .specify/memory/constitution.md                     # 用途：项目宪章（plan/tasks/validate 必读）
- skills/bk-security-redlines/SKILL.md                # 用途：三大安全红线（validate 安全维度）

# 说明：CodeReview 维度由 code-reviewer agent 经 use_skill("code-review") 加载统一 CR 方法论驱动，方法论由该 skill 承载，无需在 context 白名单显式引入 skills/code-review/

## Code scope
<本需求允许触达的代码路径白名单。subagent 不得越界读写：>
- internal/order/**
- api/order/v1/**
- pkg/common/errors/**

## Improvement notes
<若本轮为回退重入（attempts > 1），pipeline 主编排从 iteration-patches/attempt-${attempts}.md
 摘要写入此段。首轮执行时填"无"。>
- [spec.md] 补充 inventory.QueryStock 的响应字段定义（attempt-2 patch_to_req）
- [Code scope] 新增白名单路径 internal/inventory/client/**（attempt-2 patch_to_context）
- [Project background] 新增 docs/domain/inventory-state-machine.md（attempt-2 patch_to_context）

<若本轮 attempt 上一轮由 validate 阶段以 code_preserved=true 回退而来，额外追加：>
- [code_preserved] 本轮保留上一轮的代码实现，仅针对 unresolved_findings 做差量修复：
  - CRITICAL/architecture: internal/order/service/create_order.go:L124 违反分层依赖
  - HIGH/security: internal/order/repository/query.go:L56 输入校验缺失
```

### 1.4 各阶段白名单生成规则

pipeline 主编排在进入每个阶段前按下表生成 context.md。"新增"列表示本阶段相对上一阶段的增量；
"继承"表示从上一阶段 context.md 原样搬运。

| 进入阶段 | Source artifacts（新增） | Project background（新增） | Code scope |
|---------|--------------------------|--------------------------|-----------|
| `specify` | req.md | AGENTS.md 中"架构/接口/编码/安全"四类一级条目对应的二级文档；项目宪章 | 暂不生成（尚无代码触达）|
| `plan` | spec.md（spec.md 内显式引用的文档路径也加入 Project background） | 继承；补充 spec.md 新引用的文档 | 暂不生成 |
| `tasks` | plan.md / research.md / data-model.md | 继承；补充 plan.md / research.md 中确定的模块文档 | 按 plan.md 中模块范围初步生成 |
| `implement` | tasks.md | 继承 | 按 tasks.md 中任务汇总的代码路径精确列出 |
| `validate` | validate-arch/security/codereview/test-report.md（修复子 subagent 需要时） | 继承；补充 `skills/bk-security-redlines/`（codereview 维度由 code-reviewer agent 经 use_skill("code-review") 加载统一 CR 方法论驱动，方法论由该 skill 承载，无需在白名单补充 skills/code-review/）| 继承 implement 阶段 |

> pipeline 主编排在生成 context.md 前，应先检查 `iteration-patches/attempt-${attempts}.md` 是否
> 含 `patch_to_context`，合并后再写入。合并规则：patch 中的 `added` 条目追加到对应段末尾，
> `removed` 条目从对应段移除；同名条目按 patch 覆盖。

### 1.5 subagent 遵守的约定（重复强调）

- 只读 `Source artifacts` 与 `Project background` 中列出的文件；
- 写代码仅限 `Code scope` 范围；
- 不主动读取 AGENTS.md 全文；
- 发现白名单不够用时：
  - specify 阶段：追加 `questions.md` 条目 + `status=blocked_on_questions` 返回
  - 其他阶段：在回传 JSON 的 `issue` 中说明缺失文档，返回 `status=fail`

---

## 2. meta.yaml

### 2.1 文件职责

`meta.yaml` 是需求级"执行元数据"**唯一事实源**——pipeline 自治的核心契约。它承载本需求的全部状态，包括：

- 需求标识：`id` / `workspace_id` / `agent_tool` / `owner`
- 状态游标：`phase`（当前阶段）/ `attempts`（回退重入计数）/ `round`（同 phase 内代问轮次）/ `context_revision`
- 卡点信号：`last_failure`（fail 卡点）/ `pending_review`（confirm 卡点）
- 代码统计：`stats`（commit 阶段写入）
- 成本度量：`stats.cost`（pipeline 步骤 5a 实时累加 + commit 阶段以 ts_event 去重的最终对账补齐）
- 执行轨迹：`history`（每段成功 / 回退后追加）
- 内部细化游标（可选）：`implement_baseline_commit` / `specify_attempts[]` / `implement_attempts[]` / `validate_attempts[]`

字段读写权限、外部修改语义、卡点判定优先级详见 `state-mutation-guide.md`。

### 2.2 生命周期

- **创建**：pipeline 首次执行时初始化（若 `${WORKDIR}/meta.yaml` 不存在）；初始 `phase=initialized`、`attempts=1`、`round=1`、`context_revision=0`
- **写入者**：pipeline 主编排（推进 phase / 写 last_failure / 写 pending_review）+ 各子 skill（追加内部细化游标 + 写 stats）
- **外部修改者**：调用方（runner 或用户）按 `state-mutation-guide.md` 修改卡点字段后再次调起 pipeline
- **subagent 只读**：内部子 skill 通过 SUBAGENT_PROMPT 接收 meta.yaml 的渲染结果，不直接写入
- **校验**：pipeline 启动时对外部修改做合法性校验（见 `state-mutation-guide.md` §4），违规写入 `last_failure.type=mutation_invalid` 后退出

### 2.3 完整字段模板

```yaml
# ${WORKDIR}/meta.yaml
# 字段说明详见 state-mutation-guide.md §1 字段修改权限矩阵

# ===== 需求标识 =====
id: "1000000755129275824"              # TAPD 需求长 ID
workspace_id: "20000001"               # TAPD 工作空间 ID（首次执行时由调用方入参落入）
agent_tool: agent                      # agent / claude / cursor 等
owner: "jimwu"                         # 当前开发者

# ===== 状态游标 =====
phase: tasks-generated                 # 见 state-mutation-guide.md §4 phase 链
attempts: 2                            # 当前 attempt 号（回退重入 +1），从 1 起记
max_attempts: 5                        # 最大 attempt 数；达到上限后禁止再次回退重入或 retry
round: 1                               # 同 attempt 内 round 号（代问 / 原地修复 +1）
context_revision: 4                    # context.md 当前版本号，从 0 起记，单调递增

# ===== 技术路由 =====
tech_type: backend                     # 需求技术类型：backend / frontend / fullstack（由 specify 阶段写入，只读）
implement_rounds:                      # fullstack 需求实现轮次（backend=1, frontend=1, fullstack=2）
  total: 1                             # 总轮次数（backend/frontend=1，fullstack=2）
  completed: 0                         # 已完成轮次（0=尚未开始）
  current: backend                     # 当前正在执行的 round（backend / frontend）

# ===== 基线 =====
implement_baseline_commit: "a1b2c3d"   # 进入 tapd-story-implement 时 git rev-parse HEAD
                                       # 仅用于 implement / validate 极端回退时的路径级回滚
                                       # 首次进入 implement 前为空字符串

# ===== 卡点信号 =====
# 每次成功推进 phase 后，pipeline 把 last_failure 与 pending_review 都置为 null

last_failure:                          # 非空 = fail 卡点
  type: semantic                       # system | semantic | mutation_invalid
  phase: validate                      # 失败发生时所处的 phase
  message: "架构校验发现订单服务越层访问底层 DAO"   # 根因摘要 ≤200 字
  occurred_at: "2026-05-13T14:23:11+08:00"
  evidence:                            # 可选；semantic / mutation_invalid 推荐填
    - "process.log: line 1287-1305"
    - "iteration-patches/attempt-3.md"

pending_review:                        # 非空 = confirm 卡点（仅 phase=tasks-generated 时由 pipeline 写）
  artifacts: ["spec.md", "plan.md", "tasks.md"]
  ready_at: "2026-05-13T11:05:00+08:00"

# ===== 代码统计（commit 阶段写入）=====
stats:
  total: 0
  add_code: 0
  delete_code: 0
  logic_code: 0
  test_code: 0
  docs: 0
  files: 0

  # ===== 成本度量（pipeline 主编排实时累加）=====
  cost:
    # ===== 总累加（pipeline 步骤 5a 实时累加；hook 写 cost-events.jsonl 后由 5a 倒序匹配）=====
    total_duration_sec: 0.0
    total_cost_usd: 0.0                # 语义变更：现承载 credit 累加值（积分单位），不再是美元
    total_credit: 0.0                  # 与 total_cost_usd 同值的别名（过渡期保留，commit.md 模板使用）
    total_input_tokens: 0
    total_output_tokens: 0
    total_cache_tokens: 0
    total_cached_write_tokens: 0
    total_cached_miss_tokens: 0
    subagent_calls: 0

    # ===== 单次调用记录 =====
    # 5a 即时 best-effort append；commit 阶段"成本最终对账门禁"以 ts_event 为去重主键
    # 全量补齐遗漏事件。ts_event（hook 写入时刻，每条唯一）是对账去重的权威键。
    per_call:
      - stage: "specify"
        attempt: 1
        round: 1
        ts_event:      "2026-05-14T10:00:30+08:00"   # hook 写入时刻
        ts_marker:     "2026-05-14T10:00:00+08:00"   # 渲染 prompt 时刻（关联键的一部分）
        duration_sec:  45.2
        input_tokens:  28000
        output_tokens: 6500
        cache_tokens:  4200
        credit:        0.15

# ===== 子 skill 内部 round 轨迹（可选细化游标）=====
# 仅在执行过程中维护最近 attempt 的 rounds，历史 attempt 的轨迹归到 history

specify_attempts:                      # specify 阶段的代问 round 轨迹
  - attempt: 1
    rounds:
      - round: 1
        at: "2026-05-12T10:00:00+08:00"
        status: blocked_on_questions
        added_open: ["Q1", "Q2"]
        self_resolved: []
      - round: 2
        at: "2026-05-12T10:20:00+08:00"
        status: ok

implement_attempts:                    # implement 阶段的原地修复 round 轨迹
  - attempt: 2
    rounds:
      - round: 1
        at: "2026-05-13T09:10:00+08:00"
        status: fail
        issue: "订单创建服务的单元测试 TestCreateOrder_InvalidAmount 未通过"
        failed_tests: ["internal/order/service/create_order_test.go::TestCreateOrder_InvalidAmount"]
        key_errors:
          - "expected error type ErrInvalidAmount, got nil"
      - round: 2
        at: "2026-05-13T09:45:00+08:00"
        status: ok

validate_attempts:                     # validate 阶段的四段校验 + 修复 round 轨迹
  - attempt: 2
    rounds:
      - round: 1
        at: "2026-05-13T13:00:00+08:00"
        arch:  { verdict: needs_fix, findings_count: 1 }
        security: { verdict: LGTM, findings_count: 0 }
        codereview: { verdict: needs_fix, findings_count: 3 }
        test: { verdict: LGTM, findings_count: 0 }
        fix_status: ok                 # 修复 subagent 的 status
      - round: 2
        at: "2026-05-13T13:40:00+08:00"
        arch:  { verdict: LGTM, findings_count: 0 }
        security: { verdict: LGTM, findings_count: 0 }
        codereview: { verdict: LGTM, findings_count: 0 }
        test: { verdict: LGTM, findings_count: 0 }
        fix_status: skipped            # 四段全 LGTM，无需修复

# ===== 累计历史（每段成功推进 / 回退后 append 一条）=====
history:
  - at: "2026-05-12T10:20:00+08:00"
    kind: phase_advance                # phase_advance | rollback | code_preserved_rollback | resume
    phase: tech-clarified
    attempt: 1
    round: 2
    note: ""
  - at: "2026-05-12T15:42:00+08:00"
    kind: rollback
    from_phase: researched
    to_phase: tech-clarified
    attempt_delta: "1 -> 2"
    note: "spec.md 缺少库存查询接口响应字段定义"
  - at: "2026-05-13T14:25:00+08:00"
    kind: code_preserved_rollback
    from_phase: validated
    to_phase: tech-clarified
    attempt_delta: "2 -> 3"
    unresolved_findings_count: 2
    note: "架构层面分层违规，需要修订 spec.md 中的服务边界描述"
```

### 2.4 字段所有权

| 字段 | pipeline 写入 | 外部允许写入 | 读取者 |
|------|-------------|------------|--------|
| `id` | 初始化 | ❌ | 所有子 skill / runner |
| `workspace_id` | 初始化（首次执行时由调用方入参落入） | ✅（下次调起生效）| commit 子 skill 调 `stories_update` / req.md 兜底注入调 `stories_get` |
| `agent_tool` | 初始化 | ✅（下次调起生效）| pipeline 主编排 / SUBAGENT_PROMPT 渲染 |
| `owner` | 初始化 | ✅ | pipeline 主编排 / commit 子 skill |
| `phase` | 推进 / 回退时写入 | ✅（仅沿链回退或 `tasks-generated → confirmed`，详见 `state-mutation-guide.md` §4）| 所有子 skill / runner（派生缓存到 iteration-state.json）|
| `attempts` | 回退重入 / retry 时 +1（不得超过 `max_attempts`） | ❌ | 所有子 skill |
| `max_attempts` | 初始化时写入（当前固定为 `3`） | ❌ | pipeline 主编排 / runner |
| `round` | 代问 / 原地修复时 +1 | ✅（清零）| 同上 |
| `context_revision` | 重写 context.md 时 +1 | ❌ | 恢复流程 / 审计 |
| `tech_type` | specify 阶段（由 tech-lead subagent 写入）| ❌ 只读（写入后不可修改）| pipeline 路由 implement / validate-fix 使用 |
| `implement_rounds` | implement 阶段（pipeline 主编排维护）| ❌ 外部只读 | 记录 fullstack 串行实现进度 |
| `implement_baseline_commit` | implement 首次进入时由 pipeline 主编排写；回退重入重新进入时重新取一次 | ❌ | implement 子 skill 回退重入、validate 极端回退 |
| `last_failure` | fail 时写入；推进 phase 时清空 | ✅（清空 = 接受失败已被外部修复）| 调用方（runner / 用户）按 `state-mutation-guide.md` §2 卡点 3 处理 |
| `pending_review` | tasks-generated 完成时写入；外部清空后下次调起继续 | ✅（清空 + 把 phase 改为 confirmed）| 调用方按 `state-mutation-guide.md` §2 卡点 1 处理 |
| `stats` | commit 写入 | ❌ | runner 派生缓存到 iteration-state.json / report |
| `specify_attempts[]` / `implement_attempts[]` / `validate_attempts[]` | 各对应子 skill 每轮 subagent 返回后追加 round 条目 | ❌ | 下一轮 SUBAGENT_PROMPT 渲染（取 `[-1].rounds[-1]` 作为"已知失败原因"）|
| `history[]` | 每次 phase 推进 / 回退后 append | ❌ | 恢复流程 / report |

### 2.5 与 iteration-state.json 的边界

`meta.yaml` 是**需求级唯一事实源**，由 pipeline 自治；`iteration-state.json` 是**迭代级聚合视图**，由 runner 拥有。两者职责严格分离，但 runner 在 loop 中会把 meta.yaml 的部分字段同步到 iteration-state.json 作为**派生缓存**，便于全局视角。

| 字段 / 场景 | 归属（事实源） | 派生缓存 |
|------------|--------------|---------|
| 需求当前 phase（tech-clarified / researched / ...） | `meta.yaml.phase` | `iteration-state.json.stories.*.children.*.phase` |
| 需求代码贡献统计（add/delete/logic/test/docs/files）| `meta.yaml.stats` | `iteration-state.json.stories.*.children.*.stats` |
| 需求重试次数、失败上下文、round 轨迹 | `meta.yaml`（attempts / last_failure / *_attempts[] / history） | — |
| 需求回滚锚点（代码基线 commit）| `meta.yaml.implement_baseline_commit` | — |
| confirm 卡点（spec/plan/tasks 待审）| `meta.yaml.pending_review` | — |
| 迭代级状态（initialized / analyzed / implementing / bugfix / completed / reported）| `iteration-state.json.status` | — |
| 需求间依赖顺序 | `iteration-state.json.sequence` | — |
| 父需求结构 | `iteration-state.json.all_parents` / `stories.<PARENT_ID>.children` | — |
| 当前调度的需求 | `iteration-state.json.selected_story` | — |
| 迭代分支 / patch 号 | `iteration-state.json.iter_branch` / `patch` | — |

**原则**：
- pipeline **完全不读** iteration-state.json（pipeline 自治）；
- runner **不直接写** meta.yaml 的非卡点字段（如不主动改 phase / stats / history）；runner 仅在 loop 决策中按 `state-mutation-guide.md` §7 修改卡点字段（清空 last_failure / pending_review、回退 phase 等）；
- runner 在 loop / 恢复时**单向同步** meta.yaml.phase / stats → iteration-state.json，作为全局视图刷新；
- 用户独立模式下不存在 iteration-state.json；用户可按 `state-mutation-guide.md` §8 直接修改 meta.yaml。

---

## 3. iteration-patches/attempt-${N}.md

### 3.1 文件职责

回退重入时由调用方（runner 或用户）写入的改进方案。下游子 skill 重新进入时由 pipeline 主编排
读取并在 SUBAGENT_PROMPT 中引用，确保重试方向明确。

### 3.2 文件结构

```markdown
# Attempt ${N} — Improvement Notes

## 失败阶段
failed_phase: plan                     # specify | plan | tasks | implement | validate

## 回退目标
target_phase_to_re_enter: specify      # 下游重新进入的子 skill 阶段

## 根因
root_cause: |
  spec.md 未定义 inventory.QueryStock 的完整响应字段，导致 plan.md 无法确定
  订单创建流程中库存占用的异常分支处理。违反宪章"外部依赖契约需完整"的条款。

## 代码保留策略
code_preserved: true                   # validate 回退时固定为 true——保留代码做差量修复

## 未解 Findings（仅 validate 回退时需要）
unresolved_findings:                   # 上游修订时参考，下游增量修复时精确定位
  - severity: CRITICAL
    category: architecture
    location: "internal/order/service/create_order.go:L124"
    evidence: "违反分层依赖：service 层直接引用 pkg/dao/raw"
    suggestion: "在 repository 层封装 dao 访问，service 层仅依赖 repository 接口"
    source: "specs/v0.9.x/1234567/validate-arch-report.md"

## 上下文补丁（合并到 context.md）
patch_to_context:
  added:
    project_background:
      - path: "proto/inventory/v1/inventory.proto"
        purpose: "库存查询接口契约"
      - path: "docs/domain/inventory-state-machine.md"
        purpose: "库存状态机说明（含 stock_status=2 预扣语义）"
    code_scope:
      - "internal/inventory/client/**"
  removed: {}

## 需求补丁（合并到 req.md）
patch_to_req: |
  在 "外部依赖" 章节追加：
  - 调用 inventory.QueryStock 查询库存
  - 关注 stock_status=2（预扣状态）时的兜底分支：需要等待 30s 后重试一次，
    仍为预扣则返回 ErrStockReserved

## 期望改进点（下游执行时优先覆盖）
expected_improvements:
  - "spec.md 外部依赖章节必须完整列出 inventory.QueryStock 的请求/响应字段"
  - "spec.md 异常路径章节覆盖 stock_status=2 的处理策略"
  - "plan.md TDD 计划中补充 inventory client 的单元测试任务"
```

### 3.3 命名与持久化

- 文件名：`attempt-${N}.md`，N 与 `meta.yaml.attempts` 保持一致
- 归档：**永不删除**，作为审计轨迹
- 下游读取：重新进入的子 skill 只读取**最新一份**（N 最大者）
- 多份存在时：说明经历过多次回退重入，`history[]` 字段可展示全轨迹
