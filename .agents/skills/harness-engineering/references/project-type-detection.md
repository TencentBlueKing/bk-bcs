# 项目类型识别与 Skill 安装根检测

执行任何 harness-engineering 子 skill **前必须首先完成**此步骤，结果影响后续所有路径。

详细安装根说明见 [`skill-install-root.md`](skill-install-root.md)；项目自有工具见 [`project-owned-tools.md`](project-owned-tools.md)。

## 第一步：确定 `${SKILL_INSTALL_ROOT}`（亦称 `$SKILL_ROOT`）

按以下优先级依次检查，命中第一个即确定安装根：

| 检查 | 项目类型 | `${SKILL_INSTALL_ROOT}` |
|------|---------|-------------------------|
| `.agents/skills/*/SKILL.md`（或 harness-engineering） | 接入仓（推荐布局） | `.agents/skills` |
| `skills/harness-engineering/SKILL.md` | 开发仓（ai-practice 本仓） | `skills` |
| `.cursor` / `.codebuddy` / `.claude` / `.codex` → `.agents` 软链 | 接入仓 fallback | `.agents/skills` |
| `.<ide>/skills/harness-engineering/SKILL.md`（实体目录） | 接入仓 | `.<ide>/skills`；总结报告 WARN 建议迁移 |

`$SKILL_ROOT` 与 `${SKILL_INSTALL_ROOT}` 同义：执行期间路径变量，所有对项目级 skill 的文件引用均使用该变量。  
通用框架 skill（systematic-debugging、brainstorming 等）可位于用户级 IDE 目录，路径不受该变量约束。

**禁止**在生成文档中写死「唯一权威路径=.codebuddy/skills」或「=.cursor/skills」。

## 第二步：确定 `$PROJECT_DOMAIN`（产出物性质层）

在确定安装根之后，继续检测项目的**领域类型**，结果作为 `$PROJECT_DOMAIN` 变量，
影响技术规范预设选择、质量检查规则和提问策略。

按以下规则判定，两类信号同时检查：

| 信号 | 检测方式 | 说明 |
|------|---------|------|
| **Skill 信号** | 项目中存在路径**各段均非隐藏**（不以 `.` 开头）的 `SKILL.md` 文件 | 隐藏路径下的 `SKILL.md`（如 `.cursor/skills/*/SKILL.md`）是安装的开发工具，不算产出；非隐藏路径下的 `SKILL.md`（如 `skills/*/SKILL.md`、`my-skills/*/SKILL.md`）才代表项目自身产出 Skill |
| **代码信号** | 项目根目录下存在 `go.mod`、`package.json` 或 `pyproject.toml` | 有代码构建文件 |

检测命令参考：
```bash
# Skill 信号：查找所有路径段均非隐藏的 SKILL.md
find . -name "SKILL.md" | grep -v '/\.' | head -1

# 代码信号：检查根目录构建文件
ls go.mod package.json pyproject.toml 2>/dev/null | head -1
```

根据两类信号的组合确定 `$PROJECT_DOMAIN`：

| Skill 信号 | 代码信号 | `$PROJECT_DOMAIN` | 典型场景 |
|-----------|---------|------------------|---------|
| ✅ 有 | ❌ 无 | `skill-tooling` | ai-practice 本仓：`skills/*/SKILL.md` 存在，无 go.mod |
| ❌ 无 | ✅ 有 | `code-project` | 接入仓：`.agents/skills/*/SKILL.md`（隐藏），有 go.mod |
| ✅ 有 | ✅ 有 | `mixed` | 接入仓自建了非隐藏目录的 Skill，同时有代码 |
| ❌ 无 | ❌ 无 | `code-project` | 无明确信号，默认按代码项目处理 |

`$PROJECT_DOMAIN` 确定后，后续步骤按以下差异化策略执行：

| `$PROJECT_DOMAIN` | 技术规范预设 | 代码安全/评审规范 |
|-------------------|------------|----------------|
| `skill-tooling` | 选用 `skill-tooling` 类别预设（skill-spec.md） | **不适用**，跳过 |
| `code-project` | 选用前端/API/后端预设 + security/quality 必选 | **必须**包含 |
| `mixed` | 两套预设均处理 | **必须**包含（代码部分） |
