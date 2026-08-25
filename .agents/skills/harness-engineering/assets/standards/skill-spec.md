# Skill 规范

> 本文档提炼自 Anthropic 官方 Skill Creator 与项目 Writing Skills，作为编写和维护 Skill 的权威参考。

---

## 一、什么是 Skill

Skill 是一份**可复用的技术/流程参考文档**，帮助未来的 Agent 实例找到并应用有效的方法、模式或工具。

**Skill 是：** 可复用的技术、流程、设计模式、工具参考指南

**Skill 不是：** 某次问题解决过程的叙事记录

### Skill 的三种类型

| 类型 | 说明 | 示例 |
|------|------|------|
| **Technique**（技术方法） | 有具体步骤的方法论 | `test-driven-development`、`systematic-debugging` |
| **Pattern**（思维模型） | 看待问题的方式 | `dispatching-parallel-agents`、`brainstorming` |
| **Reference**（参考文档） | API 文档、语法指南、工具手册 | `harness-spec-template` |

---

## 二、目录结构（Anatomy）

```
skill-name/
├── SKILL.md              # 必须，Skill 主文件
│   ├── YAML frontmatter  # name、description 必填
│   └── Markdown 正文     # 指令与说明
└── 可选资源/
    ├── scripts/          # 可执行脚本（确定性/重复性任务）
    ├── references/       # 按需加载的参考文档
    └── assets/           # 输出用文件（模板、图标、字体等）
```

**命名规则：** Skill 目录名和 `name` 字段只允许使用字母、数字、连字符（不允许括号、下划线等特殊字符）。

---

## 三、YAML Frontmatter 规范

```yaml
---
name: skill-name-with-hyphens
description: Use when [具体触发条件和场景]
---
```

### 必填字段

| 字段 | 要求 |
|------|------|
| `name` | 字母、数字、连字符；与目录名一致 |
| `description` | 触发条件描述；frontmatter 全部内容不超过 1024 字符 |

### description 字段规则

1. **以 "Use when..." 开头**，聚焦触发时机
2. **只写触发条件，不写内部流程**（见下方 CSO 章节）
3. **第三人称**书写（会被注入系统 prompt）
4. **500 字以内**为佳，保持精简

---

## 四、渐进式加载模型（Progressive Disclosure）

Skill 采用三层加载机制，按需消耗上下文：

```
第一层：元数据（name + description）
  └─ 始终在上下文，约 100 词
     ↓ Skill 触发时
第二层：SKILL.md 正文
  └─ 建议 500 行以内；接近上限时拆分子文档
     ↓ 正文中显式引用时
第三层：资源文件（references/、scripts/、assets/）
  └─ 按需加载，无大小限制；脚本可直接执行而无需加载
```

**关键实践：**
- SKILL.md 接近 500 行时，提炼摘要保留在正文，详情移至 `references/` 并附导航指引
- 大型参考文件（>300 行）需在 SKILL.md 中附目录
- 多变体 Skill（如多云部署）按变体拆分子文档，正文只负责选择逻辑：

```
cloud-deploy/
├── SKILL.md          # 工作流入口 + 平台选择
└── references/
    ├── aws.md
    ├── gcp.md
    └── azure.md
```

---

## 五、description 字段最佳实践（CSO）

`description` 是 Agent 决定是否加载某 Skill 的**唯一依据**。写好它直接决定 Skill 能否在正确时机被触发。

### 核心原则：只写触发条件，不写工作流摘要

**为什么不能写流程摘要？**  
经测试发现：当 description 包含工作流摘要时，Agent 会以描述为捷径直接执行，跳过 SKILL.md 正文的详细指令。这会导致 Skill 中精心设计的步骤被忽略。

```yaml
# 反例：含流程摘要，Agent 会走捷径跳过正文
description: Use when executing plans - dispatches subagent per task with code review between tasks

# 反例：过多流程细节
description: Use for TDD - write test first, watch it fail, write minimal code, refactor

# 正例：只写触发条件
description: Use when executing implementation plans with independent tasks in the current session

# 正例：只写触发条件
description: Use when implementing any feature or bugfix, before writing implementation code
```

### 关键词覆盖

在 description 和 SKILL.md 正文中植入 Agent 会搜索的词：
- **错误信息**：具体报错文本
- **症状词**：flaky、hanging、race condition、inconsistent
- **工具名**：实际命令、库名、文件类型
- **同义词**：timeout/hang/freeze、cleanup/teardown/afterEach

### 触发增强

当某类场景容易被漏触发时，可在 description 中加入主动提示：

```yaml
# 普通版
description: How to build a dashboard to display internal data.

# 增强版（避免漏触发）
description: How to build a dashboard to display internal data. Use this skill whenever
  the user mentions dashboards, data visualization, internal metrics, or wants to
  display any kind of company data, even if they don't explicitly ask for a 'dashboard'.
```

---

## 六、内容写作规范

### 指令风格

- 使用**祈使句**写指令（"Read the file"，而非"You should read the file"）
- **解释原因**，而非只写"必须做什么"——当前 LLM 理解"为什么"后执行更准确，比堆砌 MUST/NEVER 更有效
- 避免过度刚性约束；如果出现大写的 ALWAYS/NEVER，反思是否可以改写为原理说明

### 示例

- **一个好例子胜过五个平庸例子**
- 选择最相关的语言（测试技术 → TypeScript/JS；系统调试 → Shell/Python）
- 好例子的标准：完整可运行、注释解释"为什么"、来自真实场景、清晰展示模式

```markdown
## 提交信息格式

**示例：**
输入：Added user authentication with JWT tokens
输出：feat(auth): implement JWT-based authentication
```

### 输出格式定义

```markdown
## 报告结构
ALWAYS use this exact template:
# [标题]
## 执行摘要
## 关键发现
## 建议
```

### 篇幅控制

| 场景 | 目标字数 |
|------|---------|
| 高频加载的 getting-started 类 Skill | < 150 词 |
| 其他频繁加载的 Skill | < 200 词 |
| 一般 Skill | < 500 词（仍需精简） |

避免重复：不要在 Skill 中重述其他已引用 Skill 的内容；用交叉引用代替复制。

---

## 七、文件组织模式

### 自包含 Skill
```
defense-in-depth/
  SKILL.md    # 所有内容内联
```
适用于：内容适中、无大型参考文档

### 含工具脚本的 Skill
```
condition-based-waiting/
  SKILL.md         # 概述 + 模式
  wait-helper.ts   # 可复用代码
```
适用于：有可复用的脚本或工具

### 含大型参考文档的 Skill
```
pptx/
  SKILL.md       # 概述 + 工作流
  references/
    pptxgenjs.md   # 600 行 API 参考
    ooxml.md       # 500 行 XML 结构
  scripts/         # 可执行工具
```
适用于：参考材料体量过大，不适合内联

---

## 八、何时创建 / 何时不创建 Skill

### 适合创建

- 该技术对你来说并非直觉显而易见
- 可跨项目复用
- 模式具有普适性，不特定于某个项目
- 他人能从中受益

### 不适合创建

- 一次性解决方案
- 业界通行的标准实践（有更好的外部文档）
- 项目特定的约定（放 `AGENTS.md` 或项目规则文件中）
- 机械性约束（如果可以用 lint/validation 自动执行，就自动化它，而不是靠文档）

---

## 九、安全原则

Skill 内容不得包含：
- 恶意代码、漏洞利用代码
- 任何可能危害系统安全的内容
- 具有误导性意图的指令

**无意外原则：** Skill 的实际行为必须与其描述的意图一致，不能让用户感到被欺骗或意外。

---

## 参考来源

- [Anthropic Skill Creator](.cursor/skills/skill-creator/SKILL.md) — 官方 Skill 创建与评测工作流
- [Writing Skills](.agents/skills/writing-skills/SKILL.md) — 项目级 Skill 编写规范与 TDD 方法论
- [agentskills.io/specification](https://agentskills.io/specification) — Frontmatter 完整字段规范
