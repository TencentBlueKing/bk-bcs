---
name: tapd-story-specify
slug: tapd-story-specify
version: 5.0.0
description: |
  迭代执行流水线技术澄清与规范生成阶段。通过 tech-lead 分两段执行：
  技术澄清段审查需求并产出澄清结论，specify 段生成 spec.md。
  两段共享统一的 questions.md 代问循环机制——遇到无法自答的问题以
  blocked_on_questions 退出，由外部代问后以 answer 指令重入继续。
  支持 blocked_on_questions 代问重入与跨阶段回退重入。
---

# 技术澄清与规范生成

## 定位

技术澄清处于**需求文档定稿**与**进入开发**之间的关键衔接点。

> **需求澄清**关注"做什么"；**技术澄清**关注"怎么做"和"能不能做"。

## 前置条件

- 当前需求 `meta.yaml.phase` 为 `initialized`
- `${WORKDIR}/req.md` 存在（由调用方注入或 pipeline 拉取）
- pipeline 主编排已按 `references/context-and-meta-template.md` 生成 `context.md` 与 `meta.yaml`

## 输入

- `${WORKDIR}/meta.yaml`（读取 `attempts` / `context_revision`）
- 当前需求 ID（由 pipeline 主编排传入）
- `${WORKDIR}/context.md`
- `${WORKDIR}/iteration-patches/attempt-${N}.md`（若为回退重入，必读）

## 执行流程

本子 skill 由两段组成，通过 `tech-lead` 隔离执行，共享统一的
`questions.md` 代问循环机制。

### 1. 重入判定（pipeline 主编排执行）

每次进入本子 skill 时，按文件系统状态判定当前段位：

| 条件 | 段位 | 动作 |
|------|------|------|
| `questions.md` 不存在 | 技术澄清段（首次） | 执行技术澄清段 |
| `questions.md` 有 `[open]` + `spec.md` 不存在 | 技术澄清段（重入） | 执行技术澄清段 |
| `questions.md` 有 `[open]` + `spec.md` 存在 | specify 段（重入） | 执行 specify 段 |
| `questions.md` 全非 open + `spec.md` 不存在 | specify 段（首次） | 执行 specify 段 |
| `questions.md` 全非 open + `spec.md` 存在 | 质量验证 | 直接进入质量验证 |

> spec.md 不存在 → 还未进入 specify 段，统一按技术澄清段重入处理。

### 2. 技术澄清段（tech-lead）

pipeline 主编排按 `../references/subagent-prompt-template.md` §2.0 **Clarify 模板**渲染
SUBAGENT_PROMPT，通过 `Task(subagent_name="tech-lead", ...)` 拉起 subagent。

渲染填充：`${ID}` / `${VERSION}` / `${WORK_DIR}` / `attempts` / `round`；
若 `attempts > 1`，附上 `iteration-patches/attempt-${attempts}.md` 摘要。

**subagent 职责**：
1. 通读 req.md，评估技术复杂度（简单/中等/复杂）
2. 按复杂度选择澄清维度（参考 `references/technical-clarification-guide.md`）
3. 对照 context.md 白名单自答（追加 `resolved_by_doc` 条目到 questions.md）
4. 无法自答的问题追加为 `open` 条目到 questions.md
5. 所有可自答的结论写入 req.md "技术澄清"章节（格式参考 `references/technical-clarification-template.md`）
6. 读取 questions.md 已有的 `answered` 条目（重入场景），将答复融入 req.md

**跳过条件**（由 subagent 在执行中自行判定）：
- 需求极简（配置变更/文案修改/简单 Bug 修复）且无技术风险
- 追加 `[dropped]` 条目 + req.md 标注"技术审查通过"
- 以 `status=ok` 返回

**回传处理**：

| status | 处理 |
|--------|------|
| `ok` | 技术澄清结论已写入 req.md → 进入 specify 段 |
| `blocked_on_questions` | questions.md 有新 [open] → pipeline 退出（blocked 卡点）|
| `fail` | 写 last_failure → pipeline 退出 |

**代问上限**：技术澄清段单 attempt 内最多 3 round。

### 3. Specify 段（tech-lead）

pipeline 主编排按 `../references/subagent-prompt-template.md` §2.1 **Specify 模板**渲染
SUBAGENT_PROMPT，通过 `Task(subagent_name="tech-lead", ...)` 拉起 subagent。

渲染填充：同技术澄清段。
若 `attempts > 1`，附上 `iteration-patches/attempt-${attempts}.md` 摘要。

> subagent 内的 speckit-specify skill、代问指令（追加 questions.md open 条目 →
> blocked_on_questions 返回）均在 subagent-prompt-template.md §2.1 中定义。

**回传处理**：

| status | 处理 |
|--------|------|
| `ok` | spec.md 已生成 → 进入质量验证 |
| `blocked_on_questions` | questions.md 有新 [open] → pipeline 退出（blocked 卡点）|
| `fail` | 写 last_failure → pipeline 退出 |

**代问上限**：specify 段单 attempt 内最多 5 round。

### 4. 质量验证（主编排内联）

对 `spec.md` 与 `req.md` 做覆盖性对比，并执行内容门禁：

| 检查项 | 通过标准 |
|--------|----------|
| 用户场景 | 至少一个主流程场景，且包含可验证的验收场景 |
| 功能需求 | 每项需求可测试、无歧义，并覆盖 req.md 的核心业务目标 |
| 成功标准 | 至少一项可量化或可观察的验证标准 |
| 边界与依赖 | 边界场景、范围、依赖与假设已记录；不适用项必须说明理由 |
| 澄清状态 | 不遗留未处理的 `[NEEDS CLARIFICATION]` 标记；无法确定的信息已写入 questions.md 的 open 条目 |
| 技术类型 | subagent 回传的 `tech_type` 为 backend / frontend / fullstack 之一 |

主编排按项目的实际 spec 模板识别对应章节，不要求固定的中英文标题。

- **OK** → 更新状态
- **遗漏或矛盾** → `status=fail`，在 issue 中列出缺失项，走回退重入（不在本 attempt 内简单重跑）

### 5. 更新状态

phase → `tech-clarified`；history 追加成功记录。
若回传 JSON 中含 `tech_type` 字段（值为 backend / frontend / fullstack），将其写入 `meta.yaml.tech_type`。

## 代问循环协议（统一）

无论技术澄清段还是 specify 段产生 `blocked_on_questions`，对外表现一致：

1. pipeline 以 blocked 卡点退出
2. 退出报告中展示 questions.md 中 [open] 条目数
3. 调用方（runner/用户）修改 questions.md（[open] → [answered]）
4. 以 `action=answer` 再次调起 pipeline
5. pipeline 重入，按 §1 重入判定表确定段位，继续执行

用户无需区分问题来自技术澄清还是 spec 生成。

## round 计数规则

技术澄清段和 specify 段各自维护 round 计数：

- 技术澄清段 round：从 1 起记，每次代问重入 +1，上限 3
- specify 段 round：技术澄清完成后重置为 1，每次代问重入 +1，上限 5
- 区分方式：`meta.yaml.specify_attempts[-1].rounds` 中标注 `stage: clarify | specify`

超限后按 `../references/reentry-protocol.md` §1.2 升级为回退重入。

## 可重入约定

1. **代问重入**（pipeline 内部）：round +1（attempts 不变）
2. **回退重入**：调用方按 `state-mutation-guide.md` §2 提供 retry 指令与
   `iteration-patches/attempt-*.md`；pipeline 仅在 `attempts < max_attempts` 时递增 attempts
3. **每次调用前**：重新生成 `context.md`，`context_revision` +1

子 skill 在执行前必须：

- 读取最新 `attempt-*.md`（若有），追加重入增量段（subagent-prompt-template.md §3）
- `spec.md` 覆盖前由 subagent 做快照 `spec.md.prev-attempt${N}-round${R}`

幂等性、产物保留、process.log 只追加等通用约束见 `../references/reentry-protocol.md`。

## 信息不足回退（回退重入入口）

后续子 skill 发现 spec.md 不充分时返回 `spec_insufficient`，
调用方把 phase 回退至 `initialized` 并写 `attempt-*.md`，再次调起时重新执行本子 skill。

## 参考资料

- SUBAGENT_PROMPT（clarify 模板）：`../references/subagent-prompt-template.md` §2.0
- SUBAGENT_PROMPT（specify 模板 + 代问指令）：`../references/subagent-prompt-template.md` §2.1
- 技术澄清维度与最佳实践：`references/technical-clarification-guide.md`
- 技术澄清章节模板：`references/technical-clarification-template.md`
- 问题文件格式：`../references/questions-md-template.md`
- 上下文与元数据模板：`../references/context-and-meta-template.md`
- 可重入协议：`../references/reentry-protocol.md`

## 产出

- `${WORKDIR}/req.md` — 原始需求（含技术澄清章节）
- `${WORKDIR}/questions.md` — 澄清问题与答复（全历史）
- `${WORKDIR}/spec.md` — 技术规范
- `${WORKDIR}/process.log` — stream-json 流式日志
- `${WORKDIR}/meta.yaml` — 更新后的 attempts / round / history
- 需求 phase 为 `tech-clarified`
