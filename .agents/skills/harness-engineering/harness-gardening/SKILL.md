---
name: harness-gardening
slug: harness-gardening
version: 1.6.5
description: |
  文档园艺——持续维护 harness-engineering 产出的文档与项目实际状态的一致性。
  支持 PR 轻量检查（代码变更后按需触发受影响维度）、全量深度扫描（十二维度完整检测）
  和定向检测（targeted，指定单个组件/维度进行针对性巡检）。
  完整扫描后：若存在 P1，向用户列出方案并请求确认，确认后与 P0 一并修复并提交；仅有 P0 时自动修复提交。
  由 harness-engineering 按维护类触发词路由调用，不独立触发。
---

# 文档园艺（Harness Gardening）

> **路径约定**：`references/` = 本子 skill 私有资源（`harness-gardening/references/`）；`../references/` = 父 skill 共享资源（`harness-engineering/references/`）；`../assets/` = 父 skill 预设库。
>
> **维度 1 红线（2026-08）：** 路径失效时只改文档引用或提 P1；**禁止** `git checkout`/`git restore` 把已删除路径加回工作区（除非用户在 P1 中明确确认恢复）。详见 `references/detection-dimensions.md`。
> **维度 9：** 清理 AGENTS 强制 workflow 措辞时**无治理仓例外**。
> **维度 12：** 根 AGENTS 项目记忆保留 + 工作单元索引（未索引 → P1；误改非根 AGENTS → 报告，默认不自动 restore）；规程见 `../references/agents-merge.md`、`../references/agents-work-units.md`。

## 核心职责

持续维护目标项目中 harness-engineering 产出的文档与项目代码/结构的一致性：

- 检测 Harness 文档（`AGENTS.md` / `docs/harness/` / `docs/standards/` / `docs/glossary.md` / `docs/dev-map/`）与实际状态的偏差
- P0 偏差（确定性、无歧义）自动修复
- P1 偏差（语义级、需人判断）向用户列出修复方案并请求确认；确认后与 P0 一并提交，拒绝/延迟的写入 proposals 文件备用

**`docs/standards/` 同步策略：单向从预设库同步，不支持用户定制。**
`docs/standards/` 下的文件以 `../assets/standards/` 预设为权威来源，gardening 负责将预设覆写到项目（**入口 `.md` + 同名 `{stem}/` 分册目录**）：
- 若项目文件与同名预设文件存在差异 → 用预设覆写项目文件（P0 自动修复）
- 若检测到新技术栈且预设库有对应文件 → 将预设复制到 `docs/standards/`（P0 自动修复）

因此 `docs/standards/` 下的文件**不应手动修改**，任何改动在下次 gardening 时都会被预设覆盖。
如需定制规范内容，应修改 `../assets/standards/` 中的对应预设文件。

## 执行清单

**开始执行前**，根据 `mode` 使用 TodoWrite 创建对应清单（全部状态 `pending`）：

**PR 模式（mode=pr）：**

| ID | 清单项 |
|----|--------|
| `grd-pr-1` | 获取变更文件列表（调用 harness-changed-files.sh，切出点 diff 或 last-commit diff） |
| `grd-pr-2` | 推断受影响维度 |
| `grd-pr-3` | 汇总并处置（P0/P1 分类、P1 确认、修复、commit） |

`grd-pr-2` 完成后，为每个被触发的维度用 TodoWrite **追加** todo 项（插入在 `grd-pr-3` 之前），ID 格式 `grd-pr-dim-N`（N 为维度编号，如 `grd-pr-dim-1`、`grd-pr-dim-8`）。

**全量模式（mode=full）：**

| ID | 清单项 |
|----|--------|
| `grd-full-1` | 读取所有 Harness 文档 |
| `grd-full-3` | 汇总检测结果（P0 / P1 / Skip 分类） |
| `grd-full-4` | P1 确认（若有 P1；无则直接标记 completed） |
| `grd-full-5` | 执行修复（P0 自动 + 用户确认的 P1） |
| `grd-full-6` | 生成 gardening-report.md 并 commit |
| `grd-full-7` | 输出总结报告 |

`grd-full-1` 完成后，按 `references/detection-dimensions.md` 的维度列表，为每个维度用 TodoWrite **追加** todo 项（插入在 `grd-full-3` 之前），ID 格式 `grd-full-dim-N`。

**定向模式（mode=targeted）：**

| ID | 清单项 |
|----|--------|
| `grd-tgt-1` | 解析 target 参数，确定目标维度 |
| `grd-tgt-dim-N` | 执行目标维度 N 检测（N 为实际维度编号） |
| `grd-tgt-2` | 汇总并处置（P0/P1 分类、P1 确认、修复、commit） |

**每完成一步（含每个维度）**：立即调用 TodoWrite 将对应条目标记为 `completed`，再继续下一步。

## 前置条件

- 目标项目已运行过 harness-generating（存在 `AGENTS.md` 和 `docs/harness/` 目录）
- 当前处于目标项目的仓库根目录
- **执行前须识别项目类型**：读取 `../references/project-type-detection.md`，确定 `$SKILL_ROOT`。

识别项目类型后，执行以下两个自动化前置步骤：

**步骤 A：环境配置（幂等，每次执行自动完成）**

```bash
eval $(bash "$SKILL_ROOT/harness-engineering/scripts/harness-setup-git.sh")
# 输出 GITATTR_MODIFIED=true/false
```

- 自动检查并写入 `.gitattributes` 两条条目（缺失时追加）：
  - `docs/dev-map/graph.json merge=graphify`
  - `docs/harness/gardening-report.md merge=ours`
- 自动检查并注册 git config merge driver（`graphify merge-driver %O %A %B`）
- 若 `GITATTR_MODIFIED=true`，将 `.gitattributes` 纳入本次最终 commit（`git add .gitattributes`）

**步骤 B：报告路径与 diff 基线（每次执行自动确定）**

```bash
eval $(bash "$SKILL_ROOT/harness-engineering/scripts/harness-diff-base.sh")
# 输出 REPORT_PATH 和 DIFF_BASE 两个环境变量
```

- `$REPORT_PATH`：功能分支（`feature/issue-N`）→ `workflows/issue-N/gardening-report.md`；其他 → `docs/harness/gardening-report.md`
- `$DIFF_BASE`：功能分支 → 沿 HEAD 遍历找切出点（首个被其他远端分支包含的 commit）；其他 → 空（由 last-commit 字段决定）

后续所有"生成/更新 gardening-report.md"步骤均使用 `$REPORT_PATH` 作为写入目标。

## 执行流程

### PR 轻量检查（mode=pr）

```
1. 获取变更文件列表（增量 diff）
   执行：CHANGED=$(bash "$SKILL_ROOT/harness-engineering/scripts/harness-changed-files.sh" "$DIFF_BASE" "$REPORT_PATH")
   - 若输出含 UPGRADE_TO_FULL → 自动升级为全量扫描（执行 mode=full 流程）
   - 若输出为空 → 输出"无新变更，无需检查"并退出

2. 根据变更文件推断受影响维度：
   - 新增/删除 `$SKILL_ROOT/` 下目录 → 维度 2（Skill 清单）+ 维度 7（工具依赖）
   - 修改 `$SKILL_ROOT/harness-engineering/references/tool-dependencies.md` → 维度 7（工具依赖）
   - 修改 `agents/*.md`（开发仓）或 `{SKILL_ROOT}/../agents/*.md`（接入仓） → 维度 7（工具依赖）
   - 修改 go.mod / package.json → 维度 4（技术规范同步）[仅 code-project / mixed]：检测新增技术栈，对比并同步预设
   - 新增/删除 `docs/standards/` 下文件 → 维度 4（技术规范同步）：对比并同步预设
   - 新增/移动/删除文件 → 维度 1（路径有效性）+ 维度 6（目录结构）+ 维度 8（Dev Map 与 IDE 集成一致性）
   - 修改 internal/ 或分层相关目录 → 维度 3（架构描述）
   - （已移除）修改 `../assets/workflow-template.md` → 维度 9 已 Skip，不再触发
   - 新增/删除 `.cursor/` / `.codebuddy/` / `.claude/` 目录 → 维度 8（Dev Map 与 IDE 集成一致性）
   - 新增/删除/修改 `**/rules/graphify.mdc` / `**/rules/graphify.md` → 维度 8
   - 新增/删除/修改 `docs/business-standards/` 下文件 → 维度 10（业务规范一致性）
   - 修改根 `AGENTS.md` / `docs/standards/` / 任意 `standards-*.mdc` / `standards-*.md` → 维度 11（Standards Rules 一致性）+ 维度 12（AGENTS 项目记忆 / 工作单元索引；根 AGENTS 变更时）
   - 新增/删除/修改任意非根 `**/AGENTS.md` → 维度 12（工作单元索引与误改局部检测）
   - 新增/删除 `.agents/`（或与 Rules 相关的 `.cursor` / `.codebuddy`）→ 维度 11
   - 若无维度被触发 → 输出"无需检查"并退出

3. 仅运行被触发的维度
   读取 references/detection-dimensions.md 获取各维度的检测规则

4. 汇总并处置
   a. 按 P0 / P1 / Skip 分类所有检测结果
   b. 若存在 P1：向用户逐条列出 P1 修复方案，等待确认
      - 用户确认：纳入本次执行
      - 用户拒绝/延迟：追加到 docs/harness/gardening-proposals.md 备用
   c. 执行全部 P0 修复 + 用户确认的 P1 修复
   d. 生成/更新 $REPORT_PATH（last-commit 字段先填占位值 "pending"）
      git add <所有修复文件> $REPORT_PATH [.gitattributes（若步骤 A 有修改）]
      git commit -m "docs(gardening): ..."（新建独立 commit，不 amend 已有提交）
   e. 写入 last-commit（仅 $DIFF_BASE 为空时执行，即 master/非 issue 分支）：
      1. git rev-parse HEAD → 获取步骤 d 产生的 commit SHA
      2. 更新 $REPORT_PATH 中 "> last-commit: " 行的值为该 SHA
      3. git add $REPORT_PATH && git commit --amend --no-edit
   f. Skip：记录到报告但不处理
   g. 若无 P1（仅 P0）：跳过步骤 b，直接执行 c → d → e
```

### 全量深度扫描（mode=full）

```
1. 读取目标项目的所有 Harness 文档
   - AGENTS.md
   - docs/harness/*.md
   - docs/standards/*.md
   - docs/glossary.md
   - docs/dev-map/*.md（若存在）
   - docs/workflow.md（若存在）

2. 依次运行十二个检测维度
   读取 references/detection-dimensions.md 获取详细规则

3. 汇总检测结果，按 P0/P1/Skip 分类

4. 处置
   a. 若存在 P1：向用户逐条列出 P1 修复方案，等待确认
      - 用户确认：纳入本次执行
      - 用户拒绝/延迟：写入 docs/harness/gardening-proposals.md 备用
   b. 执行全部 P0 修复 + 用户确认的 P1 修复
   c. 生成 $REPORT_PATH（last-commit 字段先填占位值 "pending"）
   d. git add <所有修复文件> $REPORT_PATH [.gitattributes（若步骤 A 有修改）]
      git commit -m "docs(gardening): ..."（新建独立 commit，包含修复 + 报告）
      然后写入 last-commit（仅 $DIFF_BASE 为空时执行）：
      1. git rev-parse HEAD → 获取步骤 d commit 的 SHA
      2. 更新 $REPORT_PATH 中 "> last-commit: " 行的值为该 SHA
      3. git add $REPORT_PATH && git commit --amend --no-edit
   e. Skip：记录到 $REPORT_PATH 但不处理
   f. 若无 P1（仅 P0）：跳过步骤 a，直接执行 b → c → d

5. 输出总结报告给用户（已修复项、延迟 P1 项、Skip 项）
```

### 定向检测（mode=targeted）

```
1. 解析 target 参数，确定目标维度：
   - dim-N 格式（N=1~12）→ 直接取维度 N
   - 具名 target → 按以下映射转换：
     paths        → 维度 1（路径有效性）
     skills       → 维度 2（Skill 清单一致性）
     architecture → 维度 3（架构描述一致性）
     standards    → 维度 4（技术规范同步）
     glossary     → 维度 5（词汇表完整性）
     structure    → 维度 6（目录结构一致性）
     tooling      → 维度 7（工具依赖一致性）
     dev-map      → 维度 8（Dev Map 与 IDE 集成一致性）
     workflow     → 维度 9（工作流文档同步）
     business-standards → 维度 10（业务规范一致性）
     standards-rules → 维度 11（Standards Rules 一致性）
     agents-merge / agents-memory → 维度 12（AGENTS 项目记忆保留）
   - 无法识别 → 列出所有合法 target 值并退出

2. 仅读取该维度所需的 Harness 文档（按 references/detection-dimensions.md 中维度扫描范围定义），
   读取 references/detection-dimensions.md 获取该维度的检测规则

3. 执行目标维度检测

4. 汇总并处置（流程同全量模式步骤 4，新增 h：无偏差时不 commit）：
   a. 按 P0 / P1 / Skip 分类检测结果
   b. 若存在 P1：向用户逐条列出 P1 修复方案，等待确认
      - 用户确认：纳入本次执行
      - 用户拒绝/延迟：写入 docs/harness/gardening-proposals.md 备用
   c. 执行全部 P0 修复 + 用户确认的 P1 修复
   d. 生成/更新 $REPORT_PATH：
      - 在报告中注明本次为 targeted 扫描，记录 target 值与对应维度编号
      - last-commit 字段先填占位值 "pending"
      git add <所有修复文件> $REPORT_PATH [.gitattributes（若步骤 A 有修改）]
      git commit -m "docs(gardening): targeted <target> ..."（新建独立 commit）
   e. 写入 last-commit（仅 $DIFF_BASE 为空时执行）：
      1. git rev-parse HEAD → 获取步骤 d 产生的 commit SHA
      2. 更新 $REPORT_PATH 中 "> last-commit: " 行的值为该 SHA
      3. git add $REPORT_PATH && git commit --amend --no-edit
   f. Skip：记录到报告但不处理
   g. 若无 P1（仅 P0）：跳过步骤 b，直接执行 c → d → e
   h. 若无任何偏差：输出"目标组件 <target> 未检测到偏差"并退出（不 commit）
```

## 处置分级

| 级别 | 判定条件 | 处置方式 | 示例 |
|------|---------|---------|------|
| P0 | 确定性偏差，修复无歧义 | 自动修复；若同批次有 P1 则与确认的 P1 一并提交，否则单独 commit | 路径重命名、Skill 清单新增、词汇表补条目、standards 预设同步覆写 |
| P1 | 语义级偏差，需人判断 | 向用户展示方案并请求确认；确认后与 P0 一并提交，拒绝/延迟则写入 proposals 备用 | 架构描述不一致、新技术栈预设不在预设库中需降级处理 |
| Skip | 无法确定是否为偏差 | 记录到报告但不处理 | 文档描述的是规划中架构 |

## 产出文件

```
docs/harness/
├── gardening-report.md       # 非 workflow 场景报告（merge=ours 兜底，master 周期扫描用）
└── gardening-proposals.md    # 用户拒绝/延迟的 P1 方案（执行后由下次扫描清空）

workflows/issue-N/
└── gardening-report.md       # workflow 场景报告（随 PR diff 可见，路径隔离无冲突）
```

报告路径由前置条件步骤 B（`harness-diff-base.sh`）动态确定，写入 `$REPORT_PATH`。
产出文件的 Markdown 格式模板见 `references/output-templates.md`。

## 参考资源

| 文件 | 用途 | 何时读取 |
|------|------|---------|
| `../references/project-type-detection.md` | 项目类型识别与 $SKILL_ROOT 检测 | 前置条件检查时（必须最先） |
| `../scripts/harness-setup-git.sh` | 配置 .gitattributes 和 git merge driver（运行时路径：`$SKILL_ROOT/harness-engineering/scripts/`） | 前置条件步骤 A |
| `../scripts/harness-diff-base.sh` | 确定报告路径（$REPORT_PATH）和 diff 切出点（$DIFF_BASE） | 前置条件步骤 B |
| `../scripts/harness-changed-files.sh` | 获取本次巡检的变更文件列表 | PR 模式步骤 1 |
| `references/detection-dimensions.md` | 各维度详细检测规则和修复策略（文件开头有维度概览表） | 执行检测时 |
| `../references/agents-merge.md` | 根 AGENTS 先理解再合并（维度 12） | 维度 12 |
| `../references/agents-work-units.md` | 工作单元发现与根索引（维度 12） | 维度 12 |
| `../references/standards-compliance.md` | Standards 门闩与 Rules 同步算法（维度 11 / generating 共用） | 维度 11 |
| `../scripts/sync-standards-rules.sh` | 按选用表渲染/裁剪 IDE Rules | 维度 11 P0 修复 |
| `references/output-templates.md` | gardening-report 和 gardening-proposals 格式模板 | 生成产出文件时 |
| `../assets/standards/index.yaml` | 技术规范预设索引（维度 4 使用） | 检测规范版本时 |
| `../assets/harness-spec-template.md` | 文档结构模板（维度 6 参照） | 检测目录结构时 |
| `../assets/dev-map-templates.md` | 开发地图文档模板（维度 8 修复参照） | 修复 dev map 偏差时 |
| `../assets/dev-map.gitignore` | 开发地图结果级 ignore（维度 8 写回 `docs/dev-map/.gitignore`） | 修复误 ignore / 缺失 ignore 时 |
| `../references/best-practices.md` | 最佳实践（做法 9-12 直接相关） | 生成修复方案时参考 |
| `../references/tool-dependencies.md` | Agent 工具依赖权威清单（维度 7） | 对比 tooling.md 时 |
| `../assets/workflow-template.md` | **DEPRECATED** | 维度 9 已 Skip，不再使用 |

## 清单验收

在输出总结报告后，检查 TodoWrite 清单：

- **全部 `completed`** → 执行完毕，正常退出
- **有 `pending` / `in_progress` 项** → 立即补充执行对应步骤，直至清单全绿再退出
