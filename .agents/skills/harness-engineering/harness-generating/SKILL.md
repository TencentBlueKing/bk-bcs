---
name: harness-generating
slug: harness-generating
version: 1.5.2
description: |
  Harness 规范文档生成——harness-engineering 的子 skill，负责项目上下文感知、
  信息收集、五大组件与技术规范文档的生成/修正、开发地图生成及总结报告。
  由 harness-engineering 按生成类触发词路由调用，不独立触发。
---

# Harness 规范文档生成

> **路径约定**：`references/` = 本子 skill 私有资源（`harness-generating/references/`）；`../references/` = 父 skill 共享资源（`harness-engineering/references/`）；`../assets/` = 父 skill 预设库。

## 执行清单

**开始执行前**，使用 TodoWrite 创建如下清单（全部状态 `pending`）：

| ID | 清单项 |
|----|--------|
| `gen-1` | 第一步：项目上下文感知（含项目类型识别、基本信息、已有 Harness 检测、环境工具检查） |
| `gen-2` | 第二步：信息收集与交互 |
| `gen-3` | 第三步：规范文档生成（AGENTS.md + 五大组件文档 + glossary） |
| `gen-3b` | 第三步-B：技术规范文档处理 |
| `gen-3c` | 第三步-C：开发地图生成（含 IDE 集成配置） |
| `gen-3d` | 第三步-D：工作流文档生成（**已移除，标记 completed 并 Skip**） |
| `gen-3e` | 第三步-E：初始化 `docs/business-standards/` 目录与索引骨架（幂等，已存在不覆写） |
| `gen-3f` | 第三步-F：同步 Standards IDE Rules（`sync-standards-rules.sh`，按选用裁剪） |
| `gen-4` | 第四步：定向修正（`targeted` 模式时必做，其他模式跳过并标记 completed） |
| `gen-5` | 第五步：工具依赖自审 |
| `gen-qa` | 质量检查（十二项逐一核对，在生成总结前执行） |
| `gen-6` | 第六步：生成总结报告 |

**每完成一步**：立即调用 TodoWrite 将对应条目标记为 `completed`，再继续下一步。

## 执行流程

### 第一步：项目上下文感知

扫描项目结构，收集以下信息作为规范生成的基础。

0. **项目类型识别**（必须最先执行，结果影响后续所有路径）

   读取 `../references/project-type-detection.md`，完成识别并确定 `$SKILL_ROOT`。

1. **项目基本信息**
   - 读取 `README.md`、`package.json`、`pyproject.toml` 等项目描述文件
   - 识别技术栈（语言、框架、构建工具）
   - 检测已有的文档目录结构（`docs/`、`AGENTS.md` 等）

2. **已有 Harness 组件检测**
   - 检查是否已有 `docs/harness/` 目录
   - 检查是否已有 `docs/standards/` 目录及其中的规范文件
   - 如已有规范文档，读取并理解当前状态，后续只做增量更新

3. **环境工具完备性检查**
   - **先运行** `../scripts/harness-doctor.sh`（或仓库内等价路径）收集 CLI / Skill 安装根 / 项目自有工具 present·absent；结果**只进总结报告「环境缺口」节**，**禁止**写入 `docs/harness/tooling.md` 的「环境状态 / 已就绪」列
   - 读取 `../references/tool-dependencies.md` 作为 Harness 基线数据源
   - 根据用户意图或项目已有 Skill，从 §四 选定一个或多个场景（如 v3 主线用「场景 A」，仅用澄清/评估用「场景 B」）
   - 对所选场景的检查清单逐项探测（按以下优先级）：
     - **MCP**：§一 各 MCP「环境检查」行——**立即发起 tool call**（执行调用，不是描述）；成功则记入总结报告缺口表，失败则记缺口。**不得**把就绪结果写进 tooling.md
     - **Skill / CLI**：以 `harness-doctor` 输出为准，可按场景补探测
     - **可选工具**（`docker`、`gh`、`python3` 等）：不主动检查，总结报告中注明"按需安装"
   - 生成「环境工具缺口」列表，**仅报告所选场景**涉及的缺口（未选场景不探测、不报告）
   - 缺口列表暂存，经第五步自审补全后在第六步总结报告中输出；**tooling.md 只保留契约表（所需工具 + 检测方式）**

### 第二步：信息收集与交互

根据第一步的感知结果，评估信息充分程度。对于信息不足的组件，读取
`../references/question-bank.md` 获取对应组件的提问清单，一次性向用户提出。

**信息充分度评估标准：**

| 组件 | 最低信息要求 |
|------|------------|
| 上下文工程 | 知识来源、文档结构、动态数据源 |
| 架构约束 | 分层结构、依赖规则、边界定义 |
| 熵管理 | 文档维护策略、技术债处理方式 |
| 工具能力 | 工具清单、接口规范、稳定性策略 |
| 执行与验证 | 任务流程、验证机制、可观测性方案 |

**交互原则：**
- 将所有待确认问题整理为一份结构化清单，一次性提出，避免碎片式追问
- 区分阻塞性问题（必须回答才能生成）和非阻塞性问题（可先用默认值）
- 用户未回答的非阻塞性问题，使用最佳实践默认值填充，并在文档中标注"待确认"

### 第三步：规范文档生成

读取 `../assets/harness-spec-template.md` 获取文档模板，结合收集到的信息生成规范文档。

**输出目录结构：**

```
AGENTS.md                                # 项目入口（渐进式上下文披露的第一层）
docs/
├── glossary.md                          # 词汇表（核心概念、术语、缩写定义）
├── harness/
│   ├── README.md                        # 总览与导航（渐进式上下文披露的第二层入口）
│   ├── context-engineering.md
│   ├── architectural-constraints.md
│   ├── entropy-management.md
│   ├── tooling.md
│   └── execution-verification.md
├── dev-map/
│   ├── README.md                        # 开发地图索引 + 维护规则矩阵
│   └── graph.json                       # 持久化图谱（支持增量更新与深度查询）
└── standards/
    ├── README.md                        # 导航 + Agent 加载策略 + 章节索引
    ├── skill-spec.md                    # skill-tooling 项目：Skill 编写规范（从预设同步）
    ├── security-bk-redlines.md          # code-project：代码安全三大红线（从预设同步）
    ├── quality-code-review.md           # code-project：代码评审规范（从预设同步）
    ├── frontend-{stack}.md              # code-project：前端技术栈规范
    ├── api-{stack}.md                   # code-project：接口协议规范
    └── backend-{stack}.md               # code-project：后端技术栈规范
docs/business-standards/                 # 用户自定义业务规范（gardening 永不覆写）
└── README.md                            # 业务规范索引 + frontmatter 元数据说明
```

> `docs/standards/` 下的文件均从 `../assets/standards/` 预设同步，内容不可手动修改。

披露层次：根 AGENTS（短头 + 局部入口索引 + 项目记忆）→ 工作单元 `**/AGENTS.md`（nearest 优先）→ harness/README + standards/README → 详细文档。

**关键生成规则：**
- AGENTS.md（根）：写根前按 `../references/agents-work-units.md` 执行 `git ls-files -- 'AGENTS.md' '**/AGENTS.md'`，产出工作单元清单；有非根条目则短头必须含「局部入口」索引 + **nearest / 局部优先于根**句，**默认不覆写局部 AGENTS 正文**；短头含概述、目录（二级）、关键规范、「编码前必读」门闩；**禁止**写入「开发工作流」/ `workflow-agent` /「不允许跳过」；**若已有根 AGENTS.md，按 `../references/agents-merge.md` 先理解（含四件套）再裁决**（`RETAIN-ENTRY` 等）——禁止跳过理解整文件覆写、禁止描点作为充分条件；短头 prose ≤80 行；`RETAIN-ENTRY` 与工作单元索引不因压行数删除；总结报告须含「AGENTS 理解」与「工作单元 AGENTS」节；Skill/MCP/工具依赖只进 `docs/harness/tooling.md`；模板 0-A 见 `../assets/harness-spec-template.md`
- L2 组件：`context-engineering` 知识来源含工作单元 AGENTS；`execution-verification` 验证命令对齐根/局部 AGENTS 与组件根，禁止虚构测试栈、禁止与「无单测勿声称」冲突；`architectural-constraints` 对已有分层/关系图只摘要+指针；禁止 L2 粘贴大段局部 AGENTS
- tooling.md Harness 基线 Skill：扫描 `${SKILL_INSTALL_ROOT}/*/SKILL.md`（仅顶层；安装根见 `../references/skill-install-root.md`），与 `../references/tool-dependencies.md` 交叉验证，**只列白名单内 Skill**
- tooling.md Harness 基线 MCP：以 `../references/tool-dependencies.md` §一为权威；**不**扫用户级 `~/.cursor`；个人额外 MCP 不列入基线
- tooling.md **项目自有工具**（算法见 `../references/project-owned-tools.md`，**算法文档路径勿写入 tooling.md**）：`git ls-files` ∩ 安装布局（**含 monorepo 子树**如 `apps/*/.agents/skills/*/SKILL.md`），且名不在白名单 → 写入/合并「项目自有工具」节，表头仅 **`名称 | 用途`**；用途从 `SKILL.md` `description` 压成一行，子树路径可前缀组件目录；未跟踪路径禁止写入；再生成**不得清空**该节已有行
- tooling.md **禁止写入引用说明块**：勿写「依据 / 见 `tool-dependencies.md` / `project-owned-tools.md` / `install-to-target.sh` / 权威清单路径」等引用段落；契约表本身即可，探活与准入规则留在 skill references
- **接入仓扫描结果只写 tooling.md 契约**：权威清单只读；**环境缺口只进总结报告 / `harness-doctor` stdout**，不得写入 tooling.md 状态列
- tooling.md **禁止**表头「环境状态」及单元格「已就绪 / 未安装 / 未接入」；交付前须跑 `../scripts/harness-verify.sh`
- **统一脱敏**：AGENTS.md / docs/harness/** / docs/glossary.md 禁止写入本机绝对路径（`/home/` `/Users/` `/data/go/` `/root/` `/tmp/`）、邮箱、人员英文名/个人标识表；联系人放 CODEOWNERS 或团队 wiki
- glossary.md：从五大组件提取核心术语，按分类组织，每条含中/英文名和定义
- 各组件文档：包含目标、原则、规范条目、实施指南、检查清单；信息不足处标注 `<!-- TODO: 待补充 -->`

生成各组件时，参考 `../references/best-practices.md`（§1–§5 对应五大组件）。

### 第三步-B：技术规范文档处理

在五大组件文档生成后，处理技术规范预设的选择与部署。

**首先根据 `$PROJECT_DOMAIN` 确定必选规范分支：**

| `$PROJECT_DOMAIN` | 必选规范 | 跳过规范 |
|-------------------|---------|---------|
| `code-project` | `security/bk-redlines`、`quality/code-review` | `skill-tooling` 类别 |
| `skill-tooling` | `skill-tooling/skill-spec` | `security`、`quality` 类别 |
| `mixed` | 全部必选规范 | 无 |

1. **处理必选规范**（按 `$PROJECT_DOMAIN` 决定范围）— 读取 `../assets/standards/index.yaml`：
   - `$PROJECT_DOMAIN == code-project` 或 `mixed`：找出所有 `detect: code-project` 的预设（当前为 `security/bk-redlines`、`quality/code-review`），复制对应**入口** `.md` 到 `docs/standards/`，**无需用户确认、无需技术栈匹配**
   - `$PROJECT_DOMAIN == skill-tooling` 或 `mixed`：找出所有 `detect: skill-tooling` 的预设（当前为 `skill-tooling/skill-spec`），复制对应**入口** `.md` 到 `docs/standards/`，**无需用户确认**
   - **分册目录**：若 `../assets/standards/{stem}/` 存在（`file: foo.md` → `assets/standards/foo/`），须**递归复制**整目录到 `docs/standards/{stem}/`（与入口一并部署；见 `../references/preset-management.md`）
2. **自动检测技术栈（必须跑脚本，支持 monorepo）** — 在目标仓根执行：

   ```bash
   bash "$SKILL_ROOT/harness-engineering/scripts/detect-standards.sh" --json "<workspace_root>"
   ```

   脚本读取 `../assets/standards/index.yaml`，发现 `package.json` / `go.mod` **子项目根**（忽略 `node_modules`/`.cursor` 等），按 `detect` 规则求 Level-1（语义见 `../references/preset-management.md`）。  
   - `contains_require`：**仅直接 require**（忽略 `// indirect`）  
   - `contains_require` 同一 rule 数组多项 = **OR**  
   - `require_dirs` / `any_of_files` 相对候选根；`**/*.proto` 等不扫隐藏 IDE 目录  
   - 每分类取 **primary**（第一条 Level-1）部署；同分类多命中时列候选，首版默认 primary，总结报告注明其余根  
   - **示例**：`apps/ui` Vue3+Vite → `frontend-vue3`；`apps/server` Gin+swag → `backend-gin` + `api-swagger`；有 `trpc.group/trpc-go` 但无 `proto/`+`stub/` → **不**命中 `trpc-go`  
   - 命中 `status: planned` → 不部署  
   - 已按 `$PROJECT_DOMAIN` 处理的必选横切（security/quality/skill-spec）：不经本脚本，仍按上一步复制  
3. **确认选择** — 单一 Level-1 匹配请用户确认；多匹配列候选；无匹配 → 该分类不设规范（见下）
4. **匹配策略（策略 B：无 generic 骨架）**

   | 结果 | 条件 | 行为 |
   |------|------|------|
   | Level 1 | detect 脚本 primary 命中且非 `planned` | 复制规范**入口** `.md` 到 `docs/standards/`；若存在同名 `{stem}/` 分册目录则递归复制到 `docs/standards/{stem}/`；列入「当前项目选用」；README「项目事实」记录命中根（如 `apps/ui`） |
   | 未匹配 / planned | 无 Level-1，或仅 planned | **不**生成该分类（frontend/api/backend）规范文件；不进「当前项目选用」；总结报告写明信号 + 贡献引导 |

   **禁止**使用或生成 `*-generic.md`。仓库 `assets/standards/` 不再提供 generic 文件。  
   **禁止**仅靠全文 grep / 人肉记忆子路径选型而不跑 `detect-standards.sh`。

5. **部署** — 复制 Level-1（及必选横切）的**入口** `.md` 到 `docs/standards/`；对每个 `{stem}.md`，若预设库存在 `assets/standards/{stem}/` 则递归复制到 `docs/standards/{stem}/`；动态生成 `docs/standards/README.md`（含加载预算；**项目事实**：各端命中根）
6. **贡献引导**（存在未匹配分类时）— README「未覆盖的技术栈」列出分类与探测信号；总结报告提示编写预设并注册 `index.yaml`
7. **根 README 冲突提示（R5）** — 若 Level-1 含 gin/swagger/vue 等，而根 `README.md` 仍主推冲突栈关键词（如无 gin 信号却大段 trpc-cli/`proto/` 安装，或声明 React 而命中 Vue）：**不改**根 README；在总结报告「后续建议」列出冲突信号，请人工修订

预设管理规范见 `../references/preset-management.md`。

### 第三步-C：开发地图生成

在第三步-B 完成后执行。有两种触发场景：
- `mode=full`：自动接续执行，作为主流程标准步骤
- `mode=targeted, target=dev-map`：仅执行第一步（上下文感知）+ 本步骤 + 第六步（总结报告）

**3-C.1 工具检测**

```
检查 skills/graphify/SKILL.md 是否存在

- 存在：继续执行 3-C.2（调用 graphify skill）
- 不存在：
    在总结报告中记录"graphify skill 未找到（skills/graphify/SKILL.md 不存在），知识图谱功能暂不可使用"
    跳过本步骤剩余内容，继续后续步骤（第三步-D 已移除）
```

**3-C.2 全量生成**

调用 graphify skill，全量生成：

```
/graphify .
```

（graphify skill 已配置输出目录为 docs/dev-map，graph.json → docs/dev-map/graph.json）

**git 策略（配置/规则入库 · 结果不入库）：**

对齐 graphify 官方指引与 F7：提交「让 graphify 跑起来的配置与说明」，**不**提交「跑出来的结果」。

| 类别 | 路径 | git |
|------|------|-----|
| 说明 / 约定 | `docs/dev-map/README.md` | **必须纳入**（`git add`） |
| 忽略清单 | `docs/dev-map/.gitignore` | **必须纳入**（从 `../assets/dev-map.gitignore` 复制/覆写对齐） |
| IDE 规则 | `.cursor/rules/*.mdc`、`.codebuddy/rules/*.md`、`.claude/rules/*.md` 等（见 3-C.5）；Codex 仅 `AGENTS.md` | 按安装布局同步；是否入库随目标仓对 IDE 目录的 gitignore 策略 |
| 图谱结果 | `graph.json`、报告、cache、wiki、可视化等 | **生成到本地，默认不 `git add`**（`.gitignore` 白名单：仅 `README.md` + `.gitignore`） |

硬性约束：
- **禁止**将整个 `docs/dev-map/` 写入任意 `.gitignore`（会丢掉 README / 约定落点）
- **禁止**默认 `git add docs/dev-map/graph.json`（或其它结果文件）；仅当用户**显式**要求入库时例外，并须同步放宽 `.gitignore`
- 交付前确保：`cp ../assets/dev-map.gitignore docs/dev-map/.gitignore`（或内容等价），再 `git add docs/dev-map/README.md docs/dev-map/.gitignore`
- 总结报告注明：结果已本地生成、已被 ignore、未将整目录 ignore

**3-C.3 更新 README.md 与 `.gitignore`**

1. 将 `docs/dev-map/README.md` 内容替换为（读取 `../assets/dev-map-templates.md` 中的模板）。
2. 将 `../assets/dev-map.gitignore` 复制为 `docs/dev-map/.gitignore`（覆盖对齐 canonical 清单）。

**3-C.4 清理旧文件**

若以下文件存在，删除：
```bash
git rm --ignore-unmatch docs/dev-map/source-index.md
git rm --ignore-unmatch docs/dev-map/module-index.md
git rm --ignore-unmatch docs/dev-map/module-dependencies.md
```

**增量更新策略（targeted 模式）：**

| 场景 | 行为 |
|------|------|
| `docs/dev-map/graph.json` 不存在 | 全量生成（同 3-C.2） |
| 文件存在 | 调用 graphify skill 增量更新：`/graphify . --update`，仅重新处理变更文件 |

**3-C.5 IDE 集成配置**

仅在 3-C.1 确认 graphify skill 存在时执行（graphify 不存在则跳过本节）：

```bash
bash "$SKILL_ROOT/harness-engineering/scripts/harness-ide-setup.sh" .
```

读取输出日志，将 `[OK]` / `[MERGED]` / `[SKIP]` / `[WARN]` 状态汇入第六步总结报告。若输出含 `[WARN]`（graphify 未在 PATH 中），在总结报告中提示用户安装 graphify 或确认路径。

---

### 第三步-F：同步 Standards IDE Rules（默认开启）

在第三步-B（及 README 选用表）就绪后、交付前执行。算法见 `../references/standards-compliance.md`；**禁止**另写一套路径/裁剪逻辑。

```bash
bash "$SKILL_ROOT/harness-engineering/scripts/sync-standards-rules.sh" "<workspace_root>"
```

- 按安装布局写入 Cursor `.mdc` / CodeBuddy `.md` / Claude `.md`（fallback → `.agents/rules/` 同时双格式；`.md` 含 Claude `paths`）
- Codex **不**写 `.codex/rules/`（execpolicy 专用）；门闩依赖 `AGENTS.md`
- 仅对「当前项目选用」中存在的 frontend/api/backend/security 生成对应 rule；未选用则删除多余 `standards-*`
- 无 IDE / `.agents` 布局 → 脚本 Skip；总结报告注明「Rules Skip：无 IDE 目录」
- 用户显式拒绝写入 Rules 时 Skip，并记入总结报告

将脚本 stdout（wrote/removed/skip）汇入第六步「Standards Rules」节。

---

### 第三步-D：工作流文档生成 — **已移除**

harness-engineering **不再**生成或同步 `docs/workflow.md`，**不再**在 AGENTS.md 写入 workflow-agent /「不允许跳过」。
迭代工作流由目标仓自行维护（或使用独立 Skill），不在本 skill 职责内。
本步永久 Skip：将 Todo `gen-3d` 标为 `completed` 后直接进入第三步-E。

---

### 第三步-E：初始化业务规范空间（幂等，永不覆写）

在第三步-B 完成后执行。与 `docs/standards/`（预设单向覆写）语义**相反**：`docs/business-standards/` 为**用户自有**空间，harness-gardening 永不覆写。

**执行逻辑：**

1. 判定目录是否存在：

   ```bash
   test -d docs/business-standards && echo "SKIP-已存在保留用户内容" || echo "INIT-创建骨架"
   ```

2. 若 `docs/business-standards/` **不存在** → 创建目录并生成 `README.md` 索引骨架，包含：
   - frontmatter 元数据填写说明（`tags` string 数组 + `scenarios` string 数组）；
   - 空的「规范索引」表（含一行示例，展示 tags/scenarios 写法）；
   - 「harness-gardening 永不覆写本目录」的显式声明与「按 tags/scenarios 选择性加载」的说明。
3. 若 `docs/business-standards/` **已存在** → **保留用户内容，不覆写、不删除**（严禁比对预设或覆盖，与 `docs/standards/` 覆写语义相反）。
4. 在生成的 `docs/standards/README.md`「Agent 加载策略」中登记业务规范空间条目：agent 按 tags/scenarios 选择性加载 `docs/business-standards/`（非全量强制）。

---

### 第四步：定向修正（按需）

当 `mode=targeted` 或用户指定修改某个组件时：

1. 读取该组件的现有文档
2. 根据用户要求更新内容
3. 检查是否影响其他组件（如修改架构约束可能影响工具能力约束规则），有关联影响则提示用户

**技术规范的定向修正：**

`docs/standards/` 以 `../assets/standards/` 预设为**唯一权威来源**，不支持用户定制：

1. **更换预设** — 重新执行第三步-B 检测/选择流程，覆写对应文件
2. **同步预设更新** — 对比 `docs/standards/` 与预设（入口 `.md` + 同名 `{stem}/` 分册目录），差异项直接覆写
3. **定制规范内容** — 修改 `../assets/standards/` 中的预设文件（而非 `docs/standards/`）

**增量更新策略：**

| 场景 | 行为 |
|------|------|
| `docs/standards/` 不存在 | 全量生成 |
| 文件存在且与预设一致 | 跳过 |
| 文件存在但与预设不一致 | 用预设覆写（预设为权威） |
| 检测到新技术栈 | 为新 category 复制预设，不影响已有 |
| 检测到技术栈已移除 | 不自动删除，提示用户确认 |

### 第五步：工具依赖自审

在输出总结报告之前，**必须**完成以下对账，不得跳过：

1. **对账** — 将本次已检查的工具条目与 `../references/tool-dependencies.md` §四中当前场景的检查清单逐项比对，找出满足以下任一条件的条目：
   - 清单中有、但本次未执行检查的
   - 已记录但状态为"未知/待检测"的

2. **补检** — 对每个对账缺口条目**立即发起检查**（执行动作，不得输出"需要检查"类说明，不得询问用户）：
   - MCP：对该 MCP 发起 tool call（接口见 §一对应"环境检查"行）
   - Skill：`test -f $SKILL_ROOT/<skill-name>/SKILL.md`
   - CLI：执行 `command -v <cmd>`（技术栈专属 CLI 须先确认检测条件满足）
   - 配置：§三 — 检查文件存在与关键字段

3. **合并** — 将补检结果追加到第一步暂存的缺口列表，然后进入第五步

### 第六步：生成总结

**交付前硬门禁（必须执行）：**

```bash
bash "$SKILL_ROOT/harness-engineering/scripts/harness-verify.sh" "<workspace_root>"
```

失败则：不得宣布生成完成；按报错修改产物（去掉虚假「已就绪」、删除 workflow 强制段、把 TODO 骨架移出「当前选用」等）后重跑，直到 exit 0。

完成后按 `references/report-template.md` 格式向用户输出总结报告（含已完成组件、待补充内容、环境工具缺口、后续建议）。

## 质量检查

生成文档后、总结报告前，执行以下检查（含上节 harness-verify）：

1. **完整性** — AGENTS.md、glossary.md、五大组件文档、技术规范文档、dev map 四文件是否全部生成
2. **一致性** — `tooling.md` §1.0 Skill 清单与 `$SKILL_ROOT/*/SKILL.md`（仅顶层）一致（`$SKILL_ROOT` 由第一步项目类型识别确定）；glossary.md 覆盖核心术语；组件间约束自洽
3. **披露层次** — AGENTS.md → harness/README.md → 组件文档；AGENTS.md → standards/README.md → 技术规范导航畅通；AGENTS.md → dev-map/README.md 引用正确
4. **可操作性** — 规范条目足够具体，能直接指导实施
5. **可维护性** — 标注了待补充内容和后续改进方向
6. **规范完整性** — 每个检测到的技术栈都有对应规范文档（完整预设或通用骨架）；按 `$PROJECT_DOMAIN` 检查必选规范：
   - `$PROJECT_DOMAIN == code-project` 或 `mixed`：`security-bk-redlines.md`、`quality-code-review.md` 必须存在于 `docs/standards/`
   - `$PROJECT_DOMAIN == skill-tooling` 或 `mixed`：`skill-spec.md` 必须存在于 `docs/standards/`
7. **规范与架构一致性** — 技术规范中的架构约束与 `docs/harness/architectural-constraints.md` 定义一致
8. **工具依赖完备性** — `tooling.md` 依赖表与 `../references/tool-dependencies.md` 一致；环境检查已通过「工具依赖自审」步骤补全，总结报告列出最终缺口
9. **Dev Map 与 IDE 集成完整性** — 若 graphify 可用：本地已有 `docs/dev-map/graph.json`；`README.md` 与 `.gitignore`（对齐 `assets/dev-map.gitignore`）已入库且**未**整目录 ignore；结果文件未误 `git add`；旧三文件已删除；`harness-ide-setup.sh` 已执行且无非预期错误——按布局写入 `graphify.mdc` / `graphify.md`（含实体 `.claude/rules`；fallback → `.agents/rules/`），`.codebuddy` 侧 `settings.json` hook-guard 齐全。若 graphify 不可用：在总结报告中说明原因和安装方式
10. **无 workflow 接入** — AGENTS.md / docs/harness 不含 workflow-agent、「不允许跳过工作流」、「## 开发工作流」；**不**要求生成 `docs/workflow.md`；交付前运行 `bash "$SKILL_ROOT/harness-engineering/scripts/harness-verify.sh" <workspace_root>`，失败则不得宣布完成
11. **业务规范空间完整性** — `docs/business-standards/README.md` 存在且含 frontmatter 元数据说明与「规范索引」表；若目录已存在，确认用户内容未被覆写
12. **Standards 门闩与 Rules** — AGENTS 含「编码前必读（门闩）」且强调按节/预算；`docs/standards/README.md` 含「Agent 加载步骤（强制）」与「加载预算」（或等价「按节」指引）；已执行 `sync-standards-rules.sh`（或用户拒绝已记入报告）；未选用分类不得残留对应 `standards-*` rule

## 参考资源

| 文件 | 用途 | 何时读取 |
|------|------|---------|
| `../references/project-type-detection.md` | 项目类型识别与 $SKILL_ROOT 检测 | 第一步-0（必须最先） |
| `../assets/harness-spec-template.md` | 规范文档模板 | 生成文档时 |
| `../assets/dev-map-templates.md` | 开发地图模板（README.md 目标内容） | 第三步-C 生成 dev map 时 |
| `../assets/dev-map.gitignore` | 开发地图结果级 ignore（复制为 docs/dev-map/.gitignore） | 第三步-C |
| `../assets/standards/index.yaml` | 技术规范预设索引 | 技术规范处理时 |
| `../assets/standards/*.md` | 技术规范预设**入口**文件 | 匹配后复制到目标项目 |
| `../assets/standards/{stem}/` | 大预设的分册目录（与 `{stem}.md` 成对） | 与入口一并递归复制 |
| `../references/best-practices.md` | 五大组件最佳实践详解 | 生成各组件规范时 |
| `../references/question-bank.md` | 交互提问库 | 信息不足时 |
| `../references/tool-dependencies.md` | Agent 工具依赖权威清单 | 第一步工具依赖扫描 + 第五步自审 |
| `../references/preset-management.md` | 技术规范预设管理规范 | 管理/扩展预设库时 |
| `../references/agents-merge.md` | 根 AGENTS 先理解再合并 | 第三步写 AGENTS；总结「AGENTS 理解」节 |
| `../references/agents-work-units.md` | 工作单元发现与根索引 | 第三步写根前；总结「工作单元 AGENTS」节 |
| `../references/standards-compliance.md` | Standards 门闩与 IDE Rules 同步算法 | 第三步-F |
| `../scripts/sync-standards-rules.sh` | 按选用表渲染/裁剪 IDE Rules | 第三步-F |
| `../assets/ide-rules/` | Cursor / CodeBuddy / Claude Rules 模板（Codex 无 instruction rules） | 第三步-F |
| `references/report-template.md`（本 skill 私有） | 总结报告格式模板 | 生成第六步总结时 |
| `../assets/workflow-template.md` | **DEPRECATED**（不再使用） | 第三步-D 已移除 |

## 清单验收

在输出总结报告后，检查 TodoWrite 清单：

- **全部 `completed`** → 执行完毕，正常退出
- **有 `pending` / `in_progress` 项** → 立即补充执行对应步骤，直至清单全绿再退出
