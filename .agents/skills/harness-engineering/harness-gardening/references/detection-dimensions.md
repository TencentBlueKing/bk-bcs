# 检测维度详细规则

本文档定义 harness-gardening 的检测维度，包含检测目标、扫描方法、判定条件、修复策略和示例。

## 维度概览

| 维度 | 检测目标 | 默认级别 | PR 模式触发条件 |
|------|---------|---------|----------------|
| 1. 路径有效性 | 文档中引用的路径是否存在 | P0 | 文件移动/删除 |
| 2. Skill 清单一致性 | `tooling.md` §1.0 与 `$SKILL_ROOT/*/SKILL.md`（仅顶层）及 `tool-dependencies.md` 已登记的 Skill 匹配；接入仓额外安装但未登记的跳过 | P0 | `$SKILL_ROOT/` 增删 |
| 3. 架构描述一致性 | 架构文档与代码结构匹配 | P1 | 分层目录变更 |
| 4. 技术规范同步 | 对比 `docs/standards/` 与 `../assets/standards/` 预设，覆写差异文件；检测新技术栈并补充预设 | P0 | go.mod/package.json 变更；docs/standards/ 增删 |
| 5. 词汇表完整性 | glossary.md 覆盖核心术语 | P0 | 全量模式 only |
| 6. 目录结构一致性 | 文档中的目录树与实际匹配 | P0/P1 | 文件增删 |
| 7. 工具依赖一致性 | tooling.md 基线节与 `tool-dependencies.md` 一致；无环境状态列；项目自有节与 git 跟踪对账 | P0/P1 | `$SKILL_ROOT/`、agents、git 跟踪的自研 Skill/MCP |
| 8. Dev Map 与 IDE 集成一致性 | graphify 可用时：本地 graph.json 与代码同步（结果不入库）；`README.md` + `.gitignore` 入库且未整目录 ignore；实体/fallback 下 graphify Rules（含 `.claude/rules/graphify.md`）+ CodeBuddy hook-guard；skill 不存在则 Skip | P0/Skip | 文件增删/移动/重命名；.cursor/.codebuddy/.claude/.agents/rules 与 graphify.* 变更 |
| 9. 工作流文档同步 | **已移除（Skip）**：harness 不再接入 workflow；仅清理 AGENTS.md 中残留的强制工作流措辞 | Skip / P0（残留清理） | — |
| 10. 业务规范一致性 | `docs/business-standards/` 存在时：索引登记文件是否存在 + 文档引用路径是否有效；仅报告不覆写正文；目录不存在时整个维度跳过 | P0/P1/Skip | `docs/business-standards/` 增删/修改 |
| 11. Standards Rules 一致性 | AGENTS 门闩、README 强制加载步骤、IDE `standards-*` Rules 与选用表裁剪一致；与 generating 共用 `sync-standards-rules.sh` | P0/Skip | AGENTS.md、docs/standards/、standards-* rules、.cursor/.codebuddy/.agents 变更 |
| 12. AGENTS 项目记忆保留 | 根 AGENTS 中经理解应保留的项目记忆是否被再生成误伤；规程见 `../references/agents-merge.md` | P0/Skip | AGENTS.md 变更；harness 再生成后全量 |

---

## 维度 1：路径有效性

### 检测目标

Harness 文档中引用的文件/目录路径是否在项目中真实存在。

### 扫描范围

- `AGENTS.md`
- `docs/harness/*.md`
- `docs/standards/README.md`
- `docs/glossary.md`

### 检测方法

```
1. 提取 Markdown 链接中的路径：[text](relative/path)
2. 提取反引号中的路径：`path/to/file`（仅匹配看起来像路径的模式）
3. 提取代码块中的目录树（├── / └── 格式）中的路径
4. 对每个路径检查 fs.existsSync(resolve(projectRoot, path))
5. 排除外部 URL（http:// / https://）
6. 排除锚点链接（#section）
7. 排除模板占位符（${...}）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 路径不存在，git log 可追溯到**重命名**（旧→新仍存在） | P0 | `git log --diff-filter=R --find-renames` 指向**现存**新路径 |
| 路径不存在，可通过文件名模糊匹配找到**唯一现存**候选 | P0 | `find` 唯一命中且文件存在 |
| 路径不存在，git 历史曾有该文件但工作区已删（删除/退役，非 rename） | **P1** | 不得当作「误删」自动恢复 |
| 路径不存在，无法确定新位置 | P1 | 无匹配或多个候选 |
| 引用目标文件正文含「已废弃」且路径仍存在 | Skip / 信息 | 不因「还在用」强制保留；清单侧见维度 2/7 |
| 路径存在 | PASS | — |

### 修复策略

**P0 自动修复（仅改文档引用，绝不恢复文件）：**
```
1. 通过 rename 记录或唯一 find 确定**现存**新路径
2. 在文档中替换所有旧路径为新路径
3. 如果是目录树中的条目，更新整行缩进和层级
```

**P1 生成方案（含「文件已从工作区删除」）：**
```
方案内容：
- 失效路径 + 出现位置（文件:行号）
- 若 git 仍能找到历史 blob：注明「曾存在，当前工作区无」
- 候选：A) 更新引用到替代物 B) 删除/改写引用 C) 用户确认后才允许恢复文件
- 默认建议：优先 A/B，**不要**默认恢复
```

### 显式禁止

```
禁止：
- git checkout / git restore / 从 HEAD 取出已删除路径「补回」工作区（除非用户在 P1 中明确确认「恢复文件」）
- 将「链接失效 + 历史曾有文件」直接升为 P0 自动恢复
- 以「治理仓例外」绕过上述禁止
```

### 示例

```
检测到: docs/harness/tooling.md §1.0 第 18 行引用 `skills/tapd-iteration/SKILL.md`
实际: 该文件已移动到 `skills/tapd-iteration-runner/SKILL.md`
证据: git log --diff-filter=R 显示重命名记录
级别: P0
修复: 替换路径 skills/tapd-iteration/SKILL.md → skills/tapd-iteration-runner/SKILL.md

检测到: AGENTS.md 引用 `agents/workflow-agent.md`，工作区无此文件，HEAD 历史仍有
级别: P1（退役/删除）
修复: 提案删除引用或改写替代入口；禁止 git restore agents/workflow-agent.md
```

---

## 维度 2：Skill 清单一致性

### 检测目标

`docs/harness/tooling.md` §1.0 Skill 清单表格是否与项目中实际存在的顶层 Skill 一致。

### 扫描范围

- `docs/harness/tooling.md` §1.0 中的 Skill 表格
- `$SKILL_ROOT/*/SKILL.md`（**仅顶层**，不含嵌套子 skill）

> `$SKILL_ROOT` 由 harness-gardening 前置条件中的项目类型识别确定。

### 登记规则

- **应登记**：`$SKILL_ROOT/{name}/SKILL.md`（顶层 skill，如开发仓的 `skills/code-review/SKILL.md`）且已在 `tool-dependencies.md` 中登记（接入仓）
- **不登记**：`$SKILL_ROOT/{parent}/{child}/SKILL.md`（嵌套子 skill，如 `skills/harness-engineering/harness-gardening/SKILL.md`）
- **接入仓额外安装的 Skill**：存在于 `$SKILL_ROOT` 但未在 `tool-dependencies.md` 中登记 → 不参与一致性检查（Skip），harness-generating 也不会将其写入 tooling.md
- 子 skill 的触发词合并到父 skill 在 tooling.md 的条目中（如 harness-engineering 合并 generation + gardening 触发词）
- 子 skill 目录应在 AGENTS.md **目录树**中列出，但不在 tooling.md §1.0 Skill 清单表中重复登记

### 检测方法

```
1. 从文件系统扫描 $SKILL_ROOT/*/SKILL.md（仅顶层，排除 $SKILL_ROOT/*/*/SKILL.md 嵌套子 skill）
2. 【接入仓专用】白名单过滤：读取 $SKILL_ROOT/harness-engineering/references/tool-dependencies.md，
   将步骤 1 结果与其中登记的 Skill 做交叉验证，只保留已登记的条目；
   未登记的 Skill（用户额外安装）记为 Skip，不参与后续比对；
   开发仓无需此步骤（$SKILL_ROOT 中的 Skill 均为项目维护内容，直接全量参与比对）
3. 从每个（过滤后的）SKILL.md 的 frontmatter 中提取 name、slug、description
4. 如无 frontmatter，从文件首行 # 标题提取名称
5. 读取 AGENTS.md，定位 Skill 清单表格（按 | Skill | 触发词 | 功能概要 | 格式识别）
6. 对比两侧清单
7. 单独检查 AGENTS.md 目录树是否包含已知子 skill 目录（维度 6 协作）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 文件系统中存在（且已在 tool-dependencies.md 登记）但 AGENTS.md 中无记录 | P0 | 新增 Skill 未注册 |
| 文件系统中存在但未在 tool-dependencies.md 登记（接入仓额外安装） | Skip | 不属于项目工作流 Skill，不检查 |
| AGENTS.md 中有记录但文件系统中不存在 | P0 | Skill 已删除未清理 |
| 名称/描述不一致 | P0 | SKILL.md 中的描述与 AGENTS.md 不匹配 |
| 触发词不确定 | P1 | 无法从 SKILL.md 中推导触发词 |

### 修复策略

**P0 自动修复——新增 Skill：**
```
1. 从 SKILL.md 提取 name、description、触发词
2. 确定该 Skill 属于哪个表格分区（根据目录位置判断）
3. 按字母序插入新行到对应表格
4. 格式：| **{name}** | "{trigger_words}" | {description} |
```

**P0 自动修复——删除 Skill：**
```
1. 从 AGENTS.md 的表格中移除对应行
2. 如果该行有加粗标记（**），同步检查是否有其他地方引用
```

**P0 自动修复——更新描述：**
```
1. 以 SKILL.md 中的描述为准，更新 AGENTS.md 表格中的对应字段
```

**P1 生成方案——触发词不确定：**
```
方案内容：
- Skill 名称和路径
- SKILL.md 中可能的触发词线索
- 建议用户补充触发词定义
```

### 示例

```
检测到: skills/micro-service-project-init/SKILL.md 存在
        但 AGENTS.md 中无 "micro-service-project-init" 条目
级别: P0
修复: 在 AGENTS.md 对应分区表格中新增行：
      | **micro-service-project-init** | "新建微服务项目"、"初始化项目脚手架" | 一键生成 go-micro 微服务项目骨架 |

检测到: skills/harness-engineering/harness-generating/SKILL.md 存在
        但 AGENTS.md Skill 清单中无 "harness-generating" 条目
级别: Skip（非偏差）
原因: 嵌套子 skill 不单独登记，触发词已合并到父 skill harness-engineering 条目中
```

---

## 维度 3：架构描述一致性

### 检测目标

`docs/harness/architectural-constraints.md` 中描述的分层架构、模块划分、依赖方向是否与代码实际结构一致。

### 扫描范围

- `docs/harness/architectural-constraints.md`
- 项目源码目录（`cmd/`、`internal/`、`pkg/`、`app/`、`src/` 等）
- `go.mod` / `package.json`（依赖声明）

### 检测方法

```
1. 解析架构文档中的层次定义：
   - 识别"分层"描述（如 "API 层 → Service 层 → Repository 层"）
   - 识别目录映射（如 "internal/service → Service 层"）
   - 识别依赖规则（如 "上层不可依赖下层"）

2. 扫描项目代码验证：
   - 文档中提到的目录是否存在
   - 目录内包的 import 关系是否符合声明的依赖方向
   - 是否有未在文档中注册的新模块/目录

3. 生成差异报告
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 文档中描述的目录在代码中不存在 | P1 | 可能是规划中或已重构 |
| 代码中有文档未描述的新模块 | P1 | 需人判断是否需要补充文档 |
| 代码中的依赖关系违反文档声明 | P1 | 需人判断是文档错还是代码违规 |
| 文档中的技术选型与实际不符 | P1 | 如文档写 MySQL 实际用 PostgreSQL |

### 修复策略

所有架构描述偏差均为 P1，生成方案：

```
方案内容：
- 偏差类型（目录缺失 / 新模块未注册 / 依赖违规 / 技术选型不符）
- 具体位置（文档行号 + 代码位置）
- 可能原因分析：
  A) 文档过时，需更新以反映当前代码
  B) 代码违规，需修复代码
  C) 规划中的架构，暂不处理
- 建议修复内容（如果是原因 A，给出具体的文档修改建议）
```

### 示例

```
检测到: docs/harness/architectural-constraints.md 第 28 行描述
        "internal/handler/ → API Handler 层"
        但项目中 internal/handler/ 目录不存在，
        实际 API Handler 在 internal/api/ 下
级别: P1
方案: 
  可能原因：A) 目录已重命名但文档未更新
  建议修复：将 "internal/handler/" 改为 "internal/api/"
```

---

## 维度 4：技术规范版本同步

### 检测目标

`docs/standards/` 中的规范文件是否与预设库 (`$SKILL_ROOT/harness-engineering/assets/standards/`) 保持一致（**入口 `.md` + 同名 `{stem}/` 分册目录**）。

> **策略约定（2026-08 修订）：**
> - 目标文件含 `<!-- harness:project-local -->` → **Skip 覆写**（P1 仅报告与预设漂移）。
> - 文件中「## 项目事实」章节（生成时填充）→ **永不覆写**。
> - 其余框架规则章节与预设 diff → 仍可用预设同步；但若预设含未替换占位符 `{{...}}` → 禁止 P0 覆写。
> - 定制长期规则应回写 `assets/standards/` 预设源。

### 扫描范围

- `docs/standards/*.md`（目标项目中的规范**入口**文件）
- `docs/standards/{stem}/`（大预设的分册目录，与 `{stem}.md` 成对）
- `$SKILL_ROOT/harness-engineering/assets/standards/`（预设库源：入口 + 分册）
- `$SKILL_ROOT/harness-engineering/assets/standards/index.yaml`（预设索引）

### 检测方法

```
1. 读取 docs/standards/ 下每个 .md 入口文件
2. 在 index.yaml 中查找同名预设
3. 逐字节对比项目入口与预设源入口；若预设库存在 assets/standards/{stem}/，对比 docs/standards/{stem}/ 下各分册（缺失/多余/差异均记 P0）
4. 记录差异（有差异 → P0；无同名预设 → PASS）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 与预设存在差异 | P0 | 用预设覆写 |
| 与预设完全一致 | PASS | — |
| 文件无对应预设 | PASS | 不是预设管理的文件，不处理 |

### 修复策略

**P0 自动修复：**
```
1. 从预设库复制对应入口 .md 到 docs/standards/（覆写）
2. 若预设库存在 assets/standards/{stem}/，递归复制到 docs/standards/{stem}/（覆写/补齐分册）
3. 记录覆写摘要供总结报告输出
```

### 示例

```
检测到: docs/standards/backend-go-micro.md
        与预设库 go-micro.md 存在差异（17 行不同）
级别: P0
修复: 用预设库覆写，输出差异摘要
```

---

## 维度 5：词汇表完整性

### 检测目标

`docs/glossary.md` 是否覆盖了 Harness 文档中出现的核心术语。

### 扫描范围

- `docs/harness/*.md`（五大组件文档）
- `docs/standards/README.md`
- `AGENTS.md`
- `docs/glossary.md`

### 检测方法

```
1. 从五大组件文档中提取候选术语：
   - 加粗文本：**术语**
   - 表格表头中的专有名词
   - 定义列表的标题行（: 开头的行前一行）
   - 首次出现且带括号注释的词：术语（English Term）

2. 过滤通用词汇（停用词表）：
   - 排除常见中文连接词/助词
   - 排除常见英文 the/a/is 等
   - 排除 Markdown 格式词（如代码关键字）

3. 读取 docs/glossary.md 现有条目

4. 对比：哪些候选术语不在 glossary 中
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 核心术语缺失（在 3+ 个文档中出现） | P0 | 高频使用但未定义 |
| 术语仅在 1-2 个文档中出现 | P0 | 添加为占位条目 |
| glossary 中有条目但文档中未使用 | Skip | 可能是规划中的概念 |

### 修复策略

**P0 自动修复：**
```
1. 在 glossary.md 的适当位置（按首字母排序）插入新条目
2. 格式：
   ### {术语}（{English}）
   <!-- TODO: 待补充定义 -->
3. 如果术语的英文名可从文档中提取，一并填入
4. 标记 TODO 提示用户后续补充完整定义
```

### 示例

```
检测到: "熵管理" 在 docs/harness/entropy-management.md、AGENTS.md、
        docs/harness/README.md 中共出现 5 次
        但 docs/glossary.md 中无对应条目
级别: P0
修复: 在 glossary.md 中插入：
      ### 熵管理（Entropy Management）
      <!-- TODO: 待补充定义 -->
```

---

## 维度 6：目录结构一致性

### 检测目标

`AGENTS.md` 和 `docs/harness/README.md` 中的目录树描述是否与项目实际文件结构一致。

### 扫描范围

- `AGENTS.md` 中的 "目录结构" 代码块
- `docs/harness/README.md` 中的目录树（如存在）
- 项目实际文件系统

### 检测方法

```
1. 从文档中提取目录树代码块（```...``` 包裹的树形结构）
2. 解析树形结构为路径列表：
   ├── dir1/       → dir1/
   │   ├── file.md → dir1/file.md
   └── dir2/       → dir2/
3. 获取实际文件系统对应范围的目录列表（排除 node_modules、.git 等）
4. 对比：
   - 文档中有但实际不存在的条目
   - 实际存在但文档中未列出的条目（仅关注同级目录，不要求穷举所有文件）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 实际新增了目录/关键文件 | P0 | 新增条目，补充到树中 |
| 实际删除了目录/文件，文档中仍存在 | P0（简单删除可追溯） / P1（涉及结构重组） | 根据 git log 判断 |
| 文档树的缩进/层级与实际不符 | P0 | 格式修正 |
| 大范围结构重组 | P1 | 需要重新设计目录树 |

### 修复策略

**P0 自动修复——新增条目：**
```
1. 确定新条目在树中的正确位置（按同级排序规则）
2. 插入新行，保持缩进和 ├──/└── 格式一致
3. 如果新条目是目录，带上 / 后缀
4. 如果新增是最后一项，更新前一项的 └── 为 ├──
```

**P0 自动修复——删除条目：**
```
1. 从树中移除对应行
2. 如果删除后子树为空，也移除父目录行
3. 更新最后一项的 ├── 为 └──
```

**P1 生成方案——结构重组：**
```
方案内容：
- 当前文档中的目录树
- 实际文件系统的目录树（同级别）
- 差异高亮
- 建议：直接提供重写后的目录树内容
```

### 示例

```
检测到: AGENTS.md 目录树中有 "skills/tapd-iteration/" 条目
        但该目录已不存在（被 tapd-iteration-runner 替代）
        git log 显示该目录在 3 次 commit 前被删除
级别: P0
修复: 从目录树中移除 "skills/tapd-iteration/" 行

检测到: 实际存在 "skills/harness-engineering/harness-generating/" 目录
        但 AGENTS.md 目录树中未列出
级别: P0
修复: 在 "skills/harness-engineering/" 子树中插入 "harness-generating/" 条目
      （子 skill 目录树需列出，但 tooling.md §1.0 Skill 清单表不重复登记）
```

---

## 维度 7：工具依赖一致性

### 检测目标

`docs/harness/tooling.md` 中「Agent 工具依赖」章节是否与权威清单 `tool-dependencies.md` 一致，
且是否反映当前项目实际启用的 Skill/场景（§五）。

### 扫描范围

- `docs/harness/tooling.md`（Harness 基线节 +「项目自有工具」节）
- `agents/*.md`（Agent 定义权威来源）
- 权威清单路径：`${SKILL_INSTALL_ROOT}/harness-engineering/references/tool-dependencies.md`

### 检测方法

```
1. 若 docs/harness/tooling.md 不存在 → Skip（尚未运行 harness-generating）
2. 形状检查：表头含「环境状态」或正文含「✅ 已就绪 / ❌ 未安装 / 未接入」→ P0（对齐 verify-tooling-shape）
3. 读取权威 tool-dependencies.md §一～§三、§四场景清单
4. 从 tooling.md 基线节提取 Skill/MCP/CLI（不含「项目自有工具」节）
5. 对比基线与权威清单；不探测本机是否就绪（属 harness-doctor / generating，非园艺职责）
6. 若存在「项目自有工具」节：表头应为 `名称 | 用途`；按名称与 git ls-files ∩ 安装布局对账（**含 monorepo 子树** `*/.agents/skills` 等，见 `project-owned-tools.md`）；未跟踪多装 → Skip；再生成不得清空该节
7. tooling.md 正文若出现「见 `…/project-owned-tools.md`」「权威清单路径：…」等**引用说明块** → P1（应删除；规则留在 skill references）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| tooling.md 含环境状态列或已就绪等状态单元格 | P0 | F6：状态不入库 |
| tooling.md 缺失基线 Skill/MCP 表或表格为空 | P0 | 文档已存在但章节未填充 |
| 权威清单新增依赖，基线节未同步 | P0 | 确定性条目遗漏 |
| tooling.md 基线含权威清单已移除的依赖 | P0 | 如 v3 场景仍写 `agent` CLI 为必需 |
| 项目自有节表头仍为旧式（含「仓库相对路径」等）或缺「用途」列 | P1 | 应改为 `名称 \| 用途` 并补用途 |
| 项目自有节名称在安装布局无对应 git 跟踪 Skill/MCP | P1 | 移出或先纳入 git |
| tooling.md 含权威清单/project-owned 等引用说明块 | P1 | 删除引用段，保留契约表 |
| 完全一致且形状合法 | PASS | — |

### 修复策略

**P0 自动修复——环境状态列：**
```
1. 删除「环境状态」列及「已就绪 / 未安装 / 未接入」单元格
2. 改为「检测方式 / 检测命令」列（见 harness-spec-template 模板 5）
3. 提示用户运行 harness-doctor；禁止把本机探测结果写回 tooling.md
```

**P0 自动修复——Skill/MCP/CLI（仅基线节）：**
```
1. 以权威 tool-dependencies.md 为准，重写基线表格（无环境状态列）
2. 同步 MCP / CLI 节与权威清单对应条目
3. 「项目自有工具」节保留不动（merge=ours）
```

**P0 自动修复——Agent 清单（§1.1）：**
```
新增 Agent 到 §1.1：
1. 从 agents/<name>.md frontmatter 提取 name、description
2. 在 tooling.md §1.1 表格中按字母序插入新行
   格式：| `<name>` | <定位> | <说明>（定义：`agents/<name>.md`）✅ |

从 §1.1 删除 Agent：
1. 从 §1.1 表格中移除对应行

更新 Agent 描述：
1. 以 agents/<name>.md frontmatter 中的描述为准，更新 §1.1 对应字段
```

**P1 生成方案：**
```
方案内容：
- 当前 tooling.md 场景勾选
- AGENTS.md 已登记 Skill 列表
- 建议勾选的场景（§五 A/B/C/…）
- 需用户确认后由 Agent 更新 tooling.md
```

### 示例

```
检测到: tooling.md §1.0 仍列出 CLI `agent` 为 v3 流水线必需
权威清单: 场景 A 已注明 v3 不需要 agent CLI（仅场景 A-v1 需要）
级别: P0
修复: 从 tooling.md 删除 agent 行，或移至「场景 A-v1」备注

检测到: 新增 skills/speckit-plan/SKILL.md，tooling.md 未含 speckit 结构依赖
权威清单: §二 Spec Kit 表已列出 speckit-* skill
级别: P0
修复: 在 §1.0 追加 speckit 相关行

检测到: agents/tapd-story-agent.md 或 agents/dispatcher.md 存在
        但 tooling.md §1.1 中缺少对应条目
级别: P0
修复: 在 §1.1 表格中分别登记稳定 Pipeline 与 Graph Engineering 入口：
      | `tapd-story-agent` | 稳定单需求流水线调度器 | 加载 tapd-story-pipeline（定义：`agents/tapd-story-agent.md`）✅ |
      | `dispatcher` | Graph Engineering 可选入口 | 用户显式启动后加载 graph-engineering（定义：`agents/dispatcher.md`）✅ |
```

---

## 通用规则

### 扫描排除列表

以下路径始终排除，不作为检测输入：

```
.git/
node_modules/
vendor/
dist/
build/
.cache/
*.log
```

### Commit Message 规范

P0 自动修复的 commit message 遵循以下格式：

```
docs(gardening): {action} {scope}

- {具体修改项 1}
- {具体修改项 2}
...

Auto-fixed by harness-gardening (mode={pr|full}, dimension={N})
```

示例：
```
docs(gardening): sync skill list in AGENTS.md

- Added micro-service-project-init to "微服务工程脚手架" table
- Removed deprecated tapd-iteration entry

Auto-fixed by harness-gardening (mode=full, dimension=2)
```

---

## 维度 8：Dev Map 与 IDE 集成一致性

### 检测目标

1. 本地 `docs/dev-map/graph.json` 是否存在且与当前代码同步（通过 graphify 增量更新；**结果默认不入库**）
2. 入库约定是否完整：`README.md` + `.gitignore`（对齐 `assets/dev-map.gitignore`），且**未**整目录 ignore `docs/dev-map/`
3. IDE 集成配置：按 install-to-target / `harness-ide-setup.sh` 布局检查 `graphify.mdc` / `graphify.md`（含实体 `.claude/rules`），以及 `.codebuddy/settings.json` hook-guard

### 前置条件

- dev-map 检测：仅在 `docs/dev-map/` 目录存在时执行；若目录不存在，dev-map 部分记录为 Skip。
- IDE 配置检测：仅在 graphify skill 存在时执行（两者共同依赖 graphify）。
- 必须先检测 graphify 是否可用（见工具检测步骤）。

### 工具检测

```
检查 $SKILL_ROOT/graphify/SKILL.md 是否存在

- 存在：继续执行 dev-map 检测与更新，同时执行 IDE 配置检测
- 不存在：
    整个维度 8 级别：Skip
    输出：提示"graphify skill 未找到（$SKILL_ROOT/graphify/SKILL.md 不存在），dev-map 与 IDE 集成配置检测跳过"
    退出维度 8，不阻断流程
```

### 扫描范围

- `docs/dev-map/graph.json`（本地存在性 / 新鲜度；**不**要求 tracked）
- `docs/dev-map/README.md`、`docs/dev-map/.gitignore`（须存在且宜 tracked）
- 根或其它 `.gitignore` 是否误忽略整个 `docs/dev-map/`
- `.agents/rules/graphify.mdc` / `graphify.md`（fallback 软链时）
- `.cursor/rules/graphify.mdc`（实体 `.cursor`）
- `.codebuddy/rules/graphify.md`（实体 `.codebuddy`）
- `.claude/rules/graphify.md`（实体 `.claude`；与 ide-setup / Claude 模板一致）
- `.codebuddy/settings.json`（hook-guard；软链时可能落在 `.agents/settings.json`）

### 检测方法

```
Dev Map 检测（docs/dev-map/ 存在且 graphify 可用时）：
1. 检查是否误 ignore 整个 docs/dev-map/（根或其它路径的整目录规则）：
   - 是 → P0，删除整目录 ignore；本目录用白名单 ignore（assets/dev-map.gitignore：`*` + `!README.md` + `!.gitignore`）
2. 检查 docs/dev-map/README.md：
   - 缺失或严重偏离模板 → P0，按 assets/dev-map-templates.md 写回；git add README.md
3. 检查 docs/dev-map/.gitignore：
   - 缺失，或不含白名单策略（须有独立行 `*`，且有 `!README.md` 与 `!.gitignore`）→ P0，复制 assets/dev-map.gitignore；git add .gitignore
4. 检查 docs/dev-map/graph.json 是否存在（本地）：
   - 不存在 → P0，执行全量生成到本地（不 git add 结果）
   - 存在且代码有变更 → P0，执行增量更新到本地（不 git add 结果）
5. 全量：调用 graphify skill /graphify . ；增量：/graphify . --update
6. **禁止** git add 本目录除 README.md / .gitignore 以外的文件（与 generating 第三步-C 一致）
7. 若用户曾把结果误提交：提示从索引移除（git rm --cached），不自动 force-push

IDE 集成配置检测（graphify skill 存在时执行；路径算法同 standards-compliance / harness-ide-setup.sh）：
1. 若任一 .$ide → .agents 软链（fallback）：
   检查 .agents/rules/graphify.mdc 与 .agents/rules/graphify.md 均存在
2. 若 .cursor 为实体目录：检查 .cursor/rules/graphify.mdc
3. 若 .codebuddy 为实体目录：检查 .codebuddy/rules/graphify.md
4. 若 .claude 为实体目录：检查 .claude/rules/graphify.md
5. 若 .codebuddy 存在（实体或软链）：检查 settings.json 含 hook-guard search + read
6. 若无 .cursor / .codebuddy / .claude / .agents：P0（无 IDE 集成）
   （仅有 .codex 不算 instruction rules 集成；graphify 门闩依赖 AGENTS.md，本项不因仅有 .codex 而 P0）
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| graphify skill 不存在 | Skip | 整个维度跳过，提示 skill 缺失，不阻断流程 |
| 误 ignore 整个 `docs/dev-map/` | P0 | 删除整目录 ignore，改用本目录白名单（仅 README + .gitignore） |
| `README.md` 或 `.gitignore` 缺失 / `.gitignore` 非白名单策略 | P0 | 从 assets 写回并 `git add` 这两项 |
| graphify skill 存在，本地 graph.json 不存在 | P0 | 自动全量生成到**本地**，不提交结果 |
| graphify skill 存在，本地 graph.json 过期（代码有变更） | P0 | 自动增量更新到**本地**，不提交结果 |
| docs/dev-map/ 目录不存在 | Skip | 项目尚未生成开发地图 |
| fallback 布局缺 `graphify.mdc` 或 `graphify.md` | P0 | 双格式未齐 |
| 实体 `.cursor` 缺 `graphify.mdc` | P0 | 文件缺失 |
| 实体 `.codebuddy` 缺 `rules/graphify.md` | P0 | 文件缺失 |
| 实体 `.claude` 缺 `rules/graphify.md` | P0 | 文件缺失（与 ide-setup 对称） |
| `.codebuddy/settings.json` 缺 hook-guard 条目 | P0 | jq 查询无匹配 |
| `.codebuddy` 存在但 `settings.json` 不存在 | P0 | 文件缺失 |
| `.cursor`、`.codebuddy`、`.claude`、`.agents` 均不存在 | P0 | 无任何 IDE instruction 集成配置 |

### 修复策略

**P0 自动修复——入库约定（README / .gitignore / 误整目录 ignore）：**
```
1. 若任意 .gitignore 忽略整个 docs/dev-map/ → 删除该规则
2. cp "$SKILL_ROOT/harness-engineering/assets/dev-map.gitignore" docs/dev-map/.gitignore（白名单：仅 README.md + .gitignore）
3. 按 assets/dev-map-templates.md 写回 docs/dev-map/README.md（若缺失或偏离）
4. git add docs/dev-map/README.md docs/dev-map/.gitignore
5. 不得 git add 本目录其它文件
```

**P0 自动修复——Dev Map 全量生成（本地 graph.json 不存在）：**
```
1. 调用 graphify skill：/graphify .
2. 确保 docs/dev-map/.gitignore 与 README.md 按上节就绪并 git add（仅这两项）
3. 不 git add graph.json 等结果；纳入 commit 的仅是约定文件（若有变更）
```

**P0 自动修复——Dev Map 增量更新（本地 graph.json 存在但代码有变更）：**
```
1. 调用 graphify skill：/graphify . --update
2. 仅刷新本地 graph.json；不检查、不 git add 结果文件
```

**P0 自动修复——IDE 集成配置（缺失或不完整）：**
```bash
bash "$SKILL_ROOT/harness-engineering/scripts/harness-ide-setup.sh" .
```
脚本幂等，直接重跑即完成创建或合并，读取 `[OK]` / `[MERGED]` / `[SKIP]` / `[WARN]` 状态行记录到 gardening-report。

**Skip——graphify skill 不存在：**
```
1. 记录到 gardening-report.md
2. 输出提示：graphify skill 未找到，dev-map 与 IDE 集成配置检测跳过
3. 不阻断流程，退出维度 8
```

### 示例

```
检测到：$SKILL_ROOT/graphify/SKILL.md 不存在
级别：Skip（整个维度 8）
输出：graphify skill 未找到，dev-map 与 IDE 集成配置检测跳过

检测到：docs/dev-map/graph.json 不存在（graphify skill 可用）
级别：P0
修复：调用 graphify skill /graphify .，结果留本地；git add 仅 README.md + .gitignore（若有变更）

检测到：代码文件有变更，本地 graph.json 存在
级别：P0
修复：调用 graphify skill /graphify . --update，只更新本地图谱，不提交结果

检测到：根 .gitignore 含 docs/dev-map/
级别：P0
修复：删除整目录 ignore，写入 docs/dev-map/.gitignore（结果级），保留 README

检测到：fallback（.cursor→.agents）但缺 .agents/rules/graphify.md
级别：P0
修复：bash "$SKILL_ROOT/harness-engineering/scripts/harness-ide-setup.sh" .
结果：[OK] Created .agents/rules/graphify.md（同时确保 .mdc）

检测到：实体 .claude 存在但缺 .claude/rules/graphify.md
级别：P0
修复：bash "$SKILL_ROOT/harness-engineering/scripts/harness-ide-setup.sh" .
结果：[OK] Created .claude/rules/graphify.md

检测到：.codebuddy/settings.json 存在，缺少 hook-guard read 条目
级别：P0
修复：bash "$SKILL_ROOT/harness-engineering/scripts/harness-ide-setup.sh" .
结果：[MERGED] .codebuddy/settings.json — added missing hook-guard entries
```

---

## 维度 9：工作流文档同步

> **Skip（2026-08）：** harness 已去掉 workflow 接入。本维度**不再**检测/注入/覆写 `docs/workflow.md`，**不再**要求 AGENTS.md 含「开发工作流」。
>
> **残留清理（全仓一视同仁，无「治理仓例外」）：** 若 `AGENTS.md` 仍含 `workflow-agent` /「不允许跳过工作流」/「## 开发工作流」强制段，记为 **P0**，删除该章节或相关强制句子（可改为指向 `flow-steward` / `develop-flow` 的**可选**一句，不得保留「不允许跳过」）。
> `docs/workflow.md`：不创建、不覆写、不要求存在；若存在且带 `<!-- workflow:custom -->`，视为人类可读参考，**不**因此保留 `workflow-agent` 强制段。
> `tooling.md` / 架构文档若仍将 `workflow-agent` 列为启用调度器 → **P1**（标废弃或改指向 flow-steward），不得用维度 1 把已删的 `agents/workflow-agent.md` 加回来。

### 判定条件与级别

| 检测项 | 条件 | 级别 | 说明 |
|--------|------|------|------|
| 9 | 默认 | Skip | 不同步 workflow 模板 |
| 9-clean | AGENTS.md 含 workflow 强制措辞 | P0 | 删除强制章节/句子（**含本治理仓**） |
| 9-registry | tooling/架构仍把 workflow-agent 当启用项 | P1 | 标废弃或改 flow-steward；禁止 restore 文件 |

### 修复策略

```
9-clean P0：
  - 删除 AGENTS.md 中「## 开发工作流」整节，及任何 workflow-agent /「不允许跳过工作流」句子
  - 不得从 assets/workflow-template.md 或 harness-spec-template 重新注入
  - 禁止「本仓是治理仓所以保留强制段」类例外

9-registry P1：
  - 提案：tooling §Agent 表 / 架构调度器表改为废弃说明或 flow-steward
  - 用户确认后再改；不得 git restore agents/workflow-agent.md
```

---

## 维度 10：业务规范一致性

### 检测目标

`docs/business-standards/`（用户自有业务规范空间）的**存在性与路径引用有效性**，且**仅报告、绝不覆写正文**。

### 扫描范围

- `docs/business-standards/README.md`（索引）
- `docs/business-standards/**/*.md`（业务规范文件，含 frontmatter 与正文中的路径引用）

### 判定条件

- **检测项 1（存在性）**：`docs/business-standards/README.md`「规范索引」表中登记的每个业务规范文件是否在磁盘存在。缺失 → **P0**，仅报告缺失，**不创建、不覆写正文**。
- **检测项 2（路径有效性）**：`docs/business-standards/**/*.md` 内 frontmatter/正文中引用的项目内路径是否真实存在。复用**维度 1**（本文档「维度 1：路径有效性」小节）的路径有效性判定语义，失效 → P0（被引用文档缺失）/P1（弱引用），语义与维度 1 一致。
- **目录不存在**：整个维度记 **Skip**（语义参照维度 8 dev-map skill 不存在时的 Skip）。

### 显式禁止（边界，取舍 C）

- **不**比对任何预设库、**不**覆写/修改业务规范正文。
- **不**校验业务规范正文格式、frontmatter 字段合法性、tags/scenarios 取值。
- 对缺失文件仅报告 P0，**不**代为创建。

### 修复策略

- 检测项 1 缺失：报告「索引登记的 `<file>` 不存在」，建议用户补齐文件或从索引移除该行；agent 不代改正文。
- 检测项 2 失效：报告失效路径，处置同维度 1（提示修正引用或补齐被引用文件）。

### 示例

```
检测到：docs/business-standards/README.md 索引登记 order-naming.md，但文件不存在
级别：P0
处置：仅报告，建议用户补齐 order-naming.md 或从索引移除对应行（不代为创建）

检测到：docs/business-standards/order-naming.md 引用 docs/standards/naming.md，路径不存在
级别：P0
处置：报告失效引用（复用维度 1 判定），提示修正
```

---

## 维度 11：Standards Rules 一致性

### 检测目标

接入仓「编码前必读」门闩、`docs/standards/README.md` 强制加载步骤，以及 IDE `standards-*` Rules 是否与「当前项目选用」裁剪一致。  
算法权威：`../references/standards-compliance.md`；**唯一**修复入口：`../scripts/sync-standards-rules.sh`（与 harness-generating 第三步-F 相同，禁止另写路径/裁剪逻辑）。

### 扫描范围

- `AGENTS.md`
- `docs/standards/README.md`（「当前项目选用」「Agent 加载步骤（强制）」）
- `.agents/rules/standards-*`、`.cursor/rules/standards-*`、`.codebuddy/rules/standards-*`、`.claude/rules/standards-*`
- Codex：不检查 `.codex/rules/*.md`（非 instruction 格式）；门闩以 `AGENTS.md` 为准

### 判定条件

| 条件 | 级别 | 修复 |
|------|------|------|
| 有 `docs/standards/README.md` 但 AGENTS 无「编码前必读」/「门闩」 | P0 | 按 `../assets/harness-spec-template.md` 插入门闩短段 |
| README 无「Agent 加载步骤（强制）」 | P0 | 按模板补步骤节 |
| README 无「加载预算」且无「按节」 | P0 | 按模板 7 补「加载预算」与按节加载步骤 |
| 有选用前端但缺少 `standards-frontend`（按 standards-compliance 路径） | P0 | 运行 `sync-standards-rules.sh` |
| 无选用前端却存在 `standards-frontend.*`（api/backend/security 同理） | P0 | 运行 `sync-standards-rules.sh`（脚本删多余） |
| rule 内标准路径与选用表不一致 | P0 | 重跑 sync |
| 无任何 IDE / `.agents` 布局 | Skip | 报告提示；**默认不**擅自 `ln -s` |

### 修复策略

```bash
bash "$SKILL_ROOT/harness-engineering/scripts/sync-standards-rules.sh" .
# AGENTS / README 门闩与步骤：按 harness-spec-template 对应节补齐
```

报告须写明：Rules 根、wrote/removed 列表、Skip 原因。

### 示例

```
检测到：选用 frontend-vue2.md，但 .agents/rules 无 standards-frontend.mdc/.md
级别：P0
修复：sync-standards-rules.sh → 写入双格式 frontend + gate + security（若已选用）

检测到：无前端选用，但存在 .cursor/rules/standards-frontend.mdc
级别：P0
修复：sync-standards-rules.sh → removed orphan
```

---

## 维度 12：AGENTS 项目记忆保留

### 检测目标

1. 根 `AGENTS.md` 中、对 Agent 入口决策仍必要的**项目记忆**（地图、硬约束、交付法、元策略、验证入口 / 四件套）是否在 harness 再生成或园艺过程中被误删/冲淡。  
2. **工作单元**：已存在的非根 `**/AGENTS.md` 是否在根被索引，并写明 nearest / 局部优先；园艺是否误改局部正文。  
权威规程：`../references/agents-merge.md`、`../references/agents-work-units.md`（**先理解再裁决**；禁止描点作为充分条件；默认不覆写局部）。

### 扫描范围

- 工作区根 `AGENTS.md`
- `git ls-files -- 'AGENTS.md' '**/AGENTS.md'` 得到的全部路径（工作单元清单）
- git 基线（`HEAD` 或 PR 基线）中根 `AGENTS.md` 及非根 `**/AGENTS.md`（若可得）

### 检测方法

```
1. 若工作区无根 AGENTS.md → Skip（尚未生成入口）；仍可单独报告工作单元清单供信息
2. 确认工作区含「编码前必读」门闩；缺失 → 交维度 11 处理，本维度不替代门闩补齐
3. 工作单元索引（全量 / PR 均可）：
   a. 执行发现算法，得到非根 **/AGENTS.md 清单
   b. 清单非空，且根无索引表/列表、且无「最近/局部优先于根」语义 → P1（提案补索引；角色列可「见该文件」）
   c. 清单为空 → 不要求索引节
4. 若可得 git 基线根 AGENTS：
   a. **再理解**基线与工作区：基线中承担四件套/入口记忆的内容，是否明显变弱或消失而短头/门闩仍在
   b. 若是 → P0（项目记忆误伤）
   c. 维度 9-clean 允许删除的强制 workflow 措辞不算误伤
5. 局部正文误改（相对基线）：
   a. 工作区相对基线**改写了非根 **/AGENTS.md**，且会话/指令中无用户明确要求改该子树 → P1（或证据确凿的误伤记 P0）
   b. 报告「禁止园艺改局部」；**默认不**自动 git restore 第三方局部正文（与维度 1 红线一致：优先提示；用户在 P1 中确认后方可恢复）
6. 禁止仅凭「某关键字行数下降」或「缺少某 HTML 注释」自动判定；启发式只触发「必须再理解」
7. 若无 git 基线可对比根记忆 → 步骤 4 Skip；步骤 3 索引检查仍执行
```

### 判定条件

| 情况 | 级别 | 条件 |
|------|------|------|
| 理解后确认根入口记忆被误删/冲淡 | P0 | 短头在、记忆弱或无 |
| 存在非根 AGENTS 且根未索引 / 无 nearest 语义 | P1 | 提案补索引，不编造角色 |
| 误改非根 AGENTS 正文（无用户明确要求） | P1（确凿误伤可 P0） | 报告禁止改局部；默认不自动 restore |
| 仅清理了强制 workflow 措辞 | Skip / 维度 9 | 不算本维度失败 |
| 无法取得基线对比（仅记忆对比） | Skip | 索引检查仍可判定 |

### 修复策略

```
P0（根记忆）：
1. 按 agents-merge.md 再理解基线内容，恢复被误判删除的 RETAIN-ENTRY 记忆
2. 保留工作区 harness 短头与门闩；合并去重 MERGE-HEADER
3. 不得恢复维度 9-clean 应删除的 workflow 强制段
4. 不得用模板 0-A 整文件覆写后假装已修复

P1（未索引）：
1. 按 agents-work-units.md 在根短头补「局部入口」表/列表 + nearest 优先句
2. 不读写局部 AGENTS 正文

P1/P0（误改局部）：
1. 向用户说明并提案恢复；用户确认前不自动改第三方局部文件
2. 确认后仅恢复被误改的非根路径，不借机「统一格式」批量改写
```

### 示例

```
检测到：基线 AGENTS 含详细任务路由与交付纪律；工作区仅剩 harness 短头+门闩
级别：P0
修复：理解基线后将仍必要的入口记忆合并回短头之后（保持原文结构优先）

检测到：仓库有 apps/foo/AGENTS.md，根无局部入口索引且无 nearest 句
级别：P1
修复：根短头补索引表 +「局部约定优先于根」

检测到：园艺会话改写了 packages/bar/AGENTS.md 且用户未要求
级别：P1
修复：提案恢复该文件；默认不自动 restore

检测到：工作区相对基线仅删除「不允许跳过工作流」句
级别：Skip（维度 9-clean）
```

---

## 通用规则

### 扫描排除列表

以下路径始终排除，不作为检测输入：

```
.git/
node_modules/
vendor/
dist/
build/
.cache/
*.log
```

### Commit Message 规范

P0 自动修复的 commit message 遵循以下格式：

```
docs(gardening): {action} {scope}

- {具体修改项 1}
- {具体修改项 2}
...

Auto-fixed by harness-gardening (mode={pr|full}, dimension={N})
```

示例：
```
docs(gardening): sync skill list in AGENTS.md

- Added micro-service-project-init to "微服务工程脚手架" table
- Removed deprecated tapd-iteration entry

Auto-fixed by harness-gardening (mode=full, dimension=2)
```

### 冲突处理

当多个维度对同一文件提出修改时：
1. 按维度编号顺序依次应用
2. 如果后续维度的修复与前序修复冲突，将冲突项升级为 P1
3. 所有 P0 修复在一个 commit 中提交（避免多次 commit 噪音）

