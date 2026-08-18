---
name: tapd-story-review
slug: tapd-story-review
version: 1.0.0
description: |
  TAPD 需求评审技能。用于需求澄清或需求拆分结果已回写 TAPD 后，发起评审、@指定评审人、
  读取父单和子单评论、汇总待处理意见，并在满足评审通过规则前，驱动
  重新澄清或重新拆分。Use this skill whenever the user mentions 需求评审, review需求,
  评审澄清结果, 评审拆单结果, TAPD评论处理, 需求review闭环, 需求待审批, or any workflow
  involving TAPD story review after clarification or evaluation.
metadata:
  requires:
    mcps: ["tapd"]
---

# TAPD 需求评审

## 概述

本技能用于围绕 TAPD 需求形成一个可重入的评审闭环。

职责边界：

- 发起评审：将父单流转到 `for approve`，并在父单 / 子单评论区 `@` 评审人，将单据处理人更新为评审人列表
- 读取评论：拉取父单与子单评论，识别本轮新增反馈
- 汇总意见：输出结构化的评审意见清单和评审结论
- 判断状态：识别当前是否仍在等待评审，或是否已满足评审通过规则并将父单推进为 `approved`

本技能不负责：

- 重新澄清需求
- 重新拆分子需求
- 直接修改其他 skill 的 phase

这些动作应由用户或上层 pipeline 根据本 skill 的结论决定。

## 前置条件

- TAPD MCP 服务可用
- 用户提供至少一个父需求 ID
- 首轮发起评审时，若历史上下文中不存在评审人信息，需要用户指定评审人列表
- 目标需求的澄清结果或拆分结果已经写回 TAPD
- `workspace_id` 可由用户提供，或从项目根目录 `project.json` 读取

## 输入

| 参数 | 来源 | 必需 | 说明 |
|------|------|------|------|
| 父需求 ID | 用户输入 | 是 | 支持短 ID 或 19 位长 ID |
| workspace_id | 用户输入 > `project.json` | 是 | TAPD 工作空间 ID |
| phase | 调用方传入 / 本地元数据 / 上下文推断 | 否 | `clarification` 或 `evaluation`；优先自动推断，无法判断时再询问用户 |
| mode | 调用方传入 / 本地元数据 / 上下文推断 | 否 | `request-review` / `process-feedback` / `check-status`；优先自动推断 |
| reviewers | 用户输入 / 本地元数据 / 历史评论解析 | 否 | 评审人用户名列表，使用 `@用户名` 格式；后续轮次默认沿用上一轮评审人 |
| answered_questions | 调用方传入 / `questions.md` 已回答条目 | 否 | 当上一轮因待确认问题返回 `blocked` 后，调用方应把当前 phase 最新 round 的已回答问题传回本 skill，用于继续裁决 |
| include_children | 用户输入 / phase 默认值 / 父单结构推断 | 否 | 是否将父单下子单一起纳入评审；`evaluation` 默认 `true`，`clarification` 默认 `false` |
| sync_children_status | 用户输入 | 否 | 是否将子单也流转到 `for approve`，默认 `false` |
| 背景知识 | 用户指定 > AGENTS.md 自动查找 | 否 | 架构文档、模块文档、接口文档等 |

### 参数推断与回退规则

#### `phase`

按以下优先级确定：

1. 调用方显式传入
2. 读取当前阶段 review 元数据：`docs/reqs/<父需求短ID>/<phase>-review-meta.yaml`
3. 结合当前上下文推断：
   - 正在处理澄清评审 → `clarification`
   - 正在处理拆单 / 工时评审 → `evaluation`
4. 仍无法判断 → 询问用户

#### `mode`

按以下优先级确定：

1. 调用方显式传入
2. 若用户明确只想查看当前状态 → `check-status`
3. 若存在当前阶段的 review 元数据，且父单状态为 `for approve` → `process-feedback`
4. 若不存在当前阶段的 review 元数据 → `request-review`

#### `reviewers`

按以下优先级确定：

1. 用户或调用方本轮显式指定
2. 读取当前阶段 review 元数据中的 `reviewers`：`docs/reqs/<父需求短ID>/<phase>-review-meta.yaml`
3. 从最近一条 `【需求评审发起】` 评论中解析 `@用户名`
4. 以上均无 → 询问用户

> 因此，第二次及后续复审时，若用户未重新指定评审人，默认沿用上一轮评审人。

#### `include_children`

按以下优先级确定：

1. 用户或调用方显式指定
2. `phase=evaluation` → 默认 `true`
3. `phase=clarification` → 默认 `false`

#### 子单范围

当 `include_children=true` 时，不要求用户手动提供子单列表。系统会自动：

1. 读取父单 `children_id`
2. 查询所有子单详情
3. 若父单没有子单，则自动降级为仅评审父单

## 模式说明

### `request-review`

用于发起新一轮评审。典型场景：

- 澄清结果刚写回 TAPD，准备请同事 review
- 拆单和工时评估结果已写回，准备评审父单和子单
- 根据上一轮评论修订完成后，再次请求复审

### `process-feedback`

用于读取评论并决定后续动作。典型场景：

- 评审人已在父单或子单评论区留下意见
- 需要汇总评论，识别哪些要修改、哪些要向用户确认
- 需要重新调用澄清或拆单 skill 修订内容

### `check-status`

用于只检查当前评审状态，不做任何写回。典型场景：

- 用户只想知道需求是否已被评审人通过
- 流水线轮询父单状态是否已变为 `approved`

## 执行流程

### 1. 参数收集与环境准备

#### 1.1 确定 `workspace_id`

按以下优先级确定：

1. 用户消息中显式指定 → 直接使用
2. `project.json` 中的 `workspace_id` → 使用 `read_file` 读取并解析
3. 以上均无 → 询问用户

#### 1.2 解析父需求 ID

从用户输入提取需求 ID，并同时归一化出短 ID / 长 ID：

- 若输入为短 ID：
  1. 记录为 `short_id`
  2. 调用 TAPD MCP 转换为 `long_id`
- 若输入为 19 位长 ID：
  1. 记录为 `long_id`
  2. 按仓库统一约定，将 `long_id` 的后 8 位截取为 `short_id`

约束：

- 后续查询 TAPD 评论、状态和子单时统一使用 `long_id`
- 本地文件命名统一使用 `short_id`
- 当输入为长 ID 时，不再尝试向 TAPD 反解另一套短 ID；统一使用 `long_id` 后 8 位作为 `short_id`

#### 1.3 读取父需求详情

使用 TAPD MCP `stories_get` 提取父需求信息：

```
调用参数:
  workspace_id: <workspace_id>
  id: <父需求ID>
  with_v_status: "1"
  fields: "id,name,description,owner,parent_id,children_id,priority_label,v_status"
```

记录以下字段：

- `id`
- `name`
- `description`
- `owner`
- `children_id`
- `v_status`

如果父需求不存在，终止并告知用户。

#### 1.4 按需读取子需求详情

当 `phase=evaluation` 且 `include_children=true` 时：

1. 解析 `children_id`
2. 使用 TAPD MCP 逐一查询子需求详情
3. 记录每个子单的 `id`、`name`、`description`、`v_status`

若父单没有子需求，则继续流程，但输出中明确标注“当前未检测到子需求，本轮仅评审父单”。

#### 1.5 收集背景知识

按以下优先级确定：

1. 用户显式指定背景文档路径 → 读取指定文档
2. 用户未指定 → 读取项目根目录 `AGENTS.md`，从中识别相关架构文档、模块文档、
   安全规范、接口文档等

背景知识在分析评论合理性时使用，不足以替代用户确认。

### 2. 读取本地评审元数据

为保证技能可重入，在本地维护评审元数据文件。

路径规则如下：

- `docs/reqs/<父需求短ID>/<phase>-review-meta.yaml`

建议至少记录：

```yaml
short_id: "32139656"
long_id: "1070046748132139656"
story_id: "1070046748132139656"
phase: "clarification"
round: 1
reviewers:
  - "@alice"
  - "@bob"
previous_owner: "alice"
previous_children_owners: {}
current_review_owners:
  - "@alice"
  - "@bob"
approval_rule: "any_one_reviewer_approve"
include_children: false
children_ids: []
last_request_comment_ids: []
last_processed_comment_ids: []
status: "waiting_review"
```

如果文件不存在，则视为首次进入该阶段评审。

### 2.1 评审详情文件

为避免多轮 review 返工时把父单或子单的既有内容整段覆盖丢失，同时减少本地产物数量，本
skill 在每个评审阶段只维护 **1 个供人阅读的评审详情文件**。

路径规则：

- 澄清评审：`docs/reqs/<父需求短ID>/clarification-review.md`
- 评估评审：`docs/reqs/<父需求短ID>/evaluation-review.md`

命名规则：

1. 统一使用父需求短 ID 作为目录名主体
2. 评审详情文件命名为 `<phase>-review.md`
3. 评审元数据文件命名为 `<phase>-review-meta.yaml`
4. 示例：
   - `docs/reqs/32139656/clarification-review.md`
   - `docs/reqs/32139656/clarification-review-meta.yaml`
   - `docs/reqs/32139656/evaluation-review.md`
   - `docs/reqs/32139656/evaluation-review-meta.yaml`

文件内容按轮次追加，至少包含以下章节：

- `Round N / 发起评审基线快照`：本轮发起评审时父单 `description` 的完整快照
- `Round N / 子单基线快照`：每个子单 `description` 的完整快照（仅 `evaluation` 且
  `include_children=true` 时需要）
- `Round N / 评审发起评论`：本轮发起评审时写入评论区的正文
- `Round N / 评审反馈摘要`：处理评论后的结构化意见摘要
- `Round N / 修订保护边界`：本轮返工时必须保留的内容清单与修订边界

约束如下：

- 评审详情文件按轮次追加，不应为了写入新一轮而覆盖旧轮次内容
- `needs_rework` 时，上层流程必须把“当前 TAPD 内容 + 评审详情文件中最近一轮的基线快照 /
  反馈摘要 / 修订保护边界”一起作为下一轮修订输入
- review skill 自身不直接重写父单或子单 `description`

### 2.2 评审通过规则

默认通过规则为：`any_one_reviewer_approve`。

含义：

- 当前轮次中，只要任一评审人给出 `approve` 且 `是否阻塞=no`，即可视为本轮评审通过
- 满足通过规则后，review skill 应将父单状态推进为 `approved`
- 推进为 `approved` 前，应先将处理人从评审人列表恢复为 `previous_owner`

如后续需要支持“全员通过”或更复杂规则，应通过新增 `approval_rule` 扩展，而不是修改默认语义。

## 3. `request-review` 模式

### 3.1 校验发起条件

发起评审前检查：

- 父需求 `description` 不为空
- `reviewers` 非空
- `phase=clarification` 时父单已完成澄清回写
- `phase=evaluation` 时父单已完成拆单说明回写；若要求包含子单，则子单描述也已写回

如上述任一条件不满足，终止并告知用户先完成对应前置 skill。

### 3.2 计算评审轮次

按以下优先级确定轮次：

1. 本地元数据已有轮次 → `round + 1`
2. 无本地元数据时，扫描父单评论中历史 `【需求评审发起】` 标记 → 最大轮次 + 1
3. 均无 → 当前轮次为 `1`

### 3.3 生成本轮评审详情

在发起评审前，先把本轮评审基线写入当前阶段的评审详情文件：

1. 新建或读取当前阶段评审详情文件
2. 追加 `## Round <round>` 章节
3. 写入父单 `description` 的完整快照
4. 若 `phase=evaluation` 且 `include_children=true`，在同一轮次章节中追加所有子单的
   `description` 快照
5. 预留“评审发起评论 / 评审反馈摘要 / 修订保护边界”章节占位

> 这一步的目的不是做归档，而是为后续 `needs_rework` 提供“保留原内容”的修订基线。

### 3.4 流转状态

1. 将父需求状态流转为 `for approve`
2. 若 `phase=evaluation` 且 `sync_children_status=true`，将子需求状态也流转为 `for approve`
3. 在流转前记录父单当前 `owner` 为 `previous_owner`
4. 若需要同步子单状态，则按需记录每个子单的原始处理人为 `previous_children_owners`
5. 将纳入本轮评审范围的单据处理人字段临时切换为完整评审人列表，并记录为 `current_review_owners`

> 父单状态是唯一评审完成信号。子单状态仅用于辅助展示，不作为结束条件。

### 3.5 生成评审发起评论

在父单评论区添加标准评论，并 `@` 评审人，将单据处理人字段修改为完整评审人列表：

```text
【需求评审发起】
评审阶段：clarification / evaluation
评审轮次：R<round>
评审范围：父单 / 父单+子单
父需求：<名称> #<短ID>
本轮变更摘要：
1. ...
2. ...

请评审人按以下格式反馈：
- 结论：approve / change_required / question
- 范围：父单 / 子单 #<ID>
- 是否阻塞：yes / no
- 反馈内容：...

评审人：
@alice @bob
```

> 若用户未提供“本轮变更摘要”，则由 Agent 基于当前 description 与上一轮元数据自动总结。

### 3.6 按需对子单补充评论

当 `phase=evaluation` 且 `include_children=true` 时，在每个子单评论区补充一条简化评论：

```text
【子需求待评审】
所属父需求：<父需求名称> #<父需求短ID>
评审阶段：evaluation
评审轮次：R<round>
请在本子单下评论拆分边界、依赖关系、工时和价值规模是否合理。
@alice @bob
```

### 3.7 写入本地元数据

更新当前阶段 review 元数据文件：`docs/reqs/<父需求短ID>/<phase>-review-meta.yaml`

- 当前轮次
- 评审人列表
- 子单列表
- 当前状态 `waiting_review`
- 本轮评审发起时间
- 当前阶段评审详情文件路径
- `current_review_owners`
- `approval_rule`

### 3.8 输出

输出结构化结果：

```markdown
## 评审已发起

- 父需求：xxx
- 评审阶段：clarification
- 当前轮次：R1
- 状态：for approve
- 评审人：@alice, @bob
- 下一步：等待评审人评论，随后使用 `process-feedback` 处理
```

## 4. `process-feedback` 模式

### 4.1 先检查父单状态

优先读取父需求当前 `v_status`：

- 若已为 `approved` → 先恢复处理人为 `previous_owner`，再结束并输出“评审已完成”
- 若仍为 `for approve` → 继续处理评论
- 若被改成其他状态 → 输出“状态异常”，请用户确认是否继续

### 4.2 拉取本轮评论

读取父单评论：

```
调用参数:
  workspace_id: <workspace_id>
  entry_type: "stories"
  entry_id: <父需求长ID>
```

当 `phase=evaluation` 且 `include_children=true` 时，逐一读取子单评论。

评论筛选规则：

- 仅处理当前轮次评审发起之后的新评论
- 忽略本 skill 自己发起的模板评论
- 忽略已记录在 `last_processed_comment_ids` 中的评论
- 保留评审人评论；非评审人评论默认标注为“旁路意见”，由用户决定是否纳入

若检测到当前轮次已有评审人发表有效评论，且父单当前处理人仍包含 `current_review_owners`，
则应立即将处理人恢复为 `previous_owner`。若本轮同步过子单状态，也按同样规则恢复对应子单
处理人。该恢复操作应为幂等，多次执行不应产生副作用。

### 4.3 汇总与分类评论

将评论整理为“评审意见清单”，每条至少包含：

- 评论 ID
- 评论人
- 来源单据（父单 / 子单 ID）
- 原始评论摘要
- 建议动作：`accept` / `reject` / `needs-user-confirmation`
- 影响范围：需求描述 / 子单拆分 / 依赖关系 / 工时 / 价值规模
- 是否阻塞

分类原则：

- **accept**：意见明确、低风险、与现有背景知识不冲突
- **needs-user-confirmation**：涉及业务取舍、范围调整、上线策略、边界变化
- **reject**：明显误解需求、与已确认约束冲突，或超出当前阶段范围

> 本技能可以判断评论“是否有修改价值”，但不能擅自改变关键业务方向。涉及业务取舍时必
> 须向用户确认。
>
> 汇总完成后，应将本轮“评审反馈摘要”写回当前阶段评审详情文件的 `Round <round>` 对应章节。

### 4.4 判断是否需要用户确认

若存在 `needs-user-confirmation` 项，按以下规则处理：

1. 若本次调用未传入 `answered_questions`，则输出 `verdict=blocked`
2. 将待确认问题整理为一次性问题清单
3. 优先列出阻塞性问题
4. 每条问题给出建议选项或推荐方案
5. 等待调用方或用户补充答案后，再次以 `mode=process-feedback` 调用本 skill

待确认问题的结构至少包含：

- `question_id`：如 `R2-Q1`
- `round`：所属轮次
- `source_comment_id`：来源评论 ID
- `scope`：父单 / 子单及对应 ID
- `blocking`：`yes / no`
- `question`：待确认问题正文
- `options`：可选答案列表（A/B/C...）
- `recommended_option`：推荐选项
- `impact`：影响范围
- `answer`：用户回答（首次 blocked 时为空）

调用方映射到 `questions.md` 时，应保留以上字段，至少保证 `question_id`、`round`、
`blocking`、`question`、`options`、`answer` 可稳定回传。

若已传入 `answered_questions`，则先将其与待确认问题按 `question_id` 对齐，再继续进入修订或回复阶段。

### 4.5 判断本轮结论

#### 情况 A：无有效评论

若当前轮次没有新的有效评论，输出：

- `verdict=waiting_review`
- 建议继续等待评审人

不做任何写回。

#### 情况 B：满足通过规则

若当前轮次出现任一评审人的 `approve` 且 `是否阻塞=no` 评论，则：

1. 恢复父单处理人为 `previous_owner`
2. 将父单状态推进为 `approved`
3. 若本轮同步过子单处理人，也恢复对应子单处理人
4. 输出 `verdict=approved`

#### 情况 C：评论均为问题澄清，无需改动

1. 生成评论回复草稿
2. 在父单评论区答复评审人
3. 保持 `for approve`
4. 输出 `verdict=waiting_review`

#### 情况 D：存在待确认问题，需用户补充答案

输出：

- `verdict=blocked`
- 结构化待确认问题清单
- 当前阶段评审详情文件
- 建议调用方将问题稳定写入 `questions.md`，待用户回答后再次调用 `process-feedback`

#### 情况 E：存在需要修改的意见

输出：

- `verdict=needs_rework`
- 结构化评审意见清单
- 当前阶段评审详情文件
- 建议上层流程重新执行当前阶段对应的处理 skill，并把评审详情文件中最近一轮内容一起传入下一轮修订

同时需要把以下内容补写回当前轮次章节：

- `评审反馈摘要`
- `修订保护边界`

修订保护规则：

- 不得以“重新生成一份新文档”替代对现有父单 / 子单内容的修订
- 除非评审意见明确要求删除，否则评审详情文件最近一轮“基线快照”中已有的有效章节必须保留
- 若某一轮修订需要大幅改写，也应基于评审详情文件中的最近一轮基线做增量调整，而不是忽略旧文档上下文重新起草

### 4.6 更新本地元数据

更新以下字段：

- `round`
- `last_processed_comment_ids`
- `status`
- `last_action`
- 本轮处理摘要
- 当前阶段评审详情文件路径
- `previous_owner`
- `previous_children_owners`
- `current_review_owners`
- `approval_rule`

### 4.7 输出

输出结构化结果：

```markdown
## 评审意见已处理

- 父需求：xxx
- 评审阶段：evaluation
- 当前轮次：R2
- 结论：waiting_review / needs_rework / approved / blocked
- 下一步：继续等待评审 / 由上层流程重跑当前阶段 / 用户补充确认信息
```

## 5. `check-status` 模式

仅检查父需求状态和元数据：

- 父单 `v_status=approved` → 恢复处理人为 `previous_owner` 后，输出“评审通过”
- 父单 `v_status=for approve` → 输出“评审中”
- 其他状态 → 输出“状态异常，需人工确认”

本模式不读取评论、不修改单据、不更新本地文件。

## 结束条件

以下任一条件满足时，本次评审处理可视为结束：

1. 满足 `any_one_reviewer_approve` 规则后，父需求状态被本 skill 推进为 `approved`
2. 用户明确要求终止当前评审
3. 连续 3 轮以上出现相互冲突且无法收敛的意见，升级为用户人工裁决

> 评审完成信号以父单状态为准，不以“评论清空”或“子单都已回复”作为自动结束条件。

## 评论协议建议

为提高自动化稳定性，发起评审时应尽量引导评审人使用以下格式：

```text
结论：approve / change_required / question
范围：父单 / 子单 #12345
是否阻塞：yes / no
反馈内容：...
```

若评审人未按模板评论，也应尽量兼容自然语言，只是在输出中标注“非结构化评论，需人工复核”。

## 错误处理

| 错误场景 | 处理方式 |
|---------|---------|
| TAPD MCP 不可用 | 终止执行，提示用户检查 MCP 配置 |
| 父需求不存在 | 终止执行并提示用户确认需求 ID |
| 父需求未写回 description | 终止执行，提示先运行澄清或拆单 skill |
| `reviewers` 为空 | 终止执行，要求用户指定评审人 |
| 评论读取失败 | 重试一次，仍失败则输出当前状态并提示人工检查 |
| 评审意见冲突严重 | 输出冲突摘要，请用户决策 |
| 子单缺失或部分查询失败 | 继续处理父单，输出异常子单列表 |

## 参考文件

| 文件 | 用途 | 何时读取 |
|------|------|---------|
| `../tapd-story-evaluation/references/requirement-doc-template.md` | 子需求文档结构参考 | 分析拆单评论时 |

## 产出

- TAPD 父需求和子需求评论中的评审发起记录
- `docs/reqs/<父需求短ID>/<phase>-review-meta.yaml` 评审元数据
- `docs/reqs/<父需求短ID>/<phase>-review.md` 评审详情文件
- 结构化评审意见清单
- 评审结论：`waiting_review` / `needs_rework` / `approved` / `blocked`
- 满足单人通过规则后由本 skill 自动将父单状态推进为 `approved`

## 使用示例

```
用户输入：评审需求 12345 的澄清结果，评审人 @alice @bob

系统处理：
1. 读取父需求 12345 当前内容和状态
2. 将父单流转为 for approve
3. 在父单评论区发起 R1 评审并 @alice @bob
4. 写入本地 review 元数据
5. 输出“已发起评审”
```

```
用户输入：处理需求 12345 的拆单评审意见，评审人 @alice @bob

系统处理：
1. 读取父单与子单评论
2. 识别本轮新增评论并汇总意见
3. 输出 verdict=needs_rework 与结构化意见清单
4. 由用户或上层 pipeline 决定是否重跑当前阶段
```
