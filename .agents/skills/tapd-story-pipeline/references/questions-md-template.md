# Questions.md 模板与代问循环协议

## 目录

- [§1. 为什么需要 questions.md](#1-为什么需要-questionsmd)
- [§2. 四态状态机](#2-四态状态机)
  - [2.1 状态定义](#21-状态定义)
  - [2.2 合法状态转移](#22-合法状态转移)
  - [2.3 状态一致性规则](#23-状态一致性规则恢复流程依赖)
- [§3. 条目格式](#3-条目格式)
  - [3.1 结构约束](#31-结构约束grep-友好)
  - [3.2 完整条目模板](#32-完整条目模板)
  - [3.3 字段取值](#33-字段取值)
  - [3.4 追加记录语义](#34-追加记录语义避免改写历史)
  - [3.5 示例](#35-示例完整文件)
- [§4. 代问循环协议](#4-代问循环协议)
  - [4.1 时序图](#41-时序图subagent--主会话--用户)
  - [4.2 终止条件](#42-代问循环的终止条件)
  - [4.3 与 subagent 回传 JSON 的对齐](#43-与-subagent-回传-json-的对齐)
  - [4.4 编号分配协议](#44-编号分配协议)

---

`questions.md` 是**技术澄清阶段**（`tapd-story-specify`）用于承载"subagent ↔ 主会话 ↔ 用户"
三方对话的落地文件。所有澄清类问题与答复都以**条目**形式追加到该文件，作为
`req.md` / `spec.md` 之外的第三类可追溯资产。

文件路径：`${WORKDIR}/questions.md`

---

## 1. 为什么需要 questions.md

技术澄清阶段的问题有两个来源：

1. **主会话发起**：pipeline 主编排在技术澄清流程中整理的问题清单
2. **subagent 发起**：`/speckit.specify` 执行期间 speckit 自身需要澄清的问题

两类问题如果分散在对话流里，会让主会话上下文持续膨胀，也无法被 subagent 重新启动时"再次看到"。
把它们统一落到一份**结构化可追溯、可 grep、可断点恢复**的文件里，是本机制的核心价值：

- **subagent 发现问题 → 追加条目 → 以 `status=blocked_on_questions` 返回**：不猜测、不扩展白名单
- **主会话代问 → 用户答复 → 回写条目状态**：对话回到主会话唯一入口，subagent 不直接问用户
- **重启 subagent**：再次读取 questions.md 就能拿到最新答复，继续生成 spec.md

---

## 2. 四态状态机

### 2.1 状态定义

| 状态 | 含义 | 下游（subagent / speckit.specify）处理 |
|------|------|-----------------------------------|
| `open` | 问题已提出，等待用户答复 | **阻塞性**：任一条 open 存在则 subagent 不得继续生成 spec.md；必须返回 blocked_on_questions |
| `answered` | 用户已答复 | 正常参与 spec.md 生成；pipeline 主编排把答复同步到 `req.md` 的"技术澄清"章节 |
| `resolved_by_doc` | 通过 `context.md` 白名单内文档自答 | 同 answered，不阻塞；但标注来源为文档，无需用户确认 |
| `dropped` | 后续判定非必要，显式放弃 | 不参与 spec.md 生成；保留作为审计痕迹（避免下一轮再被提同样问题）|

### 2.2 合法状态转移

```
              用户答复              判定非必要
   open ────────────────► answered ─────────────► dropped
     │                                               ▲
     │ 自答（文档查得）                                │
     └────────────────► resolved_by_doc ──────────────┘
                               │
                               └─ 用户覆盖答案 ──► answered（追加补充答复后升级）
```

**说明**：

- 状态一旦进入 `answered` / `resolved_by_doc` / `dropped`，**不得回退**到 `open`。
  若答复后被发现不正确需要重新回答，追加一条**新条目**引用旧条目（见 §3.4），不要改旧条目状态。
- `resolved_by_doc` 升级为 `answered` 的场景：用户看到 subagent 自答后提出补充意见，此时
  追加补充答复并把状态改为 `answered`（文档来源仍保留在条目里）。

### 2.3 状态一致性规则（恢复流程依赖）

恢复流程（`tapd-iteration-init`）读取 questions.md 时按以下规则检查：

- 任意 `open` 条目存在 → 本需求处于 blocked_on_questions 状态，恢复后第一步是继续代问循环
- 所有条目均非 `open` → 可以直接重启 subagent 继续 `/speckit.specify`

---

## 3. 条目格式

### 3.1 结构约束（grep 友好）

每个条目都是一个二级标题（`## Q${N}`）起始的段落，标题行**必须**严格遵循以下格式：

```
## Q${N} [${status}] — 来源：${source}
```

pipeline 主编排（以及外部调用方）据此以 `grep '^## Q.*\[open\]'` 精确锁定未决条目，**不引入任何解析歧义**。

### 3.2 完整条目模板

```markdown
## Q${N} [${status}] — 来源：${source}
**问题**：${question_text}
**影响**：${impact_description}（标注阻塞性：阻塞 | 非阻塞）
**建议候选**：
- A. ${candidate_a}（推荐理由：...）
- B. ${candidate_b}
**提出方**：${proposer} / attempt=${N} / round=${R} / ts=${ISO8601}
**答复**：${answer_text}              # 仅 answered / resolved_by_doc 填充
**答复方**：${answerer} / ts=${ISO8601}  # 仅 answered / resolved_by_doc 填充
**文档来源**：${doc_path}              # 仅 resolved_by_doc 填充
**放弃理由**：${drop_reason}           # 仅 dropped 填充
```

### 3.3 字段取值

| 字段 | 取值 |
|------|------|
| `${N}` | 自增整数，从 1 起记；**永不复用**（dropped 的编号也不回收）|
| `${status}` | `open` / `answered` / `resolved_by_doc` / `dropped` 之一 |
| `${source}` | `技术澄清`（主会话发起）/ `subagent(speckit.specify)`（subagent 发起）|
| `${proposer}` | `主会话` / `subagent(speckit.specify)` |
| `${answerer}` | `user via 主会话` / `subagent(自答)` |
| `attempt` / `round` | 来自 `meta.yaml` 当前值，标注问题是在哪一轮提出的 |
| `ts` | ISO 8601 格式时间戳，含时区（如 `2026-05-12T16:42:00+08:00`）|

### 3.4 追加记录语义（避免改写历史）

条目一旦写入不做修改性编辑，下列场景全部通过**追加新条目**表达：

| 场景 | 处理 |
|------|------|
| 用户对已 answered 条目提供补充/修正答复 | 追加新条目引用旧编号：`**关联**：Q3 的补充/修正` |
| 已 answered 条目在后续发现矛盾 | 追加新 open 条目 + 在标题/影响字段引用原条目 |
| 需求范围缩减导致原问题不再相关 | 追加 `## Qmeta [annotation]` 说明；原条目状态**不**改 |

> 这条"不改历史"规则让 questions.md 成为完整的审计记录，同时让 grep `[open]` 永远
> 只返回真正待处理的条目。

### 3.5 示例（完整文件）

```markdown
# Clarification Questions — Story 1234567

## Q1 [answered] — 来源：技术澄清
**问题**：订单过期时间默认值？
**影响**：影响 spec.md 中"订单生命周期"章节；非阻塞。
**建议候选**：
- A. 30 分钟（推荐：与现有订单模块一致）
- B. 60 分钟
**提出方**：主会话 / attempt=1 / round=1 / ts=2026-05-12T16:30:00+08:00
**答复**：统一 30 分钟，允许通过配置 override。
**答复方**：user via 主会话 / ts=2026-05-12T16:45:00+08:00

## Q2 [resolved_by_doc] — 来源：subagent(speckit.specify)
**问题**：订单号生成规则？
**影响**：影响 spec.md 中"订单创建流程"步骤 2；阻塞。
**建议候选**：
- A. 雪花算法（文档已定义）
**提出方**：subagent(speckit.specify) / attempt=1 / round=1 / ts=2026-05-12T16:48:00+08:00
**答复**：使用雪花算法，workerId 由服务启动时从环境变量读取，数据中心 ID 固定为 1。
**答复方**：subagent(自答) / ts=2026-05-12T16:48:30+08:00
**文档来源**：docs/domain/order-id.md

## Q3 [open] — 来源：subagent(speckit.specify)
**问题**：库存查询 `inventory.QueryStock` 返回字段 `stock_status=2` 的业务含义？
**影响**：决定 spec.md 中"异常分支"的定义；阻塞。
**建议候选**：
- A. 表示预扣，不计入可售库存（推荐：与领域模型一致）
- B. 表示冻结，计入可售但不可下单
**提出方**：subagent(speckit.specify) / attempt=1 / round=2 / ts=2026-05-12T17:10:00+08:00

## Q4 [dropped] — 来源：技术澄清
**问题**：是否需要在订单创建时发送站内信？
**影响**：影响 spec.md "通知"章节；非阻塞。
**建议候选**：无
**提出方**：主会话 / attempt=1 / round=1 / ts=2026-05-12T16:35:00+08:00
**放弃理由**：本迭代范围不含通知模块（与用户确认于 ts=2026-05-12T16:36:00+08:00）。

## Q5 [answered] — 来源：技术澄清
**问题**：Q1 的补充/修正
**影响**：覆盖 Q1 原答复。
**建议候选**：无
**提出方**：主会话 / attempt=1 / round=2 / ts=2026-05-12T17:20:00+08:00
**答复**：30 分钟的默认值调整为 45 分钟（运营需求变更）。
**答复方**：user via 主会话 / ts=2026-05-12T17:22:00+08:00
**关联**：Q1 的补充/修正
```

---

## 4. 代问循环协议

### 4.1 时序图（speckit-executor-agent / 主编排 / 用户）

```
pipeline 主编排                      speckit-executor-agent               用户
     │                                      │                              │
     │ 1. 技术澄清段                        │                              │
     │ ───────────────────────────────────► │                              │
     │                                      │ 1.1 读 req.md + context.md    │
     │                                      │ 1.2 技术可行性审查             │
     │                                      │    - 可自答 → resolved_by_doc │
     │                                      │    - 不可自答 → open 条目      │
     │                                      │    → status=blocked / ok 返回  │
     │◄─────────────────────────────────────│                              │
     │                                      │                              │
     │ 2. 收到 blocked_on_questions         │                              │
     │    → pipeline 退出（blocked 卡点）    │                              │
     │    → 调用方展示 [open] 条目给用户     │                              │
     │ ────────────────────────────────────────────────────────────────────►│
     │◄────────────────────────────────────────────────────────────────────│
     │   - 用户答复 → 条目 open → answered   │                              │
     │   - action=answer 再次调起 pipeline   │                              │
     │                                      │                              │
     │ 3. 重入判定 → 技术澄清段（重入）      │                              │
     │ ───────────────────────────────────► │                              │
     │                                      │ 3.1 读已 answered 条目         │
     │                                      │     融入 req.md 技术澄清章节  │
     │                                      │ 3.2 检查 DoR → 通过 → ok 返回 │
     │◄─────────────────────────────────────│                              │
     │                                      │                              │
     │ 4. 技术澄清完成 → specify 段         │                              │
     │ ───────────────────────────────────► │                              │
     │                                      │ 4.1 读 questions.md(非 open)   │
     │                                      │    + req.md + context.md       │
     │                                      │ 4.2 调用 /speckit.specify      │
     │                                      │    - 可自答 → resolved_by_doc  │
     │                                      │    - 不可自答 → open 条目      │
     │                                      │    → status=blocked / ok 返回  │
     │◄─────────────────────────────────────│                              │
     │                                      │                              │
     │ 5. 若 blocked → 同步骤 2 代问循环     │                              │
     │    若 ok → spec.md 已生成             │                              │
     │                                      │                              │
     │ 6. 质量验证 → 推进 phase=tech-clarified│                             │
```

### 4.2 代问循环的终止条件

**正常终止**：subagent 返回 `status=ok`，`produced` 含 `spec.md`；pipeline 主编排推进 phase。

**异常终止**：代问循环 round 数达到上限（默认 5），仍有 open 问题未消除。此时 pipeline 主编排：

1. 把所有 open 条目的 question_text 汇总为 `iteration-patches/attempt-${N+1}.md` 的
   `patch_to_req`（建议用户在 req.md 中补充）
2. 若 `attempts < max_attempts`，由 pipeline 主编排执行回退重入并写入
   `meta.yaml.attempts +1`；否则写
   `last_failure.message="attempts exhausted (3/3): 代问超限"`、type=semantic 后退出
3. 虽然下游阶段并未失败，但这种情况通常意味着**需求本身不具备澄清收敛条件**——
   pipeline 退出后由调用方（runner / 用户）按 `state-mutation-guide.md` §2 卡点 3 决定是否拆分需求或调整范围

> round 上限 5 是经验值，比非阻塞性问题的通常对话轮数大 2~3 倍，足够应对"用户需要查阅
> 资料后再回答"的场景。实际值可在主 SKILL.md 中配置覆盖。

### 4.3 与 subagent 回传 JSON 的对齐

subagent 返回 JSON 时必须填充 `questions_delta` 字段（仅本段 subagent 实际写入的条目）：

```jsonc
{
  "status": "blocked_on_questions",
  "questions_delta": {
    "added_open":   ["Q3", "Q7"],     // 本轮由 subagent 新增的 open 条目编号
    "self_resolved": ["Q2"]           // 本轮由 subagent 自答（resolved_by_doc）的编号
  }
}
```

主会话读取该字段后，定向到 questions.md 对应条目展开代问，不需要全文扫描。

### 4.4 编号分配协议

- **主会话新增条目**：主会话负责维护"最大编号"，追加条目时取 `max(现有编号) + 1`
- **subagent 新增条目**：subagent 追加前先 grep 当前文件取最大编号，再 +1 追加；
  并在回传 JSON 的 `questions_delta.added_open` / `self_resolved` 中声明自己写入的编号
- **并发写入防护**：当前架构下 subagent 运行时主会话不会并行写入（代问循环是**串行**的，
  subagent 返回后主会话才代问），因此无需加锁

---

## 5. 与 req.md 的同步规则

### 5.1 谁是权威？

- `questions.md` 是**对话历史与审计记录**（全量、不改历史）
- `req.md` 的"技术澄清"章节是**spec.md 的直接输入**（仅保留最新有效结论）

两者职责分工：questions.md 负责"为什么是这样"；req.md 负责"当前结论是什么"。

### 5.2 同步时机

| 时机 | 操作 |
|------|------|
| 主会话收到用户答复并置 `answered` | 同步把答复融入 `req.md` 的"技术澄清"章节（覆盖性写入相应小节）|
| subagent 追加 `resolved_by_doc` 条目 | subagent 在回传 JSON 中声明；主会话收到后同步到 req.md |
| 追加修正答复（如 Q5 修正 Q1） | 覆盖 req.md 中对应小节，老答案不保留（req.md 只保留最新）|
| 条目进入 `dropped` | req.md 中不写入（若之前误写，删除对应小节）|

### 5.3 覆盖性写入的幂等保证

req.md 的"技术澄清"章节按问题维度组织小节（如"## 外部依赖" / "## 订单生命周期"），
每次同步时**覆盖对应小节**而非追加。具体结构参考
`../tapd-story-specify/references/technical-clarification-template.md`。
