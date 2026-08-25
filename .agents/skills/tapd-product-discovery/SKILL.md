---
name: tapd-product-discovery
slug: tapd-product-discovery
version: 1.0.0
description: |
  Use when work starts before a complete PRD exists, including 产品前置,
  产品调研, 用户调研, 竞品分析, PRD, 原型, 想法建单, 老板需求,
  需求来源, 产品父单, 角色拆单, 设计子单, 前端子单, 后端子单,
  页面原型 Spec, BKUI 原型, HTML 原型, 原型评审,
  product discovery, product research, PRD drafting, or role story split.
metadata:
  requires:
    mcps: ["tapd"]
    skills: ["tapd-story-clarification", "tapd-story-evaluation", "tapd-story-pipeline"]
---

# TAPD 产品前置流程

## 概述

处理 PRD 之前的需求工作：想法记录、需求来源整理、用户调研、竞品分析、业务分析、PRD、页面原型 Spec、BKUI HTML 原型、原型评审和产品评审。流程从产品父单开始，按“想法记录 -> 产品调研 -> PRD -> 页面原型 Spec -> BKUI HTML 原型 -> 原型评审 -> 产品评审 -> 角色拆单”推进；产品方案通过后再按实际需要创建设计、前端、后端子单。

前端和后端子单创建后，继续使用已有研发流程处理；它不替代 `tapd-story-clarification`、`tapd-story-evaluation`、`tapd-story-pipeline`。

## 适用场景

- 用户只有一个想法、老板需求、用户反馈、竞品观察或业务问题，还没有完整 PRD。
- 用户希望先做用户调研、竞品分析、业务分析或原型，再整理成 PRD。
- PRD 已准备好，需要生成页面原型 Spec、BKUI HTML 原型，供产品、设计、前端评审。
- 已有产品方案通过评审，需要按设计、前端、后端拆 TAPD 子单。
- 需要说明产品、设计、前端、后端分别使用哪些 TAPD 状态。

## 不适用场景

- PRD、设计稿和接口文档都已确认，只需要研发实现：使用 `tapd-story-pipeline`。
- 已有标准化需求文档，只需要澄清或补充验收标准：使用 `tapd-story-clarification`。
- 已有父需求，需要按用户故事拆分和评分：使用 `tapd-story-evaluation`。
- 只处理代码、构建、发布或 Code Review：使用对应研发 Skill。

## 前置条件

- TAPD MCP 服务可用。
- 用户提供 workspace_id，或项目根目录 `project.json` 中存在 `workspace_id`。
- 创建或更新 TAPD 单据前，先向用户展示计划变更，确认后再调用 MCP。

## 输入

| 参数 | 来源 | 必需 | 说明 |
|------|------|------|------|
| 需求来源 | 用户输入 / TAPD 父单 | 是 | 想法、老板需求、用户反馈、竞品观察、已有草稿等 |
| 产品父单 ID | 用户输入 / 新建 | 否 | 已有父单则复用；没有则创建 |
| workspace_id | 用户输入 > project.json | 是 | TAPD 工作空间 ID |
| owner | 用户输入 > project.json | 否 | 产品负责人或当前处理人 |
| 背景知识 | 用户指定 > AGENTS.md 自动查找 | 否 | 业务说明、用户资料、竞品资料、架构约束、前后端规范等 |

## 状态流转

### 产品父单

```
backlog -> doing -> for approve -> approved -> done
```

| 状态 | 含义 |
|------|------|
| `backlog` | 需求来源已进入 TAPD，但还没有开始产品调研 |
| `doing` | 正在做调研、分析、原型或 PRD |
| `for approve` | PRD 或产品方案等待评审 |
| `approved` | 产品方案通过，可以拆分角色子单 |
| `done` | 产品侧交付完成，角色子单已创建或确认无需创建 |

### 设计子单

```
backlog -> todo(已规划) -> designing -> design approved -> done
```

| 状态 | 含义 |
|------|------|
| `backlog` | 设计子单已创建，等待安排 |
| `todo(已规划)` | 设计负责人、范围和排期已确认 |
| `designing` | 设计师正在产出交互、视觉、原型或设计说明 |
| `design approved` | 设计产物已通过评审，可供研发使用 |
| `done` | 设计材料已交付并归档 |

### 前端和后端子单

```
backlog -> for approve -> approved -> todo(已规划) -> doing -> for test -> tested -> for gray -> grayed -> done
```

前端和后端子单只记录各自研发进度，不再放产品调研和 PRD 评审内容。

## 执行流程

### 1. 判断当前入口

先判断用户处于哪一段：

| 用户输入 | 处理方式 |
|----------|----------|
| 只有想法或需求来源 | 创建或定位产品父单，进入产品前置 |
| 正在做调研 / 竞品 / PRD | 复用产品父单，推进到 `doing` |
| PRD 等待评审 | 回写产品文档，推进到 `for approve` |
| 产品方案已通过 | 推进到 `approved`，进入角色拆单 |
| 只要求研发实现 | 不走这个流程，提示使用现有研发流程 |

### 2. 准备产品父单

#### 2.1 确定 workspace_id 和 owner

优先级：
1. 用户显式提供
2. `project.json`
3. 询问用户

#### 2.2 已有产品父单

用户提供产品父单 ID 时，调用 TAPD MCP `stories_get` 读取单据：

```
workspace_id: <workspace_id>
id: <产品父单 ID>
with_v_status: "1"
```

读取后检查：
- 单据是否存在
- 是否为父单或可作为父单
- 当前状态是否适合继续推进

#### 2.3 没有产品父单

没有父单时，先整理创建计划并请用户确认。确认后调用 `stories_create`：

```
workspace_id: <workspace_id>
name: <需求名称>
description: <产品父单初始描述>
with_v_status: "1"
v_status: "backlog"
owner: <owner>
creator: <owner>
priority_label: <High|Middle|Low>
created: <当前时间>
```

父单初始描述使用 `references/product-requirement-template.md`。

### 3. 产品调研与 PRD

执行前读取 `references/product-discovery-guide.md`，按需求复杂度选择要补的内容：

- 需求来源和触发背景
- 目标用户和使用场景
- 当前问题和业务影响
- 用户调研结论
- 竞品或替代方案分析
- 项目内约束
- 产品目标和衡量指标
- 原型、流程或交互草图
- PRD 范围、验收标准和非目标

如果用户还没有开始调研，先输出调研提纲并推进产品父单到 `doing`。如果用户已经提供调研材料，直接整理为产品文档。

产品文档保存到：

```
docs/reqs/product/<产品父单短标题>.md
```

文件名从需求名称提炼，保留 4-12 个中文字符；重名时追加需求短 ID。

### 4. 页面原型 Spec 与 HTML 原型

PRD 完成后，先判断是否需要页面原型。

需要页面原型的场景：
- 新增或调整页面、表格、表单、弹窗、抽屉、图表
- 需要产品、设计、前端在拆单前确认页面结构
- PRD 中已有多个页面或复杂交互

可跳过的场景：
- 纯后端能力
- 纯配置或脚本任务
- 只改文案且不影响页面结构
- 已有可复用设计稿或原型链接

需要时读取 `references/prototype-spec-template.md`，生成页面原型 Spec，保存到：

```
docs/reqs/product/<产品父单短标题>.prototype.md
```

用户确认页面原型 Spec 后，读取 `references/bkui-html-prototype-guide.md`，生成 BKUI HTML 原型，保存到：

```
docs/reqs/product/prototypes/<产品父单短标题>-preview.html
```

原型评审通过后，再进入产品评审和角色拆单。跳过页面原型时，要在产品父单中记录原因。

### 5. 产品评审

当 PRD 已准备好，且原型完成或已记录跳过原因：

1. 向用户展示即将回写的产品文档路径和状态变更。
2. 用户确认后，读取本地产品文档内容，调用 `stories_update` 回写到产品父单 `description`。
3. 将产品父单状态推进到 `for approve`。

评审通过后，用户可以直接说“产品评审通过”或提供评审结论。确认后将产品父单推进到 `approved`。

### 6. 角色拆单

产品父单为 `approved` 后，读取 `references/role-ticket-splitting.md`，判断需要哪些子单。

| 子单 | 创建条件 |
|------|----------|
| 设计子单 | 需要交互流程、视觉稿、原型细化、信息架构或设计评审 |
| 前端子单 | 需要页面、组件、交互、表单、前端状态、前端联调或埋点 |
| 后端子单 | 需要接口、数据模型、权限、状态机、后台任务、消息或外部系统集成 |

简单需求可以跳过某些角色。每个跳过项都要在产品父单中写明原因。

#### 子单创建字段

调用 `stories_create` 前，先展示子单计划并请用户确认。确认后逐一创建：

```
workspace_id: <workspace_id>
name: <角色前缀 + 子单名称>
parent_id: <产品父单 19 位 ID>
description: <角色子单说明>
with_v_status: "1"
v_status: "backlog"
owner: <对应负责人，未知则使用 owner 并标注待分配>
creator: <owner>
priority_label: <继承产品父单或用户指定>
created: <当前时间>
```

角色前缀建议：
- `[设计] <需求名称>`
- `[前端] <需求名称>`
- `[后端] <需求名称>`

### 7. 回写父单并结束产品侧流程

子单创建完成后，更新产品父单描述，补充：

- 子单列表和链接
- 未创建角色子单的原因
- 研发侧下一步建议

用户确认后，将产品父单推进到 `done`。

产品父单 `done` 不代表功能已上线，只代表产品侧工作完成；上线状态由前端/后端子单继续记录。

### 8. 输出总结

输出格式：

```markdown
## 产品前置流程完成

| 项目 | 结果 |
|------|------|
| 产品父单 | #xxx，done |
| 产品文档 | docs/reqs/product/xxx.md |
| 页面原型 Spec | docs/reqs/product/xxx.prototype.md / 未生成：原因 |
| HTML 原型 | docs/reqs/product/prototypes/xxx-preview.html / 未生成：原因 |
| 设计子单 | #xxx / 未创建：原因 |
| 前端子单 | #xxx / 未创建：原因 |
| 后端子单 | #xxx / 未创建：原因 |
| 下一步 | 设计进入设计流程；前端/后端进入澄清、评估、规划、实现流程 |
```

## 与现有 Skill 的关系

| 阶段 | 使用 Skill |
|------|------------|
| 想法、调研、PRD、产品评审 | `tapd-product-discovery` |
| 前端/后端子单需求澄清 | `tapd-story-clarification` |
| 前端/后端子单拆分、评分 | `tapd-story-evaluation` |
| 迭代规划 | `tapd-iteration-plan` |
| 单需求研发实现 | `tapd-story-pipeline` |
| 迭代研发执行 | `tapd-iteration-runner` |

## 错误处理

| 场景 | 处理 |
|------|------|
| TAPD MCP 不可用 | 停止 TAPD 写入，输出本地文档和手动建单字段 |
| 缺少 workspace_id | 先从 `project.json` 读取；仍缺失则询问用户 |
| 产品父单不存在 | 提示用户重新确认 ID，或创建新父单 |
| 用户未确认状态变更 | 不调用 `stories_create` / `stories_update` |
| 子单创建失败 | 重试一次；仍失败则输出手动创建字段 |
| 无法判断是否需要某角色 | 标记为待确认，不自动创建该角色子单 |

## 参考文件

| 文件 | 用途 |
|------|------|
| `references/product-discovery-guide.md` | 产品调研、分析、PRD 整理指南 |
| `references/product-requirement-template.md` | 产品父单描述模板 |
| `references/prototype-spec-template.md` | 页面原型 Spec 模板 |
| `references/bkui-html-prototype-guide.md` | BKUI HTML 原型生成规范 |
| `references/role-ticket-splitting.md` | 设计、前端、后端子单拆分规则 |
