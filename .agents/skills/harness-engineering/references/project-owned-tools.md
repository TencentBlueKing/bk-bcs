# 项目自有工具

接入仓除 Harness 基线（`tool-dependencies.md` 白名单）外，可有团队自研 Skill/MCP。  
它们进入 `docs/harness/tooling.md` 的 **「项目自有工具」** 节，不改治理仓权威清单。

## 准入（唯一）

**安装布局路径 ∩ 已被 git 跟踪** → 项目自有。  
磁盘存在但未进 git（含 `install-to-target.sh` 对未跟踪布局路径的 gitignore 安装产物）→ **不入库**。  
已跟踪的布局路径（如业务仓入库的 `.claude/skills`）不得被安装脚本 ignore。

名已在 `tool-dependencies.md` 白名单 → 只进 **Harness 基线**节，不重复进项目自有。

## 路径（仓根 + monorepo 子树）

仓根（对齐 `scripts/install-to-target.sh`）：

```text
.agents/skills/<name>/SKILL.md
.cursor/skills/<name>/SKILL.md
.codebuddy/skills/<name>/SKILL.md
.claude/skills/<name>/SKILL.md
.codex/skills/<name>/SKILL.md
skills/<name>/SKILL.md          # 仅治理/开发仓源

.agents/mcp.json
.cursor/mcp.json
```

**monorepo 组件级**（任意深度，路径段中含下列布局之一）：

```text
<component>/.agents/skills/<name>/SKILL.md
<component>/.cursor/skills/<name>/SKILL.md
<component>/.codebuddy/skills/<name>/SKILL.md
<component>/.claude/skills/<name>/SKILL.md
<component>/.codex/skills/<name>/SKILL.md
```

示例：`apps/bkms-server/.agents/skills/bkms-dev-ginapi/SKILL.md`、`apps/ui/.agents/skills/vue-typescript-type-checker/SKILL.md`。

扫描（generating / doctor / gardening 对账共用语义）：

```bash
# 仓根
git ls-files -- \
  '.agents/skills/*/SKILL.md' \
  '.cursor/skills/*/SKILL.md' \
  '.codebuddy/skills/*/SKILL.md' \
  '.claude/skills/*/SKILL.md' \
  '.codex/skills/*/SKILL.md' \
  'skills/*/SKILL.md'

# 含子树：凡匹配「…/.(agents|cursor|codebuddy|claude|codex)/skills/<name>/SKILL.md」
git ls-files | grep -E '(^|/)\.(agents|cursor|codebuddy|claude|codex)/skills/[^/]+/SKILL\.md$'
```

排除：`node_modules/`、`.git/` 等（`git ls-files` 本身不含未跟踪与忽略路径）。

## tooling.md 分节

| 节 | 来源 | 再生成 |
|----|------|--------|
| Harness 基线 | `tool-dependencies.md` | 可按权威清单重写 |
| 项目自有工具 | `git ls-files` ∩ 上表（含 monorepo 子树），且非白名单 | **不得清空**；可合并扫描，手工行保留 |

表头：`名称 | 用途`（无环境状态列；**不写**类型/路径/检测方式，也不写「准入 / 见某某文档」等引用说明）。

用途来源：

- Skill：读对应 `SKILL.md` frontmatter `description`，压成一行（去换行；过长截断至约 120 字）
- 若路径在子树（非仓根安装布局）：用途可前缀组件目录，如 `apps/bkms-server — …`（取自 `…/skills/` 之前的相对路径），便于区分同名 skill
- MCP：配置内 description（若有），否则「（见 mcp 配置）」；允许手工改用途列，再生成按名称合并时保留手工用途（若新扫描无更好 description）

探活由 `harness-doctor` 按名称在**仓根 + 子树**安装布局中解析，**不依赖** tooling 表内路径列。

## harness-doctor（必查）

1. 若 tooling 有「项目自有工具」节 → 按名称逐条在安装布局（含 monorepo 子树）中 present/absent  
2. 若无节 → `git ls-files` ∩ 上表路径，有则探、无则 Skip  
3. Skill：匹配布局的 `SKILL.md` 存在且（建议）git 跟踪；MCP：清单存在且含 server 键  
4. 结果仅 stdout；单条 absent → WARN；脚本 exit 0  

## generating / gardening

- generating：扫描跟踪项（含 `apps/*/.agents/skills` 等子树）写入/合并项目自有节（仅名称+用途）；**禁止**写入未跟踪路径；**禁止**在 tooling.md 写入指向 `project-owned-tools.md` / `install-to-target.sh` / 权威清单路径等引用说明块；再生成**不得清空**该节已有行  
- gardening 维度 7：基线对白名单；项目自有名称对 git 跟踪安装布局（含 monorepo 子树）对账；未跟踪多装 Skip；再生成不抹项目自有节  
