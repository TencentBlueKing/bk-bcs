# SUBAGENT_PROMPT 骨架与各阶段模板

## 目录

- [SUBAGENT\_PROMPT 骨架与各阶段模板](#subagent_prompt-骨架与各阶段模板)
  - [目录](#目录)
  - [1. 通用骨架](#1-通用骨架)
    - [1.1 占位符清单](#11-占位符清单)
    - [1.2 通用前缀（所有阶段必须出现）](#12-通用前缀所有阶段必须出现)
    - [1.3 通用回传 JSON Schema（严格契约）](#13-通用回传-json-schema严格契约)
  - [2. 阶段差异化模板](#2-阶段差异化模板)
    - [2.0 Clarify 模板（tapd-story-specify 技术澄清段）](#20-clarify-模板tapd-story-specify-技术澄清段)
    - [2.1 Specify 模板（tapd-story-specify）](#21-specify-模板tapd-story-specify)
    - [2.2 Plan 模板（tapd-story-plan）](#22-plan-模板tapd-story-plan)
    - [2.3 Tasks-Generate 模板（tapd-story-tasks 第一段）](#23-tasks-generate-模板tapd-story-tasks-第一段)
    - [2.4 Tasks-Analyze 模板（tapd-story-tasks 第二段）](#24-tasks-analyze-模板tapd-story-tasks-第二段)
    - [2.5 Implement 模板（tapd-story-implement）](#25-implement-模板tapd-story-implement)
    - [2.6 Validate-Arch 模板（tapd-story-validate 段 #1）](#26-validate-arch-模板tapd-story-validate-段-1)
    - [2.7 Validate-Security 模板（tapd-story-validate 段 #2）](#27-validate-security-模板tapd-story-validate-段-2)
    - [2.8 Validate-CodeReview 模板（tapd-story-validate 段 #3）](#28-validate-codereview-模板tapd-story-validate-段-3)
    - [2.9 Validate-Fix 模板（tapd-story-validate 修复段，仅当需要）](#29-validate-fix-模板tapd-story-validate-修复段仅当需要)
    - [2.10 Validate-Test 模板（tapd-story-validate 段 #4）](#210-validate-test-模板tapd-story-validate-段-4)
  - [3. 重入增量段（attempts \> 1 或 round \> 1 时追加）](#3-重入增量段attempts--1-或-round--1-时追加)
    - [3.1 通用回退重入增量（attempts \> 1）](#31-通用回退重入增量attempts--1)
    - [3.2 同 attempt 内 round 重试增量（round \> 1）](#32-同-attempt-内-round-重试增量round--1)
    - [3.3 Specify 阶段代问重入增量（round \> 1 且来源是 blocked\_on\_questions）](#33-specify-阶段代问重入增量round--1-且来源是-blocked_on_questions)
  - [4. 渲染流程总结](#4-渲染流程总结)
  - [5. 验收清单（pipeline 主编排渲染前自检）](#5-验收清单pipeline-主编排渲染前自检)

---

本文件提供 pipeline 主编排在 `Task(subagent_name=<AGENT>, prompt=<PROMPT>)` 调用前用于
**渲染** SUBAGENT_PROMPT 的全部模板。

**阶段 → subagent_name 映射**（主编排按阶段选择执行 agent）：

| 阶段 | subagent_name | 说明 |
|------|---------------|------|
| clarify / specify | `tech-lead` | 技术澄清 + spec 生成 |
| plan | `tech-lead` | 开发计划制定 |
| tasks-generate / tasks-analyze | `tech-lead` | 任务拆解与分析 |
| implement | `backend-developer` 或 `frontend-developer` | 按 `meta.yaml.tech_type` 路由；fullstack 串行：后端先 |
| validate-arch | `tech-lead` | 架构合规评审 |
| validate-security | `code-reviewer` | 安全审查（经 use_skill("code-review") 加载统一 CR 方法论） |
| validate-codereview | `code-reviewer` | 代码质量审查（经 use_skill("code-review") 加载统一 CR 方法论，见 §2.8） |
| validate-test | `qa-engineer` | 测试覆盖审查（新增，见 §2.10） |
| validate-fix | `backend-developer` 或 `frontend-developer` | 按受影响文件路径路由；fullstack 串行：后端先 |

模板分两层：

- **§1 通用骨架**：所有阶段共用的指令前缀 + 行为约束 + 回传 JSON schema
- **§2 阶段差异化模板**：每个子 skill 阶段的专属指令

pipeline 主编排渲染流程：
1. 渲染 §1 通用骨架（按占位符表填充）
2. 按当前阶段从 §2 取对应模板，追加到骨架末尾
3. 若 `meta.yaml.attempts > 1` 或 round > 1，再追加 §3 "重入增量段"
4. 通过 `Task()` 发起调用

---

## 1. 通用骨架

### 1.1 占位符清单

| 占位符 | 来源 | 示例 |
|--------|------|------|
| `${ID}` | 调用方传入 / `meta.yaml.id` | `1234567` |
| `${VERSION}` | 调用方传入 / 从 `${WORK_DIR}` 路径解析 | `v0.9.x` |
| `${WORK_DIR}` | 调用方传入 / 默认规则 | `specs/v0.9.x/1234567` |
| `${ATTEMPT}` | `meta.yaml.attempts` | `2` |
| `${ROUND}` | 子 skill 维护的当前 round | `1` |
| `${TS}` | 调用时的 ISO 8601 时间戳 | `2026-05-13T14:23:11+08:00` |
| `${STAGE}` | 当前阶段名 | `specify` / `plan` / ... |
| `${REENTRY_CONTEXT}` | 重入时由 pipeline 填充修正摘要（首次为空） | 见 §3 |
| `${LOG_FILE}` | 当前阶段日志文件路径 | `${WORK_DIR}/process.log` 或 `${WORK_DIR}/process-validate-arch.log` |
| `${IMPLEMENT_BASELINE_COMMIT}` | `meta.yaml.implement_baseline_commit` | `a1b2c3d` |

**日志文件选择规则**（pipeline 主编排填充 `${LOG_FILE}`）：

| 阶段 | LOG_FILE |
|------|----------|
| clarify / specify / plan / tasks-generate / tasks-analyze / implement | `${WORK_DIR}/process.log` |
| validate-arch | `${WORK_DIR}/process-validate-arch.log` |
| validate-security | `${WORK_DIR}/process-validate-security.log` |
| validate-codereview | `${WORK_DIR}/process-validate-codereview.log` |
| validate-test | `${WORK_DIR}/process-validate-test.log` |
| validate-fix | `${WORK_DIR}/process-validate-fix.log` |

### 1.2 通用前缀（所有阶段必须出现）

```
你的唯一职责是在隔离上下文中完成一次本阶段的工作并返回结构化 JSON 摘要。

[基础信息]
- 需求 ID：${ID}
- 工作目录 work_dir：${WORK_DIR}
- 当前阶段 stage：${STAGE}
- 当前 attempt：${ATTEMPT}
- 当前 round：${ROUND}
- 调用时间：${TS}

[spec-cost-marker] work_dir=${WORK_DIR} stage=${STAGE} attempt=${ATTEMPT} round=${ROUND} ts=${TS}

[返回JSON定义]
- 来自`skills/tapd-story-pipeline/references/subagent-prompt-template.md` §1.3 通用回传 JSON Schema（严格契约）

[上下文白名单 — 强制约束]
必须按 ${WORK_DIR}/context.md 的白名单工作：
  - 仅读取 Source artifacts + Project background 中列出的文件，不扩展白名单
  - 仅当本模板的“方法论补充”步骤明确指定时，允许读取对应 agent skill 的指定 reference 文件及章节；不得读取其他 skill reference
  - validate-* 阶段可使用 `git diff ${IMPLEMENT_BASELINE_COMMIT}` 确定 Code scope 内的变更范围；不得据此读取白名单外文件内容
  - 写代码仅限 Code scope 范围；越界写入视为失败
  - 若白名单不足以完成本阶段任务：
      - specify 阶段：追加 questions.md open 条目，以 blocked_on_questions 返回
      - 其他阶段：以 status=fail 返回，在 issue 中说明缺失文档

[日志落盘]
  - 日志文件：${LOG_FILE}
  - 在阶段开始前先 echo banner 到日志文件：
      ===== [stage=${STAGE} attempt=${ATTEMPT} round=${ROUND} ts=${TS}] =====
  - 阶段执行过程中的关键事件（错误、警告、产物落盘）由 subagent 追加到 ${LOG_FILE}
  - 严禁把日志内容回传给调用方

[隔离边界]
- 不进行 git 分支操作
- 不修改 ${WORK_DIR} 之外的产物文件（implement除外）
- 不直接询问用户
- 完成后立即返回结构化 JSON
```

### 1.3 通用回传 JSON Schema（严格契约）

subagent 完成后**必须**且**仅**返回以下 JSON：

```jsonc
{
  "ok": false,

  // 三态状态：ok（成功推进）/ blocked_on_questions（仅 specify 可用）/ fail
  "status": "ok | blocked_on_questions | fail",

  // 阶段标识；validate 阶段细分为 validate-arch / validate-security / validate-codereview / validate-test / validate-fix
  "phase": "specify | plan | tasks-generate | tasks-analyze | implement | validate-arch | validate-security | validate-codereview | validate-test | validate-fix",
  "attempt": 1,
  "round": 1,
  // 本轮 subagent 新生成或更新的产物路径（相对仓库根）
  "produced": ["specs/v0.9.x/1234567/spec.md"],
  // ===== 阶段差异化字段（仅特定阶段填写，其他阶段可省略）=====

  // 仅 specify 阶段
  "questions_delta": {
    "added_open":   ["Q编号数组"],
    "self_resolved":["Q编号数组"]
  },

  // 仅 specify 阶段回传；pipeline 主编排将其写入 meta.yaml
  "tech_type": "backend | frontend | fullstack",

  // 仅 plan / tasks-analyze / validate-* 阶段
  "compliance": {
    "verdict": "pass | needs_fix | spec_insufficient | plan_insufficient | LGTM",
    "report": "{phase}-report.md"   // 指向已落盘的报告文件，pipeline 主编排按需读取
  },

  // 仅 implement / validate-fix 阶段填写
  "tests": {
    "unit":        "passed | failed | skipped",
    "integration": "passed | failed | skipped",
    "e2e":         "passed | failed | skipped"
  },

  // 通用：失败摘要（status=fail 时必填，其他状态留空）
  "issue": ""
}
```

**填写规则**：
- `produced` 列出本轮实际写入的文件
- `compliance.report` 指向落盘的报告文件，findings从报告文件按需提取
- `tests` 仅 implement / validate-fix 阶段填写，其他阶段省略
- `issue` ≤200 字

---

## 2. 阶段差异化模板

> 所有阶段统一结构：`[阶段任务] → [执行命令] → [产物落盘] → [完成后自检] → [返回 JSON 要求]`

### 2.0 Clarify 模板（tapd-story-specify 技术澄清段）

**适用阶段**：`STAGE=clarify`
**phase 推进**：不推进（技术澄清完成后由主编排进入 specify 段）

```
[阶段任务]
基于 ${WORK_DIR}/req.md 的需求描述，从技术视角审查需求并输出技术澄清结论。
参考 skills/tapd-story-specify/references/technical-clarification-guide.md 的澄清维度。
参考 skills/tapd-story-specify/references/technical-clarification-template.md 的输出格式。
上下文白名单见 ${WORK_DIR}/context.md。

[执行步骤]
1. 通读 req.md，评估技术复杂度（简单/中等/复杂）
2. 按复杂度选择澄清维度（简单→快速审查；中等→标准澄清；复杂→深度澄清）
3. 对照 context.md 白名单内文档自答：
   - 能从文档中找到答案 → 追加 resolved_by_doc 条目到 questions.md
   - 将自答结论写入 req.md 的"技术澄清"章节
4. 无法自答的问题 → 追加 open 条目到 questions.md
5. 读取 questions.md 已有的 answered 条目（代问重入场景），将答复融入 req.md 技术澄清章节

[跳过条件]
若需求满足以下全部条件，跳过详细澄清：
- 复杂度=简单（纯文案/配置变更/简单 Bug 修复）
- 无技术风险（不涉及架构变更、外部依赖、数据迁移）
跳过时：
- 追加一条 [dropped] 条目到 questions.md："技术审查通过，无需额外技术澄清"
- req.md 追加简化技术审查结论
- 以 status=ok 返回

[DoR 检查]
技术就绪检查按复杂度分级：
| 复杂度 | 必须满足 |
|--------|---------|
| 简单 | 无技术风险 |
| 中等 | + 技术方案明确 + 外部依赖已识别 + 测试策略已定义 |
| 复杂 | + 部署约束明确 + 回滚方案就绪 |

全部 DoR 项可通过 context.md 自答或已有 answered 条目满足 → ok
存在无法满足的 DoR 项 → 追加 open 条目 → blocked_on_questions

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：创建或追加 questions.md（若 ${WORK_DIR}/questions.md 不存在则
          初始化为 "# Clarification Questions — Story ${ID}"）
  Step 3：读取
          skills/tech-lead/references/requirement-analysis.md 的“1.3 NFR 清单”和
          “三、澄清问题模板”，仅对当前复杂度适用的维度生成澄清结论或问题。
  Step 4：直接执行：通读 req.md + context.md 白名单文档，
          按 [执行步骤] 完成技术澄清；结论写入 req.md 技术澄清章节，问题追加到 questions.md

[完成后自检]
1) 确认 req.md 技术澄清章节已更新（或标注跳过）
2) 确认 questions.md 格式合规（grep '^## Q' 验证）

[返回 JSON 要求]
- phase: "clarify"
- 必须填写 questions_delta（即使为空）
- ok 时 produced 含 req.md
- blocked_on_questions 时 produced 可为空
```

### 2.1 Specify 模板（tapd-story-specify）

**适用阶段**：`STAGE=specify`
**phase 推进**：`initialized → tech-clarified`

```
[阶段任务]
基于 ${WORK_DIR}/req.md 的需求描述与 ${WORK_DIR}/questions.md 中已 answered /
resolved_by_doc 的澄清结论，通过 use_skill 调用 speckit-specify 生成
${WORK_DIR}/spec.md。不创建新分支，工作目录为 ${WORK_DIR}。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：覆盖前快照——若 ${WORK_DIR}/spec.md 存在，cp 到
          ${WORK_DIR}/spec.md.prev-attempt${ATTEMPT}-round${ROUND}
  Step 3：调用 use_skill(command="speckit-specify") 加载 skill 内容到上下文
  Step 4：读取
          skills/tech-lead/references/requirement-analysis.md 的“1.3 NFR 清单”和
          “三、澄清问题模板”；将相关维度状态写入 spec.md，待澄清项写入 questions.md。
  Step 5：把 [ARGUMENTS] 段（见下）的全文作为 skill body 中 `$ARGUMENTS`
          占位符的实际值，执行 skill 内的指令
  Step 6：执行完成后落盘 ${WORK_DIR}/spec.md

[ARGUMENTS]
  基于 ${WORK_DIR}/req.md 生成规范文件。
  澄清结论见 ${WORK_DIR}/questions.md（如有，仅参考 answered / resolved_by_doc 条目）。
  上下文白名单见 ${WORK_DIR}/context.md。工作目录为 ${WORK_DIR}，不创建新分支。
  如果遇到问题，追加 open 条目到 questions.md 等待下一轮执行（问题模板参考 ./references/report-template.md）。
  执行结束时判定需求的技术类型（backend=仅涉及后端/接口；frontend=仅涉及前端/UI；
  fullstack=同时涉及前后端），在回传 JSON 的 `tech_type` 字段中填写。
  ${REENTRY_CONTEXT}

[代问指令 — 强制]
若推理中产生无法由白名单自答的新问题：
  - 追加 questions.md open 条目
  - 以 blocked_on_questions 返回，不继续生成 spec.md
若问题可由白名单文档自答：
  - 追加 resolved_by_doc 条目，继续完成

[完成后自检]
1) 确认 spec.md 已生成且可读

[返回 JSON 要求]
- phase: "specify"
- 必须填写 questions_delta（即使为空）
- status=ok 时必须填写 tech_type（backend | frontend | fullstack）
- 成功时 produced 含 spec.md
- blocked_on_questions 时 produced 可为空
```

### 2.2 Plan 模板（tapd-story-plan）

**适用阶段**：`STAGE=plan`
**phase 推进**：`tech-clarified → researched`

```
[阶段任务]
基于 ${WORK_DIR}/spec.md 通过 use_skill 调用 speckit-plan 构建开发计划，
随后就地做合规自检，产出 ${WORK_DIR}/plan-report.md。

[重入清理]
若存在 "${WORK_DIR}/plan-report.md"，使用delete_file删除。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：调用 use_skill(command="speckit-plan") 加载 skill 内容到上下文
  Step 3：读取
          skills/tech-lead/references/architecture.md 的“3.1 依赖方向检查清单”，
          对计划中的模块边界与依赖方向进行预检，并将结论写入 plan.md。
  Step 4：把 [ARGUMENTS] 段（见下）的全文作为 skill body 中 `$ARGUMENTS`
          占位符的实际值，执行 skill 内的指令

[ARGUMENTS]
  基于 ${WORK_DIR}/spec.md 以测试驱动开发模式构建计划。
  上下文白名单见 ${WORK_DIR}/context.md，仅在该白名单范围内读取背景知识。
  工作目录为 ${WORK_DIR}，不创建新分支。
  当 meta.yaml.tech_type=fullstack 时，plan.md 中的接口定义是前后端的权威契约：
  - 面向前端消费的接口使用明确的 RESTful / JSON 语义，避免仅适用于 service-to-service 的模式；
  - 列表接口须定义分页、排序、过滤的 query 参数及默认行为；
  - 响应字段须明确前端 camelCase 映射，避免未说明的深层嵌套或多态结构；
  - 若使用 Proto，须同时定义 HTTP 映射的路径与方法。
  ${REENTRY_CONTEXT}

[就地合规自检]
speckit-plan 完成后继续：
1) 读取 plan.md / research.md / data-model.md
2) 按以下 3 个维度核对：
   - 完整度：plan 是否覆盖 spec.md 所有需求
   - research 合规：技术选型是否违反架构/安全/编码规范
   - 项目宪章：是否违反 .specify/memory/constitution.md 硬约束
3) 按 ./references/report-template.md 格式输出 ${WORK_DIR}/plan-report.md

[完成后自检]
1) 确认 plan.md + plan-report.md 已生成

[返回 JSON 要求]
- phase: "plan"
- 必须填写 compliance（verdict + report）
- produced 含 plan.md / research.md / data-model.md / plan-report.md
```

### 2.3 Tasks-Generate 模板（tapd-story-tasks 第一段）

**适用阶段**：`STAGE=tasks-generate`
**phase 推进**：暂不推进（待 tasks-analyze 通过后推进）

```
[阶段任务]
基于 ${WORK_DIR}/plan.md 通过 use_skill 调用 speckit-tasks 生成全量任务清单。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：调用 use_skill(command="speckit-tasks") 加载 skill 内容到上下文
  Step 3：读取
          skills/tech-lead/references/task-decomposition.md 的“1.3 Walking Skeleton”和
          “1.4 依赖拓扑排序”；以最小可运行主流程作为首个可交付切片，显式标注任务依赖与关键路径。
  Step 4：把 [ARGUMENTS] 段（见下）的全文作为 skill body 中 `$ARGUMENTS`
          占位符的实际值，执行 skill 内的指令

[ARGUMENTS]
  基于 ${WORK_DIR}/plan.md 以 TDD 模式构建任务，
  覆盖单元/集成/端到端测试。上下文白名单见 ${WORK_DIR}/context.md。
  输出：${WORK_DIR}/tasks.md。
  ${REENTRY_CONTEXT}

[完成后自检]
1) 确认 tasks.md 已生成且可读

[返回 JSON 要求]
- phase: "tasks-generate"
- 不填 compliance
- produced 含 tasks.md
- 若 tasks.md 未生成 → status=fail
```

### 2.4 Tasks-Analyze 模板（tapd-story-tasks 第二段）

**适用阶段**：`STAGE=tasks-analyze`
**phase 推进**：`researched → tasks-generated`

```
[阶段任务]
通过 use_skill 调用 speckit-analyze 验证产物合规，按 references/report-template.md
格式输出 tasks-report.md。

[重入清理]
若存在 "${WORK_DIR}/tasks-report.md"，使用delete_file删除。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：调用 use_skill(command="speckit-analyze") 加载 skill 内容到上下文
  Step 3：把 [ARGUMENTS] 段（见下）的全文作为 skill body 中 `$ARGUMENTS`
          占位符的实际值，执行 skill 内的指令

[ARGUMENTS]
  确认 ${WORK_DIR}/ 下产物（spec.md / plan.md / research.md / tasks.md）
  没有违反宪章、架构设计、安全规范、编码规范。
  上下文白名单见 ${WORK_DIR}/context.md。
  输出报告格式严格按 references/report-template.md 模板。
  输出位置：${WORK_DIR}/tasks-report.md。
  ${REENTRY_CONTEXT}

[完成后自检]
1) 确认 tasks-report.md 已生成且含 Verdict 段

[返回 JSON 要求]
- phase: "tasks-analyze"
- 必须填写 compliance（verdict + report）
- produced 含 tasks-report.md
```

### 2.5 Implement 模板（tapd-story-implement）

**适用阶段**：`STAGE=implement`
**phase 推进**：`confirmed → implemented`

```
[阶段任务]
基于 ${WORK_DIR}/tasks.md 通过 use_skill 调用 speckit-implement 完成所有任务（TDD 模式），
通过单元/集成/端到端测试。

[关键约束]
- 代码改动仅允许触达 context.md 的 Code scope 路径；越界视为 fail
- 不创建/切换分支
- round > 1 时按原地修复策略：不回滚，针对失败点修复
- code_preserved=true 时：保留代码做差量修复

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：调用 use_skill(command="speckit-implement") 加载 skill 内容到上下文
  Step 3：按 meta.yaml.tech_type 读取以下对应实现指南的
          “TDD 执行流程”和“代码自查清单”：
          backend → skills/backend-developer/references/implementation.md；
          frontend → skills/frontend-developer/references/implementation.md；
          fullstack → 按当前实现轮次读取对应文件。
  Step 4：把 [ARGUMENTS] 段（见下）的全文作为 skill body 中 `$ARGUMENTS`
          占位符的实际值，执行 skill 内的指令

[ARGUMENTS]
  按以下优先级读取实现上下文：
  - ${WORK_DIR}/tasks.md：任务顺序、文件范围与验收测试；
  - ${WORK_DIR}/plan.md：接口契约、数据模型和架构决策的权威来源；
  - ${WORK_DIR}/spec.md：功能范围与验收标准；
  - ${WORK_DIR}/context.md：项目背景、可读取文件与 Code scope。
  基于 tasks.md 实现所有任务（TDD 模式）；不得以实现便利为由偏离 plan.md 中的接口契约。
  代码改动仅允许触达 context.md 的 Code scope 路径。
  实现完成后运行单元/集成/端到端测试并确保通过。
  ${REENTRY_CONTEXT}

[完成后自检]
1) 校验改动文件均在 Code scope 内——越界 → fail
2) 确认测试全通过，填写顶层 tests 字段（unit / integration / e2e）

[返回 JSON 要求]
- phase: "implement"
- 必须填写顶层 tests 字段
- produced 列出改动文件路径
- 不期望 blocked_on_questions
```

### 2.5.1 统一 Severity 判定标准（所有 Validate 阶段）

所有 validate agent 必须按下表对 finding 分级。项目专项规范的等级更严格时，以专项规范为准：

| 级别 | 判定条件 | 典型示例 |
|------|----------|----------|
| CRITICAL | 可造成数据损坏、敏感数据泄露、可被直接利用的安全漏洞，或生产服务不可用 | SQL 注入、未鉴权的敏感写接口、破坏性数据迁移、循环依赖导致构建失败 |
| HIGH | 违反已确认的架构或接口契约，或缺失核心用户路径测试；不修复将导致集成失败或高概率功能错误 | 依赖方向反转、响应字段偏离 plan.md、核心 happy path 无测试 |
| MEDIUM | 损害可维护性、可观测性或边界场景可靠性，但不阻断当前主流程 | 错误处理遗漏、边界条件未覆盖、重复逻辑、命名与通用语言不一致 |
| LOW | 改进建议，不影响功能正确性或可维护性基线 | 可读性优化、注释补充、非关键日志优化 |

**Verdict 映射**：
- 任一 CRITICAL 或 HIGH finding，且全部归因 `code-self` → `needs_fix`；
- 任一 CRITICAL 或 HIGH finding，且存在 `spec-insufficient` / `plan-insufficient` 根因
  → 对应输出 `spec_insufficient` / `plan_insufficient`；
- 仅有 MEDIUM / LOW finding，或无 finding → `LGTM`；
- 多个根因并存时，取最深上游根因作为 verdict。

### 2.6 Validate-Arch 模板（tapd-story-validate 段 #1）

**适用阶段**：`STAGE=validate-arch`
**phase 推进**：与 validate-security / validate-codereview 全部 LGTM 后推进到 `validated`

```
[阶段任务]
基于 context.md 列出的架构设计文档与项目宪章，并主动读取项目根 AGENTS.md（项目定位与关键规范索引），
对需求 #${ID} 的代码实现做架构合规与项目原则合规校验，
按 report-template.md 格式输出 validate-arch-report.md。

[重入清理]
若存在 ${WORK_DIR}/validate-arch-report.md，使用delete_file删除。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：直接执行评审，按下方 [评审任务] 段执行
  Step 3：输出报告到 ${WORK_DIR}/validate-arch-report.md（按 references/report-template.md 模板）

[评审任务]
  对需求 #${ID} 的代码实现做架构合规校验：
  1) 分层架构约束（依赖方向）
  2) 循环依赖
  3) 模块边界
  4) 代码改动是否在 Code scope 白名单内
  5) 项目原则合规：主动读取项目根 AGENTS.md，提取项目定位与关键规范索引，
     校验代码实现是否违反项目定位或关键规范；仅只读 AGENTS.md，不修改。
  输出报告格式必须按 references/report-template.md 模板。
  输出位置：${WORK_DIR}/validate-arch-report.md。
  ${REENTRY_CONTEXT}

[分级与 verdict 映射]
按本文件“统一 Severity 判定标准”分级并决定 Verdict；每条 finding 必须标注根因。
项目原则合规类 finding（违反 AGENTS.md 项目定位/关键规范）按统一 Severity 分级，纳入 Verdict。

[对抗性自检 — 必须执行]
在输出 validate-arch-report.md 前，额外检查：
1) plan.md 中的架构决策本身是否不合理；不得因其为上游已确认产物而默认正确。
2) 若实现困难源于 plan.md 的模块边界、依赖方向或接口约束不充分，归因
   `spec-insufficient` 或 `plan-insufficient`，而不是 `code-self`。
3) 代码是否在 plan.md 未覆盖的区域作出隐式架构决策；此类偏差必须记录为 finding。

[完成后自检]
1) 确认 validate-arch-report.md 已生成且含 Verdict 段

[返回 JSON 要求]
- phase: "validate-arch"
- 必须填写 compliance（verdict + report）
- produced 含 validate-arch-report.md
```

### 2.7 Validate-Security 模板（tapd-story-validate 段 #2）

**适用阶段**：`STAGE=validate-security`

```
[阶段任务]
基于 context.md 列出的安全规范（含 bk-security-redlines 三大红线）+ 项目安全文档，
对需求 #${ID} 的代码实现做安全校验，按 references/report-template.md 格式输出 validate-security-report.md。

[重入清理]
若存在"${WORK_DIR}/validate-security-report.md"，使用delete_file删除。

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：直接执行评审，按下方 [评审任务] 段执行
  Step 3：输出报告到 ${WORK_DIR}/validate-security-report.md（按 references/report-template.md 模板）

[评审任务]
  基于 ${WORK_DIR}/context.md 中列出的安全规范（含 skills/bk-security-redlines/ 三大红线），
  对需求 #${ID} 的代码实现做安全校验：
  1) 输入校验（边界、类型、长度、字符集、注入）
  2) 鉴权（横向/纵向越权）
  3) 敏感数据加密（硬编码密钥/凭证、加密算法合规）
  4) 常见风险（SQL 注入、XSS、路径穿越、不安全反序列化、SSRF）
  输出报告格式必须按 references/report-template.md 模板。
  输出位置：${WORK_DIR}/validate-security-report.md。
  ${REENTRY_CONTEXT}

[分级与 verdict 映射]
按本文件“统一 Severity 判定标准”分级并决定 Verdict；每条 finding 必须标注根因。

[完成后自检]
1) 确认 validate-security-report.md 已生成且含 Verdict 段

[返回 JSON 要求]
- phase: "validate-security"
- 必须填写 compliance（verdict + report）
- produced 含 validate-security-report.md
```

### 2.8 Validate-CodeReview 模板（tapd-story-validate 段 #3）

**适用阶段**：`STAGE=validate-codereview`
**执行 agent**：本段由 `Task(subagent_name="code-reviewer", prompt=<PROMPT>)` 发起
（区别于 validate-arch 段由 `tech-lead` 执行；validate-security 段同样由 `code-reviewer` 执行，
但评审维度不同，见下方 [评审任务]）。
**phase 推进**：与 validate-arch / validate-security 全部 LGTM 后推进到 `validated`

```
[重入清理]
若存在"${WORK_DIR}/validate-codereview-report.md"，使用delete_file删除。

[分级与 verdict 映射]
- 分级依据见本文件“统一 Severity 判定标准”。
- 无 CRITICAL/HIGH → Verdict=LGTM。
- 存在 CRITICAL/HIGH 且全部归因 code-self → Verdict=needs_fix。
- 存在 CRITICAL/HIGH 且任一归因 spec-insufficient / plan-insufficient → 分别输出
  Verdict=spec_insufficient / plan_insufficient。

[根因标注 — 强制]
对每条 finding，按 references/report-template.md「Finding 根因枚举」标注根因，
取值 ∈ {code-self | spec-insufficient | plan-insufficient}：
- 实现自身缺陷 → code-self（触发原地修复）
- 因 spec.md/plan.md 信息不足导致的实现问题 → spec-insufficient / plan-insufficient（触发回退重入）

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：直接执行评审，按下方 [评审任务] 段执行
  Step 3：用 write_to_file 输出报告到 ${WORK_DIR}/validate-codereview-report.md
          （按 references/report-template.md 模板）

[评审任务]
  经 use_skill("code-review") 加载统一 CR 方法论的分级评审清单，对需求 #${ID} 的完整代码实现做 CodeReview。
  背景（用于根因判断）：${WORK_DIR}/spec.md、plan.md、tasks.md。
  上下文白名单见 ${WORK_DIR}/context.md，仅在该白名单范围内读取背景知识。
  [评审维度]
   1) 代码规范（命名/格式/结构 — 以项目编码规范为准）
   2) 逻辑正确性（边界/异常/空值防御）
   3) 性能隐患（N+1、内存泄漏）
   4) 可维护性（复杂度、重复）
   5) 测试覆盖度
  按 CRITICAL / HIGH / MEDIUM / LOW 四级分级，并对每条 finding 标注根因。
  Verdict 按 [分级与 verdict 映射] 规则判定。
  输出报告必须按 references/report-template.md 模板。
  输出位置：${WORK_DIR}/validate-codereview-report.md。
  ${REENTRY_CONTEXT}

[完成后自检]
1) 确认 validate-codereview-report.md 已生成且含 Verdict 段
2) 确认每条 finding 均带根因标注

[返回 JSON 要求]
- phase: "validate-codereview"
- 必须填写 compliance（verdict + report）
- produced 含 validate-codereview-report.md
```

### 2.9 Validate-Fix 模板（tapd-story-validate 修复段，仅当需要）

**适用阶段**：`STAGE=validate-fix`
**触发条件**：四段评审中任一 verdict=needs_fix 且全部归因 code-self

```
[阶段任务]
基于四份评审报告、spec.md 和 plan.md，对需求 #${ID} 的代码做原地修复，修复后重跑测试。

[关键约束]
- 修复范围限定在 Code scope 白名单内
- 不创建/切换分支；不做 git 回滚
- 仅修复评审报告中的 finding，不触碰 LGTM 部分

[执行命令]
  Step 1：echo banner 到 ${LOG_FILE}（按 §1.2 [日志落盘]）
  Step 2：直接执行修复，按下方 [修复任务] 段执行
  Step 3：执行完成后运行单元/集成/端到端测试

[修复任务]
  先读取 ${WORK_DIR}/spec.md（验收标准）和 ${WORK_DIR}/plan.md（接口与架构契约），
  再基于以下四份评审报告，对需求 #${ID} 的代码做原地修复（不要全量重新实现）：
  - ${WORK_DIR}/validate-arch-report.md
  - ${WORK_DIR}/validate-security-report.md
  - ${WORK_DIR}/validate-codereview-report.md
  - ${WORK_DIR}/validate-test-report.md
  约束：
  1、修复范围限定在 ${WORK_DIR}/context.md 的 Code scope 白名单内。
  2、不创建/切换分支；不做 git 回滚。
  3、仅修复评审报告中的 finding，不触碰 LGTM 部分；修复不得偏离 spec.md / plan.md。
  修复完成后运行单元/集成/端到端测试，全部通过后返回。
  ${REENTRY_CONTEXT}

[完成后自检]
1) 校验改动文件均在 Code scope 内
2) 确认测试全通过，填写顶层 tests 字段（unit / integration / e2e）

[返回 JSON 要求]
- phase: "validate-fix"
- 必须填写顶层 tests 字段
- produced 列出修复的文件路径
- 不填 compliance（后续 round 重新跑四段校验）
```

### 2.10 Validate-Test 模板（tapd-story-validate 段 #4）

**适用阶段**：`STAGE=validate-test`
**执行 agent**：`qa-engineer`
**日志文件**：`${WORK_DIR}/process-validate-test.log`

```
[执行命令]

Step 1: 读取以下文件建立评审上下文：
- `${WORK_DIR}/spec.md`（了解需求范围与验收标准）
- `${WORK_DIR}/plan.md`（了解接口设计与模块边界）
- `${WORK_DIR}/tasks.md`（了解预期的测试覆盖点）
- 代码变更（通过 `git diff ${IMPLEMENT_BASELINE_COMMIT}` 获取 Code scope 内的变更范围）

Step 2: 对变更范围内的测试文件进行以下维度评审：
- **测试金字塔合理性**：单测 / 集成测试 / E2E 测试比例是否合理
- **核心路径覆盖**：主流程 Happy Path 是否有对应测试
- **边界条件覆盖**：空值、最大值、并发、重试等边界场景
- **错误路径覆盖**：错误码、异常处理、超时场景是否有测试
- **BDD 场景完整性**（如有）：Given-When-Then 描述是否清晰且完整
- finding 分级与 Verdict 必须遵循本文件“统一 Severity 判定标准”

Step 3: 按统一报告模板（下方）输出 `${WORK_DIR}/validate-test-report.md`

[报告模板]

# Validate Test Report

**需求 ID**: ${ID}
**Attempt**: ${ATTEMPT} / Round: ${ROUND}
**评审时间**: ${TS}

## Verdict: LGTM | needs_fix | spec_insufficient | plan_insufficient

## 评审摘要

[2-3 句话总结测试覆盖质量]

## Findings

| 级别 | 位置 | 问题描述 | 根因 | 建议 |
|------|------|---------|------|------|
| CRITICAL/HIGH/MEDIUM/LOW | file:line | ... | code-self / spec-insufficient / plan-insufficient | ... |

## 覆盖评估

| 维度 | 状态 | 说明 |
|------|------|------|
| 单元测试 | ✅/⚠️/❌ | ... |
| 集成测试 | ✅/⚠️/❌ | ... |
| 边界条件 | ✅/⚠️/❌ | ... |
| 错误路径 | ✅/⚠️/❌ | ... |

[归因判定]
- `LGTM`：无 CRITICAL/HIGH finding，测试覆盖达标
- `needs_fix` + `code-self`：测试文件遗漏覆盖点，开发者补测试即可
- `plan_insufficient`：tasks.md / plan.md 未包含必要的测试场景描述，需回退
- `spec_insufficient`：spec.md 的验收标准或边界条件不足，需回退
- 每条 finding 必须标注根因；有多个根因时按最深上游根因输出 verdict。

[完成后自检]
1) 确认 validate-test-report.md 已生成且含 Verdict 段

[返回 JSON 要求]
- phase: "validate-test"
- 必须填写 compliance（verdict + report）
- produced 含 validate-test-report.md
```

---

## 3. 重入增量段（attempts > 1 或 round > 1 时追加）

> pipeline 主编排在渲染时，将 §3 内容追加到 SUBAGENT_PROMPT 末尾，
> 同时将修正摘要填入各阶段命令中的 `${REENTRY_CONTEXT}` 占位符。

### 3.1 通用回退重入增量（attempts > 1）

pipeline 主编排追加到 PROMPT 末尾，**同时**将核心改进点摘要填入 `${REENTRY_CONTEXT}`：

```
[本轮为第 ${ATTEMPT} 次尝试 — 改进点指引]
请优先解决 ${WORK_DIR}/iteration-patches/attempt-${ATTEMPT}.md 中列出的：
- 失败阶段：<failed_phase>
- 根因：<root_cause>
- 期望改进点：<expected_improvements>
- 上下文补丁：<patch_to_context 摘要>
- 需求补丁：<patch_to_req 摘要>

若 code_preserved=true，保留当前代码仅做差量修复：
- 未解 Findings：<unresolved_findings 列表>

请确保本轮产出覆盖以上改进点。
```

**`${REENTRY_CONTEXT}` 填充内容**（注入到 ARGUMENTS 段内部）：
```
[重入修正] 本轮为第 ${ATTEMPT} 次尝试。改进要点：<expected_improvements 一句话摘要>。
详见 ${WORK_DIR}/iteration-patches/attempt-${ATTEMPT}.md。
```

### 3.2 同 attempt 内 round 重试增量（round > 1）

追加到 PROMPT 末尾，**同时**填入 `${REENTRY_CONTEXT}`：

```
[本轮为第 ${ATTEMPT} 次尝试的第 ${ROUND} 次重试 — 已知失败原因]
上一轮失败摘要：<issue>
失败的测试用例（若有）：<failed_tests>
关键错误证据（若有）：<key_errors>

请基于当前产物/代码定位与修复，不要全量重新实现。
```

**`${REENTRY_CONTEXT}` 填充内容**：
```
[重试修正] round=${ROUND}，上一轮失败：<issue 一句话摘要>。针对性修复，不全量重做。
```

### 3.3 Specify 阶段代问重入增量（round > 1 且来源是 blocked_on_questions）

```
[本轮为第 ${ROUND} 次代问重入 — 用户已答复]
${WORK_DIR}/questions.md 中以下条目已被用户答复：
<answered 条目列表>

请基于新答复继续 speckit-specify；若仍有新问题，继续追加 open 条目。
```

**`${REENTRY_CONTEXT}` 填充内容**：
```
[代问重入] 用户已答复 questions.md 中的问题，请基于新答复继续生成 spec.md。
```

---

## 4. 渲染流程总结

```
§1.2 通用前缀
   │
   ├─ §1.3 通用回传 JSON Schema
   │
   ├─ §2.x 当前阶段差异化模板（仅一份）
   │
   ├─ §3.1 回退重入增量（仅 attempts > 1）
   │
   ├─ §3.2 round 重试增量（仅 round > 1）
   │
   └─ §3.3 specify 代问重入增量（仅 specify + blocked_on_questions 回弹）
```

**关键**：§3 的内容同时追加到两个位置：
1. SUBAGENT_PROMPT 末尾（subagent 整体可见）
2. `${REENTRY_CONTEXT}` 占位符（注入到各阶段 [ARGUMENTS] 段内部，确保 use_skill 加载的 speckit skill 在 `$ARGUMENTS` 解析时可见）

---

## 5. 验收清单（pipeline 主编排渲染前自检）

- [ ] 所有占位符已被替换（grep `\${` 应返回空）
- [ ] `${LOG_FILE}` 已按阶段选择正确的日志路径
- [ ] `${REENTRY_CONTEXT}` 已填充（首次为空字符串，重入时为修正摘要）
- [ ] `attempts > 1` 时已追加 §3.1
- [ ] `round > 1` 时已追加 §3.2
- [ ] specify 代问重入时已追加 §3.3
- [ ] 各阶段命令中引用了 `references/report-template.md`（需要产出报告的阶段）
- [ ] `[spec-cost-marker]` 行已包含完整的 work_dir / stage / attempt / round / ts 字段
