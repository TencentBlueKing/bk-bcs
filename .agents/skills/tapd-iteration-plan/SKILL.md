---
name: tapd-iteration-plan
slug: tapd-iteration-plan
version: 1.0.0
description: |
  TAPD 迭代规划技能。基于"approved"状态的需求池，结合需求依赖关系、size 规模、
  优先级进行全局编排，将合适规模的需求规划进入指定迭代。支持新建迭代和已有迭代
  重入两种模式，自动控制迭代总规模上限（默认 1000），通过有向无环图（DAG）分析
  保证依赖需求优先入迭代。
  Use this skill whenever the user mentions 迭代规划, 规划迭代, 编排迭代,
  iteration planning, 迭代编排, 需求入迭代, 排迭代, 拉迭代, 创建迭代,
  plan iteration, sprint planning, organize sprint,
  or any workflow involving TAPD iteration scheduling and story assignment.
metadata:
  requires:
    mcps: ["tapd"]
---

# TAPD 迭代规划

## 概述

本技能针对当前需求池中"approved"状态的需求进行全局优先级与规模评估，将规模合适
的需求按依赖关系编排进入指定迭代。规划过程遵循以下核心原则：

- **依赖优先**：被依赖的需求优先入迭代，避免规划完成后无法启动开发
- **规模控制**：单个迭代总 size 不超过上限（默认 1000）
- **重入安全**：已有迭代重入时仅追加未规划需求，不影响已入迭代的需求
- **价值导向**：在规模允许的前提下，优先编排高 size、高价值的需求

## 前置条件

- TAPD MCP 服务可用（需要 `iterations_get`、`iterations_create`、`stories_get`、`stories_update` 工具）
- 用户提供迭代名称
- 待规划需求需已完成评估（v_status="approved"）
- workspace_id 与 owner 可由用户提供，或从项目根目录 `project.json` 读取

## 输入

| 参数 | 来源 | 必需 | 说明 |
|------|------|------|------|
| 迭代名称 name | 用户输入 | 是 | 迭代名称，遵循命名规范 `iteration-vMAJOR.MINOR.PATCH` |
| workspace_id | 用户输入 > project.json | 是 | TAPD 工作空间 ID |
| owner | project.json | 是 | 创建人，用于新建迭代时填入 creator 字段 |
| 需求 ID 列表 | 用户输入 | 否 | 指定参与规划的需求 ID；不提供时自动从需求池筛选 |
| 迭代规模上限size_limit | 用户输入 | 否 | 默认 1000 |

## 迭代命名规范

迭代名称必须遵循以下规范：

- 固定 `iteration-` 开头
- 版本号采用社区语义化规范 `vMAJOR.MINOR.PATCH`
- 默认 `PATCH` 位固定为 `x`（如 `iteration-v0.1.x`、`iteration-v1.2.x`）

**示例**：`iteration-v0.1.x`、`iteration-v0.9.x`、`iteration-v1.0.x`

## 执行流程

### 1. 参数收集与环境准备

#### 1.1 校验迭代名称

用户输入消息中必须包含迭代名称 `name`。如果未提供，**直接报错并终止**：

```
错误：缺少迭代名称。请提供迭代名称（如 iteration-v0.1.x）后重试。
```

校验名称是否符合命名规范（`iteration-vX.Y.Z` 格式）。不符合时给出提示但继续
执行（仅作为警告）。

#### 1.2 确定 workspace_id 与 owner

按以下优先级确定：

1. 用户消息中显式指定 → 直接使用
2. `project.json` 中的对应字段 → 使用 `read_file` 读取并解析
3. 以上均无 → 询问用户

### 2. 迭代查询或创建

#### 2.1 精确查询迭代

使用 TAPD MCP `iterations_get` 精确查询：

```
调用参数:
  workspace_id: <workspace_id>
  name: <name>
```

#### 2.2 分支判断

根据查询结果走不同流程：

##### 分支 A：未找到迭代（**新迭代流程**）

使用 TAPD MCP `iterations_create` 创建新迭代：

```
调用参数:
  workspace_id: <workspace_id>
  name: <name>
  creator: <owner>
  startdate: <今日日期，YYYY-MM-DD 格式>
  enddate: <两周后的日期，YYYY-MM-DD 格式>
```

记录返回的 `iteration_id`。**新迭代默认总规模上限**为 `total_size = size_limit`。

##### 分支 B：找到迭代（**迭代流程重入**）

记录迭代 `iteration_id`。使用 TAPD MCP `stories_get` 查询迭代下所有需求：

```
调用参数:
  workspace_id: <workspace_id>
  iteration_id: <iteration_id>
  limit: 200
```

统计已入迭代需求的 `size` 字段总和 `workload_size`，计算剩余可规划规模：

```
total_size = size_limit - workload_size
```

如果 `total_size <= 0`，告知用户该迭代规模已满，询问是否仍要继续规划（继续则
按用户提供的上限或忽略规模限制）。

### 3. 候选需求列表构建

#### 3.1 自动筛选模式

如果**用户未指定需求 ID**，或者用户**明确要求规划更多需求**：

使用 TAPD MCP `stories_get` 拉取需求池中的"approved"需求：

```
调用参数:
  workspace_id: <workspace_id>
  v_status: "approved"
  with_v_status: "1"
  limit: 100
```

从查询结果中**仅保留满足以下两个条件的需求**：

- `children_id == "|"`（该需求没有子需求，避免规划"父需求壳子"）
- `iteration_id == ""` 或字段不存在（该需求尚未规划入任何迭代）

构建候选需求 ID 列表 `candidate_ids`。

#### 3.2 用户指定模式

如果**用户指定了需求 ID 列表**：

逐一使用 TAPD MCP `stories_get` 查询需求详情，过滤掉以下需求：

- 已存在 `iteration_id`（已入其他迭代）
- `children_id != "|"`（有子需求的父需求壳子）

被过滤的需求需告知用户，列出 ID 与过滤原因，最终保留的列表作为 `candidate_ids`。

#### 3.3 混合模式

如果用户既指定了需求 ID，又要求"再规划更多"，则两种模式合并：先处理用户指定
的需求，再追加自动筛选的需求。

### 4. 迭代规划

#### 4.1 截取评估列表

从 `candidate_ids` 中取**前 20 个 ID**作为本轮评估列表 `eval_list`。20 是单轮处理
的合理上限，避免一次性 DAG 计算规模过大。

#### 4.2 拉取详情与依赖扩展

对 `eval_list` 中每个 ID，逐一使用 TAPD MCP `stories_get` 获取详情：

```
调用参数:
  workspace_id: <workspace_id>
  id: <长ID>
  with_v_status: "1"
```

解析每个需求 `description` 字段中的"依赖关系"章节（参考 `references/dependency-analysis.md`
中的依赖识别规则），提取依赖的需求 ID。

对每个识别出的依赖需求：

- 使用 `stories_get` 获取详情
- 满足 `children_id == "|"` 且 `iteration_id == ""` 且 `v_status == "approved"` 的，**追加进 `eval_list`**
- 不满足条件的，记录为"外部依赖"，告知用户但不加入规划

#### 4.3 构建依赖 DAG 与排序

针对 `eval_list` 中所有需求，构建有向无环图（DAG）：

- 节点：需求 ID
- 边：A 依赖 B 时，B → A（B 必须先于 A 完成）

排序规则（参考 `references/dependency-analysis.md`）：

1. **拓扑序优先**：依赖链中位于上游的需求（被依赖方）排在前面
2. **同层 size 降序**：同一拓扑层级内，size 大的排在前面
3. **优先级兜底**：size 相同时，优先级高的排在前面

输出有序的入迭代列表 `ordered_list`。

#### 4.4 逐个加入迭代

按 `ordered_list` 顺序逐个加入迭代。对每个需求：

```
调用 stories_update：
  workspace_id: <workspace_id>
  id: <需求长ID>
  iteration_id: <iteration_id>
  v_status: "todo"
```

每加入成功一个：

- `total_size -= 该需求的 size`
- 该需求从 `candidate_ids` 中移除
- 记录到 `planned_list`（用于最终汇总）

**终止条件**（满足任一即结束）：

- `total_size <= 0`（迭代已满）
- `candidate_ids` 已清空（无更多候选需求）
- 当前需求的 size > 剩余 `total_size` 时，跳过该需求继续尝试下一个；
  若扫描完整个 `ordered_list` 中所有剩余需求均无法放入，加入结束。

### 5. 汇总输出

输出完整汇总：

```markdown
## 迭代规划完成

**迭代信息**：
- 名称：<name>
- ID：<iteration_id>
- 模式：<新建 | 重入>
- 起止日期：<startdate> ~ <enddate>
- 规模上限：<size_limit>
- 已用规模：<size_limit - total_size>
- 剩余规模：<total_size>

**本次规划入迭代需求**（共 N 个）：

| 序号 | 需求 ID | 名称 | size | 优先级 | 依赖 |
|------|---------|------|------|--------|------|
| 1 | xxx | xxx | 80 | High | - |
| 2 | yyy | yyy | 120 | High | xxx |

**迭代核心目标**
<基于划入迭代的需求列出本次迭代的产品、特性的核心价值目标>

**未规划需求说明**：
- 因迭代规模已满，剩余 K 个候选需求未入迭代：[ID 列表]
- 因外部依赖未满足跳过：[ID 列表]
```

## 错误处理

| 错误场景 | 处理方式 |
|---------|---------|
| 未提供迭代名称 | 直接报错终止，提示用户必须提供 name |
| TAPD MCP 不可用 | 终止执行，提示用户检查 MCP 配置 |
| `iterations_create` 失败 | 重试一次，仍失败则告知用户原因，输出待创建参数供手动创建 |
| `stories_update` 失败 | 重试一次，仍失败则跳过该需求，记录到失败列表 |
| 候选需求列表为空 | 告知用户当前需求池中无可规划需求，询问是否需要先评估需求 |
| 项目根目录无 project.json | 询问用户提供 workspace_id 和 owner |
| 迭代名称不符合规范 | 警告但继续执行，规划结束后提示用户调整 |
| DAG 出现循环依赖 | 告知用户存在循环依赖（列出涉及需求 ID），按 size 降序兜底处理 |

## 参考文件

| 文件 | 用途 | 何时读取 |
|------|------|---------|
| `references/dependency-analysis.md` | 依赖识别与 DAG 构建规则 | 执行步骤 4.2 - 4.3 时 |
| `references/iteration-naming.md` | 迭代命名规范详解与示例 | 校验/提示用户命名时 |

## 产出

- 已创建或复用的 TAPD 迭代单据
- 一批 `v_status="todo"` 且填入 `iteration_id` 的需求
- 控制台输出的规划汇总报告（含迭代信息、入迭代列表、未入迭代说明）

## 使用示例

**示例 1：新建迭代并自动规划**

```
用户输入：规划迭代 iteration-v0.2.x

系统处理：
1. 校验迭代名称符合规范
2. 从 project.json 读取 workspace_id 和 owner
3. iterations_get 查询未找到 → iterations_create 创建新迭代（今日 ~ 两周后）
4. 自动筛选 v_status="approved" 且 children_id="|" 且 iteration_id="" 的需求
5. 取前 20 个构建 DAG，按拓扑序+size 降序排序
6. 逐个调用 stories_update 入迭代，直到 total_size 用完
7. 输出规划汇总
```

**示例 2：用户指定需求规划入已有迭代**

```
用户输入：把需求 12345, 67890 规划进 iteration-v0.1.x

系统处理：
1. 校验迭代名称
2. iterations_get 找到已有迭代 → 统计已入迭代 size，计算剩余 total_size
3. 查询用户指定的两个需求详情，过滤掉已入迭代或有子需求的
4. 构建 DAG（含依赖扩展），排序
5. 逐个入迭代，输出汇总
```
